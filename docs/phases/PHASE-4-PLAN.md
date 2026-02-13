# StreamVault — Faz 4: Production-Ready

> **Süre:** ~1-2 Hafta
> **Ön Koşul:** Faz 1, 2 ve 3 tamamlanmış olmalı (7 servis + tüm altyapı çalışır durumda)
> **Hedef:** Notification Service eklenmesi, distributed tracing (Jaeger), metrics (Prometheus + Grafana), centralized logging (Loki), health check ekosistemi, circuit breaker, retry policy ve chaos testing ile sistemin production ortamına hazır hale getirilmesi.
> **Sonuç:** Sistem gözlemlenebilir, dayanıklı ve hata durumlarında kendini toparlayabilen bir hale gelir. Grafana dashboard'larından tüm servislerin durumu, latency, error rate ve throughput takip edilebilir.

---

## 1. Faz 4'te Neler Ekleniyor?

| Bileşen | Dil | Yeni/Güncelleme | Açıklama |
|---------|-----|-----------------|----------|
| Notification Service | Go | 🆕 Yeni | Email, push, in-app bildirimler (WebSocket) |
| Jaeger | — | 🆕 Yeni altyapı | Distributed tracing |
| Prometheus | — | 🆕 Yeni altyapı | Metrics toplama |
| Grafana | — | 🆕 Yeni altyapı | Dashboard ve alerting |
| Loki | — | 🆕 Yeni altyapı | Centralized log aggregation |
| Promtail | — | 🆕 Yeni altyapı | Log shipper (container → Loki) |
| Tüm servisler | Go/Rust/.NET | 🔄 Güncelleme | Tracing, metrics, structured logging, health checks, resilience |

---

## 2. Yeni Proje Yapısı (Faz 1+2+3'e Eklenenler)

```
streamvault/
├── ... (Faz 1 + 2 + 3 yapısı aynen kalır)
│
├── services/
│   ├── ... (mevcut 7 servis)
│   │
│   └── notification-service/               # 🆕 Go — Notification Service
│       ├── cmd/
│       │   └── notification/
│       │       └── main.go
│       ├── internal/
│       │   ├── config/
│       │   │   └── config.go
│       │   ├── handler/
│       │   │   ├── websocket.go            # WebSocket bağlantı yönetimi
│       │   │   ├── preferences.go          # Bildirim tercihleri
│       │   │   └── history.go              # Bildirim geçmişi
│       │   ├── consumer/
│       │   │   ├── subscription_consumer.go # SubscriptionCreated/Cancelled
│       │   │   ├── encoding_consumer.go     # EncodingCompleted/Failed
│       │   │   ├── content_consumer.go      # ContentAdded (yeni içerik)
│       │   │   └── recommendation_consumer.go # Haftalık öneri digest
│       │   ├── dispatcher/
│       │   │   ├── dispatcher.go           # Event → kanal yönlendirme
│       │   │   ├── email.go                # SMTP mock sender
│       │   │   ├── push.go                 # FCM mock sender
│       │   │   └── inapp.go                # WebSocket ile in-app bildirim
│       │   ├── template/
│       │   │   ├── engine.go               # Go template engine
│       │   │   └── templates/
│       │   │       ├── welcome.html
│       │   │       ├── subscription_confirmed.html
│       │   │       ├── payment_failed.html
│       │   │       ├── new_content.html
│       │   │       └── encoding_complete.html
│       │   ├── websocket/
│       │   │   ├── hub.go                  # Connection hub (broadcast/unicast)
│       │   │   ├── client.go               # Tek client bağlantısı
│       │   │   └── message.go              # WS mesaj yapısı
│       │   ├── store/
│       │   │   ├── mongo.go                # Bildirim geçmişi (MongoDB)
│       │   │   └── preferences.go          # Kullanıcı tercihleri
│       │   ├── discovery/
│       │   │   └── consul.go
│       │   └── middleware/
│       │       └── ws_auth.go              # WebSocket JWT doğrulama
│       ├── go.mod
│       ├── Dockerfile
│       └── Makefile
│
├── observability/                           # 🆕 Tüm monitoring altyapısı
│   ├── prometheus/
│   │   └── prometheus.yml                  # Scrape konfigürasyonu
│   ├── grafana/
│   │   ├── provisioning/
│   │   │   ├── datasources/
│   │   │   │   └── datasources.yml        # Prometheus + Loki + Jaeger
│   │   │   └── dashboards/
│   │   │       ├── dashboards.yml         # Dashboard auto-provision
│   │   │       ├── services-overview.json  # Tüm servislerin genel durumu
│   │   │       ├── gateway-dashboard.json  # Gateway detay metrikleri
│   │   │       ├── streaming-dashboard.json # Streaming metrikleri
│   │   │       ├── encoding-dashboard.json  # Encoding pipeline metrikleri
│   │   │       └── business-dashboard.json  # İş metrikleri (abonelik, izleme)
│   │   └── grafana.ini
│   ├── loki/
│   │   └── loki-config.yml
│   ├── promtail/
│   │   └── promtail-config.yml            # Container log toplama
│   └── jaeger/
│       └── jaeger-config.yml
│
├── resilience/                              # 🆕 Resilience test ve config
│   ├── chaos/
│   │   ├── kill-service.sh                # Rastgele servis öldürme
│   │   ├── network-delay.sh               # Ağ gecikmesi enjekte etme
│   │   ├── cpu-stress.sh                  # CPU yükü simülasyonu
│   │   └── scenarios/
│   │       ├── encoding-failure.sh        # Encoding Service çöker
│   │       ├── payment-timeout.sh         # Ödeme servisi timeout
│   │       ├── database-slow.sh           # Veritabanı yavaşlaması
│   │       └── full-chaos.sh              # Rastgele multi-failure
│   └── load-test/
│       ├── k6-scripts/
│       │   ├── smoke.js                   # Temel sağlık testi (5 VU)
│       │   ├── load.js                    # Normal yük (50 VU, 5 dk)
│       │   ├── stress.js                  # Stres testi (200 VU, ramp)
│       │   └── spike.js                   # Ani yük artışı (10→500 VU)
│       └── Makefile
│
└── scripts/
    ├── ... (Faz 1 + 2 + 3'ten)
    ├── test-notifications.sh              # 🆕
    ├── run-chaos.sh                       # 🆕
    └── run-load-test.sh                   # 🆕
```

---

## 3. Notification Service (Go) — Detaylı Tasarım

**Sorumluluklar:**

- RabbitMQ'dan domain event'leri dinler
- Event tipine göre doğru bildirim kanalına yönlendirir (email, push, in-app)
- Template engine ile dinamik bildirim içeriği oluşturur
- WebSocket ile gerçek zamanlı in-app bildirimler
- Bildirim geçmişi (MongoDB)
- Kullanıcı bildirim tercihleri (email kapat, push kapat vs.)

**Event → Bildirim Eşlemesi:**

