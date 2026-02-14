package handler

import (
	"net/http"

	"github.com/streamvault/recommendation-service/internal/engine"
)

// RegisterRecommendationRoutes registers all recommendation HTTP routes on the given mux.
func RegisterRecommendationRoutes(mux *http.ServeMux, recommender engine.Recommender) {
	h := NewRecommendationHandler(recommender)

	mux.HandleFunc("GET /api/recommendations", h.ServeRecommendations)
	mux.HandleFunc("GET /api/recommendations/similar/{contentId}", h.ServeSimilar)
	mux.HandleFunc("GET /api/recommendations/home", h.ServeHomePage)
	mux.HandleFunc("POST /api/recommendations/feedback", h.ServeFeedback)
}
