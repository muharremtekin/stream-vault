# StreamVault — Faz 2: Core Streaming

> **Süre:** ~2-3 Hafta
> **Ön Koşul:** Faz 1 tamamlanmış olmalı (Gateway, User Service, Catalog Service çalışır durumda)
> **Hedef:** Video upload → transcode → HLS segmentlere ayır → serve akışının uçtan uca çalışması. gRPC ile servisler arası senkron iletişim, RabbitMQ ile asenkron event pipeline.
> **Sonuç:** Bir admin video yükleyebilir, sistem otomatik olarak birden fazla kaliteye dönüştürür ve kullanıcı adaptive bitrate ile izleyebilir. İzleme pozisyonu kaydedilir (kaldığın yerden devam).

---

## 1. Faz 2'de Neler Ekleniyor?

| Bileşen | Dil | Yeni/Güncelleme | Açıklama |
|---------|-----|-----------------|----------|
| Streaming Service | Go | 🆕 Yeni | HLS manifest üretimi, chunk serving, izleme pozisyonu |
| Encoding Service | Rust | 🆕 Yeni | Video transcoding pipeline, thumbnail üretimi |
| RabbitMQ | — | 🆕 Yeni altyapı | Event bus, encoding job queue |
| MinIO | — | 🆕 Yeni altyapı | S3-uyumlu object storage (video dosyaları) |
| Proto Repo | — | 🆕 Yeni | Paylaşımlı gRPC proto tanımları |
| Catalog Service | .NET | 🔄 Güncelleme | Video metadata alanları, event publishing |
| API Gateway | Go | 🔄 Güncelleme | Streaming route'ları, gRPC proxy desteği |

---

## 2. Yeni Proje Yapısı (Faz 1'e Eklenenler)

```
streamvault/
├── ... (Faz 1 yapısı aynen kalır)
│
├── proto/                                  # 🔄 Genişletildi
│   ├── user/v1/user.proto
│   ├── catalog/v1/catalog.proto
│   ├── streaming/v1/streaming.proto        # 🆕
│   ├── encoding/v1/encoding.proto          # 🆕
│   └── common/v1/common.proto              # 🆕 Paylaşımlı tipler
│
├── gateway/                                # 🔄 Güncellendi
│   ├── internal/
│   │   ├── middleware/
│   │   │   └── streaming_auth.go          # 🆕 Streaming özel auth (subscription tier kontrolü)
│   │   └── proxy/
│   │       ├── router.go                  # 🔄 Yeni streaming route'ları eklendi
│   │       └── streaming_proxy.go         # 🆕 Streaming özel proxy (range requests, chunked)
│   └── ...
│
├── services/
│   ├── user-service/                       # Faz 1'den — değişiklik yok
│   ├── catalog-service/                    # 🔄 Güncellendi
│   │   ├── src/
│   │   │   ├── CatalogService.Domain/
│   │   │   │   └── Entities/
│   │   │   │       └── Movie.cs           # 🔄 VideoStatus, StreamingInfo eklendi
│   │   │   ├── CatalogService.Application/
│   │   │   │   ├── Events/               # 🆕
│   │   │   │   │   ├── ContentUploadedEvent.cs
│   │   │   │   │   └── IEventPublisher.cs
│   │   │   │   └── Commands/
│   │   │   │       └── UpdateVideoStatus/  # 🆕
│   │   │   │           ├── UpdateVideoStatusCommand.cs
│   │   │   │           └── UpdateVideoStatusHandler.cs
│   │   │   └── CatalogService.Infrastructure/
│   │   │       └── Messaging/             # 🆕
│   │   │           └── RabbitMqEventPublisher.cs
│   │   └── ...
│   │
│   ├── streaming-service/                  # 🆕 Go — Streaming Service
│   │   ├── cmd/
│   │   │   └── streaming/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   │   └── config.go
│   │   │   ├── handler/
│   │   │   │   ├── manifest.go            # HLS manifest (.m3u8) üretimi
│   │   │   │   ├── chunk.go              # Video chunk serving
│   │   │   │   ├── progress.go           # İzleme pozisyonu kaydetme/okuma
│   │   │   │   └── upload.go             # Video upload (multipart)
│   │   │   ├── hls/
│   │   │   │   ├── master_playlist.go    # Master playlist (multi-quality)
│   │   │   │   └── media_playlist.go     # Tek kalite playlist
│   │   │   ├── storage/
│   │   │   │   ├── minio.go             # MinIO client wrapper
│   │   │   │   └── interface.go
│   │   │   ├── grpc/
│   │   │   │   ├── server.go            # gRPC server (streaming bilgi sorgulama)
│   │   │   │   └── catalog_client.go    # Catalog Service gRPC client
│   │   │   ├── progress/
│   │   │   │   ├── repository.go        # Redis'te izleme pozisyonu
│   │   │   │   └── service.go
│   │   │   ├── discovery/
│   │   │   │   └── consul.go            # Consul registration
│   │   │   └── middleware/
│   │   │       ├── subscription.go      # Abonelik tier kontrolü
│   │   │       └── concurrent.go        # Eşzamanlı izleme limiti
│   │   ├── go.mod
│   │   ├── Dockerfile
│   │   └── Makefile
│   │
│   └── encoding-service/                   # 🆕 Rust — Encoding Service
│       ├── src/
│       │   ├── main.rs                    # Actix-web/Axum bootstrap + RabbitMQ consumer
│       │   ├── config.rs                  # Konfigürasyon yönetimi
│       │   ├── api/
│       │   │   ├── mod.rs
│       │   │   ├── routes.rs             # HTTP endpoints (job status, health)
│       │   │   └── handlers.rs
│       │   ├── pipeline/
│       │   │   ├── mod.rs
│       │   │   ├── orchestrator.rs       # Pipeline adımlarını sırayla çalıştırır
│       │   │   ├── validator.rs          # Video format doğrulama
│       │   │   ├── transcoder.rs         # FFmpeg ile transcoding
│       │   │   ├── segmenter.rs          # HLS segment üretimi
│       │   │   ├── thumbnail.rs          # Thumbnail/poster üretimi
│       │   │   └── uploader.rs           # Sonuçları MinIO'ya yükleme
│       │   ├── queue/
│       │   │   ├── mod.rs
│       │   │   ├── consumer.rs           # RabbitMQ consumer (job alır)
│       │   │   └── publisher.rs          # Event publish (encoding tamamlandı)
│       │   ├── storage/
│       │   │   ├── mod.rs
│       │   │   └── minio.rs              # MinIO S3 client
│       │   ├── models/
│       │   │   ├── mod.rs
│       │   │   ├── job.rs                # EncodingJob struct
│       │   │   ├── profile.rs            # Encoding profilleri (720p, 1080p...)
│       │   │   └── status.rs             # Job durumları
│       │   └── error.rs                  # Hata tipleri
│       ├── Cargo.toml
│       ├── Dockerfile
│       └── Makefile
│
├── infrastructure/
│   ├── ... (Faz 1'den)
│   ├── rabbitmq/                          # 🆕
│   │   ├── rabbitmq.conf
│   │   └── definitions.json              # Exchange, queue, binding tanımları
│   └── minio/                             # 🆕
│       └── init-buckets.sh               # Bucket oluşturma script'i
│
└── scripts/
    ├── ... (Faz 1'den)
    ├── generate-proto.sh                  # 🔄 Tüm diller için proto üretim
    ├── upload-test-video.sh               # 🆕 Test video yükleme
    └── test-streaming.sh                  # 🆕 Streaming akışı test
```