```
┌─────────────────────────┬──────────┬──────────┬──────────┬──────────────────────────┐
│ Event                   │ Email    │ Push     │ In-App   │ Açıklama                 │
├─────────────────────────┼──────────┼──────────┼──────────┼──────────────────────────┤
│ SubscriptionCreated     │ ✓        │ ✗        │ ✓        │ Hoş geldin + fatura      │
│ SubscriptionCancelled   │ ✓        │ ✗        │ ✓        │ İptal onayı              │
│ PlanChanged             │ ✓        │ ✗        │ ✓        │ Plan değişiklik onayı    │
│ PaymentProcessed        │ ✓        │ ✗        │ ✓        │ Ödeme makbuzu            │
│ PaymentFailed           │ ✓        │ ✓        │ ✓        │ Ödeme hatası uyarısı     │
│ ContentAdded            │ ✗        │ ✓        │ ✓        │ Yeni içerik (genre match)│
│ EncodingCompleted       │ ✗        │ ✗        │ ✓(admin) │ Video hazır              │
│ EncodingFailed          │ ✓(admin) │ ✗        │ ✓(admin) │ Encoding hatası          │
│ WeeklyDigest (cron)     │ ✓        │ ✗        │ ✗        │ Haftalık öneri özeti     │
└─────────────────────────┴──────────┴──────────┴──────────┴──────────────────────────┘
```

**WebSocket Mimarisi:**

```
Client (Browser)
    │
    │ WSS /ws/notifications?token=eyJ...
    ▼
┌────────────────────────────────────────────────────┐
│  Notification Service — WebSocket Hub              │
│                                                    │
│  Hub                                               │
│  ├── clients map[userId]*Client                    │
│  ├── register chan *Client                         │
│  ├── unregister chan *Client                       │
│  └── broadcast chan Message                        │
│                                                    │
│  İş akışı:                                        │
│  1. Client WS bağlantısı açar                     │
│  2. JWT token doğrulanır (query param)             │
│  3. Hub'a register edilir                          │
│  4. RabbitMQ'dan event geldiğinde:                 │
│     → userId'ye göre ilgili client'ı bul          │
│     → Mesajı JSON olarak gönder                   │
│  5. Client disconnect olursa Hub'dan çıkarılır     │
│                                                    │
│  Heartbeat: Her 30 saniyede ping/pong              │
│  Reconnect: Client tarafında exponential backoff   │
└────────────────────────────────────────────────────┘
```

**WebSocket Mesaj Formatı:**

```json
{
  "type": "notification",
  "id": "notif-uuid",
  "category": "subscription",
  "title": "Aboneliğiniz Aktif!",
  "body": "Standard planınız başarıyla aktifleştirildi.",
  "icon": "check-circle",
  "action": {
    "type": "navigate",
    "url": "/account/subscription"
  },
  "read": false,
  "createdAt": "2026-02-13T14:30:00Z"
}
```

**HTTP API Kontratı:**

```
GET /ws/notifications?token=eyJ...
  Protocol: WebSocket upgrade
  Auth: JWT token query parametresinde

GET /api/notifications?page=1&pageSize=20&unreadOnly=false
  Headers: X-User-Id
  Response: 200 {
    "items": [
      {
        "id": "notif-uuid",
        "category": "subscription",
        "title": "Aboneliğiniz Aktif!",
        "body": "Standard planınız başarıyla aktifleştirildi.",
        "icon": "check-circle",
        "read": false,
        "createdAt": "2026-02-13T14:30:00Z"
      }
    ],
    "unreadCount": 3,
    "totalCount": 42,
    "page": 1,
    "pageSize": 20
  }

POST /api/notifications/{id}/read
  Headers: X-User-Id
  Response: 200 { "read": true }

POST /api/notifications/read-all
  Headers: X-User-Id
  Response: 200 { "markedCount": 3 }

GET /api/notifications/preferences
  Headers: X-User-Id
  Response: 200 {
    "email": { "enabled": true, "subscription": true, "newContent": false, "weeklyDigest": true },
    "push":  { "enabled": true, "newContent": true, "paymentFailed": true },
    "inApp": { "enabled": true }
  }

PUT /api/notifications/preferences
  Headers: X-User-Id
  Request: {
    "email": { "enabled": true, "subscription": true, "newContent": true, "weeklyDigest": false },
    "push":  { "enabled": false },
    "inApp": { "enabled": true }
  }
  Response: 200 { "updated": true }

GET /health
  Response: 200 { "status": "healthy", "rabbitmq": "connected", "mongodb": "connected", "wsClients": 12 }
```

**MongoDB Collections:**

```javascript
// notifications collection
{
  _id: ObjectId("..."),
  userId: "user-uuid",
  category: "subscription",       // subscription, content, encoding, payment, recommendation
  title: "Aboneliğiniz Aktif!",
  body: "Standard planınız başarıyla aktifleştirildi.",
  icon: "check-circle",
  action: { type: "navigate", url: "/account/subscription" },
  channels: ["email", "inapp"],   // Hangi kanallardan gönderildi
  channelStatus: {
    email: { sent: true, sentAt: ISODate("...") },
    inapp: { sent: true, sentAt: ISODate("...") }
  },
  read: false,
  readAt: null,
  sourceEvent: "SubscriptionCreated",
  metadata: { planName: "Standard", amount: 79.99 },
  createdAt: ISODate("2026-02-13T14:30:00Z"),
  expiresAt: ISODate("2026-05-13T14:30:00Z")     // TTL index ile 90 gün sonra sil
}

// notification_preferences collection
{
  _id: ObjectId("..."),
  userId: "user-uuid",
  email: { enabled: true, subscription: true, newContent: false, weeklyDigest: true },
  push: { enabled: true, newContent: true, paymentFailed: true },
  inApp: { enabled: true },
  updatedAt: ISODate("...")
}

// Indexes
db.notifications.createIndex({ userId: 1, createdAt: -1 })
db.notifications.createIndex({ userId: 1, read: 1 })
db.notifications.createIndex({ expiresAt: 1 }, { expireAfterSeconds: 0 })    // TTL index
db.notification_preferences.createIndex({ userId: 1 }, { unique: true })
```

**RabbitMQ Bindings:**

```
Notification Service tüm event exchange'lerini dinler:

  subscription.events exchange
  ├── routing_key = "subscription.created"  → notification.subscription queue
  ├── routing_key = "subscription.cancelled" → notification.subscription queue
  └── routing_key = "plan.changed"          → notification.subscription queue

  encoding exchange
  ├── routing_key = "job.completed"         → notification.encoding queue
  └── routing_key = "job.failed"            → notification.encoding queue

  catalog.events exchange
  └── routing_key = "content.created"       → notification.content queue

  payment.events exchange (Subscription Service'ten)
  └── routing_key = "payment.failed"        → notification.payment queue
```

**Konfigürasyon (config.yaml):**

