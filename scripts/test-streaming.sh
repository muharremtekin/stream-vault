#!/bin/bash
# StreamVault - Streaming Flow Test
# Tests: 7.2 (Streaming, progress tracking, concurrent limits)
#
# Prerequisites:
#   - All services running via docker compose (user-service seeds tier users on startup)
#   - upload-test-video.sh completed successfully (content is encoded)
#   - CONTENT_ID env var set (or auto-detected from catalog)

set -euo pipefail

API_URL="${API_URL:-http://localhost:8081}"
SEED_PASSWORD="${SEED_PASSWORD:-SeedPass123@}"
CONTENT_ID="${CONTENT_ID:-}"

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

# Helper: login and return token
login_as() {
  local EMAIL=$1
  local LABEL=$2

  local RESP
  RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"${EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")

  local STATUS
  STATUS=$(echo "$RESP" | tail -1)
  local BODY
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "200" ]; then
    echo "$BODY" | jq -r '.accessToken'
    pass "${LABEL} login -> 200"
  else
    fail "${LABEL} login -> HTTP ${STATUS} (expected 200)"
    echo ""
  fi
}

# ── Prerequisites ──────────────────────────────────────────
echo "=== StreamVault Streaming Flow Test ==="
echo "    Gateway: $API_URL"
echo ""

for cmd in curl jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: '$cmd' is required but not installed."
    exit 1
  fi
done

# ── Login Seeded Users ────────────────────────────────────
section "Setup — Login Seeded Users"

ADMIN_TOKEN=$(login_as "admin@streamvault.com" "Admin")
FREE_TOKEN=$(login_as "free@streamvault.com" "Free")
BASIC_TOKEN=$(login_as "basic@streamvault.com" "Basic")
STANDARD_TOKEN=$(login_as "standard@streamvault.com" "Standard")
PREMIUM_TOKEN=$(login_as "premium@streamvault.com" "Premium")

if [ -z "$ADMIN_TOKEN" ]; then
  echo "ERROR: Admin login failed. Aborting."
  exit 1
fi

# ── Auto-detect Content ID ────────────────────────────────
if [ -z "$CONTENT_ID" ]; then
  info "CONTENT_ID not set, detecting from catalog..."
  MOVIES_BODY=$(curl -s "${API_URL}/api/catalog/movies")
  CONTENT_ID=$(echo "$MOVIES_BODY" | jq -r '[.items[] | select(.status == "Ready")] | .[0].id // empty' 2>/dev/null)

  if [ -z "$CONTENT_ID" ] || [ "$CONTENT_ID" = "null" ]; then
    echo "ERROR: No encoded movies found. Run upload-test-video.sh first."
    exit 1
  fi
fi

info "Using content_id: ${CONTENT_ID}"

# ── 7.2 Subscription Tier Enforcement ─────────────────────
section "7.2 — Subscription Tier Enforcement"

# Free tier -> 403
if [ -n "$FREE_TOKEN" ]; then
  FREE_STREAM_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${FREE_TOKEN}")

  if [ "$FREE_STREAM_STATUS" = "403" ]; then
    pass "Free tier streaming blocked -> 403"
  else
    fail "Free tier streaming -> HTTP ${FREE_STREAM_STATUS} (expected 403)"
  fi
fi

# Basic tier -> 200 (allowed)
if [ -n "$BASIC_TOKEN" ]; then
  BASIC_STREAM_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${BASIC_TOKEN}")

  if [ "$BASIC_STREAM_STATUS" = "200" ]; then
    pass "Basic tier streaming allowed -> 200"
  else
    fail "Basic tier streaming -> HTTP ${BASIC_STREAM_STATUS} (expected 200)"
  fi
fi

# Standard tier -> 200
if [ -n "$STANDARD_TOKEN" ]; then
  STANDARD_STREAM_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${STANDARD_TOKEN}")

  if [ "$STANDARD_STREAM_STATUS" = "200" ]; then
    pass "Standard tier streaming allowed -> 200"
  else
    fail "Standard tier streaming -> HTTP ${STANDARD_STREAM_STATUS} (expected 200)"
  fi
