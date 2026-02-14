#!/bin/bash
# StreamVault - Upload & Encoding Pipeline Test
# Tests: 7.1 (Upload→Encode→Serve), 7.3 (Error scenarios), 7.4 (Consul & infra)
#
# Prerequisites:
#   - All services running via docker compose (user-service seeds admin on startup)
#   - Catalog seeded with at least one movie (seed-catalog.sh)

set -euo pipefail

API_URL="${API_URL:-http://localhost:8081}"
MINIO_URL="${MINIO_URL:-http://localhost:9000}"
CONSUL_URL="${CONSUL_URL:-http://localhost:8500}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@streamvault.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-SeedPass123@}"
VIDEO_PATH="${VIDEO_PATH:-docs/test-videos/video-5.mp4}"
POLL_INTERVAL="${POLL_INTERVAL:-10}"
POLL_TIMEOUT="${POLL_TIMEOUT:-600}"

# MinIO credentials (read from .env if not set)
if [ -z "${MINIO_ACCESS_KEY:-}" ] || [ -z "${MINIO_SECRET_KEY:-}" ]; then
  if [ -f .env ]; then
    MINIO_ACCESS_KEY=$(grep -E '^MINIO_ACCESS_KEY=' .env | cut -d= -f2)
    MINIO_SECRET_KEY=$(grep -E '^MINIO_SECRET_KEY=' .env | cut -d= -f2)
  fi
fi
MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY:-minioadmin}"
MINIO_SECRET_KEY="${MINIO_SECRET_KEY:-minioadmin}"

PASS=0
FAIL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

pass() {
  echo -e "  ${GREEN}PASS${NC} $1"
  PASS=$((PASS + 1))
}

fail() {
  echo -e "  ${RED}FAIL${NC} $1"
  FAIL=$((FAIL + 1))
}

info() {
  echo -e "  ${YELLOW}INFO${NC} $1"
}

section() {
  echo ""
  echo -e "${CYAN}[$1]${NC}"
}

# ── Prerequisites ──────────────────────────────────────────
echo "=== StreamVault Upload & Encoding Pipeline Test ==="
echo "    Gateway:  $API_URL"
echo "    MinIO:    $MINIO_URL"
echo "    Consul:   $CONSUL_URL"
echo ""

for cmd in curl jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: '$cmd' is required but not installed."
    exit 1
  fi
done

if [ ! -f "$VIDEO_PATH" ]; then
  echo "ERROR: Test video not found at '$VIDEO_PATH'"
  echo "       Run from repo root or set VIDEO_PATH env var."
  exit 1
fi

# ── 7.4 Consul Health Checks ──────────────────────────────
section "7.4 — Consul Service Health"

CONSUL_SERVICES=("user-service" "catalog-service" "streaming-service" "encoding-service")
for svc in "${CONSUL_SERVICES[@]}"; do
  HEALTH=$(curl -s "${CONSUL_URL}/v1/health/service/${svc}?passing" 2>/dev/null)
  COUNT=$(echo "$HEALTH" | jq 'length' 2>/dev/null || echo "0")
  if [ "$COUNT" -gt 0 ]; then
    pass "Consul: ${svc} is healthy (${COUNT} instance(s))"
  else
    fail "Consul: ${svc} NOT healthy or not registered"
  fi
done

# ── Admin Login ────────────────────────────────────────────
section "7.1 — Upload → Encode → Serve Pipeline"

info "Logging in as admin (${ADMIN_EMAIL})..."
LOGIN_RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${ADMIN_EMAIL}\", \"password\": \"${ADMIN_PASSWORD}\"}")

LOGIN_STATUS=$(echo "$LOGIN_RESP" | tail -1)
LOGIN_BODY=$(echo "$LOGIN_RESP" | sed '$d')

if [ "$LOGIN_STATUS" != "200" ]; then
  fail "Admin login -> HTTP ${LOGIN_STATUS} (expected 200)"
  echo "       Make sure admin user exists. Response: $(echo "$LOGIN_BODY" | head -c 200)"
  echo ""
  echo "================================"
  echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
  echo "================================"
  exit 1
fi

TOKEN=$(echo "$LOGIN_BODY" | jq -r '.accessToken')
pass "Admin login -> 200"

# ── Get Catalog Movie ID ──────────────────────────────────
info "Fetching catalog movies..."
MOVIES_RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/catalog/movies")
MOVIES_STATUS=$(echo "$MOVIES_RESP" | tail -1)
MOVIES_BODY=$(echo "$MOVIES_RESP" | sed '$d')

if [ "$MOVIES_STATUS" != "200" ]; then
  fail "Fetch movies -> HTTP ${MOVIES_STATUS} (expected 200)"
  echo ""
  echo "================================"
  echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
  echo "================================"
  exit 1