```yaml
server:
  http_port: 5008
  read_timeout: 10s
  write_timeout: 10s

websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  ping_interval: 30s
  pong_timeout: 10s
  max_connections_per_user: 5

mongodb:
  uri: "mongodb://mongo:27017"
  database: "streamvault_notifications"
  notification_ttl_days: 90

rabbitmq:
  url: "amqp://streamvault:secret@rabbitmq:5672/"
  queues:
    subscription: "notification.subscription"
    encoding: "notification.encoding"
    content: "notification.content"
    payment: "notification.payment"

email:
  mock: true
  smtp_host: "mailhog:1025"            # MailHog ile mock SMTP
  from: "noreply@streamvault.dev"

push:
  mock: true

consul:
  address: "consul:8500"
  service_name: "notification-service"
  service_port: 5008
```

---

## 4. Distributed Tracing (OpenTelemetry + Jaeger)

### 4.1 Tracing Stratejisi

Her servis OpenTelemetry SDK entegre eder. Trace'ler Jaeger'a gönderilir.

```
Kullanıcı isteği (trace başlangıcı):
    │
    │  trace-id: abc-123-def
    │  span: "GET /api/catalog/movies"
    ▼
┌──────────────┐
│  API Gateway │  span: "gateway.route"
│              │  ├── span: "gateway.auth.validate_jwt"
│              │  ├── span: "gateway.ratelimit.check"
│              │  └── span: "gateway.proxy.forward"
└──────┬───────┘
       │  trace-id header propagate edilir
       ▼
┌──────────────────┐
│  Catalog Service │  span: "catalog.get_movies"
│                  │  ├── span: "catalog.mongodb.find"
│                  │  │   attributes: { db.system: "mongodb", db.operation: "find" }
│                  │  └── span: "catalog.serialize_response"
└──────────────────┘
```

**Asenkron event'lerde trace propagation:**

```
Catalog Service (publisher)
    │  span: "catalog.publish_content_created"
    │  trace-id ve span-id mesaj header'ına eklenir
    ▼
RabbitMQ
    │
    ▼
Search Service (consumer)
    │  span: "search.consume_content_created" (linked span)
    │  ├── span: "search.elasticsearch.index"
    │  └── span: "search.ack_message"
    │
    │  Aynı trace-id ile devam eder → Jaeger'da tek trace altında görünür
```

### 4.2 Her Dil İçin Entegrasyon

**Go Servisleri (Gateway, Streaming, Search, Notification):**

```go
// Kullanılacak paketler:
// go.opentelemetry.io/otel
// go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
// go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
// go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc

// Automatic instrumentation:
// - HTTP handler'lar: otelhttp.NewHandler() ile wrap
// - HTTP client'lar: otelhttp.NewTransport() ile wrap
// - gRPC server: otelgrpc.UnaryServerInterceptor()
// - gRPC client: otelgrpc.UnaryClientInterceptor()

// Manual span:
// ctx, span := tracer.Start(ctx, "custom.operation")
// defer span.End()
// span.SetAttributes(attribute.String("content.id", contentId))
```

**Rust Servisleri (Encoding, Recommendation):**

```rust
// Kullanılacak crate'ler:
// opentelemetry = { version = "0.22", features = ["rt-tokio"] }
// opentelemetry-otlp
// tracing
// tracing-opentelemetry

// Axum middleware ile otomatik span oluşturma
// #[instrument] attribute ile fonksiyon bazlı tracing
// RabbitMQ mesajlarından trace context extraction
```

**.NET Servisleri (User, Catalog, Subscription):**

```csharp
// NuGet paketleri:
// OpenTelemetry.Extensions.Hosting
// OpenTelemetry.Instrumentation.AspNetCore
// OpenTelemetry.Instrumentation.Http
// OpenTelemetry.Instrumentation.EntityFrameworkCore
// OpenTelemetry.Instrumentation.GrpcNetClient
// OpenTelemetry.Exporter.OtlpTrace

// Program.cs'de:
// builder.Services.AddOpenTelemetry()
//   .WithTracing(tracing => tracing
//     .AddAspNetCoreInstrumentation()
//     .AddHttpClientInstrumentation()
//     .AddEntityFrameworkCoreInstrumentation()
//     .AddGrpcClientInstrumentation()
//     .AddOtlpExporter(o => o.Endpoint = new Uri("http://jaeger:4318"))
//   );
```

### 4.3 Trace'te Yakalanacak Bilgiler

```
Her span'de olması gereken attribute'lar:

Genel:
  service.name          = "streaming-service"
  service.version       = "1.0.0"
  deployment.environment = "development"

HTTP:
  http.method           = "GET"
  http.url              = "/api/catalog/movies"
  http.status_code      = 200
  http.request.duration = 45ms
  user.id               = "user-uuid"

Database:
  db.system             = "postgresql" | "mongodb" | "redis" | "elasticsearch"
  db.operation          = "SELECT" | "find" | "GET" | "search"
  db.statement          = "SELECT * FROM users WHERE id = ?"

Messaging:
  messaging.system      = "rabbitmq"
  messaging.operation   = "publish" | "consume"
  messaging.destination = "encoding.jobs"

gRPC:
  rpc.system            = "grpc"
  rpc.service           = "StreamingService"
  rpc.method            = "GetStreamingInfo"

Custom (business logic):
  content.id            = "abc"
  encoding.quality      = "1080p"
  subscription.tier     = "Standard"
  search.query          = "interstellar"
  search.result_count   = 15
```

---

## 5. Metrics (Prometheus + Grafana)

### 5.1 Prometheus Konfigürasyonu

```yaml
# observability/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'gateway'
    static_configs:
      - targets: ['gateway:8080']
    metrics_path: /metrics

  - job_name: 'user-service'
    static_configs:
      - targets: ['user-service:5001']
    metrics_path: /metrics

  - job_name: 'catalog-service'
    static_configs:
      - targets: ['catalog-service:5002']
    metrics_path: /metrics

  - job_name: 'streaming-service'
    static_configs:
      - targets: ['streaming-service:5003']
    metrics_path: /metrics

  - job_name: 'encoding-service'
    static_configs:
      - targets: ['encoding-service:5004']
    metrics_path: /metrics

  - job_name: 'search-service'
    static_configs:
      - targets: ['search-service:5005']
    metrics_path: /metrics

  - job_name: 'recommendation-service'
    static_configs:
      - targets: ['recommendation-service:5006']
    metrics_path: /metrics

  - job_name: 'subscription-service'
    static_configs:
      - targets: ['subscription-service:5007']
    metrics_path: /metrics

  - job_name: 'notification-service'
    static_configs:
      - targets: ['notification-service:5008']
    metrics_path: /metrics
```

### 5.2 Her Servisten Toplanacak Metrikler

**RED Metrikleri (tüm servisler — zorunlu):**

```
# Rate — İstek sayısı
http_requests_total{method, path, status_code, service}
grpc_requests_total{method, service, status}

# Errors — Hata sayısı
http_request_errors_total{method, path, error_type, service}

# Duration — İstek süresi (histogram)
http_request_duration_seconds{method, path, service}
grpc_request_duration_seconds{method, service}
```

**USE Metrikleri (altyapı — zorunlu):**

```
# Utilization
process_cpu_seconds_total
go_goroutines                          # Go servisleri
dotnet_threadpool_threads_total        # .NET servisleri

# Saturation
go_memstats_alloc_bytes
process_resident_memory_bytes

# Errors
db_connection_errors_total{db_system}
rabbitmq_publish_errors_total
```

