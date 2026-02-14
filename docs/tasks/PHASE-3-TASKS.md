# StreamVault — Faz 3: Smart Features Görevleri

> **Durum:** Faz 1 ve Faz 2 tamamlandı. Faz 3 ile akıllı arama (Elasticsearch), kişiselleştirilmiş öneriler (collaborative filtering), abonelik yönetimi (Saga pattern), CQRS ve Outbox pattern'leri eklenecek.

---

## 1. Altyapı Eklentileri

### 1.1 Elasticsearch
- [x] Docker Compose'a Elasticsearch servisi ekle (`docker.elastic.co/elasticsearch/elasticsearch:8.12.0`)
- [x] `discovery.type=single-node`, `xpack.security.enabled=false` ayarları
- [x] ES_JAVA_OPTS: `-Xms512m -Xmx512m`
- [x] `elasticsearch_data` volume ekle
- [x] Healthcheck çalışıyor (`curl -f http://localhost:9200/_cluster/health`)
- [x] `infrastructure/elasticsearch/init-index.sh` oluştur (elasticsearch-init container ile)
- [x] Index mapping oluştur: `streamvault-content` (Türkçe analyzer dahil)
- [x] Custom analyzer'lar: `turkish_analyzer`, `autocomplete_analyzer`, `autocomplete_search_analyzer`
- [x] Edge ngram filter (min_gram: 2, max_gram: 15) tanımla
- [x] Mapping test: DevTools veya curl ile index'in doğru oluştuğunu doğrula

### 1.2 PostgreSQL — Yeni Veritabanları
- [x] `infrastructure/postgres/init.sql` güncelle: `streamvault_subscriptions` DB ekle
- [x] `infrastructure/postgres/init.sql` güncelle: `streamvault_recommendations` DB ekle
- [x] Her iki DB için de `streamvault` kullanıcısına yetki ver

### 1.3 RabbitMQ Topology Güncellemesi
- [x] `infrastructure/rabbitmq/definitions.json` güncelle
- [x] `catalog.events` exchange ekle (topic, durable)
- [x] Queue: `search.catalog-sync` (routing_key: `content.created`, `content.updated`, `content.deleted`)
- [x] Queue: `recommendation.catalog` (routing_key: `content.created`)
- [x] `watch.events` exchange ekle (topic, durable)
- [x] Queue: `search.watch-count` (routing_key: `watch.completed`)
- [x] Queue: `recommendation.watch` (routing_key: `watch.completed`)
- [x] `user.events` exchange ekle (topic, durable)
- [x] Queue: `recommendation.ratings` (routing_key: `content.rated`)
- [x] `subscription.events` exchange ekle (topic, durable)
- [x] Queue: `user.subscription-sync` (routing_key: `subscription.created`, `subscription.cancelled`, `plan.changed`)
- [x] RabbitMQ Management UI'da tüm exchange ve queue'ların doğru göründüğünü doğrula

### 1.4 Docker Compose Güncellemeleri
- [x] Search Service container tanımı ekle (port: 5005, 50053)
- [x] Recommendation Service container tanımı ekle (port: 5006, 50054)
- [x] Subscription Service container tanımı ekle (port: 5007)
- [x] Servis bağımlılıklarını (depends_on + condition) doğru kur
- [x] Tüm yeni servisler için environment variable'ları ekle
- [x] `docker compose up` ile tüm yeni altyapı ve servislerin ayağa kalktığını doğrula

---

## 2. Proto Dosyaları & gRPC

