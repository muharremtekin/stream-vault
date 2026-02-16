# syntax=docker/dockerfile:1
# ============================================================
# Stage 1: Build the Rust binary
# ============================================================
FROM rust:1.93-alpine3.21 AS builder

RUN apk add --no-cache musl-dev protobuf-dev protoc openssl-dev openssl-libs-static

WORKDIR /build

# Copy manifests and build script first for dependency caching
COPY Cargo.toml Cargo.lock* build.rs ./

# Create stub source and proto so cargo can resolve and build dependencies
RUN mkdir -p src proto/encoding/v1 && \
    echo "fn main() {}" > src/main.rs && \
    echo 'syntax = "proto3"; package encoding.v1;' > proto/encoding/v1/encoding.proto

# Build dependencies only (cached unless Cargo.toml/Cargo.lock change)
RUN --mount=type=cache,target=/build/target \
    --mount=type=cache,target=/usr/local/cargo/registry \
    cargo build --release && \
    rm -rf src proto

# Copy real source and proto files
COPY . .

# Build the actual binary (only recompiles our crate, deps are cached)
# touch source + build.rs to invalidate cargo fingerprints from the stub build
RUN --mount=type=cache,target=/build/target \
    --mount=type=cache,target=/usr/local/cargo/registry \
    touch src/main.rs build.rs && \
    cargo build --release && \
    cp target/release/encoding-service /build/encoding-service-bin

# ============================================================
# Stage 2: Minimal runtime with FFmpeg
# ============================================================
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata ffmpeg

RUN addgroup -S encoding && adduser -S encoding -G encoding

WORKDIR /app

COPY --from=builder /build/encoding-service-bin ./encoding-service
COPY --from=builder /build/config.toml* ./

RUN mkdir -p /tmp/encoding && \
    chown -R encoding:encoding /app /tmp/encoding

USER encoding

EXPOSE 5004 50052

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:5004/health/live"]

ENTRYPOINT ["./encoding-service"]