---

## 3. gRPC Proto Tanımları

### 3.1 common/v1/common.proto

```protobuf
syntax = "proto3";
package streamvault.common.v1;
option go_package = "github.com/streamvault/proto/common/v1";
option csharp_namespace = "StreamVault.Proto.Common.V1";

message Pagination {
  int32 page = 1;
  int32 page_size = 2;
}

message PaginatedResponse {
  int32 page = 1;
  int32 page_size = 2;
  int64 total_count = 3;
  int32 total_pages = 4;
}

enum ContentType {
  CONTENT_TYPE_UNSPECIFIED = 0;
  CONTENT_TYPE_MOVIE = 1;
  CONTENT_TYPE_EPISODE = 2;
}

enum SubscriptionTier {
  SUBSCRIPTION_TIER_UNSPECIFIED = 0;
  SUBSCRIPTION_TIER_BASIC = 1;      // 720p, 1 ekran
  SUBSCRIPTION_TIER_STANDARD = 2;   // 1080p, 2 ekran
  SUBSCRIPTION_TIER_PREMIUM = 3;    // 4K, 4 ekran
}
```

### 3.2 streaming/v1/streaming.proto

```protobuf
syntax = "proto3";
package streamvault.streaming.v1;
option go_package = "github.com/streamvault/proto/streaming/v1";

import "common/v1/common.proto";
import "google/protobuf/timestamp.proto";

// Streaming Service tarafından sunulur
service StreamingService {
  // İçerik için streaming bilgisi sorgula
  rpc GetStreamingInfo(GetStreamingInfoRequest) returns (GetStreamingInfoResponse);
  // Kullanıcının izleme pozisyonunu al
  rpc GetProgress(GetProgressRequest) returns (GetProgressResponse);
  // Kullanıcının tüm devam eden izlemelerini listele ("Kaldığın Yerden Devam Et")
  rpc GetContinueWatching(GetContinueWatchingRequest) returns (GetContinueWatchingResponse);
}

message GetStreamingInfoRequest {
  string content_id = 1;
}

message GetStreamingInfoResponse {
  string content_id = 1;
  StreamingStatus status = 2;
  repeated QualityOption available_qualities = 3;
  string manifest_url = 4;          // /stream/{id}/manifest.m3u8
  int64 duration_seconds = 5;
}

enum StreamingStatus {
  STREAMING_STATUS_UNSPECIFIED = 0;
  STREAMING_STATUS_NOT_AVAILABLE = 1;
  STREAMING_STATUS_ENCODING = 2;
  STREAMING_STATUS_READY = 3;
  STREAMING_STATUS_ERROR = 4;
}

message QualityOption {
  string label = 1;                  // "720p", "1080p", "4K"
  int32 width = 2;
  int32 height = 3;
  int32 bitrate_kbps = 4;
  streamvault.common.v1.SubscriptionTier min_tier = 5;  // Bu kalite için minimum abonelik
}

message GetProgressRequest {
  string user_id = 1;
  string content_id = 2;
}

message GetProgressResponse {
  string content_id = 1;
  int64 position_seconds = 2;
  int64 duration_seconds = 3;
  double percentage = 4;
  google.protobuf.Timestamp updated_at = 5;
}

message GetContinueWatchingRequest {
  string user_id = 1;
  int32 limit = 2;
}

message GetContinueWatchingResponse {
  repeated WatchProgress items = 1;
}

message WatchProgress {
  string content_id = 1;
  string title = 2;
  string thumbnail_url = 3;
  int64 position_seconds = 4;
  int64 duration_seconds = 5;
  double percentage = 6;
  google.protobuf.Timestamp updated_at = 7;
}
```

### 3.3 encoding/v1/encoding.proto

```protobuf
syntax = "proto3";
package streamvault.encoding.v1;
option go_package = "github.com/streamvault/proto/encoding/v1";

import "google/protobuf/timestamp.proto";

// Encoding Service tarafından sunulur
service EncodingService {
  // Encoding job durumunu sorgula
  rpc GetJobStatus(GetJobStatusRequest) returns (GetJobStatusResponse);
  // Aktif job'ları listele
  rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
}

message GetJobStatusRequest {
  string job_id = 1;
}

message GetJobStatusResponse {
  string job_id = 1;
  string content_id = 2;
  JobStatus status = 3;
  double progress_percentage = 4;      // 0.0 - 100.0
  string current_step = 5;             // "validating", "transcoding_720p", "segmenting"...
  repeated EncodingOutput outputs = 6;
  string error_message = 7;
  google.protobuf.Timestamp started_at = 8;
  google.protobuf.Timestamp completed_at = 9;
}

enum JobStatus {
  JOB_STATUS_UNSPECIFIED = 0;
  JOB_STATUS_QUEUED = 1;
  JOB_STATUS_PROCESSING = 2;
  JOB_STATUS_COMPLETED = 3;
  JOB_STATUS_FAILED = 4;
  JOB_STATUS_CANCELLED = 5;
}

message EncodingOutput {
  string quality = 1;                   // "360p", "720p", "1080p", "4k"
  int32 width = 2;
  int32 height = 3;
  int32 bitrate_kbps = 4;
  int64 file_size_bytes = 5;
  int32 segment_count = 6;
  string storage_path = 7;             // MinIO path
}

message ListJobsRequest {
  JobStatus status_filter = 1;
  int32 limit = 2;
  int32 offset = 3;
}

message ListJobsResponse {
  repeated GetJobStatusResponse jobs = 1;
  int64 total_count = 2;
}
```

