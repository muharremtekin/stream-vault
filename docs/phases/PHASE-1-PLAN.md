# StreamVault — Faz 1: Foundation

> **Süre:** ~2-3 Hafta
> **Hedef:** Docker Compose üzerinde çalışan 3 servis (API Gateway, User Service, Catalog Service), JWT auth, servisler arası iletişim ve service discovery.
> **Sonuç:** Bir kullanıcı kayıt olabilir, giriş yapabilir, film/dizi kataloğunu görebilir ve watchlist'ine ekleyebilir.

---

## 1. Proje Yapısı (Monorepo)

```
streamvault/
├── docker-compose.yml
├── docker-compose.override.yml        # Local geliştirme için port/volume override
├── .env.example
├── README.md
│
├── proto/                              # Paylaşımlı gRPC proto dosyaları
│   ├── user/v1/user.proto
│   └── catalog/v1/catalog.proto
│
├── gateway/                            # Go — API Gateway
│   ├── cmd/
│   │   └── gateway/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go              # Viper ile config yönetimi
│   │   ├── middleware/
│   │   │   ├── auth.go                # JWT doğrulama middleware
│   │   │   ├── ratelimit.go           # Token bucket rate limiter
│   │   │   ├── cors.go
│   │   │   ├── logging.go            # Structured logging (request/response)
│   │   │   └── recovery.go           # Panic recovery
│   │   ├── proxy/
│   │   │   ├── router.go             # Route tanımları ve yönlendirme
│   │   │   └── upstream.go           # Upstream servis bağlantı yönetimi
│   │   ├── discovery/
│   │   │   ├── consul.go             # Consul service discovery client
│   │   │   └── resolver.go           # Servis adı → adres çözümleme
│   │   └── health/
│   │       └── handler.go            # Health check endpoint
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── Makefile
│
├── services/
│   ├── user-service/                   # .NET — User Service
│   │   ├── src/
│   │   │   ├── UserService.Api/       # ASP.NET Core Web API (Presentation)
│   │   │   │   ├── Controllers/
│   │   │   │   │   ├── AuthController.cs
│   │   │   │   │   ├── ProfileController.cs
│   │   │   │   │   └── WatchlistController.cs
│   │   │   │   ├── Middleware/
│   │   │   │   │   └── ExceptionHandlingMiddleware.cs
│   │   │   │   ├── Filters/
│   │   │   │   │   └── ValidationFilter.cs
│   │   │   │   ├── Program.cs
│   │   │   │   └── appsettings.json
│   │   │   │
│   │   │   ├── UserService.Application/ # İş mantığı katmanı (Use Cases)
│   │   │   │   ├── Commands/
│   │   │   │   │   ├── RegisterUser/
│   │   │   │   │   │   ├── RegisterUserCommand.cs
│   │   │   │   │   │   ├── RegisterUserHandler.cs
│   │   │   │   │   │   └── RegisterUserValidator.cs
│   │   │   │   │   ├── LoginUser/
│   │   │   │   │   │   ├── LoginUserCommand.cs
│   │   │   │   │   │   └── LoginUserHandler.cs
│   │   │   │   │   ├── CreateProfile/
│   │   │   │   │   │   ├── CreateProfileCommand.cs
│   │   │   │   │   │   └── CreateProfileHandler.cs
│   │   │   │   │   └── AddToWatchlist/
│   │   │   │   │       ├── AddToWatchlistCommand.cs
│   │   │   │   │       └── AddToWatchlistHandler.cs
│   │   │   │   ├── Queries/
│   │   │   │   │   ├── GetUserProfiles/
│   │   │   │   │   │   ├── GetUserProfilesQuery.cs
│   │   │   │   │   │   └── GetUserProfilesHandler.cs
│   │   │   │   │   └── GetWatchlist/
│   │   │   │   │       ├── GetWatchlistQuery.cs
│   │   │   │   │       └── GetWatchlistHandler.cs
│   │   │   │   ├── DTOs/
│   │   │   │   │   ├── UserDto.cs
│   │   │   │   │   ├── ProfileDto.cs
│   │   │   │   │   ├── AuthResponseDto.cs
│   │   │   │   │   └── WatchlistItemDto.cs
│   │   │   │   ├── Interfaces/
│   │   │   │   │   ├── IUserRepository.cs
│   │   │   │   │   ├── IProfileRepository.cs
│   │   │   │   │   ├── IWatchlistRepository.cs
│   │   │   │   │   ├── ITokenService.cs
│   │   │   │   │   └── IPasswordHasher.cs
│   │   │   │   └── Mappings/
│   │   │   │       └── MappingProfile.cs
│   │   │   │
│   │   │   ├── UserService.Domain/     # Domain modelleri (Entity'ler)
│   │   │   │   ├── Entities/
│   │   │   │   │   ├── User.cs
│   │   │   │   │   ├── Profile.cs
│   │   │   │   │   └── WatchlistItem.cs
│   │   │   │   ├── Enums/
│   │   │   │   │   ├── SubscriptionTier.cs
│   │   │   │   │   └── ProfileIcon.cs
│   │   │   │   └── Exceptions/
│   │   │   │       ├── UserNotFoundException.cs
│   │   │   │       └── DuplicateEmailException.cs
│   │   │   │
│   │   │   └── UserService.Infrastructure/ # Dış bağımlılıklar (DB, Token, Hash)
│   │   │       ├── Persistence/
│   │   │       │   ├── UserDbContext.cs
│   │   │       │   ├── Configurations/
│   │   │       │   │   ├── UserConfiguration.cs
│   │   │       │   │   ├── ProfileConfiguration.cs
│   │   │       │   │   └── WatchlistItemConfiguration.cs
│   │   │       │   └── Repositories/
│   │   │       │       ├── UserRepository.cs
│   │   │       │       ├── ProfileRepository.cs
│   │   │       │       └── WatchlistRepository.cs
│   │   │       ├── Services/
│   │   │       │   ├── JwtTokenService.cs
│   │   │       │   └── BCryptPasswordHasher.cs
│   │   │       ├── Migrations/
│   │   │       └── DependencyInjection.cs
│   │   │
│   │   ├── tests/
│   │   │   ├── UserService.UnitTests/
│   │   │   │   ├── Commands/
│   │   │   │   │   ├── RegisterUserHandlerTests.cs
│   │   │   │   │   └── LoginUserHandlerTests.cs
│   │   │   │   └── Services/
│   │   │   │       └── JwtTokenServiceTests.cs
│   │   │   └── UserService.IntegrationTests/
│   │   │       ├── AuthControllerTests.cs
│   │   │       └── CustomWebApplicationFactory.cs
│   │   │
│   │   ├── Dockerfile
│   │   └── Makefile
│   │
│   └── catalog-service/                # .NET — Catalog Service
│       ├── src/
│       │   ├── CatalogService.Api/
│       │   │   ├── Controllers/
│       │   │   │   ├── MoviesController.cs
│       │   │   │   ├── SeriesController.cs
│       │   │   │   └── GenresController.cs
│       │   │   ├── Middleware/
│       │   │   │   └── ExceptionHandlingMiddleware.cs
│       │   │   ├── Program.cs
│       │   │   └── appsettings.json
│       │   │
│       │   ├── CatalogService.Application/
│       │   │   ├── Commands/
│       │   │   │   ├── CreateMovie/
│       │   │   │   │   ├── CreateMovieCommand.cs
│       │   │   │   │   ├── CreateMovieHandler.cs
│       │   │   │   │   └── CreateMovieValidator.cs
│       │   │   │   ├── CreateSeries/
│       │   │   │   │   ├── CreateSeriesCommand.cs
│       │   │   │   │   └── CreateSeriesHandler.cs
│       │   │   │   └── AddEpisode/
│       │   │   │       ├── AddEpisodeCommand.cs
│       │   │   │       └── AddEpisodeHandler.cs
│       │   │   ├── Queries/
│       │   │   │   ├── GetMovies/
│       │   │   │   │   ├── GetMoviesQuery.cs
│       │   │   │   │   └── GetMoviesHandler.cs
│       │   │   │   ├── GetContentById/
│       │   │   │   │   ├── GetContentByIdQuery.cs
│       │   │   │   │   └── GetContentByIdHandler.cs
│       │   │   │   └── GetByGenre/
│       │   │   │       ├── GetByGenreQuery.cs
│       │   │   │       └── GetByGenreHandler.cs
│       │   │   ├── DTOs/
│       │   │   │   ├── MovieDto.cs
│       │   │   │   ├── SeriesDto.cs
│       │   │   │   ├── EpisodeDto.cs
│       │   │   │   ├── GenreDto.cs
│       │   │   │   └── ContentSummaryDto.cs
│       │   │   ├── Interfaces/
│       │   │   │   ├── IMovieRepository.cs
│       │   │   │   ├── ISeriesRepository.cs
│       │   │   │   └── IGenreRepository.cs
│       │   │   └── Mappings/
│       │   │       └── MappingProfile.cs
│       │   │
│       │   ├── CatalogService.Domain/
│       │   │   ├── Entities/
│       │   │   │   ├── Movie.cs
│       │   │   │   ├── Series.cs
│       │   │   │   ├── Season.cs
│       │   │   │   ├── Episode.cs
│       │   │   │   ├── Genre.cs
│       │   │   │   └── CastMember.cs
│       │   │   ├── Enums/
│       │   │   │   ├── ContentType.cs
│       │   │   │   ├── MaturityRating.cs
│       │   │   │   └── ContentStatus.cs
│       │   │   └── ValueObjects/
│       │   │       ├── Duration.cs
│       │   │       └── ThumbnailUrl.cs
│       │   │
│       │   └── CatalogService.Infrastructure/
│       │       ├── Persistence/
│       │       │   ├── CatalogDbContext.cs     # MongoDB context
│       │       │   ├── MongoCollectionSettings.cs
│       │       │   └── Repositories/
│       │       │       ├── MovieRepository.cs
│       │       │       ├── SeriesRepository.cs
│       │       │       └── GenreRepository.cs
│       │       ├── Seed/
│       │       │   └── CatalogSeeder.cs       # Başlangıç verileri
│       │       └── DependencyInjection.cs
│       │
│       ├── tests/
│       │   └── CatalogService.UnitTests/
│       │       └── Commands/
│       │           └── CreateMovieHandlerTests.cs
│       │
│       ├── Dockerfile
│       └── Makefile
│
├── infrastructure/
│   ├── consul/
│   │   └── config.json                # Consul agent konfigürasyonu
│   ├── postgres/
│   │   └── init.sql                   # User DB oluşturma
│   ├── mongo/
│   │   └── init-mongo.js             # Catalog DB ve collection oluşturma
│   └── redis/
│       └── redis.conf
│
└── scripts/
    ├── seed-catalog.sh                # Örnek film/dizi verisi yükleme
    ├── test-api.sh                    # Tüm endpoint'leri test eden script
    └── generate-proto.sh             # Proto dosyalarından kod üretme
```

