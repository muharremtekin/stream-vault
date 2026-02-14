# StreamVault — Faz 3: Smart Features Görevleri

> **Durum:** Faz 1 ve Faz 2 tamamlandı. Faz 3 ile akıllı arama (Elasticsearch), kişiselleştirilmiş öneriler (collaborative filtering), abonelik yönetimi (Saga pattern), CQRS ve Outbox pattern'leri eklenecek.

---

## 1. Altyapı Eklentileri

### 1.1 Elasticsearch
- [ ] Docker Compose'a Elasticsearch servisi ekle (`docker.elastic.co/elasticsearch/elasticsearch:8.12.0`)
- [ ] `discovery.type=single-node`, `xpack.security.enabled=false` ayarları
- [ ] ES_JAVA_OPTS: `-Xms512m -Xmx512m`
- [ ] `elasticsearch_data` volume ekle
- [ ] Healthcheck çalışıyor (`curl -f http://localhost:9200/_cluster/health`)
- [ ] `infrastructure/elasticsearch/elasticsearch.yml` oluştur
- [ ] Index mapping oluştur: `streamvault-content` (Türkçe analyzer dahil)
- [ ] Custom analyzer'lar: `turkish_analyzer`, `autocomplete_analyzer`, `autocomplete_search_analyzer`
- [ ] Edge ngram filter (min_gram: 2, max_gram: 15) tanımla
- [ ] Mapping test: DevTools veya curl ile index'in doğru oluştuğunu doğrula

### 1.2 PostgreSQL — Yeni Veritabanları
- [ ] `infrastructure/postgres/init.sql` güncelle: `streamvault_subscriptions` DB ekle
- [ ] `infrastructure/postgres/init.sql` güncelle: `streamvault_recommendations` DB ekle
- [ ] Her iki DB için de `streamvault` kullanıcısına yetki ver

### 1.3 RabbitMQ Topology Güncellemesi
- [ ] `infrastructure/rabbitmq/definitions.json` güncelle
- [ ] `catalog.events` exchange ekle (topic, durable)
- [ ] Queue: `search.catalog-sync` (routing_key: `content.created`, `content.updated`, `content.deleted`)
- [ ] Queue: `recommendation.catalog` (routing_key: `content.created`)
- [ ] `watch.events` exchange ekle (topic, durable)
- [ ] Queue: `search.watch-count` (routing_key: `watch.completed`)
- [ ] Queue: `recommendation.watch` (routing_key: `watch.completed`)
- [ ] `user.events` exchange ekle (topic, durable)
- [ ] Queue: `recommendation.ratings` (routing_key: `content.rated`)
- [ ] `subscription.events` exchange ekle (topic, durable)
- [ ] Queue: `user.subscription-sync` (routing_key: `subscription.created`, `subscription.cancelled`, `plan.changed`)
- [ ] RabbitMQ Management UI'da tüm exchange ve queue'ların doğru göründüğünü doğrula

### 1.4 Docker Compose Güncellemeleri
- [ ] Search Service container tanımı ekle (port: 5005, 50053)
- [ ] Recommendation Service container tanımı ekle (port: 5006, 50054)
- [ ] Subscription Service container tanımı ekle (port: 5007)
- [ ] Servis bağımlılıklarını (depends_on + condition) doğru kur
- [ ] Tüm yeni servisler için environment variable'ları ekle
- [ ] `docker compose up` ile tüm yeni altyapı ve servislerin ayağa kalktığını doğrula

---

## 2. Proto Dosyaları & gRPC

