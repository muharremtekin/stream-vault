# StreamVault — Backend Kuralları

> **Bu dosya projenin anayasasıdır.**
> Buradaki kurallar tartışmaya açık değildir. Hiçbir gerekçeyle, hiçbir "hızlıca halledelim" bahanesiyle ihlal edilemez.
> CI pipeline'da bu kurallar otomatik kontrol edilir. İhlal eden PR merge edilmez.

---

## 1. Environment & Configuration

### 1.1 Secret'lar ve Environment Değişkenleri

```
✅ DOĞRU: Tüm secret'lar ve environment değişkenleri .env dosyasında tanımlanır.
          docker-compose.yml sadece ${VARIABLE_NAME} referansı kullanır.

❌ YANLIŞ: docker-compose.yml içine hardcoded değer yazmak.
```

```yaml
# ❌ ASLA YAPMA
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: super-secret-123
      
  user-service:
    environment:
      - ConnectionStrings__DefaultConnection=Host=postgres;Password=secret

# ✅ BÖYLE YAP
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}

  user-service:
    environment:
      - ConnectionStrings__DefaultConnection=Host=postgres;Database=${POSTGRES_DB};Username=${POSTGRES_USER};Password=${POSTGRES_PASSWORD}
```

**Kurallar:**

- `docker-compose.yml` içinde **TEK BİR** hardcoded secret/password/key olmayacak.
- Tüm secret'lar `.env` dosyasında yaşar. `.env` git'e **ASLA** commit edilmez.
- `.env.example` dosyası her zaman güncel tutulur (gerçek değerler yerine placeholder).
- Default değer kullanılacaksa `${VAR:-default}` syntax'ı **sadece** tehlikesiz değişkenler için kullanılır (port numaraları, log level gibi). Secret'larda default değer **YASAKTIR**.
- Servis içi config dosyalarında (appsettings.json, config.yaml, config.toml) secret değer bulunmaz. Bunlar environment variable override ile runtime'da sağlanır.

```
# .env.example
POSTGRES_USER=streamvault
POSTGRES_PASSWORD=              # ← Boş, kullanıcı dolduracak
POSTGRES_DB=streamvault_users
JWT_SECRET=                     # ← Boş, kullanıcı dolduracak
RABBITMQ_USER=streamvault
RABBITMQ_PASS=                  # ← Boş
MINIO_ACCESS_KEY=
MINIO_SECRET_KEY=
```

### 1.2 Port Numaraları

- Port numaraları tek bir yerde tanımlanır: `docker-compose.yml`.
- Servis kodu içinde port hardcoded yazılmaz, environment variable'dan okunur.
- Servisler kendi aralarında **servis adı** ile haberleşir, localhost veya IP adresi kullanılmaz.

```yaml
# ❌ YANLIŞ
environment:
  - CATALOG_URL=http://192.168.1.5:5002
  - CATALOG_URL=http://localhost:5002

# ✅ DOĞRU
environment:
  - CATALOG_URL=http://catalog-service:5002
```

### 1.3 Config Dosyası Hiyerarşisi

```
Öncelik sırası (yüksekten düşüğe):
  1. Environment variable (runtime'da sağlanan)
  2. .env dosyası (docker-compose tarafından yüklenen)
  3. Servis config dosyası (appsettings.json, config.yaml, config.toml)
  4. Kod içindeki default değer (sadece tehlikesiz ayarlar için)

Hiçbir katmanda secret bulunmaz, secret sadece katman 1 ve 2'de yaşar.
```

---

## 2. Proje Yapısı

### 2.1 Klasör Kuralları

- Her servis kendi klasöründe yaşar. Servisler arası **doğrudan dosya import'u yasaktır.**
- Paylaşılan tek şey `proto/` klasöründeki gRPC tanımlarıdır.
- Ortak util/helper kodu paylaşılmaz. Her servis kendi util'ini yazar. DRY prensibi servis sınırlarında geçerli değildir.

```
# ❌ YANLIŞ — servisler arası import
import { hashPassword } from '../../user-service/src/utils/crypto';

# ✅ DOĞRU — her servisin kendi implementasyonu
import { hashPassword } from '../utils/crypto';
```

