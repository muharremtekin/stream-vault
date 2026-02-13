# StreamVault — Faz 3: Smart Features

> **Süre:** ~2-3 Hafta
> **Ön Koşul:** Faz 1 ve Faz 2 tamamlanmış olmalı (Gateway, User, Catalog, Streaming, Encoding çalışır durumda)
> **Hedef:** Akıllı arama (Elasticsearch), kişiselleştirilmiş öneriler (collaborative filtering), abonelik yönetimi (Saga pattern). CQRS ve Outbox pattern'lerinin gerçek implementasyonu.
> **Sonuç:** Kullanıcı içerik arayabilir, kişisel öneriler görebilir, abonelik satın alabilir/değiştirebilir/iptal edebilir. Sistem "akıllı" ve "ticari" hale gelir.

---

## 1. Faz 3'te Neler Ekleniyor?

| Bileşen | Dil | Yeni/Güncelleme | Açıklama |
|---------|-----|-----------------|----------|
| Search Service | Go | 🆕 Yeni | Elasticsearch üzerinde full-text arama, autocomplete, facet |
| Recommendation Engine | Rust | 🆕 Yeni | Collaborative filtering + content-based hibrit öneri |
| Subscription Service | .NET | 🆕 Yeni | Abonelik planları, ödeme (mock), Saga pattern |
| Elasticsearch | — | 🆕 Yeni altyapı | Full-text search engine |
| Catalog Service | .NET | 🔄 Güncelleme | Outbox pattern, event publishing genişletme |
| User Service | .NET | 🔄 Güncelleme | Subscription tier bilgisi, rating endpoint |
| API Gateway | Go | 🔄 Güncelleme | Yeni servis route'ları |

---

## 2. Yeni Proje Yapısı (Faz 1+2'ye Eklenenler)

```
streamvault/
├── ... (Faz 1 + Faz 2 yapısı aynen kalır)
│
├── proto/                                      # 🔄 Genişletildi
│   ├── ... (mevcut proto'lar)
│   ├── search/v1/search.proto                  # 🆕
│   ├── recommendation/v1/recommendation.proto  # 🆕
│   └── subscription/v1/subscription.proto      # 🆕
│
├── gateway/                                    # 🔄 Güncellendi
│   ├── internal/
│   │   └── proxy/
│   │       └── router.go                      # 🔄 Search, recommendation, subscription route'ları
│   └── ...
│
├── services/
│   ├── user-service/                           # 🔄 Güncellendi
│   │   ├── src/
│   │   │   ├── UserService.Application/
│   │   │   │   ├── Commands/
│   │   │   │   │   └── RateContent/           # 🆕
│   │   │   │   │       ├── RateContentCommand.cs
│   │   │   │   │       └── RateContentHandler.cs
│   │   │   │   └── Events/                    # 🆕
│   │   │   │       ├── ContentRatedEvent.cs
│   │   │   │       └── WatchCompletedEvent.cs
│   │   │   └── UserService.Domain/
│   │   │       └── Entities/
│   │   │           └── ContentRating.cs       # 🆕
│   │   └── ...
│   │
│   ├── catalog-service/                        # 🔄 Güncellendi
│   │   ├── src/
│   │   │   ├── CatalogService.Infrastructure/
│   │   │   │   ├── Messaging/
│   │   │   │   │   └── RabbitMqEventPublisher.cs  # 🔄 Outbox pattern eklendi
│   │   │   │   └── Outbox/                    # 🆕
│   │   │   │       ├── OutboxMessage.cs
│   │   │   │       ├── OutboxProcessor.cs     # Background service
│   │   │   │       └── OutboxRepository.cs
│   │   │   └── ...
│   │   └── ...
│   │
│   ├── streaming-service/                      # 🔄 Güncellendi
│   │   ├── internal/
│   │   │   └── handler/
│   │   │       └── progress.go                # 🔄 WatchCompleted event publish eklendi
│   │   └── ...
│   │
│   ├── search-service/                         # 🆕 Go — Search Service
│   │   ├── cmd/
│   │   │   └── search/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   │   └── config.go
│   │   │   ├── handler/
│   │   │   │   ├── search.go                 # Full-text arama endpoint
│   │   │   │   ├── autocomplete.go           # Otomatik tamamlama
│   │   │   │   ├── facets.go                 # Faceted arama (tür, yıl, rating)
│   │   │   │   └── trending.go              # Trend içerikler
│   │   │   ├── elasticsearch/
│   │   │   │   ├── client.go                # ES client wrapper
│   │   │   │   ├── indexer.go               # Document indexleme
│   │   │   │   ├── searcher.go              # Search query builder
│   │   │   │   ├── mapping.go               # Index mapping tanımları
│   │   │   │   └── analyzer.go              # Custom analyzer (Türkçe desteği)
│   │   │   ├── consumer/
│   │   │   │   ├── catalog_consumer.go      # ContentAdded/Updated event consumer
│   │   │   │   └── watch_consumer.go        # WatchCompleted consumer (popülerlik)
│   │   │   ├── model/
│   │   │   │   ├── search_document.go       # ES'e yazılan document yapısı
│   │   │   │   ├── search_request.go        # Arama isteği
│   │   │   │   └── search_response.go       # Arama sonucu
│   │   │   ├── cache/
│   │   │   │   └── redis.go                 # Popüler aramalar + sonuç cache
│   │   │   └── discovery/
│   │   │       └── consul.go
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── Makefile
│   │
│   ├── recommendation-service/                 # 🆕 Rust — Recommendation Engine
│   │   ├── src/
│   │   │   ├── main.rs
│   │   │   ├── config.rs
│   │   │   ├── api/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── routes.rs
│   │   │   │   └── handlers.rs              # HTTP + gRPC handlers
│   │   │   ├── engine/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── collaborative.rs         # User-based collaborative filtering
│   │   │   │   ├── content_based.rs         # Content-based filtering (genre, tag)
│   │   │   │   ├── hybrid.rs               # İki yaklaşımı birleştiren skor
│   │   │   │   ├── popularity.rs            # Popülerlik bazlı (cold start için)
│   │   │   │   └── similarity.rs            # Cosine similarity hesaplama
│   │   │   ├── model/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── user_profile.rs          # Kullanıcı tercih profili
│   │   │   │   ├── content_features.rs      # İçerik özellik vektörü
│   │   │   │   ├── interaction.rs           # İzleme/puan etkileşim kaydı
│   │   │   │   └── recommendation.rs        # Öneri sonuç yapısı
│   │   │   ├── store/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── postgres.rs              # Etkileşim verileri (kalıcı)
│   │   │   │   ├── redis.rs                 # Feature store (hızlı erişim)
│   │   │   │   └── cache.rs                 # Öneri sonuç cache
│   │   │   ├── consumer/
│   │   │   │   ├── mod.rs
│   │   │   │   ├── watch_consumer.rs        # WatchCompleted event
│   │   │   │   ├── rating_consumer.rs       # ContentRated event
│   │   │   │   └── catalog_consumer.rs      # ContentAdded event
│   │   │   ├── grpc/
│   │   │   │   ├── mod.rs
│   │   │   │   └── server.rs               # gRPC server (GetHomePageSections)
│   │   │   └── error.rs
│   │   ├── migrations/                      # sqlx migrations
│   │   │   ├── 001_create_interactions.sql
│   │   │   └── 002_create_user_profiles.sql
│   │   ├── Cargo.toml
│   │   ├── Dockerfile
│   │   └── Makefile
│   │
│   └── subscription-service/                   # 🆕 .NET — Subscription Service
│       ├── src/
│       │   ├── SubscriptionService.Api/
│       │   │   ├── Controllers/
│       │   │   │   ├── PlansController.cs
│       │   │   │   ├── SubscriptionsController.cs
│       │   │   │   └── PaymentsController.cs
│       │   │   ├── BackgroundServices/
│       │   │   │   ├── SagaOrchestratorService.cs     # Saga state machine
│       │   │   │   ├── SubscriptionRenewalService.cs  # Otomatik yenileme
│       │   │   │   └── OutboxProcessorService.cs      # Outbox event publishing
│       │   │   ├── Program.cs
│       │   │   └── appsettings.json
│       │   │
│       │   ├── SubscriptionService.Application/
│       │   │   ├── Commands/
│       │   │   │   ├── CreateSubscription/
│       │   │   │   │   ├── CreateSubscriptionCommand.cs
│       │   │   │   │   ├── CreateSubscriptionHandler.cs
│       │   │   │   │   └── CreateSubscriptionValidator.cs
│       │   │   │   ├── CancelSubscription/
│       │   │   │   │   ├── CancelSubscriptionCommand.cs
│       │   │   │   │   └── CancelSubscriptionHandler.cs
│       │   │   │   ├── ChangePlan/
│       │   │   │   │   ├── ChangePlanCommand.cs
│       │   │   │   │   └── ChangePlanHandler.cs
│       │   │   │   └── ProcessPayment/
│       │   │   │       ├── ProcessPaymentCommand.cs
│       │   │   │       └── ProcessPaymentHandler.cs
│       │   │   ├── Queries/
│       │   │   │   ├── GetPlans/
│       │   │   │   │   ├── GetPlansQuery.cs
│       │   │   │   │   └── GetPlansHandler.cs
│       │   │   │   ├── GetMySubscription/
│       │   │   │   │   ├── GetMySubscriptionQuery.cs
│       │   │   │   │   └── GetMySubscriptionHandler.cs
│       │   │   │   └── GetPaymentHistory/
│       │   │   │       ├── GetPaymentHistoryQuery.cs
│       │   │   │       └── GetPaymentHistoryHandler.cs
│       │   │   ├── Sagas/
│       │   │   │   ├── SubscriptionSaga.cs            # Saga state machine
│       │   │   │   ├── SubscriptionSagaState.cs       # Saga durumları
│       │   │   │   └── CompensatingActions.cs         # Geri alma aksiyonları
│       │   │   ├── DTOs/
│       │   │   │   ├── PlanDto.cs
│       │   │   │   ├── SubscriptionDto.cs
│       │   │   │   ├── PaymentDto.cs
│       │   │   │   └── InvoiceDto.cs
│       │   │   ├── Interfaces/
│       │   │   │   ├── IPlanRepository.cs
│       │   │   │   ├── ISubscriptionRepository.cs
│       │   │   │   ├── IPaymentRepository.cs
│       │   │   │   ├── IPaymentGateway.cs             # Mock ödeme arayüzü
│       │   │   │   ├── ISagaRepository.cs
│       │   │   │   └── IOutboxRepository.cs
│       │   │   └── Mappings/
│       │   │       └── MappingProfile.cs
│       │   │
│       │   ├── SubscriptionService.Domain/
│       │   │   ├── Entities/
│       │   │   │   ├── Plan.cs
│       │   │   │   ├── Subscription.cs
│       │   │   │   ├── Payment.cs
│       │   │   │   ├── Invoice.cs
│       │   │   │   ├── SagaState.cs
│       │   │   │   └── OutboxMessage.cs
│       │   │   ├── Enums/
│       │   │   │   ├── PlanTier.cs
│       │   │   │   ├── SubscriptionStatus.cs
│       │   │   │   ├── PaymentStatus.cs
│       │   │   │   └── SagaStep.cs
│       │   │   ├── Events/
│       │   │   │   ├── SubscriptionCreatedEvent.cs
│       │   │   │   ├── SubscriptionCancelledEvent.cs
│       │   │   │   ├── PlanChangedEvent.cs
│       │   │   │   └── PaymentProcessedEvent.cs
│       │   │   └── Exceptions/
│       │   │       ├── PlanNotFoundException.cs
│       │   │       ├── ActiveSubscriptionExistsException.cs
│       │   │       └── PaymentFailedException.cs
│       │   │
│       │   └── SubscriptionService.Infrastructure/
│       │       ├── Persistence/
│       │       │   ├── SubscriptionDbContext.cs
│       │       │   ├── Configurations/
│       │       │   │   ├── PlanConfiguration.cs
│       │       │   │   ├── SubscriptionConfiguration.cs
│       │       │   │   ├── PaymentConfiguration.cs
│       │       │   │   ├── SagaStateConfiguration.cs
│       │       │   │   └── OutboxMessageConfiguration.cs
│       │       │   └── Repositories/
│       │       │       ├── PlanRepository.cs
│       │       │       ├── SubscriptionRepository.cs
│       │       │       ├── PaymentRepository.cs
│       │       │       ├── SagaRepository.cs
│       │       │       └── OutboxRepository.cs
│       │       ├── Payment/
│       │       │   └── MockPaymentGateway.cs          # Simüle ödeme
│       │       ├── Messaging/
│       │       │   └── RabbitMqEventPublisher.cs
│       │       ├── Migrations/
│       │       └── DependencyInjection.cs
│       │
│       ├── tests/
│       │   ├── SubscriptionService.UnitTests/
│       │   │   ├── Sagas/
│       │   │   │   └── SubscriptionSagaTests.cs
│       │   │   └── Commands/
│       │   │       └── CreateSubscriptionHandlerTests.cs
│       │   └── SubscriptionService.IntegrationTests/
│       │       └── SagaIntegrationTests.cs
│       │
│       ├── Dockerfile
│       └── Makefile
│
├── infrastructure/
│   ├── ... (Faz 1 + Faz 2'den)
│   ├── elasticsearch/                          # 🆕
│   │   └── elasticsearch.yml
│   └── postgres/
│       └── init.sql                            # 🔄 subscription + recommendation DB'leri eklendi
│
└── scripts/
    ├── ... (Faz 1 + Faz 2'den)
    ├── seed-search-index.sh                    # 🆕 ES index'ini catalog'dan doldur
    ├── test-search.sh                          # 🆕 Arama test senaryoları
    ├── test-subscription.sh                    # 🆕 Abonelik akışı testi
    └── generate-interactions.sh                # 🆕 Fake izleme/puan verisi üretimi
```

