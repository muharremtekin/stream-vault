# ============================================================
# Stage 1: Build the gateway binary
# ============================================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Cache dependency downloads before copying the full source.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/gateway ./cmd/gateway

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S gateway && adduser -S gateway -G gateway

WORKDIR /app

COPY --from=builder /build/gateway .
COPY --from=builder /build/config.yaml* ./

RUN chown -R gateway:gateway /app

USER gateway

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health/live"]

ENTRYPOINT ["./gateway"]