fi

# Premium tier -> 200
if [ -n "$PREMIUM_TOKEN" ]; then
  PREMIUM_STREAM_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${PREMIUM_TOKEN}")

  if [ "$PREMIUM_STREAM_STATUS" = "200" ]; then
    pass "Premium tier streaming allowed -> 200"
  else
    fail "Premium tier streaming -> HTTP ${PREMIUM_STREAM_STATUS} (expected 200)"
  fi
fi

# ── 7.2 Master Playlist ──────────────────────────────────
section "7.2 — HLS Streaming (Admin)"

MANIFEST_RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}")

MANIFEST_STATUS=$(echo "$MANIFEST_RESP" | tail -1)
MANIFEST_BODY=$(echo "$MANIFEST_RESP" | sed '$d')

if [ "$MANIFEST_STATUS" = "200" ]; then
  pass "Master playlist -> 200"
else
  fail "Master playlist -> HTTP ${MANIFEST_STATUS} (expected 200)"
  echo "       Response: $(echo "$MANIFEST_BODY" | head -c 300)"
fi

# Validate master playlist content
if echo "$MANIFEST_BODY" | grep -q "#EXTM3U"; then
  pass "Master playlist has #EXTM3U header"
else
  fail "Master playlist missing #EXTM3U header"
fi

if echo "$MANIFEST_BODY" | grep -q "#EXT-X-STREAM-INF"; then
  STREAM_COUNT=$(echo "$MANIFEST_BODY" | grep -c "#EXT-X-STREAM-INF")
  pass "Master playlist has ${STREAM_COUNT} variant stream(s)"
else
  fail "Master playlist missing #EXT-X-STREAM-INF entries"
fi

if echo "$MANIFEST_BODY" | grep -q "360p"; then
  pass "Master playlist includes 360p variant"
else
  fail "Master playlist missing 360p variant"
fi

if echo "$MANIFEST_BODY" | grep -q "720p"; then
  pass "Master playlist includes 720p variant"
else
  fail "Master playlist missing 720p variant"
fi

# ── 7.2 Variant Playlist ─────────────────────────────────
VARIANT_PATH=$(echo "$MANIFEST_BODY" | grep -v "^#" | head -1)
if [ -n "$VARIANT_PATH" ]; then
  VARIANT_RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}${VARIANT_PATH}" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}")

  VARIANT_STATUS=$(echo "$VARIANT_RESP" | tail -1)
  VARIANT_BODY=$(echo "$VARIANT_RESP" | sed '$d')

  if [ "$VARIANT_STATUS" = "200" ]; then
    pass "Variant playlist -> 200 (${VARIANT_PATH})"
  else
    fail "Variant playlist -> HTTP ${VARIANT_STATUS} (${VARIANT_PATH})"
  fi

  if echo "$VARIANT_BODY" | grep -q "#EXTINF"; then
    SEGMENT_COUNT=$(echo "$VARIANT_BODY" | grep -c "#EXTINF")
    pass "Variant playlist has ${SEGMENT_COUNT} segment(s)"
  else
    fail "Variant playlist missing #EXTINF entries"
  fi

  # Extract first segment URL
  SEGMENT_PATH=$(echo "$VARIANT_BODY" | grep -v "^#" | head -1)
  if [ -n "$SEGMENT_PATH" ]; then
    VARIANT_DIR=$(dirname "$VARIANT_PATH")
    SEGMENT_URL="${VARIANT_DIR}/${SEGMENT_PATH}"

    SEG_RESP=$(curl -s -o /dev/null -w "%{http_code}" \
      "${API_URL}${SEGMENT_URL}" \
      -H "Authorization: Bearer ${ADMIN_TOKEN}")

    if [ "$SEG_RESP" = "200" ]; then
      pass "Video segment served -> 200 (${SEGMENT_PATH})"
    else
      fail "Video segment -> HTTP ${SEG_RESP} (${SEGMENT_URL})"
    fi
  fi
else
  fail "Could not extract variant playlist URL from master"
fi

# ── 7.2 Progress Tracking ────────────────────────────────
section "7.2 — Progress Tracking"