---

## 3. gRPC Proto Tanımları

### 3.1 search/v1/search.proto

```protobuf
syntax = "proto3";
package streamvault.search.v1;
option go_package = "github.com/streamvault/proto/search/v1";

service SearchService {
  // Full-text arama
  rpc Search(SearchRequest) returns (SearchResponse);
  // Otomatik tamamlama
  rpc Autocomplete(AutocompleteRequest) returns (AutocompleteResponse);
  // Trend içerikler
  rpc GetTrending(GetTrendingRequest) returns (GetTrendingResponse);
}

message SearchRequest {
  string query = 1;
  repeated string genres = 2;          // Filtre: türler
  int32 year_from = 3;                 // Filtre: yıl aralığı
  int32 year_to = 4;
  string maturity_rating = 5;          // Filtre: yaş sınıfı
  double min_rating = 6;               // Filtre: minimum puan
  string content_type = 7;             // "movie", "series", "all"
  string sort_by = 8;                  // "relevance", "rating", "year", "title"
  string sort_order = 9;               // "asc", "desc"
  int32 page = 10;
  int32 page_size = 11;
}

message SearchResponse {
  repeated SearchHit hits = 1;
  Facets facets = 2;
  int64 total_count = 3;
  int32 page = 4;
  int32 page_size = 5;
  double took_ms = 6;                  // Arama süresi
}

message SearchHit {
  string id = 1;
  string title = 2;
  string description = 3;
  string content_type = 4;             // "movie" | "series"
  int32 release_year = 5;
  repeated string genres = 6;
  double average_rating = 7;
  string thumbnail_url = 8;
  string maturity_rating = 9;
  double score = 10;                   // ES relevance skoru
  HighlightFields highlights = 11;     // Eşleşen kısımlar vurgulu
}

message HighlightFields {
  string title = 1;                    // "<em>Inter</em>stellar"
  string description = 2;
}

message Facets {
  repeated FacetBucket genres = 1;
  repeated FacetBucket years = 2;
  repeated FacetBucket maturity_ratings = 3;
  repeated FacetBucket content_types = 4;
}

message FacetBucket {
  string key = 1;
  int64 doc_count = 2;
}

message AutocompleteRequest {
  string query = 1;                    // Minimum 2 karakter
  int32 limit = 2;                     // Default 5
}

message AutocompleteResponse {
  repeated AutocompleteSuggestion suggestions = 1;
}

message AutocompleteSuggestion {
  string id = 1;
  string title = 2;
  string content_type = 3;
  string thumbnail_url = 4;
  int32 release_year = 5;
}

message GetTrendingRequest {
  string time_window = 1;             // "day", "week", "month"
  int32 limit = 2;
}

message GetTrendingResponse {
  repeated TrendingItem items = 1;
}

message TrendingItem {
  string id = 1;
  string title = 2;
  string content_type = 3;
  string thumbnail_url = 4;
  int64 view_count = 5;
  int32 rank = 6;
  int32 rank_change = 7;              // +3, -1, 0 (önceki güne göre)
}
```

