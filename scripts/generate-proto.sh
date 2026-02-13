#!/bin/bash
# StreamVault - Proto Code Generation Script
# Generates Go code from proto definitions.
# Rust (tonic-build) and C# (Grpc.Tools) handle their own generation at build time.

set -e

PROTO_DIR="$(cd "$(dirname "$0")/.." && pwd)/proto"
GATEWAY_DIR="$(cd "$(dirname "$0")/.." && pwd)/gateway"
GO_OUT_DIR="$GATEWAY_DIR/proto"
ENCODING_SERVICE_DIR="$(cd "$(dirname "$0")/.." && pwd)/services/encoding-service"

echo "=== StreamVault Proto Generator ==="
echo "Proto source: $PROTO_DIR"
echo "Go output:    $GO_OUT_DIR"

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

echo ""
echo "=== Proto generation completed ==="
echo ""
echo "Generated Go packages:"
find "$GO_OUT_DIR" -name "*.pb.go" | sort | while read f; do
  echo "  ${f#$GATEWAY_DIR/}"
done
