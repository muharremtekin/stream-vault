package proxy

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/streamvault/gateway/internal/config"
	"github.com/streamvault/gateway/internal/middleware"
)

// UpstreamManager maintains a pool of reverse proxies keyed by upstream
// address. It reuses proxies to avoid re-creating transports on every request.
type UpstreamManager struct {
	mu      sync.RWMutex
	proxies map[string]*httputil.ReverseProxy

	// Transport is the shared HTTP transport for all upstream connections.
	Transport http.RoundTripper
}

// NewUpstreamManager creates an UpstreamManager with a configurable retry
// transport. GET/HEAD requests are automatically retried on 502/503/504
// errors or network failures with exponential backoff and jitter.
func NewUpstreamManager(retryCfg config.RetryConfig) *UpstreamManager {
	baseTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	// Apply sensible defaults if config values are zero.
	if retryCfg.MaxRetries <= 0 {
		retryCfg.MaxRetries = 3
	}
	if retryCfg.InitialBackoff <= 0 {
		retryCfg.InitialBackoff = 100 * time.Millisecond
	}
	if retryCfg.MaxBackoff <= 0 {
		retryCfg.MaxBackoff = 2 * time.Second
	}
	if retryCfg.Multiplier <= 0 {
		retryCfg.Multiplier = 2.0
	}
	if retryCfg.JitterFraction <= 0 {
		retryCfg.JitterFraction = 0.2
	}

	retryTransport := &RetryTransport{
		Base:           otelhttp.NewTransport(baseTransport),
		MaxRetries:     retryCfg.MaxRetries,
		InitialBackoff: retryCfg.InitialBackoff,
		MaxBackoff:     retryCfg.MaxBackoff,
		Multiplier:     retryCfg.Multiplier,
		JitterFraction: retryCfg.JitterFraction,
	}

	return &UpstreamManager{
		proxies:   make(map[string]*httputil.ReverseProxy),
		Transport: retryTransport,
	}
}

// RetryTransport wraps an http.RoundTripper and retries idempotent requests
// (GET, HEAD) on retryable status codes (502, 503, 504) or network failures
// with exponential backoff and jitter.
type RetryTransport struct {
	Base           http.RoundTripper
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	JitterFraction float64 // 0.0–1.0, fraction of delay to randomise
}

// retryableStatus reports whether the HTTP status code warrants a retry.
func retryableStatus(code int) bool {
	return code == http.StatusBadGateway ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusGatewayTimeout
}

// nonRetryableStatus reports whether the status code must never be retried.
func nonRetryableStatus(code int) bool {
	switch code {
	case http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusConflict:
		return true
	}
	return false
}

// RoundTrip implements http.RoundTripper with retry logic for safe methods.
func (rt *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only retry idempotent methods to avoid duplicate side effects.
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return rt.Base.RoundTrip(req)
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= rt.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rt.backoff(attempt)
			log.Warn().
				Int("attempt", attempt).
				Str("path", req.URL.Path).
				Dur("delay", delay).
				Msg("retrying upstream request")

			select {
			case <-time.After(delay):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		resp, err = rt.Base.RoundTrip(req)
		if err != nil {
			// Check if context was cancelled — no point retrying.
			if req.Context().Err() != nil {
				return nil, req.Context().Err()
			}
			continue // network error, retry
		}

		// Non-retryable 4xx — return immediately.
		if nonRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Retryable 5xx — close body and retry.
		if retryableStatus(resp.StatusCode) && attempt < rt.MaxRetries {
			resp.Body.Close()
			continue
		}

		return resp, nil
	}

	return resp, err
}

// backoff computes the delay for the given retry attempt using exponential
// backoff with jitter. The delay is capped at MaxBackoff.
func (rt *RetryTransport) backoff(attempt int) time.Duration {
	delay := float64(rt.InitialBackoff)
	for i := 1; i < attempt; i++ {
		delay *= rt.Multiplier
	}

	if time.Duration(delay) > rt.MaxBackoff {
		delay = float64(rt.MaxBackoff)
	}

	// Add jitter: delay ± (jitter_fraction * delay)
	if rt.JitterFraction > 0 {
		jitter := delay * rt.JitterFraction
		delay = delay - jitter + rand.Float64()*2*jitter //nolint:gosec // jitter doesn't need crypto rand
	}

	return time.Duration(delay)
}