### 3.2 recommendation/v1/recommendation.proto

```protobuf
syntax = "proto3";
package streamvault.recommendation.v1;
option go_package = "github.com/streamvault/proto/recommendation/v1";

service RecommendationService {
  // Kullanıcı için kişisel öneriler
  rpc GetRecommendations(GetRecommendationsRequest) returns (GetRecommendationsResponse);
  // Belirli bir içeriğe benzer içerikler
  rpc GetSimilar(GetSimilarRequest) returns (GetSimilarResponse);
  // Ana sayfa section'ları (birden fazla kategori)
  rpc GetHomePageSections(GetHomePageSectionsRequest) returns (GetHomePageSectionsResponse);
}

message GetRecommendationsRequest {
  string user_id = 1;
  int32 limit = 2;
  repeated string exclude_ids = 3;     // Zaten izlenmiş içerikleri hariç tut
}

message GetRecommendationsResponse {
  repeated RecommendedItem items = 1;
  string algorithm = 2;                // "hybrid", "collaborative", "popularity"
}

message RecommendedItem {
  string content_id = 1;
  string title = 2;
  string content_type = 3;
  string thumbnail_url = 4;
  repeated string genres = 5;
  double average_rating = 6;
  double prediction_score = 7;         // Tahmin edilen beğeni skoru (0-10)
  string reason = 8;                   // "Aksiyon filmlerini sevdiğiniz için"
}

message GetSimilarRequest {
  string content_id = 1;
  int32 limit = 2;
}

message GetSimilarResponse {
  repeated SimilarItem items = 1;
}

message SimilarItem {
  string content_id = 1;
  string title = 2;
  string content_type = 3;
  string thumbnail_url = 4;
  repeated string genres = 5;
  double similarity_score = 6;         // Benzerlik skoru (0.0-1.0)
}

message GetHomePageSectionsRequest {
  string user_id = 1;
}

message GetHomePageSectionsResponse {
  repeated HomePageSection sections = 1;
}

message HomePageSection {
  string title = 1;                     // "Sizin İçin", "Trend Olanlar", "Çünkü X İzlediniz"
  string section_type = 2;             // "personal", "trending", "because_you_watched", "genre", "new"
  repeated RecommendedItem items = 3;
}
```

### 3.3 subscription/v1/subscription.proto

```protobuf
syntax = "proto3";
package streamvault.subscription.v1;
option go_package = "github.com/streamvault/proto/subscription/v1";
option csharp_namespace = "StreamVault.Proto.Subscription.V1";

import "common/v1/common.proto";
import "google/protobuf/timestamp.proto";

// Diğer servisler tarafından çağrılır (özellikle Gateway ve Streaming)
service SubscriptionQueryService {
  // Kullanıcının aktif abonelik bilgisi
  rpc GetUserSubscription(GetUserSubscriptionRequest) returns (GetUserSubscriptionResponse);
  // Kullanıcının tier'ini hızlıca öğren (erişim kontrolü için)
  rpc GetUserTier(GetUserTierRequest) returns (GetUserTierResponse);
}

message GetUserSubscriptionRequest {
  string user_id = 1;
}

message GetUserSubscriptionResponse {
  string subscription_id = 1;
  string user_id = 2;
  streamvault.common.v1.SubscriptionTier tier = 3;
  string status = 4;                   // "active", "cancelled", "expired"
  google.protobuf.Timestamp current_period_start = 5;
  google.protobuf.Timestamp current_period_end = 6;
  bool auto_renew = 7;
}

message GetUserTierRequest {
  string user_id = 1;
}

message GetUserTierResponse {
  streamvault.common.v1.SubscriptionTier tier = 1;
  bool is_active = 2;
}
```

---

## 4. Servis Detayları

### 4.1 Search Service (Go)

**Sorumluluklar:**

