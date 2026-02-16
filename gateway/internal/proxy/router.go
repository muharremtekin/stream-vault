package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/rs/zerolog/log"
	"github.com/streamvault/gateway/internal/discovery"
	"github.com/streamvault/gateway/internal/metrics"
	"github.com/streamvault/gateway/internal/middleware"
	"github.com/streamvault/gateway/internal/resilience"
)

// Route maps a URL path prefix to an upstream service.
type Route struct {
	// PathPrefix is the URL prefix that triggers this route (e.g. "/api/auth").
	PathPrefix string
	// ServiceName is the logical service name resolved through discovery.
	ServiceName string
	// StripPrefix controls whether PathPrefix is stripped before proxying.
	StripPrefix bool
	// Streaming enables a streaming-optimised reverse proxy that flushes
	// response bytes immediately instead of buffering (FlushInterval: -1).
	Streaming bool
}

// DefaultRoutes returns the standard set of gateway routes that map public
// API paths to internal micro-services.
func DefaultRoutes() []Route {
	return []Route{
		{
			PathPrefix:  "/api/auth",
			ServiceName: "user-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/users",
			ServiceName: "user-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/profile",
			ServiceName: "user-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/watchlist",
			ServiceName: "user-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/catalog",
			ServiceName: "catalog-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/stream",
			ServiceName: "streaming-service",
			StripPrefix: false,
			Streaming:   true,
		},
		{
			PathPrefix:  "/stream",
			ServiceName: "streaming-service",
			StripPrefix: false,
			Streaming:   true,
		},
		{
			PathPrefix:  "/api/encoding",
			ServiceName: "encoding-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/search",
			ServiceName: "search-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/recommendations",
			ServiceName: "recommendation-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/plans",
			ServiceName: "subscription-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/subscriptions",
			ServiceName: "subscription-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/api/notifications",
			ServiceName: "notification-service",
			StripPrefix: false,
		},
		{
			PathPrefix:  "/ws/notifications",
			ServiceName: "notification-service",
			StripPrefix: false,
			Streaming:   true,
		},
	}
}

// Router holds the route table and builds an http.ServeMux that forwards
// requests to the appropriate upstream service.
type Router struct {
	routes     []Route
	resolver   *discovery.Resolver
	manager    *UpstreamManager
	resilience *resilience.Resilience
}

// NewRouter creates a Router with the given routes, discovery resolver,
// upstream connection manager, and resilience patterns. If res is nil,
// requests are proxied without circuit breaker, bulkhead, or timeout.
func NewRouter(routes []Route, resolver *discovery.Resolver, manager *UpstreamManager, res *resilience.Resilience) *Router {
	return &Router{
		routes:     routes,
		resolver:   resolver,
		manager:    manager,
		resilience: res,
	}
}

// Handler builds an http.Handler that matches incoming requests against the
// route table and proxies them to the resolved upstream address. Unmatched
// paths receive a 404 response.
func (rt *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	for _, route := range rt.routes {
		r := route // capture loop variable

		var handler http.HandlerFunc
		if rt.resilience != nil {
			handler = rt.resilientHandler(r)
		} else {
			handler = rt.plainHandler(r)
		}

		// Register both with and without trailing slash to avoid 301 redirects.
		mux.Handle(r.PathPrefix+"/", handler)
		mux.Handle(r.PathPrefix, handler)
	}

	return mux
}

// plainHandler proxies requests without resilience patterns (backward compat).
func (rt *Router) plainHandler(r Route) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addr, err := rt.resolver.Resolve(r.ServiceName)
		if err != nil {
			log.Error().Err(err).Str("service", r.ServiceName).Msg("service resolution failed")
			middleware.WriteErrorResponse(w, http.StatusBadGateway, "service unavailable")
			return
		}

		proxy, err := rt.getProxy(r, addr)
		if err != nil {
			log.Error().Err(err).Str("address", addr).Msg("failed to create reverse proxy")
			middleware.WriteErrorResponse(w, http.StatusBadGateway, "bad gateway")
			return
		}

		proxy.ServeHTTP(w, req)
	}
}

// resilientHandler wraps the proxy call with bulkhead → timeout → circuit breaker.
func (rt *Router) resilientHandler(r Route) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		svc := r.ServiceName

		// 1. Bulkhead: fast-fail if service concurrency limit reached.
		if !rt.resilience.AcquireBulkhead(svc, req.Context()) {
			log.Warn().Str("service", svc).Msg("bulkhead full, rejecting request")
			middleware.WriteErrorResponse(w, http.StatusServiceUnavailable, "service overloaded")
			return
		}
		defer rt.resilience.ReleaseBulkhead(svc)

		// 2. Timeout: set per-service context deadline.
		timeout := rt.resilience.GetTimeout(svc)
		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()
		req = req.WithContext(ctx)

		// 3. Circuit breaker: wraps the resolve + proxy call.
		err := rt.resilience.Execute(svc, func() error {
			addr, resolveErr := rt.resolver.Resolve(svc)
			if resolveErr != nil {
				return fmt.Errorf("resolve %s: %w", svc, resolveErr)
			}

			proxy, proxyErr := rt.getProxy(r, addr)
			if proxyErr != nil {
				return fmt.Errorf("proxy %s: %w", addr, proxyErr)
			}

			sc := &resilience.StatusCapture{ResponseWriter: w, Code: http.StatusOK}
			proxy.ServeHTTP(sc, req)

			if sc.Code >= 500 {
				return fmt.Errorf("upstream %s returned %d", svc, sc.Code)
			}
			return nil
		})

		if err != nil {
			if resilience.IsCircuitError(err) {
				log.Warn().Str("service", svc).Msg("circuit breaker open, fast-failing")
				// Try graceful degradation fallback before generic 503.
				if !resilience.ServeFallback(w, req, svc) {
					middleware.WriteErrorResponse(w, http.StatusServiceUnavailable, "service temporarily unavailable")
				}
				return
			}
			// Check if the error was a context timeout (deadline exceeded).
			if req.Context().Err() != nil {
				metrics.GatewayUpstreamTimeoutsTotal.WithLabelValues(svc).Inc()
			}
			// For non-circuit errors the proxy already wrote the response to the client,
			// so we only log the failure.
			log.Debug().Err(err).Str("service", svc).Msg("upstream request failed")
		}
	}
}

// getProxy returns the appropriate reverse proxy (streaming or regular).
func (rt *Router) getProxy(r Route, addr string) (*httputil.ReverseProxy, error) {
	if r.Streaming {
		return rt.manager.GetStreamingProxy(addr, r.StripPrefix, r.PathPrefix)
	}
	return rt.manager.GetProxy(addr, r.StripPrefix, r.PathPrefix)
}
