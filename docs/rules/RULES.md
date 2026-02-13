# StreamVault — Backend Kuralları

> **Bu dosya projenin anayasasıdır.**
> Buradaki kurallar tartışmaya açık değildir. Hiçbir gerekçeyle, hiçbir "hızlıca halledelim" bahanesiyle ihlal edilemez.
> CI pipeline'da bu kurallar otomatik kontrol edilir. İhlal eden PR merge edilmez.

---
## 1. Environment & Configuration

### 1.1 Secret'lar ve Environment Değişkenleri

```yaml
# ❌ ASLA YAPMA — docker-compose.yml içine hardcoded değer
services:
  postgres:
    environment:
      POSTGRES_PASSWORD: super-secret-123

# ✅ BÖYLE YAP — tüm secret'lar .env'den gelir
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
- Default değer (`${VAR:-default}`) **sadece** tehlikesiz değişkenler için kullanılır (port, log level). Secret'larda default değer **YASAKTIR**.
- Servis içi config dosyalarında (appsettings.json, config.yaml, config.toml) secret değer bulunmaz. Runtime'da environment variable override ile sağlanır.

```
# .env.example
POSTGRES_USER=streamvault
POSTGRES_PASSWORD=              # ← Boş, kullanıcı dolduracak
POSTGRES_DB=streamvault_users
JWT_SECRET=                     # ← Boş
RABBITMQ_USER=streamvault
RABBITMQ_PASS=                  # ← Boş
MINIO_ACCESS_KEY=
MINIO_SECRET_KEY=
```

### 1.2 Port Numaraları

Port numaraları tek bir yerde tanımlanır (`docker-compose.yml`). Servis kodu içinde port hardcoded yazılmaz, environment variable'dan okunur. Servisler kendi aralarında **servis adı** ile haberleşir.

```yaml
# ❌ YANLIŞ
- CATALOG_URL=http://localhost:5002

# ✅ DOĞRU
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
- Ortak util/helper kodu paylaşılmaz. Her servis kendi util'ini yazar.

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
.env
*.pem
*.key
*.p12
**/bin/
**/obj/
**/target/
**/node_modules/
**/.next/
.idea/
.vscode/
*.swp
*.swo
.DS_Store
data/
*_data/
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

Bu kuralın **HİÇBİR** istisnası yoktur. "Performans için" bile olmaz. Her servisin veritabanı sadece o servisin tekelindedir.

### 3.2 Senkron vs Asenkron Karar Matrisi

```
"Bu çağrının sonucunu HEMEN bilmem gerekiyor mu?"  → Evet: gRPC/REST  → Hayır: RabbitMQ event
"Bu işlem başarısız olursa kullanıcı ne görür?"     → Hata mesajı: Senkron  → Fark etmez: Asenkron
```

**Asenkron olması GEREKEN:** Encoding pipeline, search index güncelleme, bildirim gönderme, recommendation güncelleme, abonelik event'leri.

**Senkron olması GEREKEN:** Erişim kontrolü (tier), JWT doğrulama, içerik bilgisi sorgulama.

### 3.3 Event İsimlendirme

```
Format: {Domain}.{Entity}.{Action} (past tense)

✅ DOĞRU:
  catalog.content.created
  catalog.content.updated
  subscription.subscription.created
  encoding.job.completed
  encoding.job.failed

❌ YANLIŞ:
  create_content          ← İmperative, domain yok
  ContentCreated          ← Namespace yok, routing key olarak kullanılamaz
```

### 3.4 Event Payload Kuralları

```json
// Her event'te ZORUNLU alanlar:
{
  "eventId": "uuid",
  "eventType": "catalog.content.created",
  "timestamp": "2026-02-13T14:30:00Z",
  "source": "catalog-service",
  "correlationId": "trace-id",
  "data": { }
}
```

- Event payload'ında **sadece** gerekli veri bulunur. Tüm entity'yi koyma.
- Consumer ek bilgi gerekiyorsa kaynak servise gRPC ile sorar.
- Event payload'ı değiştirmek breaking change'dir. Yeni alan eklenebilir, mevcut alan silinemez/tipi değiştirilemez.

---
## 4. Veritabanı

### 4.1 Migration Kuralları

- Veritabanı şeması **SADECE** migration dosyaları ile değiştirilir. Elle SQL çalıştırarak şema değiştirmek yasaktır.
- Migration dosyaları sıralı numaralandırılır ve **ASLA** değiştirilmez. Hata varsa yeni migration yazılır.
- Her migration dosyası hem `up` hem `down` (rollback) içerir.

```sql
-- ❌ YANLIŞ: docker exec -it postgres psql -c "ALTER TABLE..."
-- ✅ DOĞRU: migrations/005_add_user_phone.sql
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
idx_profiles_user_id, idx_watchlist_profile_id_content_id
-- Constraint: {tip}_{tablo}_{açıklama}
pk_users, uq_users_email, fk_profiles_user_id, ck_profiles_max_count
```