- Elasticsearch üzerinde full-text arama (Türkçe + İngilizce)
- Autocomplete (prefix + fuzzy matching)
- Faceted arama (tür, yıl, rating'e göre filtreleme ve sayaç)
- Trend içerikler listesi (izleme sayısına göre)
- CQRS read model: Catalog'dan gelen event'lerle ES index'i senkronize tutar
- Popüler aramalar (Redis sorted set)
- Sonuç caching (Redis)

**Elasticsearch Index Mapping:**

```json
PUT /streamvault-content
{
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
            "turkish_lowercase",
            "edge_ngram_filter"
          ]
        },
        "autocomplete_search_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": [
            "turkish_lowercase"
          ]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id":              { "type": "keyword" },
      "title":           {
        "type": "text",
        "analyzer": "turkish_analyzer",
        "fields": {
          "autocomplete": {
            "type": "text",
            "analyzer": "autocomplete_analyzer",
            "search_analyzer": "autocomplete_search_analyzer"
          },
          "keyword": { "type": "keyword" },
          "english": { "type": "text", "analyzer": "english" }
        }
      },
      "original_title":  {
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
      "description":     {
        "type": "text",
        "analyzer": "turkish_analyzer",
        "fields": {
          "english": { "type": "text", "analyzer": "english" }
        }
      },
      "content_type":    { "type": "keyword" },
      "genres":          { "type": "keyword" },
      "tags":            { "type": "keyword" },
      "cast_names":      { "type": "text", "analyzer": "standard" },
      "director":        { "type": "text", "analyzer": "standard" },
      "release_year":    { "type": "integer" },
      "maturity_rating": { "type": "keyword" },
      "average_rating":  { "type": "float" },
      "rating_count":    { "type": "integer" },
      "view_count":      { "type": "long" },
      "video_status":    { "type": "keyword" },
      "thumbnail_url":   { "type": "keyword", "index": false },
      "banner_url":      { "type": "keyword", "index": false },
      "duration_seconds": { "type": "integer" },
      "season_count":    { "type": "integer" },
      "episode_count":   { "type": "integer" },
      "created_at":      { "type": "date" },
      "updated_at":      { "type": "date" }
    }
  }
}
```

**CQRS Akışı — Catalog → Search Senkronizasyonu:**

```
Catalog Service (Write Model — MongoDB)
    │
    │  1. Admin yeni film ekler (POST /api/catalog/movies)
    │  2. MongoDB'ye yazar
    │  3. Outbox tablosuna event yazar (aynı transaction)
    │  4. OutboxProcessor event'i RabbitMQ'ya publish eder
    │
    ▼  RabbitMQ: catalog.events exchange
    │
    │  routing_key = "content.created" / "content.updated" / "content.deleted"
    │
    ▼
Search Service (Read Model — Elasticsearch)
    │
    │  1. catalog_consumer.go event'i alır
    │  2. Event tipine göre:
    │     - content.created → ES'e yeni document index'le
    │     - content.updated → ES'teki document'ı güncelle
    │     - content.deleted → ES'ten document'ı sil
    │  3. ACK gönder
    │
    ▼
Elasticsearch (streamvault-content index)
    │
    │  Kullanıcı arama yapar → Search Service ES'e sorgu atar → Sonuçları döner
    │
    ▼
Client: Arama sonuçları
```

**HTTP API Kontratı:**

```
GET /api/search?q=interstellar&genres=bilim-kurgu&yearFrom=2010&sort=rating_desc&page=1&pageSize=20
  Response: 200 {
    "hits": [
      {
        "id": "abc",
        "title": "Interstellar",
        "description": "Dünya'nın geleceği tehlikede...",
        "contentType": "movie",
        "releaseYear": 2014,
        "genres": ["bilim-kurgu", "dram"],
        "averageRating": 8.7,
        "thumbnailUrl": "...",
        "maturityRating": "PG13",
        "score": 12.45,
        "highlights": {
          "title": "<em>Interstellar</em>",
          "description": "...uzay ve zaman üzerine epik bir yolculuk..."
        }
      }
    ],
    "facets": {
      "genres": [
        { "key": "bilim-kurgu", "docCount": 32 },
        { "key": "dram", "docCount": 28 },
        { "key": "aksiyon", "docCount": 45 }
      ],
      "years": [
        { "key": "2024", "docCount": 12 },
        { "key": "2023", "docCount": 15 }
      ],
      "contentTypes": [
        { "key": "movie", "docCount": 85 },
        { "key": "series", "docCount": 42 }
      ]
    },
    "totalCount": 156,
    "page": 1,
    "pageSize": 20,
    "tookMs": 12.3
  }

GET /api/search/autocomplete?q=inter&limit=5
  Response: 200 {
    "suggestions": [
      { "id": "abc", "title": "Interstellar", "contentType": "movie", "thumbnailUrl": "...", "releaseYear": 2014 },
      { "id": "def", "title": "The Internship", "contentType": "movie", "thumbnailUrl": "...", "releaseYear": 2013 }
    ]
  }

GET /api/search/trending?window=week&limit=10
  Response: 200 {
    "items": [
      {
        "id": "abc",
        "title": "Interstellar",
        "contentType": "movie",
        "thumbnailUrl": "...",
        "viewCount": 15420,
        "rank": 1,
        "rankChange": 2
      }
    ]
  }

GET /health
  Response: 200 { "status": "healthy", "elasticsearch": "connected", "rabbitmq": "connected" }
```

**Redis Veri Yapıları:**

```
# Popüler aramalar (sorted set — score = arama sayısı)
popular_searches → Sorted Set
  member: "interstellar"
  score: 1542

# Arama sonucu cache (hash — TTL 5 dakika)
search_cache:{query_hash} → String (JSON)
  TTL: 300s

# Trend içerikler (sorted set — score = izleme sayısı)
trending:daily → Sorted Set
  member: "content-id-abc"
  score: 342

trending:weekly → Sorted Set
  ...

# İçerik view count (hash — günlük artış)
views:daily:{date} → Hash
  field: "content-id-abc"
  value: 342
  TTL: 8 gün
```

**Konfigürasyon (config.yaml):**

```yaml
server:
  http_port: 5005
  grpc_port: 50053

elasticsearch:
  urls: ["http://elasticsearch:9200"]
  index: "streamvault-content"
  request_timeout: 5s

rabbitmq:
  url: "amqp://streamvault:secret@rabbitmq:5672/"
  catalog_exchange: "catalog.events"
  catalog_queue: "search.catalog-sync"
  watch_exchange: "watch.events"
  watch_queue: "search.watch-count"

redis:
  url: "redis://redis:6379/3"
  cache_ttl: "5m"
  trending_ttl: "1h"

consul:
  address: "consul:8500"
  service_name: "search-service"
  service_port: 5005
```

**Öğrenme Noktaları:**

- Elasticsearch index mapping, custom analyzer (Türkçe)
- Full-text search, fuzzy matching, autocomplete (edge_ngram)
- Faceted/aggregation sorgular
- CQRS pattern (write model ≠ read model, eventual consistency)
- Event consumer ile index senkronizasyonu
- Arama sonucu caching stratejisi

---

### 4.2 Recommendation Engine (Rust)

**Sorumluluklar:**

- Kullanıcıya kişiselleştirilmiş içerik önerileri
- İçerik benzerliği hesaplama ("Bunu beğendiyseniz..." listesi)
- Ana sayfa section'ları (kişisel, trending, tür bazlı, "çünkü X izlediniz")
- 3 algoritma: collaborative filtering, content-based filtering, hibrit
- Cold start problemi çözümü (yeni kullanıcılar → popülerlik bazlı)
- Event'lerden öğrenme: izleme tamamlama, puanlama, watchlist ekleme

**Öneri Algoritmaları:**

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Hybrid Recommendation Engine                   │
│                                                                     │
│  ┌─────────────────────┐    ┌─────────────────────┐                │
│  │  Collaborative       │    │  Content-Based       │                │
│  │  Filtering           │    │  Filtering           │                │
│  │                     │    │                     │                │
│  │  "Seni senin gibi   │    │  "İzlediklerinin    │                │
│  │   izleyenlere göre  │    │   türüne/tag'ine    │                │
│  │   eşleştir"         │    │   göre benzer bul"  │                │
│  │                     │    │                     │                │
│  │  Input:             │    │  Input:             │                │
│  │  - User-Item matrix │    │  - İçerik feature   │                │
│  │    (izleme+rating)  │    │    vektörleri        │                │
│  │  - Benzer kullanıcı │    │  - User preference  │                │
│  │    bulma (cosine)   │    │    profili           │                │
│  │                     │    │                     │                │
│  │  Output:            │    │  Output:            │                │
│  │  - Tahmin skoru     │    │  - Benzerlik skoru  │                │
│  │    (0-10)           │    │    (0.0-1.0)        │                │
│  └────────┬────────────┘    └────────┬────────────┘                │
│           │                          │                              │
│           └──────────┬───────────────┘                              │
│                      ▼                                              │
│           ┌─────────────────────┐                                   │
│           │  Hybrid Scorer      │                                   │
│           │                     │                                   │
│           │  final_score =      │                                   │
│           │    α × collab_score │  ← α = collaborative ağırlık     │
│           │  + β × content_score│  ← β = content-based ağırlık     │
│           │  + γ × popularity   │  ← γ = popülerlik ağırlık        │
│           │                     │                                   │
│           │  α, β, γ kullanıcı  │                                   │
│           │  etkileşim sayısına  │                                   │
│           │  göre değişir:       │                                   │
│           │  - Yeni kullanıcı:   │                                   │
│           │    γ yüksek (cold    │                                   │
│           │    start)            │                                   │
│           │  - Aktif kullanıcı:  │                                   │
│           │    α yüksek          │                                   │
│           └─────────────────────┘                                   │
└─────────────────────────────────────────────────────────────────────┘
```

**Collaborative Filtering — User-Item Matrix:**

```
               Film A   Film B   Film C   Film D   Film E
Kullanıcı 1   [ 8.0     5.0      ?        9.0      ?    ]
Kullanıcı 2   [ 7.0     ?        6.0      8.5      4.0  ]
Kullanıcı 3   [ ?       4.0      7.0      ?        8.0  ]
Kullanıcı 4   [ 9.0     6.0      ?        ?        7.0  ]
                                  ↑
                         Bu değerleri tahmin et!

Adımlar:
1. Kullanıcı benzerliği hesapla (cosine similarity)
   sim(user1, user2) = dot(u1, u2) / (||u1|| × ||u2||)
   (Sadece ortak puanlanmış filmler üzerinden)

2. En benzer K kullanıcıyı bul (K=20)

3. Ağırlıklı ortalama ile tahmin:
   pred(user1, filmC) = Σ(sim × rating) / Σ(|sim|)
   benzer kullanıcıların filmC puanlarının ağırlıklı ortalaması

İzleme verisi de implicit feedback olarak kullanılır:
- İzleme tamamlama = implicit 7.0 puan
- %50+ izleme = implicit 5.0 puan
- Watchlist'e ekleme = implicit 6.0 puan
```

**Content-Based Filtering — Feature Vectors:**

```
Her içerik bir feature vektörüne dönüştürülür:

Film: "Interstellar"
Feature Vector: [
  genres:    [0, 0, 1, 1, 0, 0, 1, 0]   # bilim-kurgu=1, dram=1, macera=1
  tags:      [1, 0, 1, 0, 0, 1, 0, 0]   # uzay=1, zaman-yolculugu=1, aile=1
  director:  [0, 0, 1, 0, 0, ...]        # nolan=1
  year_norm: [0.85]                       # (2014-1950)/(2026-1950)
  rating_norm: [0.87]                     # 8.7/10
]

Kullanıcı preference profili:
  İzlediği ve beğendiği filmlerin feature vektörlerinin ağırlıklı ortalaması.

Öneri:
  cosine_similarity(user_preference_vector, content_feature_vector)
  En yüksek benzerliğe sahip izlenmemiş içerikler önerilir.
```

**Veritabanı (PostgreSQL — recommendation DB):**

```sql
-- Kullanıcı etkileşimleri (izleme, puanlama)
CREATE TABLE interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    content_id VARCHAR(50) NOT NULL,
    interaction_type VARCHAR(20) NOT NULL,  -- 'watch', 'rating', 'watchlist'
    value DOUBLE PRECISION,                 -- rating: 1-10, watch: completion %
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_interaction UNIQUE (user_id, content_id, interaction_type)
);

