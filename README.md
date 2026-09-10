# StreamVault

A microservices-based streaming platform backend built with .NET 8, Go, Rust, and Docker.

## Architecture

```
                                    ┌─────────────┐
                                    │   Client    │
                                    └──────┬──────┘
                                           │
                                    ┌──────▼──────┐
                                    │ API Gateway │──── Circuit Breaker, Retry,
                                    │  (Go 1.25)  │     Bulkhead, Rate Limiting
                                    │ :8081 (host)│
                                    └──┬──┬──┬──┬─┘
                  ┌────────────────────┘  │  │  └─────────────────────────┐
                  │    ┌─────────────────┘  └────────────────┐           │
   ┌──────────────┼────┼──────────────────────────────────┐  │           │
   │              │    │                                   │  │           │
┌──▼───────┐ ┌───▼────▼──┐ ┌──────────┐ ┌──────────┐ ┌───▼──▼──┐ ┌─────▼──────┐ ┌──────────┐ ┌──────────┐
│  User    │ │ Catalog   │ │Streaming │ │Encoding  │ │ Search  │ │Recommend.  │ │Subscript.│ │Notific.  │
│ Service  │ │ Service   │ │ Service  │ │ Service  │ │ Service │ │  Service   │ │ Service  │ │ Service  │
│ (.NET 8) │ │ (.NET 8)  │ │  (Go)    │ │ (Rust)   │ │  (Go)   │ │   (Go)     │ │ (.NET 8) │ │  (Go)    │
│  :5001   │ │  :5002    │ │  :5003   │ │  :5004   │ │  :5005  │ │  :5006     │ │  :5007   │ │  :5008   │
└──┬───────┘ └──┬────────┘ └──┬──────┘ └──┬───────┘ └──┬──────┘ └──┬────────┘ └──┬───────┘ └──┬───────┘
   │            │             │           │            │            │             │             │
   │            │             │           │            │            │             │             │
┌──▼──────┐ ┌──▼──────┐   ┌──▼───────────▼────────────▼────────────▼─────────────▼─────────────▼──┐
│Postgres │ │MongoDB  │   │                        RabbitMQ                                       │
│(users)  │ │(catalog)│   │   AMQP :5672  │  Management UI :15672  │  6 exchanges, 13 queues     │
│  :5432  │ │ :27017  │   └────────────────────────────┬──────────────────────────────────────────┘
└─────────┘ └─────────┘                                │
                                                       │
   ┌──────────────┐  ┌──────────────┐  ┌───────────────┤  ┌──────────────┐  ┌──────────────┐
   │   MinIO      │  │ Elastic      │  │               │  │  Postgres    │  │  Postgres    │
   │  Obj.Store   │  │ search 8.12  │  │    Redis      │  │  (recomm.)   │  │  (subscr.)   │
   │ API :9000    │  │   :9200      │  │  Cache+State  │  │   :5432      │  │   :5432      │
   │Console :9001 │  │              │  │    :6379      │  │              │  │              │
   └──────────────┘  └──────────────┘  └───────────────┘  └──────────────┘  └──────────────┘

                            ┌──────────────────────────────┐
                            │    Service Discovery         │
                            │    Consul  :8500              │
                            └──────────────────────────────┘

 ┌─────────────────────── Observability Stack (optional profile) ──────────────────────────┐
 │                                                                                         │
 │  ┌──────────┐  ┌────────────┐  ┌──────────┐  ┌────────────────────┐  ┌──────────────┐  │
 │  │  Jaeger  │  │ Prometheus │  │ Grafana  │  │ Loki              │  │  Promtail    │  │
 │  │ Tracing  │  │  Metrics   │  │Dashboards│  │ Log Aggregation   │  │ Log Shipper  │  │
 │  │ :16686   │  │   :9090    │  │  :3000   │  │     :3100         │  │    :9080     │  │
 │  └──────────┘  └────────────┘  └──────────┘  └────────────────────┘  └──────────────┘  │
 │                                                                                         │
 └─────────────────────────────────────────────────────────────────────────────────────────┘
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
| Notification Service | Go 1.25, WebSocket (gorilla/websocket), MongoDB |
| Service Discovery | HashiCorp Consul |
| Message Broker | RabbitMQ 3.13 (6 topic exchanges, 13 queues, outbox pattern) |
| Search Engine | Elasticsearch 8.12 (Turkish + English analyzers) |
| Object Storage | MinIO (S3-compatible) |
| Caching & State | Redis (Rate Limiting, Watch Progress, Search Cache, Trending, Recommendations) |
| Distributed Tracing | OpenTelemetry SDK + Jaeger (OTLP gRPC/HTTP) |
| Metrics & Dashboards | Prometheus + Grafana (4 dashboards, alert rules) |
| Centralized Logging | Loki + Promtail (structured JSON, LogQL) |
| Resilience | Circuit Breaker (gobreaker), Retry, Bulkhead, Graceful Degradation |
| Architecture Pattern | Clean Architecture + CQRS (MediatR) |
| Event Patterns | Outbox Pattern, Saga Pattern, Event-Driven Sync |
| Containerization | Docker Compose (21 containers, 3 optional profiles) |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

## Quick Start

```bash
# Clone the repository
git clone <repo-url>
cd stream-vault

