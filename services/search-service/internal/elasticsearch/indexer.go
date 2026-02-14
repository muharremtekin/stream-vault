package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"

	"github.com/streamvault/search-service/internal/model"
)

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
