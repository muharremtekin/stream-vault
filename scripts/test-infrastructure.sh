#!/bin/bash
# StreamVault - Infrastructure & Event Topology Test
# Tests: 8.4 (Event flow validation) + 8.5 (Consul & infrastructure)
#
# Prerequisites:
#   - All services running via docker compose (with --profile elasticsearch for search)

set -euo pipefail

API_URL="${API_URL:-http://localhost:8081}"
CONSUL_URL="${CONSUL_URL:-http://localhost:8500}"
RABBITMQ_URL="${RABBITMQ_URL:-http://localhost:15672}"
RABBITMQ_USER="${RABBITMQ_USER:-streamvault}"
RABBITMQ_PASS="${RABBITMQ_PASS:-${RABBITMQ_PASSWORD:-password}}"
ES_URL="${ES_URL:-http://localhost:9200}"

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

echo "=== StreamVault Infrastructure & Event Topology Test ==="
echo "    Gateway:      $API_URL"
echo "    Consul:       $CONSUL_URL"
echo "    RabbitMQ:     $RABBITMQ_URL"
echo "    Elasticsearch: $ES_URL"
echo ""

for cmd in curl jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: '$cmd' is required but not installed."
    exit 1
  fi
done

# ── 8.5 Consul Service Health ─────────────────────────────
section "8.5 — Consul Service Health"

SERVICES="user-service catalog-service streaming-service search-service recommendation-service subscription-service"

for SVC in $SERVICES; do
  RESP=$(curl -sf "${CONSUL_URL}/v1/health/service/${SVC}?passing=true" 2>/dev/null || echo "[]")
  COUNT=$(echo "$RESP" | jq 'length' 2>/dev/null || echo "0")

  if [ "$COUNT" -gt 0 ]; then
    pass "${SVC} healthy in Consul (${COUNT} instance(s))"
  else
    fail "${SVC} not healthy in Consul"
  fi
done

# Encoding service is optional (may not always run)
RESP=$(curl -sf "${CONSUL_URL}/v1/health/service/encoding-service?passing=true" 2>/dev/null || echo "[]")
COUNT=$(echo "$RESP" | jq 'length' 2>/dev/null || echo "0")
if [ "$COUNT" -gt 0 ]; then
  pass "encoding-service healthy in Consul (${COUNT} instance(s))"
else
  info "encoding-service not registered in Consul (optional)"
fi

# ── 8.5 Elasticsearch Cluster Health ──────────────────────
section "8.5 — Elasticsearch"

ES_HEALTH_RESP=$(curl -sf "${ES_URL}/_cluster/health" 2>/dev/null || echo "")
if [ -n "$ES_HEALTH_RESP" ]; then
  ES_STATUS=$(echo "$ES_HEALTH_RESP" | jq -r '.status' 2>/dev/null || echo "")
  ES_NODES=$(echo "$ES_HEALTH_RESP" | jq '.number_of_nodes' 2>/dev/null || echo "0")

  if [ "$ES_STATUS" = "green" ] || [ "$ES_STATUS" = "yellow" ]; then
    pass "Elasticsearch cluster health: ${ES_STATUS} (${ES_NODES} node(s))"
  else
    fail "Elasticsearch cluster health: ${ES_STATUS} (expected green or yellow)"
  fi
else
  fail "Elasticsearch not reachable at ${ES_URL}"
fi

# Check index exists
ES_INDEX_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${ES_URL}/streamvault-content" 2>/dev/null || echo "000")
if [ "$ES_INDEX_STATUS" = "200" ]; then
  pass "Elasticsearch index 'streamvault-content' exists"
else
  fail "Elasticsearch index 'streamvault-content' not found (HTTP ${ES_INDEX_STATUS})"
fi

# ── 8.5 PostgreSQL Databases ─────────────────────────────
section "8.5 — PostgreSQL Databases"

for DB in streamvault_subscriptions streamvault_recommendations; do
  if docker compose exec -T postgres psql -U streamvault -d "$DB" -c "SELECT 1" &>/dev/null; then
    pass "PostgreSQL database '${DB}' accessible"
  else
    fail "PostgreSQL database '${DB}' not accessible"
  fi
done

# ── 8.4 RabbitMQ Exchanges ───────────────────────────────
section "8.4 — RabbitMQ Exchanges"

