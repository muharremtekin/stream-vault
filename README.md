# StreamVault

A microservices-based streaming platform backend built with .NET 8, Go, Rust, and Docker.

## Architecture

```
                              ┌─────────────────┐
                              │     Client      │
                              └────────┬────────┘
                                       │
                              ┌────────▼────────┐
                              │   API Gateway   │
                              │    (Go 1.22)    │
                              │   :8081 (host)  │
                              └──┬────┬────┬──┬─┘
                                 │    │    │  │
          ┌──────────────────────┘    │    │  └──────────────────────┐
          │          ┌────────────────┘    └───────────────┐         │
  ┌───────▼───────┐  │  ┌────────────────┐  ┌─────────────▼───────┐ │
  │ User Service  │  │  │Catalog Service │  │ Streaming Service   │ │
  │  (.NET 8)     │  │  │  (.NET 8)      │  │    (Go 1.22)        │ │
  │ :5001 (host)  │  │  │ :5002 (host)   │  │ HTTP :5003          │ │
  └──────┬────────┘  │  └──────┬─────────┘  │ gRPC :50051         │ │
         │           │         │            └───┬────────┬────────┘ │
  ┌──────▼──────┐    │  ┌──────▼──────┐        │        │          │
  │ PostgreSQL  │    │  │   MongoDB   │        │        │          │
  │   :5432     │    │  │   :27017    │        │        │          │
  └─────────────┘    │  └─────────────┘        │        │          │
                     │                         │        │          │
                     │  ┌──────────────────────┘  ┌─────▼────────┐ │
                     │  │      ┌──────────────┐   │  Encoding    │ │
                     │  │      │   RabbitMQ   │◄──┤  Service     │◄┘
                     │  │      │ AMQP :5672   │   │  (Rust)      │
                     │  └─────►│ UI :15672    │   │ HTTP :5004   │
                     │         └──────────────┘   │ gRPC :50052  │
                     │                            └──────┬───────┘
                     │                                   │
                     │  ┌─────────────┐           ┌──────▼───────┐
                     │  │    Redis    │           │    MinIO     │
                     │  │ Rate Limit  │           │ Obj. Storage │
                     │  │ + Progress  │           │ API :9000    │
                     │  │   :6379     │           │ Console:9001 │
                     │  └─────────────┘           └──────────────┘
                     │
              ┌──────▼──────┐
              │   Consul    │
              │ (Discovery) │
              │   :8500     │
              └─────────────┘
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| API Gateway | Go 1.22, net/http, jwt/v5 |
| User Service | .NET 8, EF Core, PostgreSQL |
| Catalog Service | .NET 8, MongoDB Driver |
| Streaming Service | Go 1.22, HLS, gRPC |
| Encoding Service | Rust, FFmpeg, tonic (gRPC) |
| Service Discovery | HashiCorp Consul |
| Message Broker | RabbitMQ 3.13 (topic exchange) |
| Object Storage | MinIO (S3-compatible) |
| Caching & State | Redis (Rate Limiting + Watch Progress) |
| Architecture Pattern | Clean Architecture + CQRS (MediatR) |
| Containerization | Docker Compose |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)

## Quick Start

```bash
# Clone the repository
git clone <repo-url>
cd stream-vault

# Start all services
docker compose up --build -d

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
| Consul UI | http://localhost:8500 |
| RabbitMQ Management | http://localhost:15672 |
| MinIO Console | http://localhost:9001 |

## API Documentation (Swagger)

Both .NET services expose Swagger UI in development mode:

- **User Service**: http://localhost:5001/swagger
- **Catalog Service**: http://localhost:5002/swagger

