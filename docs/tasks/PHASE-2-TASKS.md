# StreamVault — Faz 2: Core Streaming Görevleri

> **Durum:** Faz 1 tamamlandı. Faz 2 ile video upload → transcode → HLS serve akışı, gRPC servisler arası iletişim ve RabbitMQ event pipeline kurulacak.

---

## 1. Altyapı Eklentileri

### 1.1 RabbitMQ
- [x] Docker Compose'a RabbitMQ servisi ekle (rabbitmq:3.13-management-alpine)
- [x] `infrastructure/rabbitmq/rabbitmq.conf` oluştur
- [x] `infrastructure/rabbitmq/definitions.json` ile topology tanımla (exchange, queue, binding)
- [x] Exchange: `encoding` (type: topic, durable)
- [x] Queue: `encoding.jobs` (durable, prefetch=2, routing_key: `job.new`)
- [x] Queue: `encoding.results.catalog` (durable, routing_key: `job.completed` + `job.failed`)
- [x] Dead Letter Exchange: `encoding.dlx` + `encoding.dead-letters` queue
- [x] RabbitMQ Management UI'da (15672) topology'nin doğru göründüğünü doğrula
- [x] Healthcheck çalışıyor (`rabbitmq-diagnostics -q ping`)

### 1.2 MinIO
- [x] Docker Compose'a MinIO servisi ekle (minio/minio:latest)
- [x] `infrastructure/minio/init-buckets.sh` ile bucket oluşturma script'i yaz
- [x] `minio-init` servisi ekle (mc ile bucket oluşturma)
- [x] 3 bucket oluştur: `streamvault-raw`, `streamvault-encoded`, `streamvault-thumbnails`
- [x] `streamvault-thumbnails` için anonymous download izni ayarla
- [x] MinIO Console'da (9001) bucket'ların göründüğünü doğrula
- [x] Healthcheck çalışıyor (`curl -f http://localhost:9000/minio/health/live`)

### 1.3 Docker Compose Güncellemeleri
- [x] `rabbitmq_data` ve `minio_data` volume'larını ekle
- [ ] Streaming Service container tanımı ekle (port: 5003, 50051)
- [ ] Encoding Service container tanımı ekle (CPU: 2.0, RAM: 2G limiti)
- [x] Servis bağımlılıklarını (depends_on + condition) doğru kur
- [x] Environment variable'ları ekle (MINIO, RABBITMQ, REDIS, CONSUL)
- [x] `docker compose up` ile tüm yeni altyapı servislerinin ayağa kalktığını doğrula

---

## 2. Proto Dosyaları & gRPC

- [ ] `proto/common/v1/common.proto` oluştur (Pagination, ContentType, SubscriptionTier)
- [ ] `proto/streaming/v1/streaming.proto` oluştur (StreamingService, GetStreamingInfo, GetProgress, GetContinueWatching)
- [ ] `proto/encoding/v1/encoding.proto` oluştur (EncodingService, GetJobStatus, ListJobs)
- [ ] `scripts/generate-proto.sh` güncelle — Go ve Rust için kod üretimi
- [ ] Go proto üretiminin çalıştığını doğrula (`protoc` + `protoc-gen-go`, `protoc-gen-go-grpc`)
- [ ] Rust proto üretiminin çalıştığını doğrula (`tonic-build`)

---

## 3. Streaming Service (Go)

### 3.1 Proje Kurulumu & Temel Yapı
- [ ] Go projesi oluştur (`services/streaming-service/`)
- [ ] `go.mod` oluştur, dependency'leri ekle (minio-go, go-redis, amqp091-go, grpc, consul api)
- [ ] `internal/config/config.go` — Viper ile config.yaml + env var yükleme
- [ ] `cmd/streaming/main.go` — HTTP + gRPC server bootstrap
- [ ] Dockerfile oluştur (multi-stage build)
- [ ] Makefile oluştur (build, test, lint)
- [ ] `/health` endpoint'i (MinIO, Redis bağlantı durumu)

### 3.2 Consul Entegrasyonu
- [ ] `internal/discovery/consul.go` — Consul'a self-registration
- [ ] Health check kaydı
- [ ] Graceful shutdown'da deregistration
- [ ] Consul UI'da "streaming-service" healthy göründüğünü doğrula