### 2.2 Dosya İsimlendirme

```
Go:        snake_case.go          (catalog_consumer.go, minio_client.go)
Rust:      snake_case.rs          (catalog_consumer.rs, minio_client.rs)
.NET:      PascalCase.cs          (CatalogConsumer.cs, MinioClient.cs)
Proto:     snake_case.proto       (streaming_service.proto)
Config:    kebab-case.yml/toml    (docker-compose.yml, loki-config.yml)
Script:    kebab-case.sh          (seed-catalog.sh, test-api.sh)
Migration: NNN_description.sql    (001_create_users.sql)
```

### 2.3 .gitignore — Commit Edilmeyecekler

Aşağıdakiler **ASLA** git'e commit edilmez:

```gitignore
# Secret'lar
.env
*.pem
*.key
*.p12

# Build artifact'ları
**/bin/
**/obj/
**/target/
**/node_modules/
**/.next/

# IDE
.idea/
.vscode/
*.swp
*.swo
.DS_Store

# Data volumes
data/
*_data/

# Geçici dosyalar
tmp/
temp/
*.tmp
*.log
```

---

## 3. Servis İletişimi

### 3.1 Servisler Arası Doğrudan Veritabanı Erişimi Yasaktır

```
✅ DOĞRU: Streaming Service, kullanıcı bilgisi için User Service'e gRPC/HTTP çağrısı yapar.
❌ YANLIŞ: Streaming Service, User Service'in PostgreSQL'ine doğrudan bağlanıp sorgu atar.
```

Bu kuralın **HİÇBİR** istisnası yoktur. "Performans için" bile olmaz. Her servisin kendi veritabanı sadece o servisin tekelindedir. Başka servis o veritabanına bağlanamaz, okuyamaz, yazamaz.

### 3.2 Senkron vs Asenkron Karar Matrisi

```
"Bu çağrının sonucunu HEMEN bilmem gerekiyor mu?"
  → Evet: gRPC veya REST (senkron)
  → Hayır: RabbitMQ event (asenkron)

"Bu işlem başarısız olursa kullanıcı ne görür?"
  → Hata mesajı görmeli: Senkron
  → Fark etmez, arka planda hallolur: Asenkron
```

**Asenkron olması GEREKEN işlemler:**
- Encoding pipeline (upload → transcode → segment)
- Search index güncelleme (catalog → elasticsearch)
- Bildirim gönderme (event → email/push/inapp)
- Recommendation model güncelleme (izleme verisi → model)
- Abonelik event'leri (subscription → user tier güncelleme)

**Senkron olması GEREKEN işlemler:**
- Erişim kontrolü (streaming → subscription tier kontrolü)
- JWT doğrulama (gateway → user service)
- İçerik bilgisi sorgulama (streaming → catalog)

### 3.3 Event İsimlendirme

```
Format: {Domain}.{Entity}.{Action} (past tense)

✅ DOĞRU:
  catalog.content.created
  catalog.content.updated
  catalog.content.deleted
  subscription.subscription.created
  subscription.plan.changed
  encoding.job.completed
  encoding.job.failed
  user.content.rated
  streaming.watch.completed

❌ YANLIŞ:
  create_content          ← İmperative, domain yok
  ContentCreated          ← Namespace yok, routing key olarak kullanılamaz
  catalog.createContent   ← Present tense, imperative
  new-movie-added         ← Tutarsız format
```

### 3.4 Event Payload Kuralları

```json
// Her event'te ZORUNLU alanlar:
{
  "eventId": "uuid",              // İdempotency için unique ID
  "eventType": "catalog.content.created",
  "timestamp": "2026-02-13T14:30:00Z",
  "source": "catalog-service",    // Hangi servis üretti
  "correlationId": "trace-id",    // Distributed tracing için
  "data": {                       // Event'e özel veri
    // ...
  }
}
```

- Event payload'ında **sadece** gerekli veri bulunur. Tüm entity'yi koyma.
- Consumer, event payload'ındaki veriye güvenir. Ek bilgi gerekiyorsa kaynak servise gRPC ile sorar.
- Event payload'ı değiştirmek breaking change'dir. Yeni alan eklenebilir, mevcut alan silinemez veya tipi değiştirilemez.