OpenAPI JSON specs:
- http://localhost:5001/swagger/v1/swagger.json
- http://localhost:5002/swagger/v1/swagger.json

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
│   │   │   └── UserService.Infrastructure/# EF Core, repositories
│   │   └── tests/
│   ├── catalog-service/        # Catalog Service (.NET 8)
│   │   ├── src/
│   │   │   ├── CatalogService.Api/
│   │   │   ├── CatalogService.Application/
│   │   │   ├── CatalogService.Domain/
│   │   │   └── CatalogService.Infrastructure/
│   │   └── tests/
│   ├── streaming-service/      # Streaming Service (Go)
│   │   ├── cmd/streaming/      # Entry point
│   │   ├── internal/
│   │   │   ├── handler/        # HTTP handlers (upload, manifest, progress)
│   │   │   ├── hls/            # HLS playlist generation + quality profiles
│   │   │   ├── messaging/      # RabbitMQ publisher + consumer
│   │   │   ├── middleware/     # Admin, subscription, concurrent streams
│   │   │   ├── progress/       # Redis-backed watch progress
│   │   │   └── storage/        # MinIO storage interface
│   │   ├── proto/              # Generated gRPC stubs
│   │   └── config.yaml
│   └── encoding-service/       # Encoding Service (Rust)
│       ├── src/
│       │   ├── api/            # Axum HTTP handlers + routes
│       │   ├── domain/         # Job, Status, Profile models
│       │   ├── grpc.rs         # Tonic gRPC server
│       │   ├── messaging/      # RabbitMQ consumer + publisher
│       │   ├── pipeline/       # FFmpeg transcoder, thumbnails, orchestrator
│       │   └── storage/        # MinIO client
│       ├── proto/              # Proto source files for tonic-build
│       └── config.toml
├── proto/                      # Shared proto definitions
│   ├── common/v1/              # Pagination, ContentType, SubscriptionTier
│   ├── catalog/v1/             # CatalogService RPCs
│   ├── encoding/v1/            # EncodingService RPCs
│   ├── streaming/v1/           # StreamingService RPCs
│   └── user/v1/                # UserService RPCs
├── infrastructure/
│   ├── consul/                 # Consul configuration
│   ├── mongo/                  # MongoDB init scripts
│   ├── rabbitmq/               # RabbitMQ config + topology definitions
│   └── minio/                  # MinIO bucket init script
├── scripts/                    # Utility scripts
│   ├── generate-proto.sh       # Proto code generation (Go + Rust)
│   ├── upload-test-video.sh    # Upload test video + verify encoding
│   └── test-streaming.sh       # Streaming flow integration tests
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
| PostgreSQL | 5432 | 5432 |
| MongoDB | 27017 | 27017 |
| Redis | 6379 | 6379 |
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

## RabbitMQ Topology

The encoding pipeline uses a topic exchange with dead-letter support. All topology is pre-configured via `infrastructure/rabbitmq/definitions.json`.

### Exchanges

| Exchange | Type | Durable | Purpose |
|----------|------|---------|---------|
| `encoding` | topic | Yes | Main encoding event routing |
| `encoding.dlx` | direct | Yes | Dead-letter exchange for failed messages |

### Queues & Bindings

| Queue | Routing Key(s) | Consumer | Purpose |
|-------|----------------|----------|---------|
| `encoding.jobs` | `job.new` | Encoding Service | New encoding job requests |
| `encoding.results.catalog` | `job.completed`, `job.failed` | Catalog Service | Updates video status in catalog |
| `encoding.results.streaming` | `job.completed`, `job.failed` | Streaming Service | Updates stream info cache |
| `encoding.dead-letters` | `dead-letter` (via DLX) | — | Failed messages for inspection |

### Message Flow

```
  Upload Video                   Encoding Complete/Failed
       │                                │
       ▼                                ▼
┌──────────────┐  job.new   ┌───────────────────┐  job.completed  ┌─────────────────────┐
│  Streaming   │──────────► │                   │────────────────►│encoding.results     │
│  Service     │            │ encoding exchange │                 │  .catalog (Catalog) │
│  (publisher) │            │    (topic)        │────────────────►│  .streaming (Stream)│
└──────────────┘            └────────┬──────────┘  job.failed     └─────────────────────┘
                                     │ job.new
                                     ▼
                            ┌─────────────────┐         ┌─────────────────┐
                            │ encoding.jobs   │────────►│ Encoding Service│
                            │    (queue)      │         │   (consumer)    │
                            └─────────────────┘         └─────────────────┘
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

| Tier | Max Streaming Quality | Concurrent Streams |
|------|----------------------|-------------------|
| Free | — (streaming not available) | 1 |
| Basic | 360p + 720p | 1 |
| Standard | 360p + 720p + 1080p | 2 |
| Premium | All qualities (incl. 4K) | 4 |
| Admin | All qualities (incl. 4K) | 4 |