---

## 4. Servis Detayları

### 4.1 Streaming Service (Go)

**Sorumluluklar:**

- HLS master ve media playlist (.m3u8) üretimi
- Video chunk (segment) serving — MinIO'dan okuyup istemciye iletir
- Adaptive bitrate: kullanıcının abonelik seviyesine göre mevcut kaliteleri filtreler
- İzleme pozisyonu kaydetme ve okuma (Redis)
- "Kaldığın Yerden Devam Et" listesi
- Eşzamanlı izleme limiti (Basic: 1, Standard: 2, Premium: 4)
- Video upload endpoint'i (admin) — dosyayı MinIO'ya yükler ve encoding job başlatır
- gRPC server (diğer servisler streaming durumunu sorgulayabilir)

**HTTP API Kontratı:**

```
=== Video İzleme ===

GET /stream/{contentId}/manifest.m3u8
  Headers:  X-User-Id, X-User-Tier (gateway tarafından eklenir)
  Response: 200 Content-Type: application/vnd.apple.mpegurl
  Body:
    #EXTM3U
    #EXT-X-VERSION:3
    #EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360
    /stream/{contentId}/360p/playlist.m3u8
    #EXT-X-STREAM-INF:BANDWIDTH=2400000,RESOLUTION=1280x720
    /stream/{contentId}/720p/playlist.m3u8
    #EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080
    /stream/{contentId}/1080p/playlist.m3u8
  
  Not: Kullanıcının tier'ine göre kaliteler filtrelenir.
  Basic → sadece 360p, 720p
  Standard → 360p, 720p, 1080p
  Premium → tüm kaliteler + 4K

GET /stream/{contentId}/{quality}/playlist.m3u8
  Response: 200 Content-Type: application/vnd.apple.mpegurl
  Body:
    #EXTM3U
    #EXT-X-VERSION:3
    #EXT-X-TARGETDURATION:10
    #EXT-X-MEDIA-SEQUENCE:0
    #EXTINF:10.0,
    /stream/{contentId}/{quality}/segment_000.ts
    #EXTINF:10.0,
    /stream/{contentId}/{quality}/segment_001.ts
    ...
    #EXT-X-ENDLIST

GET /stream/{contentId}/{quality}/segment_{number}.ts
  Response: 200 Content-Type: video/mp2t
  Body: Binary video segment data (MinIO'dan proxy)
  Headers: Accept-Ranges: bytes, Content-Length: ...

=== İzleme Pozisyonu ===

POST /api/stream/{contentId}/progress
  Headers: X-User-Id
  Request:  { "positionSeconds": 1847, "durationSeconds": 8820 }
  Response: 200 { "saved": true }
  Not: Her 10-30 saniyede bir client tarafından gönderilir.

GET /api/stream/{contentId}/progress
  Headers: X-User-Id
  Response: 200 {
    "contentId": "abc",
    "positionSeconds": 1847,
    "durationSeconds": 8820,
    "percentage": 20.94,
    "updatedAt": "2026-02-13T14:30:00Z"
  }

GET /api/stream/continue-watching
  Headers: X-User-Id
  Response: 200 [
    {
      "contentId": "abc",
      "title": "Interstellar",
      "thumbnailUrl": "https://minio:9000/thumbnails/abc.jpg",
      "positionSeconds": 1847,
      "durationSeconds": 8820,
      "percentage": 20.94,
      "updatedAt": "2026-02-13T14:30:00Z"
    },
    ...
  ]
  Not: Tamamlanmış (%95+) içerikler listeden düşer.

=== Video Upload (Admin) ===

POST /api/stream/upload
  Headers: X-User-Id, X-User-Role: Admin
  Body: multipart/form-data
    - file: video dosyası (mp4, mkv, avi)
    - contentId: Catalog'daki içerik ID'si
    - title: "Interstellar" (metadata)
  Response: 202 {
    "jobId": "job-uuid",
    "contentId": "abc",
    "status": "queued",
    "message": "Video encoding kuyruğuna eklendi"
  }
  Not: Dosya MinIO'ya yüklenir, RabbitMQ'ya encoding job mesajı gönderilir.

=== Sağlık ve Durum ===

GET /health
  Response: 200 { "status": "healthy", "minio": "connected", "redis": "connected" }
```

**MinIO Bucket Yapısı:**

```
streamvault-raw/                    # Ham upload edilen videolar
  └── {contentId}/
      └── original.mp4

streamvault-encoded/                # Encode edilmiş segmentler
  └── {contentId}/
      ├── 360p/
      │   ├── playlist.m3u8
      │   ├── segment_000.ts
      │   ├── segment_001.ts
      │   └── ...
      ├── 720p/
      │   ├── playlist.m3u8
      │   └── ...
      ├── 1080p/
      │   └── ...
      └── 4k/
          └── ...

streamvault-thumbnails/             # Thumbnail ve poster görselleri
  └── {contentId}/
      ├── poster.jpg                # Ana poster (yüksek çözünürlük)
      ├── thumbnail.jpg             # Küçük thumbnail
      └── preview_{timestamp}.jpg   # Timeline preview görselleri
```

**Redis Veri Yapıları:**

