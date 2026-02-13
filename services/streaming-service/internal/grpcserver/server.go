package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/streamvault/streaming-service/proto/common/v1"
	streamingv1 "github.com/streamvault/streaming-service/proto/streaming/v1"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/messaging"
	"github.com/streamvault/streaming-service/internal/progress"
	"github.com/streamvault/streaming-service/internal/storage"
)

type StreamingServer struct {
	streamingv1.UnimplementedStreamingServiceServer
	storage     storage.Storage
	progressSvc *progress.Service
	redisClient *redis.Client
	minioCfg    config.MinIOConfig
}

func NewStreamingServer(store storage.Storage, progressSvc *progress.Service, redisClient *redis.Client, minioCfg config.MinIOConfig) *StreamingServer {
	return &StreamingServer{
		storage:     store,
		progressSvc: progressSvc,
		redisClient: redisClient,
		minioCfg:    minioCfg,
	}
}

func (s *StreamingServer) GetStreamingInfo(ctx context.Context, req *streamingv1.GetStreamingInfoRequest) (*streamingv1.StreamingInfoResponse, error) {
	if req.ContentId == "" {
		return nil, status.Error(codes.InvalidArgument, "content_id is required")
	}

	// Check Redis cache
	key := fmt.Sprintf("stream-info:%s", req.ContentId)
	data, err := s.redisClient.Get(ctx, key).Result()
	if err == nil {
		var result messaging.EncodingResult
		if err := json.Unmarshal([]byte(data), &result); err == nil {
			resp := &streamingv1.StreamingInfoResponse{
				ContentId:       req.ContentId,
				Status:          commonv1.VideoStatus_VIDEO_STATUS_READY,
				DurationSeconds: result.Duration,
				ManifestPath:    fmt.Sprintf("/stream/%s/manifest.m3u8", req.ContentId),
				ThumbnailPath:   fmt.Sprintf("/thumbnails/%s/thumb.jpg", req.ContentId),
				PosterPath:      fmt.Sprintf("/thumbnails/%s/poster.jpg", req.ContentId),
			}

			for _, out := range result.Outputs {
				qi := &streamingv1.QualityInfo{
					Label:        out.Quality,
					Width:        int32(out.Width),
					Height:       int32(out.Height),
					BitrateKbps:  int32(out.BitrateKbps),
					SegmentCount: int32(out.SegmentCount),
					MinTier:      qualityToMinTier(out.Quality),
				}
				resp.AvailableQualities = append(resp.AvailableQualities, qi)
			}

			return resp, nil
		}
	}

	// Check if content exists in encoded bucket
	prefix := req.ContentId + "/"
	objects, err := s.storage.List(ctx, s.minioCfg.EncodedBucket, prefix)
	if err != nil || len(objects) == 0 {
		log.Debug().Str("content_id", req.ContentId).Msg("no encoded content found")
		return &streamingv1.StreamingInfoResponse{
			ContentId: req.ContentId,
			Status:    commonv1.VideoStatus_VIDEO_STATUS_NOT_UPLOADED,
		}, nil
	}

	return &streamingv1.StreamingInfoResponse{
		ContentId:    req.ContentId,
		Status:       commonv1.VideoStatus_VIDEO_STATUS_READY,
		ManifestPath: fmt.Sprintf("/stream/%s/manifest.m3u8", req.ContentId),
	}, nil
}

func (s *StreamingServer) GetProgress(ctx context.Context, req *streamingv1.GetProgressRequest) (*streamingv1.ProgressResponse, error) {
	if req.UserId == "" || req.ContentId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and content_id are required")
	}

	p, err := s.progressSvc.GetProgress(ctx, req.UserId, req.ContentId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get progress: %v", err)
	}

	if p == nil {
		return nil, status.Error(codes.NotFound, "no progress found")
	}

	return &streamingv1.ProgressResponse{
		UserId:          p.UserID,
		ContentId:       p.ContentID,
		PositionSeconds: p.PositionSeconds,
		DurationSeconds: p.DurationSeconds,
		Percentage:      p.Percentage,
		UpdatedAt:       timestamppb.New(time.Unix(p.UpdatedAt, 0)),
	}, nil
}

func (s *StreamingServer) GetContinueWatching(ctx context.Context, req *streamingv1.GetContinueWatchingRequest) (*streamingv1.ContinueWatchingResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	items, err := s.progressSvc.GetContinueWatching(ctx, req.UserId, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get continue watching: %v", err)
	}

	resp := &streamingv1.ContinueWatchingResponse{}
	for _, item := range items {
		resp.Items = append(resp.Items, &streamingv1.ContinueWatchingItem{
			ContentId:       item.ContentID,
			PositionSeconds: item.PositionSeconds,
			DurationSeconds: item.DurationSeconds,
			Percentage:      item.Percentage,
			UpdatedAt:       timestamppb.New(time.Unix(item.UpdatedAt, 0)),
		})
	}

	return resp, nil
}

func qualityToMinTier(quality string) commonv1.SubscriptionTier {
	switch quality {
	case "1080p":
		return commonv1.SubscriptionTier_SUBSCRIPTION_TIER_STANDARD
	case "4k":
		return commonv1.SubscriptionTier_SUBSCRIPTION_TIER_PREMIUM
	default:
		return commonv1.SubscriptionTier_SUBSCRIPTION_TIER_BASIC
	}
}

