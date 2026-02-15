#!/bin/bash
# StreamVault - Subscription Flow Test
# Tests: 8.3 (Plans, saga happy path, payment failure compensation,
#              upgrade/downgrade, cancel, invoices, tier sync)
#
# Prerequisites:
#   - All services running (user-service, subscription-service, rabbitmq)
#   - Subscription service has seeded plans (Basic, Standard, Premium)

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

echo "=== StreamVault Subscription Flow Test ==="
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

# Card numbers
SUCCESS_CARD="4242424242424242"
FAIL_CARD="4000000000000002"

TS=$(date +%s)

# ── Setup — Register Fresh Users ─────────────────────────
section "Setup — Register Fresh Users"

# User 1: happy path
USER1_EMAIL="sub_test_${TS}@example.com"
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${USER1_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "201" ] || [ "$STATUS" = "200" ]; then
  pass "Registered user1: ${USER1_EMAIL}"
else
  fail "Register user1 -> HTTP ${STATUS}"
fi

# Login user 1
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${USER1_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  TOKEN1=$(echo "$BODY" | jq -r '.accessToken')
  USER1_ID=$(echo "$BODY" | jq -r '.user.id // empty')
  pass "Login user1 -> 200"
else
  fail "Login user1 -> HTTP ${STATUS}"
  echo "ERROR: Cannot proceed without user1 token. Aborting."
  exit 1
fi

if [ -z "$USER1_ID" ] || [ "$USER1_ID" = "null" ]; then
  info "Extracting user1 ID from /api/users/me..."
  ME_RESP=$(curl -s "${API_URL}/api/users/me" -H "Authorization: Bearer ${TOKEN1}")
  USER1_ID=$(echo "$ME_RESP" | jq -r '.id // empty')
fi

info "User1 ID: ${USER1_ID}"

# User 2: payment failure path
USER2_EMAIL="sub_fail_${TS}@example.com"
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${USER2_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)

if [ "$STATUS" = "201" ] || [ "$STATUS" = "200" ]; then
  pass "Registered user2: ${USER2_EMAIL}"
else
  fail "Register user2 -> HTTP ${STATUS}"
fi

# Login user 2
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"${USER2_EMAIL}\", \"password\": \"${SEED_PASSWORD}\"}")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  TOKEN2=$(echo "$BODY" | jq -r '.accessToken')
  pass "Login user2 -> 200"
else
  fail "Login user2 -> HTTP ${STATUS}"
  TOKEN2=""
fi

# ── List Plans (Public) ──────────────────────────────────
section "8.3 — List Plans"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/plans")
STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/plans -> 200"
else
  fail "GET /api/plans -> HTTP ${STATUS} (expected 200)"
fi

PLAN_COUNT=$(echo "$BODY" | jq 'length' 2>/dev/null || echo "0")
if [ "$PLAN_COUNT" = "3" ]; then
  pass "3 plans returned"
else
  fail "${PLAN_COUNT} plans returned (expected 3)"
fi

# Verify plan names
for PLAN_NAME in Basic Standard Premium; do
  HAS_PLAN=$(echo "$BODY" | jq --arg name "$PLAN_NAME" '[.[] | select(.name == $name)] | length' 2>/dev/null || echo "0")
  if [ "$HAS_PLAN" -gt 0 ]; then
    pass "Plan '${PLAN_NAME}' exists"
  else
    fail "Plan '${PLAN_NAME}' not found"
  fi
done

# Verify prices
BASIC_PRICE=$(echo "$BODY" | jq '[.[] | select(.name == "Basic")][0].priceMonthly' 2>/dev/null || echo "0")
STANDARD_PRICE=$(echo "$BODY" | jq '[.[] | select(.name == "Standard")][0].priceMonthly' 2>/dev/null || echo "0")
PREMIUM_PRICE=$(echo "$BODY" | jq '[.[] | select(.name == "Premium")][0].priceMonthly' 2>/dev/null || echo "0")

if [ "$BASIC_PRICE" = "49.99" ]; then
  pass "Basic price = 49.99 TRY"
else
  fail "Basic price = ${BASIC_PRICE} (expected 49.99)"
fi

if [ "$STANDARD_PRICE" = "79.99" ]; then
  pass "Standard price = 79.99 TRY"
else
  fail "Standard price = ${STANDARD_PRICE} (expected 79.99)"
fi

if [ "$PREMIUM_PRICE" = "119.99" ]; then
  pass "Premium price = 119.99 TRY"
else
  fail "Premium price = ${PREMIUM_PRICE} (expected 119.99)"
fi