```javascript
// MongoDB collection: snake_case, çoğul
// movies, series, genres, notifications
// MongoDB field: camelCase
// contentType, releaseYear, averageRating, createdAt
```

### 4.3 Veri Tipi Kuralları

```
ID'ler:       PostgreSQL → UUID (gen_random_uuid())  |  MongoDB → ObjectId veya string UUID
Tarihler:     PostgreSQL → TIMESTAMPTZ (UTC)  |  MongoDB → ISODate (UTC)  |  API → ISO 8601 ("2026-02-13T14:30:00Z")
              Uygulama kodu → UTC. Timezone dönüşümü sadece presentation katmanında.
Para:         PostgreSQL → DECIMAL(10,2). ASLA float/double kullanma. Para birimi her zaman yanında saklanır.
Boolean:      Kolon adı is_ veya has_ ile başlar: is_active, is_kids, has_downloads
```

### 4.4 Soft Delete Yasağı

Bu projede soft delete (is_deleted, deleted_at) **KULLANILMAZ**. Silinen veri silinir (hard delete). Geçmiş tutmak gerekiyorsa ayrı bir audit/history tablosu oluşturulur.

Sebep: Soft delete her sorguda `WHERE deleted_at IS NULL` eklemeyi gerektirir, unutulduğunda silinmiş veri geri gelir, index'leri kirletir.

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

❌ YANLIŞ:
  GET    /api/getMovies                    ← Fiil kullanma
  POST   /api/catalog/movies/create        ← Gereksiz fiil, POST zaten create
  GET    /api/catalog/Movies               ← Büyük harf
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
  200 OK              → GET, PUT, PATCH başarılı
  201 Created         → POST başarılı (yeni kaynak oluştu)
  202 Accepted        → Asenkron işlem kabul edildi (encoding job)
  204 No Content      → DELETE başarılı (body dönmez)
