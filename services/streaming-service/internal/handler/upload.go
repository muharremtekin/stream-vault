package handler

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/messaging"
	"github.com/streamvault/streaming-service/internal/storage"
)

type UploadHandler struct {
	storage   storage.Storage
	publisher messaging.Publisher
	uploadCfg config.UploadConfig
	minioCfg  config.MinIOConfig
}

func NewUploadHandler(store storage.Storage, pub messaging.Publisher, uploadCfg config.UploadConfig, minioCfg config.MinIOConfig) *UploadHandler {
	return &UploadHandler{
		storage:   store,
		publisher: pub,
		uploadCfg: uploadCfg,
		minioCfg:  minioCfg,
	}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")

	reader, err := r.MultipartReader()
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "multipart form data required")
		return
	}

	var contentID string
	var fileReader io.Reader
	var fileSize int64
	var fileContentType string
	var fileName string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			WriteErrorResponse(w, http.StatusBadRequest, "error reading multipart data")
			return
		}

		switch part.FormName() {
		case "content_id":
			data, _ := io.ReadAll(part)
			contentID = string(data)
		case "file":
			fileName = part.FileName()
			fileContentType = part.Header.Get("Content-Type")
			if fileContentType == "" {
				fileContentType = mime.TypeByExtension("." + fileExt(fileName))
			}
			// Buffer to temp file so NextPart() doesn't invalidate the reader.
			tmpFile, err := os.CreateTemp("", "upload-*.tmp")
			if err != nil {
				WriteErrorResponse(w, http.StatusInternalServerError, "failed to create temp file")
				return
			}
			defer os.Remove(tmpFile.Name())
			written, err := io.Copy(tmpFile, part)
			if err != nil {
				tmpFile.Close()
				WriteErrorResponse(w, http.StatusInternalServerError, "failed to buffer upload")
				return
			}
			if _, err := tmpFile.Seek(0, 0); err != nil {
				tmpFile.Close()
				WriteErrorResponse(w, http.StatusInternalServerError, "failed to seek temp file")
				return
			}
			fileReader = tmpFile
			fileSize = written
		}

		if contentID != "" && fileReader != nil {
			break
		}
	}

	// Close temp file after upload completes.
	if closer, ok := fileReader.(io.Closer); ok {
		defer closer.Close()
	}

	if contentID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "content_id is required")
		return
	}
	if fileReader == nil {
		WriteErrorResponse(w, http.StatusBadRequest, "file is required")
		return
	}

	if !h.isAllowedType(fileContentType) {
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("unsupported file type: %s", fileContentType))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	key := fmt.Sprintf("%s/original.mp4", contentID)

	if err := h.storage.Upload(ctx, h.minioCfg.RawBucket, key, fileReader, fileSize, fileContentType); err != nil {
		log.Error().Err(err).Str("content_id", contentID).Msg("failed to upload file to minio")
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to upload file")
		return
	}

	jobID := uuid.New().String()
	job := messaging.EncodingJob{
		JobID:        jobID,
		ContentID:    contentID,
		SourcePath:   key,
		SourceBucket: h.minioCfg.RawBucket,
		RequestedBy:  userID,
		CreatedAt:    time.Now().UTC(),
	}

	if err := h.publisher.PublishEncodingJob(ctx, job); err != nil {
		log.Error().Err(err).Str("job_id", jobID).Msg("failed to publish encoding job")
		WriteErrorResponse(w, http.StatusInternalServerError, "failed to queue encoding job")
		return
	}

	log.Info().
		Str("job_id", jobID).
		Str("content_id", contentID).
		Str("uploaded_by", userID).
		Msg("video uploaded and encoding job queued")

	WriteJSON(w, http.StatusAccepted, map[string]string{
		"job_id":     jobID,
		"content_id": contentID,
		"status":     "queued",
		"message":    "Upload successful, encoding job queued",
	})
}

func (h *UploadHandler) isAllowedType(contentType string) bool {
	for _, allowed := range h.uploadCfg.AllowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func fileExt(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i+1:]
		}
	}
	return ""
}
