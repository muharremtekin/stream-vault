package cache

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/elasticsearch"
	"github.com/streamvault/search-service/internal/metrics"
	"github.com/streamvault/search-service/internal/model"
)

// CachedSearcher wraps an Elasticsearch Searcher with Redis caching.
// It satisfies the handler.Searcher interface.
type CachedSearcher struct {
	searcher *elasticsearch.Searcher
	cache    *Cache
}

// NewCachedSearcher creates a new CachedSearcher.
func NewCachedSearcher(searcher *elasticsearch.Searcher, cache *Cache) *CachedSearcher {
	return &CachedSearcher{
		searcher: searcher,
		cache:    cache,
	}
}

// Search tries the cache first, then falls back to Elasticsearch.
func (cs *CachedSearcher) Search(ctx context.Context, req model.SearchRequest) (*model.SearchResponse, error) {
	if resp, err := cs.cache.GetSearchResult(ctx, req); err == nil {
		metrics.SearchCacheHitsTotal.Inc()
		log.Debug().Str("query", req.Query).Msg("search cache hit")
		return resp, nil
	}
	metrics.SearchCacheMissesTotal.Inc()

	resp, err := cs.searcher.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	cs.cache.SetSearchResult(ctx, req, resp)

	if req.Query != "" {
		cs.cache.TrackSearch(ctx, req.Query)
	}

	return resp, nil
}

// Autocomplete tries the cache first, then falls back to Elasticsearch.
func (cs *CachedSearcher) Autocomplete(ctx context.Context, req model.AutocompleteRequest) (*model.AutocompleteResponse, error) {
	if resp, err := cs.cache.GetAutocompleteResult(ctx, req); err == nil {
		metrics.SearchCacheHitsTotal.Inc()
		log.Debug().Str("query", req.Query).Msg("autocomplete cache hit")
		return resp, nil
	}
	metrics.SearchCacheMissesTotal.Inc()

	resp, err := cs.searcher.Autocomplete(ctx, req)
	if err != nil {
		return nil, err
	}

	cs.cache.SetAutocompleteResult(ctx, req, resp)
	return resp, nil
}
