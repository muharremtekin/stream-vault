# StreamVault — Faz 4: Production-Ready Görevleri

> **Durum:** Faz 1, 2 ve 3 tamamlandı. Faz 4 ile Notification Service, distributed tracing (OpenTelemetry + Jaeger), metrics (Prometheus + Grafana), centralized logging (Loki + Promtail), resilience pattern'leri (circuit breaker, retry, bulkhead), health check ekosistemi, chaos testing ve load testing eklenecek.

---

## 1. Observability Altyapısı

### 1.1 Jaeger
- [x] Docker Compose'a Jaeger servisi ekle (`jaegertracing/all-in-one:1.54`)
- [x] Port: 16686 (UI), 4317 (OTLP gRPC), 4318 (OTLP HTTP)
- [x] `COLLECTOR_OTLP_ENABLED=true` environment ayarı
- [x] `observability/jaeger/jaeger-config.yml` oluştur
- [ ] Jaeger UI'da (http://localhost:16686) erişim doğrulaması

### 1.2 Prometheus
- [x] Docker Compose'a Prometheus servisi ekle (`prom/prometheus:v2.50.0`)
- [x] Port: 9090
- [x] `observability/prometheus/prometheus.yml` oluştur — scrape konfigürasyonu
- [x] Tüm 9 servis için `scrape_configs` tanımla (gateway:8080, user-service:5001, catalog-service:5002, streaming-service:5003, encoding-service:5004, search-service:5005, recommendation-service:5006, subscription-service:5007, notification-service:5008)
- [x] `--storage.tsdb.retention.time=30d` ayarı
- [x] `prometheus_data` volume ekle
- [ ] Prometheus UI'da (http://localhost:9090/targets) target'ların "UP" göründüğünü doğrula

### 1.3 Grafana
- [x] Docker Compose'a Grafana servisi ekle (`grafana/grafana:10.3.0`)
- [x] Port: 3000, Admin: `admin` / `streamvault`
- [x] `observability/grafana/grafana.ini` oluştur
- [x] `observability/grafana/provisioning/datasources/datasources.yml` — Prometheus + Loki + Jaeger datasource tanımları
- [x] `observability/grafana/provisioning/dashboards/dashboards.yml` — Dashboard auto-provision ayarı
- [x] `grafana_data` volume ekle
- [ ] Grafana UI'da (http://localhost:3000) datasource'ların çalıştığını doğrula

### 1.4 Loki
- [x] Docker Compose'a Loki servisi ekle (`grafana/loki:2.9.0`)
- [x] Port: 3100
- [x] `observability/loki/loki-config.yml` oluştur
- [x] `loki_data` volume ekle
- [ ] Loki API'nin çalıştığını doğrula (`http://localhost:3100/ready`)

### 1.5 Promtail
- [x] Docker Compose'a Promtail servisi ekle (`grafana/promtail:2.9.0`)
- [x] `observability/promtail/promtail-config.yml` oluştur
- [x] Container log toplama konfigürasyonu (`/var/lib/docker/containers` mount)
- [x] Docker socket mount (`/var/run/docker.sock:ro`)
- [x] Pipeline stages: JSON parse (level, service, traceId), label extraction, timestamp parse
- [x] Loki'ye push URL: `http://loki:3100/loki/api/v1/push`
- [ ] Promtail'in Loki'ye log gönderdiğini doğrula

### 1.6 Docker Compose Güncellemeleri
- [x] `docker-compose.yml` — 5 observability servisi ekle (jaeger, prometheus, grafana, loki, promtail)
- [x] `docker-compose.override.yml` — Geliştirme port mapping'leri
- [x] Tüm yeni volume tanımları (`prometheus_data`, `grafana_data`, `loki_data`)
- [x] Servis bağımlılıkları (grafana depends_on: prometheus, loki, jaeger; promtail depends_on: loki)
- [ ] `docker compose up` ile 20 container'ın 2 dakika içinde ayağa kalktığını doğrula

---

## 2. Notification Service (Go)

### 2.1 Proje Kurulumu & Temel Yapı
- [x] Go projesi oluştur (`services/notification-service/`)
- [x] `go.mod` oluştur, dependency'leri ekle (gorilla/websocket, amqp091-go, mongo-driver, consul api, viper, zerolog)
- [x] `internal/config/config.go` — Viper ile config.yaml + env var yükleme (`NOTIFICATION_` prefix)
- [x] `cmd/notification/main.go` — HTTP + WebSocket server bootstrap
- [x] Dockerfile oluştur (multi-stage build)
- [x] Makefile oluştur (build, test, lint)
- [x] `config.yaml` — server (port: 5008), websocket, mongodb, rabbitmq, email, push, consul ayarları
- [x] GET /health endpoint (MongoDB, RabbitMQ bağlantı durumu, wsClients sayısı)

### 2.2 Consul Entegrasyonu
- [x] `internal/discovery/consul.go` — Consul'a self-registration
- [x] Health check kaydı
- [x] Graceful shutdown'da deregistration
- [x] Consul UI'da "notification-service" healthy göründüğünü doğrula

