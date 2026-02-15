#!/bin/bash
# StreamVault - Search Flow Test
# Tests: 8.1 (Full-text search, autocomplete, facets, trending, CQRS sync, cache)
#
# Prerequisites:
#   - All services running via docker compose --profile elasticsearch
#   - Catalog seeded (seed-catalog.sh)
#   - Search index seeded (seed-search-index.sh)

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

echo "=== StreamVault Search Flow Test ==="
echo "    Gateway: $API_URL"
echo ""

for cmd in curl jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: '$cmd' is required but not installed."
    exit 1
  fi
done

# ── Setup — Login ────────────────────────────────────────
section "Setup — Login"

TOKEN=$(login_as "admin@streamvault.com" "Admin")

if [ -z "$TOKEN" ]; then
  echo "ERROR: Admin login failed. Aborting."
  exit 1
fi

# ── Full-Text Search ─────────────────────────────────────
section "8.1 — Full-Text Search"

# Search for a known title
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search?q=Interstellar" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Search 'Interstellar' -> 200"
else
  fail "Search 'Interstellar' -> HTTP ${STATUS} (expected 200)"
fi

TOTAL=$(echo "$BODY" | jq '.totalCount' 2>/dev/null || echo "0")
if [ "$TOTAL" -ge 1 ]; then
  pass "Search 'Interstellar' returned ${TOTAL} result(s)"
else
  fail "Search 'Interstellar' returned 0 results (expected >= 1)"
fi

# Check highlights
HAS_HIGHLIGHT=$(echo "$BODY" | jq '[.items[] | select(.highlights != null and .highlights.title != null)] | length' 2>/dev/null || echo "0")
if [ "$HAS_HIGHLIGHT" -ge 1 ]; then
  pass "Search results contain highlights"
else
  fail "Search results missing highlights"
fi

HAS_EM=$(echo "$BODY" | jq -r '[.items[].highlights.title // empty] | .[0]' 2>/dev/null || echo "")
if echo "$HAS_EM" | grep -q "<em>"; then
  pass "Highlights contain <em> tags"
else
  fail "Highlights missing <em> tags"
fi

# Search for nonexistent content
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search?q=zzzyyyxxx99988877" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Search gibberish -> 200"
else
  fail "Search gibberish -> HTTP ${STATUS} (expected 200)"
fi

TOTAL=$(echo "$BODY" | jq '.totalCount' 2>/dev/null || echo "-1")
if [ "$TOTAL" = "0" ]; then
  pass "Search gibberish returned 0 results"
else
  fail "Search gibberish returned ${TOTAL} results (expected 0)"
fi

# ── Autocomplete ─────────────────────────────────────────
section "8.1 — Autocomplete"

# Valid autocomplete (2+ chars)
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search/autocomplete?q=In" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Autocomplete 'In' -> 200"
else
  fail "Autocomplete 'In' -> HTTP ${STATUS} (expected 200)"
fi

SUGGESTION_COUNT=$(echo "$BODY" | jq '.suggestions | length' 2>/dev/null || echo "0")
if [ "$SUGGESTION_COUNT" -ge 1 ]; then
  pass "Autocomplete 'In' returned ${SUGGESTION_COUNT} suggestion(s)"
else
  fail "Autocomplete 'In' returned 0 suggestions (expected >= 1)"
fi

# Check suggestion structure
FIRST_TITLE=$(echo "$BODY" | jq -r '.suggestions[0].title // empty' 2>/dev/null)
FIRST_ID=$(echo "$BODY" | jq -r '.suggestions[0].contentId // empty' 2>/dev/null)
if [ -n "$FIRST_TITLE" ] && [ -n "$FIRST_ID" ]; then
  pass "Autocomplete suggestion has title='${FIRST_TITLE}' and contentId"
else
  fail "Autocomplete suggestion missing title or contentId"
fi

# Below minimum (1 char) -> 400
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search/autocomplete?q=x" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)

if [ "$STATUS" = "400" ]; then
  pass "Autocomplete 1 char -> 400 (min 2 chars)"
else
  fail "Autocomplete 1 char -> HTTP ${STATUS} (expected 400)"
fi

# Autocomplete for another known title
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search/autocomplete?q=Dark" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  HAS_DARK=$(echo "$BODY" | jq '[.suggestions[] | select(.title | test("Dark";"i"))] | length' 2>/dev/null || echo "0")
  if [ "$HAS_DARK" -ge 1 ]; then
    pass "Autocomplete 'Dark' contains matching suggestion"
  else
    fail "Autocomplete 'Dark' has no suggestion containing 'Dark'"
  fi
else
  fail "Autocomplete 'Dark' -> HTTP ${STATUS} (expected 200)"
fi

