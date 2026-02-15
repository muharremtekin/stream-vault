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

var esIndexTracer = otel.Tracer("search-service/elasticsearch")

// Indexer handles document indexing operations against Elasticsearch.
type Indexer struct {
	client *Client
}

// NewIndexer creates a new Indexer.
func NewIndexer(client *Client) *Indexer {
	return &Indexer{client: client}
}

// IndexDocument indexes or updates a single document.
func (ix *Indexer) IndexDocument(ctx context.Context, doc model.SearchDocument) error {
	ctx, span := esIndexTracer.Start(ctx, "elasticsearch.index",
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.operation", "index"),
			attribute.String("db.elasticsearch.index", ix.client.indexName),
			attribute.String("db.elasticsearch.doc_id", doc.ID),
		),
	)
	defer span.End()

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshalling document: %w", err)
	}

	res, err := ix.client.es.Index(
		ix.client.indexName,
		bytes.NewReader(body),
		ix.client.es.Index.WithContext(ctx),
		ix.client.es.Index.WithDocumentID(doc.ID),
	)
	if err != nil {
		return fmt.Errorf("indexing document %s: %w", doc.ID, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("indexing document %s: status %d, body: %s", doc.ID, res.StatusCode, string(respBody))
	}

	log.Debug().Str("id", doc.ID).Str("title", doc.Title).Msg("document indexed")
	return nil
}

// BulkIndex indexes multiple documents in a single bulk request.
func (ix *Indexer) BulkIndex(ctx context.Context, docs []model.SearchDocument) error {
	ctx, span := esIndexTracer.Start(ctx, "elasticsearch.bulk_index",
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.operation", "bulk"),
			attribute.String("db.elasticsearch.index", ix.client.indexName),
			attribute.Int("db.elasticsearch.doc_count", len(docs)),
		),
	)
	defer span.End()

	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, doc := range docs {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": ix.client.indexName,
				"_id":    doc.ID,
			},
		}
		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return fmt.Errorf("encoding bulk meta for %s: %w", doc.ID, err)
		}
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return fmt.Errorf("encoding bulk body for %s: %w", doc.ID, err)
		}
	}

	res, err := ix.client.es.Bulk(
		bytes.NewReader(buf.Bytes()),
		ix.client.es.Bulk.WithContext(ctx),
		ix.client.es.Bulk.WithIndex(ix.client.indexName),
	)
	if err != nil {
		return fmt.Errorf("bulk indexing: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk indexing: status %d, body: %s", res.StatusCode, string(respBody))
	}

	var bulkResp struct {
		Errors bool `json:"errors"`
		Items  []struct {
			Index struct {
				ID     string `json:"_id"`
				Status int    `json:"status"`
				Error  *struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"error,omitempty"`
			} `json:"index"`
		} `json:"items"`
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading bulk response: %w", err)
	}
	if err := json.Unmarshal(body, &bulkResp); err != nil {
		return fmt.Errorf("parsing bulk response: %w", err)
	}

	if bulkResp.Errors {
		for _, item := range bulkResp.Items {
			if item.Index.Error != nil {
				log.Error().
					Str("id", item.Index.ID).
					Str("error_type", item.Index.Error.Type).
					Str("reason", item.Index.Error.Reason).
					Msg("bulk index item failed")
			}
		}
		return fmt.Errorf("bulk indexing completed with errors")
	}

	log.Info().Int("count", len(docs)).Msg("bulk index completed")
	return nil
}

// DeleteDocument removes a document from the index.
func (ix *Indexer) DeleteDocument(ctx context.Context, id string) error {
	ctx, span := esIndexTracer.Start(ctx, "elasticsearch.delete",
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.operation", "delete"),
			attribute.String("db.elasticsearch.index", ix.client.indexName),
			attribute.String("db.elasticsearch.doc_id", id),
		),
	)
	defer span.End()

	res, err := ix.client.es.Delete(
		ix.client.indexName,
		id,
		ix.client.es.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("deleting document %s: %w", id, err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("deleting document %s: status %d, body: %s", id, res.StatusCode, string(respBody))
	}

	log.Debug().Str("id", id).Msg("document deleted")
	return nil
}

// IncrementViewCount atomically increments the view_count field of a document.
func (ix *Indexer) IncrementViewCount(ctx context.Context, id string) error {
	ctx, span := esIndexTracer.Start(ctx, "elasticsearch.update",
		trace.WithAttributes(
			attribute.String("db.system", "elasticsearch"),
			attribute.String("db.operation", "update"),
			attribute.String("db.elasticsearch.index", ix.client.indexName),
			attribute.String("db.elasticsearch.doc_id", id),
		),
	)
	defer span.End()

	script := `{"script":{"source":"ctx._source.view_count += 1","lang":"painless"},"upsert":{"view_count":1}}`

	res, err := ix.client.es.Update(
		ix.client.indexName,
		id,
		strings.NewReader(script),
		ix.client.es.Update.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("incrementing view count for %s: %w", id, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("incrementing view count for %s: status %d, body: %s", id, res.StatusCode, string(respBody))
	}

	log.Debug().Str("id", id).Msg("view count incremented")
	return nil
}