fi

CONTENT_ID=$(echo "$MOVIES_BODY" | jq -r '.items[0].id')
MOVIE_TITLE=$(echo "$MOVIES_BODY" | jq -r '.items[0].title')

if [ -z "$CONTENT_ID" ] || [ "$CONTENT_ID" = "null" ]; then
  fail "No movies in catalog. Run seed-catalog.sh first."
  echo ""
  echo "================================"
  echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
  echo "================================"
  exit 1
fi

pass "Catalog movie found: \"${MOVIE_TITLE}\" (${CONTENT_ID})"

# ── Upload Test Video ──────────────────────────────────────
VIDEO_SIZE=$(stat -c%s "$VIDEO_PATH" 2>/dev/null || stat -f%z "$VIDEO_PATH")
info "Uploading $(basename "$VIDEO_PATH") ($(numfmt --to=iec "$VIDEO_SIZE" 2>/dev/null || echo "${VIDEO_SIZE} bytes"))..."

UPLOAD_RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/stream/upload" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "content_id=${CONTENT_ID}" \
  -F "file=@${VIDEO_PATH};type=video/mp4")

UPLOAD_STATUS=$(echo "$UPLOAD_RESP" | tail -1)
UPLOAD_BODY=$(echo "$UPLOAD_RESP" | sed '$d')

if [ "$UPLOAD_STATUS" = "202" ]; then
  JOB_ID=$(echo "$UPLOAD_BODY" | jq -r '.job_id')
  pass "Video upload -> 202 Accepted (job_id: ${JOB_ID})"
else
  fail "Video upload -> HTTP ${UPLOAD_STATUS} (expected 202)"
  echo "       Response: $(echo "$UPLOAD_BODY" | head -c 300)"
  JOB_ID=""
fi

# ── Poll Encoding Job Status ──────────────────────────────
if [ -n "$JOB_ID" ]; then
  info "Polling encoding job status (every ${POLL_INTERVAL}s, timeout ${POLL_TIMEOUT}s)..."

  ELAPSED=0
  LAST_STATUS=""
  ENCODING_OK=false

  while [ "$ELAPSED" -lt "$POLL_TIMEOUT" ]; do
    JOB_RESP=$(curl -s "${API_URL}/api/encoding/jobs/${JOB_ID}" \
      -H "Authorization: Bearer ${TOKEN}" 2>/dev/null)

    JOB_STATUS=$(echo "$JOB_RESP" | jq -r '.status // empty' 2>/dev/null)

    if [ -n "$JOB_STATUS" ] && [ "$JOB_STATUS" != "$LAST_STATUS" ]; then
      info "Job status: ${JOB_STATUS} (${ELAPSED}s)"
      LAST_STATUS="$JOB_STATUS"
    fi

    if [ "$JOB_STATUS" = "completed" ]; then
      ENCODING_OK=true
      break
    elif [ "$JOB_STATUS" = "failed" ]; then
      ERROR_MSG=$(echo "$JOB_RESP" | jq -r '.error_message // "unknown"' 2>/dev/null)
      fail "Encoding job failed: ${ERROR_MSG}"
      break
    fi

    sleep "$POLL_INTERVAL"
    ELAPSED=$((ELAPSED + POLL_INTERVAL))
  done

  if [ "$ENCODING_OK" = true ]; then
    pass "Encoding completed successfully (${ELAPSED}s)"
  elif [ "$ELAPSED" -ge "$POLL_TIMEOUT" ]; then
    fail "Encoding timed out after ${POLL_TIMEOUT}s (last status: ${LAST_STATUS})"
  fi
fi

# ── Verify MinIO Outputs ──────────────────────────────────
section "7.1 — MinIO Output Verification"

check_minio_object() {
  local BUCKET=$1
  local KEY=$2
  local LABEL=$3

  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -u "${MINIO_ACCESS_KEY}:${MINIO_SECRET_KEY}" \
    "${MINIO_URL}/${BUCKET}/${KEY}" 2>/dev/null)

  if [ "$HTTP_CODE" = "200" ]; then
    pass "MinIO: ${LABEL}"
  else
    fail "MinIO: ${LABEL} (HTTP ${HTTP_CODE})"
  fi
}

check_minio_prefix() {
  local BUCKET=$1
  local PREFIX=$2
  local PATTERN=$3
  local LABEL=$4

  LIST_RESP=$(curl -s \
    -u "${MINIO_ACCESS_KEY}:${MINIO_SECRET_KEY}" \
    "${MINIO_URL}/${BUCKET}?prefix=${PREFIX}&list-type=2" 2>/dev/null)

  MATCH_COUNT=$(echo "$LIST_RESP" | grep -o "$PATTERN" | wc -l)

  if [ "$MATCH_COUNT" -gt 0 ]; then
    pass "MinIO: ${LABEL} (${MATCH_COUNT} match(es))"
  else
    fail "MinIO: ${LABEL} (0 matches)"
  fi
}