### 3.3 MinIO Storage
- [ ] `internal/storage/interface.go` — Storage interface tanımı
- [ ] `internal/storage/minio.go` — MinIO client wrapper (upload, download, list, delete)
- [ ] Raw bucket'a dosya yükleme testi
- [ ] Encoded bucket'tan dosya okuma testi

### 3.4 Video Upload (Admin)
- [ ] `internal/handler/upload.go` — POST /api/stream/upload (multipart/form-data)
- [ ] Dosyayı MinIO `streamvault-raw/{contentId}/original.mp4` konumuna yükle
- [ ] Admin rolü kontrolü (X-User-Role header)
- [ ] RabbitMQ'ya encoding job mesajı gönder
- [ ] 202 Accepted response (jobId, contentId, status)

### 3.5 HLS Manifest & Segment Serving
- [ ] `internal/hls/master_playlist.go` — Master playlist üretimi (multi-quality, tier filtreli)
- [ ] `internal/hls/media_playlist.go` — Tek kalite playlist üretimi
- [ ] `internal/handler/manifest.go` — GET /stream/{contentId}/manifest.m3u8
- [ ] `internal/handler/chunk.go` — GET /stream/{contentId}/{quality}/segment_{number}.ts
- [ ] Tier'e göre kalite filtreleme (Basic: 360p+720p, Standard: +1080p, Premium: +4K)
- [ ] MinIO'dan segment okuyup client'a proxy
- [ ] Content-Type header'ları (application/vnd.apple.mpegurl, video/mp2t)
- [ ] Accept-Ranges ve Content-Length header'ları

