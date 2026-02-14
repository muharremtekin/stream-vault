# StreamVault — RULES.md Uyumluluk Düzeltme Görevleri

> **Durum:** RULES.md taraması sonucu tespit edilen ~81 ihlal için düzeltme görevleri.
> **Öncelik:** KRİTİK > YÜKSEK > ORTA > DÜŞÜK
> **Tarih:** 2026-02-13

---

## 1. [KRİTİK] Environment & Secret Düzeltmeleri (Kural 1.1)

### 1.1 docker-compose.yml — Secret Default Değerleri Kaldır
- [x] `POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-secret}` → `${POSTGRES_PASSWORD}` (default kaldır)
- [x] `RABBITMQ_DEFAULT_PASS: ${RABBITMQ_PASS:-secret}` → `${RABBITMQ_PASS}` (default kaldır)
- [x] `MINIO_ROOT_USER: ${MINIO_ACCESS_KEY:-minioadmin}` → `${MINIO_ACCESS_KEY}` (2 yerde)
- [x] `MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY:-minioadmin}` → `${MINIO_SECRET_KEY}` (2 yerde)
- [x] `ConnectionStrings__DefaultConnection` içindeki `${POSTGRES_PASSWORD:-secret}` → `${POSTGRES_PASSWORD}`
- [x] `RabbitMQ__ConnectionString` içindeki `${RABBITMQ_PASS:-secret}` → `${RABBITMQ_PASS}`
- [x] `STREAMING_MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}` → `${MINIO_ACCESS_KEY}`
- [x] `STREAMING_MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}` → `${MINIO_SECRET_KEY}`
- [x] `STREAMING_RABBITMQ_URL` içindeki `${RABBITMQ_PASS:-secret}` → `${RABBITMQ_PASS}`
- [x] `ENCODING_MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}` → `${MINIO_ACCESS_KEY}`
- [x] `ENCODING_MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}` → `${MINIO_SECRET_KEY}`
- [x] `ENCODING_RABBITMQ_URL` içindeki `${RABBITMQ_PASS:-secret}` → `${RABBITMQ_PASS}`
- [x] `RABBITMQ_DEFAULT_USER: ${RABBITMQ_USER:-streamvault}` → username de kontrol et, secret değilse OK

### 1.2 Config Dosyalarından Hardcoded Secret'ları Temizle
- [x] `services/user-service/src/UserService.Api/appsettings.json` — `ConnectionStrings.DefaultConnection` içindeki `Password=postgres` kaldır, placeholder yap veya boş bırak
- [x] `services/user-service/src/UserService.Api/appsettings.json` — `Jwt.Secret` değerini kaldır, environment variable'dan override edilecek şekilde düzenle (boş string veya placeholder)
- [x] `services/encoding-service/config.toml` — `rabbitmq.url = "amqp://guest:guest@localhost:5672/%2f"` → credential'sız placeholder yap
- [x] `services/catalog-service/src/CatalogService.Infrastructure/Messaging/EncodingResultConsumer.cs` — `?? "amqp://streamvault:secret@localhost:5672/"` fallback'indeki credential kaldır

### 1.3 Code Default'larından Credential'ları Temizle
- [x] `services/encoding-service/src/config.rs` — `set_default("rabbitmq.url", "amqp://guest:guest@localhost:5672/")` → credential olmadan `"amqp://localhost:5672/"` yap
- [x] `services/streaming-service/internal/config/config.go` — `SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")` → credential olmadan
- [x] Tüm servislerin config default'larını tarayıp credential içerenleri temizle

### 1.4 .env ve .env.example Güncelle
- [x] `.env` dosyasında tüm secret'ların dolu olduğunu doğrula (docker-compose artık default kullanmayacak)
- [x] `.env.example` dosyasını güncelle — yeni eklenen tüm environment variable'lar için placeholder ekle
- [x] `.env.example`'da secret alanlarının boş bırakıldığını doğrula

---

## 2. [YÜKSEK] Docker & Altyapı Düzeltmeleri (Kural 7)

### 2.1 `latest` Tag'lerini Spesifik Versiyona Çevir (Kural 7.1)
- [x] `docker-compose.yml` — `minio/minio:latest` → `minio/minio:RELEASE.2025-09-07T16-13-09Z`
- [x] `docker-compose.yml` — `minio/mc:latest` → `minio/mc:RELEASE.2025-08-13T08-35-41Z`