---

## 4. Veritabanı

### 4.1 Migration Kuralları

- Veritabanı şeması **SADECE** migration dosyaları ile değiştirilir.
- Elle SQL çalıştırarak şema değiştirmek yasaktır.
- Migration dosyaları sıralı numaralandırılır ve **ASLA** değiştirilmez. Hata varsa yeni migration yazılır.
- Her migration dosyası hem `up` hem `down` (rollback) içerir.

```
# ❌ YANLIŞ
docker exec -it postgres psql -c "ALTER TABLE users ADD COLUMN phone VARCHAR(20)"

# ✅ DOĞRU
# migrations/005_add_user_phone.sql içinde:
-- Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
-- Down
ALTER TABLE users DROP COLUMN phone;
```

### 4.2 Naming Convention

```sql
-- Tablo: snake_case, çoğul
users, profiles, watchlist_items, refresh_tokens, saga_states

-- Kolon: snake_case
user_id, created_at, password_hash, is_active

-- Index: idx_{tablo}_{kolonlar}
idx_profiles_user_id
idx_watchlist_profile_id_content_id

-- Constraint: {tip}_{tablo}_{açıklama}
pk_users                        -- Primary key
uq_users_email                  -- Unique
fk_profiles_user_id             -- Foreign key
ck_profiles_max_count           -- Check
```

```javascript
// MongoDB collection: snake_case, çoğul
movies, series, genres, notifications, notification_preferences

// MongoDB field: camelCase
contentType, releaseYear, averageRating, createdAt
```

### 4.3 Veri Tipi Kuralları

```
ID'ler:
  PostgreSQL → UUID (gen_random_uuid())
  MongoDB    → ObjectId (default) veya string UUID

Tarihler:
  PostgreSQL → TIMESTAMPTZ (timezone-aware, her zaman UTC)
  MongoDB    → ISODate (her zaman UTC)
  API response → ISO 8601 string ("2026-02-13T14:30:00Z")
  Uygulama kodu → UTC. Timezone dönüşümü sadece presentation katmanında.

Para:
  PostgreSQL → DECIMAL(10,2) veya DECIMAL(12,4)
  ASLA float/double kullanma.
  Para birimi her zaman yanında saklanır (amount + currency).

Boolean:
  Kolon adı is_ veya has_ ile başlar: is_active, is_kids, has_downloads
```

### 4.4 Soft Delete Yasağı

```
Bu projede soft delete (is_deleted, deleted_at) KULLANILMAZ.

Silinen veri silinir (hard delete).
Geçmiş tutmak gerekiyorsa ayrı bir audit/history tablosu oluşturulur.

Sebep: Soft delete her sorguda WHERE deleted_at IS NULL eklemeyi gerektirir,
       unutulduğunda silinmiş veri geri gelir, index'leri kirletir.
```

---

## 5. API Tasarımı

### 5.1 URL Yapısı

```
Format: /api/{servis-alanı}/{kaynak}

✅ DOĞRU:
  GET    /api/catalog/movies
  GET    /api/catalog/movies/{id}
  POST   /api/catalog/movies
  PUT    /api/catalog/movies/{id}
  DELETE /api/catalog/movies/{id}
  GET    /api/catalog/movies/{id}/streaming-info
  GET    /api/users/me/profiles
  POST   /api/users/me/profiles/{profileId}/watchlist

❌ YANLIŞ:
  GET    /api/getMovies                    ← Fiil kullanma
  GET    /api/catalog/movie                ← Tekil kullanma
  POST   /api/catalog/movies/create        ← Gereksiz fiil, POST zaten create
  GET    /api/catalog/Movies               ← Büyük harf
  GET    /api/v1/catalog/movies            ← Bu projede versioning yok (henüz)
  POST   /api/catalog/deleteMovie/{id}     ← DELETE method kullan
```

### 5.2 HTTP Method Kullanımı

```
GET     → Veri oku (body göndermez, idempotent)
POST    → Yeni kaynak oluştur (201 döner, Location header ile)
PUT     → Kaynağı tamamen güncelle (idempotent)
PATCH   → Kaynağı kısmen güncelle
DELETE  → Kaynağı sil (idempotent, 204 döner)
```