CREATE INDEX idx_interactions_user ON interactions(user_id);
CREATE INDEX idx_interactions_content ON interactions(content_id);
CREATE INDEX idx_interactions_type ON interactions(interaction_type);

-- Kullanıcı tercih profili (önceden hesaplanmış)
CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY,
    genre_weights JSONB NOT NULL DEFAULT '{}',     -- {"bilim-kurgu": 0.8, "dram": 0.6}
    tag_weights JSONB NOT NULL DEFAULT '{}',       -- {"uzay": 0.9, "aile": 0.4}
    avg_rating DOUBLE PRECISION DEFAULT 0,
    interaction_count INT DEFAULT 0,
    last_computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- İçerik feature vektörleri (önceden hesaplanmış)
CREATE TABLE content_features (
    content_id VARCHAR(50) PRIMARY KEY,
    feature_vector JSONB NOT NULL,                  -- Normalize edilmiş feature vektörü
    genres JSONB NOT NULL DEFAULT '[]',
    tags JSONB NOT NULL DEFAULT '[]',
    avg_rating DOUBLE PRECISION DEFAULT 0,
    view_count BIGINT DEFAULT 0,
    last_computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Önceden hesaplanmış benzerlik matrisi (top-N benzerlikleri tut)
CREATE TABLE content_similarity (
    content_id_a VARCHAR(50) NOT NULL,
    content_id_b VARCHAR(50) NOT NULL,
    similarity_score DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (content_id_a, content_id_b)
);

CREATE INDEX idx_similarity_a ON content_similarity(content_id_a, similarity_score DESC);
```

**Redis Feature Store (Hızlı Erişim):**

```
# Kullanıcı tercih profili (cache — TTL 1 saat)
rec:user_profile:{userId} → Hash
  genre_weights: '{"bilim-kurgu": 0.8, "dram": 0.6}'
  tag_weights: '{"uzay": 0.9}'
  interaction_count: 47
  TTL: 3600s

# İçerik feature vektörü (cache — TTL 6 saat)
rec:content_features:{contentId} → Hash
  feature_vector: '[0.8, 0.6, 0.0, ...]'
  genres: '["bilim-kurgu", "dram"]'
  avg_rating: 8.7
  TTL: 21600s

# Önceden hesaplanmış öneri sonuçları (cache — TTL 30 dakika)
rec:recommendations:{userId} → String (JSON array)
  TTL: 1800s

# "Bu içeriğe benzer" sonuçları (cache — TTL 6 saat)
rec:similar:{contentId} → String (JSON array)
  TTL: 21600s
```

**HTTP API Kontratı:**

```
GET /api/recommendations?limit=20
  Headers: X-User-Id
  Response: 200 {
    "items": [
      {
        "contentId": "abc",
        "title": "Inception",
        "contentType": "movie",
        "thumbnailUrl": "...",
        "genres": ["bilim-kurgu", "aksiyon"],
        "averageRating": 8.8,
        "predictionScore": 9.1,
        "reason": "Interstellar'ı beğendiğiniz için"
      }
    ],
    "algorithm": "hybrid"
  }

GET /api/recommendations/similar/{contentId}?limit=10
  Response: 200 {
    "items": [
      {
        "contentId": "def",
        "title": "The Martian",
        "contentType": "movie",
        "thumbnailUrl": "...",
        "genres": ["bilim-kurgu", "macera"],
        "similarityScore": 0.87
      }
    ]
  }

GET /api/recommendations/home
  Headers: X-User-Id
  Response: 200 {
    "sections": [
      {
        "title": "Sizin İçin Seçtiklerimiz",
        "sectionType": "personal",
        "items": [ ... ]
      },
      {
        "title": "Türkiye'de Trend",
        "sectionType": "trending",
        "items": [ ... ]
      },
      {
        "title": "Interstellar İzlediğiniz İçin",
        "sectionType": "because_you_watched",
        "items": [ ... ]
      },
      {
        "title": "Bilim Kurgu",
        "sectionType": "genre",
        "items": [ ... ]
      },
      {
        "title": "Yeni Eklenenler",
        "sectionType": "new",
        "items": [ ... ]
      }
    ]
  }

POST /api/recommendations/feedback
  Headers: X-User-Id
  Request:  { "contentId": "abc", "action": "not_interested" }
  Response: 200 { "acknowledged": true }

GET /health
  Response: 200 { "status": "healthy", "postgres": "connected", "redis": "connected" }
```

**Konfigürasyon (config.toml):**

```toml
[server]
http_port = 5006
grpc_port = 50054

[postgres]
url = "postgres://streamvault:secret@postgres:5432/streamvault_recommendations"
max_connections = 10

[redis]
url = "redis://redis:6379/4"
profile_ttl_seconds = 3600
recommendation_ttl_seconds = 1800
similar_ttl_seconds = 21600

[rabbitmq]
url = "amqp://streamvault:secret@rabbitmq:5672/"
watch_queue = "recommendation.watch-events"
rating_queue = "recommendation.rating-events"
catalog_queue = "recommendation.catalog-events"

[engine]
collaborative_weight = 0.5          # α
content_based_weight = 0.3          # β
popularity_weight = 0.2             # γ
cold_start_threshold = 5            # Bu kadar etkileşimden azı → cold start
similar_top_n = 50                  # Her içerik için top-50 benzerlik sakla
recompute_interval_hours = 6        # Profil ve benzerlik yeniden hesaplama

[consul]
address = "consul:8500"
service_name = "recommendation-service"
service_port = 5006
```

**Öğrenme Noktaları:**

- Collaborative filtering (user-based, cosine similarity)
- Content-based filtering (feature vektörü, TF-IDF benzeri yaklaşım)
- Cold start problemi ve çözüm stratejileri
- Rust'ta matrix/vektör işlemleri (ndarray veya manuel)
- Feature store pattern (Redis'te önceden hesaplanmış veriler)
- Batch hesaplama vs online hesaplama trade-off'u
- gRPC server (tonic crate)

---

### 4.3 Subscription Service (.NET 8)

**Sorumluluklar:**

- Abonelik planı tanımlama ve listeleme (Basic, Standard, Premium)
- Abonelik oluşturma (Saga pattern ile çok adımlı iş akışı)
- Plan değiştirme (upgrade/downgrade)
- Abonelik iptal etme (dönem sonuna kadar aktif kalır)
- Otomatik yenileme (background job)
- Mock ödeme sistemi (başarılı/başarısız senaryolar)
- Fatura geçmişi
- Outbox pattern ile güvenilir event publishing

**Domain Modelleri:**

```
Plan
├── Id: Guid
├── Name: string                         # "Basic", "Standard", "Premium"
├── Tier: enum (Basic, Standard, Premium)
├── PriceMonthly: decimal                # 49.99, 79.99, 119.99 (TRY)
├── MaxScreens: int                      # 1, 2, 4
├── MaxQuality: string                   # "720p", "1080p", "4K"
├── Features: List<string>              # ["HD", "Ad-free", "Downloads"]
├── IsActive: bool
├── CreatedAt: DateTime
└── UpdatedAt: DateTime

Subscription
├── Id: Guid
├── UserId: Guid
├── PlanId: Guid
├── Plan: Plan (navigation)
├── Status: enum (PendingPayment, Active, Cancelled, Expired, Failed)
├── CurrentPeriodStart: DateTime
├── CurrentPeriodEnd: DateTime
├── AutoRenew: bool (default: true)
├── CancelledAt: DateTime?
├── CancellationReason: string?
├── CreatedAt: DateTime
└── UpdatedAt: DateTime

Payment
├── Id: Guid
├── SubscriptionId: Guid
├── Amount: decimal
├── Currency: string (TRY)
├── Status: enum (Pending, Succeeded, Failed, Refunded)
├── PaymentMethod: string                # "mock_card_visa", "mock_card_mc"
├── TransactionId: string                # Mock payment gateway ID
├── FailureReason: string?
├── ProcessedAt: DateTime?
├── CreatedAt: DateTime
└── UpdatedAt: DateTime

Invoice
├── Id: Guid
├── SubscriptionId: Guid
├── PaymentId: Guid
├── InvoiceNumber: string                # "INV-2026-000001"
├── Amount: decimal
├── Currency: string
├── PeriodStart: DateTime
├── PeriodEnd: DateTime
├── IssuedAt: DateTime
└── PdfUrl: string?

SagaState (Saga pattern durumu)
├── Id: Guid
├── SagaType: string                     # "CreateSubscription", "ChangePlan"
├── CorrelationId: string                # İlişkili işlem ID'si
├── CurrentStep: enum
├── Status: enum (InProgress, Completed, Compensating, Failed)
├── StateData: string (JSON)             # Adım verisi
├── CreatedAt: DateTime
├── UpdatedAt: DateTime
└── CompletedAt: DateTime?

OutboxMessage (Outbox pattern)
├── Id: Guid
├── EventType: string                    # "SubscriptionCreated"
├── Payload: string (JSON)
├── CreatedAt: DateTime
├── ProcessedAt: DateTime?
├── RetryCount: int
└── Error: string?
```

**Saga Pattern — Abonelik Oluşturma Akışı:**

```
CreateSubscription Saga
═══════════════════════

Normal Akış (Happy Path):
─────────────────────────

  Step 1: VALIDATE_PLAN
  ├── Plan mevcut mu ve aktif mi kontrol et
  ├── Kullanıcının zaten aktif aboneliği var mı kontrol et
  └── Başarılı → Step 2'ye geç

  Step 2: CREATE_PENDING_SUBSCRIPTION
  ├── Subscription kaydı oluştur (Status: PendingPayment)
  ├── Outbox'a yaz (aynı DB transaction)
  └── Başarılı → Step 3'e geç

  Step 3: PROCESS_PAYMENT
  ├── MockPaymentGateway.ChargeAsync(amount, paymentMethod)
  ├── Payment kaydı oluştur
  ├── Başarılı → Step 4'e geç
  └── Başarısız → COMPENSATE

  Step 4: ACTIVATE_SUBSCRIPTION
  ├── Subscription.Status = Active
  ├── CurrentPeriodStart = now
  ├── CurrentPeriodEnd = now + 30 days
  └── Başarılı → Step 5'e geç

  Step 5: CREATE_INVOICE
  ├── Invoice kaydı oluştur
  └── Başarılı → Step 6'ya geç

  Step 6: PUBLISH_EVENTS
  ├── SubscriptionCreated event → RabbitMQ
  │   → User Service: kullanıcının tier bilgisini güncelle
  │   → (Faz 4) Notification Service: hoş geldin emaili
  └── Saga tamamlandı ✓


Compensating Akış (Hata Durumu):
────────────────────────────────

  Step 3 başarısız (ödeme reddedildi):
  ├── COMPENSATE Step 2: Subscription'ı sil veya Status = Failed yap
  ├── COMPENSATE Step 1: (bir şey yapmaya gerek yok)
  └── PaymentFailed event publish et

  Compensating action'lar ters sırada çalışır (3 → 2 → 1).
  Her adım idempotent olmalı (tekrar çalışsa aynı sonucu vermeli).
```

```
ChangePlan Saga (Plan Değiştirme)
═════════════════════════════════

  Step 1: VALIDATE
  ├── Yeni plan mevcut ve farklı mı?
  ├── Aktif abonelik var mı?
  └── Fiyat farkı hesapla (upgrade: ek ücret, downgrade: kredi)

  Step 2: PROCESS_PRICE_DIFFERENCE
  ├── Upgrade: Kalan gün için fark ücreti al
  ├── Downgrade: Kalan gün için kredi kaydet
  └── Payment kaydı oluştur

  Step 3: UPDATE_SUBSCRIPTION
  ├── PlanId güncelle
  ├── Fiyat farkına göre period ayarla
  └── Yeni Invoice oluştur

  Step 4: PUBLISH_EVENTS
  ├── PlanChanged event
  │   → User Service: tier güncelle
  │   → Streaming Service: kalite limiti güncelle
  └── Saga tamamlandı ✓
```

**Outbox Pattern — Güvenilir Event Publishing:**

```
Sorun: Bir DB işlemi ve RabbitMQ publish ayrı transaction'lar.
       DB başarılı + RabbitMQ başarısız = tutarsızlık.

Çözüm: Outbox Pattern

  1. Subscription Handler:
     BEGIN TRANSACTION
       INSERT INTO subscriptions (...)
       INSERT INTO outbox_messages (event_type, payload, created_at)
     COMMIT

  2. OutboxProcessorService (Background):
     Her 5 saniyede bir:
       SELECT * FROM outbox_messages WHERE processed_at IS NULL ORDER BY created_at LIMIT 50
       For each message:
         → RabbitMQ'ya publish et
         → processed_at = NOW() olarak güncelle
       Hata durumunda:
         → retry_count++
         → 3 denemeden sonra error logla, skip et

  Bu sayede DB ve event publishing atomik davranır.
  En kötü durumda duplicate event olabilir (idempotent consumer gerekir).
```

**Mock Payment Gateway:**

```csharp
// Test senaryoları için mock ödeme sistemi
public class MockPaymentGateway : IPaymentGateway
{
    // Kart numarasına göre sonuç belirlenir:
    // "4242424242424242" → Başarılı
    // "4000000000000002" → Reddedildi (insufficient funds)
    // "4000000000000069" → Süresi dolmuş kart
    // "4000000000000127" → Genel hata
    // Random delay: 200-800ms (gerçekçi latency simülasyonu)
}
```

**API Kontratı:**

```
GET /api/plans
  Response: 200 [
    {
      "id": "guid",
      "name": "Basic",
      "tier": "Basic",
      "priceMonthly": 49.99,
      "currency": "TRY",
      "maxScreens": 1,
      "maxQuality": "720p",
      "features": ["Reklamsız izleme", "Mobil indirme"]
    },
    {
      "id": "guid",
      "name": "Standard",
      "tier": "Standard",
      "priceMonthly": 79.99,
      "currency": "TRY",
      "maxScreens": 2,
      "maxQuality": "1080p",
      "features": ["Reklamsız izleme", "Mobil indirme", "Full HD"]
    },
    {
      "id": "guid",
      "name": "Premium",
      "tier": "Premium",
      "priceMonthly": 119.99,
      "currency": "TRY",
      "maxScreens": 4,
      "maxQuality": "4K",
      "features": ["Reklamsız izleme", "Mobil indirme", "Ultra HD 4K", "Dolby Atmos"]
    }
  ]

POST /api/subscriptions
  Headers: X-User-Id
  Request: {
    "planId": "guid",
    "paymentMethod": {
      "type": "card",
      "cardNumber": "4242424242424242",
      "expiryMonth": 12,
      "expiryYear": 2028,
      "cvv": "123"
    }
  }
  Response: 201 {
    "subscriptionId": "guid",
    "planName": "Standard",
    "status": "Active",
    "currentPeriodStart": "2026-02-13T00:00:00Z",
    "currentPeriodEnd": "2026-03-15T00:00:00Z",
    "payment": {
      "amount": 79.99,
      "currency": "TRY",
      "status": "Succeeded",
      "transactionId": "mock_txn_abc123"
    }
  }
  Error: 402 {
    "error": "Ödeme başarısız",
    "details": "Kart reddedildi: yetersiz bakiye"
  }
  Error: 409 {
    "error": "Aktif aboneliğiniz zaten mevcut"
  }

GET /api/subscriptions/me
  Headers: X-User-Id
  Response: 200 {
    "id": "guid",
    "plan": { "name": "Standard", "tier": "Standard", "priceMonthly": 79.99 },
    "status": "Active",
    "currentPeriodStart": "2026-02-13T00:00:00Z",
    "currentPeriodEnd": "2026-03-15T00:00:00Z",
    "autoRenew": true,
    "daysRemaining": 30
  }
  Error: 404 { "error": "Aktif abonelik bulunamadı" }

PUT /api/subscriptions/me/plan
  Headers: X-User-Id
  Request:  { "newPlanId": "premium-guid" }
  Response: 200 {
    "subscriptionId": "guid",
    "previousPlan": "Standard",
    "newPlan": "Premium",
    "priceDifference": 40.00,
    "effectiveDate": "2026-02-13T00:00:00Z"
  }

POST /api/subscriptions/me/cancel
  Headers: X-User-Id
  Request:  { "reason": "Çok pahalı" }
  Response: 200 {
    "subscriptionId": "guid",
    "status": "Cancelled",
    "activeUntil": "2026-03-15T00:00:00Z",
    "message": "Aboneliğiniz mevcut dönem sonuna kadar aktif kalacaktır"
  }

GET /api/subscriptions/me/invoices
  Headers: X-User-Id
  Response: 200 [
    {
      "id": "guid",
      "invoiceNumber": "INV-2026-000042",
      "amount": 79.99,
      "currency": "TRY",
      "periodStart": "2026-02-13T00:00:00Z",
      "periodEnd": "2026-03-15T00:00:00Z",
      "issuedAt": "2026-02-13T10:00:00Z",
      "status": "Paid"
    }
  ]

GET /health
  Response: 200 { "status": "healthy", "postgres": "connected", "rabbitmq": "connected" }
```

**Veritabanı (PostgreSQL — subscription DB):**

```sql
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    tier VARCHAR(20) NOT NULL,
    price_monthly DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
    max_screens INT NOT NULL,
    max_quality VARCHAR(10) NOT NULL,
    features JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'PendingPayment',
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_active_subscription UNIQUE (user_id) -- 1 aktif abonelik/kullanıcı
);

CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_period_end ON subscriptions(current_period_end);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
    status VARCHAR(20) NOT NULL DEFAULT 'Pending',
    payment_method VARCHAR(50),
    transaction_id VARCHAR(100),
    failure_reason TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_subscription ON payments(subscription_id);

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    payment_id UUID NOT NULL REFERENCES payments(id),
    invoice_number VARCHAR(20) NOT NULL UNIQUE,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE saga_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_type VARCHAR(50) NOT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    current_step VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'InProgress',
    state_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_saga_status ON saga_states(status);
CREATE INDEX idx_saga_correlation ON saga_states(correlation_id);

CREATE TABLE outbox_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    retry_count INT NOT NULL DEFAULT 0,
    error TEXT
);

CREATE INDEX idx_outbox_unprocessed ON outbox_messages(created_at) WHERE processed_at IS NULL;

-- Seed planları
INSERT INTO plans (name, tier, price_monthly, max_screens, max_quality, features) VALUES
  ('Basic', 'Basic', 49.99, 1, '720p', '["Reklamsız izleme", "Mobil indirme"]'),
  ('Standard', 'Standard', 79.99, 2, '1080p', '["Reklamsız izleme", "Mobil indirme", "Full HD"]'),
  ('Premium', 'Premium', 119.99, 4, '4K', '["Reklamsız izleme", "Mobil indirme", "Ultra HD 4K", "Dolby Atmos"]');
