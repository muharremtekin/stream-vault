package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/model"
)

const (
	searchCacheTTL       = 5 * time.Minute
	autocompleteCacheTTL = 5 * time.Minute
)

// Cache provides Redis-based caching for search results and trending data.
type Cache struct {
	client *redis.Client
}

// NewCache creates a new Cache.
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// GetSearchResult retrieves a cached search result.
func (c *Cache) GetSearchResult(ctx context.Context, req model.SearchRequest) (*model.SearchResponse, error) {
	key := fmt.Sprintf("search_cache:%s", hashRequest(req))
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var resp model.SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshalling cached search result: %w", err)
	}
	return &resp, nil
}

// SetSearchResult stores a search result in cache.
func (c *Cache) SetSearchResult(ctx context.Context, req model.SearchRequest, resp *model.SearchResponse) {
	key := fmt.Sprintf("search_cache:%s", hashRequest(req))
	data, err := json.Marshal(resp)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal search result for cache")
		return
	}
	if err := c.client.Set(ctx, key, data, searchCacheTTL).Err(); err != nil {
		log.Warn().Err(err).Msg("failed to cache search result")
	}
}

// GetAutocompleteResult retrieves a cached autocomplete result.
func (c *Cache) GetAutocompleteResult(ctx context.Context, req model.AutocompleteRequest) (*model.AutocompleteResponse, error) {
	key := fmt.Sprintf("autocomplete_cache:%s", hashRequest(req))
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var resp model.AutocompleteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshalling cached autocomplete result: %w", err)
	}
	return &resp, nil
}

// SetAutocompleteResult stores an autocomplete result in cache.
func (c *Cache) SetAutocompleteResult(ctx context.Context, req model.AutocompleteRequest, resp *model.AutocompleteResponse) {
	key := fmt.Sprintf("autocomplete_cache:%s", hashRequest(req))
	data, err := json.Marshal(resp)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal autocomplete result for cache")
		return
	}
	if err := c.client.Set(ctx, key, data, autocompleteCacheTTL).Err(); err != nil {
		log.Warn().Err(err).Msg("failed to cache autocomplete result")
	}
}

// GetTrending retrieves a cached trending response.
func (c *Cache) GetTrending(ctx context.Context, key string) (*model.TrendingResponse, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var resp model.TrendingResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshalling cached trending result: %w", err)
	}
	return &resp, nil
}

// SetTrending stores a trending response in cache.
func (c *Cache) SetTrending(ctx context.Context, key string, resp *model.TrendingResponse, ttl time.Duration) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal trending result for cache")
		return
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Warn().Err(err).Msg("failed to cache trending result")
	}
}

// TrackSearch increments the search frequency counter for a query term.
func (c *Cache) TrackSearch(ctx context.Context, query string) {
	if err := c.client.ZIncrBy(ctx, "popular_searches", 1, query).Err(); err != nil {
		log.Warn().Err(err).Str("query", query).Msg("failed to track search query")
	}
}

// hashRequest creates a deterministic SHA256 hash of a request struct for cache keying.
func hashRequest(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