- [x] `proto/search/v1/search.proto` oluştur (SearchService: Search, Autocomplete, GetTrending)
- [x] SearchRequest/SearchResponse mesajları (filtreleme, sıralama, sayfalama)
- [x] SearchHit, HighlightFields, Facets, FacetBucket mesajları
- [x] AutocompleteRequest/AutocompleteResponse mesajları
- [x] GetTrendingRequest/GetTrendingResponse mesajları
- [x] `proto/recommendation/v1/recommendation.proto` oluştur (RecommendationService: GetRecommendations, GetSimilar, GetHomePageSections)
- [x] GetRecommendationsRequest/Response mesajları (kişisel öneriler)
- [x] GetSimilarRequest/Response mesajları (benzer içerikler)
- [x] GetHomePageSectionsRequest/Response mesajları (ana sayfa section'ları)
- [x] RecommendedItem, SimilarItem, HomePageSection mesajları
- [x] `proto/subscription/v1/subscription.proto` oluştur (SubscriptionQueryService: GetUserSubscription, GetUserTier)
- [x] `google/protobuf/timestamp.proto` import'u, `common/v1/common.proto` bağımlılığı
- [x] `scripts/generate-proto.sh` güncelle — yeni proto dosyaları için Go ve Rust kod üretimi
- [x] Proto üretiminin tüm servisler için çalıştığını doğrula

---

## 3. Search Service (Go)

### 3.1 Proje Kurulumu & Temel Yapı
- [x] Go projesi oluştur (`services/search-service/`)
- [x] `go.mod` oluştur, dependency'leri ekle (elastic/go-elasticsearch, go-redis, amqp091-go, grpc, consul api, viper)
- [x] `internal/config/config.go` — Viper ile config.yaml + env var yükleme
- [x] `cmd/search/main.go` — HTTP + gRPC server bootstrap
- [x] Dockerfile oluştur (multi-stage build)
- [x] Makefile oluştur (build, test, lint)
- [x] `/health` endpoint'i (Elasticsearch, Redis, RabbitMQ bağlantı durumu)

### 3.2 Consul Entegrasyonu
- [x] `internal/discovery/consul.go` — Consul'a self-registration
- [x] Health check kaydı
- [x] Graceful shutdown'da deregistration
- [x] Consul UI'da "search-service" healthy göründüğünü doğrula

### 3.3 Elasticsearch Client & Index
- [x] `internal/elasticsearch/client.go` — ES client wrapper
- [x] `internal/elasticsearch/mapping.go` — Index mapping tanımları (Türkçe analyzer dahil)
- [x] `internal/elasticsearch/analyzer.go` — Custom analyzer konfigürasyonu
- [x] Startup'ta index yoksa otomatik oluşturma
- [x] Index mapping doğrulama

### 3.4 Elasticsearch Indexer
- [x] `internal/elasticsearch/indexer.go` — Document indexleme
- [x] Tek document indexleme (create/update)
- [x] Bulk indexleme (initial sync için)
- [x] Document silme
- [x] `internal/model/search_document.go` — ES'e yazılan document yapısı

### 3.5 Full-Text Arama
- [x] `internal/elasticsearch/searcher.go` — Search query builder
- [x] Multi-match query (title, description, original_title, cast_names)
- [x] Türkçe ve İngilizce analyzer ile çapraz dil arama
- [x] Highlight desteği (title, description alanlarında eşleşen kısımlar vurgulu)
- [x] `internal/handler/search.go` — GET /api/search endpoint
- [x] `internal/model/search_request.go` — Arama isteği modeli
- [x] `internal/model/search_response.go` — Arama sonucu modeli

### 3.6 Faceted Arama & Filtreleme
- [x] `internal/handler/facets.go` — Faceted arama endpoint'i
- [x] Genre filtresi (keyword aggregation)
- [x] Yıl aralığı filtresi (range query)
- [x] Maturity rating filtresi
- [x] Minimum rating filtresi
- [x] Content type filtresi (movie/series)
- [x] Sıralama seçenekleri: relevance, rating, year, title (asc/desc)
- [x] Sayfalama (page, pageSize)
- [x] Facet sayaçları (her filtre değeri için döküman sayısı)

### 3.7 Autocomplete
- [x] `internal/handler/autocomplete.go` — GET /api/search/autocomplete endpoint
- [x] Edge ngram ile prefix matching (title.autocomplete, original_title.autocomplete)
- [x] Fuzzy matching desteği
- [x] Minimum 2 karakter gereksinimi
- [x] Limit parametresi (default: 5)
- [x] Response: id, title, contentType, thumbnailUrl, releaseYear

### 3.8 Trend İçerikler
- [x] `internal/handler/trending.go` — GET /api/search/trending endpoint
- [x] Redis sorted set ile trend listesi (`trending:daily`, `trending:weekly`)
- [x] Zaman penceresi parametresi (day, week, month)
- [x] Rank ve rank değişimi hesaplama
- [x] View count bazlı sıralama

### 3.9 RabbitMQ Consumer — Catalog Sync
- [x] `internal/consumer/catalog_consumer.go` — catalog.events exchange consumer
- [x] `content.created` event → ES'e yeni document indexle
- [x] `content.updated` event → ES'teki document'ı güncelle
- [x] `content.deleted` event → ES'ten document'ı sil
- [x] ACK/NACK mekanizması
- [x] Hata durumunda retry

### 3.10 RabbitMQ Consumer — Watch Count
- [x] `internal/consumer/watch_consumer.go` — watch.events exchange consumer
- [x] `watch.completed` event → view_count artır (ES document update)
- [x] Redis'te günlük view count güncelle (`views:daily:{date}`)
- [x] Trend sorted set güncelle

### 3.11 Redis Cache
- [x] `internal/cache/redis.go` — Redis cache wrapper
- [x] Arama sonucu cache'leme (TTL: 5 dakika, key: `search_cache:{query_hash}`)
- [x] Popüler aramalar kaydı (sorted set: `popular_searches`)
- [x] Trend içerikler cache (TTL: 1 saat)

### 3.12 gRPC Server
- [x] gRPC server implementasyonu (port: 50053)
- [x] Search RPC
- [x] Autocomplete RPC
- [x] GetTrending RPC

### 3.13 Initial Data Sync
- [x] `scripts/seed-search-index.sh` — Mevcut catalog verisini ES'e toplu indexleme script'i
- [x] Catalog Service API'dan tüm içerikleri çek → ES'e bulk index

---

## 4. Recommendation Engine (Go)

### 4.1 Proje Kurulumu & Temel Yapı
- [x] Go projesi oluştur (`services/recommendation-service/`)
- [x] `go.mod` oluştur, dependency'leri ekle (pgx, go-redis, amqp091-go, grpc, consul api, viper)
- [x] `internal/config/config.go` — config.yaml + Viper ile env var yükleme
- [x] `cmd/recommendation/main.go` — HTTP + gRPC server bootstrap, RabbitMQ consumer başlatma
- [x] Dockerfile oluştur (multi-stage build)
- [x] Makefile oluştur (build, test, lint)
- [x] GET /health endpoint (PostgreSQL, Redis, RabbitMQ bağlantı durumu)

### 4.2 Consul Entegrasyonu
- [x] Consul'a self-registration
- [x] Health check kaydı
- [x] Graceful shutdown'da deregistration

### 4.3 PostgreSQL — Migration & Models
- [x] `migrations/001_create_interactions.sql` — interactions tablosu
- [x] `migrations/002_create_user_profiles.sql` — user_profiles tablosu
- [x] `migrations/003_create_content_features.sql` — content_features tablosu
- [x] `migrations/004_create_content_similarity.sql` — content_similarity tablosu
- [x] `internal/model/interaction.go` — Etkileşim modeli (watch, rating, watchlist)
- [x] `internal/model/user_profile.go` — Kullanıcı tercih profili
- [x] `internal/model/content_features.go` — İçerik özellik vektörü
- [x] `internal/model/recommendation.go` — Öneri sonuç yapısı
- [x] `internal/store/postgres.go` — PostgreSQL repository (pgx)

### 4.4 Redis Feature Store
- [x] `internal/store/redis.go` — Redis feature store
- [x] Kullanıcı profil cache (`rec:user_profile:{userId}`, TTL: 1 saat)
- [x] İçerik feature cache (`rec:content_features:{contentId}`, TTL: 6 saat)
- [x] Öneri sonuç cache (`rec:recommendations:{userId}`, TTL: 30 dakika)
- [x] Benzer içerik cache (`rec:similar:{contentId}`, TTL: 6 saat)

### 4.5 Content-Based Filtering
- [x] `internal/engine/content_based.go` — Content-based filtering motoru
- [x] İçerik feature vektörü oluşturma (genres, tags, director, year, rating)
- [x] Kullanıcı preference profili hesaplama (izleme geçmişi ağırlıklı ortalaması)
- [x] `internal/engine/similarity.go` — Cosine similarity hesaplama
- [x] İçerik-içerik benzerlik matrisi oluşturma (top-N sakla)
- [x] Kullanıcı-içerik benzerlik skoru hesaplama

### 4.6 Collaborative Filtering
- [x] `internal/engine/collaborative.go` — User-based collaborative filtering
- [x] User-item matrix oluşturma (izleme + rating verisi)
- [x] Kullanıcı benzerliği hesaplama (cosine similarity, ortak puanlanmış içerikler üzerinden)
- [x] En benzer K kullanıcıyı bulma (K=20)
- [x] Ağırlıklı ortalama ile puan tahmini
- [x] Implicit feedback dönüşümü (izleme tamamlama → 7.0, %50+ → 5.0, watchlist → 6.0)

### 4.7 Hybrid Scorer & Popularity
- [x] `internal/engine/hybrid.go` — Hibrit skor hesaplama
- [x] `final_score = α × collab_score + β × content_score + γ × popularity`
- [x] Ağırlıklar kullanıcı etkileşim sayısına göre dinamik (cold start: γ yüksek, aktif: α yüksek)
- [x] `internal/engine/popularity.go` — Popülerlik bazlı öneri (cold start fallback)
- [x] Cold start eşiği: 5 etkileşimden az → popülerlik bazlı

### 4.8 HTTP API
- [x] `internal/handler/routes.go` — HTTP endpoint tanımları
- [x] `internal/handler/recommendation.go` — Handler implementasyonları
- [x] GET /api/recommendations — Kişisel öneriler (X-User-Id header, limit parametresi)
- [x] GET /api/recommendations/similar/{contentId} — Benzer içerikler (limit parametresi)
- [x] GET /api/recommendations/home — Ana sayfa section'ları (personal, trending, because_you_watched, genre, new)
- [x] POST /api/recommendations/feedback — Geri bildirim (not_interested)
- [x] Response'larda `algorithm` alanı (hybrid, collaborative, popularity)
- [x] Response'larda `reason` alanı ("Interstellar'ı beğendiğiniz için")

### 4.9 gRPC Server
- [x] gRPC server implementasyonu (port: 50054)
- [x] GetRecommendations RPC
- [x] GetSimilar RPC
- [x] GetHomePageSections RPC

### 4.10 RabbitMQ Consumers
- [x] `internal/consumer/watch_consumer.go` — WatchCompleted event consumer
- [x] Etkileşim kaydet (interactions tablosu, type: watch)
- [x] Kullanıcı profili güncelleme tetikle
- [x] `internal/consumer/rating_consumer.go` — ContentRated event consumer
- [x] Etkileşim kaydet (interactions tablosu, type: rating)
- [x] Kullanıcı profili güncelleme tetikle
- [x] `internal/consumer/catalog_consumer.go` — ContentAdded event consumer
- [x] Content features tablosuna yeni içerik ekle
- [x] Feature vektörü hesapla

### 4.11 Batch Hesaplama
- [x] Profil ve benzerlik matrisi periyodik yeniden hesaplama (background task, 6 saat aralık)
- [x] Content similarity matrix güncelleme
- [x] User profile güncelleme

### 4.12 Fake Data Generation
- [x] `scripts/generate-interactions.sh` — Fake izleme/puan verisi üretme script'i
- [x] Mevcut seed kullanıcılar ve catalog içerikleri üzerinden rastgele etkileşimler

---

## 5. Subscription Service (.NET 8)

### 5.1 Proje Kurulumu
- [x] .NET solution oluştur (`services/subscription-service/`)
- [x] 4 proje ekle: `SubscriptionService.Api`, `SubscriptionService.Application`, `SubscriptionService.Domain`, `SubscriptionService.Infrastructure`
- [x] NuGet paketleri: MediatR, FluentValidation, AutoMapper, Npgsql.EntityFrameworkCore, RabbitMQ.Client
- [x] `Program.cs` — DI, MediatR, EF Core, RabbitMQ konfigürasyonu
- [x] `appsettings.json` — Bağlantı bilgileri
- [x] Dockerfile oluştur
- [x] Makefile oluştur

### 5.2 Domain Modelleri
- [x] `Domain/Entities/Plan.cs` — Abonelik planı (Id, Name, Tier, PriceMonthly, MaxScreens, MaxQuality, Features)
- [x] `Domain/Entities/Subscription.cs` — Abonelik (UserId, PlanId, Status, PeriodStart/End, AutoRenew, CancelledAt)
- [x] `Domain/Entities/Payment.cs` — Ödeme (SubscriptionId, Amount, Status, TransactionId, FailureReason)
- [x] `Domain/Entities/Invoice.cs` — Fatura (InvoiceNumber, Amount, PeriodStart/End)
- [x] `Domain/Entities/SagaState.cs` — Saga durumu (SagaType, CurrentStep, Status, StateData)
- [x] `Domain/Entities/OutboxMessage.cs` — Outbox mesajı (EventType, Payload, ProcessedAt, RetryCount)
- [x] `Domain/Enums/PlanTier.cs` — Basic, Standard, Premium
- [x] `Domain/Enums/SubscriptionStatus.cs` — PendingPayment, Active, Cancelled, Expired, Failed
- [x] `Domain/Enums/PaymentStatus.cs` — Pending, Succeeded, Failed, Refunded
- [x] `Domain/Enums/SagaStep.cs` — ValidatePlan, CreateSubscription, ProcessPayment, Activate, CreateInvoice, PublishEvents
- [x] `Domain/Events/` — SubscriptionCreatedEvent, SubscriptionCancelledEvent, PlanChangedEvent, PaymentProcessedEvent
- [x] `Domain/Exceptions/` — PlanNotFoundException, ActiveSubscriptionExistsException, PaymentFailedException

### 5.3 Infrastructure — Persistence
- [x] `Infrastructure/Persistence/SubscriptionDbContext.cs` — EF Core DbContext
- [x] `Infrastructure/Persistence/Configurations/PlanConfiguration.cs`
- [x] `Infrastructure/Persistence/Configurations/SubscriptionConfiguration.cs`
- [x] `Infrastructure/Persistence/Configurations/PaymentConfiguration.cs`
- [x] `Infrastructure/Persistence/Configurations/SagaStateConfiguration.cs`
- [x] `Infrastructure/Persistence/Configurations/OutboxMessageConfiguration.cs`
- [x] EF Core migration oluştur ve uygula
- [x] Seed data: 3 plan (Basic: 49.99 TRY, Standard: 79.99 TRY, Premium: 119.99 TRY)

### 5.4 Infrastructure — Repositories
- [x] `Application/Interfaces/IPlanRepository.cs` interface
- [x] `Application/Interfaces/ISubscriptionRepository.cs` interface
- [x] `Application/Interfaces/IPaymentRepository.cs` interface
- [x] `Application/Interfaces/ISagaRepository.cs` interface
- [x] `Application/Interfaces/IOutboxRepository.cs` interface
- [x] `Application/Interfaces/IPaymentGateway.cs` interface
- [x] `Infrastructure/Persistence/Repositories/PlanRepository.cs`
- [x] `Infrastructure/Persistence/Repositories/SubscriptionRepository.cs`
- [x] `Infrastructure/Persistence/Repositories/PaymentRepository.cs`
- [x] `Infrastructure/Persistence/Repositories/SagaRepository.cs`
- [x] `Infrastructure/Persistence/Repositories/OutboxRepository.cs`

### 5.5 Mock Payment Gateway
- [x] `Infrastructure/Payment/MockPaymentGateway.cs`
- [x] Kart numarasına göre sonuç: `4242...4242` → Başarılı
- [x] `4000...0002` → Reddedildi (insufficient funds)
- [x] `4000...0069` → Süresi dolmuş kart
- [x] `4000...0127` → Genel hata
- [x] Random delay: 200-800ms (gerçekçi latency simülasyonu)

### 5.6 Saga Pattern — Abonelik Oluşturma
- [x] `Application/Sagas/SubscriptionSagaState.cs` — Saga durumları tanımla
- [x] `Application/Sagas/SubscriptionSaga.cs` — Saga state machine
- [x] Step 1: VALIDATE_PLAN — Plan mevcut/aktif mi, kullanıcının aktif aboneliği var mı
- [x] Step 2: CREATE_PENDING_SUBSCRIPTION — Subscription kaydı (Status: PendingPayment) + Outbox
- [x] Step 3: PROCESS_PAYMENT — MockPaymentGateway.ChargeAsync + Payment kaydı
- [x] Step 4: ACTIVATE_SUBSCRIPTION — Status = Active, period ayarla
- [x] Step 5: CREATE_INVOICE — Invoice kaydı oluştur
- [x] Step 6: PUBLISH_EVENTS — SubscriptionCreated event → RabbitMQ
- [x] `Application/Sagas/CompensatingActions.cs` — Geri alma aksiyonları
- [x] Compensate Step 3 başarısız: Subscription sil/Failed yap
- [x] Her adım idempotent olmalı

### 5.7 Saga Pattern — Plan Değiştirme
- [x] ChangePlan Saga: Validate → ProcessPriceDifference → UpdateSubscription → PublishEvents
- [x] Upgrade: Kalan gün için fark ücreti hesapla ve al
- [x] Downgrade: Kalan gün için kredi hesapla
- [x] PlanChanged event publish

### 5.8 Commands & Handlers
- [x] `Application/Commands/CreateSubscription/` — Command, Handler, Validator
- [x] `Application/Commands/CancelSubscription/` — Command, Handler
- [x] `Application/Commands/ChangePlan/` — Command, Handler
- [x] `Application/Commands/ProcessPayment/` — Command, Handler

### 5.9 Queries & Handlers
- [x] `Application/Queries/GetPlans/` — Query, Handler (plan listesi)
- [x] `Application/Queries/GetMySubscription/` — Query, Handler (aktif abonelik)
- [x] `Application/Queries/GetPaymentHistory/` — Query, Handler (fatura geçmişi)
- [x] `Application/DTOs/` — PlanDto, SubscriptionDto, PaymentDto, InvoiceDto
- [x] `Application/Mappings/MappingProfile.cs` — AutoMapper profili

### 5.10 API Controllers
- [x] `Api/Controllers/PlansController.cs` — GET /api/plans (public)
- [x] `Api/Controllers/SubscriptionsController.cs` — POST /api/subscriptions, GET /api/subscriptions/me, PUT /api/subscriptions/me/plan, POST /api/subscriptions/me/cancel, GET /api/subscriptions/me/invoices
- [x] `Api/Controllers/PaymentsController.cs` — Ödeme bilgileri (opsiyonel)
- [x] X-User-Id header'dan kullanıcı bilgisi alma

### 5.11 Background Services
- [x] `Api/BackgroundServices/SagaOrchestratorService.cs` — Saga state machine çalıştırıcı
- [x] `Api/BackgroundServices/OutboxProcessorService.cs` — Outbox event publishing (her 5 saniye)
- [x] `Api/BackgroundServices/SubscriptionRenewalService.cs` — Otomatik yenileme (period bitiminde)
- [x] Outbox: unprocessed mesajları al → RabbitMQ publish → processedAt güncelle
- [x] Outbox: max 3 retry, sonra error logla

### 5.12 Consul Entegrasyonu
- [x] Consul'a self-registration
- [x] Health check kaydı
- [x] `/health` endpoint (PostgreSQL, RabbitMQ bağlantı durumu)

---

## 6. Mevcut Servis Güncellemeleri

### 6.1 User Service — Rating Endpoint
- [x] `Domain/Entities/ContentRating.cs` — Yeni entity (UserId, ContentId, Rating, RatedAt)
- [x] EF Core migration: content_ratings tablosu ekle
- [x] `Application/Commands/RateContent/RateContentCommand.cs`
- [x] `Application/Commands/RateContent/RateContentHandler.cs`
- [x] POST /api/users/me/ratings endpoint (request: contentId, rating)
- [x] GET /api/users/me/ratings endpoint (kullanıcının tüm puanları)
- [x] `Application/Events/ContentRatedEvent.cs` — RabbitMQ'ya publish (`user.events` exchange, `content.rated`)

### 6.2 User Service — Subscription Tier Güncelleme
- [x] `SubscriptionCreated` event consumer ekle (`user.subscription-sync` queue)
- [x] `SubscriptionCancelled` event consumer ekle
- [x] `PlanChanged` event consumer ekle
- [x] Kullanıcının `Role/Tier` alanını güncelle (Free → Basic/Standard/Premium)

### 6.3 Catalog Service — Outbox Pattern
- [x] `Infrastructure/Outbox/OutboxMessage.cs` — Outbox entity
- [x] `Infrastructure/Outbox/OutboxRepository.cs` — Outbox repository
- [x] `Infrastructure/Outbox/OutboxProcessor.cs` — Background service
- [x] Her write operasyonunda (create, update, delete) outbox collection'a event yaz
- [x] MongoDB multi-document transaction kullanımı (replica set)
- [x] OutboxProcessor: her 5 saniyede unprocessed mesajları al → RabbitMQ publish
- [x] Mevcut direkt RabbitMQ publish'i outbox pattern'e geçir
- [x] `catalog.events` exchange'e event publish (content.created, content.updated, content.deleted)

### 6.4 Streaming Service — WatchCompleted Event
- [x] `internal/handler/progress.go` güncelle
- [x] %90+ izleme tamamlandığında WatchCompleted event publish et
- [x] `watch.events` exchange'e publish (routing_key: `watch.completed`)
- [x] Event payload: userId, contentId, completionPercentage, watchedAt
- [x] Aynı içerik için duplicate event engelleme (Redis flag)

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
8. **Recommendation — Temel** → Go projesi, PostgreSQL migrations, model tanımları, Redis feature store
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
