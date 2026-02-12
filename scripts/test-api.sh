#!/bin/bash
# StreamVault - API Test Script
# Tests all endpoints end-to-end

set -e

API_URL="${API_URL:-http://localhost:8080}"
PASS=0
FAIL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

test_endpoint() {
  local METHOD=$1
  local PATH=$2
  local EXPECTED_STATUS=$3
  local DATA=$4
  local AUTH_HEADER=$5

  local HEADERS="-H 'Content-Type: application/json'"
  if [ -n "$AUTH_HEADER" ]; then
    HEADERS="$HEADERS -H 'Authorization: Bearer $AUTH_HEADER'"
  fi

  if [ -n "$DATA" ]; then
    ACTUAL_STATUS=$(curl -s -o /tmp/sv_response.json -w "%{http_code}" \
      -X "$METHOD" "$API_URL$PATH" \
      -H "Content-Type: application/json" \
      ${AUTH_HEADER:+-H "Authorization: Bearer $AUTH_HEADER"} \
      -d "$DATA")
  else
    ACTUAL_STATUS=$(curl -s -o /tmp/sv_response.json -w "%{http_code}" \
      -X "$METHOD" "$API_URL$PATH" \
      -H "Content-Type: application/json" \
      ${AUTH_HEADER:+-H "Authorization: Bearer $AUTH_HEADER"})
  fi

  if [ "$ACTUAL_STATUS" = "$EXPECTED_STATUS" ]; then
    echo -e "  ${GREEN}PASS${NC} $METHOD $PATH -> $ACTUAL_STATUS"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}FAIL${NC} $METHOD $PATH -> $ACTUAL_STATUS (expected $EXPECTED_STATUS)"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== StreamVault API Test Suite ==="
echo ""

# Health check
echo "[Health]"
test_endpoint GET /health 200

# Auth - Register
echo ""
echo "[Auth]"
TIMESTAMP=$(date +%s)
TEST_EMAIL="test${TIMESTAMP}@example.com"

test_endpoint POST /api/auth/register 201 \
  "{\"email\": \"$TEST_EMAIL\", \"password\": \"TestPass123!\"}"

# Auth - Login
test_endpoint POST /api/auth/login 200 \
  "{\"email\": \"$TEST_EMAIL\", \"password\": \"TestPass123!\"}"

TOKEN=$(cat /tmp/sv_response.json | jq -r '.accessToken')

# Auth - Unauthorized access
test_endpoint GET /api/users/me 401

# User - Get current user
echo ""
echo "[User]"
test_endpoint GET /api/users/me 200 "" "$TOKEN"

# Profile - Create
echo ""
echo "[Profile]"
test_endpoint POST /api/users/me/profiles 201 \
  '{"name": "Test Profile", "icon": "Avatar1", "isKids": false}' "$TOKEN"

PROFILE_ID=$(cat /tmp/sv_response.json | jq -r '.id')

# Profile - List
test_endpoint GET /api/users/me/profiles 200 "" "$TOKEN"

# Catalog - Movies
echo ""
echo "[Catalog]"
test_endpoint GET /api/catalog/movies 200 "" "$TOKEN"
test_endpoint GET /api/catalog/series 200 "" "$TOKEN"
test_endpoint GET /api/catalog/genres 200 "" "$TOKEN"

# Watchlist
echo ""
echo "[Watchlist]"
test_endpoint POST "/api/users/me/profiles/${PROFILE_ID}/watchlist" 201 \
  '{"contentId": "test-movie-1", "contentType": "Movie"}' "$TOKEN"

test_endpoint GET "/api/users/me/profiles/${PROFILE_ID}/watchlist" 200 "" "$TOKEN"

# Rate limiting test
echo ""
echo "[Rate Limit]"
echo "  Sending 10 rapid requests..."
for i in $(seq 1 10); do
  curl -s -o /dev/null "$API_URL/health" &
done
wait
echo "  Rate limit test completed"

# Summary
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

# Cleanup
rm -f /tmp/sv_response.json

if [ $FAIL -gt 0 ]; then
  exit 1
fi