# ── Faceted Search & Filtering ───────────────────────────
section "8.1 — Faceted Search & Filtering"

# Content type filter
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search?q=&content_type=movie" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Search content_type=movie -> 200"
  TOTAL=$(echo "$BODY" | jq '.totalCount' 2>/dev/null || echo "0")
  if [ "$TOTAL" -ge 1 ]; then
    pass "Movie filter returned ${TOTAL} result(s)"
    ALL_MOVIES=$(echo "$BODY" | jq '[.items[] | select(.contentType != "movie")] | length' 2>/dev/null || echo "0")
    if [ "$ALL_MOVIES" = "0" ]; then
      pass "All results are movies"
    else
      fail "Found ${ALL_MOVIES} non-movie result(s) with movie filter"
    fi
  else
    fail "Movie filter returned 0 results"
  fi
else
  fail "Search content_type=movie -> HTTP ${STATUS} (expected 200)"
fi

# Year range filter
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search?q=&year_from=2010&year_to=2020" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Search year_from=2010&year_to=2020 -> 200"
  OUT_OF_RANGE=$(echo "$BODY" | jq '[.items[] | select(.releaseYear < 2010 or .releaseYear > 2020)] | length' 2>/dev/null || echo "0")
  if [ "$OUT_OF_RANGE" = "0" ]; then
    pass "All results within year range 2010-2020"
  else
    fail "Found ${OUT_OF_RANGE} result(s) outside year range"
  fi
else
  fail "Search year range -> HTTP ${STATUS} (expected 200)"
fi

# Sort by rating descending
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search?q=&sort=rating&order=desc&pageSize=10" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Search sort=rating&order=desc -> 200"
  ITEM_COUNT=$(echo "$BODY" | jq '.items | length' 2>/dev/null || echo "0")
  if [ "$ITEM_COUNT" -ge 2 ]; then
    FIRST_RATING=$(echo "$BODY" | jq '.items[0].averageRating' 2>/dev/null || echo "0")
    SECOND_RATING=$(echo "$BODY" | jq '.items[1].averageRating' 2>/dev/null || echo "0")
    SORTED=$(echo "$FIRST_RATING >= $SECOND_RATING" | bc -l 2>/dev/null || echo "1")
    if [ "$SORTED" = "1" ]; then
      pass "Results sorted by rating descending (${FIRST_RATING} >= ${SECOND_RATING})"
    else
      fail "Results not sorted by rating (${FIRST_RATING} < ${SECOND_RATING})"
    fi
  else
    info "Only ${ITEM_COUNT} result(s), skipping sort verification"
  fi
else
  fail "Search sort by rating -> HTTP ${STATUS} (expected 200)"
fi

# Check facets
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search?q=" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  GENRE_FACETS=$(echo "$BODY" | jq '.facets.genreFacets | length' 2>/dev/null || echo "0")
  CT_FACETS=$(echo "$BODY" | jq '.facets.contentTypeFacets | length' 2>/dev/null || echo "0")

  if [ "$GENRE_FACETS" -gt 0 ]; then
    pass "Genre facets present (${GENRE_FACETS} bucket(s))"
  else
    fail "Genre facets empty"
  fi

  if [ "$CT_FACETS" -gt 0 ]; then
    pass "Content type facets present (${CT_FACETS} bucket(s))"
  else
    fail "Content type facets empty"
  fi
else
  fail "Search with facets -> HTTP ${STATUS} (expected 200)"
fi

# ── Trending ─────────────────────────────────────────────
section "8.1 — Trending Content"

# Trending is public (no auth needed)
RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search/trending?window=week")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Trending (week) -> 200"
  IS_ARRAY=$(echo "$BODY" | jq '.items | type' 2>/dev/null || echo "")
  if [ "$IS_ARRAY" = '"array"' ]; then
    TREND_COUNT=$(echo "$BODY" | jq '.items | length' 2>/dev/null || echo "0")
    pass "Trending items is array (${TREND_COUNT} item(s))"
  else
    fail "Trending items is not an array"
  fi
else
  fail "Trending (week) -> HTTP ${STATUS} (expected 200)"
fi

RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search/trending?window=day")

STATUS=$(echo "$RESP" | tail -1)

if [ "$STATUS" = "200" ]; then
  pass "Trending (day) -> 200"
else
  fail "Trending (day) -> HTTP ${STATUS} (expected 200)"
fi

# ── Pagination ───────────────────────────────────────────
section "8.1 — Pagination"

RESP=$(curl -s -w "\n%{http_code}" \
  "${API_URL}/api/search?q=&page=1&pageSize=2" \
  -H "Authorization: Bearer ${TOKEN}")

