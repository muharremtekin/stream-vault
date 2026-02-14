package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/cache"
	"github.com/streamvault/search-service/internal/model"
	"github.com/streamvault/search-service/internal/trending"
)

var validWindows = map[string]bool{"day": true, "week": true, "month": true}

// TrendingHandler handles the trending content endpoint.
type TrendingHandler struct {
	trendingSvc *trending.Service
	cache       *cache.Cache
}

// NewTrendingHandler creates a new TrendingHandler.
func NewTrendingHandler(trendingSvc *trending.Service, cache *cache.Cache) *TrendingHandler {
	return &TrendingHandler{
		trendingSvc: trendingSvc,
		cache:       cache,
	}
}

// ServeHTTP handles GET /api/search/trending requests.
func (h *TrendingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := model.ParseTrendingRequest(r)

	if !validWindows[req.TimeWindow] {
		WriteErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "window must be day, week, or month")
		return
	}

	cacheKey := fmt.Sprintf("trending_cache:%s", req.TimeWindow)
	if cached, err := h.cache.GetTrending(r.Context(), cacheKey); err == nil {
		WriteJSON(w, http.StatusOK, cached)
		return
	}

	resp, err := h.trendingSvc.GetTrending(r.Context(), req.TimeWindow, req.Limit)
	if err != nil {
		log.Error().Err(err).Str("window", req.TimeWindow).Msg("trending failed")
		WriteErrorResponse(w, r, http.StatusInternalServerError, "TRENDING_ERROR", "trending request failed")
		return
	}

	h.cache.SetTrending(r.Context(), cacheKey, resp, 1*time.Hour)

	WriteJSON(w, http.StatusOK, resp)
}