if [ "${ENCODING_OK:-false}" = true ]; then
  # HLS playlists
  check_minio_object "streamvault-encoded" "${CONTENT_ID}/360p/playlist.m3u8" "360p playlist exists"
  check_minio_object "streamvault-encoded" "${CONTENT_ID}/720p/playlist.m3u8" "720p playlist exists"

  # HLS segments
  check_minio_prefix "streamvault-encoded" "${CONTENT_ID}/360p/" "\.ts" "360p segments (.ts files)"
  check_minio_prefix "streamvault-encoded" "${CONTENT_ID}/720p/" "\.ts" "720p segments (.ts files)"

  # Thumbnails
  check_minio_object "streamvault-thumbnails" "${CONTENT_ID}/poster.jpg" "Poster image exists"
  check_minio_object "streamvault-thumbnails" "${CONTENT_ID}/thumb_300x170.jpg" "Thumbnail image exists"
else
  info "Skipping MinIO verification (encoding did not complete)"
fi

# ── Verify Catalog Update ─────────────────────────────────
section "7.1 — Catalog Status Verification"

if [ "${ENCODING_OK:-false}" = true ]; then
  STREAMING_RESP=$(curl -s "${API_URL}/api/catalog/movies/${CONTENT_ID}/streaming-info")
  VIDEO_STATUS=$(echo "$STREAMING_RESP" | jq -r '.videoStatus // empty' 2>/dev/null)

  if [ "$VIDEO_STATUS" = "Ready" ]; then
    pass "Catalog videoStatus = Ready"
  else
    fail "Catalog videoStatus = '${VIDEO_STATUS}' (expected 'Ready')"
  fi

  QUALITY_COUNT=$(echo "$STREAMING_RESP" | jq '.availableQualities | length' 2>/dev/null || echo "0")
  if [ "$QUALITY_COUNT" -ge 2 ]; then
    pass "Catalog has ${QUALITY_COUNT} quality level(s) (>= 2)"
  else
    fail "Catalog has ${QUALITY_COUNT} quality level(s) (expected >= 2)"
  fi

  MANIFEST_URL=$(echo "$STREAMING_RESP" | jq -r '.manifestUrl // empty' 2>/dev/null)
  if [ -n "$MANIFEST_URL" ]; then
    pass "Catalog manifestUrl populated: ${MANIFEST_URL}"
  else
    fail "Catalog manifestUrl is empty"
  fi
else
  info "Skipping catalog verification (encoding did not complete)"
fi

# ── 7.3 Error Scenarios ───────────────────────────────────
section "7.3 — Error Scenarios"

# Test: upload unsupported file type
TMPFILE=$(mktemp /tmp/invalid-upload-XXXXXX.txt)
echo "this is not a video file" > "$TMPFILE"

ERR_RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/stream/upload" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "content_id=error-test-content" \
  -F "file=@${TMPFILE};type=text/plain")

ERR_STATUS=$(echo "$ERR_RESP" | tail -1)
rm -f "$TMPFILE"

if [ "$ERR_STATUS" = "400" ]; then
  pass "Invalid file type rejected -> 400"
else
  fail "Invalid file type -> HTTP ${ERR_STATUS} (expected 400)"
fi

# Test: upload without auth
NOAUTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${API_URL}/api/stream/upload" \
  -F "content_id=test" \
  -F "file=@${VIDEO_PATH};type=video/mp4")

if [ "$NOAUTH_STATUS" = "401" ]; then
  pass "Upload without auth -> 401"
else
  fail "Upload without auth -> HTTP ${NOAUTH_STATUS} (expected 401)"
fi

# Test: upload without content_id
NOID_RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/stream/upload" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "file=@${VIDEO_PATH};type=video/mp4")

NOID_STATUS=$(echo "$NOID_RESP" | tail -1)

if [ "$NOID_STATUS" = "400" ]; then
  pass "Upload without content_id -> 400"
else
  fail "Upload without content_id -> HTTP ${NOID_STATUS} (expected 400)"
fi

# ── Summary ────────────────────────────────────────────────
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

# Export content ID for streaming test script
if [ "${ENCODING_OK:-false}" = true ]; then
  echo ""
  info "Content ID for streaming tests: ${CONTENT_ID}"
  echo "  Run: CONTENT_ID=${CONTENT_ID} ./scripts/test-streaming.sh"
fi

rm -f /tmp/sv_response.json

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