# Create local configuration, then fill every required blank secret.
cp .env.example .env

# Start core services (without Elasticsearch or Observability)
docker compose up --build -d

# Start with Search (Elasticsearch)
docker compose --profile elasticsearch up --build -d

# Start with Observability (Jaeger, Prometheus, Grafana, Loki, Promtail)
docker compose --profile observability up --build -d

# Start everything (all 21 containers)
docker compose --profile elasticsearch --profile observability up --build -d

# Verify services are running
docker compose ps
```

Once running, all services will be accessible:

| Service | URL | Notes |
|---------|-----|-------|
| API Gateway | http://localhost:8081 | Main entry point |
| User Service | http://localhost:5001 | |
| Catalog Service | http://localhost:5002 | |
| Streaming Service | http://localhost:5003 | |
| Encoding Service | http://localhost:5004 | |
| Search Service | http://localhost:5005 | `elasticsearch` profile |
| Recommendation Service | http://localhost:5006 | |
| Subscription Service | http://localhost:5007 | |
| Notification Service | http://localhost:5008 | WebSocket: `ws://localhost:5008/ws/notifications` |
| Consul UI | http://localhost:8500 | |
| RabbitMQ Management | http://localhost:15672 | Credentials from `RABBITMQ_USER` / `RABBITMQ_PASS` |
| MinIO Console | http://localhost:9001 | minioadmin / minioadmin |
| Elasticsearch | http://localhost:9200 | `elasticsearch` profile |
| Grafana | http://localhost:3000 | admin / streamvault (`observability` profile) |
| Prometheus | http://localhost:9090 | `observability` profile |
| Jaeger UI | http://localhost:16686 | `observability` profile |
| Loki | http://localhost:3100 | `observability` profile |

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

### 38. WebSocket Notifications (Real-Time)

```bash
# Install wscat: npm install -g wscat
# Connect with JWT token
wscat -c "ws://localhost:8081/ws/notifications?token=$TOKEN"
```

Once connected, you'll receive real-time notifications as JSON messages:

```json
{
  "type": "notification",
  "id": "67a1b2c3d4e5f6a7b8c9d0e1",
  "category": "subscription",
  "title": "Welcome to StreamVault!",
  "body": "Your Standard plan is now active.",
  "icon": "check-circle",
  "action": "/account/subscription",
  "read": false,
  "createdAt": "2026-02-14T12:00:00Z"
}
```

Max 5 concurrent WebSocket connections per user. Heartbeat: ping every 30s, pong timeout 10s.

### 39. List Notifications

```bash
curl -s "http://localhost:8081/api/notifications?page=1&pageSize=20&unreadOnly=false" \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "items": [
    {
      "id": "67a1b2c3d4e5f6a7b8c9d0e1",
      "userId": "550e8400-...",
      "category": "subscription",
      "title": "Welcome to StreamVault!",
      "body": "Your Standard plan is now active.",
      "icon": "check-circle",
      "action": "/account/subscription",
      "read": false,
      "channels": {"inApp": "sent", "email": "sent"},
      "createdAt": "2026-02-14T12:00:00Z"
    }
  ],
  "unreadCount": 3,
  "totalCount": 15,
  "page": 1,
  "pageSize": 20,
  "totalPages": 1
}
```

Query parameters: `page` (default 1), `pageSize` (default 20, max 100), `unreadOnly` (boolean).

### 40. Mark Notification as Read

```bash
curl -s -X POST http://localhost:8081/api/notifications/NOTIFICATION_ID/read \
  -H "Authorization: Bearer $TOKEN"
```

Response: 204 No Content

### 41. Mark All Notifications as Read

```bash
curl -s -X POST http://localhost:8081/api/notifications/read-all \
  -H "Authorization: Bearer $TOKEN"
```

Response: 204 No Content

### 42. Get Notification Preferences

```bash
curl -s http://localhost:8081/api/notifications/preferences \
  -H "Authorization: Bearer $TOKEN"
```

Response (200):
```json
{
  "userId": "550e8400-...",
  "email": true,
  "push": true,
  "inApp": true,
  "updatedAt": "2026-02-14T12:00:00Z"
}
```

### 43. Update Notification Preferences

```bash
curl -s -X PUT http://localhost:8081/api/notifications/preferences \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "email": false,
    "push": true,
    "inApp": true
  }'
```

Response (200): Updated preferences object. At least one channel must be provided.

### 44. Multi-Level Health Checks

```bash
# Liveness — process alive, no dependency checks
curl -s http://localhost:8081/health/live

# Readiness — all downstream services healthy
curl -s http://localhost:8081/health/ready

# Full diagnostic — component status + latency
curl -s http://localhost:8081/health
```

Response for `/health` (200):
```json
{
  "status": "healthy",
  "service": "gateway",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "startupCompleted": true,
  "components": {
    "user-service": {"status": "healthy", "latency": "5ms"},
    "catalog-service": {"status": "healthy", "latency": "3ms"},
    "streaming-service": {"status": "healthy", "latency": "2ms"},
    "encoding-service": {"status": "healthy", "latency": "8ms"},
    "search-service": {"status": "healthy", "latency": "4ms"},
    "recommendation-service": {"status": "healthy", "latency": "3ms"},
    "subscription-service": {"status": "healthy", "latency": "6ms"},
    "notification-service": {"status": "healthy", "latency": "4ms"},
    "redis": {"status": "healthy", "latency": "1ms"},
    "consul": {"status": "healthy", "latency": "2ms"}
  }
}
```

