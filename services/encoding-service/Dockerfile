# ============================================================
# Stage 1: Build the Rust binary
# ============================================================
FROM rust:1.85-alpine AS builder

RUN apk add --no-cache musl-dev protobuf-dev protoc

WORKDIR /build

# Cache dependency downloads
COPY Cargo.toml Cargo.lock* ./
COPY build.rs ./
RUN mkdir -p src proto/encoding/v1 && \
    echo "fn main() {}" > src/main.rs && \
    touch proto/encoding/v1/encoding.proto && \
    cargo build --release 2>/dev/null || true && \
    rm -rf src proto

# Copy source and proto files
COPY . .

RUN cargo build --release

# ============================================================
# Stage 2: Minimal runtime with FFmpeg
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata ffmpeg

RUN addgroup -S encoding && adduser -S encoding -G encoding

WORKDIR /app

COPY --from=builder /build/target/release/encoding-service .
COPY --from=builder /build/config.toml* ./

RUN mkdir -p /tmp/encoding && \
    chown -R encoding:encoding /app /tmp/encoding

USER encoding

EXPOSE 5004 50052

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:5004/health"]

ENTRYPOINT ["./encoding-service"]
