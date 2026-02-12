# StreamVault — Faz 1: Kalan Görevler

> **Durum:** Proje iskeleti oluşturuldu (134 dosya, 9365 satır). Servislerin çalışır hale getirilmesi gerekiyor.

---

## 1. .NET Solution & Proje Dosyaları

- [x] User Service için `dotnet new sln` ile solution oluştur
- [x] 4 proje ekle: `UserService.Api`, `UserService.Application`, `UserService.Domain`, `UserService.Infrastructure`
- [x] Projeler arası referansları ekle (Api → Application → Domain, Infrastructure → Application)
- [x] NuGet paketlerini ekle: MediatR, FluentValidation, AutoMapper, EF Core (Npgsql), BCrypt.Net-Next, Swashbuckle
- [x] Catalog Service için `dotnet new sln` ile solution oluştur
- [x] 4 proje ekle: `CatalogService.Api`, `CatalogService.Application`, `CatalogService.Domain`, `CatalogService.Infrastructure`
- [x] NuGet paketlerini ekle: MediatR, FluentValidation, AutoMapper, MongoDB.Driver, Swashbuckle
- [x] Test projeleri oluştur: xUnit, Moq, FluentAssertions, Microsoft.AspNetCore.Mvc.Testing

## 2. User Service — Çalışır Hale Getirme

### 2.1 Veritabanı & Migration
- [x] EF Core ile ilk migration oluştur (`dotnet ef migrations add InitialCreate`)
- [x] PostgreSQL bağlantısını doğrula
- [x] Migration'ın `docker compose up` sırasında otomatik uygulanmasını sağla

### 2.2 Auth (Register + Login)
- [x] RegisterUserHandler: email benzersizlik kontrolü, BCrypt ile şifre hashleme, kullanıcı kaydetme
- [x] LoginUserHandler: email/şifre doğrulama, JWT access token + refresh token üretimi
- [x] JwtTokenService: token üretimi ve doğrulama implementasyonu
- [x] BCryptPasswordHasher: Hash ve Verify implementasyonu
- [x] Refresh token endpoint: yeni access token üretimi, eski token'ı revoke etme
- [x] Postman/curl ile uçtan uca test

### 2.3 Profil Yönetimi
- [x] CreateProfileHandler: max 5 profil sınırı validasyonu
- [x] GetUserProfilesHandler: kullanıcının profillerini listeleme
- [x] ProfileController endpoint'lerinin çalışması

### 2.4 Watchlist
- [x] AddToWatchlistHandler: duplicate kontrolü (aynı içerik tekrar eklenemesin)
- [x] GetWatchlistHandler: profil bazlı watchlist listeleme
- [x] Watchlist silme endpoint'i
- [x] WatchlistController endpoint'lerinin çalışması

### 2.5 Consul Entegrasyonu
- [x] Startup'ta Consul'a self-registration implementasyonu
- [x] `/health` endpoint'inin health check olarak Consul'a kaydedilmesi
- [x] Graceful shutdown'da Consul'dan deregistration
- [x] Consul UI'da servisin "healthy" göründüğünü doğrula

## 3. Catalog Service — Çalışır Hale Getirme

### 3.1 MongoDB Bağlantısı
- [x] MongoDB bağlantısını doğrula
- [x] Collection'ların ve index'lerin otomatik oluşturulmasını sağla
- [x] BsonClassMap konfigürasyonlarını ekle

### 3.2 Movie CRUD
- [x] CreateMovieHandler: film ekleme (Admin only)
- [x] GetMoviesHandler: sayfalama, filtreleme (genre, year), sıralama (rating, year, title)
- [x] GetContentByIdHandler: tekil film detayı
- [x] MoviesController endpoint'lerinin çalışması

### 3.3 Series CRUD
- [x] CreateSeriesHandler: dizi ekleme (Admin only)
- [x] AddEpisodeHandler: sezon/bölüm ekleme
- [x] Series listeleme ve detay endpoint'leri
- [x] SeriesController endpoint'lerinin çalışması

### 3.4 Genre & Filtreleme
- [x] GenreRepository: tüm türleri listeleme
- [x] GetByGenreHandler: slug ile tür bazlı içerik listeleme (movie + series karışık)
- [x] Genre başına contentCount hesaplama
- [x] GenresController endpoint'lerinin çalışması

### 3.5 Seed Data
- [x] CatalogSeeder'ı 20-30 gerçek film/dizi ile tamamla
- [x] En az 8-10 farklı genre ekle
- [x] Seed'in uygulama başlangıcında otomatik çalışmasını sağla (boş DB kontrolü ile)

### 3.6 Consul Entegrasyonu
- [x] User Service ile aynı pattern: self-registration, health check, deregistration