Status values: `healthy` (all OK), `degraded` (at least 1 component unhealthy), `unhealthy` (critical components down). All 9 application services expose `/health/live`, `/health/ready`, and `/health`.

## Project Structure

```
stream-vault/
├── gateway/                    # API Gateway (Go)
│   ├── cmd/gateway/            # Entry point
│   ├── internal/
│   │   ├── config/             # Viper configuration
│   │   ├── discovery/          # Consul service resolver
│   │   ├── health/             # Health check handler
│   │   ├── middleware/         # Auth, CORS, rate limiting, logging
│   │   ├── proxy/              # Reverse proxy, routing, upstream
│   │   └── resilience/         # Circuit breaker, retry, bulkhead, degradation
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
│   ├── subscription-service/   # Subscription Service (.NET 8)
│   │   ├── src/
│   │   │   ├── SubscriptionService.Api/           # Controllers, background services
│   │   │   ├── SubscriptionService.Application/   # Commands, queries, sagas
│   │   │   ├── SubscriptionService.Domain/        # Entities, enums, events
│   │   │   └── SubscriptionService.Infrastructure/# EF Core, repositories, mock payment
│   │   └── tests/
│   └── notification-service/   # Notification Service (Go)
│       ├── cmd/notification/   # Entry point
│       ├── internal/
│       │   ├── auth/           # JWT validation (WebSocket query param)
│       │   ├── config/         # Viper configuration
│       │   ├── consumer/       # RabbitMQ consumers (subscription, encoding, content)
│       │   ├── dispatcher/     # Event → channel routing (email, push, inapp)
│       │   ├── discovery/      # Consul registration
│       │   ├── handler/        # HTTP + WebSocket handlers
│       │   ├── middleware/     # Recovery, request ID, logging
│       │   ├── store/          # MongoDB stores (notifications, preferences)
│       │   ├── template/       # HTML email templates (Go html/template)
│       │   └── websocket/      # Hub/Client architecture, heartbeat
│       └── config.yaml
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
│   ├── rabbitmq/               # RabbitMQ config + topology (6 exchanges, 13 queues)
│   ├── redis/                  # Redis configuration
│   └── minio/                  # MinIO bucket init script
├── observability/
│   ├── grafana/
│   │   ├── grafana.ini                        # Grafana server config
│   │   └── provisioning/
│   │       ├── datasources/datasources.yml    # Prometheus, Loki, Jaeger
│   │       ├── dashboards/
│   │       │   ├── dashboards.yml             # Auto-provision config
│   │       │   ├── services-overview.json     # Services Overview dashboard
│   │       │   ├── streaming-dashboard.json   # Streaming & Encoding dashboard
│   │       │   ├── business-dashboard.json    # Business Metrics dashboard
│   │       │   └── gateway-dashboard.json     # Gateway dashboard
│   │       └── alerting/                      # Alert rules (error rate, latency, service down)
│   ├── jaeger/jaeger-config.yml               # Jaeger collector config
│   ├── loki/loki-config.yml                   # Loki storage + retention config
│   ├── prometheus/prometheus.yml              # Scrape configs for all 9 services
│   └── promtail/promtail-config.yml           # Docker log collection + pipeline stages
├── scripts/                    # Utility scripts
│   ├── generate-proto.sh       # Proto code generation (Go + Rust)
│   ├── upload-test-video.sh    # Upload test video + verify encoding
│   ├── test-streaming.sh       # Streaming flow integration tests
│   ├── test-search.sh          # Search service integration tests
│   ├── test-subscription.sh    # Subscription saga integration tests
│   ├── seed-search-index.sh    # Bulk index catalog data into Elasticsearch
│   └── generate-interactions.sh # Generate fake watch/rating data for recommendations
├── docker-compose.yml          # Service definitions (21 containers)
└── docker-compose.override.yml # Development port mappings
```

## Port Reference

### Application Services (9)

| Service | Language | Container Port | Host Port | Profile |
|---------|----------|---------------|-----------|---------|
| API Gateway | Go | 8080 | 8081 | default |
| User Service | .NET 8 | 8080 | 5001 | default |
| Catalog Service | .NET 8 | 5100 | 5002 | default |
| Streaming Service (HTTP) | Go | 5003 | 5003 | default |
| Streaming Service (gRPC) | Go | 50051 | 50051 | default |
| Encoding Service (HTTP) | Rust | 5004 | 5004 | default |
| Encoding Service (gRPC) | Rust | 50052 | 50052 | default |
| Search Service (HTTP) | Go | 5005 | 5005 | elasticsearch |
| Search Service (gRPC) | Go | 50053 | 50053 | elasticsearch |
| Recommendation Service (HTTP) | Go | 5006 | 5006 | default |
| Recommendation Service (gRPC) | Go | 50054 | 50054 | default |
| Subscription Service | .NET 8 | 5007 | 5007 | default |
| Notification Service | Go | 5008 | 5008 | default |

### Data Stores (6)

