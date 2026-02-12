package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// UpstreamManager maintains a pool of reverse proxies keyed by upstream
// address. It reuses proxies to avoid re-creating transports on every request.
type UpstreamManager struct {
	mu      sync.RWMutex
	proxies map[string]*httputil.ReverseProxy

	// Transport is the shared HTTP transport for all upstream connections.
	Transport http.RoundTripper
}

// NewUpstreamManager creates an UpstreamManager with a sensible default
// transport configuration.
func NewUpstreamManager() *UpstreamManager {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	return &UpstreamManager{
		proxies:   make(map[string]*httputil.ReverseProxy),
		Transport: transport,
	}
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

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":"upstream service unavailable"}`))
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
