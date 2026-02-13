package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/hls"
	"github.com/streamvault/streaming-service/internal/storage"
)

type ManifestHandler struct {
	storage  storage.Storage
	minioCfg config.MinIOConfig
}

func NewManifestHandler(store storage.Storage, minioCfg config.MinIOConfig) *ManifestHandler {
	return &ManifestHandler{
		storage:  store,
		minioCfg: minioCfg,
	}
}

func (h *ManifestHandler) MasterPlaylist(w http.ResponseWriter, r *http.Request) {
	contentId := r.PathValue("contentId")
	if contentId == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "content_id is required")
		return
	}

	role := r.Header.Get("X-User-Role")
	if role == "" || role == "Admin" {
		role = "Premium" // Admin gets highest tier, empty defaults to Premium for quality filtering
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Find available qualities by listing encoded bucket
	prefix := contentId + "/"
	objects, err := h.storage.List(ctx, h.minioCfg.EncodedBucket, prefix)
	if err != nil {
		log.Error().Err(err).Str("content_id", contentId).Msg("failed to list encoded content")
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to retrieve content")
		return
	}

	// Determine which quality directories exist
	availableDirs := make(map[string]bool)
	for _, obj := range objects {
		parts := strings.Split(strings.TrimPrefix(obj.Key, prefix), "/")
		if len(parts) > 0 {
			availableDirs[parts[0]] = true
		}
	}

	var available []hls.QualityProfile
	for _, p := range hls.AllProfiles {
		if availableDirs[p.Label] {
			available = append(available, p)
		}
	}

	if len(available) == 0 {
		WriteErrorResponse(w, http.StatusNotFound, "no encoded content available")
		return
	}

	// Filter by subscription tier
	filtered := hls.FilterByTier(role, available)
	if len(filtered) == 0 {
		WriteErrorResponse(w, http.StatusForbidden, "no qualities available for your subscription tier")
		return
	}

	playlist := hls.GenerateMasterPlaylist(contentId, filtered)

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(playlist))
}

func (h *ManifestHandler) MediaPlaylist(w http.ResponseWriter, r *http.Request) {
	contentId := r.PathValue("contentId")
	quality := r.PathValue("quality")

	if contentId == "" || quality == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "content_id and quality are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	key := contentId + "/" + quality + "/playlist.m3u8"
	reader, info, err := h.storage.Download(ctx, h.minioCfg.EncodedBucket, key)
	if err != nil {
		if storage.IsNotFound(err) {
			log.Debug().Str("key", key).Msg("playlist not found")
			WriteErrorResponse(w, http.StatusNotFound, "playlist not found")
		} else {
			log.Error().Err(err).Str("key", key).Msg("failed to download media playlist")
			WriteErrorResponse(w, http.StatusBadGateway, "storage error")
		}
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Content-Length", formatInt64(info.Size))
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
}

func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