| Service | Container Port | Host Port | Profile |
|---------|---------------|-----------|---------|
| PostgreSQL | 5432 | 5432 | default |
| MongoDB | 27017 | 27017 | default |
| Redis | 6379 | 6379 | default |
| Elasticsearch | 9200 | 9200 | elasticsearch |
| MinIO (API) | 9000 | 9000 | default |
| MinIO (Console) | 9001 | 9001 | default |

### Infrastructure (2)

| Service | Container Port | Host Port | Profile |
|---------|---------------|-----------|---------|
| Consul (UI) | 8500 | 8500 | default |
| RabbitMQ (AMQP) | 5672 | 5672 | default |
| RabbitMQ (Management UI) | 15672 | 15672 | default |

### Observability Stack (5)

| Service | Container Port | Host Port | Profile |
|---------|---------------|-----------|---------|
| Jaeger (UI) | 16686 | 16686 | observability |
| Jaeger (OTLP gRPC) | 4317 | 4317 | observability |
| Jaeger (OTLP HTTP) | 4318 | 4318 | observability |
| Prometheus | 9090 | 9090 | observability |
| Grafana | 3000 | 3000 | observability |
| Loki | 3100 | 3100 | observability |
| Promtail | 9080 | 9080 | observability |

**Total: 21 containers** (9 app + 6 data + 2 infra + 5 observability, excluding init containers). Resource limits: ~13.5 CPU cores, ~9.5 GB RAM when all profiles active.

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
- **Real-time notifications** via WebSocket (in-app, mock email via SMTP, mock push)
- **Notification preferences** per user (enable/disable email, push, in-app channels)
- **Notification history** with pagination, mark as read, auto-expire (90-day TTL)
- **Distributed tracing** with OpenTelemetry + Jaeger (all 9 services, async trace propagation)
- **Prometheus metrics** with RED metrics + business metrics on all services (`/metrics` endpoint)
- **Grafana dashboards** (4 pre-provisioned: Overview, Streaming, Business, Gateway + alert rules)
- **Centralized logging** with Loki + Promtail (structured JSON, LogQL queries, trace correlation)
- **Circuit breaker** per downstream service (gobreaker v2, configurable thresholds)
- **Bulkhead pattern** with per-service concurrency limits (streaming: 200, search: 100, others: 30-50)
- **Retry with exponential backoff** + jitter (max 3 retries, retryable status codes only)
- **Graceful degradation** strategies (empty results fallback when services are down)
- **Multi-level health checks** (live/ready/startup on all services, Gateway aggregation)
- **Graceful shutdown** with signal handling (SIGTERM, in-flight request completion, Consul deregistration)

## Runtime Configuration

- `JWT_SECRET` is shared by the gateway, User Service, and Notification Service and must be at least 32 bytes. `JWT_EXPIRES_IN_MINUTES` controls both the JWT `exp` value and the `expiresIn` response field.
- User, Catalog, and Subscription registration consistently use `ServiceRegistration:Name` and `ServiceRegistration:Port`; `Service:Host` supplies the advertised host. Compose derives ASP.NET listening, Consul registration, Gateway upstream, readiness-check, and container health-check ports from the matching `*_SERVICE_PORT` value.
- `ENCODING_MAX_CONCURRENT_JOBS` must be between 1 and 65535. The consumer automatically raises RabbitMQ prefetch to at least this value, so increasing concurrency is not capped by the file default. `ENCODING_MAX_RETRIES` counts retries after the initial attempt, so `0` means one attempt and the default `3` means up to four total attempts. Before this correction, the default happened to make only three total attempts.
- `CATALOG_DATABASE_NAME` defaults to the canonical `streamvault_catalog` name used by Mongo initialization and Catalog Service.
- `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_WS_URL` are public browser configuration. Compose passes them as Docker build arguments because Next.js embeds them into client bundles at build time; rebuild the web image after changing either value, and never put a secret in a `NEXT_PUBLIC_*` value.

Initialization scripts run only when their database volume is empty. Existing Mongo installations that contain the legacy `stream_vault_catalog` database can set `CATALOG_DATABASE_NAME=stream_vault_catalog` until data is migrated manually. Existing PostgreSQL installations keep their original bootstrap role: either keep `POSTGRES_USER` unchanged or have a DBA create/grant the replacement role and update application credentials. No volume deletion or automatic data migration is performed.

## RabbitMQ Topology

Topology is stored in `infrastructure/rabbitmq/definitions.json` and imported by the one-shot `rabbitmq-init` service after the broker creates the env-configured default user. The file is mounted as a Compose config, so changing its contents recreates the init container and reapplies topology on the next `docker compose up`. User records, password hashes, and permissions are intentionally absent from the repository definitions. Applications wait for the topology import to complete before starting.

