# StreamVault

A microservices-based streaming platform backend built with .NET 8, Go, and Docker.

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
                          └──┬──────────┬───┘
                             │          │
              ┌──────────────▼──┐  ┌────▼──────────────┐
              │  User Service   │  │  Catalog Service   │
              │  (.NET 8)       │  │  (.NET 8)          │
              │  :5001 (host)   │  │  :5002 (host)      │
              └──────┬──────────┘  └──────┬─────────────┘
                     │                    │
              ┌──────▼──────┐      ┌──────▼──────┐
              │ PostgreSQL  │      │   MongoDB   │
              │   :5432     │      │   :27017    │
              └─────────────┘      └─────────────┘

        ┌─────────────┐      ┌─────────────┐
        │    Consul    │      │    Redis     │
        │ (Discovery)  │      │(Rate Limit)  │
        │   :8500      │      │   :6379      │
        └─────────────┘      └─────────────┘
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| API Gateway | Go 1.22, chi router, jwt/v5 |
| User Service | .NET 8, EF Core, PostgreSQL |
| Catalog Service | .NET 8, MongoDB Driver |
| Service Discovery | HashiCorp Consul |
| Rate Limiting | Redis + Token Bucket |
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
| Consul UI | http://localhost:8500 |

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

## Project Structure

```
stream-vault/
├── gateway/                    # API Gateway (Go)
│   ├── cmd/gateway/            # Entry point
│   ├── internal/
│   │   ├── middleware/          # Auth, CORS, rate limiting, logging
│   │   └── proxy/              # Reverse proxy, routing, upstream
│   └── config.yaml             # Gateway configuration
├── services/
│   ├── user-service/           # User Service (.NET 8)
│   │   ├── src/
│   │   │   ├── UserService.Api/           # Controllers, middleware
│   │   │   ├── UserService.Application/   # CQRS commands/queries
│   │   │   ├── UserService.Domain/        # Entities, value objects
│   │   │   └── UserService.Infrastructure/# EF Core, repositories
│   │   └── tests/
│   └── catalog-service/        # Catalog Service (.NET 8)
│       ├── src/
│       │   ├── CatalogService.Api/
│       │   ├── CatalogService.Application/
│       │   ├── CatalogService.Domain/
│       │   └── CatalogService.Infrastructure/
│       └── tests/
├── infrastructure/
│   └── consul/                 # Consul configuration
├── docker-compose.yml          # Service definitions
└── docker-compose.override.yml # Development port mappings
```

## Port Reference

| Service | Container Port | Host Port |
|---------|---------------|-----------|
| API Gateway | 8080 | 8081 |
| User Service | 8080 | 5001 |
| Catalog Service | 5100 | 5002 |
| PostgreSQL | 5432 | 5432 |
| MongoDB | 27017 | 27017 |
| Redis | 6379 | 6379 |
| Consul | 8500 | 8500 |

## Key Features

- **JWT Authentication** with access + refresh token flow
- **Role-based access control** (User / Admin)
- **Multi-profile support** (up to 5 profiles per user)
- **Watchlist management** per profile
- **Movie & Series catalog** with genres, filtering, pagination, and sorting
- **Service discovery** via Consul with health checks
- **Rate limiting** (IP-based token bucket via Redis)
- **Correlation ID** propagation across services
- **Structured logging** (Serilog for .NET, zerolog for Go)
- **Consistent error responses** across all services