**Business Metrikleri (servise özel):**

```
# Gateway
gateway_active_connections
gateway_rate_limit_hits_total
gateway_circuit_breaker_state{service, state}       # closed/open/half-open

# Streaming Service
streaming_active_viewers
streaming_concurrent_viewers{tier}
streaming_segment_serve_duration_seconds{quality}
streaming_progress_saves_total
streaming_bandwidth_bytes_total{quality}

# Encoding Service
encoding_jobs_total{status}                          # queued/processing/completed/failed
encoding_job_duration_seconds{quality}
encoding_active_jobs
encoding_queue_depth
encoding_ffmpeg_cpu_usage

# Search Service
search_queries_total{type}                           # fulltext/autocomplete/trending
search_query_duration_seconds
search_results_count{type}
search_cache_hit_ratio
search_index_document_count

# Recommendation Service
recommendation_requests_total{algorithm}             # collaborative/content/hybrid/popularity
recommendation_computation_duration_seconds
recommendation_cache_hit_ratio
recommendation_cold_start_fallback_total

# Subscription Service
subscription_active_total{tier}                      # Basic/Standard/Premium
subscription_created_total
subscription_cancelled_total
subscription_renewed_total
payment_processed_total{status}                      # succeeded/failed
saga_completed_total{saga_type, result}              # success/compensated/failed
saga_duration_seconds{saga_type}
outbox_messages_pending
outbox_process_duration_seconds

# Notification Service
notification_sent_total{channel, category}           # email/push/inapp × subscription/content/...
notification_failed_total{channel, reason}
notification_ws_active_connections
notification_ws_messages_sent_total
```

### 5.3 Grafana Dashboard'ları

**Dashboard 1: Services Overview**

```
┌─────────────────────────────────────────────────────────────────┐
│  StreamVault — Services Overview                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │ Requests │ │  Errors  │ │ Latency  │ │ Services │          │
│  │  1.2K/s  │ │  0.3%    │ │  45ms    │ │  8/8 ✓   │          │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
│                                                                 │
│  [Request Rate by Service — Line Chart]                         │
│  gateway ████████████████████ 1200/s                            │
│  catalog ██████████ 580/s                                       │
│  search  ████████ 420/s                                         │
│  streaming ██████ 310/s                                         │
│  ...                                                            │
│                                                                 │
│  [Error Rate by Service — Line Chart]                           │
│  [P50/P95/P99 Latency by Service — Line Chart]                 │
│                                                                 │
│  [Service Health Matrix — Table]                                │
│  Service          │ Status │ Uptime  │ Error%  │ P99          │
│  gateway          │ ✅     │ 99.99%  │ 0.01%   │ 120ms        │
│  user-service     │ ✅     │ 99.98%  │ 0.05%   │ 85ms         │
│  catalog-service  │ ✅     │ 99.99%  │ 0.02%   │ 95ms         │
│  ...              │        │         │         │              │
└─────────────────────────────────────────────────────────────────┘
```

**Dashboard 2: Streaming & Encoding**

```
┌─────────────────────────────────────────────────────────────────┐
│  StreamVault — Streaming & Encoding                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │ Active   │ │ Segment  │ │ Encoding │ │ Queue    │          │
│  │ Viewers  │ │ Latency  │ │ Jobs     │ │ Depth    │          │
│  │   142    │ │  12ms    │ │  3 ▶️    │ │  7       │          │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
│                                                                 │
│  [Active Viewers by Quality — Stacked Area]                     │
│  [Segment Serve Duration by Quality — Histogram]                │
│  [Encoding Job Pipeline — Gantt style]                          │
│  [Bandwidth Usage — Line Chart]                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Dashboard 3: Business Metrics**

```
┌─────────────────────────────────────────────────────────────────┐
│  StreamVault — Business Metrics                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │ Active   │ │ MRR      │ │ Search   │ │ Churn    │          │
│  │ Subs     │ │ (TRY)    │ │ Queries  │ │ Rate     │          │
│  │  1,247   │ │ 98,540   │ │  420/s   │ │  2.3%    │          │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
│                                                                 │
│  [Subscriptions by Tier — Pie Chart]                            │
│  [Daily Active Viewers — Line Chart]                            │
│  [Top Searched Terms — Table]                                   │
│  [Payment Success Rate — Gauge]                                 │
│  [Recommendation Click-Through Rate — Line Chart]               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Centralized Logging (Loki + Promtail)

### 6.1 Log Formatı Standardizasyonu

Tüm servisler aynı JSON log formatını kullanır:

```json
{
  "timestamp": "2026-02-13T14:30:00.123Z",
  "level": "info",
  "service": "streaming-service",
  "traceId": "abc123def456",
  "spanId": "789ghi",
  "message": "Segment served successfully",
  "fields": {
    "contentId": "abc",
    "quality": "1080p",
    "segmentNumber": 42,
    "durationMs": 12,
    "userId": "user-uuid",
    "clientIp": "192.168.1.100"
  }
}
```

**Log Seviyeleri:**

```
ERROR   — Kullanıcıyı etkileyen hatalar (DB bağlantı kaybı, ödeme hatası)
WARN    — Potansiyel sorunlar (rate limit yaklaşma, retry, degraded mode)
INFO    — Önemli iş olayları (kullanıcı kaydı, abonelik, encoding tamamlanma)
DEBUG   — Detaylı akış (sadece geliştirme ortamında)
```

**Her Dil İçin Kütüphane:**

```
Go:     zerolog (JSON native, yüksek performans)
Rust:   tracing + tracing-subscriber (JSON formatter)
.NET:   Serilog + Serilog.Formatting.Compact (JSON)
```

### 6.2 Promtail Konfigürasyonu

```yaml
# observability/promtail/promtail-config.yml
server:
  http_listen_port: 9080

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: containers
    static_configs:
      - targets: [localhost]
        labels:
          job: streamvault
          __path__: /var/log/containers/*.log

    pipeline_stages:
      - json:
          expressions:
            level: level
            service: service
            traceId: traceId
            message: message
      - labels:
          level:
          service:
      - timestamp:
          source: timestamp
          format: RFC3339Nano
```

### 6.3 Loki'de Faydalı Sorgular

```logql
# Tüm error logları
{job="streamvault"} |= "error" | json | level="error"

# Belirli bir servisin logları
{service="encoding-service"} | json | level="error"

# Belirli bir trace'in tüm logları (servisler arası)
{job="streamvault"} | json | traceId="abc123def456"

# Encoding hatalarının son 1 saat
{service="encoding-service"} | json | level="error" | line_format "{{.message}}"

# Yavaş istekler (>500ms)
{service="gateway"} | json | durationMs > 500

# Ödeme hatalarının sayısı
count_over_time({service="subscription-service"} | json | message=~".*payment.*failed.*" [1h])
```

---

## 7. Resilience Patterns

### 7.1 Circuit Breaker (API Gateway)

