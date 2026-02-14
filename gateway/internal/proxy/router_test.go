package proxy

import (
	"testing"
)

func TestDefaultRoutes_ReturnsExpectedCount(t *testing.T) {
	routes := DefaultRoutes()
	if len(routes) != 12 {
		t.Errorf("expected 12 routes, got %d", len(routes))
	}
}

func TestDefaultRoutes_AllRoutesHaveRequiredFields(t *testing.T) {
	routes := DefaultRoutes()
	for _, r := range routes {
		if r.PathPrefix == "" {
			t.Error("route has empty PathPrefix")
		}
		if r.ServiceName == "" {
			t.Errorf("route %q has empty ServiceName", r.PathPrefix)
		}
	}
}

func TestDefaultRoutes_StreamingServiceRoutes(t *testing.T) {
	routes := DefaultRoutes()
	streamingRoutes := make([]Route, 0)
	for _, r := range routes {
		if r.ServiceName == "streaming-service" {
			streamingRoutes = append(streamingRoutes, r)
		}
	}

	if len(streamingRoutes) != 2 {
		t.Fatalf("expected 2 streaming-service routes, got %d", len(streamingRoutes))
	}

	for _, r := range streamingRoutes {
		if !r.Streaming {
			t.Errorf("streaming-service route %q should have Streaming=true", r.PathPrefix)
		}
	}
}

func TestDefaultRoutes_EncodingServiceRoute(t *testing.T) {
	routes := DefaultRoutes()
	found := false
	for _, r := range routes {
		if r.ServiceName == "encoding-service" {
			found = true
			if r.Streaming {
				t.Errorf("encoding-service route %q should not have Streaming=true", r.PathPrefix)
			}
			if r.PathPrefix != "/api/encoding" {
				t.Errorf("expected /api/encoding prefix, got %q", r.PathPrefix)
			}
		}
	}
	if !found {
		t.Error("encoding-service route not found")
	}
}

func TestDefaultRoutes_ContainsExpectedPrefixes(t *testing.T) {
	routes := DefaultRoutes()
	expected := []string{
		"/api/auth",
		"/api/users",
		"/api/profile",
		"/api/watchlist",
		"/api/catalog",
		"/api/stream",
		"/stream",
		"/api/encoding",
		"/api/search",
		"/api/recommendations",
		"/api/plans",
		"/api/subscriptions",
	}

	prefixSet := make(map[string]bool)
	for _, r := range routes {
		prefixSet[r.PathPrefix] = true
	}

	for _, prefix := range expected {
		if !prefixSet[prefix] {
			t.Errorf("expected route prefix %q not found", prefix)
		}
	}
}

func TestDefaultRoutes_SearchServiceRoute(t *testing.T) {
	routes := DefaultRoutes()
	found := false
	for _, r := range routes {
		if r.ServiceName == "search-service" {
			found = true
			if r.Streaming {
				t.Errorf("search-service route %q should not have Streaming=true", r.PathPrefix)
			}
			if r.PathPrefix != "/api/search" {
				t.Errorf("expected /api/search prefix, got %q", r.PathPrefix)
			}
		}
	}
	if !found {
		t.Error("search-service route not found")
	}
}

func TestDefaultRoutes_RecommendationServiceRoute(t *testing.T) {
	routes := DefaultRoutes()
	found := false
	for _, r := range routes {
		if r.ServiceName == "recommendation-service" {
			found = true
			if r.Streaming {
				t.Errorf("recommendation-service route %q should not have Streaming=true", r.PathPrefix)
			}
			if r.PathPrefix != "/api/recommendations" {
				t.Errorf("expected /api/recommendations prefix, got %q", r.PathPrefix)
			}
		}
	}
	if !found {
		t.Error("recommendation-service route not found")
	}
}

func TestDefaultRoutes_SubscriptionServiceRoutes(t *testing.T) {
	routes := DefaultRoutes()
	subRoutes := make([]Route, 0)
	for _, r := range routes {
		if r.ServiceName == "subscription-service" {
			subRoutes = append(subRoutes, r)
		}
	}

	if len(subRoutes) != 2 {
		t.Fatalf("expected 2 subscription-service routes, got %d", len(subRoutes))
	}

	prefixes := make(map[string]bool)
	for _, r := range subRoutes {
		prefixes[r.PathPrefix] = true
		if r.Streaming {
			t.Errorf("subscription-service route %q should not have Streaming=true", r.PathPrefix)
		}
	}

	if !prefixes["/api/plans"] {
		t.Error("expected /api/plans route for subscription-service")
	}
	if !prefixes["/api/subscriptions"] {
		t.Error("expected /api/subscriptions route for subscription-service")
	}
}

func TestDefaultRoutes_CatalogServiceNoStripPrefix(t *testing.T) {
	routes := DefaultRoutes()
	for _, r := range routes {
		if r.ServiceName == "catalog-service" {
			if r.StripPrefix {
				t.Errorf("catalog-service route %q should have StripPrefix=false", r.PathPrefix)
			}
		}
	}
}