### 2.3 MongoDB Store
- [x] `internal/store/mongo.go` — Bildirim geçmişi store
- [x] `notifications` collection — CRUD işlemleri (create, list by userId, mark as read, mark all as read)
- [x] Index: `{ userId: 1, createdAt: -1 }` (bildirim listeleme)
- [x] Index: `{ userId: 1, read: 1 }` (okunmamış sayacı)
- [x] TTL index: `{ expiresAt: 1 }` ile 90 gün sonra otomatik silme
- [x] `internal/store/preferences.go` — Kullanıcı bildirim tercihleri store
- [x] `notification_preferences` collection — CRUD işlemleri (get, upsert)
- [x] Index: `{ userId: 1 }` (unique)
- [x] Varsayılan tercihler (tüm kanallar açık)

### 2.4 WebSocket Hub & Client
- [x] `internal/websocket/message.go` — WS mesaj yapısı (type, id, category, title, body, icon, action, read, createdAt)
- [x] `internal/websocket/client.go` — Tek client bağlantısı (conn, userId, send channel)
- [x] `internal/websocket/hub.go` — Connection hub (clients map, register/unregister channel, broadcast channel)
- [x] Hub.Run() — goroutine ile register/unregister/broadcast dinleme
- [x] Unicast: belirli userId'ye mesaj gönderme
- [x] Heartbeat: 30 saniyede ping, 10 saniye pong timeout
- [x] Bağlantı kopma durumunda Hub'dan otomatik çıkarma

### 2.5 WebSocket Auth Middleware
- [x] `internal/middleware/ws_auth.go` — WebSocket JWT doğrulama
- [x] Query parameter'dan token okuma (`?token=eyJ...`)
- [x] JWT doğrulama (User Service ile aynı secret/issuer)
- [x] userId extraction ve context'e ekleme
- [x] Geçersiz token durumunda 401 + bağlantı reddi

### 2.6 WebSocket Handler
- [x] `internal/handler/websocket.go` — GET /ws/notifications WebSocket upgrade
- [x] gorilla/websocket Upgrader konfigürasyonu (buffer size: 1024)
- [x] Auth middleware sonrası Client oluşturma ve Hub'a register
- [x] ReadPump: client'tan gelen mesajları oku (ping/pong, ack)
- [x] WritePump: send channel'dan mesajları client'a gönder
- [x] Max connections per user: 5

### 2.7 RabbitMQ Consumers
- [x] `internal/consumer/subscription_consumer.go` — `notification.subscription` queue consumer
  - [x] `subscription.created` event → "Hoş geldin" bildirimi (email + inapp)
  - [x] `subscription.cancelled` event → "İptal onay" bildirimi (email + inapp)
  - [x] `plan.changed` event → "Plan değişiklik" bildirimi (email + inapp)
- [x] `internal/consumer/encoding_consumer.go` — `notification.encoding` queue consumer
  - [x] `job.completed` event → "Video hazır" bildirimi (inapp, admin only)
  - [x] `job.failed` event → "Encoding hatası" bildirimi (email + inapp, admin only)
- [x] `internal/consumer/content_consumer.go` — `notification.content` queue consumer
  - [x] `content.created` event → "Yeni içerik" bildirimi (push + inapp)
- [ ] `internal/consumer/recommendation_consumer.go` — Haftalık öneri digest (opsiyonel, cron bazlı)
- [x] Tüm consumer'larda ACK/NACK mekanizması

