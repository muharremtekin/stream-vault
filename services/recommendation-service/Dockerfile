# Stage 1: Build the recommendation service binary
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/recommendation ./cmd/recommendation

# Stage 2: Minimal runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S recommendation && adduser -S recommendation -G recommendation

WORKDIR /app

COPY --from=builder /build/recommendation .
COPY --from=builder /build/config.yaml* ./

RUN chown -R recommendation:recommendation /app

USER recommendation

EXPOSE 5006 50054

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:5006/health/live"]

ENTRYPOINT ["./recommendation"]