### 2.2 Tüm Servislere `restart` Policy Ekle (Kural 7.2)
- [x] `postgres` servisine `restart: unless-stopped` ekle
- [x] `mongo` servisine `restart: unless-stopped` ekle
- [x] `redis` servisine `restart: unless-stopped` ekle
- [x] `consul` servisine `restart: unless-stopped` ekle
- [x] `rabbitmq` servisine `restart: unless-stopped` ekle
- [x] `minio` servisine `restart: unless-stopped` ekle
- [x] `gateway` servisine `restart: unless-stopped` ekle
- [x] `user-service` servisine `restart: unless-stopped` ekle
- [x] `catalog-service` servisine `restart: unless-stopped` ekle
- [x] `streaming-service` servisine `restart: unless-stopped` ekle
- [x] `encoding-service` servisine `restart: unless-stopped` ekle

### 2.3 Application Servislerine `healthcheck` Ekle (Kural 7.2)
- [x] `gateway` — healthcheck tanımla (wget ile `/health`)
- [x] `user-service` — healthcheck tanımla
- [x] `catalog-service` — healthcheck tanımla
- [x] `streaming-service` — healthcheck tanımla
- [x] `encoding-service` — healthcheck tanımla

### 2.4 `depends_on` Koşullarını Düzelt (Kural 7.2)
- [x] `gateway` → `consul`: `service_started` → `service_healthy` yap
- [x] `user-service` → `consul`: `service_started` → `service_healthy` yap
- [x] `catalog-service` → `consul`: `service_started` → `service_healthy` yap
- [x] `streaming-service` → `consul`: `service_started` → `service_healthy` yap
- [x] Consul servisine healthcheck tanımla (prerequisite)

### 2.5 Eksik `.dockerignore` Ekle (Kural 7.1)
- [x] `services/streaming-service/.dockerignore` oluştur (Go standart: `.git`, `*.test`, `vendor/`, `tmp/`, etc.)

---

## 3. [YÜKSEK] Observability Endpoint'leri (Kural 13.1)

### 3.1 Gateway — Health Endpoint'leri Ayır
- [x] `/health` → `/health/live` (liveness — sadece process alive)
- [x] `/health/ready` ekle (readiness — downstream bağımlılık kontrolü)
- [ ] `/metrics` endpoint'i ekle (Prometheus format) — Faz 4'e bırakıldı

### 3.2 User Service — Health Endpoint'leri Ayır
- [x] `/health` → `/health/live` + `/health/ready` olarak ayır
- [ ] `/metrics` endpoint'i ekle — Faz 4'e bırakıldı

### 3.3 Catalog Service — Health Endpoint'leri Ayır
- [x] `/health` → `/health/live` + `/health/ready` olarak ayır
- [ ] `/metrics` endpoint'i ekle — Faz 4'e bırakıldı

### 3.4 Streaming Service — Health Endpoint'leri Ayır
- [x] `/health` → `/health/live` + `/health/ready` olarak ayır
- [ ] `/metrics` endpoint'i ekle — Faz 4'e bırakıldı

### 3.5 Encoding Service — Health Endpoint'leri Ayır
- [x] `/health` → `/health/live` + `/health/ready` olarak ayır
- [ ] `/metrics` endpoint'i ekle — Faz 4'e bırakıldı

> **Not:** Bu görevler Faz 4 (Production-Ready) planıyla örtüşüyor. Faz 4'te daha kapsamlı yapılacaksa sadece `/health/live` ve `/health/ready` ayrımı şimdi yapılıp `/metrics` Faz 4'e bırakılabilir.

---

## 4. [YÜKSEK] Event Payload Düzeltmeleri (Kural 3.4)

### 4.1 Encoding Service — Event Envelope Ekle
- [x] `services/encoding-service/src/messaging/models.rs` — `EncodingResult` struct'ına zorunlu envelope alanları ekle:
  - `event_id: String` (UUID)
  - `event_type: String` (örn: `"encoding.job.completed"`)
  - `timestamp: String` (ISO 8601 UTC)
  - `source: String` (`"encoding-service"`)
  - `correlation_id: String` (trace-id)
- [x] `services/encoding-service/src/pipeline/orchestrator.rs` — Publish sırasında envelope alanlarını doldur

### 4.2 Catalog Service — Consumer'ı Güncel Envelope'a Uyumla
- [x] `services/catalog-service/src/CatalogService.Infrastructure/Messaging/EncodingResultMessage.cs` — Yeni envelope alanlarını deserialize edecek şekilde güncelle
- [x] Consumer'da `eventType`, `source`, `correlationId` alanlarını loglama ve trace'e ekle

