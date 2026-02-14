#!/bin/bash
# StreamVault - Proto Code Generation Script
# Generates Go code from proto definitions.
# Rust (tonic-build) and C# (Grpc.Tools) handle their own generation at build time.

set -e

PROTO_DIR="$(cd "$(dirname "$0")/.." && pwd)/proto"
GATEWAY_DIR="$(cd "$(dirname "$0")/.." && pwd)/gateway"
GO_OUT_DIR="$GATEWAY_DIR/proto"
ENCODING_SERVICE_DIR="$(cd "$(dirname "$0")/.." && pwd)/services/encoding-service"
STREAMING_SERVICE_DIR="$(cd "$(dirname "$0")/.." && pwd)/services/streaming-service"
STREAMING_GO_OUT_DIR="$STREAMING_SERVICE_DIR/proto"
SEARCH_SERVICE_DIR="$(cd "$(dirname "$0")/.." && pwd)/services/search-service"
SEARCH_GO_OUT_DIR="$SEARCH_SERVICE_DIR/proto"
RECOMMENDATION_SERVICE_DIR="$(cd "$(dirname "$0")/.." && pwd)/services/recommendation-service"

echo "=== StreamVault Proto Generator ==="
echo "Proto source:       $PROTO_DIR"
echo "Go output (gw):     $GO_OUT_DIR"
echo "Go output (stream): $STREAMING_GO_OUT_DIR"
echo "Go output (search): $SEARCH_GO_OUT_DIR"

# Check dependencies
command -v protoc >/dev/null 2>&1 || { echo "ERROR: protoc is not installed"; exit 1; }
command -v protoc-gen-go >/dev/null 2>&1 || { echo "ERROR: protoc-gen-go is not installed. Run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"; exit 1; }
command -v protoc-gen-go-grpc >/dev/null 2>&1 || { echo "ERROR: protoc-gen-go-grpc is not installed. Run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"; exit 1; }

# Clean previous generated files
echo "[0/3] Cleaning previous generated Go files..."
rm -rf "$GO_OUT_DIR"
mkdir -p "$GO_OUT_DIR"

# Generate Go code for gateway
echo "[1/3] Generating Go code..."
for proto_file in $(find "$PROTO_DIR" -name "*.proto" | sort); do
  relative_path="${proto_file#$PROTO_DIR/}"
  echo "  Processing: $relative_path"
  protoc \
    --proto_path="$PROTO_DIR" \
    --go_out="$GO_OUT_DIR" \
    --go-grpc_out="$GO_OUT_DIR" \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    "$proto_file"
done

# Generate Go code for streaming service (only common + streaming protos)
echo "[1.5/3] Generating Go code for streaming service..."
rm -rf "$STREAMING_GO_OUT_DIR"
mkdir -p "$STREAMING_GO_OUT_DIR"
for proto_file in common/v1/common.proto streaming/v1/streaming.proto; do
  echo "  Processing (streaming-service): $proto_file"
  protoc \
    --proto_path="$PROTO_DIR" \
    --go_out="$STREAMING_GO_OUT_DIR" \
    --go-grpc_out="$STREAMING_GO_OUT_DIR" \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    --go_opt=Mcommon/v1/common.proto=github.com/streamvault/streaming-service/proto/common/v1 \
    --go_opt=Mstreaming/v1/streaming.proto=github.com/streamvault/streaming-service/proto/streaming/v1 \
    --go-grpc_opt=Mcommon/v1/common.proto=github.com/streamvault/streaming-service/proto/common/v1 \
    --go-grpc_opt=Mstreaming/v1/streaming.proto=github.com/streamvault/streaming-service/proto/streaming/v1 \
    "$PROTO_DIR/$proto_file"
done

# Generate Go code for search service (only common + search protos)
echo "[1.6/3] Generating Go code for search service..."
if [ -d "$SEARCH_SERVICE_DIR" ]; then
  rm -rf "$SEARCH_GO_OUT_DIR"
  mkdir -p "$SEARCH_GO_OUT_DIR"
  for proto_file in common/v1/common.proto search/v1/search.proto; do
    echo "  Processing (search-service): $proto_file"
    protoc \
      --proto_path="$PROTO_DIR" \
      --go_out="$SEARCH_GO_OUT_DIR" \
      --go-grpc_out="$SEARCH_GO_OUT_DIR" \
      --go_opt=paths=source_relative \
      --go-grpc_opt=paths=source_relative \
      --go_opt=Mcommon/v1/common.proto=github.com/streamvault/search-service/proto/common/v1 \
      --go_opt=Msearch/v1/search.proto=github.com/streamvault/search-service/proto/search/v1 \
      --go-grpc_opt=Mcommon/v1/common.proto=github.com/streamvault/search-service/proto/common/v1 \
      --go-grpc_opt=Msearch/v1/search.proto=github.com/streamvault/search-service/proto/search/v1 \
      "$PROTO_DIR/$proto_file"
  done
else
  echo "  Search service not found at $SEARCH_SERVICE_DIR. Skipping."
fi

# C# code generation note
echo "[2/3] C# code generation..."
echo "  C# code generation is handled by Grpc.Tools NuGet package at build time."
echo "  Ensure proto files are referenced in .csproj files with <Protobuf> items."

# Rust code generation - copy proto files for tonic-build
echo "[3/3] Rust code generation (tonic-build)..."
if [ -d "$ENCODING_SERVICE_DIR" ]; then
  RUST_PROTO_DIR="$ENCODING_SERVICE_DIR/proto"
  echo "  Copying proto files to $RUST_PROTO_DIR for tonic-build..."
  rm -rf "$RUST_PROTO_DIR"
  cp -r "$PROTO_DIR" "$RUST_PROTO_DIR"
  echo "  Proto files copied. Rust code is generated at build time via build.rs."
else
  echo "  Encoding service not found at $ENCODING_SERVICE_DIR."
  echo "  Skipping proto copy. When the service exists, tonic-build in build.rs"
  echo "  will generate Rust code from proto files at compile time."
fi

# Copy proto files for recommendation service (Rust / tonic-build)
if [ -d "$RECOMMENDATION_SERVICE_DIR" ]; then
  RUST_REC_PROTO_DIR="$RECOMMENDATION_SERVICE_DIR/proto"
  echo "  Copying proto files to $RUST_REC_PROTO_DIR for tonic-build..."
  rm -rf "$RUST_REC_PROTO_DIR"
  cp -r "$PROTO_DIR" "$RUST_REC_PROTO_DIR"
  echo "  Proto files copied for recommendation service."
else
  echo "  Recommendation service not found at $RECOMMENDATION_SERVICE_DIR. Skipping."
fi

echo ""
echo "=== Proto generation completed ==="
echo ""
echo "Generated Go packages:"
find "$GO_OUT_DIR" -name "*.pb.go" | sort | while read f; do
  echo "  ${f#$GATEWAY_DIR/}"
done