On a clean volume, RabbitMQ creates `RABBITMQ_USER` with `RABBITMQ_PASS` and full default-vhost permissions. On an existing volume, RabbitMQ preserves existing users and ignores changed default-user env values. To rotate credentials, use an authorized `rabbitmqctl change_password <user> <new-password>` operation inside the broker, update `.env`, then recreate the broker, init, and application containers without deleting the volume. If the env password does not match the existing user, `rabbitmq-init` fails instead of starting applications against an uninitialized topology.

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
| `notification.subscription` | `subscription.events` | `subscription.created`, `subscription.cancelled`, `plan.changed` | Notification Service | Welcome/cancel/change notifications |
| `notification.encoding` | `encoding` | `job.completed`, `job.failed` | Notification Service | Encoding status notifications (admin) |
| `notification.content` | `catalog.events` | `content.created` | Notification Service | New content notifications |

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
                                       ├─subscription.created/cancelled/plan.changed──► user.subscription-sync ──► User Service (tier update)
                                       └─subscription.created/cancelled/plan.changed──► notification.subscription ──► Notification Service
```

**Notification Events:**
```
[encoding] ──job.completed/failed──► notification.encoding ──► Notification Service (admin only)

[catalog.events] ──content.created──► notification.content ──► Notification Service (all users)

Notification Service ──dispatcher──► Email (mock SMTP) / Push (mock log) / In-App (WebSocket)
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

## Observability Stack

The observability stack runs under the `observability` Docker Compose profile. Enable it with:

```bash
docker compose --profile observability up -d
```

### Distributed Tracing (Jaeger)

**Access:** http://localhost:16686

All 9 application services are instrumented with OpenTelemetry and export traces to Jaeger via OTLP (gRPC port 4317, HTTP port 4318).

**Finding Traces:**

1. Open Jaeger UI → select a service from the "Service" dropdown (e.g., `gateway`, `catalog-service`)
2. Optionally filter by operation, tags, or min/max duration
3. Click "Find Traces" to see matching traces with timeline visualization

**Example Trace Chains:**

| Type | Flow | What to Look For |
|------|------|------------------|
| Synchronous HTTP | Client → Gateway → Catalog Service → MongoDB | Single trace with nested spans showing full request lifecycle |
| Async Event | Streaming Service publish → Encoding Service consume | Linked span — separate trace IDs connected via `links` |
| gRPC | Streaming Service ↔ Encoding Service | `rpc.system=grpc` spans with method names |
| Database | Any service → PostgreSQL/MongoDB/Redis/ES | `db.system` attribute, query details in span attributes |

**Useful Search Tips:**

```
# Find traces with errors
Tags: error=true

# Find slow requests (> 1 second)
Min Duration: 1s

# Find specific trace by ID
Search by Trace ID field (copy from logs or Grafana)

# Find requests to a specific endpoint
Tags: http.target=/api/catalog/movies
```

**Span Attributes:** `http.method`, `http.status_code`, `http.target`, `db.system`, `db.statement`, `messaging.system`, `messaging.destination`, `rpc.system`, `rpc.method`

### Metrics (Prometheus)

**Access:** http://localhost:9090

Prometheus scrapes all 9 services at 10-15 second intervals. Check target health at http://localhost:9090/targets.

**Useful PromQL Queries:**

```promql
# Request rate per service (requests/second)
rate(http_requests_total[5m])

# Error rate percentage
sum(rate(http_request_errors_total[5m])) / sum(rate(http_requests_total[5m])) * 100

# P95 latency by service
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service))

# P99 latency for a specific service
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{service="gateway"}[5m]))

# Active encoding jobs
encoding_active_jobs

# Circuit breaker state (0=closed, 1=half-open, 2=open)
gateway_circuit_breaker_state

# Rate limit hits
rate(gateway_rate_limit_hits_total[5m])

# WebSocket active connections
notification_ws_active_connections

# Subscription tier distribution
subscription_active_total

# Search cache hit ratio
search_cache_hit_ratio
```

**Business Metrics Available:**

| Service | Key Metrics |
|---------|-------------|
| Gateway | `gateway_active_connections`, `gateway_rate_limit_hits_total`, `gateway_circuit_breaker_state`, `gateway_bulkhead_rejects_total` |
| Streaming | `streaming_active_viewers`, `streaming_concurrent_viewers{tier}`, `streaming_segment_serve_duration_seconds{quality}`, `streaming_bandwidth_bytes_total` |
| Encoding | `encoding_jobs_total{status}`, `encoding_job_duration_seconds{quality}`, `encoding_active_jobs`, `encoding_queue_depth` |
| Search | `search_queries_total{type}`, `search_query_duration_seconds`, `search_cache_hit_ratio`, `search_index_document_count` |
| Recommendation | `recommendation_requests_total{algorithm}`, `recommendation_computation_duration_seconds`, `recommendation_cache_hit_ratio`, `recommendation_cold_start_fallback_total` |
| Subscription | `subscription_active_total{tier}`, `subscription_created_total`, `subscription_cancelled_total`, `payment_processed_total{status}`, `saga_completed_total{saga_type,result}` |
| Notification | `notification_sent_total{channel,category}`, `notification_failed_total{channel,reason}`, `notification_ws_active_connections`, `notification_ws_messages_sent_total` |

### Dashboards (Grafana)

**Access:** http://localhost:3000 — Login: `admin` / `streamvault`

4 dashboards are auto-provisioned on startup. Datasources (Prometheus, Loki, Jaeger) are pre-configured.

