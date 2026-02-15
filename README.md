# StreamVault

A microservices-based streaming platform backend built with .NET 8, Go, Rust, and Docker.

## Architecture

```
                                    ┌─────────────┐
                                    │   Client    │
                                    └──────┬──────┘
                                           │
                                    ┌──────▼──────┐
                                    │ API Gateway │
                                    │  (Go 1.25)  │
                                    │ :8081 (host)│
                                    └──┬──┬──┬──┬─┘
                  ┌────────────────────┘  │  │  └────────────────────┐
                  │    ┌─────────────────┘  └────────────────┐     │
   ┌──────────────┼────┼──────────────────────────────────┐  │     │
   │              │    │                                   │  │     │
┌──▼───────┐ ┌───▼────▼──┐ ┌──────────┐ ┌──────────┐ ┌───▼──▼──┐ ┌▼───────────┐ ┌──────────────┐
│  User    │ │ Catalog   │ │Streaming │ │Encoding  │ │ Search  │ │Recommend.  │ │Subscription  │
│ Service  │ │ Service   │ │ Service  │ │ Service  │ │ Service │ │  Service   │ │  Service     │
│ (.NET 8) │ │ (.NET 8)  │ │  (Go)   │ │ (Rust)   │ │  (Go)  │ │   (Go)    │ │  (.NET 8)    │
│  :5001   │ │  :5002    │ │  :5003  │ │  :5004   │ │ :5005  │ │  :5006    │ │   :5007      │
└──┬───────┘ └──┬────────┘ └──┬──────┘ └──┬───────┘ └──┬─────┘ └──┬───────┘ └──┬───────────┘
   │            │             │           │            │           │            │
   │            │             │           │            │           │            │
┌──▼──────┐ ┌──▼──────┐   ┌──▼───────────▼──┐     ┌───▼────────┐ │         ┌──▼──────┐
│Postgres │ │MongoDB  │   │    RabbitMQ     │     │Elastic     │ │         │Postgres │
│(users)  │ │(catalog)│   │  AMQP :5672    │◄────┤search 8.12 │ │         │(subscr.)│
│  :5432  │ │ :27017  │   │  UI :15672     │     │  :9200     │ │         │  :5432  │
└─────────┘ └─────────┘   └───────┬────────┘     └────────────┘ │         └─────────┘
                                  │                              │
                            ┌─────▼──────┐              ┌───────▼──────┐
                            │   MinIO    │              │  Postgres    │
                            │ Obj.Store  │              │  (recomm.)   │
                            │ API :9000  │              │   :5432      │
                            │Console:9001│              └──────────────┘
                            └────────────┘
                 ┌─────────────┐           ┌─────────────┐
                 │    Redis    │           │   Consul    │
                 │ Cache+State │           │ (Discovery) │
                 │   :6379     │           │   :8500     │
                 └─────────────┘           └─────────────┘
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| API Gateway | Go 1.25, net/http, jwt/v5 |
| User Service | .NET 8, EF Core, PostgreSQL |
| Catalog Service | .NET 8, MongoDB Driver |
| Streaming Service | Go 1.25, HLS, gRPC |
| Encoding Service | Rust, FFmpeg, tonic (gRPC) |
| Search Service | Go 1.25, Elasticsearch 8.12, gRPC |
| Recommendation Service | Go 1.25, PostgreSQL, gRPC |
| Subscription Service | .NET 8, EF Core, PostgreSQL, Saga Pattern |
| Service Discovery | HashiCorp Consul |
| Message Broker | RabbitMQ 3.13 (6 topic exchanges, outbox pattern) |
| Search Engine | Elasticsearch 8.12 (Turkish + English analyzers) |
| Object Storage | MinIO (S3-compatible) |
| Caching & State | Redis (Rate Limiting, Watch Progress, Search Cache, Trending, Recommendations) |
| Architecture Pattern | Clean Architecture + CQRS (MediatR) |
| Event Patterns | Outbox Pattern, Saga Pattern, Event-Driven Sync |
| Containerization | Docker Compose |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

## Quick Start

```bash
# Clone the repository
git clone <repo-url>
cd stream-vault

# Start core services (without Elasticsearch)
docker compose up --build -d

# Start all services including Search (requires Elasticsearch)
docker compose --profile elasticsearch up --build -d

# Verify services are running
docker compose ps
```

Once running, all services will be accessible:

| Service | URL |
|---------|-----|
| API Gateway | http://localhost:8081 |
| User Service | http://localhost:5001 |
| Catalog Service | http://localhost:5002 |
| Streaming Service | http://localhost:5003 |
| Encoding Service | http://localhost:5004 |
| Search Service | http://localhost:5005 |
| Recommendation Service | http://localhost:5006 |
| Subscription Service | http://localhost:5007 |
| Consul UI | http://localhost:8500 |
| RabbitMQ Management | http://localhost:15672 |
| MinIO Console | http://localhost:9001 |
| Elasticsearch | http://localhost:9200 |

## API Documentation (Swagger)

All .NET services expose Swagger UI in development mode:

- **User Service**: http://localhost:5001/swagger
- **Catalog Service**: http://localhost:5002/swagger
- **Subscription Service**: http://localhost:5007/swagger

OpenAPI JSON specs:
- http://localhost:5001/swagger/v1/swagger.json
- http://localhost:5002/swagger/v1/swagger.json
- http://localhost:5007/swagger/v1/swagger.json

## API Examples

All examples use the API Gateway (`localhost:8081`). You can also hit services directly on their individual ports.

### 1. Register a User

```bash
curl -s -X POST http://localhost:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "SecurePass123!"
  }'
