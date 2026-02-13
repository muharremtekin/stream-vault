#!/bin/sh
set -e

echo "Waiting for MinIO to be ready..."
until mc alias set sv http://minio:9000 "${MINIO_ROOT_USER:-minioadmin}" "${MINIO_ROOT_PASSWORD:-minioadmin}" 2>/dev/null; do
  echo "MinIO not ready yet, retrying in 2 seconds..."
  sleep 2
done

echo "MinIO is ready. Creating buckets..."

mc mb sv/streamvault-raw --ignore-existing
mc mb sv/streamvault-encoded --ignore-existing
mc mb sv/streamvault-thumbnails --ignore-existing

# Allow anonymous download for thumbnails
mc anonymous set download sv/streamvault-thumbnails

echo "Buckets created successfully:"
mc ls sv/