## 4. API Gateway — Çalışır Hale Getirme

### 4.1 Build & Dependency
- [x] `go mod tidy` ile dependency'leri çek
- [x] Projenin derlendiğini doğrula (`go build ./...`)
- [x] `config.yaml` dosyası oluştur

### 4.2 Reverse Proxy
- [x] User Service'e yönlendirmenin çalışması (/api/auth/*, /api/users/*)
- [x] Catalog Service'e yönlendirmenin çalışması (/api/catalog/*)
- [x] Request/response header'larının doğru iletilmesi

### 4.3 JWT Auth Middleware
- [x] Public route'ların auth bypass etmesi (/api/auth/register, /api/auth/login, /health)
- [x] Geçerli JWT ile X-User-Id ve X-User-Role header'larının eklenmesi
- [x] Geçersiz/eksik JWT'de 401 dönmesi

### 4.4 Rate Limiting
- [x] Redis bağlantısının çalışması
- [x] Token bucket algoritmasının IP bazlı çalışması
- [x] Limit aşımında 429 + Retry-After header dönmesi

### 4.5 Consul Discovery
- [x] Consul'dan servis adreslerini çözümleme
- [x] Round-robin load balancing
- [x] Health check aggregation endpoint'i (/health)

## 5. Docker & Entegrasyon

- [x] Tüm Dockerfile'ların build edilmesi
- [x] `docker compose up` ile tüm sistemin 60 saniye içinde ayağa kalkması
- [x] Servisler arası ağ iletişiminin çalışması
- [x] Volume mount'ların doğru çalışması
- [x] Environment variable'ların doğru geçmesi

### 5.1 Uçtan Uca Test Akışı
- [x] Register → kullanıcı oluşturuldu (201)
- [x] Login → JWT token alındı (200)
- [x] Refresh → yeni access token alındı (200)
- [x] GET /api/users/me → kullanıcı bilgisi (200)
- [x] POST profile → profil oluşturuldu (201)
- [x] GET /api/catalog/movies → film listesi sayfalı (200)
- [x] GET /api/catalog/genres → türler listelendi (200)
- [x] POST watchlist → watchlist'e eklendi (201)
- [x] DELETE watchlist → watchlist'ten çıkarıldı (204)
- [x] Auth olmadan korumalı endpoint → 401
- [x] Rate limit aşımı → 429

## 6. Cross-Cutting Concerns

### 6.1 Logging
- [ ] User Service: Serilog entegrasyonu, JSON formatında structured logging
- [ ] Catalog Service: Serilog entegrasyonu
- [ ] Gateway: zerolog zaten skeleton'da var, çalıştığını doğrula
- [ ] Correlation ID: her request'e unique ID atanması, servisler arası iletilmesi

### 6.2 Error Handling
- [ ] User Service: ExceptionHandlingMiddleware'in tüm exception türlerini yakalaması
- [ ] Catalog Service: aynı pattern
- [ ] Gateway: timeout ve retry mekanizması
- [ ] Tüm servislerde tutarlı hata response formatı: `{ status, error, message, timestamp }`

## 7. Test & Kalite

### 7.1 Unit Tests
- [ ] RegisterUserHandlerTests: başarılı kayıt, duplicate email, geçersiz input
- [ ] LoginUserHandlerTests: başarılı giriş, yanlış şifre, olmayan kullanıcı
- [ ] JwtTokenServiceTests: token üretimi, doğrulama, süresi dolmuş token
- [ ] CreateMovieHandlerTests: başarılı oluşturma, geçersiz input
- [ ] Gateway middleware testleri: auth, rate limit, cors
- [ ] Hedef: minimum %60 code coverage

### 7.2 Integration Tests
- [ ] AuthControllerTests: register + login akışı, WebApplicationFactory ile
- [ ] CustomWebApplicationFactory: test DB konfigürasyonu (in-memory veya test container)

## 8. Dokümantasyon

- [ ] README.md: proje açıklaması, mimari diyagram, kurulum adımları, API kullanım örnekleri
- [ ] Swagger/OpenAPI: User Service ve Catalog Service için aktif
- [ ] API örnekleri: curl komutları ile temel akışlar

---

## Önerilen Sıralama

1. **.NET Solution dosyaları** → build alınabilir hale getir
2. **User Service** → auth akışı çalışsın
3. **Docker Compose** → altyapı servisleri + User Service ayağa kalksın
4. **API Gateway** → derleme + User Service'e proxy
5. **Catalog Service** → MongoDB + CRUD + seed data
6. **Consul entegrasyonu** → tüm servisler registered
7. **Cross-cutting** → logging, error handling, correlation ID
8. **Testler** → unit + integration, %60 coverage
9. **Dokümantasyon** → README, Swagger