```
# İzleme pozisyonu
progress:{userId}:{contentId} → Hash
  positionSeconds: 1847
  durationSeconds: 8820
  updatedAt: "2026-02-13T14:30:00Z"
  TTL: 90 gün

# "Kaldığın Yerden Devam Et" — sorted set (score = timestamp)
continue:{userId} → Sorted Set
  member: {contentId}
  score: 1739448600 (unix timestamp)

# Eşzamanlı izleme — set
concurrent:{userId} → Set
  members: ["session-uuid-1", "session-uuid-2"]
  TTL: 5 dakika (heartbeat ile yenilenir)

# Streaming bilgi cache
stream_info:{contentId} → Hash
  status: "ready"
  qualities: "360p,720p,1080p"
  segmentCount360: 88
  segmentCount720: 88
  segmentCount1080: 88
  durationSeconds: 8820
  TTL: 1 saat
```

**Konfigürasyon (config.yaml):**

```yaml
server:
  http_port: 5003
  grpc_port: 50051
  read_timeout: 30s
  write_timeout: 60s           # Video chunk serving için uzun timeout

minio:
  endpoint: "minio:9000"
  access_key: "${MINIO_ACCESS_KEY}"
  secret_key: "${MINIO_SECRET_KEY}"
  use_ssl: false
  raw_bucket: "streamvault-raw"
  encoded_bucket: "streamvault-encoded"
  thumbnail_bucket: "streamvault-thumbnails"

redis:
  url: "redis://redis:6379/1"   # Farklı DB (Gateway'den ayrı)
  progress_ttl: "2160h"         # 90 gün
  concurrent_ttl: "5m"

rabbitmq:
  url: "amqp://guest:guest@rabbitmq:5672/"
  encoding_exchange: "encoding"
  encoding_queue: "encoding.jobs"
  result_queue: "encoding.results"

consul:
  address: "consul:8500"
  service_name: "streaming-service"
  service_port: 5003

concurrent_limits:
  basic: 1
  standard: 2
  premium: 4

segment:
  duration_seconds: 10
  default_quality: "720p"
```

**Öğrenme Noktaları:**

- HLS (HTTP Live Streaming) protokolü ve adaptive bitrate nasıl çalışır
- Go'da gRPC server ve client implementasyonu
- MinIO (S3 uyumlu) object storage kullanımı
- Redis'te çeşitli veri yapıları (Hash, Sorted Set, Set)
- Go'da multipart file upload ve büyük dosya handling
- Concurrent access kontrolü (eşzamanlı izleme limiti)
- Go channels ile RabbitMQ publisher

---

### 4.2 Encoding Service (Rust)

**Sorumluluklar:**

- RabbitMQ'dan encoding job'ları consume eder
- Ham videoyu doğrular (format, codec, boyut kontrolü)
- FFmpeg ile birden fazla kaliteye transcode eder (360p, 720p, 1080p, 4K)
- HLS segmentlerine ayırır (10 saniyelik .ts dosyaları)
- Her kalite için playlist.m3u8 üretir
- Thumbnail ve poster görselleri oluşturur
- Sonuçları MinIO'ya yükler
- İş durumunu (progress) Redis üzerinden raporlar
- Tamamlandığında event publish eder → Catalog Service video durumunu günceller
- gRPC server ile job durum sorgulama

**Encoding Profilleri:**

```
┌────────────┬────────────┬────────────┬───────────┬────────────────────┐
│ Profil     │ Çözünürlük │ Bitrate    │ Min Tier  │ FFmpeg Preset      │
├────────────┼────────────┼────────────┼───────────┼────────────────────┤
│ 360p       │ 640x360    │ 800 kbps   │ Basic     │ -preset fast       │
│ 720p       │ 1280x720   │ 2400 kbps  │ Basic     │ -preset medium     │
│ 1080p      │ 1920x1080  │ 5000 kbps  │ Standard  │ -preset medium     │
│ 4K         │ 3840x2160  │ 15000 kbps │ Premium   │ -preset slow       │
└────────────┴────────────┴────────────┴───────────┴────────────────────┘

Audio: AAC 128kbps (tüm profiller)
Segment süresi: 10 saniye
Container: MPEG-TS (.ts)
Codec: H.264 (libx264)
```

**Pipeline Akışı:**

```
RabbitMQ (encoding.jobs queue)
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│  Encoding Pipeline Orchestrator                             │
│                                                             │
│  Step 1: VALIDATE                                           │
│  ├── MinIO'dan ham videoyu indir (temp dizine)              │
│  ├── ffprobe ile video bilgilerini oku                      │
│  ├── Format, codec, boyut kontrolü                          │
│  └── Kaynak çözünürlüğe göre hangi profillerin              │
│      üretileceğini belirle (720p kaynak → 360p + 720p)      │
│                                                             │
│  Step 2: TRANSCODE (her profil için paralel)                │
│  ├── FFmpeg ile hedef çözünürlüğe dönüştür                  │
│  ├── İlerleme: FFmpeg stdout parse → Redis'e yaz            │
│  └── Çıktı: {contentId}_{quality}.mp4                       │
│                                                             │
│  Step 3: SEGMENT                                            │
│  ├── Her kalite dosyasını 10s HLS segmentlerine ayır        │
│  ├── Her kalite için playlist.m3u8 üret                     │
│  └── Çıktı: segment_000.ts, segment_001.ts, ...            │
│                                                             │
│  Step 4: THUMBNAIL                                          │
│  ├── Videonun %10, %30, %50, %70, %90 noktalarından kare   │
│  ├── En iyi kareyi poster olarak seç (en yüksek kontrast)  │
│  └── Thumbnail resize (300x170, 600x340)                    │
│                                                             │
│  Step 5: UPLOAD                                             │
│  ├── Tüm segmentleri MinIO'ya yükle (streamvault-encoded)   │
│  ├── Thumbnail'leri MinIO'ya yükle (streamvault-thumbnails)  │
│  └── Temp dosyaları temizle                                 │
│                                                             │
│  Step 6: NOTIFY                                             │
│  ├── RabbitMQ'ya EncodingCompleted event publish et         │
│  └── Job durumunu "completed" olarak güncelle               │
└─────────────────────────────────────────────────────────────┘
    │
    ▼
RabbitMQ (encoding.results exchange)
    │
    ├──► Catalog Service: Video status güncelle (Ready)
    └──► (Faz 4'te) Notification Service: Admin'e bildirim
```

