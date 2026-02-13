# ============================================================
# Stage 1: Build the streaming service binary
# ============================================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Cache dependency downloads before copying the full source.
COPY services/streaming-service/go.mod services/streaming-service/go.sum ./
RUN go mod download

COPY services/streaming-service/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/streaming ./cmd/streaming

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S streaming && adduser -S streaming -G streaming

WORKDIR /app

COPY --from=builder /build/streaming .
COPY --from=builder /build/config.yaml* ./

RUN chown -R streaming:streaming /app

USER streaming

EXPOSE 5003 50051

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:5003/health"]

ENTRYPOINT ["./streaming"]