```

Response (201):
```json
{
  "accessToken": "eyJhbG...",
  "refreshToken": "eyJhbG...",
  "expiresIn": 3600,
  "user": {
    "id": "550e8400-...",
    "email": "demo@example.com",
    "role": "User"
  }
}
```

### 2. Login

```bash
curl -s -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "SecurePass123!"
  }'
```

Save the `accessToken` from the response for subsequent requests:

```bash
TOKEN="eyJhbG..."
```

### 3. Get Current User

```bash
curl -s http://localhost:8081/api/users/me \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Refresh Token

```bash
curl -s -X POST http://localhost:8081/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "eyJhbG..."
  }'
```

### 5. Create a Profile

```bash
curl -s -X POST http://localhost:8081/api/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "userId": "USER_ID_FROM_LOGIN",
    "name": "Main Profile",
    "icon": 1,
    "isKids": false
  }'
```

### 6. List Movies (paginated)

```bash
# Basic listing
curl -s http://localhost:8081/api/catalog/movies

# With filters
curl -s "http://localhost:8081/api/catalog/movies?genre=Action&sort=releaseYear&page=1&pageSize=10"
```

### 7. List Genres

```bash
curl -s http://localhost:8081/api/catalog/genres
```

### 8. Get Genre Content by Slug

```bash
curl -s "http://localhost:8081/api/catalog/genres/action?page=1&pageSize=10"
```

### 9. Add to Watchlist

```bash
curl -s -X POST http://localhost:8081/api/watchlist \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "profileId": "PROFILE_ID",
    "contentId": "MOVIE_MONGO_ID",
    "contentType": 0
  }'
```

`contentType`: 0 = Movie, 1 = Series, 2 = Episode, 3 = Documentary

### 10. Remove from Watchlist

```bash
curl -s -X DELETE http://localhost:8081/api/watchlist/WATCHLIST_ITEM_ID \
  -H "Authorization: Bearer $TOKEN"
```

### 11. Unauthorized Access (expect 401)

```bash
curl -s http://localhost:8081/api/users/me
```

### 12. Health Check

```bash
curl -s http://localhost:8081/health
```

### 13. Upload Video (Admin)

```bash
# Login as admin first
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@streamvault.io", "password": "Admin123!"}' | jq -r '.accessToken')

# Upload a video file for a catalog content item
curl -s -X POST http://localhost:8081/api/stream/upload \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "content_id=MOVIE_MONGO_ID" \
  -F "file=@/path/to/video.mp4"
```

Response (202):
```json
{
  "job_id": "a1b2c3d4-...",
  "content_id": "MOVIE_MONGO_ID",
  "status": "queued",
  "message": "Upload successful, encoding job queued"
}
```

### 14. Get HLS Master Playlist

```bash
curl -s http://localhost:8081/stream/CONTENT_ID/manifest.m3u8 \
  -H "Authorization: Bearer $TOKEN"
```

Response (200, Content-Type: `application/vnd.apple.mpegurl`):
```
#EXTM3U
#EXT-X-VERSION:3

#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360,NAME="360p"
/stream/CONTENT_ID/360p/playlist.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=2500000,RESOLUTION=1280x720,NAME="720p"
/stream/CONTENT_ID/720p/playlist.m3u8
```

