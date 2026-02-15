package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/streamvault/search-service/internal/model"
)

var esSearchTracer = otel.Tracer("search-service/elasticsearch")

// Searcher builds and executes search queries against Elasticsearch.
type Searcher struct {
	client *Client
}

// NewSearcher creates a new Searcher.
func NewSearcher(client *Client) *Searcher {
	return &Searcher{client: client}
}

// Search performs a full-text search with filters, highlights, facets, and pagination.
func (s *Searcher) Search(ctx context.Context, req model.SearchRequest) (*model.SearchResponse, error) {
	ctx, span := esSearchTracer.Start(ctx, "elasticsearch.search",
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.operation", "search"),
			attribute.String("db.elasticsearch.index", s.client.indexName),
			attribute.String("search.query", req.Query),
		),
	)
	defer span.End()

	query := s.buildSearchQuery(req)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encoding search query: %w", err)
	}

	log.Debug().RawJSON("query", buf.Bytes()).Msg("executing search")

	res, err := s.client.es.Search(
		s.client.es.Search.WithContext(ctx),
		s.client.es.Search.WithIndex(s.client.indexName),
		s.client.es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("executing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search returned status %d: %s", res.StatusCode, string(body))
	}

	return s.parseSearchResponse(res.Body, req.Page, req.PageSize)
}

func (s *Searcher) buildSearchQuery(req model.SearchRequest) map[string]interface{} {
	boolQuery := map[string]interface{}{}

	// Must clause: full-text query
	if req.Query != "" {
		boolQuery["must"] = []interface{}{
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":     req.Query,
					"fields":    []string{"title^3", "title.english^2", "original_title^2", "description", "description.english", "cast_names", "director"},
					"type":      "best_fields",
					"fuzziness": "AUTO",
				},
			},
		}
	} else {
		boolQuery["must"] = []interface{}{
			map[string]interface{}{"match_all": map[string]interface{}{}},
		}
	}

	// Filter clauses
	filters := s.buildFilters(req)
	if len(filters) > 0 {
		boolQuery["filter"] = filters
	}

	// Build sort
	sort := s.buildSort(req)

	// Build full query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"from": req.From(),
		"size": req.PageSize,
		"sort": sort,
		"aggs": s.buildAggregations(),
	}

	// Add highlight if there's a text query
	if req.Query != "" {
		query["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"title":       map[string]interface{}{},
				"description": map[string]interface{}{"fragment_size": 150, "number_of_fragments": 2},
			},
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
		}
	}

	return query
}

func (s *Searcher) buildFilters(req model.SearchRequest) []interface{} {
	var filters []interface{}

	if len(req.Genres) > 0 {
		filters = append(filters, map[string]interface{}{
			"terms": map[string]interface{}{"genres": req.Genres},
		})
	}

	if req.YearFrom > 0 || req.YearTo > 0 {
		rangeFilter := map[string]interface{}{}
		if req.YearFrom > 0 {
			rangeFilter["gte"] = req.YearFrom
		}
		if req.YearTo > 0 {
			rangeFilter["lte"] = req.YearTo
		}
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{"release_year": rangeFilter},
		})
	}

	if len(req.MaturityRatings) > 0 {
		filters = append(filters, map[string]interface{}{
			"terms": map[string]interface{}{"maturity_rating": req.MaturityRatings},
		})
	}

	if req.MinRating > 0 {
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"average_rating": map[string]interface{}{"gte": req.MinRating},
			},
		})
	}

	if req.ContentType != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"content_type": req.ContentType},
		})
	}

	return filters
}

func (s *Searcher) buildSort(req model.SearchRequest) []interface{} {
	order := "desc"
	if strings.EqualFold(req.SortOrder, "asc") {
		order = "asc"
	}

	switch strings.ToLower(req.SortField) {
	case "rating":
		return []interface{}{
			map[string]interface{}{"average_rating": map[string]interface{}{"order": order}},
			"_score",
		}
	case "year":
		return []interface{}{
			map[string]interface{}{"release_year": map[string]interface{}{"order": order}},
			"_score",
		}
	case "title":
		return []interface{}{
			map[string]interface{}{"title.keyword": map[string]interface{}{"order": order}},
		}
	default:
		return []interface{}{"_score"}
	}
}

func (s *Searcher) buildAggregations() map[string]interface{} {
	return map[string]interface{}{
		"genres": map[string]interface{}{
			"terms": map[string]interface{}{"field": "genres", "size": 30},
		},
		"release_years": map[string]interface{}{
			"terms": map[string]interface{}{"field": "release_year", "size": 50, "order": map[string]interface{}{"_key": "desc"}},
		},
		"maturity_ratings": map[string]interface{}{
			"terms": map[string]interface{}{"field": "maturity_rating", "size": 10},
		},
		"content_types": map[string]interface{}{
			"terms": map[string]interface{}{"field": "content_type", "size": 10},
		},
	}
}