**RabbitMQ Mesaj Formatları:**

```json
// encoding.jobs queue'ya gönderilen mesaj (Streaming Service → Encoding Service)
{
  "jobId": "job-uuid",
  "contentId": "catalog-content-id",
  "title": "Interstellar",
  "source": {
    "bucket": "streamvault-raw",
    "key": "abc/original.mp4"
  },
  "profiles": ["360p", "720p", "1080p"],
  "priority": 5,
  "requestedBy": "admin-user-id",
  "requestedAt": "2026-02-13T10:00:00Z"
}

// encoding.results exchange'e publish edilen event (Encoding Service → Catalog + diğerleri)
{
  "eventType": "EncodingCompleted",
  "jobId": "job-uuid",
  "contentId": "catalog-content-id",
  "status": "completed",
  "outputs": [
    {
      "quality": "360p",
      "width": 640,
      "height": 360,
      "bitrateKbps": 800,
      "fileSizeBytes": 52428800,
      "segmentCount": 88,
      "storagePath": "abc/360p/"
    },
    {
      "quality": "720p",
      "width": 1280,
      "height": 720,
      "bitrateKbps": 2400,
      "fileSizeBytes": 157286400,
      "segmentCount": 88,
      "storagePath": "abc/720p/"
    }
  ],
  "thumbnails": {
    "poster": "abc/poster.jpg",
    "thumbnail": "abc/thumbnail.jpg"
  },
  "durationSeconds": 8820,
  "startedAt": "2026-02-13T10:00:05Z",
  "completedAt": "2026-02-13T10:15:30Z"
}

// Hata durumunda
{
  "eventType": "EncodingFailed",
  "jobId": "job-uuid",
  "contentId": "catalog-content-id",
  "status": "failed",
  "error": {
    "step": "transcode",
    "message": "FFmpeg exited with code 1: unsupported codec",
    "details": "Input codec 'vp9' is not supported for HLS transcoding"
  },
  "startedAt": "2026-02-13T10:00:05Z",
  "failedAt": "2026-02-13T10:00:12Z"
}
```

**HTTP API (Durum Sorgulama):**

```
GET /api/encoding/jobs/{jobId}
  Response: 200 {
    "jobId": "job-uuid",
    "contentId": "abc",
    "status": "processing",
    "progressPercentage": 45.2,
    "currentStep": "transcoding_1080p",
    "outputs": [ ... ],
    "startedAt": "2026-02-13T10:00:05Z"
  }

GET /api/encoding/jobs?status=processing&limit=10
  Response: 200 {
    "jobs": [ ... ],
    "totalCount": 3
  }

GET /health
  Response: 200 {
    "status": "healthy",
    "rabbitmq": "connected",
    "minio": "connected",
    "ffmpeg": "available",
    "activeJobs": 2
  }
```

**Konfigürasyon (config.toml):**

```toml
[server]
http_port = 5004
grpc_port = 50052

[minio]
endpoint = "minio:9000"
access_key = "${MINIO_ACCESS_KEY}"
secret_key = "${MINIO_SECRET_KEY}"
use_ssl = false
raw_bucket = "streamvault-raw"
encoded_bucket = "streamvault-encoded"
thumbnail_bucket = "streamvault-thumbnails"

[rabbitmq]
url = "amqp://guest:guest@rabbitmq:5672/"
job_queue = "encoding.jobs"
result_exchange = "encoding.results"
prefetch_count = 2                     # Aynı anda max 2 job işle
retry_max = 3
retry_delay_ms = 5000

[encoding]
temp_dir = "/tmp/encoding"
max_file_size_gb = 10
segment_duration_seconds = 10
ffmpeg_path = "/usr/bin/ffmpeg"
ffprobe_path = "/usr/bin/ffprobe"

[encoding.profiles.360p]
width = 640
height = 360
bitrate_kbps = 800
preset = "fast"

[encoding.profiles.720p]
width = 1280
height = 720
bitrate_kbps = 2400
preset = "medium"

[encoding.profiles.1080p]
width = 1920
height = 1080
bitrate_kbps = 5000
preset = "medium"

[encoding.profiles.4k]
width = 3840
height = 2160
bitrate_kbps = 15000
preset = "slow"

[redis]
url = "redis://redis:6379/2"

[consul]
address = "consul:8500"
service_name = "encoding-service"
service_port = 5004
```

**Öğrenme Noktaları:**