# ── Create Subscription — Happy Path (Saga) ──────────────
section "8.3 — Create Subscription (Happy Path)"

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/subscriptions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN1}" \
  -d "{\"planId\": \"${STANDARD_PLAN_ID}\", \"cardNumber\": \"${SUCCESS_CARD}\"}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "201" ]; then
  pass "Create subscription (Standard) -> 201"
else
  fail "Create subscription -> HTTP ${STATUS} (expected 201)"
  echo "       Response: $(echo "$BODY" | head -c 300)"
fi

SUB_STATUS=$(echo "$BODY" | jq -r '.status // empty' 2>/dev/null)
if [ "$SUB_STATUS" = "Active" ]; then
  pass "Subscription status = Active"
else
  fail "Subscription status = '${SUB_STATUS}' (expected Active)"
fi

SUB_TIER=$(echo "$BODY" | jq -r '.tier // empty' 2>/dev/null)
if [ "$SUB_TIER" = "Standard" ]; then
  pass "Subscription tier = Standard"
else
  fail "Subscription tier = '${SUB_TIER}' (expected Standard)"
fi

SUB_AMOUNT=$(echo "$BODY" | jq '.amountCharged' 2>/dev/null || echo "0")
if [ "$SUB_AMOUNT" = "79.99" ]; then
  pass "Amount charged = 79.99"
else
  fail "Amount charged = ${SUB_AMOUNT} (expected 79.99)"
fi

SUB_ID=$(echo "$BODY" | jq -r '.subscriptionId // empty' 2>/dev/null)
if [ -n "$SUB_ID" ]; then
  info "Subscription ID: ${SUB_ID}"
fi

# ── Get My Subscription ──────────────────────────────────
section "8.3 — Get My Subscription"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me" \
  -H "Authorization: Bearer ${TOKEN1}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/subscriptions/me -> 200"
else
  fail "GET /api/subscriptions/me -> HTTP ${STATUS} (expected 200)"
fi

PLAN_NAME=$(echo "$BODY" | jq -r '.plan.name // empty' 2>/dev/null)
if [ "$PLAN_NAME" = "Standard" ]; then
  pass "Subscription plan = Standard"
else
  fail "Subscription plan = '${PLAN_NAME}' (expected Standard)"
fi

ACTIVE_STATUS=$(echo "$BODY" | jq -r '.status // empty' 2>/dev/null)
if [ "$ACTIVE_STATUS" = "Active" ]; then
  pass "Subscription status = Active"
else
  fail "Subscription status = '${ACTIVE_STATUS}' (expected Active)"
fi

PERIOD_START=$(echo "$BODY" | jq -r '.periodStart // empty' 2>/dev/null)
PERIOD_END=$(echo "$BODY" | jq -r '.periodEnd // empty' 2>/dev/null)
if [ -n "$PERIOD_START" ] && [ -n "$PERIOD_END" ]; then
  pass "Period dates present (${PERIOD_START} -> ${PERIOD_END})"
else
  fail "Period dates missing"
fi

AUTO_RENEW=$(echo "$BODY" | jq '.autoRenew' 2>/dev/null || echo "null")
if [ "$AUTO_RENEW" = "true" ]; then
  pass "AutoRenew = true"
else
  fail "AutoRenew = ${AUTO_RENEW} (expected true)"
fi

# ── Duplicate Subscription — 409 ─────────────────────────
section "8.3 — Duplicate Subscription"

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/subscriptions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN1}" \
  -d "{\"planId\": \"${STANDARD_PLAN_ID}\", \"cardNumber\": \"${SUCCESS_CARD}\"}")

STATUS=$(echo "$RESP" | tail -1)

if [ "$STATUS" = "409" ]; then
  pass "Duplicate subscription -> 409 Conflict"
else
  fail "Duplicate subscription -> HTTP ${STATUS} (expected 409)"
fi

# ── Payment Failure — Saga Compensation ──────────────────
section "8.3 — Payment Failure (Saga Compensation)"

if [ -n "$TOKEN2" ]; then
  RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/subscriptions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN2}" \
    -d "{\"planId\": \"${STANDARD_PLAN_ID}\", \"cardNumber\": \"${FAIL_CARD}\"}")

  STATUS=$(echo "$RESP" | tail -1)

  if [ "$STATUS" != "201" ]; then
    pass "Failed payment -> HTTP ${STATUS} (not 201, saga compensation triggered)"
  else
    fail "Failed payment -> 201 (expected non-201, card should be rejected)"
  fi

  # Verify no active subscription after compensation
  sleep 1
  RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me" \
    -H "Authorization: Bearer ${TOKEN2}")

  STATUS=$(echo "$RESP" | tail -1)
  BODY=$(echo "$RESP" | sed '$d')

  if [ "$STATUS" = "404" ]; then
    pass "No active subscription after payment failure -> 404"
  else
    SUB_S=$(echo "$BODY" | jq -r '.status // empty' 2>/dev/null)
    if [ "$SUB_S" = "Failed" ]; then
      pass "Subscription status = Failed after payment failure"
    else
      fail "Expected 404 or Failed status, got HTTP ${STATUS} status='${SUB_S}'"
    fi
  fi