---

## 5. [ORTA] Güvenlik Düzeltmeleri (Kural 8.1)

### 5.1 JWT Payload'dan Kişisel Veriyi Kaldır
- [x] `services/user-service/src/UserService.Infrastructure/Services/JwtTokenService.cs` — JWT claim'lerden `email` alanını kaldır
- [x] Gateway veya downstream servislerde email claim'e bağımlılık varsa kontrol et ve düzelt
- [x] Email gerekiyorsa `sub` (userId) üzerinden servis çağrısıyla alınmalı

---

## 6. [ORTA] DB Naming Convention Düzeltmeleri (Kural 4.2)

### 6.1 EF Core Constraint/Index İsimlendirmesini Düzelt
- [x] Yeni migration oluştur: constraint ve index isimlerini kurala uygun hale getir
  - `PK_users` → `pk_users`
  - `PK_profiles` → `pk_profiles`
  - `PK_refresh_tokens` → `pk_refresh_tokens`
  - `PK_watchlist_items` → `pk_watchlist_items`
  - `FK_profiles_users_user_id` → `fk_profiles_user_id`
  - `FK_refresh_tokens_users_user_id` → `fk_refresh_tokens_user_id`
  - `FK_watchlist_items_profiles_profile_id` → `fk_watchlist_items_profile_id`
  - `IX_profiles_user_id` → `idx_profiles_user_id`
  - `IX_refresh_tokens_token` → `idx_refresh_tokens_token`
  - `IX_refresh_tokens_user_id` → `idx_refresh_tokens_user_id`
  - `IX_users_email` → `idx_users_email`
  - `IX_watchlist_items_profile_id_content_id` → `idx_watchlist_items_profile_id_content_id`
- [x] EF Core'da `HasName()` ile future migration'lar için convention'ı override et (veya `IModelCustomizer` ile global naming convention uygula)

---

## 7. [ORTA] .gitignore Eksik Pattern'leri Ekle (Kural 2.3)

### 7.1 Eksik Pattern'leri Ekle
- [ ] `*.pem` ekle
- [ ] `*.key` ekle
- [ ] `*.p12` ekle
- [ ] `**/target/` ekle (Rust build output)
- [ ] `**/node_modules/` ekle
- [ ] `**/.next/` ekle
- [ ] `data/` ekle
- [ ] `*_data/` ekle
- [ ] `tmp/` ekle
- [ ] `temp/` ekle
- [ ] `*.tmp` ekle

---

## 8. [DÜŞÜK] Git Commit Mesajı Düzeltmesi (Kural 10.2)

### 8.1 Format Dışı Commit
- [ ] `7a6aa89` — `add phase 3, 4 and 5 plans docs` — bu commit geçmişte kaldığı için düzeltilemez (rewrite gerekir), ileride dikkat edilmesi yeterli
- [ ] Gelecek commit'ler için `{type}({scope}): {description}` formatına uyulması hatırlatılmalı

---

## Özet

| Bölüm | Görev Sayısı | Öncelik | Tahmini Etki |
|-------|:------------:|---------|-------------|
| 1. Secret/Env Düzeltmeleri | ~20 | KRİTİK | Güvenlik açığını kapatır |
| 2. Docker Altyapı | ~22 | YÜKSEK | Stabilite ve dayanıklılık |
| 3. Health Endpoint'leri | ~15 | YÜKSEK | Observability (Faz 4 ile örtüşür) |
| 4. Event Payload | ~4 | YÜKSEK | Servisler arası uyumluluk |
| 5. JWT Güvenlik | ~3 | ORTA | Kişisel veri koruması |
| 6. DB Naming | ~12 | ORTA | Konvansiyon tutarlılığı |
| 7. .gitignore | ~11 | ORTA | Repo temizliği |
| 8. Git Commit | ~1 | DÜŞÜK | Sadece hatırlatma |
| **TOPLAM** | **~88** | | |

---

> **Uygulama Stratejisi:**
> 1. Bölüm 1 (Secret/Env) en önce yapılmalı — güvenlik riski.
> 2. Bölüm 2 (Docker) ve 7 (.gitignore) tek seferde toplu yapılabilir.
> 3. Bölüm 3 (Health) Faz 4 planıyla birleştirilebilir.
> 4. Bölüm 4 (Event payload) encoding pipeline'ı etkileyeceği için dikkatli test gerekir.
> 5. Bölüm 6 (DB naming) yeni migration gerektirir — mevcut datayı bozmamak için `IF EXISTS` kontrolü ile.
