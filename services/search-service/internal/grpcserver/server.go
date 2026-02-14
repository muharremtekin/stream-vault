package grpcserver

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/streamvault/search-service/internal/handler"
	"github.com/streamvault/search-service/internal/model"
	"github.com/streamvault/search-service/internal/trending"
	commonv1 "github.com/streamvault/search-service/proto/common/v1"
	searchv1 "github.com/streamvault/search-service/proto/search/v1"
)

// SearchServer implements the SearchServiceServer gRPC interface.
type SearchServer struct {
	searchv1.UnimplementedSearchServiceServer
	searcher    handler.Searcher
	trendingSvc *trending.Service
}

// NewSearchServer creates a new SearchServer.
func NewSearchServer(searcher handler.Searcher, trendingSvc *trending.Service) *SearchServer {
	return &SearchServer{
		searcher:    searcher,
		trendingSvc: trendingSvc,
	}
}

// Search performs full-text search with filtering, sorting, and pagination.
func (s *SearchServer) Search(ctx context.Context, req *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	modelReq := protoToSearchRequest(req)

	resp, err := s.searcher.Search(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search failed: %v", err)
	}

	return searchResponseToProto(resp), nil
}

// Autocomplete returns title suggestions for prefix matching.
func (s *SearchServer) Autocomplete(ctx context.Context, req *searchv1.AutocompleteRequest) (*searchv1.AutocompleteResponse, error) {
	if len(req.GetQuery()) < 2 {
		return nil, status.Error(codes.InvalidArgument, "query must be at least 2 characters")
	}

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	modelReq := model.AutocompleteRequest{
		Query: req.GetQuery(),
		Limit: limit,
	}

	resp, err := s.searcher.Autocomplete(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "autocomplete failed: %v", err)
	}

	protoResp := &searchv1.AutocompleteResponse{
		Suggestions: make([]*searchv1.AutocompleteSuggestion, len(resp.Suggestions)),
	}
	for i, sug := range resp.Suggestions {
		protoResp.Suggestions[i] = &searchv1.AutocompleteSuggestion{
			ContentId:    sug.ContentID,
			Title:        sug.Title,
			ContentType:  contentTypeToProto(sug.ContentType),
			ThumbnailUrl: sug.ThumbnailURL,
			ReleaseYear:  int32(sug.ReleaseYear),
		}
	}

	return protoResp, nil
}

// GetTrending returns trending content for a given time window.
func (s *SearchServer) GetTrending(ctx context.Context, req *searchv1.GetTrendingRequest) (*searchv1.GetTrendingResponse, error) {
	window := timeWindowToString(req.GetTimeWindow())

	limit := int(req.GetLimit())
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	resp, err := s.trendingSvc.GetTrending(ctx, window, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "trending failed: %v", err)
	}

	protoResp := &searchv1.GetTrendingResponse{
		Items: make([]*searchv1.TrendingItem, len(resp.Items)),
	}
	for i, item := range resp.Items {
		protoResp.Items[i] = &searchv1.TrendingItem{
			ContentId:     item.ContentID,
			Title:         item.Title,
			ThumbnailUrl:  item.ThumbnailURL,
			ContentType:   contentTypeToProto(item.ContentType),
			ReleaseYear:   int32(item.ReleaseYear),
			AverageRating: item.AverageRating,
			Rank:          int32(item.Rank),
			RankChange:    int32(item.RankChange),
			ViewCount:     item.ViewCount,
		}
	}

	return protoResp, nil
}

func protoToSearchRequest(req *searchv1.SearchRequest) model.SearchRequest {
	modelReq := model.SearchRequest{
		Query:     req.GetQuery(),
		SortField: sortFieldToString(req.GetSortField()),
		SortOrder: sortOrderToString(req.GetSortOrder()),
		Page:      1,
		PageSize:  20,
	}

	if p := req.GetPagination(); p != nil {
		if p.Page > 0 {
			modelReq.Page = int(p.Page)
		}
		if p.PageSize > 0 && p.PageSize <= 100 {
			modelReq.PageSize = int(p.PageSize)
		}
	}

	if f := req.GetFilters(); f != nil {
		modelReq.Genres = f.GetGenres()
		modelReq.YearFrom = int(f.GetYearFrom())
		modelReq.YearTo = int(f.GetYearTo())
		modelReq.MaturityRatings = f.GetMaturityRatings()
		modelReq.MinRating = f.GetMinRating()
		if f.GetContentType() != commonv1.ContentType_CONTENT_TYPE_UNSPECIFIED {
			modelReq.ContentType = contentTypeFromProto(f.GetContentType())
		}
	}

	return modelReq
}

