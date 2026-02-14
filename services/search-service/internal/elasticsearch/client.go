package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/config"
)

// Client wraps the Elasticsearch client with index-aware methods.
type Client struct {
	es        *elasticsearch.Client
	indexName string
}

// NewClient creates a new Elasticsearch client.
func NewClient(cfg config.ElasticsearchConfig) (*Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.URL},
	})
	if err != nil {
		return nil, fmt.Errorf("creating elasticsearch client: %w", err)
	}

	return &Client{
		es:        es,
		indexName: cfg.IndexName,
	}, nil
}

// ES returns the underlying elasticsearch client for advanced use.
func (c *Client) ES() *elasticsearch.Client {
	return c.es
}

// IndexName returns the configured index name.
func (c *Client) IndexName() string {
	return c.indexName
}

// EnsureIndex checks if the index exists and logs a warning if it does not.
// The index is expected to be created by init-index.sh at infrastructure startup.
func (c *Client) EnsureIndex(ctx context.Context) error {
	res, err := c.es.Indices.Exists([]string{c.indexName},
		c.es.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("checking index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		log.Warn().Str("index", c.indexName).
			Msg("index does not exist — expected to be created by elasticsearch-init container")
		return nil
	}

	if res.IsError() {
		return fmt.Errorf("checking index existence: status %d", res.StatusCode)
	}

	log.Info().Str("index", c.indexName).Msg("elasticsearch index verified")
	return nil
}

// HealthCheck verifies the Elasticsearch cluster is reachable and healthy.
func (c *Client) HealthCheck(ctx context.Context) error {
	res, err := c.es.Cluster.Health(
		c.es.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elasticsearch health check: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch health check returned status %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading health response: %w", err)
	}

	var health struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &health); err != nil {
		return fmt.Errorf("parsing health response: %w", err)
	}

	if health.Status == "red" {
		return fmt.Errorf("elasticsearch cluster status is red")
	}

	return nil
}