| Dashboard | File | Panels |
|-----------|------|--------|
| **Services Overview** | `services-overview.json` | Total requests/s, error rate %, avg latency, healthy service count, request rate by service, error rate by service, P50/P95/P99 latency, service health matrix |
| **Streaming & Encoding** | `streaming-dashboard.json` | Active viewers, segment latency, active encoding jobs, queue depth, viewers by quality, segment serve duration histogram, encoding job pipeline, bandwidth usage |
| **Business Metrics** | `business-dashboard.json` | Active subscriptions, MRR (TRY), search queries/s, churn rate, subscriptions by tier (pie), daily active viewers, payment success rate (gauge), notifications by channel (stacked bar) |
| **Gateway** | `gateway-dashboard.json` | Route-level request rate + latency, rate limit hit count, circuit breaker state visualization |

**Alert Rules:**

| Alert | Condition | Severity |
|-------|-----------|----------|
| High Error Rate | Error rate > 5% for 5 minutes | Critical |
| High Latency | P99 latency > 2s for 5 minutes | Warning |
| Service Down | Service health check failing for 1 minute | Critical |

### Centralized Logging (Loki)

**Access:** Via Grafana Explore tab → select "Loki" datasource

All services output structured JSON logs. Promtail collects Docker container logs and ships them to Loki with service labels extracted from container names.

**LogQL Query Examples:**

```logql
# All logs from a specific service
{service="gateway"}

# Error logs across all services
{job="streamvault"} | json | level="error"

# Logs for a specific trace ID (cross-service correlation)
{job="streamvault"} | json | trace_id="4bf92f3577b34da6a3ce929d0e0e4736"

# Request logs with high latency
{service="streaming-service"} | json | latency > 1000

# Count errors per service over time
count_over_time({job="streamvault"} | json | level="error" [5m])

# Search for specific error messages
{service="encoding-service"} |= "ffmpeg" |= "error"

# Filter by HTTP status code
{service="gateway"} | json | status >= 500
```

**Log → Trace Correlation:**

1. In Grafana Explore, run a Loki query
2. Expand a log entry → find the `trace_id` field
3. Click the trace_id value → Grafana navigates to the Jaeger trace view
4. See the full request path across all services for that trace

**Pipeline Stages (Promtail):**

Promtail normalizes log formats across Go (zerolog), .NET (Serilog), and Rust (tracing):
- Go fields: `level`, `msg`, `trace_id`, `span_id`, `service`
- .NET fields: `@l` → `level`, `@t` → timestamp, `TraceId`, `SpanId`, `ServiceName`
- Health check logs (`/health/live`, `/health/ready`) are automatically filtered out to reduce noise

## Resilience Patterns

The API Gateway implements multiple resilience patterns to prevent cascading failures across the microservice architecture. All resilience configuration is in `gateway/config.yaml`.

### Circuit Breaker

Each downstream service has an independent circuit breaker instance (via `sony/gobreaker` v2).

```
State Transitions:

  CLOSED ──(5 consecutive failures)──► OPEN ──(30s timeout)──► HALF-OPEN
     ▲                                                            │
     │                                                            │
     └──────────(3 consecutive successes)─────────────────────────┘
                                                                  │
                                                   (any failure)──► OPEN
```

| Parameter | Value | Description |
|-----------|-------|-------------|
| `failure_threshold` | 5 | Consecutive failures to trip circuit |
| `success_threshold` | 3 | Consecutive successes in half-open to close |
| `timeout` | 30s | Time in open state before trying half-open |

When a circuit is **OPEN**, the gateway returns `503 Service Unavailable` immediately without contacting the downstream service, with a service-specific fallback response (see Graceful Degradation below).

**Metric:** `gateway_circuit_breaker_state{service}` — 0=closed, 1=half-open, 2=open

### Retry Policy

The gateway retries failed requests with exponential backoff and jitter.

| Parameter | Value |
|-----------|-------|
| `max_retries` | 3 |
| `initial_backoff` | 100ms |
| `max_backoff` | 2s |
| `multiplier` | 2.0 |
| `jitter_fraction` | 0.2 (20% random) |

**Retryable status codes:** 502, 503, 504. **Non-retryable:** 400, 401, 403, 404, 409.

Backoff sequence example: ~100ms → ~200ms → ~400ms (with random jitter to prevent thundering herd).

### Bulkhead (Concurrency Limiting)

Per-service semaphore limits prevent one service's load from exhausting gateway resources.

| Service | Max Concurrent Requests |
|---------|------------------------|
| streaming-service | 200 (video streaming is long-running) |
| search-service | 100 |
| user-service | 50 |
| catalog-service | 50 |
| encoding-service | 30 |
| recommendation-service | 30 |
| subscription-service | 30 |
| notification-service | 30 |

When a service's concurrency limit is reached, the gateway returns `503 Service Unavailable` immediately. Other services remain unaffected.

**Metric:** `gateway_bulkhead_rejects_total{service}`

### Timeout Hierarchy

```
Client ──(300s)──► Gateway ──(5s default)──► Downstream Service
                                 │
                                 ├── streaming-service: 60s (HLS segment serving)
                                 └── all others: 5s
```

Inner timeouts are always shorter than outer timeouts to ensure clean error propagation.

### Graceful Degradation

When a circuit breaker opens, the gateway returns a service-specific fallback instead of a generic error.