```
Gateway'de her downstream servis için ayrı circuit breaker:

             ┌──────────────────────────────────────────────┐
             │           Circuit Breaker State Machine        │
             │                                                │
             │  CLOSED ──── hata eşiği aşıldı ────► OPEN     │
             │    ▲                                    │       │
             │    │                                    │       │
             │    │ başarılı                    timeout sonrası│
             │    │                                    │       │
             │    │                                    ▼       │
             │    └──── başarılı ◄──── HALF-OPEN               │
             │           istek        (1 istek dener)          │
             └──────────────────────────────────────────────┘

Konfigürasyon (servis başına):
  failure_threshold: 5          # 5 ardışık hata → OPEN
  success_threshold: 3          # 3 ardışık başarı → CLOSED
  timeout: 30s                  # OPEN → HALF-OPEN bekleme süresi
  max_concurrent: 100           # Bulkhead: max eşzamanlı istek

OPEN durumda:
  → İstek downstream'e gitmez
  → Hızlıca 503 Service Unavailable döner
  → Fallback response (varsa cache'ten eski veri)

Prometheus metriği:
  gateway_circuit_breaker_state{service="catalog-service", state="open"} 1
```

**Go implementasyonu için:** `sony/gobreaker` veya `afex/hystrix-go` kütüphanesi.

### 7.2 Retry Policy (Tüm Servisler)

```
HTTP İstekleri (servisler arası):
  max_retries: 3
  initial_backoff: 100ms
  max_backoff: 2s
  backoff_multiplier: 2.0
  retryable_status_codes: [502, 503, 504]
  non_retryable_status_codes: [400, 401, 403, 404, 409]

gRPC İstekleri:
  max_retries: 3
  initial_backoff: 100ms
  max_backoff: 1s
  retryable_codes: [UNAVAILABLE, DEADLINE_EXCEEDED]

RabbitMQ Consumer:
  max_retries: 3
  retry_delay: [1s, 5s, 30s]          # Progresif delay
  dead_letter_after: 3 başarısız deneme

İlke:
  - İdempotent istekler → retry güvenli
  - Non-idempotent (ödeme gibi) → retry tehlikeli, idempotency key kullan
  - Jitter ekle (thundering herd önlemek için)
```

### 7.3 Timeout Hiyerarşisi

```
Client → Gateway: 30s (genel timeout)
  Gateway → Downstream Service: 5s (servis çağrısı)
    Service → Database: 3s
    Service → Redis: 1s
    Service → gRPC call: 3s

Kural: İç katman timeout'u her zaman dış katmandan küçük olmalı.
Gateway 5s bekler ama iç servis 10s beklerse → Gateway timeout ama servis çalışmaya devam eder (resource leak).

Streaming Service istisna:
  Gateway → Streaming: 60s (video chunk için uzun timeout)
  Streaming → MinIO: 30s (büyük dosya okuma)
```

### 7.4 Bulkhead Pattern

```
Her downstream servis için ayrı connection pool:

Gateway:
  user-service:     max_concurrent = 50
  catalog-service:  max_concurrent = 50
  streaming-service: max_concurrent = 200  (video serving yoğun)
  search-service:   max_concurrent = 100
  others:           max_concurrent = 30

Amaç: Bir servis yavaşladığında diğerlerinin connection'larını tüketmesini önle.
Eğer catalog-service 50 concurrent'a ulaştıysa, yeni istekler hemen 503 döner
ama search-service etkilenmez.
```

### 7.5 Graceful Degradation Stratejileri

```
┌────────────────────────┬───────────────────────────────────────────────────┐
│ Hata Durumu            │ Degradation Stratejisi                           │
├────────────────────────┼───────────────────────────────────────────────────┤
│ Search Service çöktü   │ Catalog Service'e fallback (basit filtreleme)    │
│ Recommendation çöktü   │ Popülerlik bazlı statik liste dön (Redis cache) │
│ Subscription svc çöktü │ Cache'teki son bilinen tier ile devam et         │
│ Encoding Service çöktü │ Upload kabul et, queue'da beklet                 │
│ Redis çöktü            │ Rate limit devre dışı, progress kaydetme atla   │
│ MongoDB çöktü          │ Catalog: ES'ten oku (read model hâlâ çalışır)   │
│ Elasticsearch çöktü    │ Arama devre dışı, diğer özellikler devam        │
│ RabbitMQ çöktü         │ Outbox'ta biriktir, RabbitMQ gelince gönder     │
│ Notification svc çöktü │ Bildirimler kaybolur ama core flow etkilenmez   │
└────────────────────────┴───────────────────────────────────────────────────┘
```

---

## 8. Health Check Ekosistemi

### 8.1 Health Check Seviyeleri

Her serviste 3 seviye health check:

```
GET /health/live           → Kubernetes liveness probe
  Servis process'i çalışıyor mu?
  Sadece "up" veya "down" döner.
  Hızlı olmalı (<100ms), dış bağımlılık kontrol etmez.

GET /health/ready          → Kubernetes readiness probe
  Servis trafik almaya hazır mı?
  Dış bağımlılıkları kontrol eder (DB, Redis, RabbitMQ).
  Hazır değilse load balancer'dan çıkar.

GET /health/startup        → Kubernetes startup probe
  Servis başlatma tamamlandı mı?
  Migration, seed data, cache ısınma vs.
  Uzun sürebilir, bir kerelik kontrol.

GET /health                → Detaylı sağlık raporu (Consul + monitoring)
  Tüm bağımlılıkların durumu, versiyon bilgisi.
  Response:
  {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "3d 14h 22m",
    "checks": {
      "postgresql": { "status": "healthy", "latencyMs": 2 },
      "redis": { "status": "healthy", "latencyMs": 1 },
      "rabbitmq": { "status": "healthy", "latencyMs": 3 },
      "consul": { "status": "healthy" },
      "disk": { "status": "healthy", "freePercent": 72 }
    }
  }
```

### 8.2 Gateway Health Aggregation

```
GET /health (Gateway)
  → Tüm downstream servislerin /health/ready endpoint'ini çağırır
  → Sonuçları toplar

  Response:
  {
    "status": "degraded",              // healthy | degraded | unhealthy
    "gateway": "healthy",
    "services": {
      "user-service": "healthy",
      "catalog-service": "healthy",
      "streaming-service": "healthy",
      "encoding-service": "unhealthy",   ← Bu yüzden degraded
      "search-service": "healthy",
      "recommendation-service": "healthy",
      "subscription-service": "healthy",
      "notification-service": "healthy"
    }
  }
```

---

## 9. Chaos Testing

### 9.1 Chaos Senaryoları

**Senaryo 1: Servis Ölümü**