### 2.8 Dispatcher
- [x] `internal/dispatcher/dispatcher.go` — Event → kanal yönlendirme mantığı
- [x] Kullanıcı tercihlerine göre kanal filtreleme (email kapalı ise email göndermeme)
- [x] Event tipine göre varsayılan kanal haritası (plan'daki tablo)
- [x] `internal/dispatcher/email.go` — SMTP mock sender (MailHog: localhost:1025)
- [x] `internal/dispatcher/push.go` — FCM mock sender (log'a yaz)
- [x] `internal/dispatcher/inapp.go` — WebSocket Hub üzerinden in-app bildirim
- [x] Her kanal için gönderim durumu kaydı (channelStatus)

### 2.9 Template Engine
- [x] `internal/template/engine.go` — Go html/template engine wrapper
- [x] Template yükleme ve cache'leme
- [x] Dinamik veri binding (kullanıcı adı, plan ismi, tutar, içerik başlığı)
- [x] `internal/template/templates/welcome.html`
- [x] `internal/template/templates/subscription_confirmed.html`
- [x] `internal/template/templates/payment_failed.html`
- [x] `internal/template/templates/new_content.html`
- [x] `internal/template/templates/encoding_complete.html`

### 2.10 HTTP API
- [x] `internal/handler/history.go` — Bildirim geçmişi endpoint'leri
  - [x] GET /api/notifications?page=1&pageSize=20&unreadOnly=false — Bildirim listesi (X-User-Id header)
  - [x] POST /api/notifications/{id}/read — Tek bildirim okundu işaretle
  - [x] POST /api/notifications/read-all — Tüm bildirimleri okundu işaretle
- [x] `internal/handler/preferences.go` — Bildirim tercihleri endpoint'leri
  - [x] GET /api/notifications/preferences — Kullanıcının tercihlerini getir
  - [x] PUT /api/notifications/preferences — Tercihleri güncelle
- [x] Response format: items, unreadCount, totalCount, page, pageSize

### 2.11 RabbitMQ Topology Güncellemesi
- [x] `infrastructure/rabbitmq/definitions.json` güncelle
- [x] `notification.subscription` queue ekle (binding: subscription.events exchange → `subscription.created`, `subscription.cancelled`, `plan.changed`)
- [x] `notification.encoding` queue ekle (binding: encoding exchange → `job.completed`, `job.failed`)
- [x] `notification.content` queue ekle (binding: catalog.events exchange → `content.created`)
- [ ] `notification.payment` queue ekle (binding: payment.events exchange → `payment.failed`)
- [x] RabbitMQ Management UI'da yeni queue'ların göründüğünü doğrula

### 2.12 Docker Compose — Notification Service
- [x] `docker-compose.yml` — notification-service container tanımı
- [x] Port: 5008:5008
- [x] Environment: MONGODB_URI, MONGODB_DATABASE, RABBITMQ_URL, CONSUL_ADDRESS, OTEL_EXPORTER_OTLP_ENDPOINT
- [x] depends_on: mongo (service_healthy), rabbitmq (service_healthy)
- [x] Healthcheck konfigürasyonu

---

## 3. Distributed Tracing (OpenTelemetry + Jaeger)

### 3.1 Go Servisleri — OTel Entegrasyonu
- [x] Gateway (`gateway/`) — OpenTelemetry SDK entegrasyonu
  - [x] `go.opentelemetry.io/otel`, `otelhttp`, `otelgrpc` paketleri ekle
  - [x] HTTP handler'ları `otelhttp.NewHandler()` ile wrap
  - [x] HTTP client'ları `otelhttp.NewTransport()` ile wrap
  - [x] OTLP exporter: `http://jaeger:4317`
  - [x] Service name: `gateway`, version: `1.0.0`
  - [ ] Custom span'lar: `gateway.route`, `gateway.auth.validate_jwt`, `gateway.ratelimit.check`, `gateway.proxy.forward`
- [x] Streaming Service (`services/streaming-service/`) — OTel entegrasyonu
  - [x] HTTP handler ve gRPC server interceptor'ları ekle
  - [x] Custom span'lar: segment serve, progress save
  - [x] RabbitMQ mesaj header'larına trace context injection
- [x] Notification Service (`services/notification-service/`) — OTel entegrasyonu
  - [x] HTTP handler ve WebSocket işlemleri için span
  - [x] RabbitMQ consumer'larda trace context extraction
- [x] Search Service — OTel entegrasyonu
  - [x] Elasticsearch ve Redis işlemleri için custom span

### 3.2 Rust Servisleri — OTel Entegrasyonu
- [x] Encoding Service (`services/encoding-service/`) — OpenTelemetry entegrasyonu
  - [x] `opentelemetry`, `opentelemetry-otlp`, `tracing-opentelemetry` crate'leri ekle (`Cargo.toml`)
  - [x] `tracing` subscriber'a OpenTelemetry layer ekle
  - [x] Axum middleware ile otomatik HTTP span oluşturma
  - [x] `#[instrument]` attribute ile pipeline fonksiyonlarında span
  - [x] RabbitMQ mesajlarından trace context extraction (linked span)
  - [x] OTLP exporter: `http://jaeger:4318`, service name: `encoding-service`
- [x] Recommendation Service — aynı OTel entegrasyonu

### 3.3 .NET Servisleri — OTel Entegrasyonu
- [x] User Service (`services/user-service/`) — OpenTelemetry entegrasyonu
  - [x] NuGet: `OpenTelemetry.Extensions.Hosting`, `OpenTelemetry.Instrumentation.AspNetCore`, `OpenTelemetry.Instrumentation.Http`, `OpenTelemetry.Instrumentation.EntityFrameworkCore`, `OpenTelemetry.Exporter.OtlpTrace`
  - [x] `Program.cs` — `AddOpenTelemetry().WithTracing(...)` konfigürasyonu
  - [x] AspNetCore, HttpClient, EF Core automatic instrumentation
  - [x] OTLP exporter: `http://jaeger:4317`, service name: `user-service`
- [x] Catalog Service (`services/catalog-service/`) — aynı OTel entegrasyonu + MongoDB custom span
- [x] Subscription Service — aynı OTel entegrasyonu + EF Core instrumentation

### 3.4 Asenkron Event Trace Propagation
- [x] RabbitMQ mesaj header'larına `traceparent` ve `tracestate` ekleme (publisher tarafında)
- [x] Consumer tarafında header'dan trace context extraction
- [x] Linked span oluşturma (publish span → consume span bağlantısı)
- [x] Go: `propagation.TraceContext{}` ile inject/extract
- [x] Rust: `opentelemetry::global::get_text_map_propagator()` ile inject/extract
- [x] .NET: Activity propagation (otomatik)

### 3.5 Trace Doğrulama
- [x] Jaeger UI'da tüm 8+ servis görünüyor (service list)
- [x] Senkron istek zinciri: Gateway → Catalog Service → MongoDB — tek trace altında
- [x] Asenkron event zinciri: Catalog publish → Search consume — linked span olarak
- [x] gRPC çağrıları trace'te görünüyor (Streaming ↔ Encoding)
- [x] Database sorguları span olarak trace'te yer alıyor (PostgreSQL, MongoDB, Redis, Elasticsearch)
- [x] Span attribute'ları doğru: `http.method`, `http.status_code`, `db.system`, `messaging.system`

---

## 4. Prometheus Metrikleri

### 4.1 Go Servisleri — /metrics Endpoint
- [x] Gateway — Prometheus metrics endpoint (`GET /metrics`)
  - [x] `prometheus/client_golang` paketi ekle
  - [x] RED metrikleri: `http_requests_total`, `http_request_errors_total`, `http_request_duration_seconds` (histogram)
  - [x] Business: `gateway_active_connections`, `gateway_rate_limit_hits_total`, `gateway_circuit_breaker_state{service, state}`
- [x] Streaming Service — Prometheus metrics endpoint
  - [x] RED metrikleri (HTTP + gRPC)
  - [x] Business: `streaming_active_viewers`, `streaming_concurrent_viewers{tier}`, `streaming_segment_serve_duration_seconds{quality}`, `streaming_progress_saves_total`, `streaming_bandwidth_bytes_total{quality}`
- [x] Notification Service — Prometheus metrics endpoint
  - [x] RED metrikleri
  - [x] Business: `notification_sent_total{channel, category}`, `notification_failed_total{channel, reason}`, `notification_ws_active_connections`, `notification_ws_messages_sent_total`
- [x] Search Service — Prometheus metrics endpoint
  - [x] Business: `search_queries_total{type}`, `search_query_duration_seconds`, `search_results_count{type}`, `search_cache_hit_ratio`, `search_index_document_count`

### 4.2 Rust Servisleri — /metrics Endpoint
- [x] Encoding Service — Prometheus metrics endpoint
  - [x] `prometheus` veya `metrics` crate ekle (`Cargo.toml`)
  - [x] RED metrikleri (HTTP + gRPC)
  - [x] Business: `encoding_jobs_total{status}`, `encoding_job_duration_seconds{quality}`, `encoding_active_jobs`, `encoding_queue_depth`, `encoding_ffmpeg_cpu_usage`
- [x] Recommendation Service — Prometheus metrics endpoint
  - [x] Business: `recommendation_requests_total{algorithm}`, `recommendation_computation_duration_seconds`, `recommendation_cache_hit_ratio`, `recommendation_cold_start_fallback_total`

### 4.3 .NET Servisleri — /metrics Endpoint
- [x] User Service — Prometheus metrics endpoint
  - [x] NuGet: `prometheus-net.AspNetCore` veya OpenTelemetry metrics exporter
  - [x] RED metrikleri + USE metrikleri (`dotnet_threadpool_threads_total`, `process_resident_memory_bytes`)
- [x] Catalog Service — aynı metrics entegrasyonu
- [x] Subscription Service — Prometheus metrics endpoint
  - [x] Business: `subscription_active_total{tier}`, `subscription_created_total`, `subscription_cancelled_total`, `payment_processed_total{status}`, `saga_completed_total{saga_type, result}`, `saga_duration_seconds{saga_type}`, `outbox_messages_pending`, `outbox_process_duration_seconds`

### 4.4 Prometheus Doğrulama
- [ ] Prometheus UI'da (`http://localhost:9090/targets`) tüm servisler "UP" görünüyor
- [ ] `http_requests_total` metriği tüm servislerden toplanabiliyor
- [ ] PromQL ile basit sorgu çalışıyor: `rate(http_requests_total[5m])`
- [ ] Histogram metrikleri doğru bucket'larda veri içeriyor

---

## 5. Grafana Dashboard'ları

### 5.1 Services Overview Dashboard
- [x] `observability/grafana/provisioning/dashboards/services-overview.json` oluştur
- [x] Stat panel'ler: Total Requests/s, Error Rate %, Ortalama Latency, Healthy Services sayısı
- [x] Request Rate by Service — Line Chart
- [x] Error Rate by Service — Line Chart
- [x] P50/P95/P99 Latency by Service — Line Chart
- [x] Service Health Matrix — Table (service, status, uptime, error%, P99)

### 5.2 Streaming & Encoding Dashboard
- [x] `observability/grafana/provisioning/dashboards/streaming-dashboard.json` oluştur
- [x] Stat panel'ler: Active Viewers, Segment Latency, Active Encoding Jobs, Queue Depth
- [x] Active Viewers by Quality — Stacked Area Chart
- [x] Segment Serve Duration by Quality — Histogram
- [x] Encoding Job Pipeline — durum bazlı sayaç
- [x] Bandwidth Usage — Line Chart

### 5.3 Business Metrics Dashboard
- [x] `observability/grafana/provisioning/dashboards/business-dashboard.json` oluştur
- [x] Stat panel'ler: Active Subscriptions, MRR (TRY), Search Queries/s, Churn Rate
- [x] Subscriptions by Tier — Pie Chart
- [x] Daily Active Viewers — Line Chart
- [x] Payment Success Rate — Gauge
- [x] Notification Sent by Channel — Stacked Bar Chart

### 5.4 Gateway Dashboard (Opsiyonel)
- [x] `observability/grafana/provisioning/dashboards/gateway-dashboard.json` oluştur
- [x] Route bazlı request rate ve latency
- [x] Rate limit hit sayısı
- [x] Circuit breaker state görselleştirmesi

### 5.5 Alert Kuralları
- [x] Error rate > %5 → alert (en az 1 rule tanımlı)
- [x] P99 latency > 2s → alert
- [x] Servis down → alert
- [ ] Encoding queue depth > 50 → alert (opsiyonel)
- [x] Alert notification channel tanımla (Grafana webhook veya log)

---

## 6. Centralized Logging (Structured JSON)

### 6.1 Go Servisleri — zerolog
- [ ] Gateway — zerolog entegrasyonu
  - [ ] Mevcut log çağrılarını zerolog JSON formatına geçir
  - [ ] Her log entry'de: `timestamp`, `level`, `service`, `traceId`, `spanId`, `message`, `fields`
  - [ ] HTTP request logging middleware (method, path, status, duration)
- [ ] Streaming Service — zerolog entegrasyonu
- [ ] Notification Service — zerolog entegrasyonu
- [ ] Search Service — zerolog entegrasyonu

### 6.2 Rust Servisleri — tracing + JSON
- [ ] Encoding Service — tracing-subscriber JSON formatter
  - [ ] `tracing_subscriber::fmt::layer().json()` konfigürasyonu
  - [ ] trace-id ve span-id log'lara otomatik ekleme (tracing-opentelemetry layer ile)
- [ ] Recommendation Service — aynı JSON logging entegrasyonu

### 6.3 .NET Servisleri — Serilog
- [ ] User Service — Serilog entegrasyonu
  - [ ] NuGet: `Serilog.AspNetCore`, `Serilog.Formatting.Compact`
  - [ ] `Program.cs` — `UseSerilog()` konfigürasyonu, JSON formatter
  - [ ] Enricher: trace-id, span-id, service name
- [ ] Catalog Service — aynı Serilog entegrasyonu
- [ ] Subscription Service — aynı Serilog entegrasyonu

### 6.4 Loki Doğrulama
- [ ] Loki'de `{service="gateway"}` sorgusu çalışıyor
- [ ] `{job="streamvault"} | json | level="error"` ile hata logları filtrelenebiliyor
- [ ] Belirli bir trace-id ile tüm servislerin logları çekiliyor (`| json | traceId="abc123"`)
- [ ] Grafana'da Explore sekmesinde Loki datasource ile log paneli çalışıyor
- [ ] Log'dan Jaeger trace'ine link (traceId field üzerinden)

---

## 7. Resilience Patterns

### 7.1 Circuit Breaker (Gateway)
- [ ] `gateway/internal/middleware/circuit_breaker.go` — Circuit breaker implementasyonu
- [ ] `sony/gobreaker` veya `afex/hystrix-go` kütüphanesi ekle
- [ ] Her downstream servis için ayrı circuit breaker instance
- [ ] Konfigürasyon: `failure_threshold: 5`, `success_threshold: 3`, `timeout: 30s`
- [ ] State'ler: CLOSED → OPEN (hata eşiği aşıldı) → HALF-OPEN (timeout sonrası 1 istek dener)
- [ ] OPEN durumda hızlıca 503 Service Unavailable döndür
- [ ] Fallback response (varsa cache'ten eski veri)
- [ ] `gateway_circuit_breaker_state{service, state}` Prometheus metriği
- [ ] `gateway/config.yaml` — circuit breaker konfigürasyonu (servis başına ayarlar)

### 7.2 Retry Policy (Tüm Servisler)
- [ ] Gateway — HTTP retry policy
  - [ ] `max_retries: 3`, `initial_backoff: 100ms`, `max_backoff: 2s`, `backoff_multiplier: 2.0`
  - [ ] Retryable status codes: 502, 503, 504
  - [ ] Non-retryable: 400, 401, 403, 404, 409
  - [ ] Jitter ekleme (thundering herd önleme)
- [ ] Go servisleri — gRPC retry policy
  - [ ] Retryable codes: UNAVAILABLE, DEADLINE_EXCEEDED
- [ ] Tüm servisler — RabbitMQ consumer retry
  - [ ] Progressive delay: 1s, 5s, 30s
  - [ ] Max 3 başarısız deneme sonrası dead letter queue'ya düşürme
  - [ ] Idempotency key kullanımı (non-idempotent işlemler için)

### 7.3 Timeout Hiyerarşisi
- [ ] Client → Gateway: 30s (genel timeout)
- [ ] Gateway → Downstream Service: 5s (servis çağrısı)
- [ ] Service → Database: 3s (PostgreSQL, MongoDB)
- [ ] Service → Redis: 1s
- [ ] Service → gRPC call: 3s
- [ ] Gateway → Streaming Service: 60s (video chunk için uzun timeout — istisna)
- [ ] Streaming → MinIO: 30s (büyük dosya okuma)
- [ ] Tüm servisler için context timeout propagation
- [ ] İç timeout < dış timeout kuralının sağlanması

### 7.4 Bulkhead Pattern (Gateway)
- [ ] `gateway/internal/middleware/bulkhead.go` — Servis başına connection pool limiti
- [ ] user-service: max_concurrent = 50
- [ ] catalog-service: max_concurrent = 50
- [ ] streaming-service: max_concurrent = 200 (video serving yoğun)
- [ ] search-service: max_concurrent = 100
- [ ] Diğer servisler: max_concurrent = 30
- [ ] Limit aşımında hemen 503 döndür (diğer servisler etkilenmez)
- [ ] `gateway/config.yaml` — bulkhead limitleri konfigürasyonu

### 7.5 Graceful Degradation Stratejileri
- [ ] Search Service çöktü → Catalog Service'e fallback (basit filtreleme)
- [ ] Recommendation çöktü → Popülerlik bazlı statik liste dön (Redis cache)
- [ ] Subscription Service çöktü → Cache'teki son bilinen tier ile devam et
- [ ] Encoding Service çöktü → Upload kabul et, queue'da beklet
- [ ] Redis çöktü → Rate limit devre dışı, progress kaydetme atla
- [ ] MongoDB çöktü → Elasticsearch'ten oku (read model çalışır)
- [ ] Elasticsearch çöktü → Arama devre dışı, diğer özellikler devam
- [ ] RabbitMQ çöktü → Outbox'ta biriktir, RabbitMQ gelince gönder
- [ ] Notification Service çöktü → Core flow etkilenmez, bildirimler kaybolur

---

## 8. Health Check Ekosistemi

### 8.1 Go Servisleri — Health Endpoints
- [ ] Gateway — 3 seviye health check
  - [ ] GET /health/live — Process çalışıyor mu (hızlı, dış bağımlılık yok)
  - [ ] GET /health/ready — Trafik almaya hazır mı (Consul, Redis, downstream servisler)
  - [ ] GET /health/startup — Başlatma tamamlandı mı
  - [ ] GET /health — Detaylı rapor (tüm bağımlılıklar, versiyon, uptime)
- [ ] Streaming Service — 3 seviye health check (MinIO, Redis, RabbitMQ kontrolleri)
- [ ] Notification Service — 3 seviye health check (MongoDB, RabbitMQ, WebSocket hub durumu)
- [ ] Search Service — 3 seviye health check (Elasticsearch, Redis kontrolleri)

### 8.2 Rust Servisleri — Health Endpoints
- [ ] Encoding Service — 3 seviye health check
  - [ ] `/health/live`, `/health/ready`, `/health/startup`
  - [ ] Mevcut `/health` endpoint'ini genişlet: RabbitMQ, MinIO, FFmpeg kontrolleri
- [ ] Recommendation Service — 3 seviye health check (PostgreSQL, Redis kontrolleri)

### 8.3 .NET Servisleri — Health Endpoints
- [ ] User Service — 3 seviye health check
  - [ ] `Microsoft.Extensions.Diagnostics.HealthChecks` + `AspNetCore.HealthChecks.NpgSql` + `AspNetCore.HealthChecks.Redis`
  - [ ] `/health/live`, `/health/ready`, `/health/startup`
  - [ ] PostgreSQL, Redis bağımlılık kontrolü
- [ ] Catalog Service — 3 seviye health check (MongoDB kontrolü)
- [ ] Subscription Service — 3 seviye health check (PostgreSQL, RabbitMQ kontrolü)

### 8.4 Gateway Health Aggregation
- [ ] GET /health (Gateway) — Tüm downstream servislerin `/health/ready` endpoint'ini çağır
- [ ] Sonuçları topla: `healthy` (tüm OK), `degraded` (en az 1 unhealthy), `unhealthy` (kritik down)
- [ ] Timeout: her servis için 2s (yavaş servis tüm raporu engellemesin)
- [ ] Paralel health check (goroutine ile)

### 8.5 Consul Health Check Entegrasyonu
- [ ] Tüm servislerin Consul health check'i `/health/ready` endpoint'ini kullansın
- [ ] Health check interval: 10s
- [ ] Deregister critical service after: 60s
- [ ] Docker Compose healthcheck'leri `/health/live` kullansın

---

## 9. Graceful Shutdown

### 9.1 Go Servisleri
- [ ] Gateway — SIGTERM handling
  - [ ] `signal.NotifyContext` ile shutdown signal yakalama
  - [ ] HTTP server graceful shutdown (`server.Shutdown(ctx)`)
  - [ ] Mevcut isteklerin tamamlanmasını bekle (timeout: 30s)
  - [ ] Consul'dan deregistration
  - [ ] Redis, RabbitMQ bağlantılarını düzgünce kapat
- [ ] Streaming Service — aynı graceful shutdown pattern
- [ ] Notification Service — WebSocket hub'daki client'lara close mesajı gönder + graceful shutdown
- [ ] Search Service — aynı graceful shutdown pattern

### 9.2 Rust Servisleri
- [ ] Encoding Service — `tokio::signal::ctrl_c()` + SIGTERM handling
  - [ ] Aktif encoding job'ların tamamlanmasını bekle (veya cancel et ve requeue)
  - [ ] Consul deregistration, RabbitMQ/MinIO bağlantılarını kapat
- [ ] Recommendation Service — aynı graceful shutdown pattern

### 9.3 .NET Servisleri
- [ ] User Service — `IHostApplicationLifetime.ApplicationStopping` event
  - [ ] Consul deregistration
  - [ ] EF Core DbContext dispose
  - [ ] Background service'lerin düzgün durdurulması
- [ ] Catalog Service — aynı graceful shutdown
- [ ] Subscription Service — Outbox processor'ın mevcut batch'i tamamlaması

---

## 10. Chaos Testing

### 10.1 Chaos Script'leri
- [ ] `resilience/chaos/kill-service.sh` — Rastgele servis öldürme script'i
- [ ] `resilience/chaos/network-delay.sh` — Ağ gecikmesi enjekte etme (tc ile)
- [ ] `resilience/chaos/cpu-stress.sh` — CPU yükü simülasyonu

### 10.2 Senaryo 1: Servis Ölümü (Encoding Service)
- [ ] `resilience/chaos/scenarios/encoding-failure.sh` oluştur
- [ ] Encoding Service'e job gönder (video upload)
- [ ] 5 saniye bekle → encoding-service container'ı durdur (`docker compose stop encoding-service`)
- [ ] Gateway /health degraded olmalı
- [ ] RabbitMQ'da `encoding.jobs` queue'sunda mesaj birikmeli
- [ ] Diğer servisler (catalog, search, streaming) çalışmaya devam etmeli
- [ ] Encoding Service'i geri getir → 30 saniye sonra biriken job'lar tamamlanmış olmalı
- [ ] Sonuçları dokümente et

### 10.3 Senaryo 2: Veritabanı Yavaşlaması
- [ ] `resilience/chaos/scenarios/database-slow.sh` oluştur
- [ ] PostgreSQL'e 500ms network delay ekle (`tc qdisc add dev eth0 root netem delay 500ms 100ms`)
- [ ] Yavaşlama altında 20 istek gönder, response time'ları kaydet
- [ ] Circuit breaker durumunu kontrol et
- [ ] Grafana'da latency spike görünmeli
- [ ] Delay'i kaldır, servisin recovery yaptığını doğrula
- [ ] Sonuçları dokümente et

### 10.4 Senaryo 3: RabbitMQ Kesintisi
- [ ] `resilience/chaos/scenarios/rabbitmq-failure.sh` oluştur
- [ ] Outbox'lu bir işlem yap (abonelik oluştur)
- [ ] RabbitMQ'yu durdur (`docker compose stop rabbitmq`)
- [ ] Event üreten işlem yap (catalog'a film ekle) — core API başarılı olmalı
- [ ] Outbox tablosunda pending mesajlar birikmeli
- [ ] RabbitMQ'yu geri getir → Outbox processor mesajları göndermeli
- [ ] Sonuçları dokümente et

### 10.5 Senaryo 4: Full Chaos (Opsiyonel)
- [ ] `resilience/chaos/scenarios/full-chaos.sh` — Rastgele multi-failure
- [ ] Birden fazla servis ve altyapı bileşeni aynı anda bozulma
- [ ] Sistemin graceful degradation göstermesi
- [ ] Recovery sonrası tüm servislerin healthy duruma dönmesi

---

## 11. Load Testing (k6)

### 11.1 k6 Kurulum & Yapı
- [ ] `resilience/load-test/k6-scripts/` dizini oluştur
- [ ] `resilience/load-test/Makefile` — smoke, load, stress, spike komutları
- [ ] `scripts/run-load-test.sh` — k6 çalıştırma wrapper script'i

### 11.2 Smoke Test
- [ ] `resilience/load-test/k6-scripts/smoke.js` oluştur
- [ ] 5 VU, 1 dakika süreli
- [ ] Threshold: p(95) < 500ms, error rate < %1
- [ ] Senaryo: Register → Login → Catalog listele → Search → Recommendations
- [ ] Tüm endpoint'ler 200 dönüyor

### 11.3 Load Test
- [ ] `resilience/load-test/k6-scripts/load.js` oluştur
- [ ] 50 VU, 5 dakika süreli
- [ ] Normal kullanıcı akışı simülasyonu
- [ ] Threshold: p(95) < 1000ms, error rate < %2

### 11.4 Stress Test
- [ ] `resilience/load-test/k6-scripts/stress.js` oluştur
- [ ] Stages: ramp-up 50 → sabit 50 → artır 100 → sabit 100 → stres 200 → sabit 200 → ramp-down
- [ ] Threshold: p(95) < 2000ms, error rate < %5
- [ ] Grafana'da yük altındaki metrikleri gözlemle
- [ ] Darboğazları belirle ve dokümente et

### 11.5 Spike Test
- [ ] `resilience/load-test/k6-scripts/spike.js` oluştur
- [ ] Ani yük artışı: 10 → 500 VU
- [ ] Sistem davranışını gözlemle (graceful degradation vs crash)
- [ ] Recovery süresini ölç

---

## 12. API Gateway Güncellemeleri

### 12.1 Notification Route'ları
- [x] GET `/ws/notifications` → notification-service (WebSocket upgrade, Auth: JWT query param)
- [x] GET `/api/notifications` → notification-service (Auth: ✓)
- [x] POST `/api/notifications/{id}/read` → notification-service (Auth: ✓)
- [x] POST `/api/notifications/read-all` → notification-service (Auth: ✓)
- [x] GET `/api/notifications/preferences` → notification-service (Auth: ✓)
- [x] PUT `/api/notifications/preferences` → notification-service (Auth: ✓)

### 12.2 WebSocket Proxy
- [x] Gateway'de WebSocket upgrade proxy implementasyonu
- [x] `/ws/notifications` route'unda HTTP → WS upgrade forwarding
- [x] JWT token'ı query parameter olarak downstream'e ilet

### 12.3 Config Güncellemesi
- [x] `gateway/config.yaml` — notification-service tanımı ekle
- [x] Consul'dan notification-service çözümle
- [ ] Circuit breaker, bulkhead, timeout ayarları notification-service için

---

## 13. Uçtan Uca Entegrasyon & Test

### 13.1 Notification Akışı Testi
- [ ] `scripts/test-notifications.sh` oluştur
- [ ] WebSocket bağlantısı kur (wscat veya curl ile)
- [ ] Abonelik oluştur → WebSocket'ten "Hoş geldin" bildirimi geldiğini doğrula
- [ ] Encoding tamamlandığında admin'e in-app bildirim geldiğini doğrula
- [ ] GET /api/notifications → geçmiş bildirimler dönüyor
- [ ] POST /api/notifications/{id}/read → read: true
- [ ] Tercih güncelle: email kapat → email bildirim gönderilmediğini doğrula

### 13.2 Tracing Akışı Testi
- [ ] Gateway üzerinden catalog isteği gönder → Jaeger'da trace bul
- [ ] Gateway → Catalog → MongoDB span zinciri tek trace altında
- [ ] Video upload → encoding → Jaeger'da asenkron trace (linked span)
- [ ] gRPC çağrısı trace'te görünüyor

### 13.3 Metrics & Dashboard Testi
- [ ] Prometheus'ta tüm servislerden metrik toplandığını doğrula
- [ ] Grafana Services Overview dashboard'unda gerçek veri görünüyor
- [ ] Grafana Streaming dashboard'unda encoding job metrikleri görünüyor
- [ ] En az 1 alert rule çalışıyor

### 13.4 Logging Testi
- [ ] Tüm servisler JSON formatında log yazıyor
- [ ] Loki'de servis bazlı filtreleme çalışıyor
- [ ] trace-id ile servisler arası log korelasyonu çalışıyor
- [ ] Grafana'da log paneli çalışıyor

### 13.5 Resilience Testi
- [ ] Circuit breaker: downstream servisi durdur → 503 dönüyor → servis geri gelince CLOSED'a dönüyor
- [ ] Retry: geçici hatalarda otomatik tekrar çalışıyor
- [ ] Graceful shutdown: `docker compose stop <service>` → mevcut istekler tamamlanıyor → clean exit
- [ ] Health check: live/ready/startup endpoint'leri tüm servislerde çalışıyor

### 13.6 Chaos & Load Test Sonuçları
- [ ] En az 3 chaos senaryosu çalıştırılmış ve sonuçları dokümente edilmiş
- [ ] k6 smoke test geçiyor (tüm endpoint'ler sağlıklı)
- [ ] Stress test altında graceful degradation görünüyor
- [ ] Chaos sonrası otomatik recovery çalışıyor

### 13.7 Consul & Altyapı
- [ ] Tüm servisler Consul'da "healthy" görünüyor (9 uygulama servisi)
- [ ] Observability stack tamamen çalışıyor (Jaeger, Prometheus, Grafana, Loki)
- [ ] `docker compose up` ile 20 container 2 dakika içinde ayağa kalkıyor

---

## 14. Dokümantasyon

- [ ] README.md güncelle: Faz 4 mimari diyagram, observability stack, port haritası (20 container)
- [ ] Notification Service API kullanım örnekleri (curl + wscat komutları)
- [ ] Grafana dashboard erişim bilgileri ve ekran görüntüleri
- [ ] Jaeger UI kullanım rehberi (trace arama, servis grafiği)
- [ ] Loki log sorgulama örnekleri (LogQL)
- [ ] Resilience pattern'leri açıklaması (circuit breaker, retry, timeout, bulkhead)
- [ ] Chaos testing sonuç raporu
- [ ] Load testing sonuç raporu (darboğazlar, optimizasyon önerileri)
- [ ] Troubleshooting rehberi (common issues + çözümleri)
- [ ] Final port haritası: 8 uygulama + 1 notification + 5 data store + 1 message broker + 5 observability + 1 service discovery = 20 container

---

## Önerilen Sıralama

1. **Observability Altyapısı** → Docker Compose'a Jaeger, Prometheus, Grafana, Loki, Promtail ekle, UI'ların çalıştığını doğrula
2. **Notification Service — Temel** → Go projesi, config, Consul, MongoDB store, health endpoint
3. **Notification Service — WebSocket** → Hub/Client mimarisi, JWT auth, heartbeat
4. **Notification Service — Consumer & Dispatcher** → RabbitMQ consumer'lar, event → kanal yönlendirme, template engine
5. **Notification Service — API** → Bildirim listesi, okundu işaretle, tercih yönetimi
6. **RabbitMQ Topology** → Notification queue'ları ve binding'leri definitions.json'a ekle
7. **Gateway — Notification Route'ları** → HTTP + WebSocket proxy, config güncelle
8. **OpenTelemetry — Go Servisleri** → Gateway, Streaming, Search, Notification'a OTel
9. **OpenTelemetry — Rust + .NET** → Encoding, Recommendation, User, Catalog, Subscription'a OTel
10. **Asenkron Trace Propagation** → RabbitMQ mesajlarında trace context injection/extraction
11. **Prometheus Metrikleri** → Tüm servislere /metrics endpoint, RED + business metrikleri
12. **Grafana Dashboard'ları** → 3 dashboard: Overview, Streaming, Business + alert rule'ları
13. **Structured JSON Logging** → Tüm servislerde zerolog/tracing/Serilog + trace-id
14. **Loki Entegrasyonu** → Log sorgulama, Grafana'da log paneli
15. **Resilience — Circuit Breaker** → Gateway'e circuit breaker ekle
16. **Resilience — Retry + Timeout + Bulkhead** → Tüm servisler için policy'ler
17. **Graceful Degradation** → Fallback stratejileri
18. **Health Check Ekosistemi** → live/ready/startup endpoint'leri, Gateway aggregation
19. **Graceful Shutdown** → Tüm servislerde SIGTERM handling, clean exit
20. **Chaos Testing** → 3 senaryo çalıştır, sonuçları dokümente et
21. **Load Testing** → k6 smoke, load, stress test
22. **Dokümantasyon** → README, API örnekleri, mimari diyagram, troubleshooting
