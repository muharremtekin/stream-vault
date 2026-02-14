#!/bin/bash
# StreamVault - API Test Script
# Tests all endpoints end-to-end through the API Gateway

set -e

API_URL="${API_URL:-http://localhost:8081}"
PASS=0
FAIL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

test_endpoint() {
  local METHOD=$1
  local ENDPOINT=$2
  local EXPECTED_STATUS=$3
  local DATA=$4
  local AUTH_HEADER=$5

  if [ -n "$DATA" ]; then
    ACTUAL_STATUS=$(curl -s -o /tmp/sv_response.json -w "%{http_code}" \
      -X "$METHOD" "$API_URL$ENDPOINT" \
      -H "Content-Type: application/json" \
      ${AUTH_HEADER:+-H "Authorization: Bearer $AUTH_HEADER"} \
      -d "$DATA")
  else
    ACTUAL_STATUS=$(curl -s -o /tmp/sv_response.json -w "%{http_code}" \
      -X "$METHOD" "$API_URL$ENDPOINT" \
      -H "Content-Type: application/json" \
      ${AUTH_HEADER:+-H "Authorization: Bearer $AUTH_HEADER"})
  fi

  if [ "$ACTUAL_STATUS" = "$EXPECTED_STATUS" ]; then
    echo -e "  ${GREEN}PASS${NC} $METHOD $ENDPOINT -> $ACTUAL_STATUS"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}FAIL${NC} $METHOD $ENDPOINT -> $ACTUAL_STATUS (expected $EXPECTED_STATUS)"
    echo -e "       Response: $(cat /tmp/sv_response.json | head -c 200)"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== StreamVault API Test Suite ==="
echo "    Gateway: $API_URL"
echo ""

# ── Health ──────────────────────────────────────────────────
echo "[Health]"
test_endpoint GET /health 200

# ── Auth: Register ──────────────────────────────────────────
echo ""
echo "[Auth]"
TIMESTAMP=$(date +%s)
TEST_EMAIL="test${TIMESTAMP}@example.com"
TEST_PASSWORD="TestPass123@"

test_endpoint POST /api/auth/register 201 \
  "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD\"}"

# ── Auth: Login ─────────────────────────────────────────────
test_endpoint POST /api/auth/login 200 \
  "{\"email\": \"$TEST_EMAIL\", \"password\": \"$TEST_PASSWORD\"}"

LOGIN_RESPONSE=$(cat /tmp/sv_response.json)
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.accessToken')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refreshToken')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user.id')

# ── Auth: Refresh Token ────────────────────────────────────
test_endpoint POST /api/auth/refresh 200 \
  "{\"refreshToken\": \"$REFRESH_TOKEN\"}"

# Use the new access token from now on
TOKEN=$(jq -r '.accessToken' /tmp/sv_response.json)

# ── Auth: Unauthorized Access ──────────────────────────────
test_endpoint GET /api/users/me 401

# ── User: Get Current User ─────────────────────────────────
echo ""
echo "[User]"
test_endpoint GET /api/users/me 200 "" "$TOKEN"

# ── Profile: Create ────────────────────────────────────────
echo ""
echo "[Profile]"
test_endpoint POST /api/profile 201 \
  "{\"userId\": \"$USER_ID\", \"name\": \"Test Profile\", \"icon\": 1, \"isKids\": false}" "$TOKEN"

PROFILE_ID=$(jq -r '.id' /tmp/sv_response.json)

# ── Profile: List ──────────────────────────────────────────
test_endpoint GET "/api/profile/user/$USER_ID" 200 "" "$TOKEN"

# ── Catalog: Movies ────────────────────────────────────────
echo ""
echo "[Catalog]"
test_endpoint GET /api/catalog/movies 200

MOVIE_ID=$(jq -r '.items[0].id' /tmp/sv_response.json)

# ── Catalog: Genres ────────────────────────────────────────
test_endpoint GET /api/catalog/genres 200

# ── Catalog: Series ────────────────────────────────────────
test_endpoint GET /api/catalog/series 200

# ── Watchlist: Add ─────────────────────────────────────────
echo ""
echo "[Watchlist]"
test_endpoint POST /api/watchlist 201 \
  "{\"profileId\": \"$PROFILE_ID\", \"contentId\": \"$MOVIE_ID\", \"contentType\": 0}" "$TOKEN"

WATCHLIST_ID=$(jq -r '.id' /tmp/sv_response.json)

# ── Watchlist: List ────────────────────────────────────────
test_endpoint GET "/api/watchlist/profile/$PROFILE_ID" 200 "" "$TOKEN"

# ── Watchlist: Remove ──────────────────────────────────────
test_endpoint DELETE "/api/watchlist/$WATCHLIST_ID" 204 "" "$TOKEN"

# ── Rate Limiting ──────────────────────────────────────────
echo ""
echo "[Rate Limit]"
echo -e "  ${YELLOW}INFO${NC} Sending rapid parallel requests to exhaust burst bucket..."
RATE_LOG=$(mktemp)

# Send bursts of parallel requests to overwhelm the token bucket
for batch in $(seq 1 10); do
  for i in $(seq 1 50); do
    curl -s -o /dev/null -w "%{http_code}\n" "$API_URL/health" >> "$RATE_LOG" &
  done
  wait
done

RATE_429=$(grep -c "429" "$RATE_LOG" 2>/dev/null || echo "0")
rm -f "$RATE_LOG"

if [ "$RATE_429" -gt 0 ]; then
  echo -e "  ${GREEN}PASS${NC} Rate limit triggered ($RATE_429 x 429 responses)"
  PASS=$((PASS + 1))
else
  echo -e "  ${RED}FAIL${NC} Rate limit NOT triggered (0 x 429 in 500 parallel requests)"
  FAIL=$((FAIL + 1))
fi

# ── Summary ────────────────────────────────────────────────
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

# Cleanup
rm -f /tmp/sv_response.json

if [ $FAIL -gt 0 ]; then
  exit 1
fi