```bash
#!/bin/bash
# resilience/chaos/scenarios/encoding-failure.sh

echo "=== Chaos: Encoding Service'i öldürme ==="

echo "1. Encoding Service'e job gönder..."
# Upload video → encoding job başlatılır
curl -X POST http://localhost:8080/api/stream/upload -F "file=@test.mp4" -F "contentId=chaos-test"

echo "2. 5 saniye bekle (encoding başlasın)..."
sleep 5

echo "3. Encoding Service'i öldür..."
docker compose stop encoding-service

echo "4. 30 saniye bekle..."
sleep 30

echo "5. Kontroller:"
echo "   - Gateway /health degraded olmalı"
curl -s http://localhost:8080/health | jq .

echo "   - RabbitMQ'da mesaj birikmeli"
curl -s http://localhost:15672/api/queues/%2F/encoding.jobs -u streamvault:secret | jq .messages

echo "   - Diğer servisler çalışmaya devam etmeli"
curl -s http://localhost:8080/api/catalog/movies | jq .totalCount

echo "6. Encoding Service'i geri getir..."
docker compose start encoding-service

echo "7. 30 saniye bekle (recovery)..."
sleep 30

echo "8. Job tamamlanmış olmalı"
curl -s http://localhost:8080/api/encoding/jobs?status=completed | jq .
echo "=== Senaryo tamamlandı ==="
```

**Senaryo 2: Veritabanı Yavaşlaması**

```bash
#!/bin/bash
# resilience/chaos/scenarios/database-slow.sh

echo "=== Chaos: PostgreSQL yavaşlaması ==="

echo "1. PostgreSQL'e latency ekle (tc ile network delay)..."
docker compose exec postgres tc qdisc add dev eth0 root netem delay 500ms 100ms

echo "2. Yavaşlama altında istek gönder..."
for i in {1..20}; do
  START=$(date +%s%N)
  curl -s http://localhost:8080/api/users/me -H "Authorization: Bearer $TOKEN" > /dev/null
  END=$(date +%s%N)
  DURATION=$(( ($END - $START) / 1000000 ))
  echo "   İstek $i: ${DURATION}ms"
done

echo "3. Circuit breaker durumunu kontrol et..."
curl -s http://localhost:8080/metrics | grep circuit_breaker

echo "4. Grafana'da latency spike görünmeli"
echo "5. Latency'yi kaldır..."
docker compose exec postgres tc qdisc del dev eth0 root

echo "=== Senaryo tamamlandı ==="
```

**Senaryo 3: RabbitMQ Kesintisi**

```bash
#!/bin/bash
# resilience/chaos/scenarios/rabbitmq-failure.sh

echo "=== Chaos: RabbitMQ kesintisi ==="

echo "1. Outbox'lu bir işlem yap (abonelik oluştur)..."
curl -X POST http://localhost:8080/api/subscriptions -H "..." -d '{...}'

echo "2. RabbitMQ'yu durdur..."
docker compose stop rabbitmq

echo "3. Yeni event üreten işlem yap (katalog'a film ekle)..."
curl -X POST http://localhost:8080/api/catalog/movies -H "..." -d '{...}'

echo "4. Kontroller:"
echo "   - Outbox tablosunda pending mesajlar birikmeli"
echo "   - Core API'ler çalışmaya devam etmeli (catalog ekleme başarılı)"
echo "   - Search index güncellenmemeli (eventual consistency)"

echo "5. RabbitMQ'yu geri getir..."
docker compose start rabbitmq

sleep 15

echo "6. Outbox processor mesajları göndermeli..."
echo "   - Search index güncellenmiş olmalı"
echo "   - Notification gönderilmiş olmalı"
echo "=== Senaryo tamamlandı ==="
```

### 9.2 Load Testing (k6)

**Smoke Test — Temel sağlık:**

```javascript
// resilience/load-test/k6-scripts/smoke.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 5,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE = 'http://localhost:8080';

export default function () {
  // Register
  const registerRes = http.post(`${BASE}/api/auth/register`, JSON.stringify({
    email: `loadtest_${__VU}_${__ITER}@test.com`,
    password: 'Test123!',
  }), { headers: { 'Content-Type': 'application/json' } });

  check(registerRes, { 'register 201': (r) => r.status === 201 });

  // Login
  const loginRes = http.post(`${BASE}/api/auth/login`, JSON.stringify({
    email: `loadtest_${__VU}_${__ITER}@test.com`,
    password: 'Test123!',
  }), { headers: { 'Content-Type': 'application/json' } });

  const token = JSON.parse(loginRes.body).accessToken;
  const authHeaders = { headers: { Authorization: `Bearer ${token}` } };

  // Catalog
  const catalogRes = http.get(`${BASE}/api/catalog/movies?page=1&pageSize=10`, authHeaders);
  check(catalogRes, { 'catalog 200': (r) => r.status === 200 });

  // Search
  const searchRes = http.get(`${BASE}/api/search?q=interstellar`, authHeaders);
  check(searchRes, { 'search 200': (r) => r.status === 200 });

  // Recommendations
  const recRes = http.get(`${BASE}/api/recommendations/home`, authHeaders);
  check(recRes, { 'recommendations 200': (r) => r.status === 200 });

  sleep(1);
}
```

**Stress Test — Yük altında davranış:**

```javascript
// resilience/load-test/k6-scripts/stress.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },    // Ramp up
    { duration: '3m', target: 50 },    // Sabit yük
    { duration: '1m', target: 100 },   // Artır
    { duration: '3m', target: 100 },   // Sabit
    { duration: '1m', target: 200 },   // Stres
    { duration: '2m', target: 200 },   // Stres altında
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],   // 2 saniye altında
    http_req_failed: ['rate<0.05'],      // %5 altında hata
  },
};

// ... test senaryosu
```

---

## 10. Altyapı Güncellemeleri

### 10.1 Docker Compose Eklentileri

```yaml
  # Faz 1+2+3 servislerinin hepsi aynen kalır, aşağıdakiler eklenir:

  # --- Notification Service ---
  notification-service:
    build: ./services/notification-service
    ports: ["5008:5008"]
    environment:
      - MONGODB_URI=mongodb://mongo:27017
      - MONGODB_DATABASE=streamvault_notifications
      - RABBITMQ_URL=amqp://streamvault:secret@rabbitmq:5672/
      - CONSUL_ADDRESS=consul:8500
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
    depends_on:
      mongo:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy

  # --- Observability Stack ---
  jaeger:
    image: jaegertracing/all-in-one:1.54
    ports:
      - "16686:16686"   # Jaeger UI
      - "4317:4317"     # OTLP gRPC
      - "4318:4318"     # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true

  prometheus:
    image: prom/prometheus:v2.50.0
    ports: ["9090:9090"]
    volumes:
      - ./observability/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.retention.time=30d'

  grafana:
    image: grafana/grafana:10.3.0
    ports: ["3000:3000"]
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=streamvault
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./observability/grafana/provisioning:/etc/grafana/provisioning
      - ./observability/grafana/grafana.ini:/etc/grafana/grafana.ini
    depends_on:
      - prometheus
      - loki
      - jaeger

  loki:
    image: grafana/loki:2.9.0
    ports: ["3100:3100"]
    volumes:
      - ./observability/loki/loki-config.yml:/etc/loki/local-config.yaml
      - loki_data:/loki
    command: -config.file=/etc/loki/local-config.yaml

  promtail:
    image: grafana/promtail:2.9.0
    volumes:
      - ./observability/promtail/promtail-config.yml:/etc/promtail/config.yml
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
    command: -config.file=/etc/promtail/config.yml
    depends_on:
      - loki

volumes:
  # ... Faz 1+2+3 volume'ları +
  prometheus_data:
  grafana_data:
  loki_data:
```