```

**Öğrenme Noktaları:**

- Saga pattern (orchestration-based, state machine)
- Outbox pattern (güvenilir event publishing)
- Compensating transactions (geri alma)
- İdempotent consumer tasarımı
- Background service'ler (.NET Hosted Services)
- Mock dış bağımlılıklar (test edilebilir ödeme sistemi)
- Hangfire veya .NET BackgroundService ile zamanlı görevler

---

## 5. Mevcut Servislerdeki Güncellemeler

### 5.1 User Service — Rating ve Event Publishing

```
Yeni Endpoint:

POST /api/users/me/ratings
  Request:  { "contentId": "abc", "rating": 8.5 }
  Response: 201 { "id": "guid", "contentId": "abc", "rating": 8.5 }

GET /api/users/me/ratings
  Response: 200 [ { "contentId": "abc", "rating": 8.5, "ratedAt": "..." } ]

Yeni Event:
  ContentRated → RabbitMQ → Recommendation Engine

Güncelleme:
  User entity'sine subscriptionTier alanı eklenir.
  SubscriptionCreated event geldiğinde tier güncellenir.
```

### 5.2 Catalog Service — Outbox Pattern Eklenmesi

```
Mevcut durum: RabbitMQ'ya direkt publish ediyor.
Güncelleme: Outbox pattern'e geçiş.

1. Her write operasyonunda (create, update, delete):
   BEGIN TRANSACTION
     MongoDB write (içerik)
     Outbox collection'a event yaz
   COMMIT

2. OutboxProcessor (background):
   - Her 5 saniyede outbox'tan unprocessed mesajları al
   - RabbitMQ'ya publish et
   - processedAt güncelle

Not: MongoDB'de multi-document transaction kullanılacak (replica set gerekir).
Alternatif: Change streams ile outbox'a gerek kalmaz, ama öğrenme amaçlı outbox daha iyi.
```

### 5.3 Streaming Service — WatchCompleted Event

```
Mevcut durum: İzleme pozisyonu kaydediyor.
Güncelleme: %90+ izleme tamamlandığında WatchCompleted event publish et.

progress handler'da:
  if positionSeconds / durationSeconds >= 0.90 {
    → RabbitMQ'ya WatchCompleted event publish et
    → Search Service: view_count artır
    → Recommendation Engine: etkileşim kaydet
  }
```

---

## 6. RabbitMQ Topology Güncellemesi

```
Faz 2'den mevcut:
  encoding exchange (topic) → encoding.jobs, encoding.results.catalog

