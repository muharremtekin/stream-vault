#!/bin/bash
# StreamVault - End-to-End Full Flow Test
# Tasks 8.7: Register → Subscribe → Watch → Rate → Search → Recommendation
#             Plan upgrade/downgrade, tier sync, cancel, invoices
#
# Prerequisites:
#   - All services running including elasticsearch profile
#   - docker compose --profile elasticsearch up -d

set -euo pipefail

API_URL="${API_URL:-http://localhost:8081}"
SEED_PASSWORD="${SEED_PASSWORD:-SeedPass123@}"

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

echo "=== StreamVault End-to-End Flow Test ==="
echo "    Gateway: $API_URL"
echo ""

for cmd in curl jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: '$cmd' is required but not installed."
    exit 1
  fi
done

# Plan GUIDs (from DbSeeder.cs)
BASIC_PLAN_ID="a1b2c3d4-0001-0001-0001-000000000001"
STANDARD_PLAN_ID="a1b2c3d4-0001-0001-0001-000000000002"
PREMIUM_PLAN_ID="a1b2c3d4-0001-0001-0001-000000000003"

SUCCESS_CARD="4242424242424242"
TS=$(date +%s)

# ══════════════════════════════════════════════════════════
# FLOW 1: Register → Subscribe → Watch → Rate → Search → Recommend
# ══════════════════════════════════════════════════════════

section "Flow 1 — Register & Login"

E2E_EMAIL="e2e_${TS}@example.com"
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${E2E_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)

if [ "$STATUS" = "201" ] || [ "$STATUS" = "200" ]; then
  pass "Registered user: ${E2E_EMAIL}"
else
  fail "Register user -> HTTP ${STATUS}"