Client hatası:
  400 Bad Request     → Validasyon hatası (body'de detay)
  401 Unauthorized    → Token yok veya geçersiz
  403 Forbidden       → Token geçerli ama yetkisiz
  404 Not Found       → Kaynak bulunamadı
  409 Conflict        → Çakışma (duplicate email, aktif abonelik var)
  422 Unprocessable   → İş kuralı ihlali (max 5 profil)
  429 Too Many Req    → Rate limit aşıldı
Server hatası:
  500 Internal Error  → Beklenmeyen hata (loglanır, detay client'a verilmez)
  502 Bad Gateway     → Downstream servis erişilemez
  503 Unavailable     → Circuit breaker açık
  504 Gateway Timeout → Downstream servis timeout
```

### 5.4 Hata Response Formatı

Tüm servisler aynı hata formatını kullanır. İstisna yok.

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "İstek doğrulama hatası",
    "details": [
      { "field": "email", "message": "Geçerli bir email adresi giriniz" }
    ],
    "traceId": "abc-123-def"
  }
}
```

```
❌ YANLIŞ — tutarsız formatlar: { "msg": "..." }, { "error": "not found" }, veya plain string
```

### 5.5 Sayfalama Formatı

```json
// GET /api/catalog/movies?page=1&pageSize=20
{
  "items": [ ... ],
  "page": 1,
  "pageSize": 20,
  "totalCount": 156,
  "totalPages": 8
}
```

- `pageSize`: 1-100, default 20. `page`: 1'den başlar (0 değil).
- Tüm list endpoint'leri sayfalama kullanır (genre listesi gibi sabit/küçük listeler hariç).

### 5.6 Tarih Formatı

```
API'deki tüm tarihler ISO 8601 formatında ve UTC'dir.
✅ DOĞRU: "2026-02-13T14:30:00Z"
❌ YANLIŞ: "13/02/2026", 1739451000 (unix timestamp), "2026-02-13T14:30:00+03:00" (offset)
```

---
## 6. Kod Kalitesi

### 6.1 Yorum ve Dokümantasyon

```go
// Kod NEDEN yapıldığını açıklar, NE yapıldığını değil. İyi isimlendirme yorum ihtiyacını ortadan kaldırır.

// ❌ YANLIŞ — ne yapıldığını açıklayan gereksiz yorum
// Kullanıcıyı veritabanından getir
user, err := repo.GetUserByID(ctx, userID)

// ✅ DOĞRU — neden yapıldığını açıklayan anlamlı yorum
// Soft-delete kontrolü yapmıyoruz çünkü projede hard delete kullanılıyor.
// Silinen kullanıcıların token'ları JWT expiry ile doğal olarak geçersiz olur.
user, err := repo.GetUserByID(ctx, userID)

// ❌ TODO, FIXME, HACK yorumları commit edilmez. Ya düzelt ya da issue aç.
```

### 6.2 Hata Yönetimi

Hataları **ASLA** yutma. Her hata ya handle edilir ya da yukarıya fırlatılır.

```go
// ❌ YANLIŞ — hatayı yutmak
result, _ := db.Query(ctx, sql)

// ❌ YANLIŞ — sadece loglamak ama akışı devam ettirmek
result, err := db.Query(ctx, sql)
if err != nil {
    log.Error("db error", err)
}

// ✅ DOĞRU
result, err := db.Query(ctx, sql)
if err != nil {
    return fmt.Errorf("failed to query users: %w", err)
}
```

```csharp
// ❌ YANLIŞ — boş catch
try { await _repository.SaveAsync(entity); }
catch (Exception) { }

// ✅ DOĞRU
try { await _repository.SaveAsync(entity); }
catch (DbUpdateException ex) {
    _logger.LogError(ex, "Failed to save entity {EntityId}", entity.Id);
    throw;
}
```

### 6.3 Loglama Kuralları

```
Log seviyesi: ERROR (müdahale gerektiren) → WARN (potansiyel sorun) → INFO (iş olayları) → DEBUG (geliştirme)

Log'larda ASLA: Şifre (hash dahil), tam JWT token, kredi kartı, kişisel veri (email kısaltılabilir: a***@test.com)
Log'larda HER ZAMAN: Trace ID, servis adı, timestamp (ISO 8601 UTC), ilgili entity ID
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
- Bir fonksiyon TEK bir iş yapar. 50 satır iyi, 100+ kesinlikle bölünmeli.
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

- Container içinde **root olarak çalıştırma**. Her zaman `USER nobody` veya özel user.
- `.dockerignore` dosyası zorunlu. Gereksiz dosyalar image'a girmez.
- Her image'a `HEALTHCHECK` tanımla veya docker-compose'da `healthcheck` kullan.
- Image tag'inde `latest` kullanma. Her zaman spesifik versiyon (`postgres:16-alpine`).

### 7.2 Docker Compose Kuralları

```yaml
# ❌ YANLIŞ — depends_on/healthcheck/restart/limits yok
services:
  user-service:
    build: ./services/user-service

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
# ❌ YANLIŞ — data için bind mount
volumes:
  - ./data/postgres:/var/lib/postgresql/data

# ✅ DOĞRU — named volume (data) + bind mount (config, read-only)
volumes:
  postgres_data:
services:
  postgres:
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./infrastructure/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql
```

---
## 8. Güvenlik

### 8.1 JWT Kuralları

```
- Access token: max 60 dakika. Refresh token: 30 gün.
- JWT secret minimum 256-bit (32+ karakter).
- JWT payload'da şifre, kişisel veri bulunmaz.
- Token payload:
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

- Tüm API endpoint'leri (public olanlar hariç) JWT doğrulamasından geçer.
- Gateway JWT doğruladıktan sonra `X-User-Id` ve `X-User-Role` header'larını ekler.
- Downstream servisler bu header'lara güvenir, token'ı tekrar doğrulamaz.
- Downstream servisler Gateway'siz doğrudan erişilemez (network policy).
- Admin endpoint'leri role kontrolü yapar. Rate limiting tüm endpoint'lerde aktif.
- CORS sadece bilinen origin'lere açık.

### 8.3 Password Kuralları

- Şifreler BCrypt ile hash'lenir (cost factor: minimum 10).
- Düz metin şifre **ASLA** loglanmaz, **ASLA** veritabanına yazılmaz.
- Minimum 8 karakter. Büyük harf + küçük harf + rakam zorunlu.

### 8.4 Input Validation

- Tüm input'lar servis girişinde validate edilir. Veritabanına ulaşmadan önce temiz veri garantilenir.
- Validation business logic'ten ayrıdır (.NET: FluentValidation, Go: validator, Rust: validator crate).
- String alanlara max uzunluk, sayısal alanlara min/max aralığı zorunlu. Sınırsız string yasaktır.
- Enum değerleri whitelist ile kontrol edilir.

---
## 9. Test

### 9.1 Test Zorunlulukları

```
Her servis için minimum:
  - Unit test: Handler/Command handler testleri
  - Integration test: API endpoint testleri (in-memory DB ile)
Test olmadan PR merge edilmez.

Test isimlendirme: {MethodName}_{Senaryo}_{BeklenenSonuç}
  RegisterUser_WithValidData_ReturnsCreated
  RegisterUser_WithDuplicateEmail_ReturnsConflict
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
feature/{servis}/{kısa-açıklama}    → feature/user-service/add-profile-endpoint
fix/{servis}/{kısa-açıklama}        → fix/catalog-service/pagination-off-by-one
chore/{kısa-açıklama}               → chore/update-docker-compose
docs/{kısa-açıklama}                → docs/api-documentation
```

### 10.2 Commit Mesajı Formatı

```
{tip}({kapsam}): {açıklama}
Tipler: feat, fix, refactor, test, docs, chore, ci  |  Kapsam: servis adı veya genel alan

✅ DOĞRU:
  feat(user-service): add profile creation endpoint
  fix(gateway): handle timeout on catalog-service calls
  test(subscription): add saga compensation tests

❌ YANLIŞ:
  fixed stuff, update, WIP, asdfasdf
```

### 10.3 PR Kuralları

- PR açıklaması ne yapıldığını ve **NEDEN** yapıldığını anlatır. Her PR tek bir konuya odaklanır.
- Yeni environment variable → `.env.example` güncelle.
- Yeni endpoint → API dokümantasyonu güncelle, hata formatı standart olsun.
- Veritabanı değişikliği → migration dosyası olmalı.
- Yeni bağımlılık → sebebi açıklanmalı.

---
## 11. Dil Spesifik Kurallar

### 11.1 Go

```go
// Error handling: Her error kontrol edilir. err != nil → return veya handle. İstisna yok.

// Context: Her fonksiyon ilk parametre olarak context.Context alır.
func GetUser(ctx context.Context, id string) (*User, error)

// Goroutine: Her goroutine'in yaşam döngüsü kontrol edilir. Fire-and-forget yasaktır.
// ❌ go processItem(item)
// ✅ errgroup kullan:
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return processItem(ctx, item) })
if err := g.Wait(); err != nil { ... }

// Naming: MixedCaps (exported), mixedCaps (unexported). Getter'da Get prefix'i kullanılmaz:
// ❌ user.GetName()  ✅ user.Name()
```

### 11.2 Rust

```rust
// unwrap() ve expect() production kodunda YASAKTIR. Test kodu hariç Result<T,E>/Option<T> kullan.

// ❌ YANLIŞ
let user = db.find_user(id).await.unwrap();
let config = std::fs::read_to_string("config.toml").expect("config yok");

// ✅ DOĞRU
let user = db.find_user(id).await
    .map_err(|e| AppError::Database(e))?;
let config = std::fs::read_to_string("config.toml")
    .map_err(|e| AppError::Config(format!("config okunamadı: {}", e)))?;

// clone(): Gereksiz clone yasaktır. Borrow checker ile çalış. Gerekiyorsa yanına yorum ekle.
// unsafe: Bu projede unsafe blok YASAKTIR. İstisna yok.
```

### 11.3 .NET (C#)

```csharp
// async/await: Async method varsa ASLA .Result veya .Wait() kullanma — deadlock oluşturur.
// ❌ var user = _repository.GetByIdAsync(id).Result;
// ✅ var user = await _repository.GetByIdAsync(id);

// DI: new ile service oluşturma. Her şey DI container'dan gelir (constructor injection).

// Nullable: Nullable reference types etkin. CS8600-8605 uyarıları suppress edilmez, düzeltilir.

// String: interpolation kullan.
// ❌ "User " + user.Id + " not found"
// ✅ $"User {user.Id} not found"
```

---
## 12. Performans Kuralları

### 12.1 N+1 Query Yasağı

```csharp
// Bir liste çekerken her eleman için ayrı sorgu atmak YASAKTIR.

// ❌ YANLIŞ
var users = await _context.Users.ToListAsync();
foreach (var user in users) {
    user.Profiles = await _context.Profiles
        .Where(p => p.UserId == user.Id).ToListAsync();  // N sorgu daha!
}

// ✅ DOĞRU
var users = await _context.Users
    .Include(u => u.Profiles)
    .ToListAsync();  // Tek sorgu
```

### 12.2 Unbounded Query Yasağı

```
Limit olmadan tüm tabloyu çekmek YASAKTIR.

❌ SELECT * FROM movies;              ✅ SELECT * FROM movies LIMIT 100 OFFSET 0;
❌ _context.Movies.ToListAsync();     ✅ _context.Movies.Take(100).Skip(0).ToListAsync();
```

### 12.3 Büyük Dosya İşleme

- Video upload gibi büyük dosya işlemlerinde dosya **ASLA** tamamen belleğe alınmaz (streaming I/O kullan).
- Multipart upload kullanılır. Geçici dosyalar işlem bitince silinir.
- Dosya boyutu limiti config'den gelir, hardcoded değildir.

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

Her senkron çağrı ve asenkron mesajda trace context (trace-id, span-id) propagate edilir. Her log satırında trace-id bulunur. Trace olmadan production'a çıkılmaz.

### 13.3 Metric İsimlendirme

```
Format: {servis}_{ne}_{birim}_{tip}

✅ DOĞRU:
  http_request_duration_seconds (histogram)
  streaming_active_viewers (gauge)

❌ YANLIŞ:
  request_time             ← Birim yok
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