// GetProxy returns a reverse proxy for the given upstream address. If
// stripPrefix is true the matching pathPrefix is removed from the forwarded
// request path. Proxies are cached and reused across requests.
func (um *UpstreamManager) GetProxy(address string, stripPrefix bool, pathPrefix string) (*httputil.ReverseProxy, error) {
	cacheKey := fmt.Sprintf("%s|%v|%s", address, stripPrefix, pathPrefix)

	um.mu.RLock()
	if proxy, ok := um.proxies[cacheKey]; ok {
		um.mu.RUnlock()
		return proxy, nil
	}
	um.mu.RUnlock()

	// Ensure address has a scheme.
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	target, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("parsing upstream url %q: %w", address, err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			if stripPrefix && pathPrefix != "" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, pathPrefix)
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}
			}

			// Preserve the original path when not stripping.
			if req.URL.RawPath != "" {
				if stripPrefix && pathPrefix != "" {
					req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, pathPrefix)
					if req.URL.RawPath == "" {
						req.URL.RawPath = "/"
					}
				}
			}
		},
		Transport: um.Transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error().
				Err(err).
				Str("upstream", address).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("upstream request failed")

			middleware.WriteErrorResponse(w, http.StatusBadGateway, "upstream service unavailable")
		},
		ModifyResponse: func(resp *http.Response) error {
			log.Debug().
				Str("upstream", address).
				Int("status", resp.StatusCode).
				Msg("upstream response received")
			return nil
		},
	}

	um.mu.Lock()
	um.proxies[cacheKey] = proxy
	um.mu.Unlock()

	return proxy, nil
}

// GetStreamingProxy returns a reverse proxy optimised for streaming responses.
// It behaves like GetProxy but sets FlushInterval to -1 so that response bytes
// are forwarded to the client immediately instead of being buffered. This is
// critical for HLS manifests and video segments.
func (um *UpstreamManager) GetStreamingProxy(address string, stripPrefix bool, pathPrefix string) (*httputil.ReverseProxy, error) {
	cacheKey := fmt.Sprintf("%s|%v|%s|streaming", address, stripPrefix, pathPrefix)

	um.mu.RLock()
	if proxy, ok := um.proxies[cacheKey]; ok {
		um.mu.RUnlock()
		return proxy, nil
	}
	um.mu.RUnlock()

	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	target, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("parsing upstream url %q: %w", address, err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			if stripPrefix && pathPrefix != "" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, pathPrefix)
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}
			}

			if req.URL.RawPath != "" {
				if stripPrefix && pathPrefix != "" {
					req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, pathPrefix)
					if req.URL.RawPath == "" {
						req.URL.RawPath = "/"
					}
				}
			}
		},
		Transport:     um.Transport,
		FlushInterval: -1,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error().
				Err(err).
				Str("upstream", address).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("upstream request failed")

			middleware.WriteErrorResponse(w, http.StatusBadGateway, "upstream service unavailable")
		},
		ModifyResponse: func(resp *http.Response) error {
			log.Debug().
				Str("upstream", address).
				Int("status", resp.StatusCode).
				Msg("upstream response received")
			return nil
		},
	}

	um.mu.Lock()
	um.proxies[cacheKey] = proxy
	um.mu.Unlock()

	return proxy, nil
}

// RemoveProxy evicts a cached proxy for the given address. This is useful
// when a service instance becomes unhealthy and should no longer be used.
func (um *UpstreamManager) RemoveProxy(address string) {
	um.mu.Lock()
	defer um.mu.Unlock()

	for key := range um.proxies {
		if strings.HasPrefix(key, address+"|") {
			delete(um.proxies, key)
		}
	}
}