fi

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${E2E_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  TOKEN=$(echo "$BODY" | jq -r '.accessToken')
  USER_ID=$(echo "$BODY" | jq -r '.user.id // empty')
  pass "Login -> 200"
else
  fail "Login -> HTTP ${STATUS}"
  echo "ERROR: Cannot proceed without token. Aborting."
  exit 1
fi

if [ -z "$USER_ID" ] || [ "$USER_ID" = "null" ]; then
  ME_RESP=$(curl -s "${API_URL}/api/users/me" -H "Authorization: Bearer ${TOKEN}")
  USER_ID=$(echo "$ME_RESP" | jq -r '.id // empty')
fi
info "User ID: ${USER_ID}"

# ── Subscribe to Standard Plan ───────────────────────────
section "Flow 1 — Subscribe (Standard Plan)"

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/subscriptions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"planId\": \"${STANDARD_PLAN_ID}\", \"cardNumber\": \"${SUCCESS_CARD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "201" ] || [ "$STATUS" = "200" ]; then
  pass "Subscribe to Standard -> ${STATUS}"
else
  fail "Subscribe -> HTTP ${STATUS}"
  echo "       Response: $(echo "$BODY" | head -c 300)"
fi

SUB_STATUS=$(echo "$BODY" | jq -r '.status // empty' 2>/dev/null)
if [ "$SUB_STATUS" = "Active" ]; then
  pass "Subscription status = Active"
else
  fail "Subscription status = '${SUB_STATUS}' (expected Active)"
fi

# ── Verify Tier Sync ─────────────────────────────────────
section "Flow 1 — Tier Sync (Subscription -> User Service)"

info "Polling /api/users/me for up to 15 seconds..."
TIER_SYNCED=false
for i in $(seq 1 5); do
  sleep 3
  ME_RESP=$(curl -s "${API_URL}/api/users/me" -H "Authorization: Bearer ${TOKEN}")
  USER_ROLE=$(echo "$ME_RESP" | jq -r '.role // empty' 2>/dev/null)
  if [ -z "$USER_ROLE" ]; then
    USER_ROLE=$(echo "$ME_RESP" | jq -r '.tier // empty' 2>/dev/null)
  fi
  info "  Attempt ${i}/5: role = '${USER_ROLE}'"
  if [ "$USER_ROLE" = "Standard" ]; then
    TIER_SYNCED=true
    break
  fi
done

if [ "$TIER_SYNCED" = "true" ]; then
  pass "User tier synced to Standard"
else
  fail "User tier not synced to Standard within 15s (got '${USER_ROLE}')"
fi

# ── Re-login to get JWT with updated tier ─────────────────
section "Flow 1 — Re-login (Updated Tier)"

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${E2E_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  TOKEN=$(echo "$BODY" | jq -r '.accessToken')
  pass "Re-login -> 200 (JWT now has Standard tier)"
else
  fail "Re-login -> HTTP ${STATUS}"
fi

# ── Browse Catalog ────────────────────────────────────────
section "Flow 1 — Browse Catalog"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/catalog/movies" \
  -H "Authorization: Bearer ${TOKEN}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/catalog/movies -> 200"
else
  fail "GET /api/catalog/movies -> HTTP ${STATUS}"
fi

MOVIE_COUNT=$(echo "$BODY" | jq '.items | length' 2>/dev/null || echo "0")
if [ "$MOVIE_COUNT" -ge 1 ]; then
  MOVIE_ID=$(echo "$BODY" | jq -r '.items[0].id // empty' 2>/dev/null)
  MOVIE_TITLE=$(echo "$BODY" | jq -r '.items[0].title // empty' 2>/dev/null)
  pass "Catalog has ${MOVIE_COUNT} movie(s), using: '${MOVIE_TITLE}'"
else
  fail "Catalog empty (${MOVIE_COUNT} items)"
  MOVIE_ID=""
  MOVIE_TITLE=""
fi

# ── Streaming Quality Access (Standard Tier) ─────────────
section "Flow 1 — Streaming Quality Access (Standard Tier)"

if [ -n "$MOVIE_ID" ]; then
  # Standard tier should pass subscription middleware (not 403)
  RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}/stream/${MOVIE_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)

  if [ "$STATUS" = "403" ]; then
    fail "Streaming manifest -> 403 (Standard tier should have access)"
  else
    pass "Streaming manifest -> HTTP ${STATUS} (not 403, subscription check passed)"
  fi
else
  info "Skipping streaming access test (no movie from catalog)"
fi

# ── Watch Content (Simulate Watching) ────────────────────
section "Flow 1 — Watch Content"

if [ -n "$MOVIE_ID" ]; then
  # Save progress at 90% to trigger WatchCompleted event
  RESP=$(curl -s -w "\n%{http_code}" -X POST \
    "${API_URL}/api/stream/${MOVIE_ID}/progress" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "{\"position_seconds\": 3240, \"duration_seconds\": 3600}")
  STATUS=$(echo "$RESP" | tail -1)

  if [ "$STATUS" = "200" ]; then
    pass "Save watch progress (90%) -> 200"
  else
    fail "Save watch progress -> HTTP ${STATUS}"
  fi

  # Verify progress is saved
  RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}/api/stream/${MOVIE_ID}/progress" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "200" ]; then
    PERCENTAGE=$(echo "$BODY" | jq '.percentage // 0' 2>/dev/null)
    pass "Get watch progress -> 200 (percentage: ${PERCENTAGE}%)"
  else
    fail "Get watch progress -> HTTP ${STATUS}"
  fi

  # Verify continue-watching list
  RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}/api/stream/continue-watching" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "200" ]; then
    CW_COUNT=$(echo "$BODY" | jq '.items | length' 2>/dev/null || echo "0")
    pass "Continue-watching -> 200 (${CW_COUNT} item(s))"
  else
    fail "Continue-watching -> HTTP ${STATUS}"
  fi

  # Wait for WatchCompleted event to propagate
  info "Waiting 3s for WatchCompleted event propagation..."
  sleep 3
else
  info "Skipping watch content (no movie from catalog)"
fi

# ── Rate Content ──────────────────────────────────────────
section "Flow 1 — Rate Content"