### 10.2 Final Port Haritası

| Servis | İç Port | Dış Port | Protokol | Açıklama |
|--------|---------|----------|----------|----------|
| **Application Services** | | | | |
| API Gateway | 8080 | 8080 | HTTP | Ana giriş noktası |
| User Service | 5001 | — | HTTP | Auth, profil, watchlist |
| Catalog Service | 5002 | — | HTTP | Film/dizi metadata |
| Streaming Service | 5003 | 5003 | HTTP | Video serving |
| Streaming Service | 50051 | 50051 | gRPC | Servisler arası |
| Encoding Service | 5004 | — | HTTP | Video transcoding |
| Encoding Service | 50052 | 50052 | gRPC | Servisler arası |
| Search Service | 5005 | 5005 | HTTP | Full-text arama |
| Search Service | 50053 | 50053 | gRPC | Servisler arası |
| Recommendation Service | 5006 | 5006 | HTTP | İçerik önerileri |
| Recommendation Service | 50054 | 50054 | gRPC | Servisler arası |
| Subscription Service | 5007 | — | HTTP | Abonelik yönetimi |
| Notification Service | 5008 | 5008 | HTTP+WS | Bildirimler |
| **Data Stores** | | | | |
| PostgreSQL | 5432 | 5432 | TCP | User, Subscription, Recommendation DB |
| MongoDB | 27017 | 27017 | TCP | Catalog, Notification DB |
| Redis | 6379 | 6379 | TCP | Cache, sessions, feature store |
| Elasticsearch | 9200 | 9200 | HTTP | Search index |
| MinIO | 9000 | 9000 | HTTP | Object storage API |
| MinIO Console | 9001 | 9001 | HTTP | Object storage UI |
| **Messaging** | | | | |
| RabbitMQ | 5672 | 5672 | AMQP | Message broker |
| RabbitMQ UI | 15672 | 15672 | HTTP | Management UI |
| **Observability** | | | | |
| Jaeger UI | 16686 | 16686 | HTTP | Distributed tracing UI |
| Jaeger OTLP | 4317 | 4317 | gRPC | Trace collector |
| Jaeger OTLP | 4318 | 4318 | HTTP | Trace collector |
| Prometheus | 9090 | 9090 | HTTP | Metrics UI |
| Grafana | 3000 | 3000 | HTTP | Dashboard UI |
| Loki | 3100 | 3100 | HTTP | Log aggregation |
| **Service Discovery** | | | | |
| Consul | 8500 | 8500 | HTTP | Service discovery UI |

**Toplam: 8 uygulama servisi + 5 data store + 1 message broker + 5 observability aracı + 1 service discovery = 20 container**

### 10.3 Tüm Servislere Eklenen Ortak Değişiklikler

Her servis şu güncellemeleri alır:

```
1. OpenTelemetry SDK entegrasyonu
   → Trace export (Jaeger'a OTLP ile)
   → Automatic instrumentation (HTTP, gRPC, DB)
   → Context propagation (trace-id header'ları)

2. Prometheus metrics endpoint
   → GET /metrics (Prometheus format)
   → RED metrikleri + servise özel business metrikleri

3. Structured JSON logging
   → trace-id ve span-id log'lara eklenir
   → Loki ile Jaeger trace'lerine link oluşturulabilir

4. Health check endpoints
   → /health/live, /health/ready, /health/startup

5. Graceful shutdown
   → SIGTERM yakalanır
   → Mevcut istekler tamamlanır
   → DB/Redis/RabbitMQ bağlantıları düzgünce kapatılır
   → Consul'dan deregister edilir

6. Environment-based config
   → OTEL_EXPORTER_OTLP_ENDPOINT
   → OTEL_SERVICE_NAME
```

---

## 11. Haftalık İlerleme Planı

### Hafta 1: Notification Service + Observability Altyapısı

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Observability stack kurulumu | Docker Compose'a Jaeger, Prometheus, Grafana, Loki, Promtail ekle. UI'ların çalıştığını doğrula. Grafana'ya datasource'lar ekle. |
| 2 | Notification Service — Temel | Go projesi oluştur, RabbitMQ consumer'lar (subscription, encoding, content event'leri). MongoDB'de bildirim kaydetme. Template engine. |
| 3 | Notification Service — WebSocket | Hub/Client mimarisi, JWT auth, bağlantı yönetimi, heartbeat. Browser'dan test. |
| 4 | Notification Service — Dispatcher | Event → kanal yönlendirme, kullanıcı tercihleri, email mock (MailHog), bildirim geçmişi endpoint'leri. |
| 5 | OpenTelemetry — Go servisleri | Gateway, Streaming, Search, Notification servislerine OTel entegrasyonu. Jaeger'da trace'lerin göründüğünü doğrula. |

### Hafta 2: Metrics + Resilience + Chaos

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | OpenTelemetry — Rust + .NET | Encoding, Recommendation'a Rust OTel. User, Catalog, Subscription'a .NET OTel. Asenkron event'lerde trace propagation. |
| 2 | Prometheus metrikleri | Tüm servislere /metrics endpoint ekle. RED + business metrikleri. Prometheus'un scrape ettiğini doğrula. |
| 3 | Grafana dashboard'ları | 3 dashboard oluştur: Services Overview, Streaming & Encoding, Business Metrics. Alert rule'ları tanımla. |
| 4 | Resilience — Circuit breaker + Retry | Gateway'e circuit breaker ekle. Tüm servislere retry policy. Timeout hiyerarşisi. Bulkhead pattern. Graceful degradation fallback'leri. |
| 5 | Structured logging + Loki | Tüm servislerde JSON log standardizasyonu. trace-id ekleme. Loki'de log sorgulama. Grafana'da log paneli. |

### Opsiyonel Hafta 3: Chaos Testing + Load Testing + Polish

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Health check ekosistemi | Tüm servislere live/ready/startup endpoint'leri. Gateway health aggregation. Consul health check entegrasyonu. |
| 2 | Chaos testing | 3 chaos senaryosu çalıştır: servis ölümü, DB yavaşlaması, RabbitMQ kesintisi. Sonuçları dokümante et. |
| 3 | Load testing | k6 ile smoke, load, stress test. Darboğazları belirle. Grafana'da yük altındaki metrikleri gözlemle. |
| 4 | Graceful shutdown + Final fixes | Tüm servislerde SIGTERM handling. Consul deregistration. Son bug fix'ler ve edge case'ler. |
| 5 | Final dokümantasyon | Tüm README'ler güncelle. Architecture Decision Records (ADR) yaz. Runbook hazırla (common issues + çözümleri). |

---

## 12. Faz 4 Bitiş Kriterleri (Definition of Done)

### Notification Service