func searchResponseToProto(resp *model.SearchResponse) *searchv1.SearchResponse {
	protoResp := &searchv1.SearchResponse{
		Hits: make([]*searchv1.SearchHit, len(resp.Items)),
		Pagination: &commonv1.PaginatedResponse{
			Page:       int32(resp.Page),
			PageSize:   int32(resp.PageSize),
			TotalCount: resp.TotalCount,
			TotalPages: int32(resp.TotalPages),
		},
	}

	for i, hit := range resp.Items {
		protoHit := &searchv1.SearchHit{
			ContentId:     hit.ContentID,
			Title:         hit.Title,
			Description:   hit.Description,
			ContentType:   contentTypeToProto(hit.ContentType),
			ThumbnailUrl:  hit.ThumbnailURL,
			ReleaseYear:   int32(hit.ReleaseYear),
			AverageRating: hit.AverageRating,
		}

		if hit.Highlights != nil {
			protoHit.Highlights = &searchv1.HighlightFields{
				Title:       hit.Highlights.Title,
				Description: hit.Highlights.Description,
			}
		}

		protoResp.Hits[i] = protoHit
	}

	if resp.Facets != nil {
		protoResp.Facets = facetsToProto(resp.Facets)
	}

	return protoResp
}

func facetsToProto(f *model.Facets) *searchv1.Facets {
	return &searchv1.Facets{
		GenreFacets:       facetBucketsToProto(f.GenreFacets),
		YearFacets:        facetBucketsToProto(f.YearFacets),
		MaturityFacets:    facetBucketsToProto(f.MaturityFacets),
		ContentTypeFacets: facetBucketsToProto(f.ContentTypeFacets),
	}
}

func facetBucketsToProto(buckets []model.FacetBucket) []*searchv1.FacetBucket {
	result := make([]*searchv1.FacetBucket, len(buckets))
	for i, b := range buckets {
		result[i] = &searchv1.FacetBucket{
			Key:      b.Key,
			DocCount: b.DocCount,
		}
	}
	return result
}

func sortFieldToString(sf searchv1.SortField) string {
	switch sf {
	case searchv1.SortField_SORT_FIELD_RATING:
		return "rating"
	case searchv1.SortField_SORT_FIELD_YEAR:
		return "year"
	case searchv1.SortField_SORT_FIELD_TITLE:
		return "title"
	default:
		return "relevance"
	}
}

func sortOrderToString(so searchv1.SortOrder) string {
	switch so {
	case searchv1.SortOrder_SORT_ORDER_ASC:
		return "asc"
	default:
		return "desc"
	}
}

func timeWindowToString(tw searchv1.TimeWindow) string {
	switch tw {
	case searchv1.TimeWindow_TIME_WINDOW_DAY:
		return "day"
	case searchv1.TimeWindow_TIME_WINDOW_MONTH:
		return "month"
	default:
		return "week"
	}
}

func contentTypeToProto(ct string) commonv1.ContentType {
	switch ct {
	case "movie":
		return commonv1.ContentType_CONTENT_TYPE_MOVIE
	case "series":
		return commonv1.ContentType_CONTENT_TYPE_SERIES
	default:
		return commonv1.ContentType_CONTENT_TYPE_UNSPECIFIED
	}
}

func contentTypeFromProto(ct commonv1.ContentType) string {
	switch ct {
	case commonv1.ContentType_CONTENT_TYPE_MOVIE:
		return "movie"
	case commonv1.ContentType_CONTENT_TYPE_SERIES:
		return "series"
	default:
		return ""
	}
}