func (s *Searcher) parseSearchResponse(body io.Reader, page, pageSize int) (*model.SearchResponse, error) {
	var esResp esSearchResponse
	if err := json.NewDecoder(body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("parsing search response: %w", err)
	}

	hits := make([]model.SearchHit, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		searchHit := model.SearchHit{
			ContentID:      hit.Source.ID,
			Title:          hit.Source.Title,
			Description:    hit.Source.Description,
			ContentType:    hit.Source.ContentType,
			ThumbnailURL:   hit.Source.ThumbnailURL,
			ReleaseYear:    hit.Source.ReleaseYear,
			AverageRating:  hit.Source.AverageRating,
			MaturityRating: hit.Source.MaturityRating,
			Genres:         hit.Source.Genres,
		}

		if hit.Highlight != nil {
			hl := &model.Highlight{}
			if titles, ok := hit.Highlight["title"]; ok && len(titles) > 0 {
				hl.Title = titles[0]
			}
			if descs, ok := hit.Highlight["description"]; ok && len(descs) > 0 {
				hl.Description = descs[0]
			}
			if hl.Title != "" || hl.Description != "" {
				searchHit.Highlights = hl
			}
		}

		hits = append(hits, searchHit)
	}

	facets := s.parseAggregations(esResp.Aggregations)

	resp := model.NewSearchResponse(hits, facets, page, pageSize, esResp.Hits.Total.Value)
	return &resp, nil
}

func (s *Searcher) parseAggregations(aggs map[string]esAggregation) *model.Facets {
	if aggs == nil {
		return nil
	}

	parseBuckets := func(agg esAggregation) []model.FacetBucket {
		buckets := make([]model.FacetBucket, 0, len(agg.Buckets))
		for _, b := range agg.Buckets {
			buckets = append(buckets, model.FacetBucket{
				Key:      fmt.Sprintf("%v", b.Key),
				DocCount: b.DocCount,
			})
		}
		return buckets
	}

	return &model.Facets{
		GenreFacets:       parseBuckets(aggs["genres"]),
		YearFacets:        parseBuckets(aggs["release_years"]),
		MaturityFacets:    parseBuckets(aggs["maturity_ratings"]),
		ContentTypeFacets: parseBuckets(aggs["content_types"]),
	}
}

// Autocomplete returns title suggestions matching the given prefix.
func (s *Searcher) Autocomplete(ctx context.Context, req model.AutocompleteRequest) (*model.AutocompleteResponse, error) {
	ctx, span := esSearchTracer.Start(ctx, "elasticsearch.autocomplete",
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.operation", "search"),
			attribute.String("db.elasticsearch.index", s.client.indexName),
			attribute.String("search.query", req.Query),
		),
	)
	defer span.End()

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"title.autocomplete", "original_title.autocomplete"},
				"type":   "best_fields",
			},
		},
		"size": req.Limit,
		"_source": []string{
			"id", "title", "content_type", "thumbnail_url", "release_year",
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("encoding autocomplete query: %w", err)
	}

	res, err := s.client.es.Search(
		s.client.es.Search.WithContext(ctx),
		s.client.es.Search.WithIndex(s.client.indexName),
		s.client.es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("executing autocomplete: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("autocomplete returned status %d: %s", res.StatusCode, string(body))
	}

	var esResp esSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("parsing autocomplete response: %w", err)
	}

	suggestions := make([]model.AutocompleteSuggestion, 0, len(esResp.Hits.Hits))
	for _, hit := range esResp.Hits.Hits {
		suggestions = append(suggestions, model.AutocompleteSuggestion{
			ContentID:    hit.Source.ID,
			Title:        hit.Source.Title,
			ContentType:  hit.Source.ContentType,
			ThumbnailURL: hit.Source.ThumbnailURL,
			ReleaseYear:  hit.Source.ReleaseYear,
		})
	}

	return &model.AutocompleteResponse{Suggestions: suggestions}, nil
}

// --- Elasticsearch response types ---

type esSearchResponse struct {
	Hits         esHits                     `json:"hits"`
	Aggregations map[string]esAggregation   `json:"aggregations"`
}

type esHits struct {
	Total esTotal `json:"total"`
	Hits  []esHit `json:"hits"`
}

type esTotal struct {
	Value    int64  `json:"value"`
	Relation string `json:"relation"`
}

type esHit struct {
	Source    model.SearchDocument   `json:"_source"`
	Score     float64                `json:"_score"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
}

type esAggregation struct {
	Buckets []esBucket `json:"buckets"`
}

type esBucket struct {
	Key      interface{} `json:"key"`
	DocCount int64       `json:"doc_count"`
}