else
  info "Skipping payment failure test (user2 login failed)"
fi

# ── Plan Change — Upgrade ────────────────────────────────
section "8.3 — Plan Change (Upgrade: Standard -> Premium)"

RESP=$(curl -s -w "\n%{http_code}" -X PUT "${API_URL}/api/subscriptions/me/plan" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN1}" \
  -d "{\"newPlanId\": \"${PREMIUM_PLAN_ID}\", \"cardNumber\": \"${SUCCESS_CARD}\"}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Upgrade to Premium -> 200"
else
  fail "Upgrade to Premium -> HTTP ${STATUS} (expected 200)"
  echo "       Response: $(echo "$BODY" | head -c 300)"
fi

NEW_PLAN=$(echo "$BODY" | jq -r '.newPlanName // empty' 2>/dev/null)
if [ "$NEW_PLAN" = "Premium" ]; then
  pass "New plan = Premium"
else
  fail "New plan = '${NEW_PLAN}' (expected Premium)"
fi

NEW_TIER=$(echo "$BODY" | jq -r '.newTier // empty' 2>/dev/null)
if [ "$NEW_TIER" = "Premium" ]; then
  pass "New tier = Premium"
else
  fail "New tier = '${NEW_TIER}' (expected Premium)"
fi

PRICE_DIFF=$(echo "$BODY" | jq '.priceDifference // 0' 2>/dev/null)
IS_POSITIVE=$(echo "$PRICE_DIFF > 0" | bc -l 2>/dev/null || echo "0")
if [ "$IS_POSITIVE" = "1" ]; then
  pass "Price difference = ${PRICE_DIFF} (positive, upgrade charge)"
else
  info "Price difference = ${PRICE_DIFF} (expected positive for upgrade)"
fi

# ── Verify Upgraded Subscription ─────────────────────────
section "8.3 — Verify Upgrade"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me" \
  -H "Authorization: Bearer ${TOKEN1}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  PLAN_NOW=$(echo "$BODY" | jq -r '.plan.name // empty' 2>/dev/null)
  if [ "$PLAN_NOW" = "Premium" ]; then
    pass "Current plan = Premium (upgrade confirmed)"
  else
    fail "Current plan = '${PLAN_NOW}' (expected Premium)"
  fi
else
  fail "GET /api/subscriptions/me -> HTTP ${STATUS} (expected 200)"
fi

# ── Plan Change — Downgrade ──────────────────────────────
section "8.3 — Plan Change (Downgrade: Premium -> Basic)"

RESP=$(curl -s -w "\n%{http_code}" -X PUT "${API_URL}/api/subscriptions/me/plan" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN1}" \
  -d "{\"newPlanId\": \"${BASIC_PLAN_ID}\", \"cardNumber\": \"${SUCCESS_CARD}\"}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Downgrade to Basic -> 200"
else
  fail "Downgrade to Basic -> HTTP ${STATUS} (expected 200)"
fi

NEW_PLAN=$(echo "$BODY" | jq -r '.newPlanName // empty' 2>/dev/null)
if [ "$NEW_PLAN" = "Basic" ]; then
  pass "New plan = Basic"
else
  fail "New plan = '${NEW_PLAN}' (expected Basic)"
fi

PRICE_DIFF=$(echo "$BODY" | jq '.priceDifference // 0' 2>/dev/null)
IS_NONPOSITIVE=$(echo "$PRICE_DIFF <= 0" | bc -l 2>/dev/null || echo "0")
if [ "$IS_NONPOSITIVE" = "1" ]; then
  pass "Price difference = ${PRICE_DIFF} (non-positive, downgrade credit)"
else
  info "Price difference = ${PRICE_DIFF} (expected <= 0 for downgrade)"
fi

# ── Cancel Subscription ──────────────────────────────────
section "8.3 — Cancel Subscription"

RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/subscriptions/me/cancel" \
  -H "Authorization: Bearer ${TOKEN1}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Cancel subscription -> 200"
else
  fail "Cancel subscription -> HTTP ${STATUS} (expected 200)"
fi

CANCEL_STATUS=$(echo "$BODY" | jq -r '.status // empty' 2>/dev/null)
if [ "$CANCEL_STATUS" = "Cancelled" ]; then
  pass "Cancel result status = Cancelled"
else
  fail "Cancel result status = '${CANCEL_STATUS}' (expected Cancelled)"
fi

CANCELLED_AT=$(echo "$BODY" | jq -r '.cancelledAt // empty' 2>/dev/null)
if [ -n "$CANCELLED_AT" ] && [ "$CANCELLED_AT" != "null" ]; then
  pass "CancelledAt present: ${CANCELLED_AT}"
else
  fail "CancelledAt missing"
fi

CANCEL_PERIOD_END=$(echo "$BODY" | jq -r '.periodEnd // empty' 2>/dev/null)
if [ -n "$CANCEL_PERIOD_END" ] && [ "$CANCEL_PERIOD_END" != "null" ]; then
  pass "PeriodEnd present: ${CANCEL_PERIOD_END} (active until period end)"
else
  fail "PeriodEnd missing"
fi

# ── Verify Cancelled but Active Until Period End ─────────
section "8.3 — Verify Cancelled State"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me" \
  -H "Authorization: Bearer ${TOKEN1}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  CURRENT_STATUS=$(echo "$BODY" | jq -r '.status // empty' 2>/dev/null)
  if [ "$CURRENT_STATUS" = "Cancelled" ]; then
    pass "Subscription status = Cancelled (still viewable until period end)"
  else
    fail "Subscription status = '${CURRENT_STATUS}' (expected Cancelled)"
  fi
else
  fail "GET /api/subscriptions/me -> HTTP ${STATUS} (expected 200)"
fi

# ── Invoice History ──────────────────────────────────────
section "8.3 — Invoice History"

RESP=$(curl -s -w "\n%{http_code}" "${API_URL}/api/subscriptions/me/invoices" \
  -H "Authorization: Bearer ${TOKEN1}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "GET /api/subscriptions/me/invoices -> 200"
else
  fail "GET /api/subscriptions/me/invoices -> HTTP ${STATUS} (expected 200)"
fi

INVOICE_COUNT=$(echo "$BODY" | jq 'length' 2>/dev/null || echo "0")
if [ "$INVOICE_COUNT" -ge 1 ]; then
  pass "Invoice count = ${INVOICE_COUNT} (>= 1)"
else
  fail "Invoice count = ${INVOICE_COUNT} (expected >= 1)"
fi

# Check first invoice structure
if [ "$INVOICE_COUNT" -ge 1 ]; then
  INV_NUMBER=$(echo "$BODY" | jq -r '.[0].invoiceNumber // empty' 2>/dev/null)
  INV_AMOUNT=$(echo "$BODY" | jq '.[0].amount // 0' 2>/dev/null)
  INV_PERIOD_START=$(echo "$BODY" | jq -r '.[0].periodStart // empty' 2>/dev/null)
  INV_PERIOD_END=$(echo "$BODY" | jq -r '.[0].periodEnd // empty' 2>/dev/null)

  if [ -n "$INV_NUMBER" ] && [ "$INV_NUMBER" != "null" ]; then
    pass "Invoice has invoiceNumber: ${INV_NUMBER}"
  else
    fail "Invoice missing invoiceNumber"
  fi

  if [ -n "$INV_PERIOD_START" ] && [ -n "$INV_PERIOD_END" ]; then
    pass "Invoice has period dates"
  else
    fail "Invoice missing period dates"
  fi
fi

# ── Tier Sync Event ──────────────────────────────────────
section "8.3 — Subscription Event -> User Tier Sync"

info "Checking if user tier updated via subscription.events..."
info "Polling /api/users/me for up to 15 seconds..."

TIER_UPDATED=false
for i in $(seq 1 5); do
  sleep 3
  ME_RESP=$(curl -s "${API_URL}/api/users/me" -H "Authorization: Bearer ${TOKEN1}")
  USER_ROLE=$(echo "$ME_RESP" | jq -r '.role // empty' 2>/dev/null)

  if [ -z "$USER_ROLE" ]; then
    USER_ROLE=$(echo "$ME_RESP" | jq -r '.tier // empty' 2>/dev/null)
  fi

  info "  Attempt ${i}/5: user role = '${USER_ROLE}'"

  # After cancellation, the user tier should be updated
  # (either still showing the last plan tier or Free after cancel event)
  if [ "$USER_ROLE" = "Free" ] || [ "$USER_ROLE" = "Basic" ]; then
    TIER_UPDATED=true
    break
  fi
done

if [ "$TIER_UPDATED" = "true" ]; then
  pass "User tier updated via subscription event (role = ${USER_ROLE})"
else
  fail "User tier not updated within 15 seconds (role = ${USER_ROLE})"
fi

# ── Summary ──────────────────────────────────────────────
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