Available qualities depend on subscription tier (see [Subscription Tiers](#subscription-tiers)).

### 15. Get Quality-Specific Playlist

```bash
curl -s http://localhost:8081/stream/CONTENT_ID/720p/playlist.m3u8 \
  -H "Authorization: Bearer $TOKEN"
```

### 16. Get Video Segment

```bash
curl -s http://localhost:8081/stream/CONTENT_ID/720p/segment-0.ts \
  -H "Authorization: Bearer $TOKEN" \
  --output segment.ts
```

### 17. Save Watch Progress

```bash
curl -s -X POST http://localhost:8081/api/stream/CONTENT_ID/progress \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "position_seconds": 300,
    "duration_seconds": 7200
  }'
```

Response (200):
```json
{"status": "saved"}
```

### 18. Get Watch Progress

```bash
curl -s http://localhost:8081/api/stream/CONTENT_ID/progress \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "user_id": "550e8400-...",
  "content_id": "CONTENT_ID",
  "position_seconds": 300,
  "duration_seconds": 7200,
  "percentage": 4.17,
  "updated_at": 1707900000
}
```

### 19. Continue Watching List

```bash
curl -s "http://localhost:8081/api/stream/continue-watching?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "items": [
    {
      "user_id": "550e8400-...",
      "content_id": "CONTENT_ID",
      "position_seconds": 300,
      "duration_seconds": 7200,
      "percentage": 4.17,
      "updated_at": 1707900000
    }
  ]
}
```

### 20. Get Encoding Job Status (Admin)

```bash
curl -s http://localhost:8081/api/encoding/jobs/JOB_ID \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Response (200):
```json
{
  "jobId": "a1b2c3d4-...",
  "contentId": "MOVIE_MONGO_ID",
  "status": "completed",
  "progressPercentage": 100.0,
  "currentStep": "completed",
  "outputs": [
    {
      "quality": "720p",
      "width": 1280,
      "height": 720,
      "bitrateKbps": 2800,
      "fileSizeBytes": 50000,
      "segmentCount": 12,
      "playlistPath": "CONTENT_ID/720p/playlist.m3u8",
      "storagePath": "CONTENT_ID/720p/"
    }
  ],
  "createdAt": "2026-02-14T12:00:00Z",
  "startedAt": "2026-02-14T12:00:01Z",
  "completedAt": "2026-02-14T12:05:30Z"
}
```

Valid status values: `queued`, `processing`, `completed`, `failed`, `cancelled`

### 21. List Encoding Jobs (Admin)

```bash
# All jobs
curl -s http://localhost:8081/api/encoding/jobs \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Filter by status with pagination
curl -s "http://localhost:8081/api/encoding/jobs?status=processing&limit=10&offset=0" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Response (200):
```json
{
  "jobs": [],
  "totalCount": 0
}
```

Query parameters: `status`, `content_id`, `limit` (max 100), `offset`

### 22. Rate Content

```bash
curl -s -X POST http://localhost:8081/api/users/me/ratings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "contentId": "MOVIE_MONGO_ID",
    "rating": 8.5
  }'
```

### 23. Get My Ratings

```bash
curl -s http://localhost:8081/api/users/me/ratings \
  -H "Authorization: Bearer $TOKEN"
```

### 24. Full-Text Search

```bash
curl -s "http://localhost:8081/api/search?q=inception&page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "items": [
    {
      "contentId": "...",
      "title": "Inception",
      "description": "A thief who steals...",
      "contentType": "movie",
      "thumbnailUrl": "/thumbnails/...",
      "releaseYear": 2010,
      "averageRating": 8.8,
      "maturityRating": "PG-13",
      "genres": ["Sci-Fi", "Action", "Thriller"],
      "highlights": {
        "title": "<em>Inception</em>",
        "description": "...corporate espionage using dream-sharing technology..."
      }
    }
  ],
  "facets": {
    "genreFacets": [{"key": "Sci-Fi", "docCount": 5}],
    "yearFacets": [{"key": "2010", "docCount": 3}],
    "maturityFacets": [{"key": "PG-13", "docCount": 8}],
    "contentTypeFacets": [{"key": "movie", "docCount": 12}]
  },
  "page": 1,
  "pageSize": 10,
  "totalCount": 1,
  "totalPages": 1
}
```

### 25. Search with Filters

```bash
curl -s "http://localhost:8081/api/search?q=&genres=Action,Sci-Fi&year_from=2010&year_to=2024&min_rating=7.0&content_type=movie&sort=rating&order=desc&page=1&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"
```

Filter parameters: `genres` (comma-separated), `year_from`, `year_to`, `min_rating`, `maturity_ratings`, `content_type` (movie/series). Sort options: `relevance`, `rating`, `year`, `title`.

### 26. Autocomplete

```bash
curl -s "http://localhost:8081/api/search/autocomplete?q=inc&limit=5" \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "suggestions": [
    {
      "contentId": "...",
      "title": "Inception",
      "contentType": "movie",
      "thumbnailUrl": "/thumbnails/...",
      "releaseYear": 2010
    }
  ]
}
```

Minimum 2 characters required. Default limit: 5, max: 20.

### 27. Trending Content (Public)

```bash
curl -s "http://localhost:8081/api/search/trending?window=week&limit=10"
```

Response (200):
```json
{
  "items": [
    {
      "contentId": "...",
      "title": "The Dark Knight",
      "thumbnailUrl": "/thumbnails/...",
      "contentType": "movie",
      "releaseYear": 2008,
      "averageRating": 9.0,
      "rank": 1,
      "rankChange": 2,
      "viewCount": 5000
    }
  ]
}
```

Window options: `day`, `week` (default), `month`.

### 28. Get Personalized Recommendations

```bash
curl -s "http://localhost:8081/api/recommendations?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "items": [
    {
      "contentId": "...",
      "title": "Interstellar",
      "thumbnailUrl": "/thumbnails/...",
      "contentType": "movie",
      "releaseYear": 2014,
      "averageRating": 8.6,
      "score": 0.92,
      "algorithm": "hybrid",
      "reason": "Based on your watch history"
    }
  ]
}
```

Algorithm values: `hybrid`, `collaborative`, `content-based`, `popularity`. Users with fewer than 5 interactions receive popularity-based recommendations (cold start).

### 29. Get Similar Content

```bash
curl -s "http://localhost:8081/api/recommendations/similar/CONTENT_ID?limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "items": [
    {
      "contentId": "...",
      "title": "The Prestige",
      "thumbnailUrl": "/thumbnails/...",
      "contentType": "movie",
      "releaseYear": 2006,
      "averageRating": 8.5,
      "similarityScore": 0.85
    }
  ]
}
```

### 30. Home Page Sections

```bash
curl -s http://localhost:8081/api/recommendations/home \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "sections": [
    {
      "sectionType": "personal",
      "title": "Recommended for You",
      "items": [{"contentId": "...", "title": "...", "score": 0.92}]
    },
    {
      "sectionType": "trending",
      "title": "Trending Now",
      "items": []
    },
    {
      "sectionType": "because_you_watched",
      "title": "Because You Watched Inception",
      "items": []
    },
    {
      "sectionType": "genre",
      "title": "Top in Sci-Fi",
      "items": []
    },
    {
      "sectionType": "new",
      "title": "New Releases",
      "items": []
    }
  ]
}
```

Anonymous users (no token) receive only `trending` and `new` sections.

### 31. List Subscription Plans (Public)

```bash
curl -s http://localhost:8081/api/plans
```

Response (200):
```json
[
  {
    "id": "...",
    "name": "Basic",
    "tier": "Basic",
    "priceMonthly": 49.99,
    "maxScreens": 1,
    "maxQuality": "720p",
    "features": "SD + HD streaming, 1 screen",
    "isActive": true
  },
  {
    "id": "...",
    "name": "Standard",
    "tier": "Standard",
    "priceMonthly": 79.99,
    "maxScreens": 2,
    "maxQuality": "1080p",
    "features": "Full HD streaming, 2 screens",
    "isActive": true
  },
  {
    "id": "...",
    "name": "Premium",
    "tier": "Premium",
    "priceMonthly": 119.99,
    "maxScreens": 4,
    "maxQuality": "4K",
    "features": "4K + HDR streaming, 4 screens",
    "isActive": true
  }
]
```

### 32. Create Subscription (Saga)

```bash
PLAN_ID="<plan-guid-from-plans-response>"

curl -s -X POST http://localhost:8081/api/subscriptions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "planId": "'$PLAN_ID'",
    "cardNumber": "4242424242424242"
  }'
```

Response (201):
```json
{
  "subscriptionId": "...",
  "userId": "...",
  "planId": "...",
  "planName": "Standard",
  "tier": "Standard",
  "status": "Active",
  "periodStart": "2026-02-14T00:00:00Z",
  "periodEnd": "2026-03-16T00:00:00Z",
  "amountCharged": 79.99,
  "currency": "TRY"
}
```

This triggers a 6-step Saga (see [Saga Pattern](#saga-pattern)). Test card numbers:

| Card Number | Result |
|-------------|--------|
| `4242424242424242` | Payment succeeds |
| `4000000000000002` | Declined (insufficient funds) |
| `4000000000000069` | Expired card |
| `4000000000000127` | General error |

### 33. Get My Subscription

```bash
curl -s http://localhost:8081/api/subscriptions/me \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "id": "...",
  "userId": "...",
  "plan": {
    "id": "...",
    "name": "Standard",
    "tier": "Standard",
    "priceMonthly": 79.99,
    "maxScreens": 2,
    "maxQuality": "1080p",
    "features": "Full HD streaming, 2 screens",
    "isActive": true
  },
  "status": "Active",
  "periodStart": "2026-02-14T00:00:00Z",
  "periodEnd": "2026-03-16T00:00:00Z",
  "autoRenew": true,
  "cancelledAt": null,
  "createdAt": "2026-02-14T00:00:00Z"
}
```

### 34. Change Plan (Upgrade/Downgrade)

```bash
NEW_PLAN_ID="<new-plan-guid>"

curl -s -X PUT http://localhost:8081/api/subscriptions/me/plan \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "newPlanId": "'$NEW_PLAN_ID'",
    "cardNumber": "4242424242424242"
  }'
```

Response (200):
```json
{
  "subscriptionId": "...",
  "oldPlanId": "...",
  "oldPlanName": "Standard",
  "newPlanId": "...",
  "newPlanName": "Premium",
  "newTier": "Premium",
  "priceDifference": 40.00,
  "status": "Active"
}
```

Upgrade charges prorated difference: `(newDailyRate - oldDailyRate) * remainingDays`. Downgrade creates a credit.

### 35. Cancel Subscription

```bash
curl -s -X POST http://localhost:8081/api/subscriptions/me/cancel \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "subscriptionId": "...",
  "status": "Cancelled",
  "cancelledAt": "2026-02-14T10:00:00Z",
  "periodEnd": "2026-03-16T00:00:00Z"
}
```

Subscription remains active until `periodEnd`. The user's tier is downgraded to Free via the `subscription.events` exchange.

### 36. Get Invoices

```bash
curl -s http://localhost:8081/api/subscriptions/me/invoices \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
[
  {
    "id": "...",
    "subscriptionId": "...",
    "invoiceNumber": "INV-20260214-...",
    "amount": 79.99,
    "currency": "TRY",
    "periodStart": "2026-02-14T00:00:00Z",
    "periodEnd": "2026-03-16T00:00:00Z",
    "issuedAt": "2026-02-14T00:00:00Z"
  }
]
```

### 37. Get Payment History

```bash
curl -s "http://localhost:8081/api/subscriptions/me/payments?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

## Project Structure

```
stream-vault/
├── gateway/                    # API Gateway (Go)
│   ├── cmd/gateway/            # Entry point
│   ├── internal/
│   │   ├── config/             # Viper configuration
│   │   ├── discovery/          # Consul service resolver
│   │   ├── health/             # Health check handler
│   │   ├── middleware/         # Auth, CORS, rate limiting, streaming auth
│   │   └── proxy/              # Reverse proxy, routing, upstream
│   ├── proto/                  # Generated gRPC stubs
│   └── config.yaml             # Gateway configuration
├── services/
│   ├── user-service/           # User Service (.NET 8)
│   │   ├── src/
│   │   │   ├── UserService.Api/           # Controllers, middleware
│   │   │   ├── UserService.Application/   # CQRS commands/queries
│   │   │   ├── UserService.Domain/        # Entities, value objects
│   │   │   └── UserService.Infrastructure/# EF Core, repositories, consumers
│   │   └── tests/
│   ├── catalog-service/        # Catalog Service (.NET 8)
│   │   ├── src/
│   │   │   ├── CatalogService.Api/
│   │   │   ├── CatalogService.Application/
│   │   │   ├── CatalogService.Domain/
│   │   │   └── CatalogService.Infrastructure/  # Includes outbox pattern
│   │   └── tests/
│   ├── streaming-service/      # Streaming Service (Go)
│   │   ├── cmd/streaming/      # Entry point
│   │   ├── internal/
│   │   │   ├── handler/        # HTTP handlers (upload, manifest, progress)
│   │   │   ├── hls/            # HLS playlist generation + quality profiles
│   │   │   ├── messaging/      # RabbitMQ publisher + consumer
│   │   │   ├── middleware/     # Admin, subscription, concurrent streams
│   │   │   ├── progress/       # Redis-backed watch progress + WatchCompleted events
│   │   │   └── storage/        # MinIO storage interface
│   │   ├── proto/              # Generated gRPC stubs
│   │   └── config.yaml
│   ├── encoding-service/       # Encoding Service (Rust)
│   │   ├── src/
│   │   │   ├── api/            # Axum HTTP handlers + routes
│   │   │   ├── domain/         # Job, Status, Profile models
│   │   │   ├── grpc.rs         # Tonic gRPC server
│   │   │   ├── messaging/      # RabbitMQ consumer + publisher
│   │   │   ├── pipeline/       # FFmpeg transcoder, thumbnails, orchestrator
│   │   │   └── storage/        # MinIO client
│   │   ├── proto/              # Proto source files for tonic-build
│   │   └── config.toml
│   ├── search-service/         # Search Service (Go)
│   │   ├── cmd/search/         # Entry point
│   │   ├── internal/
│   │   │   ├── cache/          # Redis-backed search + trending cache
│   │   │   ├── config/         # Viper configuration
│   │   │   ├── consumer/       # RabbitMQ consumers (catalog sync, watch count)
│   │   │   ├── discovery/      # Consul registration
│   │   │   ├── elasticsearch/  # ES client, indexer, searcher, mapping
│   │   │   ├── grpcserver/     # gRPC SearchService implementation
│   │   │   ├── handler/        # HTTP handlers (search, autocomplete, trending)
│   │   │   ├── middleware/     # Recovery, request ID
│   │   │   ├── model/          # Request/response models
│   │   │   └── trending/       # Redis sorted set trending tracker
│   │   ├── proto/              # Generated gRPC stubs
│   │   └── config.yaml
│   ├── recommendation-service/ # Recommendation Service (Go)
│   │   ├── cmd/recommendation/ # Entry point
│   │   ├── internal/
│   │   │   ├── batch/          # 6h periodic recomputation scheduler
│   │   │   ├── cache/          # Redis feature/recommendation cache
│   │   │   ├── config/         # Viper configuration
│   │   │   ├── consumer/       # RabbitMQ consumers (watch, rating, catalog)
│   │   │   ├── discovery/      # Consul registration
│   │   │   ├── engine/         # Hybrid recommendation engine
│   │   │   │   ├── content_based.go   # Content-based filtering
│   │   │   │   ├── collaborative.go   # User-based collaborative filtering
│   │   │   │   ├── popularity.go      # Popularity fallback (cold start)
│   │   │   │   ├── hybrid.go          # Weighted hybrid scorer
│   │   │   │   └── similarity.go      # Cosine similarity
│   │   │   ├── grpcserver/     # gRPC RecommendationService implementation
│   │   │   ├── handler/        # HTTP handlers (recommendations, similar, home)
│   │   │   ├── middleware/     # Recovery, request ID
│   │   │   ├── model/          # Domain models (interaction, profile, features)
│   │   │   └── repository/     # PostgreSQL repository + embedded migrations
│   │   ├── proto/              # Generated gRPC stubs
│   │   └── config.yaml
│   └── subscription-service/   # Subscription Service (.NET 8)
│       ├── src/
│       │   ├── SubscriptionService.Api/           # Controllers, background services
│       │   ├── SubscriptionService.Application/   # Commands, queries, sagas
│       │   ├── SubscriptionService.Domain/        # Entities, enums, events
│       │   └── SubscriptionService.Infrastructure/# EF Core, repositories, mock payment
│       └── tests/
├── proto/                      # Shared proto definitions
│   ├── common/v1/              # Pagination, ContentType, SubscriptionTier
│   ├── catalog/v1/             # CatalogService RPCs
│   ├── encoding/v1/            # EncodingService RPCs
│   ├── search/v1/              # SearchService RPCs
│   ├── recommendation/v1/      # RecommendationService RPCs
│   ├── streaming/v1/           # StreamingService RPCs
│   ├── subscription/v1/        # SubscriptionQueryService RPCs
│   └── user/v1/                # UserService RPCs
├── infrastructure/
│   ├── consul/                 # Consul configuration
│   ├── elasticsearch/          # ES index mapping + init script
│   ├── mongo/                  # MongoDB init scripts
│   ├── postgres/               # PostgreSQL init (users, subscriptions, recommendations DBs)
│   ├── rabbitmq/               # RabbitMQ config + topology (6 exchanges, 10 queues)
│   ├── redis/                  # Redis configuration
│   └── minio/                  # MinIO bucket init script
├── scripts/                    # Utility scripts
│   ├── generate-proto.sh       # Proto code generation (Go + Rust)
│   ├── upload-test-video.sh    # Upload test video + verify encoding
│   ├── test-streaming.sh       # Streaming flow integration tests
│   ├── test-search.sh          # Search service integration tests
│   ├── test-subscription.sh    # Subscription saga integration tests
│   ├── seed-search-index.sh    # Bulk index catalog data into Elasticsearch
│   └── generate-interactions.sh # Generate fake watch/rating data for recommendations
├── docker-compose.yml          # Service definitions
└── docker-compose.override.yml # Development port mappings
```

## Port Reference

| Service | Container Port | Host Port |
|---------|---------------|-----------|
| API Gateway | 8080 | 8081 |
| User Service | 8080 | 5001 |
| Catalog Service | 5100 | 5002 |
| Streaming Service (HTTP) | 5003 | 5003 |
| Streaming Service (gRPC) | 50051 | 50051 |
| Encoding Service (HTTP) | 5004 | 5004 |
| Encoding Service (gRPC) | 50052 | 50052 |
| Search Service (HTTP) | 5005 | 5005 |
| Search Service (gRPC) | 50053 | 50053 |
| Recommendation Service (HTTP) | 5006 | 5006 |
| Recommendation Service (gRPC) | 50054 | 50054 |
| Subscription Service | 5007 | 5007 |
| PostgreSQL | 5432 | 5432 |
| MongoDB | 27017 | 27017 |
| Redis | 6379 | 6379 |
| Elasticsearch | 9200 | 9200 |
| Consul | 8500 | 8500 |
| RabbitMQ (AMQP) | 5672 | 5672 |
| RabbitMQ (Management UI) | 15672 | 15672 |
| MinIO (API) | 9000 | 9000 |
| MinIO (Console) | 9001 | 9001 |

## Key Features

- **JWT Authentication** with access + refresh token flow
- **Role-based access control** (User / Admin)
- **Multi-profile support** (up to 5 profiles per user)
- **Watchlist management** per profile
- **Content ratings** per user with event publishing
- **Movie & Series catalog** with genres, filtering, pagination, and sorting
- **Service discovery** via Consul with health checks
- **Rate limiting** (IP-based token bucket via Redis)
- **Correlation ID** propagation across services
- **Structured logging** (Serilog for .NET, zerolog for Go, tracing for Rust)
- **Consistent error responses** across all services
- **HLS adaptive bitrate streaming** (360p / 720p / 1080p / 4K)
- **Video upload & encoding pipeline** (FFmpeg-based multi-quality transcoding)
- **Subscription tier-based quality filtering** (Basic→720p, Standard→1080p, Premium→4K)
- **Watch progress tracking** with continue-watching list (Redis)
- **Concurrent stream limiting** per subscription tier (1 / 2 / 4 streams)
- **Event-driven encoding** via RabbitMQ (topic exchange with dead-letter support)
- **S3-compatible object storage** (MinIO) for raw uploads, encoded HLS, and thumbnails
- **gRPC inter-service communication** with shared proto definitions
- **Full-text search** with Turkish + English analyzers (Elasticsearch)
- **Faceted search** with genre, year, maturity rating, and content type filters
- **Autocomplete** with edge n-gram prefix matching (2-15 characters)
- **Trending content** tracking with daily/weekly/monthly Redis sorted sets
- **Personalized recommendations** via hybrid engine (content-based + collaborative filtering + popularity)
- **Content similarity** scoring using cosine similarity on feature vectors
- **Cold start handling** with popularity-based fallback (under 5 interactions)
- **Home page sections** (personal, trending, because you watched, genre, new releases)
- **Subscription management** with Saga pattern (create, change plan, cancel)
- **Mock payment gateway** with test card simulation
- **Proration on plan changes** (upgrade charges difference, downgrade credits)
- **Outbox pattern** for reliable event publishing (Catalog + Subscription services)
- **Event-driven data synchronization** across Search, Recommendation, and User services

## RabbitMQ Topology

All topology is pre-configured via `infrastructure/rabbitmq/definitions.json`. The system uses 6 exchanges and 10 queues for event-driven communication between services.

### Exchanges

| Exchange | Type | Durable | Purpose |
|----------|------|---------|---------|
| `encoding` | topic | Yes | Encoding job lifecycle (new, completed, failed) |
| `encoding.dlx` | direct | Yes | Dead-letter exchange for failed encoding messages |
| `catalog.events` | topic | Yes | Catalog content changes (created, updated, deleted) |
| `watch.events` | topic | Yes | Watch completion events from Streaming Service |
| `user.events` | topic | Yes | User rating events from User Service |
| `subscription.events` | topic | Yes | Subscription lifecycle (created, cancelled, plan changed) |

### Queues & Bindings

| Queue | Exchange | Routing Key(s) | Consumer | Purpose |
|-------|----------|----------------|----------|---------|
| `encoding.jobs` | `encoding` | `job.new` | Encoding Service | New encoding job requests |
| `encoding.results.catalog` | `encoding` | `job.completed`, `job.failed` | Catalog Service | Updates video status |
| `encoding.results.streaming` | `encoding` | `job.completed`, `job.failed` | Streaming Service | Updates stream info cache |
| `encoding.dead-letters` | `encoding.dlx` | `dead-letter` | -- | Failed messages for inspection |
| `search.catalog-sync` | `catalog.events` | `content.created`, `content.updated`, `content.deleted` | Search Service | ES index sync |
| `recommendation.catalog` | `catalog.events` | `content.created` | Recommendation Service | Content feature extraction |
| `search.watch-count` | `watch.events` | `watch.completed` | Search Service | View count + trending update |
| `recommendation.watch` | `watch.events` | `watch.completed` | Recommendation Service | Watch interaction recording |
| `recommendation.ratings` | `user.events` | `content.rated` | Recommendation Service | Rating interaction recording |
| `user.subscription-sync` | `subscription.events` | `subscription.created`, `subscription.cancelled`, `plan.changed` | User Service | Tier synchronization |

### Message Flows

**Encoding Pipeline:**
```
Streaming Service ──job.new──► [encoding] ──► encoding.jobs ──► Encoding Service
                                  │
                                  ├─job.completed──► encoding.results.catalog ──► Catalog Service
                                  └─job.completed──► encoding.results.streaming ──► Streaming Service
```

**Catalog Sync (Outbox Pattern):**
```
Catalog Service ──outbox──► [catalog.events]
                                  │
                                  ├─content.created/updated/deleted──► search.catalog-sync ──► Search Service (ES index)
                                  └─content.created──► recommendation.catalog ──► Recommendation Service
```

**Watch Events:**
```
Streaming Service ──watch.completed──► [watch.events]
                                           │
                                           ├──► search.watch-count ──► Search Service (trending)
                                           └──► recommendation.watch ──► Recommendation Service
```

**Rating Events:**
```
User Service ──content.rated──► [user.events] ──► recommendation.ratings ──► Recommendation Service
```

**Subscription Events (Outbox Pattern):**
```
Subscription Service ──outbox──► [subscription.events]
                                       │
                                       └─subscription.created/cancelled/plan.changed──► user.subscription-sync ──► User Service (tier update)
```

## MinIO Bucket Structure

Three buckets are created automatically by `infrastructure/minio/init-buckets.sh` on first startup.

| Bucket | Access | Purpose |
|--------|--------|---------|
| `streamvault-raw` | Private | Original uploaded video files |
| `streamvault-encoded` | Private | HLS-encoded output (playlists + segments) |
| `streamvault-thumbnails` | Public (anonymous download) | Generated poster images and thumbnails |

### Object Key Layout

```
streamvault-raw/
└── {contentId}/
    └── original.mp4

streamvault-encoded/
└── {contentId}/
    ├── 360p/
    │   ├── playlist.m3u8
    │   ├── segment-0.ts
    │   └── segment-N.ts
    ├── 720p/
    │   ├── playlist.m3u8
    │   └── ...
    ├── 1080p/
    │   └── ...
    └── 4k/
        └── ...

streamvault-thumbnails/
└── {contentId}/
    ├── poster.jpg
    └── thumb_300x170.jpg
```

## Subscription Tiers

| Tier | Max Streaming Quality | Concurrent Streams | Price (TRY/month) |
|------|----------------------|-------------------|-------------------|
| Free | — (streaming not available) | 1 | — |
| Basic | 360p + 720p | 1 | 49.99 |
| Standard | 360p + 720p + 1080p | 2 | 79.99 |
| Premium | All qualities (incl. 4K) | 4 | 119.99 |
| Admin | All qualities (incl. 4K) | 4 | — |

Plans are managed by the Subscription Service. Use `GET /api/plans` to list plans and `POST /api/subscriptions` to subscribe. Tier changes propagate automatically to the User Service via the `subscription.events` exchange, updating streaming quality access and concurrent stream limits.

## Elasticsearch

The Search Service uses Elasticsearch 8.12 for full-text search, faceted filtering, and autocomplete. Index mapping is initialized by `infrastructure/elasticsearch/init-index.sh`.

### Index: `streamvault-content`

Settings: 1 shard, 0 replicas (development configuration).

### Custom Analyzers

| Analyzer | Tokenizer | Filters | Purpose |
|----------|-----------|---------|---------|
| `turkish_analyzer` | standard | turkish_lowercase, turkish_stop, turkish_stemmer | Full-text search with Turkish language support |
| `autocomplete_analyzer` | standard | lowercase, edge_ngram (2-15 chars) | Index-time autocomplete token generation |
| `autocomplete_search_analyzer` | standard | lowercase | Search-time query (no edge n-gram expansion) |

### Field Mapping

| Field | Type | Analyzer | Notes |
|-------|------|----------|-------|
| `id` | keyword | — | Content unique identifier |
| `title` | text | turkish_analyzer | + `autocomplete`, `keyword`, `english` sub-fields |
| `original_title` | text | english | + `autocomplete` sub-field |
| `description` | text | turkish_analyzer | + `english` sub-field |
| `content_type` | keyword | — | `movie` / `series` filter |
| `genres` | keyword | — | Genre facet filtering (array) |
| `tags` | keyword | — | Tag filtering (array) |
| `cast_names` | text | standard | Cast member search |
| `director` | text | standard | Director search |
| `release_year` | integer | — | Year range filtering |
| `maturity_rating` | keyword | — | PG, PG-13, R filter |
| `average_rating` | float | — | Rating filter + sort |
| `rating_count` | integer | — | Number of ratings |
| `view_count` | long | — | View count for trending |
| `video_status` | keyword | — | Video processing status |
| `thumbnail_url` | keyword | — | Not indexed, display only |
| `duration_seconds` | integer | — | Content duration |
| `season_count` | integer | — | Series only |
| `episode_count` | integer | — | Series only |
| `created_at` | date | — | New content sorting |
| `updated_at` | date | — | Last modification date |

## Outbox Pattern

The Outbox Pattern ensures reliable event publishing without distributed transactions. Instead of publishing to RabbitMQ directly within a business operation, events are written to a local outbox table/collection in the same database transaction. A background processor then polls the outbox and publishes to RabbitMQ.

### Services Using Outbox

| Service | Outbox Storage | Exchange | Event Types |
|---------|---------------|----------|-------------|
| Catalog Service | MongoDB collection (`outbox_messages`) | `catalog.events` | `content.created`, `content.updated`, `content.deleted` |
| Subscription Service | PostgreSQL table (`OutboxMessages`) | `subscription.events` | `subscription.created`, `subscription.cancelled`, `plan.changed` |

### Processing

- **Poll interval:** 5 seconds
- **Batch size:** 50 messages per poll
- **Max retries:** 3 attempts per message
- **On success:** Sets `ProcessedAt` timestamp
- **On failure:** Increments `RetryCount`, logs error
- **After max retries:** Message is logged but not retried (available for manual inspection)

### Flow

```
Business Operation (e.g., Create Movie, Create Subscription)
       │
       ▼
┌─────────────────────────┐
│    Same Transaction     │
│  ┌───────────────────┐  │
│  │ Save Domain Data  │  │
│  └────────┬──────────┘  │
│           │              │
│  ┌────────▼──────────┐  │
│  │ Write Outbox Msg  │  │
│  └───────────────────┘  │
└─────────────────────────┘
       │
       │  Background Processor (every 5s)
       ▼
┌──────────────────┐      ┌───────────────┐      ┌───────────────┐
│ Read Unprocessed │─────►│ Publish to    │─────►│ Mark as       │
│ Outbox Messages  │      │ RabbitMQ      │      │ Processed     │
└──────────────────┘      └───────────────┘      └───────────────┘
```

**Guarantees:** At-least-once delivery. Consumers must be idempotent to handle potential duplicate messages.

## Saga Pattern

The Subscription Service uses the Saga pattern (orchestration-based) to coordinate multi-step business operations. Each saga step is tracked in a `SagaState` entity, and compensating actions are triggered on failure.

### CreateSubscription Saga (6 Steps)

```
Step 1: VALIDATE_PLAN
  ├── Verify plan exists and is active
  └── Check user has no active subscription
         │
Step 2: CREATE_SUBSCRIPTION
  ├── Create Subscription record (Status: PendingPayment)
  └── Compensate on failure: delete subscription
         │
Step 3: PROCESS_PAYMENT
  ├── Call MockPaymentGateway.ChargeAsync()
  ├── Create Payment record
  └── Compensate on failure: refund payment, mark subscription Failed
         │
Step 4: ACTIVATE_SUBSCRIPTION
  ├── Set Status = Active
  └── Set PeriodStart/PeriodEnd (30 days)
         │
Step 5: CREATE_INVOICE
  └── Create Invoice record with invoice number
         │
Step 6: PUBLISH_EVENTS
  ├── Write SubscriptionCreated event to Outbox
  └── → User Service updates user tier via subscription.events exchange
```

### ChangePlan Saga (4 Steps)

```
Step 1: VALIDATE_CHANGE
  ├── Verify new plan exists and is active
  ├── Verify user has active subscription
  └── Calculate proration (remaining days, price difference)
         │
Step 2: PROCESS_PRICE_DIFFERENCE
  ├── Upgrade: charge (newDailyRate - oldDailyRate) × remainingDays
  └── Downgrade: create credit Payment record
         │
Step 3: UPDATE_SUBSCRIPTION
  └── Update PlanId to new plan
         │
Step 4: PUBLISH_CHANGE_EVENTS
  └── Write PlanChanged event to Outbox
```

### Compensating Actions

On any saga step failure:
1. Saga status is set to `Compensating`
2. If a payment transaction exists, it is refunded via the payment gateway
3. If a subscription was created, its status is set to `Failed`
4. Saga status is set to `Failed`

All saga steps are **idempotent** -- if a saga restarts, it checks for existing resources (subscription, payment, invoice) before creating new ones.

### Cancel Subscription

Cancellation does **not** use a saga since it is a simple atomic operation:
1. Set subscription status to `Cancelled` and `AutoRenew = false`
2. Write outbox message in the same transaction
3. Subscription remains active until `PeriodEnd`

### Mock Payment Test Cards

| Card Number | Result |
|-------------|--------|
| `4242424242424242` | Payment succeeds |
| `4000000000000002` | Declined (insufficient funds) |
| `4000000000000069` | Expired card |
| `4000000000000127` | General error |