EXCHANGES_RESP=$(curl -sf -u "${RABBITMQ_USER}:${RABBITMQ_PASS}" \
  "${RABBITMQ_URL}/api/exchanges/%2f" 2>/dev/null || echo "[]")

EXPECTED_EXCHANGES="catalog.events watch.events user.events subscription.events encoding encoding.dlx"

for EX in $EXPECTED_EXCHANGES; do
  FOUND=$(echo "$EXCHANGES_RESP" | jq --arg name "$EX" '[.[] | select(.name == $name)] | length' 2>/dev/null || echo "0")
  if [ "$FOUND" -gt 0 ]; then
    pass "Exchange '${EX}' exists"
  else
    fail "Exchange '${EX}' not found"
  fi
done

# ── 8.4 RabbitMQ Queues ──────────────────────────────────
section "8.4 — RabbitMQ Queues"

QUEUES_RESP=$(curl -sf -u "${RABBITMQ_USER}:${RABBITMQ_PASS}" \
  "${RABBITMQ_URL}/api/queues/%2f" 2>/dev/null || echo "[]")

EXPECTED_QUEUES="search.catalog-sync recommendation.catalog search.watch-count recommendation.watch recommendation.ratings user.subscription-sync encoding.jobs encoding.results.catalog encoding.results.streaming encoding.dead-letters"

for Q in $EXPECTED_QUEUES; do
  FOUND=$(echo "$QUEUES_RESP" | jq --arg name "$Q" '[.[] | select(.name == $name)] | length' 2>/dev/null || echo "0")
  if [ "$FOUND" -gt 0 ]; then
    pass "Queue '${Q}' exists"
  else
    fail "Queue '${Q}' not found"
  fi
done

# ── 8.4 RabbitMQ Bindings ────────────────────────────────
section "8.4 — RabbitMQ Bindings"

BINDINGS_RESP=$(curl -sf -u "${RABBITMQ_USER}:${RABBITMQ_PASS}" \
  "${RABBITMQ_URL}/api/bindings/%2f" 2>/dev/null || echo "[]")

# Helper: check binding exists (source -> destination with routing_key)
check_binding() {
  local SRC=$1
  local DST=$2
  local KEY=$3
  local FOUND
  FOUND=$(echo "$BINDINGS_RESP" | jq --arg src "$SRC" --arg dst "$DST" --arg key "$KEY" \
    '[.[] | select(.source == $src and .destination == $dst and .routing_key == $key)] | length' 2>/dev/null || echo "0")
  if [ "$FOUND" -gt 0 ]; then
    pass "Binding: ${SRC} -> ${DST} [${KEY}]"
  else
    fail "Binding missing: ${SRC} -> ${DST} [${KEY}]"
  fi
}

# Catalog events -> Search
check_binding "catalog.events" "search.catalog-sync" "content.created"
check_binding "catalog.events" "search.catalog-sync" "content.updated"
check_binding "catalog.events" "search.catalog-sync" "content.deleted"

# Catalog events -> Recommendation
check_binding "catalog.events" "recommendation.catalog" "content.created"

# Watch events -> Search + Recommendation
check_binding "watch.events" "search.watch-count" "watch.completed"
check_binding "watch.events" "recommendation.watch" "watch.completed"

# User events -> Recommendation
check_binding "user.events" "recommendation.ratings" "content.rated"

# Subscription events -> User Service
check_binding "subscription.events" "user.subscription-sync" "subscription.created"
check_binding "subscription.events" "user.subscription-sync" "subscription.cancelled"
check_binding "subscription.events" "user.subscription-sync" "plan.changed"

# Encoding bindings
check_binding "encoding" "encoding.jobs" "job.new"
check_binding "encoding" "encoding.results.catalog" "job.completed"
check_binding "encoding" "encoding.results.streaming" "job.completed"
check_binding "encoding.dlx" "encoding.dead-letters" "dead-letter"

# ── 8.5 Gateway Health ───────────────────────────────────
section "8.5 — Gateway Health"

GW_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${API_URL}/health" 2>/dev/null || echo "000")
if [ "$GW_STATUS" = "200" ]; then
  pass "Gateway health -> 200"
else
  fail "Gateway health -> HTTP ${GW_STATUS} (expected 200)"
fi

# ── Summary ──────────────────────────────────────────────
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
