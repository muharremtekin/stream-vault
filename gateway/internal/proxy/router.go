package proxy

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/streamvault/gateway/internal/discovery"
)

// Route maps a URL path prefix to an upstream service.
type Route struct {
	// PathPrefix is the URL prefix that triggers this route (e.g. "/api/auth").
	PathPrefix string
	// ServiceName is the logical service name resolved through discovery.
	ServiceName string
	// StripPrefix controls whether PathPrefix is stripped before proxying.
	StripPrefix bool
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
	}
}

// Router holds the route table and builds an http.ServeMux that forwards
// requests to the appropriate upstream service.
type Router struct {
	routes   []Route
	resolver *discovery.Resolver
	manager  *UpstreamManager
}

// NewRouter creates a Router with the given routes, discovery resolver, and
// upstream connection manager.
func NewRouter(routes []Route, resolver *discovery.Resolver, manager *UpstreamManager) *Router {
	return &Router{
		routes:   routes,
		resolver: resolver,
		manager:  manager,
	}
}

// Handler builds an http.Handler that matches incoming requests against the
// route table and proxies them to the resolved upstream address. Unmatched
// paths receive a 404 response.
func (rt *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	for _, route := range rt.routes {
		r := route // capture loop variable

		handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			addr, err := rt.resolver.Resolve(r.ServiceName)
			if err != nil {
				log.Error().Err(err).Str("service", r.ServiceName).Msg("service resolution failed")
				http.Error(w, `{"error":"service unavailable"}`, http.StatusBadGateway)
				return
			}

			proxy, err := rt.manager.GetProxy(addr, r.StripPrefix, r.PathPrefix)
			if err != nil {
				log.Error().Err(err).Str("address", addr).Msg("failed to create reverse proxy")
				http.Error(w, `{"error":"bad gateway"}`, http.StatusBadGateway)
				return
			}

			proxy.ServeHTTP(w, req)
		})

		// Register both with and without trailing slash to avoid 301 redirects.
		mux.Handle(r.PathPrefix+"/", handler)
		mux.Handle(r.PathPrefix, handler)
	}

	return mux
}