### 5.3 HTTP Status Code Kuralları

```
Başarı:
  200 OK                → GET, PUT, PATCH başarılı
  201 Created           → POST başarılı (yeni kaynak oluştu)
  202 Accepted          → Asenkron işlem kabul edildi (encoding job)
  204 No Content        → DELETE başarılı (body dönmez)

Client hatası:
  400 Bad Request       → Validasyon hatası (body'de detay)
  401 Unauthorized      → Token yok veya geçersiz
  403 Forbidden         → Token geçerli ama yetkisiz (admin değil)
  404 Not Found         → Kaynak bulunamadı
  409 Conflict          → Çakışma (duplicate email, aktif abonelik var)
  422 Unprocessable     → İş kuralı ihlali (max 5 profil)
  429 Too Many Requests → Rate limit aşıldı

Server hatası:
  500 Internal Error    → Beklenmeyen hata (loglanır, detay client'a verilmez)
  502 Bad Gateway       → Downstream servis erişilemez
  503 Service Unavailable → Circuit breaker açık
  504 Gateway Timeout   → Downstream servis timeout
```

### 5.4 Hata Response Formatı

Tüm servisler aynı hata formatını kullanır. İstisna yok.

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "İstek doğrulama hatası",
    "details": [
      {
        "field": "email",
        "message": "Geçerli bir email adresi giriniz"
      },
      {
        "field": "password",
        "message": "Şifre en az 8 karakter olmalıdır"
      }
    ],
    "traceId": "abc-123-def"
  }
}
```

```
❌ YANLIŞ — tutarsız hata formatları:
  { "msg": "hata oluştu" }
  { "error": "not found" }
  { "Error": { "Message": "..." } }
  { "errors": ["hata1", "hata2"] }
  "Bir hata oluştu"                        ← String response
```

### 5.5 Sayfalama Formatı

```json
// Request:
// GET /api/catalog/movies?page=1&pageSize=20

// Response — her zaman aynı wrapper:
{
  "items": [ ... ],
  "page": 1,
  "pageSize": 20,
  "totalCount": 156,
  "totalPages": 8
}
```

- `pageSize` minimum 1, maksimum 100. Default 20.
- `page` 1'den başlar (0 değil).
- Tüm list endpoint'leri sayfalama kullanır, sayfalama olmadan liste dönmek yasaktır (genre listesi gibi sabit/küçük listeler hariç).

### 5.6 Tarih Formatı

```
API'deki tüm tarihler ISO 8601 formatında ve UTC'dir.

✅ DOĞRU: "2026-02-13T14:30:00Z"
❌ YANLIŞ: "13/02/2026"
❌ YANLIŞ: "Feb 13, 2026"
❌ YANLIŞ: 1739451000 (unix timestamp)
❌ YANLIŞ: "2026-02-13T14:30:00+03:00" (timezone offset)
```

---

## 6. Kod Kalitesi

### 6.1 Yorum ve Dokümantasyon

```
Kod NEDEN yapıldığını açıklar, NE yapıldığını değil.
İyi isimlendirme yorum ihtiyacını ortadan kaldırır.
```

```go
// ❌ YANLIŞ — ne yapıldığını açıklayan gereksiz yorum
// Kullanıcıyı veritabanından getir
user, err := repo.GetUserByID(ctx, userID)

// ✅ DOĞRU — neden yapıldığını açıklayan anlamlı yorum
// Soft-delete kontrolü yapmıyoruz çünkü projede hard delete kullanılıyor.
// Silinen kullanıcıların token'ları JWT expiry ile doğal olarak geçersiz olur.
user, err := repo.GetUserByID(ctx, userID)
```

```
// ❌ YANLIŞ — TODO, FIXME, HACK yorumları
// TODO: Bunu sonra düzelt
// HACK: Geçici çözüm
// FIXME: Bu bazen çalışmıyor

// Bunlar commit edilmez. Ya düzelt ya da issue aç.
```

### 6.2 Hata Yönetimi

```
Hataları ASLA yutma. Her hata ya handle edilir ya da yukarıya fırlatılır.
```

```go
// ❌ YANLIŞ — hatayı yutmak
result, _ := db.Query(ctx, sql)