- [ ] `proto/search/v1/search.proto` oluştur (SearchService: Search, Autocomplete, GetTrending)
- [ ] SearchRequest/SearchResponse mesajları (filtreleme, sıralama, sayfalama)
- [ ] SearchHit, HighlightFields, Facets, FacetBucket mesajları
- [ ] AutocompleteRequest/AutocompleteResponse mesajları
- [ ] GetTrendingRequest/GetTrendingResponse mesajları
- [ ] `proto/recommendation/v1/recommendation.proto` oluştur (RecommendationService: GetRecommendations, GetSimilar, GetHomePageSections)
- [ ] GetRecommendationsRequest/Response mesajları (kişisel öneriler)
- [ ] GetSimilarRequest/Response mesajları (benzer içerikler)
- [ ] GetHomePageSectionsRequest/Response mesajları (ana sayfa section'ları)
- [ ] RecommendedItem, SimilarItem, HomePageSection mesajları
- [ ] `proto/subscription/v1/subscription.proto` oluştur (SubscriptionQueryService: GetUserSubscription, GetUserTier)
- [ ] `google/protobuf/timestamp.proto` import'u, `common/v1/common.proto` bağımlılığı
- [ ] `scripts/generate-proto.sh` güncelle — yeni proto dosyaları için Go ve Rust kod üretimi
- [ ] Proto üretiminin tüm servisler için çalıştığını doğrula

---

## 3. Search Service (Go)

### 3.1 Proje Kurulumu & Temel Yapı
- [ ] Go projesi oluştur (`services/search-service/`)
- [ ] `go.mod` oluştur, dependency'leri ekle (elastic/go-elasticsearch, go-redis, amqp091-go, grpc, consul api, viper)
- [ ] `internal/config/config.go` — Viper ile config.yaml + env var yükleme
- [ ] `cmd/search/main.go` — HTTP + gRPC server bootstrap
- [ ] Dockerfile oluştur (multi-stage build)
- [ ] Makefile oluştur (build, test, lint)
- [ ] `/health` endpoint'i (Elasticsearch, Redis, RabbitMQ bağlantı durumu)

### 3.2 Consul Entegrasyonu
- [ ] `internal/discovery/consul.go` — Consul'a self-registration
- [ ] Health check kaydı
- [ ] Graceful shutdown'da deregistration
- [ ] Consul UI'da "search-service" healthy göründüğünü doğrula

### 3.3 Elasticsearch Client & Index
- [ ] `internal/elasticsearch/client.go` — ES client wrapper
- [ ] `internal/elasticsearch/mapping.go` — Index mapping tanımları (Türkçe analyzer dahil)
- [ ] `internal/elasticsearch/analyzer.go` — Custom analyzer konfigürasyonu
- [ ] Startup'ta index yoksa otomatik oluşturma
- [ ] Index mapping doğrulama

### 3.4 Elasticsearch Indexer
- [ ] `internal/elasticsearch/indexer.go` — Document indexleme
- [ ] Tek document indexleme (create/update)
- [ ] Bulk indexleme (initial sync için)
- [ ] Document silme
- [ ] `internal/model/search_document.go` — ES'e yazılan document yapısı

### 3.5 Full-Text Arama
- [ ] `internal/elasticsearch/searcher.go` — Search query builder
- [ ] Multi-match query (title, description, original_title, cast_names)
- [ ] Türkçe ve İngilizce analyzer ile çapraz dil arama
- [ ] Highlight desteği (title, description alanlarında eşleşen kısımlar vurgulu)
- [ ] `internal/handler/search.go` — GET /api/search endpoint
- [ ] `internal/model/search_request.go` — Arama isteği modeli
- [ ] `internal/model/search_response.go` — Arama sonucu modeli

### 3.6 Faceted Arama & Filtreleme
- [ ] `internal/handler/facets.go` — Faceted arama endpoint'i
- [ ] Genre filtresi (keyword aggregation)
- [ ] Yıl aralığı filtresi (range query)
- [ ] Maturity rating filtresi
- [ ] Minimum rating filtresi
- [ ] Content type filtresi (movie/series)
- [ ] Sıralama seçenekleri: relevance, rating, year, title (asc/desc)
- [ ] Sayfalama (page, pageSize)
- [ ] Facet sayaçları (her filtre değeri için döküman sayısı)

### 3.7 Autocomplete
- [ ] `internal/handler/autocomplete.go` — GET /api/search/autocomplete endpoint
- [ ] Edge ngram ile prefix matching (title.autocomplete, original_title.autocomplete)
- [ ] Fuzzy matching desteği
- [ ] Minimum 2 karakter gereksinimi
- [ ] Limit parametresi (default: 5)
- [ ] Response: id, title, contentType, thumbnailUrl, releaseYear

### 3.8 Trend İçerikler
- [ ] `internal/handler/trending.go` — GET /api/search/trending endpoint
- [ ] Redis sorted set ile trend listesi (`trending:daily`, `trending:weekly`)
- [ ] Zaman penceresi parametresi (day, week, month)
- [ ] Rank ve rank değişimi hesaplama
- [ ] View count bazlı sıralama

### 3.9 RabbitMQ Consumer — Catalog Sync
- [ ] `internal/consumer/catalog_consumer.go` — catalog.events exchange consumer
- [ ] `content.created` event → ES'e yeni document indexle
- [ ] `content.updated` event → ES'teki document'ı güncelle
- [ ] `content.deleted` event → ES'ten document'ı sil
- [ ] ACK/NACK mekanizması
- [ ] Hata durumunda retry

### 3.10 RabbitMQ Consumer — Watch Count
- [ ] `internal/consumer/watch_consumer.go` — watch.events exchange consumer
- [ ] `watch.completed` event → view_count artır (ES document update)
- [ ] Redis'te günlük view count güncelle (`views:daily:{date}`)
- [ ] Trend sorted set güncelle

### 3.11 Redis Cache
- [ ] `internal/cache/redis.go` — Redis cache wrapper
- [ ] Arama sonucu cache'leme (TTL: 5 dakika, key: `search_cache:{query_hash}`)
- [ ] Popüler aramalar kaydı (sorted set: `popular_searches`)
- [ ] Trend içerikler cache (TTL: 1 saat)

### 3.12 gRPC Server
- [ ] gRPC server implementasyonu (port: 50053)
- [ ] Search RPC
- [ ] Autocomplete RPC
- [ ] GetTrending RPC

### 3.13 Initial Data Sync
- [ ] `scripts/seed-search-index.sh` — Mevcut catalog verisini ES'e toplu indexleme script'i
- [ ] Catalog Service API'dan tüm içerikleri çek → ES'e bulk index

---

## 4. Recommendation Engine (Rust)

### 4.1 Proje Kurulumu & Temel Yapı
- [ ] Cargo projesi oluştur (`services/recommendation-service/`)
- [ ] `Cargo.toml` — dependency'ler (axum, tokio, sqlx, lapin, redis, tonic, serde, tracing)
- [ ] `src/config.rs` — config.toml + env var yükleme
- [ ] `src/main.rs` — Axum + Tonic bootstrap, RabbitMQ consumer başlatma
- [ ] `src/error.rs` — Hata tipleri (thiserror)
- [ ] Dockerfile oluştur (multi-stage build)
- [ ] Makefile oluştur (build, test, lint)
- [ ] GET /health endpoint (PostgreSQL, Redis, RabbitMQ bağlantı durumu)

### 4.2 Consul Entegrasyonu
- [ ] Consul'a self-registration
- [ ] Health check kaydı
- [ ] Graceful shutdown'da deregistration

### 4.3 PostgreSQL — Migration & Models
- [ ] `migrations/001_create_interactions.sql` — interactions tablosu
- [ ] `migrations/002_create_user_profiles.sql` — user_profiles tablosu
- [ ] `migrations/003_create_content_features.sql` — content_features tablosu
- [ ] `migrations/004_create_content_similarity.sql` — content_similarity tablosu
- [ ] `src/model/interaction.rs` — Etkileşim modeli (watch, rating, watchlist)
- [ ] `src/model/user_profile.rs` — Kullanıcı tercih profili
- [ ] `src/model/content_features.rs` — İçerik özellik vektörü
- [ ] `src/model/recommendation.rs` — Öneri sonuç yapısı
- [ ] `src/store/postgres.rs` — PostgreSQL repository (sqlx)

### 4.4 Redis Feature Store
- [ ] `src/store/redis.rs` — Redis feature store
- [ ] Kullanıcı profil cache (`rec:user_profile:{userId}`, TTL: 1 saat)
- [ ] İçerik feature cache (`rec:content_features:{contentId}`, TTL: 6 saat)
- [ ] Öneri sonuç cache (`rec:recommendations:{userId}`, TTL: 30 dakika)
- [ ] Benzer içerik cache (`rec:similar:{contentId}`, TTL: 6 saat)

### 4.5 Content-Based Filtering
- [ ] `src/engine/content_based.rs` — Content-based filtering motoru
- [ ] İçerik feature vektörü oluşturma (genres, tags, director, year, rating)
- [ ] Kullanıcı preference profili hesaplama (izleme geçmişi ağırlıklı ortalaması)
- [ ] `src/engine/similarity.rs` — Cosine similarity hesaplama
- [ ] İçerik-içerik benzerlik matrisi oluşturma (top-N sakla)
- [ ] Kullanıcı-içerik benzerlik skoru hesaplama

### 4.6 Collaborative Filtering
- [ ] `src/engine/collaborative.rs` — User-based collaborative filtering
- [ ] User-item matrix oluşturma (izleme + rating verisi)
- [ ] Kullanıcı benzerliği hesaplama (cosine similarity, ortak puanlanmış içerikler üzerinden)
- [ ] En benzer K kullanıcıyı bulma (K=20)
- [ ] Ağırlıklı ortalama ile puan tahmini
- [ ] Implicit feedback dönüşümü (izleme tamamlama → 7.0, %50+ → 5.0, watchlist → 6.0)

### 4.7 Hybrid Scorer & Popularity
- [ ] `src/engine/hybrid.rs` — Hibrit skor hesaplama
- [ ] `final_score = α × collab_score + β × content_score + γ × popularity`
- [ ] Ağırlıklar kullanıcı etkileşim sayısına göre dinamik (cold start: γ yüksek, aktif: α yüksek)
- [ ] `src/engine/popularity.rs` — Popülerlik bazlı öneri (cold start fallback)
- [ ] Cold start eşiği: 5 etkileşimden az → popülerlik bazlı

### 4.8 HTTP API
- [ ] `src/api/routes.rs` — HTTP endpoint tanımları
- [ ] `src/api/handlers.rs` — Handler implementasyonları
- [ ] GET /api/recommendations — Kişisel öneriler (X-User-Id header, limit parametresi)
- [ ] GET /api/recommendations/similar/{contentId} — Benzer içerikler (limit parametresi)
- [ ] GET /api/recommendations/home — Ana sayfa section'ları (personal, trending, because_you_watched, genre, new)
- [ ] POST /api/recommendations/feedback — Geri bildirim (not_interested)
- [ ] Response'larda `algorithm` alanı (hybrid, collaborative, popularity)
- [ ] Response'larda `reason` alanı ("Interstellar'ı beğendiğiniz için")

### 4.9 gRPC Server
- [ ] gRPC server implementasyonu (port: 50054)
- [ ] GetRecommendations RPC
- [ ] GetSimilar RPC
- [ ] GetHomePageSections RPC

### 4.10 RabbitMQ Consumers
- [ ] `src/consumer/watch_consumer.rs` — WatchCompleted event consumer
- [ ] Etkileşim kaydet (interactions tablosu, type: watch)
- [ ] Kullanıcı profili güncelleme tetikle
- [ ] `src/consumer/rating_consumer.rs` — ContentRated event consumer
- [ ] Etkileşim kaydet (interactions tablosu, type: rating)
- [ ] Kullanıcı profili güncelleme tetikle
- [ ] `src/consumer/catalog_consumer.rs` — ContentAdded event consumer
- [ ] Content features tablosuna yeni içerik ekle
- [ ] Feature vektörü hesapla

### 4.11 Batch Hesaplama
- [ ] Profil ve benzerlik matrisi periyodik yeniden hesaplama (background task, 6 saat aralık)
- [ ] Content similarity matrix güncelleme
- [ ] User profile güncelleme

### 4.12 Fake Data Generation
- [ ] `scripts/generate-interactions.sh` — Fake izleme/puan verisi üretme script'i
- [ ] Mevcut seed kullanıcılar ve catalog içerikleri üzerinden rastgele etkileşimler

---

## 5. Subscription Service (.NET 8)

### 5.1 Proje Kurulumu
- [ ] .NET solution oluştur (`services/subscription-service/`)
- [ ] 4 proje ekle: `SubscriptionService.Api`, `SubscriptionService.Application`, `SubscriptionService.Domain`, `SubscriptionService.Infrastructure`
- [ ] NuGet paketleri: MediatR, FluentValidation, AutoMapper, Npgsql.EntityFrameworkCore, RabbitMQ.Client
- [ ] `Program.cs` — DI, MediatR, EF Core, RabbitMQ konfigürasyonu
- [ ] `appsettings.json` — Bağlantı bilgileri
- [ ] Dockerfile oluştur
- [ ] Makefile oluştur

### 5.2 Domain Modelleri
- [ ] `Domain/Entities/Plan.cs` — Abonelik planı (Id, Name, Tier, PriceMonthly, MaxScreens, MaxQuality, Features)
- [ ] `Domain/Entities/Subscription.cs` — Abonelik (UserId, PlanId, Status, PeriodStart/End, AutoRenew, CancelledAt)
- [ ] `Domain/Entities/Payment.cs` — Ödeme (SubscriptionId, Amount, Status, TransactionId, FailureReason)
- [ ] `Domain/Entities/Invoice.cs` — Fatura (InvoiceNumber, Amount, PeriodStart/End)
- [ ] `Domain/Entities/SagaState.cs` — Saga durumu (SagaType, CurrentStep, Status, StateData)
- [ ] `Domain/Entities/OutboxMessage.cs` — Outbox mesajı (EventType, Payload, ProcessedAt, RetryCount)
- [ ] `Domain/Enums/PlanTier.cs` — Basic, Standard, Premium
- [ ] `Domain/Enums/SubscriptionStatus.cs` — PendingPayment, Active, Cancelled, Expired, Failed
- [ ] `Domain/Enums/PaymentStatus.cs` — Pending, Succeeded, Failed, Refunded
- [ ] `Domain/Enums/SagaStep.cs` — ValidatePlan, CreateSubscription, ProcessPayment, Activate, CreateInvoice, PublishEvents
- [ ] `Domain/Events/` — SubscriptionCreatedEvent, SubscriptionCancelledEvent, PlanChangedEvent, PaymentProcessedEvent
- [ ] `Domain/Exceptions/` — PlanNotFoundException, ActiveSubscriptionExistsException, PaymentFailedException

### 5.3 Infrastructure — Persistence
- [ ] `Infrastructure/Persistence/SubscriptionDbContext.cs` — EF Core DbContext
- [ ] `Infrastructure/Persistence/Configurations/PlanConfiguration.cs`
- [ ] `Infrastructure/Persistence/Configurations/SubscriptionConfiguration.cs`
- [ ] `Infrastructure/Persistence/Configurations/PaymentConfiguration.cs`
- [ ] `Infrastructure/Persistence/Configurations/SagaStateConfiguration.cs`
- [ ] `Infrastructure/Persistence/Configurations/OutboxMessageConfiguration.cs`
- [ ] EF Core migration oluştur ve uygula
- [ ] Seed data: 3 plan (Basic: 49.99 TRY, Standard: 79.99 TRY, Premium: 119.99 TRY)

### 5.4 Infrastructure — Repositories
- [ ] `Application/Interfaces/IPlanRepository.cs` interface
- [ ] `Application/Interfaces/ISubscriptionRepository.cs` interface
- [ ] `Application/Interfaces/IPaymentRepository.cs` interface
- [ ] `Application/Interfaces/ISagaRepository.cs` interface
- [ ] `Application/Interfaces/IOutboxRepository.cs` interface
- [ ] `Application/Interfaces/IPaymentGateway.cs` interface
- [ ] `Infrastructure/Persistence/Repositories/PlanRepository.cs`
- [ ] `Infrastructure/Persistence/Repositories/SubscriptionRepository.cs`
- [ ] `Infrastructure/Persistence/Repositories/PaymentRepository.cs`
- [ ] `Infrastructure/Persistence/Repositories/SagaRepository.cs`
- [ ] `Infrastructure/Persistence/Repositories/OutboxRepository.cs`

### 5.5 Mock Payment Gateway
- [ ] `Infrastructure/Payment/MockPaymentGateway.cs`
- [ ] Kart numarasına göre sonuç: `4242...4242` → Başarılı
- [ ] `4000...0002` → Reddedildi (insufficient funds)
- [ ] `4000...0069` → Süresi dolmuş kart
- [ ] `4000...0127` → Genel hata
- [ ] Random delay: 200-800ms (gerçekçi latency simülasyonu)

### 5.6 Saga Pattern — Abonelik Oluşturma
- [ ] `Application/Sagas/SubscriptionSagaState.cs` — Saga durumları tanımla
- [ ] `Application/Sagas/SubscriptionSaga.cs` — Saga state machine
- [ ] Step 1: VALIDATE_PLAN — Plan mevcut/aktif mi, kullanıcının aktif aboneliği var mı
- [ ] Step 2: CREATE_PENDING_SUBSCRIPTION — Subscription kaydı (Status: PendingPayment) + Outbox
- [ ] Step 3: PROCESS_PAYMENT — MockPaymentGateway.ChargeAsync + Payment kaydı
- [ ] Step 4: ACTIVATE_SUBSCRIPTION — Status = Active, period ayarla
- [ ] Step 5: CREATE_INVOICE — Invoice kaydı oluştur
- [ ] Step 6: PUBLISH_EVENTS — SubscriptionCreated event → RabbitMQ
- [ ] `Application/Sagas/CompensatingActions.cs` — Geri alma aksiyonları
- [ ] Compensate Step 3 başarısız: Subscription sil/Failed yap
- [ ] Her adım idempotent olmalı

### 5.7 Saga Pattern — Plan Değiştirme
- [ ] ChangePlan Saga: Validate → ProcessPriceDifference → UpdateSubscription → PublishEvents
- [ ] Upgrade: Kalan gün için fark ücreti hesapla ve al
- [ ] Downgrade: Kalan gün için kredi hesapla
- [ ] PlanChanged event publish

### 5.8 Commands & Handlers
- [ ] `Application/Commands/CreateSubscription/` — Command, Handler, Validator
- [ ] `Application/Commands/CancelSubscription/` — Command, Handler
- [ ] `Application/Commands/ChangePlan/` — Command, Handler
- [ ] `Application/Commands/ProcessPayment/` — Command, Handler

### 5.9 Queries & Handlers
- [ ] `Application/Queries/GetPlans/` — Query, Handler (plan listesi)
- [ ] `Application/Queries/GetMySubscription/` — Query, Handler (aktif abonelik)
- [ ] `Application/Queries/GetPaymentHistory/` — Query, Handler (fatura geçmişi)
- [ ] `Application/DTOs/` — PlanDto, SubscriptionDto, PaymentDto, InvoiceDto
- [ ] `Application/Mappings/MappingProfile.cs` — AutoMapper profili

### 5.10 API Controllers
- [ ] `Api/Controllers/PlansController.cs` — GET /api/plans (public)
- [ ] `Api/Controllers/SubscriptionsController.cs` — POST /api/subscriptions, GET /api/subscriptions/me, PUT /api/subscriptions/me/plan, POST /api/subscriptions/me/cancel, GET /api/subscriptions/me/invoices
- [ ] `Api/Controllers/PaymentsController.cs` — Ödeme bilgileri (opsiyonel)
- [ ] X-User-Id header'dan kullanıcı bilgisi alma

### 5.11 Background Services
- [ ] `Api/BackgroundServices/SagaOrchestratorService.cs` — Saga state machine çalıştırıcı
- [ ] `Api/BackgroundServices/OutboxProcessorService.cs` — Outbox event publishing (her 5 saniye)
- [ ] `Api/BackgroundServices/SubscriptionRenewalService.cs` — Otomatik yenileme (period bitiminde)
- [ ] Outbox: unprocessed mesajları al → RabbitMQ publish → processedAt güncelle
- [ ] Outbox: max 3 retry, sonra error logla

### 5.12 Consul Entegrasyonu
- [ ] Consul'a self-registration
- [ ] Health check kaydı
- [ ] `/health` endpoint (PostgreSQL, RabbitMQ bağlantı durumu)

---

## 6. Mevcut Servis Güncellemeleri

### 6.1 User Service — Rating Endpoint
- [ ] `Domain/Entities/ContentRating.cs` — Yeni entity (UserId, ContentId, Rating, RatedAt)
- [ ] EF Core migration: content_ratings tablosu ekle
- [ ] `Application/Commands/RateContent/RateContentCommand.cs`
- [ ] `Application/Commands/RateContent/RateContentHandler.cs`
- [ ] POST /api/users/me/ratings endpoint (request: contentId, rating)
- [ ] GET /api/users/me/ratings endpoint (kullanıcının tüm puanları)
- [ ] `Application/Events/ContentRatedEvent.cs` — RabbitMQ'ya publish (`user.events` exchange, `content.rated`)

### 6.2 User Service — Subscription Tier Güncelleme
- [ ] `SubscriptionCreated` event consumer ekle (`user.subscription-sync` queue)
- [ ] `SubscriptionCancelled` event consumer ekle
- [ ] `PlanChanged` event consumer ekle
- [ ] Kullanıcının `Role/Tier` alanını güncelle (Free → Basic/Standard/Premium)

### 6.3 Catalog Service — Outbox Pattern
- [ ] `Infrastructure/Outbox/OutboxMessage.cs` — Outbox entity
- [ ] `Infrastructure/Outbox/OutboxRepository.cs` — Outbox repository
- [ ] `Infrastructure/Outbox/OutboxProcessor.cs` — Background service
- [ ] Her write operasyonunda (create, update, delete) outbox collection'a event yaz
- [ ] MongoDB multi-document transaction kullanımı (replica set)
- [ ] OutboxProcessor: her 5 saniyede unprocessed mesajları al → RabbitMQ publish
- [ ] Mevcut direkt RabbitMQ publish'i outbox pattern'e geçir
- [ ] `catalog.events` exchange'e event publish (content.created, content.updated, content.deleted)

### 6.4 Streaming Service — WatchCompleted Event
- [ ] `internal/handler/progress.go` güncelle
- [ ] %90+ izleme tamamlandığında WatchCompleted event publish et
- [ ] `watch.events` exchange'e publish (routing_key: `watch.completed`)
- [ ] Event payload: userId, contentId, completionPercentage, watchedAt
- [ ] Aynı içerik için duplicate event engelleme (Redis flag)

---

## 7. API Gateway Güncellemeleri

### 7.1 Search Route'ları
- [ ] GET `/api/search` → search-service (Auth: ✓)
- [ ] GET `/api/search/autocomplete` → search-service (Auth: ✓)
- [ ] GET `/api/search/trending` → search-service (Auth: ✗, public)

### 7.2 Recommendation Route'ları
- [ ] GET `/api/recommendations` → recommendation-service (Auth: ✓)
- [ ] GET `/api/recommendations/similar/{id}` → recommendation-service (Auth: ✓)
- [ ] GET `/api/recommendations/home` → recommendation-service (Auth: ✓)
- [ ] POST `/api/recommendations/feedback` → recommendation-service (Auth: ✓)

### 7.3 Subscription Route'ları
- [ ] GET `/api/plans` → subscription-service (Auth: ✗, public)
- [ ] POST `/api/subscriptions` → subscription-service (Auth: ✓)
- [ ] GET `/api/subscriptions/me` → subscription-service (Auth: ✓)
- [ ] PUT `/api/subscriptions/me/plan` → subscription-service (Auth: ✓)
- [ ] POST `/api/subscriptions/me/cancel` → subscription-service (Auth: ✓)
- [ ] GET `/api/subscriptions/me/invoices` → subscription-service (Auth: ✓)

### 7.4 User Service Yeni Route'ları
- [ ] POST `/api/users/me/ratings` → user-service (Auth: ✓)
- [ ] GET `/api/users/me/ratings` → user-service (Auth: ✓)

### 7.5 Config Güncelleme
- [ ] `gateway/config.yaml` — search-service, recommendation-service, subscription-service tanımları
- [ ] Consul'dan yeni servisleri çözümleme
- [ ] Public route listesini güncelle (`/api/search/trending`, `/api/plans`)

---

## 8. Uçtan Uca Entegrasyon & Test

### 8.1 Search Akışı Testi
- [ ] `scripts/test-search.sh` — Arama test senaryoları
- [ ] Catalog'a içerik ekle → ES index'inde otomatik göründüğünü doğrula (CQRS)
- [ ] Full-text arama sonuçları ilgili ve highlight edilmiş dönüyor
- [ ] Autocomplete 2+ karakterde öneriler sunuyor
- [ ] Faceted arama (genre, yıl, rating) filtreleme ve sayaçları doğru
- [ ] Trend içerikler listesi izleme sayısına göre sıralı
- [ ] Arama sonuç cache'leme çalışıyor (Redis)

### 8.2 Recommendation Akışı Testi
- [ ] Fake etkileşim verisi üret (generate-interactions.sh)
- [ ] Kişisel öneriler kullanıcının izleme geçmişine göre anlamlı sonuçlar veriyor
- [ ] "Benzer içerikler" doğru genre/tag eşleşmeleri gösteriyor
- [ ] Cold start durumunda (yeni kullanıcı) popülerlik bazlı öneriler dönüyor
- [ ] Ana sayfa section'ları (personal, trending, genre, new) doğru dönüyor
- [ ] WatchCompleted event → recommendation engine etkileşim kaydı
- [ ] ContentRated event → recommendation engine etkileşim kaydı

### 8.3 Subscription Akışı Testi
- [ ] `scripts/test-subscription.sh` — Abonelik akışı test script'i
- [ ] 3 plan (Basic, Standard, Premium) listelenebiliyor
- [ ] Abonelik oluşturma Saga ile çalışıyor (happy path: başarılı ödeme → aktif abonelik)
- [ ] Mock ödeme başarısız olduğunda saga compensating action çalışıyor (subscription → Failed)
- [ ] Plan değiştirme (upgrade/downgrade) fiyat farkı hesaplıyor
- [ ] Abonelik iptal edildiğinde dönem sonuna kadar aktif kalıyor
- [ ] Fatura geçmişi görüntülenebiliyor
- [ ] SubscriptionCreated event → User Service tier güncelleniyor

### 8.4 Event Akışı Doğrulama
- [ ] WatchCompleted event hem Search hem Recommendation'a ulaşıyor
- [ ] ContentRated event Recommendation Engine'e ulaşıyor
- [ ] Catalog outbox pattern düzgün çalışıyor (content.created/updated/deleted)
- [ ] Subscription outbox pattern düzgün çalışıyor
- [ ] RabbitMQ'da 4 exchange ve ilgili queue'lar doğru topology ile mevcut

### 8.5 Consul & Altyapı
- [ ] Tüm yeni servisler Consul'da "healthy" görünüyor (search, recommendation, subscription)
- [ ] Elasticsearch cluster health: green/yellow
- [ ] Yeni PostgreSQL veritabanları erişilebilir

### 8.6 Unit & Integration Tests
- [ ] Subscription Service: Saga unit testleri (happy path + compensation)
- [ ] Subscription Service: CreateSubscriptionHandler testleri
- [ ] Search Service: searcher, indexer, consumer testleri
- [ ] Recommendation Engine: similarity, collaborative, hybrid score testleri
- [ ] Gateway: yeni route testleri

### 8.7 Tam Akış Testi
- [ ] Register → Subscribe (Standard plan) → İçerik izle → Rate → Search → Recommendation
- [ ] Tier güncelleme sonrası streaming kalite erişimi doğrula
- [ ] Plan upgrade → fiyat farkı → yeni tier yansıması

---

## 9. Dokümantasyon

- [ ] README.md güncelle: Faz 3 mimari diyagram, yeni servisler, port haritası
- [ ] Search API kullanım örnekleri (curl komutları)
- [ ] Recommendation API kullanım örnekleri
- [ ] Subscription API kullanım örnekleri (Saga akışı dahil)
- [ ] RabbitMQ topology güncellemesi açıklaması (4 exchange, queue'lar)
- [ ] Elasticsearch index yapısı açıklaması
- [ ] Outbox pattern açıklaması

---

## Önerilen Sıralama

1. **Altyapı** → Docker Compose'a Elasticsearch ekle, PostgreSQL'e yeni DB'ler ekle, RabbitMQ topology güncelle
2. **Proto dosyaları** → search, recommendation, subscription proto'larını yaz, kod üretim script'ini güncelle
3. **Search Service — Temel** → Go projesi, config, Consul, health, ES client, index mapping
4. **Search Service — Arama** → Full-text search, faceted arama, filtreleme, sıralama, highlight
5. **Search Service — Autocomplete + Trending** → Edge ngram autocomplete, Redis trend listesi, cache
6. **Catalog Service — Outbox** → Outbox pattern ekle, catalog.events exchange'e event publish
7. **Search Service — Consumer** → Catalog event consumer ile ES index senkronizasyonu, watch count consumer
8. **Recommendation — Temel** → Rust projesi, PostgreSQL migrations, model tanımları, Redis feature store
9. **Recommendation — Content-Based** → Feature vektörü, cosine similarity, benzer içerik endpoint
10. **Recommendation — Collaborative** → User-item matrix, kullanıcı benzerliği, hibrit skor, cold start fallback
11. **Recommendation — Home Page + Events** → Ana sayfa section'ları, RabbitMQ consumer'lar
12. **Subscription Service — Domain + Persistence** → .NET solution, entity'ler, EF Core, migration, plan seed
13. **Subscription Service — Saga** → Saga state machine, mock payment, compensating actions
14. **Subscription Service — Outbox + Events** → Outbox pattern, background service'ler, event publish
15. **Subscription Service — API** → Controller'lar, plan değiştirme, iptal, fatura
16. **User Service — Güncelleme** → Rating endpoint, ContentRated event, subscription tier consumer
17. **Streaming Service — Güncelleme** → WatchCompleted event publish
18. **Gateway — Güncelleme** → Tüm yeni route'ları ekle, config güncelle
19. **Uçtan Uca Test** → Tam akış doğrulama, test script'leri, saga testleri
20. **Dokümantasyon** → README, API örnekleri, mimari diyagram
