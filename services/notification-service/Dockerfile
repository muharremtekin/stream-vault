FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/notification ./cmd/notification

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S notification && adduser -S notification -G notification

WORKDIR /app

COPY --from=builder /build/notification .
COPY --from=builder /build/config.yaml ./

RUN chown -R notification:notification /app

USER notification

EXPOSE 5008 50055

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:5008/health/live"]

ENTRYPOINT ["./notification"]
