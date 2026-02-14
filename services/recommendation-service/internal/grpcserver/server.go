package grpcserver

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/streamvault/recommendation-service/internal/engine"
	commonv1 "github.com/streamvault/recommendation-service/proto/common/v1"
	recv1 "github.com/streamvault/recommendation-service/proto/recommendation/v1"
)

// RecommendationServer implements the RecommendationServiceServer gRPC interface.
type RecommendationServer struct {
	recv1.UnimplementedRecommendationServiceServer
	recommender engine.Recommender
}

// NewRecommendationServer creates a new RecommendationServer.
func NewRecommendationServer(recommender engine.Recommender) *RecommendationServer {
	return &RecommendationServer{recommender: recommender}
}

// GetRecommendations implements RecommendationServiceServer.
func (s *RecommendationServer) GetRecommendations(
	ctx context.Context, req *recv1.GetRecommendationsRequest,
) (*recv1.GetRecommendationsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}

	items, err := s.recommender.GetRecommendations(ctx, req.GetUserId(), limit)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Msg("grpc: failed to get recommendations")
		return nil, status.Errorf(codes.Internal, "failed to get recommendations: %v", err)
	}

	protoItems := make([]*recv1.RecommendedItem, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, recommendedItemToProto(item))
	}

	return &recv1.GetRecommendationsResponse{Items: protoItems}, nil
}

// GetSimilar implements RecommendationServiceServer.
func (s *RecommendationServer) GetSimilar(
	ctx context.Context, req *recv1.GetSimilarRequest,
) (*recv1.GetSimilarResponse, error) {
	if req.GetContentId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "content_id is required")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 10
	}

	items, err := s.recommender.GetSimilar(ctx, req.GetContentId(), limit)
	if err != nil {
		log.Error().Err(err).Str("content_id", req.GetContentId()).Msg("grpc: failed to get similar content")
		return nil, status.Errorf(codes.Internal, "failed to get similar content: %v", err)
	}

	protoItems := make([]*recv1.SimilarItem, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, similarItemToProto(item))
	}

	return &recv1.GetSimilarResponse{Items: protoItems}, nil
}

// GetHomePageSections implements RecommendationServiceServer.
func (s *RecommendationServer) GetHomePageSections(
	ctx context.Context, req *recv1.GetHomePageSectionsRequest,
) (*recv1.GetHomePageSectionsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	sections, err := s.recommender.GetHomePageSections(ctx, req.GetUserId())
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Msg("grpc: failed to get home page sections")
		return nil, status.Errorf(codes.Internal, "failed to get home page sections: %v", err)
	}

	protoSections := make([]*recv1.HomePageSection, 0, len(sections))
	for _, section := range sections {
		protoItems := make([]*recv1.RecommendedItem, 0, len(section.Items))
		for _, item := range section.Items {
			protoItems = append(protoItems, recommendedItemToProto(item))
		}
		protoSections = append(protoSections, &recv1.HomePageSection{
			SectionType: sectionTypeToProto(section.SectionType),
			Title:       section.Title,
			Items:       protoItems,
		})
	}

	return &recv1.GetHomePageSectionsResponse{Sections: protoSections}, nil
}

// --- Conversion helpers ---

func recommendedItemToProto(item engine.RecommendedItem) *recv1.RecommendedItem {
	return &recv1.RecommendedItem{
		ContentId:     item.ContentID,
		Title:         item.Title,
		ThumbnailUrl:  item.ThumbnailURL,
		ContentType:   contentTypeToProto(item.ContentType),
		ReleaseYear:   int32(item.ReleaseYear),
		AverageRating: item.AverageRating,
		Score:         item.Score,
		Algorithm:     item.Algorithm,
		Reason:        item.Reason,
	}
}

func similarItemToProto(item engine.SimilarItem) *recv1.SimilarItem {
	return &recv1.SimilarItem{
		ContentId:       item.ContentID,
		Title:           item.Title,
		ThumbnailUrl:    item.ThumbnailURL,
		ContentType:     contentTypeToProto(item.ContentType),
		ReleaseYear:     int32(item.ReleaseYear),
		AverageRating:   item.AverageRating,
		SimilarityScore: item.SimilarityScore,
	}
}

func sectionTypeToProto(st string) recv1.SectionType {
	switch st {
	case "personal":
		return recv1.SectionType_SECTION_TYPE_PERSONAL
	case "trending":
		return recv1.SectionType_SECTION_TYPE_TRENDING
	case "because_you_watched":
		return recv1.SectionType_SECTION_TYPE_BECAUSE_YOU_WATCHED
	case "genre":
		return recv1.SectionType_SECTION_TYPE_GENRE
	case "new":
		return recv1.SectionType_SECTION_TYPE_NEW
	default:
		return recv1.SectionType_SECTION_TYPE_UNSPECIFIED
	}
}

func contentTypeToProto(ct string) commonv1.ContentType {
	switch ct {
	case "movie":
		return commonv1.ContentType_CONTENT_TYPE_MOVIE
	case "series":
		return commonv1.ContentType_CONTENT_TYPE_SERIES
	case "episode":
		return commonv1.ContentType_CONTENT_TYPE_EPISODE
	case "documentary":
		return commonv1.ContentType_CONTENT_TYPE_DOCUMENTARY
	default:
		return commonv1.ContentType_CONTENT_TYPE_UNSPECIFIED
	}
}