- [ ] WebSocket ile gerçek zamanlı bildirim gönderilebiliyor
- [ ] Subscription, encoding, content event'leri doğru bildirim kanalına yönleniyor
- [ ] Bildirim tercihleri (email/push/inapp açma-kapama) çalışıyor
- [ ] Bildirim geçmişi listeleme ve okundu işaretleme çalışıyor
- [ ] Email template'leri render ediliyor (MailHog ile doğrulanmış)

### Distributed Tracing

- [ ] Jaeger UI'da tüm servisler görünüyor
- [ ] Senkron istek zinciri tek trace altında izlenebiliyor (Gateway → Service → DB)
- [ ] Asenkron event zinciri linked span olarak görünüyor (Publish → Consume)
- [ ] gRPC çağrıları trace'te görünüyor
- [ ] Database sorguları span olarak trace'te yer alıyor

### Metrics & Dashboards

- [ ] Tüm servisler /metrics endpoint'i sunuyor
- [ ] Prometheus tüm servislerden metrikleri topluyor
- [ ] Grafana'da 3 dashboard çalışıyor (Overview, Streaming, Business)
- [ ] En az 1 alert rule tanımlı ve çalışıyor (ör: error rate > %5)

### Logging

- [ ] Tüm servisler structured JSON log yazıyor
- [ ] Log'larda trace-id mevcut (Loki → Jaeger bağlantısı)
- [ ] Loki'de servis bazlı log filtreleme çalışıyor
- [ ] Grafana'da log paneli kurulu

### Resilience

- [ ] Gateway'de circuit breaker çalışıyor (OPEN durumda 503 dönüyor)
- [ ] Retry policy tanımlı ve çalışıyor (geçici hatalarda otomatik tekrar)
- [ ] Timeout hiyerarşisi doğru (iç timeout < dış timeout)
- [ ] Health check endpoint'leri (live/ready/startup) tüm servislerde mevcut
- [ ] Gateway health aggregation çalışıyor
- [ ] Graceful shutdown tüm servislerde çalışıyor (SIGTERM → clean exit)

### Chaos & Load Testing

- [ ] En az 3 chaos senaryosu çalıştırılmış ve sonuçları dokümante edilmiş
- [ ] k6 ile smoke test geçiyor (tüm endpoint'ler sağlıklı)
- [ ] Stress test altında sistem graceful degradation gösteriyor
- [ ] Chaos testi sonrası sistem otomatik recovery yapıyor

### Genel

- [ ] `docker compose up` ile 20 container 2 dakika içinde ayağa kalkıyor
- [ ] Tüm servisler Consul'da "healthy" görünüyor
- [ ] README.md güncel: kurulum, mimari diyagram, troubleshooting
- [ ] Projenin tamamı için toplam test coverage %50+

---

## 13. Proje Tamamlandığında — Öğrenilen Her Şey

```
┌──────────────────────────────────────────────────────────────────────────┐
│                    StreamVault — Öğrenme Haritası                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  🏗️ Mimari Pattern'ler                                                   │
│  ├── API Gateway Pattern                                                 │
│  ├── Database per Service                                                │
│  ├── CQRS (Command Query Responsibility Segregation)                     │
│  ├── Event-Driven Architecture                                           │
│  ├── Saga Pattern (Orchestration)                                        │
│  ├── Outbox Pattern                                                      │
│  ├── Circuit Breaker                                                     │
│  ├── Bulkhead Pattern                                                    │
│  ├── Service Discovery                                                   │
│  ├── Strangler Fig (mevcut sisteme microservice ekleme)                  │
│  └── Polyglot Persistence + Polyglot Programming                         │
│                                                                          │
│  🔧 Teknolojiler                                                        │
│  ├── Go: HTTP server, gRPC, goroutine, channel, reverse proxy            │
│  ├── Rust: async/await (tokio), FFmpeg binding, error handling           │
│  ├── .NET: Clean Architecture, EF Core, MediatR, Background Services     │
│  ├── gRPC + Protocol Buffers                                             │
│  ├── RabbitMQ: exchange, queue, binding, DLX, consumer pattern           │
│  ├── PostgreSQL, MongoDB, Redis, Elasticsearch, MinIO                    │
│  ├── Docker Compose: multi-container orchestration                       │
│  ├── Consul: service discovery, health checking                          │
│  ├── HLS: adaptive bitrate streaming                                     │
│  ├── FFmpeg: video transcoding                                           │
│  └── JWT: authentication, refresh token                                  │
│                                                                          │
│  📊 Observability                                                        │
│  ├── OpenTelemetry: distributed tracing                                  │
│  ├── Jaeger: trace visualization                                         │
│  ├── Prometheus: metrics collection                                      │
│  ├── Grafana: dashboards, alerting                                       │
│  ├── Loki: centralized logging                                           │
│  └── Structured logging: JSON, correlation ID                            │
│                                                                          │
│  🛡️ Resilience                                                          │
│  ├── Circuit breaker, retry, timeout                                     │
│  ├── Graceful degradation                                                │
│  ├── Chaos testing                                                       │
│  ├── Load testing (k6)                                                   │
│  ├── Health checks (live/ready/startup)                                  │
│  └── Graceful shutdown                                                   │
│                                                                          │
│  💡 Kavramlar                                                            │
│  ├── Eventual consistency                                                │
│  ├── Idempotency                                                         │
│  ├── Compensating transactions                                           │
│  ├── Collaborative filtering                                             │
│  ├── Content-based filtering                                             │
│  ├── Cold start problem                                                  │
│  ├── Backpressure                                                        │
│  └── Trace context propagation                                           │
│                                                                          │
│  📈 Toplam                                                               │
│  ├── 8 microservice (3 Go + 2 Rust + 3 .NET)                            │
│  ├── 5 data store                                                        │
│  ├── 20 Docker container                                                 │
│  ├── 50+ REST endpoint                                                   │
│  ├── 10+ gRPC method                                                     │
│  ├── 5+ RabbitMQ exchange, 15+ queue                                     │
│  └── 3 Grafana dashboard                                                 │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 14. Sonraki Adımlar (Projeyi İleriye Taşımak İstersen)

Faz 4 tamamlandıktan sonra projeyi daha da geliştirebilirsin:

- **Kubernetes'e geçiş** — Docker Compose → Helm charts, pod autoscaling, ingress controller
- **CI/CD pipeline** — GitHub Actions ile build, test, push, deploy
- **Frontend** — React veya Next.js ile Netflix benzeri bir arayüz
- **API versioning** — /v1, /v2 ile geriye uyumlu API geliştirme
- **Rate limiting gelişmiş** — Kullanıcı bazlı, endpoint bazlı, sliding window
- **A/B testing** — Recommendation algoritmaları arasında A/B test altyapısı
- **Multi-region** — CockroachDB veya Cassandra ile dağıtık veri
- **Event sourcing** — Catalog Service'te tam event sourcing
- **GraphQL Gateway** — REST yerine GraphQL BFF (Backend for Frontend)
- **Service mesh** — Istio veya Linkerd ile ağ seviyesinde observability ve güvenlik