// ❌ YANLIŞ — sadece loglamak ama akışı devam ettirmek
result, err := db.Query(ctx, sql)
if err != nil {
    log.Error("db error", err)
    // sonra result'ı nil olarak kullanmaya devam...
}

// ✅ DOĞRU
result, err := db.Query(ctx, sql)
if err != nil {
    return fmt.Errorf("failed to query users: %w", err)
}
```

```csharp
// ❌ YANLIŞ — boş catch
try {
    await _repository.SaveAsync(entity);
} catch (Exception) { }

// ❌ YANLIŞ — generic catch ile detay kaybı
try {
    await _repository.SaveAsync(entity);
} catch (Exception ex) {
    throw new Exception("Kaydetme hatası");  // Orijinal exception kaybedildi
}

// ✅ DOĞRU
try {
    await _repository.SaveAsync(entity);
} catch (DbUpdateException ex) {
    _logger.LogError(ex, "Failed to save entity {EntityId}", entity.Id);
    throw;  // Veya domain exception'a wrap et
}
```

### 6.3 Loglama Kuralları

```
Log seviyesi seçimi:
  ERROR  → Kullanıcıyı etkileyen, müdahale gerektiren hatalar
  WARN   → Potansiyel sorun ama sistem çalışmaya devam ediyor
  INFO   → Önemli iş olayları (kullanıcı kaydı, ödeme, encoding tamamlanma)
  DEBUG  → Geliştirme amaçlı detay (production'da kapalı)

Log'larda ASLA bulunmayacaklar:
  ❌ Şifre (hash dahil)
  ❌ Tam JWT token
  ❌ Kredi kartı numarası
  ❌ Kişisel veri (email kısaltılabilir: a***@test.com)

Log'larda HER ZAMAN bulunacaklar:
  ✅ Trace ID (correlation)
  ✅ Servis adı
  ✅ Timestamp (ISO 8601, UTC)
  ✅ İlgili entity ID (userId, contentId, jobId)
```

```go
// ❌ YANLIŞ
log.Println("error occurred")
fmt.Printf("user logged in: %s, password: %s\n", email, password)

// ✅ DOĞRU
logger.Info("user logged in",
    "userId", user.ID,
    "email", maskEmail(user.Email),
    "traceId", ctx.Value("traceId"),
)
```

### 6.4 Fonksiyon/Method Kuralları

```
- Bir fonksiyon TEK bir iş yapar.
- Fonksiyon 50 satırı geçmiyorsa iyi, 100 satırı geçiyorsa kesinlikle bölünmeli.
- Parametre sayısı 5'i geçmemelidir. Fazlaysa struct/object kullan.
- Boolean parametre yerine iki ayrı fonksiyon tercih et.
```

```go
// ❌ YANLIŞ
func ProcessUser(id string, sendEmail bool, updateCache bool, isAdmin bool) { ... }

// ✅ DOĞRU
func ProcessUser(id string, opts ProcessOptions) { ... }

type ProcessOptions struct {
    SendEmail   bool
    UpdateCache bool
    IsAdmin     bool
}
```

---

## 7. Docker & Altyapı

### 7.1 Dockerfile Kuralları

```dockerfile
# Her Dockerfile multi-stage build kullanır.
# Final image'da build tool'ları bulunmaz.

# ❌ YANLIŞ — tek stage, dev dependency'ler production'da
FROM golang:1.22
COPY . .
RUN go build -o app .
CMD ["./app"]

# ✅ DOĞRU — multi-stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/app /usr/local/bin/app
USER nobody:nobody
EXPOSE 8080
CMD ["app"]
```

**Ek kurallar:**

- Container içinde **root olarak çalıştırma**. Her zaman `USER nobody` veya özel user.
- `.dockerignore` dosyası zorunlu. `node_modules`, `.git`, `data/` gibi gereksiz dosyalar image'a girmez.
- Her image'a `HEALTHCHECK` tanımla veya docker-compose'da `healthcheck` kullan.
- Image tag'inde `latest` kullanma. Her zaman spesifik versiyon (`postgres:16-alpine`, `redis:7-alpine`).

### 7.2 Docker Compose Kuralları

```yaml
# Her servis:
# 1. depends_on + condition ile bağımlılıklarını tanımlar
# 2. healthcheck tanımlar
# 3. restart policy tanımlar
# 4. Resource limit tanımlar (CPU yoğun servisler için)

# ❌ YANLIŞ
services:
  user-service:
    build: ./services/user-service
    # Bağımlılık yok, healthcheck yok, restart yok

# ✅ DOĞRU
services:
  user-service:
    build: ./services/user-service
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      consul:
        condition: service_started
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5001/health/live"]
      interval: 10s
      timeout: 3s
      retries: 3
      start_period: 15s
```

### 7.3 Volume Kuralları

```yaml
# Named volume kullan, bind mount kullanma (data için).
# Bind mount sadece config dosyaları için.

# ❌ YANLIŞ — data için bind mount
volumes:
  - ./data/postgres:/var/lib/postgresql/data

# ✅ DOĞRU — named volume
volumes:
  postgres_data:

services:
  postgres:
    volumes:
      - postgres_data:/var/lib/postgresql/data                      # Named volume (data)
      - ./infrastructure/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql  # Bind mount (config, read-only)
```

---

## 8. Güvenlik

### 8.1 JWT Kuralları

```
- Access token süresi: 60 dakika (MAX).
- Refresh token süresi: 30 gün.
- JWT secret minimum 256-bit (32+ karakter).
- JWT payload'da şifre, kişisel veri bulunmaz.
- Token payload'ı:
  {
    "sub": "user-uuid",        ← Zorunlu
    "email": "a@b.com",        ← Opsiyonel
    "role": "User",            ← Zorunlu
    "iss": "streamvault",      ← Zorunlu
    "iat": ...,                ← Zorunlu
    "exp": ...                 ← Zorunlu
  }
```

### 8.2 API Güvenlik Kuralları

```
- Tüm API endpoint'leri (public olanlar hariç) JWT doğrulamasından geçer.
- Gateway, JWT'yi doğruladıktan sonra X-User-Id ve X-User-Role header'larını ekler.
- Downstream servisler bu header'lara güvenir, token'ı tekrar doğrulamaz.
- Downstream servisler Gateway'siz doğrudan erişilemez (network policy).
- Admin endpoint'leri role kontrolü yapar (X-User-Role: Admin).
- Rate limiting tüm endpoint'lerde aktiftir.
- CORS sadece bilinen origin'lere açıktır.
```

### 8.3 Password Kuralları

```
- Şifreler BCrypt ile hash'lenir (cost factor: minimum 10).
- Düz metin şifre ASLA loglanmaz, ASLA veritabanına yazılmaz.
- Minimum 8 karakter. Büyük harf + küçük harf + rakam zorunlu.
```

### 8.4 Input Validation

```
- Tüm input'lar servis girişinde validate edilir. Veritabanına
  ulaşmadan önce temiz veri garantilenir.
- Validation business logic'ten ayrıdır
  (.NET: FluentValidation, Go: validator, Rust: validator crate).
- String alanlar max uzunluk sınırına sahiptir. Sınırsız string yasaktır.
- Sayısal alanlar min/max aralığına sahiptir.
- Enum değerleri whitelist ile kontrol edilir.
```

---

## 9. Test

### 9.1 Test Zorunlulukları

```
Her servis için minimum:
  - Unit test: Handler/Command handler testleri
  - Integration test: API endpoint testleri (in-memory DB ile)

Test olmadan PR merge edilmez.

Test isimlendirme:
  {MethodName}_{Senaryo}_{BeklenenSonuç}

  RegisterUser_WithValidData_ReturnsCreated
  RegisterUser_WithDuplicateEmail_ReturnsConflict
  RegisterUser_WithWeakPassword_ReturnsBadRequest
```

### 9.2 Test'te Yapılmayacaklar

```
❌ Test'te gerçek veritabanına bağlanma (in-memory veya testcontainers kullan)
❌ Test'te gerçek RabbitMQ'ya bağlanma (mock kullan)
❌ Test'te sleep/delay ile zamanlama (deterministic olmalı)
❌ Test'ler birbirine bağımlı olmamalı (her test izole çalışmalı)
❌ Test'te hardcoded port kullanma (random port ata)
```

---

## 10. Git & Workflow

### 10.1 Branch İsimlendirme

```
feature/{servis}/{kısa-açıklama}
  feature/user-service/add-profile-endpoint
  feature/gateway/circuit-breaker
  feature/encoding-service/thumbnail-generation

fix/{servis}/{kısa-açıklama}
  fix/catalog-service/pagination-off-by-one
  fix/gateway/cors-header-missing

chore/{kısa-açıklama}
  chore/update-docker-compose
  chore/add-eslint-config

docs/{kısa-açıklama}
  docs/api-documentation
  docs/architecture-update
```

### 10.2 Commit Mesajı Formatı

```
{tip}({kapsam}): {açıklama}

Tipler: feat, fix, refactor, test, docs, chore, ci
Kapsam: servis adı veya genel alan

✅ DOĞRU:
  feat(user-service): add profile creation endpoint
  fix(gateway): handle timeout on catalog-service calls
  test(subscription): add saga compensation tests
  chore(docker): upgrade postgres to 16.2
  docs(readme): add architecture diagram

❌ YANLIŞ:
  fixed stuff
  update
  WIP
  asdfasdf
  .
```

### 10.3 PR Kuralları

```
- PR açıklaması ne yapıldığını ve NEDEN yapıldığını anlatır.
- Her PR tek bir konuya odaklanır. "User service + gateway + catalog fix" tek PR'da olmaz.
- PR'da yeni environment variable eklendiyse .env.example da güncellenmiş olmalı.
- PR'da yeni endpoint eklendiyse API dokümantasyonu güncellenmiş olmalı.
- PR'da veritabanı değişikliği varsa migration dosyası olmalı.
- PR'da yeni bağımlılık eklendiyse sebebi açıklanmalı.
```

---

## 11. Dil Spesifik Kurallar

### 11.1 Go

```go
// Error handling: Her error kontrol edilir.
// err != nil → return veya handle. İstisna yok.

// Context: Her fonksiyon ilk parametre olarak context.Context alır.
func GetUser(ctx context.Context, id string) (*User, error)

// Goroutine: Her goroutine'in yaşam döngüsü kontrol edilir.
// Fire-and-forget goroutine yasaktır. WaitGroup veya errgroup kullan.

// ❌ YANLIŞ
go processItem(item)  // Kim bekleyecek? Hata olursa ne olacak?

// ✅ DOĞRU
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error {
    return processItem(ctx, item)
})
if err := g.Wait(); err != nil { ... }