STATUS=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" = "200" ]; then
  pass "Paginated search -> 200"

  PAGE=$(echo "$BODY" | jq '.page' 2>/dev/null || echo "0")
  PAGE_SIZE=$(echo "$BODY" | jq '.pageSize' 2>/dev/null || echo "0")
  TOTAL_PAGES=$(echo "$BODY" | jq '.totalPages' 2>/dev/null || echo "0")
  ITEM_COUNT=$(echo "$BODY" | jq '.items | length' 2>/dev/null || echo "0")

  if [ "$PAGE" = "1" ]; then
    pass "Page = 1"
  else
    fail "Page = ${PAGE} (expected 1)"
  fi

  if [ "$PAGE_SIZE" = "2" ]; then
    pass "PageSize = 2"
  else
    fail "PageSize = ${PAGE_SIZE} (expected 2)"
  fi

  if [ "$ITEM_COUNT" -le 2 ]; then
    pass "Items count = ${ITEM_COUNT} (<= pageSize 2)"
  else
    fail "Items count = ${ITEM_COUNT} (expected <= 2)"
  fi

  if [ "$TOTAL_PAGES" -ge 1 ]; then
    pass "TotalPages = ${TOTAL_PAGES} (>= 1)"
  else
    fail "TotalPages = ${TOTAL_PAGES} (expected >= 1)"
  fi
else
  fail "Paginated search -> HTTP ${STATUS} (expected 200)"
fi

# ── CQRS — Catalog to ES Sync ───────────────────────────
section "8.1 — CQRS Sync (Catalog -> Elasticsearch)"

UNIQUE_TITLE="TestMovie_CQRS_$(date +%s)"

info "Creating movie '${UNIQUE_TITLE}' in catalog..."

CREATE_RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_URL}/api/catalog/movies" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"title\":\"${UNIQUE_TITLE}\",\"description\":\"CQRS test movie for search sync validation.\",\"releaseYear\":2025,\"durationMinutes\":120,\"maturityRating\":2,\"genres\":[\"bilim-kurgu\"],\"director\":\"Test Director\",\"cast\":[{\"name\":\"Test Actor\",\"role\":\"Lead\"}],\"thumbnailUrl\":\"https://placeholder.co/300x450\",\"bannerUrl\":\"https://placeholder.co/1920x600\"}")

CREATE_STATUS=$(echo "$CREATE_RESP" | tail -1)
CREATE_BODY=$(echo "$CREATE_RESP" | sed '$d')

if [ "$CREATE_STATUS" = "201" ]; then
  pass "Created test movie -> 201"
  MOVIE_ID=$(echo "$CREATE_BODY" | jq -r '.id // empty' 2>/dev/null)
  if [ -n "$MOVIE_ID" ]; then
    info "Movie ID: ${MOVIE_ID}"
  fi
else
  fail "Create test movie -> HTTP ${CREATE_STATUS} (expected 201)"
  info "Skipping CQRS sync verification"
fi

if [ "$CREATE_STATUS" = "201" ]; then
  info "Waiting for outbox sync to Elasticsearch (polling up to 20s)..."
  FOUND=false
  for i in $(seq 1 10); do
    sleep 2
    SEARCH_RESP=$(curl -s "${API_URL}/api/search?q=${UNIQUE_TITLE}" \
      -H "Authorization: Bearer ${TOKEN}")
    SEARCH_TOTAL=$(echo "$SEARCH_RESP" | jq '.totalCount' 2>/dev/null || echo "0")
    if [ "$SEARCH_TOTAL" -ge 1 ]; then
      FOUND=true
      break
    fi
    info "  Attempt ${i}/10: not yet in search index..."
  done

  if [ "$FOUND" = "true" ]; then
    pass "CQRS sync: movie appeared in search index"
  else
    fail "CQRS sync: movie did not appear in search index within 20s"
  fi
fi

# ── Cache Verification ───────────────────────────────────
section "8.1 — Search Cache"

RESP1=$(curl -s "${API_URL}/api/search?q=Interstellar" \
  -H "Authorization: Bearer ${TOKEN}")
TOTAL1=$(echo "$RESP1" | jq '.totalCount' 2>/dev/null || echo "-1")

RESP2=$(curl -s "${API_URL}/api/search?q=Interstellar" \
  -H "Authorization: Bearer ${TOKEN}")
TOTAL2=$(echo "$RESP2" | jq '.totalCount' 2>/dev/null || echo "-2")

if [ "$TOTAL1" = "$TOTAL2" ] && [ "$TOTAL1" != "-1" ]; then
  pass "Cache: consecutive searches return same totalCount (${TOTAL1})"
else
  fail "Cache: totalCount mismatch (${TOTAL1} vs ${TOTAL2})"
fi

# ── Summary ──────────────────────────────────────────────
echo ""
echo "================================"
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"
echo "================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