- Rust'ta async programlama (tokio runtime)
- FFmpeg'i subprocess olarak çalıştırma ve stdout stream parse
- RabbitMQ consumer pattern (prefetch, ack/nack, dead letter)
- Pipeline/orchestrator pattern (adım adım iş akışı)
- Error handling ve retry mekanizmaları (Rust'ın Result tipi)
- Büyük dosya işleme (streaming I/O, temp dosya yönetimi)
- MinIO S3 SDK kullanımı

---

## 5. Catalog Service Güncellemeleri

Faz 1'deki Catalog Service'e şu alanlar ve özellikler eklenir:

### 5.1 Domain Değişiklikleri

```csharp
// Movie entity'sine eklenen alanlar
public class Movie
{
    // ... Faz 1 alanları aynen kalır ...

    // 🆕 Faz 2 eklentileri
    public VideoStatus VideoStatus { get; set; } = VideoStatus.NotUploaded;
    public StreamingInfo? StreamingInfo { get; set; }
}

public enum VideoStatus
{
    NotUploaded,    // Henüz video yüklenmemiş
    Uploading,      // Yükleniyor
    Queued,         // Encoding kuyruğunda
    Encoding,       // Encoding devam ediyor
    Ready,          // İzlenmeye hazır
    Error           // Encoding hatası
}

public class StreamingInfo
{
    public int DurationSeconds { get; set; }
    public List<QualityInfo> AvailableQualities { get; set; } = new();
    public string ManifestPath { get; set; }       // MinIO path
    public string ThumbnailPath { get; set; }
    public string PosterPath { get; set; }
    public DateTime EncodedAt { get; set; }
}

public class QualityInfo
{
    public string Label { get; set; }              // "720p"
    public int Width { get; set; }
    public int Height { get; set; }
    public int BitrateKbps { get; set; }
    public int SegmentCount { get; set; }
}
```

### 5.2 RabbitMQ Consumer — EncodingCompleted Event

```
Encoding Service → encoding.results exchange → Catalog Service consumer

Catalog Service şunu yapar:
1. EncodingCompleted event'ini dinler
2. İlgili Movie/Episode'u bulur
3. VideoStatus'u "Ready" olarak günceller
4. StreamingInfo'yu doldurur (available qualities, duration, paths)
5. UpdatedAt'i günceller
```

### 5.3 Yeni Endpoint'ler

```
GET /api/catalog/movies/{id}/streaming-info
  Response: 200 {
    "videoStatus": "Ready",
    "durationSeconds": 8820,
    "availableQualities": [
      { "label": "720p", "width": 1280, "height": 720, "bitrateKbps": 2400 }
    ],
    "manifestUrl": "/stream/{id}/manifest.m3u8",
    "thumbnailUrl": "https://minio:9000/streamvault-thumbnails/{id}/thumbnail.jpg"
  }
  Error: 404 { "error": "Video henüz yüklenmemiş" }
```

---

## 6. API Gateway Güncellemeleri

### 6.1 Yeni Route'lar

| Method | Path | Downstream | Auth | Özel |
|--------|------|------------|------|------|
| GET | `/stream/{id}/manifest.m3u8` | streaming-service | ✓ | Tier header eklenir |
| GET | `/stream/{id}/{quality}/playlist.m3u8` | streaming-service | ✓ | — |
| GET | `/stream/{id}/{quality}/segment_*.ts` | streaming-service | ✓ | Range request proxy |
| POST | `/api/stream/{id}/progress` | streaming-service | ✓ | — |
| GET | `/api/stream/{id}/progress` | streaming-service | ✓ | — |
| GET | `/api/stream/continue-watching` | streaming-service | ✓ | — |
| POST | `/api/stream/upload` | streaming-service | ✓ | Admin only, multipart |
| GET | `/api/encoding/jobs/{id}` | encoding-service | ✓ | Admin only |
| GET | `/api/encoding/jobs` | encoding-service | ✓ | Admin only |

### 6.2 Streaming Özel Middleware

```
streaming_auth middleware:
1. Standart JWT doğrulama (Faz 1'den mevcut)
2. User Service'e gRPC ile kullanıcının subscription tier'ını sor
   (veya JWT payload'ına tier ekle — daha performanslı)
3. X-User-Tier header'ını ekle
4. Streaming Service bu header'a göre kaliteleri filtreler

Multipart proxy:
- /api/stream/upload için Content-Type: multipart/form-data'yı
  olduğu gibi downstream'e iletir (buffering yapmaz, streaming proxy)
```

---

## 7. Altyapı Güncellemeleri

### 7.1 Docker Compose Eklentileri

```yaml
  # Faz 1 servislerinin hepsi aynen kalır, aşağıdakiler eklenir:

  rabbitmq:
    image: rabbitmq:3.13-management-alpine
    ports:
      - "5672:5672"    # AMQP
      - "15672:15672"  # Management UI
    environment:
      RABBITMQ_DEFAULT_USER: ${RABBITMQ_USER:-streamvault}
      RABBITMQ_DEFAULT_PASS: ${RABBITMQ_PASS:-secret}
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
      - ./infrastructure/rabbitmq/definitions.json:/etc/rabbitmq/definitions.json
      - ./infrastructure/rabbitmq/rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "-q", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"    # API
      - "9001:9001"    # Console
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY:-minioadmin}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY:-minioadmin}
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 10s
      timeout: 5s
      retries: 5

  minio-init:
    image: minio/mc:latest
    depends_on:
      minio:
        condition: service_healthy
    entrypoint: >
      /bin/sh -c "
      mc alias set sv http://minio:9000 minioadmin minioadmin;
      mc mb sv/streamvault-raw --ignore-existing;
      mc mb sv/streamvault-encoded --ignore-existing;
      mc mb sv/streamvault-thumbnails --ignore-existing;
      mc anonymous set download sv/streamvault-thumbnails;
      echo 'Buckets created successfully';
      "

  streaming-service:
    build: ./services/streaming-service
    ports: ["5003:5003", "50051:50051"]
    environment:
      - MINIO_ENDPOINT=minio:9000
      - MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}
      - MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}
      - REDIS_URL=redis://redis:6379/1
      - RABBITMQ_URL=amqp://${RABBITMQ_USER:-streamvault}:${RABBITMQ_PASS:-secret}@rabbitmq:5672/
      - CONSUL_ADDRESS=consul:8500
    depends_on:
      minio:
        condition: service_healthy
      redis:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
      consul:
        condition: service_started

  encoding-service:
    build: ./services/encoding-service
    environment:
      - MINIO_ENDPOINT=minio:9000
      - MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}
      - MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}
      - REDIS_URL=redis://redis:6379/2
      - RABBITMQ_URL=amqp://${RABBITMQ_USER:-streamvault}:${RABBITMQ_PASS:-secret}@rabbitmq:5672/
      - CONSUL_ADDRESS=consul:8500
    depends_on:
      minio:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
      consul:
        condition: service_started
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G    # FFmpeg bellek yoğun

volumes:
  # ... Faz 1 volume'ları +
  rabbitmq_data:
  minio_data:
```

### 7.2 Güncellenmiş Port Haritası

| Servis | İç Port | Dış Port | Protokol | Açıklama |
|--------|---------|----------|----------|----------|
| API Gateway | 8080 | 8080 | HTTP | Ana giriş noktası |
| User Service | 5001 | — | HTTP | Gateway üzerinden |
| Catalog Service | 5002 | — | HTTP | Gateway üzerinden |
| Streaming Service | 5003 | 5003 | HTTP | Video serving |
| Streaming Service | 50051 | 50051 | gRPC | Servisler arası |
| Encoding Service | 5004 | — | HTTP | Job status |
| Encoding Service | 50052 | 50052 | gRPC | Servisler arası |
| PostgreSQL | 5432 | 5432 | TCP | User DB |
| MongoDB | 27017 | 27017 | TCP | Catalog DB |
| Redis | 6379 | 6379 | TCP | Cache + sessions |
| Consul | 8500 | 8500 | HTTP | Service discovery UI |
| RabbitMQ | 5672 | 5672 | AMQP | Message broker |
| RabbitMQ UI | 15672 | 15672 | HTTP | Management UI |
| MinIO | 9000 | 9000 | HTTP | Object storage API |
| MinIO Console | 9001 | 9001 | HTTP | Object storage UI |

### 7.3 RabbitMQ Topology

```
Exchange: encoding (type: topic, durable: true)
  │
  ├── Binding: routing_key = "job.new"
  │   └── Queue: encoding.jobs (durable, prefetch=2)
  │       └── Consumer: Encoding Service
  │
  ├── Binding: routing_key = "job.completed"
  │   └── Queue: encoding.results.catalog (durable)
  │       └── Consumer: Catalog Service
  │
  └── Binding: routing_key = "job.failed"
      └── Queue: encoding.results.catalog (same queue, farklı routing key)
          └── Consumer: Catalog Service

Dead Letter Exchange: encoding.dlx (type: direct)
  └── Queue: encoding.dead-letters
      └── Başarısız mesajlar burada birikir (manuel inceleme)

definitions.json ile bu topology otomatik oluşturulur.
```

---

## 8. Uçtan Uca Akış

### 8.1 Video Upload ve Encoding

```
Admin (Postman/UI)
    │
    │  POST /api/stream/upload
    │  (multipart: file + contentId)
    ▼
┌──────────────┐
│  API Gateway │── JWT doğrula, Admin rolü kontrol
└──────┬───────┘
       │
       ▼
┌─────────────────────┐
│  Streaming Service  │
│  1. Dosyayı oku     │
│  2. MinIO'ya yükle  │    ──►  MinIO: streamvault-raw/{contentId}/original.mp4
│  3. Job oluştur     │
│  4. RabbitMQ'ya gön │    ──►  RabbitMQ: encoding.jobs queue
│  5. 202 Accepted    │
└─────────────────────┘
       │
       ▼  (Asenkron — RabbitMQ üzerinden)
┌─────────────────────┐
│  Encoding Service   │
│  1. Job'u al        │    ◄──  RabbitMQ: encoding.jobs queue
│  2. MinIO'dan indir │    ◄──  MinIO: streamvault-raw/...
│  3. FFmpeg transcode│    ──►  360p, 720p, 1080p dosyaları
│  4. HLS segment     │    ──►  .ts segmentler + .m3u8 playlist'ler
│  5. Thumbnail üret  │    ──►  poster.jpg, thumbnail.jpg
│  6. MinIO'ya yükle  │    ──►  MinIO: streamvault-encoded/{contentId}/...
│  7. Event publish   │    ──►  RabbitMQ: encoding.results
└─────────────────────┘
       │
       ▼  (Asenkron — RabbitMQ üzerinden)
┌─────────────────────┐
│  Catalog Service    │
│  1. Event'i dinle   │    ◄──  RabbitMQ: encoding.results.catalog queue
│  2. Movie'yi bul    │
│  3. VideoStatus =   │
│     Ready           │
│  4. StreamingInfo   │
│     güncelle        │
└─────────────────────┘
```

### 8.2 Video İzleme

```
Kullanıcı (Video Player)
    │
    │  GET /stream/{id}/manifest.m3u8
    ▼
┌──────────────┐
│  API Gateway │── JWT doğrula
│              │── X-User-Id, X-User-Tier ekle
└──────┬───────┘
       │
       ▼
┌─────────────────────┐
│  Streaming Service  │
│  1. Content ID'yi   │
│     kontrol et      │
│  2. Tier'e göre     │
│     kaliteleri      │
│     filtrele        │
│  3. Master playlist │
│     üret ve dön     │
└─────────────────────┘
       │
       ▼  (Player otomatik seçer)

    GET /stream/{id}/720p/playlist.m3u8
       │
       ▼
┌─────────────────────┐
│  Streaming Service  │
│  Media playlist dön │
└─────────────────────┘
       │
       ▼  (Player segmentleri sırayla ister)

    GET /stream/{id}/720p/segment_000.ts
    GET /stream/{id}/720p/segment_001.ts
    ...
       │
       ▼
┌─────────────────────┐
│  Streaming Service  │   ◄──►  MinIO: streamvault-encoded/...
│  MinIO'dan oku,     │
│  client'a ilet      │
└─────────────────────┘

    Her 15 saniyede bir:
    POST /api/stream/{id}/progress
    { "positionSeconds": 1847 }
       │
       ▼
┌─────────────────────┐
│  Streaming Service  │   ──►  Redis: progress:{userId}:{contentId}
│  Pozisyonu kaydet   │   ──►  Redis: continue:{userId} sorted set
└─────────────────────┘
```

---

## 9. Haftalık İlerleme Planı

### Hafta 1: Altyapı + Streaming Service

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Altyapı eklentileri | Docker Compose'a RabbitMQ ve MinIO ekle. Management UI'ların çalıştığını doğrula. Bucket'ları oluştur. RabbitMQ topology'yi definitions.json ile kur. |
| 2 | Proto dosyaları | `proto/` altında common, streaming, encoding proto'larını yaz. Go ve Rust için kod üretim script'ini hazırla (`generate-proto.sh`). |
| 3 | Streaming Service — Temel | Go projesi oluştur, config, Consul kaydı, health endpoint. MinIO client ile bağlantı. Basit bir dosya upload/download testi. |
| 4 | Streaming Service — HLS | Master ve media playlist üretimi, segment serving (MinIO'dan oku → client'a ilet). Statik test segmentleriyle çalıştığını doğrula. |
| 5 | Streaming Service — Progress | Redis'te izleme pozisyonu kaydetme/okuma, "continue watching" listesi, eşzamanlı izleme limiti. gRPC server implementasyonu. |

### Hafta 2: Encoding Service (Rust)

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Rust projesi kurulumu | Cargo projesi, Axum web framework, config, health endpoint. Dockerfile (FFmpeg dahil). Derleme ve çalışma testi. |
| 2 | RabbitMQ consumer | lapin crate ile RabbitMQ'ya bağlan, job queue'dan mesaj al, ack/nack mekanizması, dead letter queue. Basit bir echo consumer ile test. |
| 3 | FFmpeg pipeline — Transcode | `std::process::Command` ile FFmpeg çağırma, stdout progress parse, birden fazla kaliteye transcode. Tek bir test videoyla çalıştır. |
| 4 | FFmpeg pipeline — Segment + Thumbnail | HLS segmentlere ayırma, playlist üretimi, thumbnail extraction. Tüm sonuçları MinIO'ya yükleme. |
| 5 | Pipeline orchestrator | Tüm adımları birleştir: validate → transcode → segment → thumbnail → upload → notify. Error handling, retry, cleanup. EncodingCompleted event publish. |

### Hafta 3: Entegrasyon + Gateway Güncelleme

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Catalog Service güncelleme | VideoStatus ve StreamingInfo alanlarını ekle. RabbitMQ consumer ekle (EncodingCompleted event). streaming-info endpoint. |
| 2 | Gateway güncelleme | Streaming route'ları ekle. Tier bazlı middleware. Multipart upload proxy. Encoding admin route'ları. |
| 3 | Uçtan uca test | Tam akış: upload → encoding → catalog güncelleme → manifest → segment serve → progress kaydetme. `test-streaming.sh` script'i. |
| 4 | Error handling ve resilience | Encoding hata senaryoları (bozuk dosya, desteklenmeyen codec). Dead letter queue. Retry mekanizması. Timeout'lar. |
| 5 | Performans ve temizlik | MinIO'dan segment serving optimizasyonu (caching headers, range requests). Logging iyileştirmesi. README güncelleme. |

---

## 10. Faz 2 Bitiş Kriterleri (Definition of Done)

- [ ] `docker compose up` ile RabbitMQ ve MinIO dahil tüm sistem ayağa kalkıyor
- [ ] RabbitMQ Management UI'da exchange, queue ve binding'ler görünüyor
- [ ] MinIO Console'da 3 bucket (raw, encoded, thumbnails) mevcut
- [ ] Admin bir video dosyası upload edebiliyor (POST /api/stream/upload → 202)
- [ ] Encoding Service job'u otomatik alıyor ve işlemeye başlıyor
- [ ] FFmpeg ile en az 2 kaliteye (360p, 720p) başarılı transcode yapılıyor
- [ ] HLS segmentler (.ts) ve playlist'ler (.m3u8) MinIO'da doğru yapıda oluşuyor
- [ ] Thumbnail ve poster görselleri üretiliyor
- [ ] Encoding tamamlandığında Catalog Service otomatik güncelleniyor (VideoStatus: Ready)
- [ ] Master playlist kullanıcının tier'ine göre doğru kaliteleri döndürüyor
- [ ] Video segmentler HLS uyumlu bir player'da (hls.js, VLC) oynatılabiliyor
- [ ] İzleme pozisyonu kaydediliyor ve "kaldığın yerden devam et" çalışıyor
- [ ] Eşzamanlı izleme limiti çalışıyor (2. cihazda Basic kullanıcı engelleniyorsa)
- [ ] Encoding hata durumunda dead letter queue'ya düşüyor, Catalog'da status "Error" oluyor
- [ ] gRPC ile Streaming Service'ten streaming bilgisi sorgulanabiliyor
- [ ] Proto dosyalarından Go ve Rust kodu başarıyla üretiliyor
- [ ] Tüm yeni servisler Consul'da "healthy" görünüyor
- [ ] Encoding job ilerleme yüzdesi Redis üzerinden takip edilebiliyor

---

## 11. Test Videoları

Geliştirme ve test için kullanılacak videolar:

```bash
# FFmpeg ile test videoları üretme (gerçek video indirmeye gerek yok)

# 30 saniyelik test videosu (renk barları + sayaç)
ffmpeg -f lavfi -i testsrc=duration=30:size=1920x1080:rate=30 \
       -f lavfi -i sine=frequency=440:duration=30 \
       -c:v libx264 -preset fast -c:a aac \
       test-30s-1080p.mp4

# 5 dakikalık daha uzun test videosu
ffmpeg -f lavfi -i testsrc2=duration=300:size=1920x1080:rate=30 \
       -f lavfi -i sine=frequency=440:duration=300 \
       -c:v libx264 -preset fast -c:a aac \
       test-5min-1080p.mp4

# 4K test videosu (Premium tier testi için)
ffmpeg -f lavfi -i testsrc=duration=30:size=3840x2160:rate=30 \
       -c:v libx264 -preset ultrafast -c:a aac \
       test-30s-4k.mp4
```

---

## 12. Faz 3'e Hazırlık

Faz 2 tamamlandığında, Faz 3 (Smart Features) için şu temeller hazır olacak:

- **RabbitMQ altyapısı hazır:** Yeni exchange/queue eklemek kolay. Search Service ve Recommendation Engine aynı event bus'ı kullanacak.
- **Event-driven pattern yerleşmiş:** ContentUploaded, EncodingCompleted event'leri zaten akıyor. Faz 3'te ContentAdded, WatchCompleted gibi yeni event'ler eklenecek.
- **gRPC altyapısı hazır:** Proto repo mevcut, kod üretim script'i çalışıyor. Recommendation Engine'in gRPC'si hızla eklenebilir.
- **İzleme verisi birikmeye başlamış:** progress ve continue-watching verileri Recommendation Engine'in input'u olacak.
- **MinIO ve Redis pattern'leri oturmuş:** Elasticsearch cache'i Redis'e, Search index snapshot'ları MinIO'ya yazılabilir.

> **Not:** Faz 3'te en büyük zorluk CQRS pattern'i (Catalog yazıyor → Search okuyor) ve Saga pattern'i (Subscription akışı) olacak. Bunlar mimari açıdan Faz 2'den daha karmaşık ama altyapı artık hazır.
