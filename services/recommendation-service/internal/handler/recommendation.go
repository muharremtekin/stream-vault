package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/recommendation-service/internal/engine"
	"github.com/streamvault/recommendation-service/internal/metrics"
)

// RecommendationHandler handles recommendation HTTP endpoints.
type RecommendationHandler struct {
	recommender engine.Recommender
}

// NewRecommendationHandler creates a new RecommendationHandler.
func NewRecommendationHandler(recommender engine.Recommender) *RecommendationHandler {
	return &RecommendationHandler{recommender: recommender}
}

// ServeRecommendations handles GET /api/recommendations.
func (h *RecommendationHandler) ServeRecommendations(w http.ResponseWriter, r *http.Request) {
	metrics.RecommendationRequestsTotal.WithLabelValues("personal").Inc()

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_USER_ID", "X-User-Id header is required")
		return
	}

	limit := parseLimit(r, 20, 100)

	items, err := h.recommender.GetRecommendations(r.Context(), userID, limit)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to get recommendations")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "RECOMMENDATION_ERROR", "failed to generate recommendations")
		return
	}

	if items == nil {
		items = []engine.RecommendedItem{}
	}

	WriteJSON(w, http.StatusOK, recommendationsResponse{Items: items})
}

// ServeSimilar handles GET /api/recommendations/similar/{contentId}.
func (h *RecommendationHandler) ServeSimilar(w http.ResponseWriter, r *http.Request) {
	metrics.RecommendationRequestsTotal.WithLabelValues("similar").Inc()

	contentID := r.PathValue("contentId")
	if contentID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_CONTENT_ID", "contentId path parameter is required")
		return
	}

	limit := parseLimit(r, 10, 50)

	items, err := h.recommender.GetSimilar(r.Context(), contentID, limit)
	if err != nil {
		log.Error().Err(err).Str("content_id", contentID).Msg("failed to get similar content")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "SIMILAR_ERROR", "failed to get similar content")
		return
	}

	if items == nil {
		items = []engine.SimilarItem{}
	}

	WriteJSON(w, http.StatusOK, similarResponse{Items: items})
}

// ServeHomePage handles GET /api/recommendations/home.
func (h *RecommendationHandler) ServeHomePage(w http.ResponseWriter, r *http.Request) {
	metrics.RecommendationRequestsTotal.WithLabelValues("home").Inc()

	userID := r.Header.Get("X-User-Id")
	// userID is optional for home page — anonymous users get trending + new

	sections, err := h.recommender.GetHomePageSections(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to get home page sections")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "HOME_PAGE_ERROR", "failed to generate home page")
		return
	}

	if sections == nil {
		sections = []engine.HomePageSection{}
	}

	WriteJSON(w, http.StatusOK, homePageResponse{Sections: sections})
}

// ServeFeedback handles POST /api/recommendations/feedback.
func (h *RecommendationHandler) ServeFeedback(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		WriteErrorResponse(w, r, http.StatusBadRequest, "MISSING_USER_ID", "X-User-Id header is required")
		return
	}

	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, r, http.StatusBadRequest, "INVALID_BODY", "invalid JSON request body")
		return
	}

	if req.ContentID == "" {
		WriteValidationError(w, r, []FieldError{{Field: "contentId", Message: "contentId is required"}})
		return
	}
	if req.FeedbackType == "" {
		req.FeedbackType = "not_interested"
	}

	if err := h.recommender.RecordFeedback(r.Context(), userID, req.ContentID, req.FeedbackType); err != nil {
		log.Error().Err(err).Str("user_id", userID).Str("content_id", req.ContentID).Msg("failed to record feedback")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "FEEDBACK_ERROR", "failed to record feedback")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Response types ---

type recommendationsResponse struct {
	Items []engine.RecommendedItem `json:"items"`
}

type similarResponse struct {
	Items []engine.SimilarItem `json:"items"`
}

type homePageResponse struct {
	Sections []engine.HomePageSection `json:"sections"`
}

type feedbackRequest struct {
	ContentID    string `json:"contentId"`
	FeedbackType string `json:"feedbackType"`
}

// parseLimit extracts and clamps the limit query parameter.
func parseLimit(r *http.Request, defaultVal, maxVal int) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return defaultVal
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return defaultVal
	}
	if limit > maxVal {
		return maxVal
	}
	return limit
}
