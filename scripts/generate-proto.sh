#!/bin/bash
# StreamVault - Proto Code Generation Script
# Generates Go and C# code from proto definitions

set -e

PROTO_DIR="$(cd "$(dirname "$0")/.." && pwd)/proto"
GATEWAY_DIR="$(cd "$(dirname "$0")/.." && pwd)/gateway"

echo "=== StreamVault Proto Generator ==="

# Check dependencies
command -v protoc >/dev/null 2>&1 || { echo "ERROR: protoc is not installed"; exit 1; }

# Generate Go code for gateway
echo "[1/2] Generating Go code..."
for proto_file in $(find "$PROTO_DIR" -name "*.proto"); do
  echo "  Processing: $proto_file"
  protoc \
    --proto_path="$PROTO_DIR" \
    --go_out="$GATEWAY_DIR" \
    --go-grpc_out="$GATEWAY_DIR" \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    "$proto_file"
done

# Generate C# code for .NET services
echo "[2/2] Generating C# code..."
echo "  C# code generation is handled by Grpc.Tools NuGet package at build time."
echo "  Ensure proto files are referenced in .csproj files."

echo ""
echo "=== Proto generation completed ==="