// Naming: MixedCaps (exported), mixedCaps (unexported)
// Getter'da Get prefix'i kullanılmaz:
//   ❌ user.GetName()
//   ✅ user.Name()
```

### 11.2 Rust

```rust
// unwrap() ve expect() production kodunda YASAKTIR.
// Test kodu hariç her yerde Result<T, E> veya Option<T> pattern matching kullan.

// ❌ YANLIŞ
let user = db.find_user(id).await.unwrap();
let config = std::fs::read_to_string("config.toml").expect("config yok");

// ✅ DOĞRU
let user = db.find_user(id).await
    .map_err(|e| AppError::Database(e))?;

let config = std::fs::read_to_string("config.toml")
    .map_err(|e| AppError::Config(format!("config okunamadı: {}", e)))?;

// clone(): Gereksiz clone yasaktır. Borrow checker ile çalış.
// Eğer clone gerekiyorsa yanına yorum ekle (neden borrow yeterli değil).

// unsafe: Bu projede unsafe blok YASAKTIR. İstisna yok.
```

### 11.3 .NET (C#)

```csharp
// async/await: Async method varsa ASLA .Result veya .Wait() kullanma.
// Deadlock oluşturur.

// ❌ YANLIŞ
var user = _repository.GetByIdAsync(id).Result;
_service.ProcessAsync(data).Wait();