Faz 3'te eklenenler:

  catalog.events exchange (topic, durable)
  ├── routing_key = "content.created"
  │   ├── Queue: search.catalog-sync        → Search Service
  │   └── Queue: recommendation.catalog     → Recommendation Engine
  ├── routing_key = "content.updated"
  │   └── Queue: search.catalog-sync        → Search Service
  └── routing_key = "content.deleted"
      └── Queue: search.catalog-sync        → Search Service

  watch.events exchange (topic, durable)
  ├── routing_key = "watch.completed"
  │   ├── Queue: search.watch-count         → Search Service (view count)
  │   └── Queue: recommendation.watch       → Recommendation Engine
  └── routing_key = "watch.progress"
      └── (Faz 3'te kullanılmaz, ileride analytics)

  user.events exchange (topic, durable)
  └── routing_key = "content.rated"
      └── Queue: recommendation.ratings     → Recommendation Engine

  subscription.events exchange (topic, durable)
  ├── routing_key = "subscription.created"
  │   └── Queue: user.subscription-sync     → User Service (tier güncelle)
  ├── routing_key = "subscription.cancelled"
  │   └── Queue: user.subscription-sync     → User Service
  └── routing_key = "plan.changed"
      └── Queue: user.subscription-sync     → User Service
```

---

## 7. Altyapı Güncellemeleri

### 7.1 Docker Compose Eklentileri

```yaml
  # Faz 1 + Faz 2 servislerinin hepsi aynen kalır, aşağıdakiler eklenir:

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.12.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports: ["9200:9200"]
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:9200/_cluster/health || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 10

  search-service:
    build: ./services/search-service
    ports: ["5005:5005", "50053:50053"]
    environment:
      - ELASTICSEARCH_URL=http://elasticsearch:9200
      - RABBITMQ_URL=amqp://streamvault:secret@rabbitmq:5672/
      - REDIS_URL=redis://redis:6379/3
      - CONSUL_ADDRESS=consul:8500
    depends_on:
      elasticsearch:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
      consul:
        condition: service_started

  recommendation-service:
    build: ./services/recommendation-service
    ports: ["5006:5006", "50054:50054"]
    environment:
      - DATABASE_URL=postgres://streamvault:secret@postgres:5432/streamvault_recommendations
      - RABBITMQ_URL=amqp://streamvault:secret@rabbitmq:5672/
      - REDIS_URL=redis://redis:6379/4
      - CONSUL_ADDRESS=consul:8500
    depends_on:
      postgres:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
      redis:
        condition: service_healthy
      consul:
        condition: service_started

  subscription-service:
    build: ./services/subscription-service
    environment:
      - ASPNETCORE_ENVIRONMENT=Development
      - ConnectionStrings__DefaultConnection=Host=postgres;Database=streamvault_subscriptions;Username=streamvault;Password=secret
      - RabbitMq__Url=amqp://streamvault:secret@rabbitmq:5672/
      - Consul__Address=http://consul:8500
      - ServiceRegistration__Name=subscription-service
      - ServiceRegistration__Port=5007
    depends_on:
      postgres:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
      consul:
        condition: service_started

volumes:
  # ... Faz 1 + Faz 2 volume'ları +
  elasticsearch_data:
```

### 7.2 Güncellenmiş Port Haritası

| Servis | İç Port | Dış Port | Protokol |
|--------|---------|----------|----------|
| ... (Faz 1+2 servisleri) | ... | ... | ... |
| Search Service | 5005 | 5005 | HTTP |
| Search Service | 50053 | 50053 | gRPC |
| Recommendation Service | 5006 | 5006 | HTTP |
| Recommendation Service | 50054 | 50054 | gRPC |
| Subscription Service | 5007 | — | HTTP |
| Elasticsearch | 9200 | 9200 | HTTP |

### 7.3 PostgreSQL — Yeni Veritabanları

```sql
-- infrastructure/postgres/init.sql güncellenmesi
CREATE DATABASE streamvault_users;
CREATE DATABASE streamvault_subscriptions;     -- 🆕
CREATE DATABASE streamvault_recommendations;   -- 🆕
```

---

## 8. API Gateway — Yeni Route'lar

| Method | Path | Downstream | Auth |
|--------|------|------------|------|
| GET | `/api/search` | search-service | ✓ |
| GET | `/api/search/autocomplete` | search-service | ✓ |
| GET | `/api/search/trending` | search-service | ✗ |
| GET | `/api/recommendations` | recommendation-service | ✓ |
| GET | `/api/recommendations/similar/{id}` | recommendation-service | ✓ |
| GET | `/api/recommendations/home` | recommendation-service | ✓ |
| POST | `/api/recommendations/feedback` | recommendation-service | ✓ |
| GET | `/api/plans` | subscription-service | ✗ |
| POST | `/api/subscriptions` | subscription-service | ✓ |
| GET | `/api/subscriptions/me` | subscription-service | ✓ |
| PUT | `/api/subscriptions/me/plan` | subscription-service | ✓ |
| POST | `/api/subscriptions/me/cancel` | subscription-service | ✓ |
| GET | `/api/subscriptions/me/invoices` | subscription-service | ✓ |
| POST | `/api/users/me/ratings` | user-service | ✓ |
| GET | `/api/users/me/ratings` | user-service | ✓ |

---

## 9. Haftalık İlerleme Planı

### Hafta 1: Search Service + Elasticsearch

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Elasticsearch kurulumu | Docker Compose'a ES ekle. Index mapping oluştur (Türkçe analyzer dahil). Kibana veya DevTools ile test. |
| 2 | Search Service — Temel | Go projesi oluştur, ES client, health, Consul kaydı. Basit bir full-text search endpoint. |
| 3 | Search Service — Gelişmiş arama | Faceted search, filtreleme (genre, year, rating), sıralama, sayfalama. Highlight desteği. |
| 4 | Search Service — Autocomplete + Trending | Edge_ngram ile autocomplete. Redis sorted set ile trending listesi. Popüler arama kaydı. |
| 5 | Catalog → Search senkronizasyonu | Catalog'a outbox pattern ekle. Search Service'te RabbitMQ consumer ile ES index senkronizasyonu. Mevcut catalog verisini ES'e toplu indexle. |

### Hafta 2: Recommendation Engine + Subscription Service

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Recommendation — Rust projesi + DB | Cargo projesi, Axum, PostgreSQL migrations, model tanımları. Fake interaction verisi üretme script'i. |
| 2 | Recommendation — Content-based | İçerik feature vektörü oluşturma, cosine similarity, benzer içerik endpoint. Redis feature store. |
| 3 | Recommendation — Collaborative | User-item matrix, kullanıcı benzerliği, puan tahmini. Cold start → popülerlik fallback. Hibrit skor. |
| 4 | Recommendation — Home page + Events | GetHomePageSections gRPC, section üretimi. RabbitMQ consumer (WatchCompleted, ContentRated). |
| 5 | Subscription Service — Domain + Saga | .NET solution, plan seed, CRUD. Saga state machine implementasyonu. Mock payment gateway. |

### Hafta 3: Entegrasyon + Outbox + Test

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Subscription — Outbox + Events | Outbox pattern implementasyonu. SubscriptionCreated/Cancelled event publish. User Service consumer (tier güncelleme). |
| 2 | Subscription — ChangePlan + Cancel + Invoices | Plan değiştirme saga'sı, iptal akışı, fatura oluşturma, otomatik yenileme background job. |
| 3 | User Service güncelleme | Rating endpoint, ContentRated event publish, subscriptionTier alanı, event consumer. |
| 4 | Gateway + Uçtan uca test | Tüm yeni route'ları Gateway'e ekle. Tam akış testi: register → subscribe → izle → rate → search → recommendation. |
| 5 | Test + Dokümantasyon | Saga unit testleri, search testleri, fake veri ile recommendation kalitesi testi. README ve API docs güncelleme. |

---

## 10. Faz 3 Bitiş Kriterleri (Definition of Done)

- [ ] Elasticsearch index oluşturulmuş, Türkçe analyzer çalışıyor
- [ ] Full-text arama sonuçları ilgili ve highlight edilmiş dönüyor
- [ ] Autocomplete 2+ karakterde öneriler sunuyor
- [ ] Faceted arama (genre, yıl, rating) filtreleme ve sayaçları doğru
- [ ] Trend içerikler listesi izleme sayısına göre sıralı
- [ ] Catalog'a yeni içerik eklendiğinde ES index otomatik güncelleniyor (CQRS)
- [ ] Outbox pattern ile güvenilir event publishing çalışıyor
- [ ] Kişisel öneriler kullanıcının izleme geçmişine göre anlamlı sonuçlar veriyor
- [ ] "Benzer içerikler" doğru genre/tag eşleşmeleri gösteriyor
- [ ] Cold start durumunda (yeni kullanıcı) popülerlik bazlı öneriler dönüyor
- [ ] Ana sayfa section'ları (kişisel, trend, genre, yeni) gRPC ile dönüyor
- [ ] 3 abonelik planı (Basic, Standard, Premium) listelenebiliyor
- [ ] Abonelik oluşturma Saga pattern ile çalışıyor (happy path + compensation)
- [ ] Mock ödeme başarısız olduğunda saga compensating action çalışıyor
- [ ] Plan değiştirme (upgrade/downgrade) fiyat farkı hesaplıyor
- [ ] Abonelik iptal edildiğinde dönem sonuna kadar aktif kalıyor
- [ ] Fatura geçmişi görüntülenebiliyor
- [ ] SubscriptionCreated event ile User Service'teki tier güncelleniyor
- [ ] WatchCompleted event hem Search hem Recommendation'a ulaşıyor
- [ ] ContentRated event Recommendation Engine'e ulaşıyor
- [ ] Tüm yeni servisler Consul'da "healthy" görünüyor
- [ ] RabbitMQ'da 4 exchange ve ilgili queue'lar doğru topology ile mevcut

---

## 11. Faz 4'e Hazırlık

Faz 3 tamamlandığında, Faz 4 (Production-Ready) için şu temeller hazır olacak:

- **Event altyapısı olgunlaşmış:** 4 exchange, 10+ queue, event-driven mimari tam oturmuş. Notification Service kolayca bağlanabilir.
- **Outbox pattern yaygınlaşmış:** Catalog ve Subscription'da çalışıyor, diğer servislere de uygulanabilir.
- **gRPC yaygınlaşmış:** 4 servis gRPC konuşuyor. Distributed tracing için gRPC interceptor eklemek kolay.
- **Servis sayısı yeterli:** 7 servis ile gerçek dünya monitoring, tracing ve resilience sorunları gözlemlenebilir.
- **Veri akışı zengin:** İzleme, puanlama, arama, abonelik verileri → Prometheus metrics ve Grafana dashboard'ları için anlamlı metrikler.

> **Not:** Faz 4'ün ana odağı kod yazmak değil, sistemi "production-ready" hale getirmektir: observability (Jaeger, Prometheus, Grafana, ELK), resilience (circuit breaker, retry, bulkhead) ve chaos testing. Bu fazda yeni business logic az, ama operasyonel olgunluk çok yüksek olacak.
