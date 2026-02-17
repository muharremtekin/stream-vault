package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/messaging"
	"github.com/streamvault/streaming-service/internal/storage"
)

// InfoHandler serves the HTTP streaming info endpoint consumed by the frontend VideoPlayer.
// It mirrors the gRPC GetStreamingInfo logic but returns camelCase JSON matching
// the frontend TypeScript StreamingInfo type.
type InfoHandler struct {
	storage     storage.Storage
	redisClient *redis.Client
	minioCfg    config.MinIOConfig
}

func NewInfoHandler(store storage.Storage, redisClient *redis.Client, minioCfg config.MinIOConfig) *InfoHandler {
	return &InfoHandler{
		storage:     store,
		redisClient: redisClient,
		minioCfg:    minioCfg,
	}
}

type streamingInfoResponse struct {
	VideoStatus        string                `json:"videoStatus"`
	DurationSeconds    int64                 `json:"durationSeconds"`
	AvailableQualities []qualityInfoResponse `json:"availableQualities"`
	ManifestUrl        string                `json:"manifestUrl,omitempty"`
	ThumbnailUrl       *string               `json:"thumbnailUrl"`
	PosterUrl          *string               `json:"posterUrl"`
	EncodedAt          *string               `json:"encodedAt"`
}

type qualityInfoResponse struct {
	Label        string `json:"label"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	BitrateKbps  int    `json:"bitrateKbps"`
	SegmentCount int    `json:"segmentCount"`
}

// GetStreamingInfo handles GET /api/stream/{contentId}/info.
// VideoStatus values returned: "Ready" | "Pending"
// ("Pending" covers not-uploaded, queued, and encoding states from the streaming service perspective)
func (h *InfoHandler) GetStreamingInfo(w http.ResponseWriter, r *http.Request) {
	contentID := r.PathValue("contentId")
	if contentID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "content_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Primary: check Redis cache populated by encoding result consumer
	key := fmt.Sprintf("stream-info:%s", contentID)
	data, err := h.redisClient.Get(ctx, key).Result()
	if err == nil {
		var result messaging.EncodingResult
		if jsonErr := json.Unmarshal([]byte(data), &result); jsonErr == nil {
			h.writeReadyResponse(w, contentID, &result)
			return
		}
	}

	// Fallback: check if encoded content exists in MinIO
	prefix := contentID + "/"
	objects, listErr := h.storage.List(ctx, h.minioCfg.EncodedBucket, prefix)
	if listErr != nil {
		log.Error().Err(listErr).Str("content_id", contentID).Msg("failed to list encoded content")
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to check content availability")
		return
	}

	if len(objects) == 0 {
		WriteJSON(w, http.StatusOK, streamingInfoResponse{
			VideoStatus:        "Pending",
			AvailableQualities: []qualityInfoResponse{},
		})
		return
	}

	// Content exists in MinIO but Redis cache is gone — return Ready with no quality details
	h.writeReadyResponse(w, contentID, nil)
}

func (h *InfoHandler) writeReadyResponse(w http.ResponseWriter, contentID string, result *messaging.EncodingResult) {
	manifestUrl := fmt.Sprintf("/stream/%s/manifest.m3u8", contentID)
	thumbnailUrl := fmt.Sprintf("/thumbnails/%s/thumb.jpg", contentID)
	posterUrl := fmt.Sprintf("/thumbnails/%s/poster.jpg", contentID)

	resp := streamingInfoResponse{
		VideoStatus:        "Ready",
		ManifestUrl:        manifestUrl,
		ThumbnailUrl:       &thumbnailUrl,
		PosterUrl:          &posterUrl,
		AvailableQualities: []qualityInfoResponse{},
	}

	if result != nil {
		resp.DurationSeconds = result.Duration
		resp.EncodedAt = &result.CompletedAt
		for _, out := range result.Outputs {
			resp.AvailableQualities = append(resp.AvailableQualities, qualityInfoResponse{
				Label:        out.Quality,
				Width:        out.Width,
				Height:       out.Height,
				BitrateKbps:  out.BitrateKbps,
				SegmentCount: out.SegmentCount,
			})
		}
	}

	WriteJSON(w, http.StatusOK, resp)
}