---

## 2. Servis Detayları

### 2.1 API Gateway (Go)

**Sorumluluklar:**

- Gelen tüm HTTP isteklerini alır, doğrular ve ilgili downstream servise yönlendirir
- JWT token doğrulama (token'ı parse eder, `userId` ve `role` bilgisini header'a ekleyerek downstream'e iletir)
- Rate limiting (IP bazlı, token bucket algoritması, Redis üzerinde)
- Request/response logging (structured JSON logs)
- CORS yönetimi
- Health check aggregation (tüm servislerin sağlık durumunu toplar)
- Consul üzerinden service discovery (servis adı → IP:port çözümleme)

**Route Tablosu:**

| Method | Path | Downstream | Auth |
|--------|------|------------|------|
| POST | `/api/auth/register` | user-service | ✗ |
| POST | `/api/auth/login` | user-service | ✗ |
| POST | `/api/auth/refresh` | user-service | ✗ |
| GET | `/api/users/me` | user-service | ✓ |
| GET | `/api/users/me/profiles` | user-service | ✓ |
| POST | `/api/users/me/profiles` | user-service | ✓ |
| GET | `/api/users/me/watchlist` | user-service | ✓ |
| POST | `/api/users/me/watchlist` | user-service | ✓ |
| DELETE | `/api/users/me/watchlist/{id}` | user-service | ✓ |
| GET | `/api/catalog/movies` | catalog-service | ✓ |
| GET | `/api/catalog/movies/{id}` | catalog-service | ✓ |
| GET | `/api/catalog/series` | catalog-service | ✓ |
| GET | `/api/catalog/series/{id}` | catalog-service | ✓ |
| GET | `/api/catalog/genres` | catalog-service | ✓ |
| GET | `/api/catalog/genres/{slug}/content` | catalog-service | ✓ |
| POST | `/api/catalog/movies` | catalog-service | ✓ (Admin) |
| POST | `/api/catalog/series` | catalog-service | ✓ (Admin) |
| GET | `/health` | gateway (local) | ✗ |

**Konfigürasyon (config.yaml):**

```yaml
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 15s

jwt:
  secret: "${JWT_SECRET}"
  issuer: "streamvault"

rate_limit:
  requests_per_second: 100
  burst: 200
  redis_url: "redis://redis:6379/0"

consul:
  address: "consul:8500"
  health_check_interval: 10s

services:
  user:
    name: "user-service"
    prefix: "/api/auth,/api/users"
    timeout: 5s
    retry: 2
  catalog:
    name: "catalog-service"
    prefix: "/api/catalog"
    timeout: 5s
    retry: 2

logging:
  level: "info"
  format: "json"
```

**Öğrenme Noktaları:**

- Go'da HTTP reverse proxy nasıl yazılır (`httputil.ReverseProxy`)
- Middleware chain pattern (auth → ratelimit → log → proxy)
- Consul client ile service discovery
- Redis ile distributed rate limiting
- Graceful shutdown

---

### 2.2 User Service (.NET 8)

**Sorumluluklar:**

- Kullanıcı kaydı (email + şifre, BCrypt hash)
- JWT access token + refresh token üretimi
- Profil yönetimi (bir hesapta max 5 profil — Netflix modeli)
- Watchlist yönetimi (profil bazlı)
- Consul'a kendini kaydetme (self-registration)

**Domain Modelleri:**

```
User
├── Id: Guid
├── Email: string (unique)
├── PasswordHash: string
├── Role: enum (User, Admin)
├── CreatedAt: DateTime
├── UpdatedAt: DateTime
├── Profiles: List<Profile>
│   └── Profile
│       ├── Id: Guid
│       ├── Name: string (max 20 char)
│       ├── Icon: enum (Avatar1..Avatar10)
│       ├── IsKids: bool
│       ├── CreatedAt: DateTime
│       └── Watchlist: List<WatchlistItem>
│           └── WatchlistItem
│               ├── Id: Guid
│               ├── ContentId: string       # Catalog'daki film/dizi ID
│               ├── ContentType: enum (Movie, Series)
│               ├── AddedAt: DateTime
│               └── Note: string?
└── RefreshTokens: List<RefreshToken>
    └── RefreshToken
        ├── Token: string
        ├── ExpiresAt: DateTime
        ├── CreatedAt: DateTime
        └── RevokedAt: DateTime?
```

**API Kontratı:**

```
POST /api/auth/register
  Request:  { "email": "ali@test.com", "password": "Pass123!" }
  Response: 201 { "userId": "guid", "message": "Kayıt başarılı" }

POST /api/auth/login
  Request:  { "email": "ali@test.com", "password": "Pass123!" }
  Response: 200 {
    "accessToken": "eyJ...",
    "refreshToken": "abc...",
    "expiresIn": 3600,
    "user": { "id": "guid", "email": "ali@test.com", "role": "User" }
  }

POST /api/auth/refresh
  Request:  { "refreshToken": "abc..." }
  Response: 200 { "accessToken": "eyJ...", "refreshToken": "def...", "expiresIn": 3600 }

GET /api/users/me
  Headers:  X-User-Id (gateway tarafından eklenir)
  Response: 200 { "id": "guid", "email": "ali@test.com", "role": "User", "profileCount": 2 }

GET /api/users/me/profiles
  Response: 200 [
    { "id": "guid", "name": "Ali", "icon": "Avatar3", "isKids": false },
    { "id": "guid", "name": "Çocuklar", "icon": "Avatar7", "isKids": true }
  ]

POST /api/users/me/profiles
  Request:  { "name": "Ali", "icon": "Avatar3", "isKids": false }
  Response: 201 { "id": "guid", "name": "Ali", ... }
  Error:    400 { "error": "Maksimum 5 profil oluşturabilirsiniz" }

GET /api/users/me/profiles/{profileId}/watchlist
  Response: 200 [
    { "id": "guid", "contentId": "abc", "contentType": "Movie", "addedAt": "2026-02-13T..." }
  ]

POST /api/users/me/profiles/{profileId}/watchlist
  Request:  { "contentId": "abc", "contentType": "Movie" }
  Response: 201 { "id": "guid", ... }

DELETE /api/users/me/profiles/{profileId}/watchlist/{itemId}
  Response: 204
```

**Veritabanı:** PostgreSQL

```sql
-- User tablosu
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'User',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Profil tablosu
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(20) NOT NULL,
    icon VARCHAR(20) NOT NULL DEFAULT 'Avatar1',
    is_kids BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_max_profiles CHECK (/* uygulama katmanında kontrol */)
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);

-- Watchlist tablosu
CREATE TABLE watchlist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    content_id VARCHAR(50) NOT NULL,
    content_type VARCHAR(10) NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note TEXT,
    CONSTRAINT uq_watchlist_unique UNIQUE (profile_id, content_id)
);

CREATE INDEX idx_watchlist_profile_id ON watchlist_items(profile_id);

-- Refresh token tablosu
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
```

**Öğrenme Noktaları:**

- Clean Architecture (.NET'te katmanlı yapı)
- CQRS with MediatR (Command/Query ayrımı)
- FluentValidation ile input doğrulama
- EF Core migrations ve configuration
- JWT token üretimi ve refresh token akışı
- Repository pattern
- Global exception handling middleware

---

### 2.3 Catalog Service (.NET 8)

**Sorumluluklar:**

- Film ve dizi metadata CRUD operasyonları
- Tür (genre) yönetimi
- Sayfalama, filtreleme ve sıralama
- Başlangıç verisi (seed data) ile gerçekçi içerik
- Admin-only içerik ekleme

**Domain Modelleri:**

```
Movie
├── Id: string (MongoDB ObjectId)
├── Title: string
├── OriginalTitle: string?
├── Description: string
├── ReleaseYear: int
├── Duration: Duration (hours, minutes)
├── MaturityRating: enum (G, PG, PG13, R, NC17)
├── Genres: List<string>
├── Cast: List<CastMember>
│   └── CastMember
│       ├── Name: string
│       ├── Role: string
│       └── PhotoUrl: string?
├── Director: string
├── ThumbnailUrl: string
├── BannerUrl: string
├── TrailerUrl: string?
├── AverageRating: double
├── RatingCount: int
├── Status: enum (Draft, Published, Archived)
├── Tags: List<string>
├── CreatedAt: DateTime
└── UpdatedAt: DateTime

Series
├── Id: string
├── Title: string
├── Description: string
├── ReleaseYear: int
├── MaturityRating: enum
├── Genres: List<string>
├── Cast: List<CastMember>
├── Creator: string
├── ThumbnailUrl: string
├── BannerUrl: string
├── Status: enum
├── Seasons: List<Season>
│   └── Season
│       ├── SeasonNumber: int
│       ├── Title: string?
│       ├── ReleaseYear: int
│       └── Episodes: List<Episode>
│           └── Episode
│               ├── EpisodeNumber: int
│               ├── Title: string
│               ├── Description: string
│               ├── Duration: Duration
│               └── ThumbnailUrl: string
├── CreatedAt: DateTime
└── UpdatedAt: DateTime

Genre
├── Id: string
├── Name: string
├── Slug: string (url-friendly: "bilim-kurgu")
├── Description: string
└── IconUrl: string?
```

**API Kontratı:**

```
GET /api/catalog/movies?page=1&pageSize=20&genre=aksiyon&sort=rating_desc&year=2025
  Response: 200 {
    "items": [ { "id": "abc", "title": "Film Adı", "releaseYear": 2025, ... } ],
    "page": 1,
    "pageSize": 20,
    "totalCount": 156,
    "totalPages": 8
  }

GET /api/catalog/movies/{id}
  Response: 200 { "id": "abc", "title": "...", "cast": [...], "seasons": null, ... }
  Error:    404 { "error": "İçerik bulunamadı" }

GET /api/catalog/series?page=1&pageSize=20
  Response: 200 { "items": [...], "page": 1, ... }

GET /api/catalog/series/{id}
  Response: 200 { "id": "def", "title": "...", "seasons": [ { "episodes": [...] } ] }

GET /api/catalog/genres
  Response: 200 [
    { "id": "1", "name": "Aksiyon", "slug": "aksiyon", "contentCount": 45 },
    { "id": "2", "name": "Bilim Kurgu", "slug": "bilim-kurgu", "contentCount": 32 }
  ]

GET /api/catalog/genres/{slug}/content?page=1&pageSize=20
  Response: 200 { "items": [ /* movies + series mixed */ ], ... }

POST /api/catalog/movies (Admin only)
  Request: {
    "title": "Yeni Film",
    "description": "...",
    "releaseYear": 2026,
    "durationMinutes": 142,
    "maturityRating": "PG13",
    "genres": ["aksiyon", "bilim-kurgu"],
    "director": "Yönetmen Adı",
    "cast": [{ "name": "Oyuncu", "role": "Karakter" }],
    "thumbnailUrl": "https://...",
    "bannerUrl": "https://..."
  }
  Response: 201 { "id": "new-id", ... }

POST /api/catalog/series (Admin only)
  Request: { ... benzer yapı ... }
  Response: 201 { "id": "new-id", ... }

POST /api/catalog/series/{id}/seasons/{seasonNum}/episodes (Admin only)
  Request: { "episodeNumber": 1, "title": "Pilot", "description": "...", "durationMinutes": 55 }
  Response: 201
```

**Veritabanı:** MongoDB

```javascript
// movies collection
{
  _id: ObjectId("..."),
  title: "Interstellar",
  originalTitle: "Interstellar",
  description: "Dünya'nın geleceği tehlikede...",
  releaseYear: 2014,
  duration: { hours: 2, minutes: 49 },
  maturityRating: "PG13",
  genres: ["bilim-kurgu", "dram", "macera"],
  cast: [
    { name: "Matthew McConaughey", role: "Cooper", photoUrl: null }
  ],
  director: "Christopher Nolan",
  thumbnailUrl: "https://placeholder.co/300x450",
  bannerUrl: "https://placeholder.co/1920x600",
  averageRating: 8.7,
  ratingCount: 1240,
  status: "Published",
  tags: ["uzay", "zaman-yolculugu", "aile"],
  createdAt: ISODate("2026-01-01"),
  updatedAt: ISODate("2026-01-01")
}

// series collection
{
  _id: ObjectId("..."),
  title: "Breaking Bad",
  description: "Lise kimya öğretmeni...",
  releaseYear: 2008,
  maturityRating: "R",
  genres: ["dram", "gerilim", "suç"],
  creator: "Vince Gilligan",
  thumbnailUrl: "...",
  status: "Published",
  seasons: [
    {
      seasonNumber: 1,
      title: "Sezon 1",
      releaseYear: 2008,
      episodes: [
        { episodeNumber: 1, title: "Pilot", description: "...",
          duration: { hours: 0, minutes: 58 }, thumbnailUrl: "..." }
      ]
    }
  ],
  createdAt: ISODate("..."),
  updatedAt: ISODate("...")
}

// genres collection
{
  _id: ObjectId("..."),
  name: "Bilim Kurgu",
  slug: "bilim-kurgu",
  description: "Gelecek teknolojiler ve uzay keşifleri",
  iconUrl: null
}

// MongoDB Indexes
db.movies.createIndex({ genres: 1 })
db.movies.createIndex({ releaseYear: -1 })
db.movies.createIndex({ status: 1, averageRating: -1 })
db.movies.createIndex({ title: "text", description: "text" })
db.series.createIndex({ genres: 1 })
db.series.createIndex({ status: 1 })
db.genres.createIndex({ slug: 1 }, { unique: true })
```

**Seed Data:** 20-30 gerçek film/dizi ile başlangıç verisi (başlık ve metadata gerçek, URL'ler placeholder).

**Öğrenme Noktaları:**

- MongoDB ile .NET (MongoDB.Driver)
- Document modelleme vs relational (nested vs reference)
- Sayfalama pattern'i (cursor-based vs offset)
- CQRS read/write model ayrımı
- Slug üretimi ve URL-friendly routing

---

## 3. Altyapı (Docker Compose)

### 3.1 Servisler ve Portlar

| Servis | İç Port | Dış Port | Açıklama |
|--------|---------|----------|----------|
| API Gateway | 8080 | 8080 | Ana giriş noktası |
| User Service | 5001 | — | Gateway üzerinden erişilir |
| Catalog Service | 5002 | — | Gateway üzerinden erişilir |
| PostgreSQL | 5432 | 5432 | User Service DB |
| MongoDB | 27017 | 27017 | Catalog Service DB |
| Redis | 6379 | 6379 | Rate limit + cache |
| Consul | 8500 | 8500 | Service discovery UI |

### 3.2 Docker Compose Yapısı

```yaml
version: '3.8'

services:
  # --- Infrastructure ---
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: streamvault_users
      POSTGRES_USER: ${POSTGRES_USER:-streamvault}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-secret}
    ports: ["5432:5432"]
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./infrastructure/postgres/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U streamvault"]
      interval: 5s
      timeout: 3s
      retries: 5

  mongo:
    image: mongo:7
    environment:
      MONGO_INITDB_DATABASE: streamvault_catalog
    ports: ["27017:27017"]
    volumes:
      - mongo_data:/data/db
      - ./infrastructure/mongo/init-mongo.js:/docker-entrypoint-initdb.d/init-mongo.js
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    volumes:
      - ./infrastructure/redis/redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  consul:
    image: hashicorp/consul:1.17
    ports:
      - "8500:8500"   # UI
      - "8600:8600/udp" # DNS
    command: agent -server -bootstrap-expect=1 -ui -client=0.0.0.0
    volumes:
      - consul_data:/consul/data

  # --- Application Services ---
  gateway:
    build: ./gateway
    ports: ["8080:8080"]
    environment:
      - JWT_SECRET=${JWT_SECRET:-super-secret-key-change-in-production}
      - CONSUL_ADDRESS=consul:8500
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      consul:
        condition: service_started
      redis:
        condition: service_healthy

  user-service:
    build: ./services/user-service
    environment:
      - ASPNETCORE_ENVIRONMENT=Development
      - ConnectionStrings__DefaultConnection=Host=postgres;Database=streamvault_users;Username=${POSTGRES_USER:-streamvault};Password=${POSTGRES_PASSWORD:-secret}
      - Jwt__Secret=${JWT_SECRET:-super-secret-key-change-in-production}
      - Jwt__Issuer=streamvault
      - Jwt__AccessTokenExpirationMinutes=60
      - Jwt__RefreshTokenExpirationDays=30
      - Consul__Address=http://consul:8500
      - ServiceRegistration__Name=user-service
      - ServiceRegistration__Port=5001
    depends_on:
      postgres:
        condition: service_healthy
      consul:
        condition: service_started

  catalog-service:
    build: ./services/catalog-service
    environment:
      - ASPNETCORE_ENVIRONMENT=Development
      - MongoDB__ConnectionString=mongodb://mongo:27017
      - MongoDB__DatabaseName=streamvault_catalog
      - Consul__Address=http://consul:8500
      - ServiceRegistration__Name=catalog-service
      - ServiceRegistration__Port=5002
    depends_on:
      mongo:
        condition: service_healthy
      consul:
        condition: service_started

volumes:
  postgres_data:
  mongo_data:
  consul_data:
```

### 3.3 Ağ Akışı

```
Client (Browser/Postman)
    │
    ▼ HTTP :8080
┌──────────────┐
│  API Gateway │──── Redis (rate limit check)
│     (Go)     │──── Consul (service discovery)
└──────┬───────┘
       │
       ├── /api/auth/*, /api/users/* ──► User Service (.NET) ──► PostgreSQL
       │
       └── /api/catalog/*             ──► Catalog Service (.NET) ──► MongoDB
```

---

## 4. Consul Service Discovery Akışı

### 4.1 Servis Kaydı (Her .NET servis başlarken)

```
1. User Service ayağa kalkar
2. Startup'ta Consul'a HTTP PUT ile kendini kaydeder:
   PUT /v1/agent/service/register
   {
     "ID": "user-service-{hostname}",
     "Name": "user-service",
     "Address": "user-service",
     "Port": 5001,
     "Check": {
       "HTTP": "http://user-service:5001/health",
       "Interval": "10s",
       "Timeout": "3s"
     }
   }
3. Consul periyodik olarak /health endpoint'ini çağırarak servisin sağlığını kontrol eder
4. Servis kapanırken Consul'dan kaydını siler (graceful deregistration)
```

### 4.2 Servis Keşfi (Gateway tarafı)

```
1. Gateway'e /api/users/me isteği gelir
2. Route tablosundan "user-service" adını bulur
3. Consul'a sorar: GET /v1/health/service/user-service?passing=true
4. Consul sağlıklı instance listesini döner
5. Gateway round-robin ile birini seçer ve isteği yönlendirir
6. Sonucu istemciye döner
```

---

## 5. JWT Auth Akışı

```
┌──────────┐     POST /api/auth/login      ┌─────────┐
│  Client  │ ─────────────────────────────► │ Gateway │
└──────────┘                                └────┬────┘
                                                 │ (auth gerektirmeyen route, direkt proxy)
                                                 ▼
                                          ┌──────────────┐
                                          │ User Service │
                                          │  1. Email/pw doğrula
                                          │  2. JWT üret (access + refresh)
                                          │  3. Response dön
                                          └──────────────┘

┌──────────┐  GET /api/catalog/movies      ┌─────────┐
│  Client  │  Authorization: Bearer eyJ... │ Gateway │
└──────────┘ ─────────────────────────────►└────┬────┘
                                                │ 1. JWT'yi doğrula (secret ile verify)
                                                │ 2. Token geçerliyse:
                                                │    - X-User-Id: <userId> header ekle
                                                │    - X-User-Role: <role> header ekle
                                                │ 3. Downstream'e yönlendir
                                                ▼
                                         ┌────────────────┐
                                         │ Catalog Service │
                                         │ (X-User-Id header'ından
                                         │  kullanıcıyı tanır)
                                         └────────────────┘
```

**JWT Payload:**

```json
{
  "sub": "user-guid",
  "email": "ali@test.com",
  "role": "User",
  "iss": "streamvault",
  "iat": 1739448000,
  "exp": 1739451600
}
```

---

## 6. Haftalık İlerleme Planı

### Hafta 1: Altyapı + User Service

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Repo + Docker Compose | Proje yapısını oluştur, docker-compose ile PostgreSQL, MongoDB, Redis, Consul ayağa kaldır. Her birinin çalıştığını doğrula. |
| 2 | User Service — Domain + Infrastructure | .NET solution oluştur (4 proje). Domain entity'leri, EF Core DbContext, migration, PostgreSQL bağlantısı. |
| 3 | User Service — Auth (Register + Login) | RegisterUserCommand/Handler, LoginUserCommand/Handler, JwtTokenService, BCryptPasswordHasher. Postman ile test. |
| 4 | User Service — Profile + Watchlist | CRUD operasyonları, max 5 profil validasyonu, watchlist ekleme/silme. |
| 5 | User Service — Consul + Health | Consul'a self-registration, /health endpoint, graceful deregistration. Consul UI'da servisin göründüğünü doğrula. |

### Hafta 2: API Gateway + Catalog Service

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | API Gateway — Temel Proxy | Go projesi oluştur, route tanımları, httputil.ReverseProxy ile user-service'e yönlendirme. Hardcoded adres ile başla. |
| 2 | API Gateway — Auth + Rate Limit | JWT doğrulama middleware, Redis ile token bucket rate limiter. Middleware chain'i kur. |
| 3 | API Gateway — Consul Discovery | Consul client entegrasyonu, servis adı → adres çözümleme, health check aggregation. |
| 4 | Catalog Service — Domain + API | .NET solution, MongoDB bağlantısı, Movie/Series CRUD, sayfalama, filtreleme. |
| 5 | Catalog Service — Seed + Genre | Seed data (20-30 film/dizi), genre CRUD, slug routing. Consul'a kayıt. |

### Hafta 3: Entegrasyon + Test + İyileştirme

| Gün | Görev | Detay |
|-----|-------|-------|
| 1 | Uçtan Uca Test | Tüm akışı test et: register → login → katalog listele → watchlist'e ekle. test-api.sh script'i yaz. |
| 2 | Error Handling | Global exception middleware, hata response formatı standardizasyonu, Gateway'de timeout/retry. |
| 3 | Logging | Structured logging (Serilog .NET, zerolog Go), correlation ID (her request'e unique ID), request/response log. |
| 4 | Unit + Integration Tests | User Service: handler testleri, JWT testleri. Gateway: middleware testleri. En az %60 coverage hedefi. |
| 5 | Dokümantasyon + Refactor | README güncelle, API docs (Swagger), code review, teknik borç temizliği. |

---

## 7. Faz 1 Bitiş Kriterleri (Definition of Done)

Faz 1'in tamamlanmış sayılması için aşağıdaki tüm maddeler sağlanmalıdır:

- [ ] `docker compose up` ile tüm sistem 60 saniye içinde ayağa kalkıyor
- [ ] Kullanıcı kayıt olabiliyor ve JWT token alabiliyor
- [ ] Refresh token ile yeni access token alınabiliyor
- [ ] Profil oluşturulabiliyor (max 5 sınırı çalışıyor)
- [ ] Katalog listesi sayfalama ile getirilebiliyor
- [ ] Türe göre filtreleme çalışıyor
- [ ] Watchlist'e ekleme/çıkarma yapılabiliyor
- [ ] API Gateway tüm istekleri doğru servise yönlendiriyor
- [ ] JWT olmadan korumalı endpoint'lere erişim 401 dönüyor
- [ ] Rate limiter çalışıyor (aşırı istekte 429 dönüyor)
- [ ] Consul UI'da tüm servisler "healthy" görünüyor
- [ ] Seed data ile en az 20 film/dizi yüklü
- [ ] Structured log'lar JSON formatında yazılıyor
- [ ] Unit test coverage minimum %60
- [ ] README.md güncel ve kurulum adımları çalışıyor

---

## 8. Faz 2'ye Hazırlık

Faz 1 tamamlandığında, Faz 2 (Core Streaming) için şu temeller hazır olacak:

- **gRPC Proto dosyaları:** `proto/` klasörü mevcut, Faz 2'de Streaming → User Service arası gRPC eklenecek
- **Service Discovery:** Yeni servisler Consul'a kayıt olarak otomatik keşfedilebilecek
- **Auth altyapısı:** Streaming erişim kontrolü JWT + subscription tier bilgisine dayanacak
- **Catalog entegrasyonu:** Streaming Service, Catalog'daki content ID'lerini referans alacak
- **Docker Compose:** Yeni servisler eklemek `docker-compose.yml`'a blok eklemek kadar kolay

> **Not:** Faz 2'de RabbitMQ eklenmesi ve event-driven architecture'a geçiş, en büyük mimari sıçrama olacak.
