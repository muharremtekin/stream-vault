# ============================================================
# Stage 1: Build the search service binary
# ============================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Cache dependency downloads before copying the full source.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/search ./cmd/search

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S search && adduser -S search -G search

WORKDIR /app

COPY --from=builder /build/search .
COPY --from=builder /build/config.yaml* ./

RUN chown -R search:search /app

USER search

EXPOSE 5005 50053

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:5005/health/live"]

ENTRYPOINT ["./search"]