if [ -n "$MOVIE_ID" ]; then
  RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/users/me/ratings" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "{\"contentId\": \"${MOVIE_ID}\", \"rating\": 8}")
  STATUS=$(echo "$RESP" | tail -1)

  if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
    pass "Rate content -> ${STATUS}"
  else
    fail "Rate content -> HTTP ${STATUS}"
  fi

  # Verify rating saved
  RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/users/me/ratings" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "200" ]; then
    RATING_COUNT=$(echo "$BODY" | jq 'length' 2>/dev/null || echo "0")
    if [ "$RATING_COUNT" -ge 1 ]; then
      pass "User ratings count = ${RATING_COUNT} (>= 1)"
    else
      fail "User ratings count = ${RATING_COUNT} (expected >= 1)"
    fi
  else
    fail "GET /api/users/me/ratings -> HTTP ${STATUS}"
  fi
else
  info "Skipping rating (no movie from catalog)"
fi

# ── Search ────────────────────────────────────────────────
section "Flow 1 — Search"

if [ -n "$MOVIE_TITLE" ]; then
  # Wait a bit for ES indexing (catalog event → search service consumer)
  sleep 2

  SEARCH_Q=$(echo "$MOVIE_TITLE" | head -c 20)
  RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/search?q=$(echo "$SEARCH_Q" | jq -sRr @uri)" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "200" ]; then
    pass "Search '${SEARCH_Q}' -> 200"
  else
    fail "Search -> HTTP ${STATUS}"
  fi

  TOTAL=$(echo "$BODY" | jq '.totalCount // 0' 2>/dev/null)
  if [ "$TOTAL" -ge 1 ]; then
    pass "Search totalCount = ${TOTAL} (>= 1)"
  else
    fail "Search totalCount = ${TOTAL} (expected >= 1)"
  fi
else
  info "Skipping search (no movie title)"
fi

# ── Autocomplete ──────────────────────────────────────────
section "Flow 1 — Autocomplete"

if [ -n "$MOVIE_TITLE" ]; then
  AUTO_Q=$(echo "$MOVIE_TITLE" | head -c 3)
  RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/search/autocomplete?q=$(echo "$AUTO_Q" | jq -sRr @uri)" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "200" ]; then
    pass "Autocomplete '${AUTO_Q}' -> 200"
  else
    fail "Autocomplete -> HTTP ${STATUS}"
  fi

  SUGGESTIONS=$(echo "$BODY" | jq '.suggestions | length' 2>/dev/null || echo "0")
  if [ "$SUGGESTIONS" -ge 1 ]; then
    pass "Autocomplete suggestions = ${SUGGESTIONS} (>= 1)"
  else
    info "Autocomplete suggestions = ${SUGGESTIONS} (may be 0 if index is empty)"
  fi
else
  info "Skipping autocomplete (no movie title)"
fi

# ── Trending (Public) ─────────────────────────────────────
section "Flow 1 — Trending (Public)"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/search/trending?window=week")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/search/trending -> 200 (no auth)"
else
  fail "GET /api/search/trending -> HTTP ${STATUS}"
fi

# ── Recommendations ───────────────────────────────────────
section "Flow 1 — Recommendations"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/recommendations" \
  -H "Authorization: Bearer ${TOKEN}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/recommendations -> 200"
else
  fail "GET /api/recommendations -> HTTP ${STATUS}"
fi

# ── Home Page Sections ────────────────────────────────────
section "Flow 1 — Home Page Sections"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/recommendations/home" \
  -H "Authorization: Bearer ${TOKEN}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/recommendations/home -> 200"
else
  fail "GET /api/recommendations/home -> HTTP ${STATUS}"
fi

SECTION_COUNT=$(echo "$BODY" | jq '.sections | length' 2>/dev/null || echo "0")
if [ "$SECTION_COUNT" -ge 1 ]; then
  pass "Home page sections count = ${SECTION_COUNT} (>= 1)"
else
  info "Home page sections = ${SECTION_COUNT} (may be empty for cold-start)"
fi

# ── Similar Content Recommendations ──────────────────────
section "Flow 1 — Similar Content Recommendations"