| Service | Fallback Behavior |
|---------|-------------------|
| search-service | Returns empty results `{"items":[], "totalCount":0}` — UI shows "no results" |
| recommendation-service | Returns empty recommendations `{"items":[]}` — UI hides section |
| notification-service | Returns message about delayed notifications — core flows continue |
| encoding-service | Returns message about queued jobs — uploads still accepted |

All fallback responses use HTTP 503 with `"status": "degraded"` in the JSON body.

## Health Check Ecosystem

All 9 application services implement a 3-level health check pattern compatible with Kubernetes probes.

### Health Check Levels

| Endpoint | Purpose | Dependencies Checked | Used By |
|----------|---------|---------------------|---------|
| `GET /health/live` | Liveness probe — process alive | None (always returns 200) | Docker Compose healthcheck |
| `GET /health/ready` | Readiness probe — can serve traffic | All critical dependencies | Consul health check |
| `GET /health` | Full diagnostic report | All dependencies + latency | Gateway aggregation, debugging |

### Service Dependencies

| Service | Dependencies Checked in `/health/ready` |
|---------|---------------------------------------|
| Gateway | Consul, Redis, all downstream services |
| User Service | PostgreSQL, RabbitMQ |
| Catalog Service | MongoDB, RabbitMQ |
| Streaming Service | MinIO, Redis, RabbitMQ |
| Encoding Service | RabbitMQ, MinIO, FFmpeg |
| Search Service | Elasticsearch, Redis |
| Recommendation Service | PostgreSQL, Redis |
| Subscription Service | PostgreSQL, RabbitMQ |
| Notification Service | MongoDB, RabbitMQ |

### Gateway Health Aggregation

`GET /health` on the gateway calls each downstream service's `/health/ready` endpoint in parallel (2s timeout per service) and aggregates results:

```json
{
  "status": "healthy | degraded | unhealthy",
  "service": "gateway",
  "version": "1.0.0",
  "uptime": "2h15m30s",
  "components": {
    "user-service": {"status": "healthy", "latency": "5ms"},
    "catalog-service": {"status": "healthy", "latency": "3ms"},
    "redis": {"status": "healthy", "latency": "1ms"}
  }
}
```

- **healthy** — all components OK
- **degraded** — at least 1 non-critical component unhealthy (service still operational)
- **unhealthy** — critical components down

## Structured Logging

All services output structured JSON logs with consistent fields for correlation and filtering.

### Log Formats by Language

**Go services** (Gateway, Streaming, Search, Recommendation, Notification) — zerolog:

```json
{
  "level": "info",
  "service": "gateway",
  "method": "GET",
  "path": "/api/catalog/movies",
  "status": 200,
  "latency": "12ms",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "correlation_id": "req-abc123",
  "time": "2026-02-14T10:30:45.123Z",
  "message": "request completed"
}
```

**.NET services** (User, Catalog, Subscription) — Serilog RenderedCompactJsonFormatter:

```json
{
  "@t": "2026-02-14T10:30:45.123Z",
  "@l": "Information",
  "@m": "User authenticated",
  "ServiceName": "user-service",
  "TraceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "SpanId": "00f067aa0ba902b7"
}
```

**Rust services** (Encoding) — tracing-subscriber JSON:

```json
{
  "timestamp": "2026-02-14T10:30:45.123Z",
  "level": "INFO",
  "message": "job processing started",
  "target": "encoding_service::pipeline",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7"
}
```

### Correlation Fields

| Field | Description | Source |
|-------|-------------|--------|
| `trace_id` | OpenTelemetry trace identifier | Extracted from span context |
| `span_id` | OpenTelemetry span identifier | Extracted from span context |
| `correlation_id` | Request correlation ID | `X-Correlation-Id` HTTP header |
| `service` | Service name | Static per service |

These fields allow filtering logs across all services for a single request flow using Loki LogQL.

## Chaos Testing

> **Status:** Planned but not yet executed. Scripts and scenarios will be added to `resilience/chaos/`.

### Planned Scenarios

**Scenario 1: Service Death (Encoding Service)**
1. Submit an encoding job (video upload)
2. Wait 5 seconds → stop encoding-service container (`docker compose stop encoding-service`)
3. Verify: Gateway `/health` reports `degraded` status
4. Verify: RabbitMQ `encoding.jobs` queue shows unprocessed messages
5. Verify: Other services (catalog, search, streaming) continue working normally
6. Restart encoding-service → verify queued jobs are processed within 30 seconds

**Scenario 2: Database Slowdown**
1. Inject 500ms network delay into PostgreSQL container
2. Send 20 requests to user-service and subscription-service
3. Verify: Response times increase but requests eventually succeed
4. Verify: Circuit breaker state changes visible in Grafana
5. Verify: Latency spike appears in Prometheus metrics
6. Remove delay → verify service recovery

**Scenario 3: RabbitMQ Outage**
1. Perform a business operation that generates outbox events (e.g., create subscription)
2. Stop RabbitMQ (`docker compose stop rabbitmq`)
3. Verify: API operations still succeed (outbox pattern buffers events)
4. Verify: Outbox table shows pending messages
5. Restart RabbitMQ → verify outbox processor delivers buffered messages

### Expected Results