### 3.6 İzleme Pozisyonu (Progress)
- [ ] `internal/progress/repository.go` — Redis'te izleme pozisyonu (Hash: `progress:{userId}:{contentId}`)
- [ ] `internal/progress/service.go` — İş mantığı (kaydet, oku, continue-watching listesi)
- [ ] `internal/handler/progress.go` — POST /api/stream/{contentId}/progress (pozisyon kaydet)
- [ ] GET /api/stream/{contentId}/progress (pozisyon oku)
- [ ] GET /api/stream/continue-watching (sorted set'ten liste)
- [ ] Tamamlanmış içerikleri (%95+) listeden düşür
- [ ] Redis TTL ayarları (progress: 90 gün, concurrent: 5 dakika)

### 3.7 Eşzamanlı İzleme Limiti
- [ ] `internal/middleware/concurrent.go` — Redis Set ile aktif session takibi
- [ ] Tier bazlı limit kontrolü (Basic: 1, Standard: 2, Premium: 4)
- [ ] Heartbeat mekanizması (TTL yenileme)
- [ ] Limit aşımında hata dönüşü

### 3.8 Abonelik Kontrolü
- [ ] `internal/middleware/subscription.go` — X-User-Tier header kontrolü
- [ ] Kalite bazlı erişim kontrolü

### 3.9 gRPC Server
- [ ] `internal/grpc/server.go` — gRPC server implementasyonu (port: 50051)
- [ ] GetStreamingInfo RPC — içerik streaming bilgisi
- [ ] GetProgress RPC — kullanıcı izleme pozisyonu
- [ ] GetContinueWatching RPC — devam eden izlemeler listesi
- [ ] `internal/grpc/catalog_client.go` — Catalog Service gRPC client (opsiyonel)

### 3.10 RabbitMQ Publisher
- [ ] Encoding job publish (encoding exchange, routing_key: job.new)
- [ ] EncodingCompleted/EncodingFailed event consume (result queue'dan)
- [ ] Stream info cache'i (Redis) güncelle

---

## 4. Encoding Service (Rust)

### 4.1 Proje Kurulumu & Temel Yapı
- [ ] Cargo projesi oluştur (`services/encoding-service/`)
- [ ] `Cargo.toml` — dependency'ler (axum, tokio, lapin, aws-sdk-s3/minio, tonic, serde, tracing)
- [ ] `src/config.rs` — config.toml + env var yükleme
- [ ] `src/main.rs` — Axum bootstrap + RabbitMQ consumer başlatma
- [ ] `src/error.rs` — Hata tipleri (thiserror)
- [ ] Dockerfile oluştur (multi-stage build, FFmpeg dahil)
- [ ] Makefile oluştur (build, test, lint)
- [ ] GET /health endpoint (RabbitMQ, MinIO, FFmpeg bağlantı durumu)

### 4.2 Consul Entegrasyonu
- [ ] `src/main.rs` içinde Consul'a self-registration
- [ ] Health check kaydı
- [ ] Graceful shutdown'da deregistration

### 4.3 RabbitMQ Consumer
- [ ] `src/queue/consumer.rs` — lapin ile RabbitMQ consumer
- [ ] `encoding.jobs` queue'dan job mesajı alma
- [ ] Ack/Nack mekanizması
- [ ] Dead letter queue'ya düşen başarısız mesajlar
- [ ] Prefetch count: 2 (aynı anda max 2 job)
- [ ] Retry mekanizması (max 3, delay 5s)

### 4.4 RabbitMQ Publisher
- [ ] `src/queue/publisher.rs` — event publish
- [ ] EncodingCompleted event (encoding.results exchange, routing_key: job.completed)
- [ ] EncodingFailed event (routing_key: job.failed)
- [ ] Mesaj formatları JSON (jobId, contentId, outputs, thumbnails, duration)

### 4.5 MinIO Storage
- [ ] `src/storage/minio.rs` — S3 client wrapper
- [ ] Raw bucket'tan ham video indirme (temp dizine)
- [ ] Encoded bucket'a segment yükleme
- [ ] Thumbnails bucket'a görsel yükleme
- [ ] Büyük dosya streaming I/O

### 4.6 Models
- [ ] `src/models/job.rs` — EncodingJob struct
- [ ] `src/models/profile.rs` — Encoding profilleri (360p, 720p, 1080p, 4K)
- [ ] `src/models/status.rs` — Job durumları (Queued, Processing, Completed, Failed, Cancelled)

### 4.7 Pipeline — Validator
- [ ] `src/pipeline/validator.rs` — Video format doğrulama
- [ ] ffprobe ile video bilgilerini oku (codec, çözünürlük, süre, boyut)
- [ ] Desteklenen format kontrolü
- [ ] Max dosya boyutu kontrolü (10 GB)
- [ ] Kaynak çözünürlüğe göre hedef profilleri belirle (720p kaynak → sadece 360p + 720p)

### 4.8 Pipeline — Transcoder
- [ ] `src/pipeline/transcoder.rs` — FFmpeg ile transcoding
- [ ] `std::process::Command` ile FFmpeg çağırma
- [ ] Her profil için ayrı transcode (H.264/libx264, AAC 128kbps)
- [ ] FFmpeg stdout parse → ilerleme yüzdesi Redis'e yaz
- [ ] Paralel transcoding (profiller arası)

### 4.9 Pipeline — Segmenter
- [ ] `src/pipeline/segmenter.rs` — HLS segment üretimi
- [ ] Her kalite dosyasını 10 saniyelik .ts segmentlerine ayır
- [ ] Her kalite için playlist.m3u8 üret
- [ ] Segment numaralandırma (segment_000.ts, segment_001.ts, ...)

### 4.10 Pipeline — Thumbnail
- [ ] `src/pipeline/thumbnail.rs` — Thumbnail ve poster üretimi
- [ ] Videonun %10, %30, %50, %70, %90 noktalarından kare çıkar
- [ ] Poster (yüksek çözünürlük) ve thumbnail (300x170, 600x340) resize
- [ ] Timeline preview görselleri

### 4.11 Pipeline — Uploader
- [ ] `src/pipeline/uploader.rs` — Sonuçları MinIO'ya yükleme
- [ ] Tüm segmentleri `streamvault-encoded/{contentId}/{quality}/` altına yükle
- [ ] Playlist dosyalarını yükle
- [ ] Thumbnail'leri `streamvault-thumbnails/{contentId}/` altına yükle
- [ ] Temp dosyaları temizle

### 4.12 Pipeline — Orchestrator
- [ ] `src/pipeline/orchestrator.rs` — Pipeline adımlarını sırayla çalıştır
- [ ] Akış: validate → transcode → segment → thumbnail → upload → notify
- [ ] Her adımda Redis'e progress güncelle
- [ ] Error handling: hata durumunda cleanup ve fail event publish
- [ ] Job durumunu güncelle (Processing → Completed/Failed)

### 4.13 HTTP API
- [ ] `src/api/routes.rs` — HTTP endpoint tanımları
- [ ] `src/api/handlers.rs` — Handler implementasyonları
- [ ] GET /api/encoding/jobs/{jobId} — Job durum sorgulama
- [ ] GET /api/encoding/jobs?status=processing&limit=10 — Job listeleme
- [ ] Response formatları (jobId, status, progressPercentage, currentStep, outputs)

### 4.14 gRPC Server
- [ ] gRPC server implementasyonu (port: 50052)
- [ ] GetJobStatus RPC
- [ ] ListJobs RPC

---

## 5. Catalog Service Güncellemeleri

### 5.1 Domain Değişiklikleri
- [ ] `VideoStatus` enum ekle (NotUploaded, Uploading, Queued, Encoding, Ready, Error)
- [ ] `StreamingInfo` sınıfı ekle (DurationSeconds, AvailableQualities, ManifestPath, ThumbnailPath, PosterPath, EncodedAt)
- [ ] `QualityInfo` sınıfı ekle (Label, Width, Height, BitrateKbps, SegmentCount)
- [ ] `Movie` entity'sine `VideoStatus` ve `StreamingInfo?` alanlarını ekle
- [ ] MongoDB BsonClassMap güncellemesi

### 5.2 RabbitMQ Entegrasyonu
- [ ] `CatalogService.Infrastructure/Messaging/RabbitMqEventPublisher.cs` oluştur
- [ ] `CatalogService.Application/Events/IEventPublisher.cs` interface tanımla
- [ ] `CatalogService.Application/Events/ContentUploadedEvent.cs` oluştur
- [ ] RabbitMQ consumer ekle — `encoding.results.catalog` queue dinle
- [ ] EncodingCompleted event'i handle et: Movie bul, VideoStatus = Ready, StreamingInfo güncelle
- [ ] EncodingFailed event'i handle et: VideoStatus = Error

### 5.3 Yeni Command & Handler
- [ ] `UpdateVideoStatusCommand.cs` oluştur
- [ ] `UpdateVideoStatusHandler.cs` — MediatR handler
- [ ] Validasyon: geçerli content ID, geçerli status geçişi

### 5.4 Yeni Endpoint
- [ ] GET /api/catalog/movies/{id}/streaming-info — Video streaming bilgisi
- [ ] Response: videoStatus, durationSeconds, availableQualities, manifestUrl, thumbnailUrl
- [ ] 404: "Video henüz yüklenmemiş"

### 5.5 NuGet Paketleri
- [ ] RabbitMQ.Client paketi ekle
- [ ] DI konfigürasyonu (RabbitMQ connection, consumer registration)

---

## 6. API Gateway Güncellemeleri

### 6.1 Yeni Route'lar
- [ ] Streaming route'ları ekle: `/stream/{id}/manifest.m3u8`, `/stream/{id}/{quality}/playlist.m3u8`, `/stream/{id}/{quality}/segment_*.ts`
- [ ] Progress route'ları: `/api/stream/{id}/progress` (GET+POST), `/api/stream/continue-watching`
- [ ] Upload route'u: `/api/stream/upload` (Admin only, multipart proxy)
- [ ] Encoding route'ları: `/api/encoding/jobs/{id}`, `/api/encoding/jobs` (Admin only)

### 6.2 Streaming Middleware
- [ ] `internal/middleware/streaming_auth.go` — Subscription tier kontrolü
- [ ] JWT payload'dan veya User Service'ten tier bilgisi al
- [ ] `X-User-Tier` header'ını downstream'e ekle
- [ ] Admin route'lar için rol kontrolü

### 6.3 Streaming Proxy
- [ ] `internal/proxy/streaming_proxy.go` — Streaming özel proxy
- [ ] Range request desteği (Accept-Ranges header forwarding)
- [ ] Multipart upload proxy (buffering yapmadan streaming)
- [ ] Chunked transfer encoding desteği

### 6.4 Config Güncelleme
- [ ] `gateway/config.yaml` — Streaming ve Encoding servis tanımları
- [ ] Consul'dan streaming-service ve encoding-service çözümleme

---

## 7. Uçtan Uca Entegrasyon & Test

### 7.1 Upload → Encode → Serve Akışı
- [ ] `scripts/upload-test-video.sh` — Test video yükleme script'i
- [ ] FFmpeg ile test videosu üret (30s, 1080p)
- [ ] Admin olarak video upload et (POST /api/stream/upload → 202)
- [ ] Encoding Service'in job'u otomatik aldığını doğrula
- [ ] FFmpeg ile en az 2 kaliteye (360p, 720p) başarılı transcode
- [ ] HLS segmentler (.ts) ve playlist'ler (.m3u8) MinIO'da doğru yapıda oluşuyor
- [ ] Thumbnail ve poster görselleri üretiliyor
- [ ] Encoding tamamlandığında Catalog Service otomatik güncelleniyor (VideoStatus: Ready)

### 7.2 Streaming Testi
- [ ] `scripts/test-streaming.sh` — Streaming akışı test script'i
- [ ] Master playlist tier'e göre doğru kaliteleri döndürüyor
- [ ] Video segmentler HLS uyumlu player'da oynatılabiliyor (hls.js, VLC)
- [ ] İzleme pozisyonu kaydediliyor
- [ ] "Kaldığın Yerden Devam Et" listesi çalışıyor
- [ ] Eşzamanlı izleme limiti çalışıyor

### 7.3 Hata Senaryoları
- [ ] Bozuk/desteklenmeyen video dosyası → dead letter queue
- [ ] Encoding hata → Catalog'da VideoStatus = Error
- [ ] MinIO bağlantı kesintisi → graceful error
- [ ] RabbitMQ bağlantı kesintisi → retry mekanizması

### 7.4 Consul & Altyapı
- [ ] Tüm yeni servisler Consul'da "healthy" görünüyor
- [ ] Encoding job ilerleme yüzdesi Redis üzerinden takip edilebiliyor
- [ ] Proto dosyalarından Go ve Rust kodu başarıyla üretiliyor

### 7.5 Unit & Integration Tests
- [ ] Streaming Service: handler testleri, HLS playlist üretimi, progress logic
- [ ] Encoding Service: pipeline adımları, RabbitMQ consumer/publisher, model testleri
- [ ] Gateway: yeni route ve middleware testleri
- [ ] Catalog Service: UpdateVideoStatus handler, RabbitMQ consumer testleri

---

## 8. Dokümantasyon

- [ ] README.md güncelle: Faz 2 mimari diyagram, yeni servisler, port haritası
- [ ] Streaming API kullanım örnekleri (curl komutları)
- [ ] Encoding API kullanım örnekleri
- [ ] RabbitMQ topology açıklaması
- [ ] MinIO bucket yapısı açıklaması

---

## Önerilen Sıralama

1. **Altyapı** → Docker Compose'a RabbitMQ ve MinIO ekle, topology kur, bucket oluştur
2. **Proto dosyaları** → gRPC tanımları yaz, kod üretim script'ini hazırla
3. **Streaming Service — Temel** → Go projesi, config, Consul, health, MinIO client
4. **Streaming Service — HLS** → Manifest üretimi, segment serving
5. **Streaming Service — Progress & Upload** → Redis'te izleme pozisyonu, video upload, RabbitMQ publish
6. **Encoding Service — Temel** → Rust projesi, config, health, Dockerfile (FFmpeg dahil)
7. **Encoding Service — RabbitMQ** → Consumer, publisher, mesaj formatları
8. **Encoding Service — Pipeline** → FFmpeg transcode, segment, thumbnail, orchestrator
9. **Catalog Service** → Domain güncellemeleri, RabbitMQ consumer, streaming-info endpoint
10. **Gateway** → Yeni route'lar, streaming middleware, multipart proxy
11. **Uçtan Uca Test** → Tam akış doğrulama, test script'leri
12. **Hata Senaryoları** → Dead letter, retry, timeout
13. **Dokümantasyon** → README, API örnekleri