if [ -n "$MOVIE_ID" ]; then
  RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}/api/recommendations/similar/${MOVIE_ID}" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "200" ]; then
    pass "GET /api/recommendations/similar/{id} -> 200"
  else
    fail "GET /api/recommendations/similar/{id} -> HTTP ${STATUS}"
  fi

  SIMILAR_COUNT=$(echo "$BODY" | jq '.items | length' 2>/dev/null || echo "0")
  if [ "$SIMILAR_COUNT" -ge 1 ]; then
    pass "Similar content items = ${SIMILAR_COUNT} (>= 1)"
  else
    info "Similar content items = ${SIMILAR_COUNT} (may be 0 without interaction data)"
  fi
else
  info "Skipping similar content (no movie from catalog)"
fi

# ══════════════════════════════════════════════════════════
# FLOW 2: Plan Upgrade & Downgrade
# ══════════════════════════════════════════════════════════

section "Flow 2 — Upgrade (Standard -> Premium)"

RESP=$(curl -s -w "\n%{http_code}" -X PUT "${API_URL}/api/subscriptions/me/plan" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"newPlanId\": \"${PREMIUM_PLAN_ID}\", \"cardNumber\": \"${SUCCESS_CARD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Upgrade to Premium -> 200"
else
  fail "Upgrade to Premium -> HTTP ${STATUS}"
  echo "       Response: $(echo "$BODY" | head -c 300)"
fi

PRICE_DIFF=$(echo "$BODY" | jq '.priceDifference // 0' 2>/dev/null)
IS_POSITIVE=$(echo "$PRICE_DIFF > 0" | bc -l 2>/dev/null || echo "0")
if [ "$IS_POSITIVE" = "1" ]; then
  pass "Upgrade price difference = ${PRICE_DIFF} (positive)"
else
  info "Upgrade price difference = ${PRICE_DIFF}"
fi

# Verify subscription updated
RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me" \
  -H "Authorization: Bearer ${TOKEN}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

PLAN_NOW=$(echo "$BODY" | jq -r '.plan.name // empty' 2>/dev/null)
if [ "$PLAN_NOW" = "Premium" ]; then
  pass "Current plan = Premium"
else
  fail "Current plan = '${PLAN_NOW}' (expected Premium)"
fi

# Verify tier sync
info "Polling for Premium tier sync..."
TIER_SYNCED=false
for i in $(seq 1 5); do
  sleep 3
  ME_RESP=$(curl -s "${API_URL}/api/users/me" -H "Authorization: Bearer ${TOKEN}")
  USER_ROLE=$(echo "$ME_RESP" | jq -r '.role // empty' 2>/dev/null)
  if [ -z "$USER_ROLE" ]; then
    USER_ROLE=$(echo "$ME_RESP" | jq -r '.tier // empty' 2>/dev/null)
  fi
  info "  Attempt ${i}/5: role = '${USER_ROLE}'"
  if [ "$USER_ROLE" = "Premium" ]; then
    TIER_SYNCED=true
    break
  fi
done

if [ "$TIER_SYNCED" = "true" ]; then
  pass "User tier synced to Premium"
else
  fail "User tier not synced to Premium within 15s (got '${USER_ROLE}')"
fi

# ── Re-login to get JWT with Premium tier ─────────────────
section "Flow 2 — Re-login (Premium Tier)"

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${E2E_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  TOKEN=$(echo "$BODY" | jq -r '.accessToken')
  pass "Re-login -> 200 (JWT now has Premium tier)"
else
  fail "Re-login -> HTTP ${STATUS}"
fi

# ── Streaming Quality Access (Premium Tier) ──────────────
section "Flow 2 — Streaming Quality Access (Premium Tier)"

if [ -n "$MOVIE_ID" ]; then
  RESP=$(curl -s -w "\n%{http_code}" \
    "${API_URL}/stream/${MOVIE_ID}/manifest.m3u8" \
    -H "Authorization: Bearer ${TOKEN}")
  STATUS=$(echo "$RESP" | tail -1)

  if [ "$STATUS" = "403" ]; then
    fail "Streaming manifest -> 403 (Premium tier should have full access)"
  else
    pass "Streaming manifest (Premium) -> HTTP ${STATUS} (not 403, full access)"
  fi
else
  info "Skipping streaming access test (no movie from catalog)"
fi

# ── Downgrade ─────────────────────────────────────────────
section "Flow 2 — Downgrade (Premium -> Basic)"

RESP=$(curl -s -w "\n%{http_code}" -X PUT "${API_URL}/api/subscriptions/me/plan" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"newPlanId\": \"${BASIC_PLAN_ID}\", \"cardNumber\": \"${SUCCESS_CARD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Downgrade to Basic -> 200"
else
  fail "Downgrade to Basic -> HTTP ${STATUS}"
fi

PRICE_DIFF=$(echo "$BODY" | jq '.priceDifference // 0' 2>/dev/null)
IS_NONPOSITIVE=$(echo "$PRICE_DIFF <= 0" | bc -l 2>/dev/null || echo "0")
if [ "$IS_NONPOSITIVE" = "1" ]; then
  pass "Downgrade price difference = ${PRICE_DIFF} (non-positive, credit)"
else
  info "Downgrade price difference = ${PRICE_DIFF}"
fi

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me" \
  -H "Authorization: Bearer ${TOKEN}")
BODY=$(echo "$RESP" | sed '$d')
PLAN_NOW=$(echo "$BODY" | jq -r '.plan.name // empty' 2>/dev/null)
if [ "$PLAN_NOW" = "Basic" ]; then
  pass "Current plan = Basic"
else
  fail "Current plan = '${PLAN_NOW}' (expected Basic)"
fi

# ══════════════════════════════════════════════════════════
# FLOW 3: Invoice, Cancel
# ══════════════════════════════════════════════════════════

# Check invoices BEFORE cancel (subscription must be findable)
section "Flow 3 — Invoice History"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me/invoices" \
  -H "Authorization: Bearer ${TOKEN}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/subscriptions/me/invoices -> 200"
else
  fail "GET /api/subscriptions/me/invoices -> HTTP ${STATUS}"
fi

IS_ARRAY=$(echo "$BODY" | jq 'type' 2>/dev/null || echo "")
if [ "$IS_ARRAY" = '"array"' ]; then
  INVOICE_COUNT=$(echo "$BODY" | jq 'length' 2>/dev/null || echo "0")
  if [ "$INVOICE_COUNT" -ge 1 ]; then
    pass "Invoice count = ${INVOICE_COUNT} (>= 1)"
  else
    fail "Invoice count = ${INVOICE_COUNT} (expected >= 1)"
  fi

  if [ "$INVOICE_COUNT" -ge 1 ]; then
    INV_NUMBER=$(echo "$BODY" | jq -r '.[0].invoiceNumber // empty' 2>/dev/null)
    if [ -n "$INV_NUMBER" ] && [ "$INV_NUMBER" != "null" ]; then
      pass "Invoice has invoiceNumber: ${INV_NUMBER}"
    else
      fail "Invoice missing invoiceNumber"
    fi
  fi
else
  fail "Invoice response is not an array"
fi

# ── Cancel Subscription ──────────────────────────────────
section "Flow 3 — Cancel Subscription"

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/subscriptions/me/cancel" \
  -H "Authorization: Bearer ${TOKEN}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Cancel subscription -> 200"
else
  fail "Cancel subscription -> HTTP ${STATUS}"
fi

CANCEL_STATUS=$(echo "$BODY" | jq -r '.status // empty' 2>/dev/null)
if [ "$CANCEL_STATUS" = "Cancelled" ]; then
  pass "Subscription status = Cancelled"
else
  fail "Subscription status = '${CANCEL_STATUS}' (expected Cancelled)"
fi

CANCELLED_AT=$(echo "$BODY" | jq -r '.cancelledAt // empty' 2>/dev/null)
if [ -n "$CANCELLED_AT" ] && [ "$CANCELLED_AT" != "null" ]; then
  pass "CancelledAt present"
else
  fail "CancelledAt missing"
fi

# ══════════════════════════════════════════════════════════
# SUMMARY
# ══════════════════════════════════════════════════════════
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