# Save progress (using Basic user)
if [ -n "$BASIC_TOKEN" ]; then
  SAVE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${API_URL}/api/stream/${CONTENT_ID}/progress" \
    -H "Authorization: Bearer ${BASIC_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"position_seconds": 120, "duration_seconds": 3600}')

  if [ "$SAVE_STATUS" = "200" ]; then
    pass "Save progress (Basic) -> 200"
  else
    fail "Save progress (Basic) -> HTTP ${SAVE_STATUS} (expected 200)"
  fi

  # Get progress
  PROGRESS_RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}/api/stream/${CONTENT_ID}/progress" \
    -H "Authorization: Bearer ${BASIC_TOKEN}")

  PROGRESS_STATUS=$(echo "$PROGRESS_RESP" | tail -1)
  PROGRESS_BODY=$(echo "$PROGRESS_RESP" | sed '$d')

  if [ "$PROGRESS_STATUS" = "200" ]; then
    POSITION=$(echo "$PROGRESS_BODY" | jq '.position_seconds' 2>/dev/null)
    if [ "$POSITION" = "120" ]; then
      pass "Get progress -> position_seconds = 120"
    else
      fail "Get progress -> position_seconds = ${POSITION} (expected 120)"
    fi
  else
    fail "Get progress -> HTTP ${PROGRESS_STATUS} (expected 200)"
  fi

  # Continue watching list
  CW_RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}/api/stream/continue-watching" \
    -H "Authorization: Bearer ${BASIC_TOKEN}")

  CW_STATUS=$(echo "$CW_RESP" | tail -1)
  CW_BODY=$(echo "$CW_RESP" | sed '$d')

  if [ "$CW_STATUS" = "200" ]; then
    pass "Continue-watching list -> 200"

    CW_HAS_CONTENT=$(echo "$CW_BODY" | jq --arg cid "$CONTENT_ID" \
      '[.items[] | select(.content_id == $cid)] | length' 2>/dev/null || echo "0")

    if [ "$CW_HAS_CONTENT" -gt 0 ]; then
      pass "Continue-watching contains content_id ${CONTENT_ID}"
    else
      fail "Continue-watching does not contain content_id ${CONTENT_ID}"
    fi
  else
    fail "Continue-watching list -> HTTP ${CW_STATUS} (expected 200)"
  fi
fi

# ── 7.2 Concurrent Stream Limit ──────────────────────────
section "7.2 — Concurrent Stream Limit"

# Basic tier: max 1 concurrent stream — send 2 sequential, expect 429 on 2nd
if [ -n "$BASIC_TOKEN" ]; then
  info "Testing Basic tier limit (max=1, sending 2 sequential requests)..."

  # First request registers a session
  curl -s -o /dev/null \
    "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${BASIC_TOKEN}" \
    -H "X-Session-Id: basic-session-1"

  # Second request with different session should be blocked
  BASIC_SECOND=$(curl -s -o /dev/null -w "%{http_code}" \
    "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${BASIC_TOKEN}" \
    -H "X-Session-Id: basic-session-2")

  if [ "$BASIC_SECOND" = "429" ]; then
    pass "Basic tier: 2nd concurrent stream blocked -> 429"
  else
    fail "Basic tier: 2nd concurrent stream -> HTTP ${BASIC_SECOND} (expected 429)"
  fi
fi

# Admin tier: max 4 — occupy 4, then 5th should fail
info "Testing Admin tier limit (max=4, sending 5 sequential requests)..."

for i in $(seq 1 4); do
  curl -s -o /dev/null \
    "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "X-Session-Id: admin-session-${i}"
done

FIFTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  "${API_URL}/stream/${CONTENT_ID}/manifest.m3u8" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "X-Session-Id: admin-session-5")

if [ "$FIFTH_STATUS" = "429" ]; then
  pass "Admin tier: 5th concurrent stream blocked -> 429"
else
  fail "Admin tier: 5th concurrent stream -> HTTP ${FIFTH_STATUS} (expected 429)"
fi

# ── Summary ────────────────────────────────────────────────
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