// ✅ DOĞRU
var user = await _repository.GetByIdAsync(id);
await _service.ProcessAsync(data);

// Dependency Injection: new ile service oluşturma. Her şey DI container'dan gelir.
// ❌ var service = new UserService(new UserRepository(new DbContext()));
// ✅ Constructor injection kullan.

// Nullable: Nullable reference types etkin. null dönebilecek her şey ? ile işaretli.
// CS8600-8605 uyarıları suppress edilmez, düzeltilir.

// String concatenation: String.Format veya interpolation kullan.
// ❌ "User " + user.Id + " not found"
// ✅ $"User {user.Id} not found"
```

---

## 12. Performans Kuralları

### 12.1 N+1 Query Yasağı

```
Bir liste çekerken her eleman için ayrı sorgu atmak YASAKTIR.

❌ YANLIŞ:
  var users = await _context.Users.ToListAsync();
  foreach (var user in users) {
      user.Profiles = await _context.Profiles
          .Where(p => p.UserId == user.Id).ToListAsync();  // N sorgu daha!
  }

✅ DOĞRU:
  var users = await _context.Users
      .Include(u => u.Profiles)
      .ToListAsync();  // Tek sorgu
```

### 12.2 Unbounded Query Yasağı

```
Limit olmadan tüm tabloyu çekmek YASAKTIR.