| Scenario | Expected Behavior |
|----------|-------------------|
| Service death | Circuit breaker opens, fallback response served, recovery when service returns |
| Database slow | Increased latency, no data loss, circuit breaker may trip at extreme delays |
| RabbitMQ down | Synchronous APIs work, events buffered in outbox, delivery on recovery |

## Load Testing

> **Status:** Planned but not yet executed. k6 scripts will be added to `resilience/load-test/k6-scripts/`.

### Test Plan

| Test Type | VUs | Duration | Threshold |
|-----------|-----|----------|-----------|
| **Smoke** | 5 | 1 min | P95 < 500ms, error rate < 1% |
| **Load** | 50 | 5 min | P95 < 1000ms, error rate < 2% |
| **Stress** | 50 → 100 → 200 | 10 min (staged) | P95 < 2000ms, error rate < 5% |
| **Spike** | 10 → 500 | 5 min | Graceful degradation, no crashes |

### User Flow Scenario

Each virtual user executes this flow:

1. Register / Login → obtain JWT token
2. Browse catalog (list movies, genres)
3. Search for content (full-text + autocomplete)
4. Get recommendations (personalized + similar)
5. Start streaming (manifest + segments)
6. Save watch progress

### Running Tests

```bash
# Install k6
# brew install k6  (macOS)
# sudo snap install k6  (Linux)

# Run smoke test
k6 run resilience/load-test/k6-scripts/smoke.js

# Run with Prometheus output (for Grafana visualization)
k6 run --out experimental-prometheus-rw resilience/load-test/k6-scripts/load.js
```

## Troubleshooting

### Service Discovery & Health

| Problem | Solution |
|---------|----------|
| Service not appearing in Consul | Check `CONSUL_ADDRESS` env var, verify `/health/ready` returns 200, check container logs |
| Service shows "critical" in Consul | Verify the service's health endpoint, check dependency connectivity |
| All services unhealthy after restart | Wait for startup period (up to 30s), check infrastructure containers first (postgres, mongo, redis) |

### Gateway & Routing

| Problem | Solution |
|---------|----------|
| 503 from Gateway | Circuit breaker may be open — check `gateway_circuit_breaker_state` metric or Gateway logs. Wait 30s for half-open retry |
| 429 Too Many Requests | Rate limit exceeded — default 100 req/min per IP. Wait or check `gateway_rate_limit_hits_total` |
| 401 Unauthorized | Token expired — refresh with `POST /api/auth/refresh`. Check JWT secret matches between User Service and Gateway |
| WebSocket connection refused | Verify JWT token is valid, check max connections per user (limit: 5), ensure notification-service is running |

### Observability

| Problem | Solution |
|---------|----------|
| No traces in Jaeger | Verify `OTEL_EXPORTER_OTLP_ENDPOINT` env var, check jaeger container is running, ensure observability profile is active |
| No logs in Loki | Check Promtail container, verify docker socket mount (`/var/run/docker.sock`), check Promtail logs for errors |
| Grafana dashboards empty | Verify Prometheus targets are UP (`http://localhost:9090/targets`), check that services expose `/metrics`, wait for scrape interval (15s) |
| Grafana "No data" on Loki panels | Verify Loki datasource URL (`http://loki:3100`) in Grafana, check Loki ready status (`http://localhost:3100/ready`) |

### Data Stores

| Problem | Solution |
|---------|----------|
| MongoDB connection failed | Check `MONGODB_URI` env var, verify mongo container is healthy (`docker compose ps mongo`) |
| PostgreSQL connection refused | Check if postgres container is ready, verify database exists (`streamvault_users`, `streamvault_subscriptions`, `streamvault_recommendations`) |
| Redis connection timeout | Verify redis container is running, check port 6379 is not blocked |
| Elasticsearch cluster red | Check ES container logs, verify sufficient memory (`ES_JAVA_OPTS=-Xms256m -Xmx256m`), check disk space |

### Encoding & Streaming

| Problem | Solution |
|---------|----------|
| Encoding job stuck at "processing" | Check FFmpeg availability in container, verify MinIO connectivity, check encoding-service logs for FFmpeg errors |
| Upload returns 0 bytes | Ensure multipart form field order doesn't matter (service buffers file to temp), check file size limits |
| HLS manifest 404 | Verify content has been encoded (check job status), ensure MinIO `streamvault-encoded` bucket has the content directory |
| Streaming quality restricted | Check user's subscription tier — Free has no streaming, Basic=720p max. Upgrade via `PUT /api/subscriptions/me/plan` |

### RabbitMQ & Events

| Problem | Solution |
|---------|----------|
| Queues not created | Check `docker compose ps rabbitmq-init` and its logs, then verify the definitions in the Management UI. An existing volume whose password differs from `.env` requires an explicit credential rotation; do not delete the volume. |
| Events not consumed | Check consumer service logs for connection errors, verify queue bindings in Management UI |
| Outbox messages pending | Check RabbitMQ connectivity, verify outbox processor background service is running (check service logs for "outbox" entries) |

### Port Conflicts

If you see `port is already allocated` errors, check which process is using the port:

```bash
# Find process using a port
lsof -i :5001

# Or use netstat
netstat -tlnp | grep 5001
```

Refer to the [Port Reference](#port-reference) table for the complete port map. Modify `docker-compose.override.yml` to change host port mappings if needed.
