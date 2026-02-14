#!/bin/sh
set -e

ES_URL="http://elasticsearch:9200"
INDEX_NAME="streamvault-content"

echo "Waiting for Elasticsearch to be ready..."
until curl -sf "${ES_URL}/_cluster/health" > /dev/null 2>&1; do
  echo "Elasticsearch not ready yet, retrying in 3 seconds..."
  sleep 3
done

echo "Elasticsearch is ready. Checking if index '${INDEX_NAME}' exists..."

if curl -sf "${ES_URL}/${INDEX_NAME}" > /dev/null 2>&1; then
  echo "Index '${INDEX_NAME}' already exists. Skipping creation."
  exit 0
fi

echo "Creating index '${INDEX_NAME}' with Turkish analyzer and autocomplete mapping..."

curl -sf -X PUT "${ES_URL}/${INDEX_NAME}" \
  -H "Content-Type: application/json" \
  -d '{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "filter": {
        "turkish_stop": {
          "type": "stop",
          "stopwords": "_turkish_"
        },
        "turkish_lowercase": {
          "type": "lowercase",
          "language": "turkish"
        },
        "turkish_stemmer": {
          "type": "stemmer",
          "language": "turkish"
        },
        "edge_ngram_filter": {
          "type": "edge_ngram",
          "min_gram": 2,
          "max_gram": 15
        }
      },
      "analyzer": {
        "turkish_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "turkish_lowercase",
            "turkish_stop",
            "turkish_stemmer"
          ]
        },
        "autocomplete_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase",
            "edge_ngram_filter"
          ]
        },
        "autocomplete_search_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "lowercase"
          ]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "title": {
        "type": "text",
        "analyzer": "turkish_analyzer",
        "fields": {
          "autocomplete": {
            "type": "text",
            "analyzer": "autocomplete_analyzer",
            "search_analyzer": "autocomplete_search_analyzer"
          },
          "keyword": {
            "type": "keyword"
          },
          "english": {
            "type": "text",
            "analyzer": "english"
          }
        }
      },
      "original_title": {
        "type": "text",
        "analyzer": "english",
        "fields": {
          "autocomplete": {
            "type": "text",
            "analyzer": "autocomplete_analyzer",
            "search_analyzer": "autocomplete_search_analyzer"
          }
        }
      },
      "description": {
        "type": "text",
        "analyzer": "turkish_analyzer",
        "fields": {
          "english": {
            "type": "text",
            "analyzer": "english"
          }
        }
      },
      "content_type": {
        "type": "keyword"
      },
      "genres": {
        "type": "keyword"
      },
      "tags": {
        "type": "keyword"
      },
      "cast_names": {
        "type": "text",
        "analyzer": "standard"
      },
      "director": {
        "type": "text",
        "analyzer": "standard"
      },
      "release_year": {
        "type": "integer"
      },
      "maturity_rating": {
        "type": "keyword"
      },
      "average_rating": {
        "type": "float"
      },
      "rating_count": {
        "type": "integer"
      },
      "view_count": {
        "type": "long"
      },
      "video_status": {
        "type": "keyword"
      },
      "thumbnail_url": {
        "type": "keyword",
        "index": false
      },
      "banner_url": {
        "type": "keyword",
        "index": false
      },
      "duration_seconds": {
        "type": "integer"
      },
      "season_count": {
        "type": "integer"
      },
      "episode_count": {
        "type": "integer"
      },
      "created_at": {
        "type": "date"
      },
      "updated_at": {
        "type": "date"
      }
    }
  }
}'

echo ""
echo "Index '${INDEX_NAME}' created successfully."
echo "Verifying index mapping..."
curl -sf "${ES_URL}/${INDEX_NAME}/_mapping" | head -c 200
echo ""
echo "Done."