❌ YANLIŞ:
  SELECT * FROM movies;
  db.movies.find({});
  _context.Movies.ToListAsync();

✅ DOĞRU:
  SELECT * FROM movies LIMIT 100 OFFSET 0;
  db.movies.find({}).limit(100).skip(0);
  _context.Movies.Take(100).Skip(0).ToListAsync();
```

### 12.3 Büyük Dosya İşleme

```
Video upload gibi büyük dosya işlemlerinde:
  - Dosya ASLA tamamen belleğe alınmaz (streaming I/O kullan).
  - Multipart upload kullanılır.
  - Geçici dosyalar işlem bitince silinir.
  - Dosya boyutu limiti config'den gelir, hardcoded değildir.
```

---

## 13. Observability Kuralları

### 13.1 Her Servisin Zorunlu Endpoint'leri

```
GET /health/live     → 200 {"status": "up"}        (liveness)
GET /health/ready    → 200 veya 503                 (readiness)
GET /metrics         → Prometheus format             (metrics)

Bu 3 endpoint her serviste ZORUNLUDUR. İstisna yok.
```

### 13.2 Trace Context Propagation

```
Her senkron çağrıda trace context (trace-id, span-id) propagate edilir.
Her asenkron mesajda trace context message header'ına eklenir.
Her log satırında trace-id bulunur.

Trace olmadan production'a çıkılmaz.
```

### 13.3 Metric İsimlendirme

```
Format: {servis}_{ne}_{birim}_{tip}

✅ DOĞRU:
  http_request_duration_seconds (histogram)
  http_requests_total (counter)
  streaming_active_viewers (gauge)
  encoding_job_duration_seconds (histogram)

❌ YANLIŞ:
  request_time             ← Birim yok
  numberOfRequests         ← camelCase, birim yok
  streaming_viewers_count  ← _count suffix'i histogram için reserved
```

---

## 14. Checklist — Her PR Öncesi

Bir PR açmadan önce şu kontrolleri yap:

```
[ ] docker-compose.yml içinde hardcoded secret yok
[ ] .env.example güncellendi (yeni variable eklendiyse)
[ ] Yeni endpoint varsa hata response'u standart formatta
[ ] Yeni endpoint varsa input validation mevcut
[ ] Yeni veritabanı değişikliği varsa migration dosyası var
[ ] Yeni event varsa isimlendirme kuralına uygun
[ ] Test yazıldı ve geçiyor
[ ] Log'larda kişisel veri yok
[ ] Fonksiyonlar 100 satırı geçmiyor
[ ] Error'lar yutulmuyor (handle veya propagate)
[ ] Unbounded query yok (limit/pagination var)
[ ] Commit mesajları formata uygun
[ ] README güncel (yeni bir şey eklendiyse)
```

---

> **Son söz:** Bu kurallar "güzel olurdu" değil "olmazsa olmaz"dır.
> Bir kural anlamsız geliyorsa tartışılabilir ve bu dosya güncellenebilir.
> Ama güncellenmediği sürece herkes bu kurallara uyar — insan da, agent da.
