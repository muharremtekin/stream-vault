package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/metrics"
	"github.com/streamvault/streaming-service/internal/storage"
)

type SegmentHandler struct {
	storage  storage.Storage
	minioCfg config.MinIOConfig
}

func NewSegmentHandler(store storage.Storage, minioCfg config.MinIOConfig) *SegmentHandler {
	return &SegmentHandler{
		storage:  store,
		minioCfg: minioCfg,
	}
}

func (h *SegmentHandler) ServeSegment(w http.ResponseWriter, r *http.Request) {
	contentId := r.PathValue("contentId")
	quality := r.PathValue("quality")
	segment := r.PathValue("segment")

	if contentId == "" || quality == "" || segment == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "content_id, quality, and segment are required")
		return
	}

	start := time.Now()
	metrics.StreamingActiveViewers.Inc()
	defer metrics.StreamingActiveViewers.Dec()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	key := fmt.Sprintf("%s/%s/%s", contentId, quality, segment)
	reader, info, err := h.storage.Download(ctx, h.minioCfg.EncodedBucket, key)
	if err != nil {
		if storage.IsNotFound(err) {
			log.Debug().Str("key", key).Msg("segment not found")
			WriteErrorResponse(w, http.StatusNotFound, "segment not found")
		} else {
			log.Error().Err(err).Str("key", key).Msg("failed to download segment")
			WriteErrorResponse(w, http.StatusBadGateway, "storage error")
		}
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	n, err := io.Copy(w, reader)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("error streaming segment to client")
	}

	metrics.StreamingSegmentServeDuration.WithLabelValues(quality).Observe(time.Since(start).Seconds())
	metrics.StreamingBandwidthBytesTotal.WithLabelValues(quality).Add(float64(n))
}
