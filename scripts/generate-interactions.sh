#!/bin/bash
# StreamVault - Fake Interaction Data Generator
# Generates random watch/rating interactions for testing the recommendation engine.
# Requires: running catalog service, user service, and PostgreSQL.

set -e

API_URL="${API_URL:-http://localhost:8081}"
POSTGRES_URL="${POSTGRES_URL:-postgres://streamvault:password@localhost:5432/streamvault_recommendations}"

echo "=== StreamVault Interaction Generator ==="
echo ""

# ─── Step 1: Login to get tokens and user IDs ───

declare -A USER_IDS
USERS=("premium@streamvault.com" "standard@streamvault.com" "basic@streamvault.com" "free@streamvault.com")
PASSWORD="SeedPass123@"

echo "[1/4] Fetching user IDs..."

for email in "${USERS[@]}"; do
  response=$(curl -sf -X POST "$API_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$email\", \"password\": \"$PASSWORD\"}" 2>/dev/null || true)

  if [ -z "$response" ]; then
    echo "   WARNING: Could not login as $email, skipping"
    continue
  fi

  user_id=$(echo "$response" | jq -r '.user.id // empty')
  if [ -z "$user_id" ]; then
    echo "   WARNING: No user ID for $email, skipping"
    continue
  fi

  USER_IDS["$email"]="$user_id"
  echo "   $email -> $user_id"
done

if [ ${#USER_IDS[@]} -eq 0 ]; then
  echo "ERROR: No users found. Make sure the user service is running with seed data."
  exit 1
fi

# ─── Step 2: Fetch content IDs from catalog ───

echo ""
echo "[2/4] Fetching content from catalog..."

MOVIE_IDS=()
MOVIE_TITLES=()
MOVIE_DATA=$(curl -sf "$API_URL/api/catalog/movies?page=1&pageSize=100" 2>/dev/null || true)
if [ -n "$MOVIE_DATA" ]; then
  while IFS= read -r line; do
    id=$(echo "$line" | jq -r '.id')
    title=$(echo "$line" | jq -r '.title')
    MOVIE_IDS+=("$id")
    MOVIE_TITLES+=("$title")
    echo "   Movie: $title ($id)"
  done < <(echo "$MOVIE_DATA" | jq -c '.items[]? // empty')
fi

SERIES_IDS=()
SERIES_TITLES=()
SERIES_DATA=$(curl -sf "$API_URL/api/catalog/series?page=1&pageSize=100" 2>/dev/null || true)
if [ -n "$SERIES_DATA" ]; then
  while IFS= read -r line; do
    id=$(echo "$line" | jq -r '.id')
    title=$(echo "$line" | jq -r '.title')
    SERIES_IDS+=("$id")
    SERIES_TITLES+=("$title")
    echo "   Series: $title ($id)"
  done < <(echo "$SERIES_DATA" | jq -c '.items[]? // empty')
fi

ALL_IDS=("${MOVIE_IDS[@]}" "${SERIES_IDS[@]}")
ALL_TITLES=("${MOVIE_TITLES[@]}" "${SERIES_TITLES[@]}")

if [ ${#ALL_IDS[@]} -eq 0 ]; then
  echo "ERROR: No content found. Run scripts/seed-catalog.sh first."
  exit 1
fi

echo "   Found ${#ALL_IDS[@]} content items total"

# ─── Step 3: Generate interactions via psql ───

echo ""
echo "[3/4] Generating interactions..."

# Build SQL for all interactions
SQL=""
TOTAL=0

# Helper to add an interaction
add_interaction() {
  local user_id="$1"
  local content_id="$2"
  local type="$3"
  local rating="$4"
  local completion="$5"
  local days_ago="$6"

  if [ "$rating" = "null" ]; then
    rating_val="NULL"
  else
    rating_val="$rating"
  fi

  SQL+="INSERT INTO interactions (user_id, content_id, interaction_type, rating, completion_pct, created_at) VALUES ('$user_id', '$content_id', '$type', $rating_val, $completion, NOW() - INTERVAL '$days_ago days') ON CONFLICT DO NOTHING;
"
  TOTAL=$((TOTAL + 1))
}

# ── User 1 (premium): Sci-fi lover ──
# Watches and rates sci-fi highly: Interstellar, Inception, The Matrix
USER1="${USER_IDS[premium@streamvault.com]}"
if [ -n "$USER1" ]; then
  echo "   User 1 (premium): Sci-fi enthusiast"
  for i in "${!ALL_IDS[@]}"; do
    title="${ALL_TITLES[$i]}"
    cid="${ALL_IDS[$i]}"
    case "$title" in
      "Interstellar")
        add_interaction "$USER1" "$cid" "complete" "null" 100 14
        add_interaction "$USER1" "$cid" "rate" 9.5 100 13
        ;;
      "Inception")
        add_interaction "$USER1" "$cid" "complete" "null" 100 10
        add_interaction "$USER1" "$cid" "rate" 9.0 100 9
        ;;
      "The Matrix")
        add_interaction "$USER1" "$cid" "complete" "null" 100 7
        add_interaction "$USER1" "$cid" "rate" 8.5 100 6
        ;;
      "Stranger Things")
        add_interaction "$USER1" "$cid" "view" "null" 65 5
        add_interaction "$USER1" "$cid" "rate" 7.0 65 4
        ;;
      "The Dark Knight")
        add_interaction "$USER1" "$cid" "view" "null" 30 3
        ;;
    esac
  done
fi

# ── User 2 (standard): Crime/drama fan ──
USER2="${USER_IDS[standard@streamvault.com]}"
if [ -n "$USER2" ]; then
  echo "   User 2 (standard): Crime/drama fan"
  for i in "${!ALL_IDS[@]}"; do
    title="${ALL_TITLES[$i]}"
    cid="${ALL_IDS[$i]}"
    case "$title" in
      "The Dark Knight")
        add_interaction "$USER2" "$cid" "complete" "null" 100 12
        add_interaction "$USER2" "$cid" "rate" 9.0 100 11
        ;;
      "Pulp Fiction")
        add_interaction "$USER2" "$cid" "complete" "null" 100 10
        add_interaction "$USER2" "$cid" "rate" 9.5 100 9
        ;;
      "Breaking Bad")
        add_interaction "$USER2" "$cid" "complete" "null" 100 8
        add_interaction "$USER2" "$cid" "rate" 10.0 100 7
        ;;
      "Interstellar")
        add_interaction "$USER2" "$cid" "view" "null" 45 6
        add_interaction "$USER2" "$cid" "rate" 6.0 45 5
        ;;
      "The Witcher")
        add_interaction "$USER2" "$cid" "view" "null" 70 3
        add_interaction "$USER2" "$cid" "rate" 7.5 70 2
        ;;
    esac
  done
fi

# ── User 3 (basic): Mixed viewer ──
USER3="${USER_IDS[basic@streamvault.com]}"
if [ -n "$USER3" ]; then
  echo "   User 3 (basic): Mixed viewer"
  for i in "${!ALL_IDS[@]}"; do
    title="${ALL_TITLES[$i]}"
    cid="${ALL_IDS[$i]}"
    case "$title" in
      "Interstellar")
        add_interaction "$USER3" "$cid" "complete" "null" 100 15
        add_interaction "$USER3" "$cid" "rate" 7.0 100 14
        ;;
      "The Dark Knight")
        add_interaction "$USER3" "$cid" "view" "null" 80 12
        add_interaction "$USER3" "$cid" "rate" 7.5 80 11
        ;;
      "Breaking Bad")
        add_interaction "$USER3" "$cid" "view" "null" 55 10
        add_interaction "$USER3" "$cid" "rate" 6.5 55 9
        ;;
      "Stranger Things")
        add_interaction "$USER3" "$cid" "complete" "null" 100 7
        add_interaction "$USER3" "$cid" "rate" 8.0 100 6
        ;;
      "The Matrix")
        add_interaction "$USER3" "$cid" "view" "null" 40 4
        ;;
      "Pulp Fiction")
        add_interaction "$USER3" "$cid" "view" "null" 60 2
        add_interaction "$USER3" "$cid" "rate" 7.0 60 1
        ;;
      "Inception")
        add_interaction "$USER3" "$cid" "complete" "null" 100 20
        ;;
      "The Witcher")
        add_interaction "$USER3" "$cid" "view" "null" 25 1
        ;;
    esac
  done
fi

# ── User 4 (free): New user — cold start test ──
USER4="${USER_IDS[free@streamvault.com]}"
if [ -n "$USER4" ]; then
  echo "   User 4 (free): New user (cold start)"
  for i in "${!ALL_IDS[@]}"; do
    title="${ALL_TITLES[$i]}"
    cid="${ALL_IDS[$i]}"
    case "$title" in
      "Inception")
        add_interaction "$USER4" "$cid" "view" "null" 50 2
        ;;
      "Stranger Things")
        add_interaction "$USER4" "$cid" "view" "null" 20 1
        ;;
    esac
  done
fi

if [ -z "$SQL" ]; then
  echo "   WARNING: No interactions generated (missing user IDs or content)"
  exit 0
fi

# Execute via psql in docker or direct
echo "   Inserting $TOTAL interactions..."

if command -v psql &>/dev/null; then
  echo "$SQL" | psql "$POSTGRES_URL" -q 2>/dev/null
elif docker compose ps postgres --status running -q &>/dev/null 2>&1; then
  echo "$SQL" | docker compose exec -T postgres psql -U streamvault -d streamvault_recommendations -q 2>/dev/null
else
  echo "   Trying docker exec..."
  CONTAINER=$(docker ps --filter "name=postgres" --format '{{.Names}}' | head -1)
  if [ -n "$CONTAINER" ]; then
    echo "$SQL" | docker exec -i "$CONTAINER" psql -U streamvault -d streamvault_recommendations -q 2>/dev/null
  else
    echo "ERROR: Cannot connect to PostgreSQL. Install psql or run via Docker."
    exit 1
  fi
fi

echo "   Done! $TOTAL interactions inserted."

# ─── Step 4: Seed content features ───

echo ""
echo "[4/4] Seeding content features..."

# Insert content features from catalog data so the engine can compute vectors
CF_SQL=""

for i in "${!MOVIE_IDS[@]}"; do
  cid="${MOVIE_IDS[$i]}"
  title="${MOVIE_TITLES[$i]}"

  # Extract genres from the movie data
  genres=$(echo "$MOVIE_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .genres | @csv' 2>/dev/null | sed "s/\"//g")
  director=$(echo "$MOVIE_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .director // ""' 2>/dev/null)
  year=$(echo "$MOVIE_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .releaseYear // 2000' 2>/dev/null)
  rating=$(echo "$MOVIE_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .averageRating // 0' 2>/dev/null)

  # Convert CSV genres to PostgreSQL array
  pg_genres=$(echo "$genres" | awk -F',' '{for(i=1;i<=NF;i++) printf "\"%s\"%s", $i, (i<NF?",":""); print ""}' | sed 's/^/{/;s/$/}/')

  CF_SQL+="INSERT INTO content_features (content_id, title, genres, tags, director, release_year, avg_rating, feature_vector, updated_at)
VALUES ('$cid', '$(echo "$title" | sed "s/'/''/g")', '$pg_genres', '{}', '$(echo "$director" | sed "s/'/''/g")', $year, $rating, '{}', NOW())
ON CONFLICT (content_id) DO UPDATE SET title=EXCLUDED.title, genres=EXCLUDED.genres, director=EXCLUDED.director, release_year=EXCLUDED.release_year, avg_rating=EXCLUDED.avg_rating, updated_at=NOW();
"
done

for i in "${!SERIES_IDS[@]}"; do
  cid="${SERIES_IDS[$i]}"
  title="${SERIES_TITLES[$i]}"

  genres=$(echo "$SERIES_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .genres | @csv' 2>/dev/null | sed "s/\"//g")
  creator=$(echo "$SERIES_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .creator // ""' 2>/dev/null)
  year=$(echo "$SERIES_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .releaseYear // 2000' 2>/dev/null)
  rating=$(echo "$SERIES_DATA" | jq -r --arg t "$title" '.items[] | select(.title == $t) | .averageRating // 0' 2>/dev/null)

  pg_genres=$(echo "$genres" | awk -F',' '{for(i=1;i<=NF;i++) printf "\"%s\"%s", $i, (i<NF?",":""); print ""}' | sed 's/^/{/;s/$/}/')

  CF_SQL+="INSERT INTO content_features (content_id, title, genres, tags, director, release_year, avg_rating, feature_vector, updated_at)
VALUES ('$cid', '$(echo "$title" | sed "s/'/''/g")', '$pg_genres', '{}', '$(echo "$creator" | sed "s/'/''/g")', $year, $rating, '{}', NOW())
ON CONFLICT (content_id) DO UPDATE SET title=EXCLUDED.title, genres=EXCLUDED.genres, director=EXCLUDED.director, release_year=EXCLUDED.release_year, avg_rating=EXCLUDED.avg_rating, updated_at=NOW();
"
done

if [ -n "$CF_SQL" ]; then
  if command -v psql &>/dev/null; then
    echo "$CF_SQL" | psql "$POSTGRES_URL" -q 2>/dev/null
  elif docker compose ps postgres --status running -q &>/dev/null 2>&1; then
    echo "$CF_SQL" | docker compose exec -T postgres psql -U streamvault -d streamvault_recommendations -q 2>/dev/null
  else
    CONTAINER=$(docker ps --filter "name=postgres" --format '{{.Names}}' | head -1)
    if [ -n "$CONTAINER" ]; then
      echo "$CF_SQL" | docker exec -i "$CONTAINER" psql -U streamvault -d streamvault_recommendations -q 2>/dev/null
    fi
  fi
  echo "   Content features seeded for ${#ALL_IDS[@]} items."
fi

echo ""
echo "=== Generation complete ==="
echo ""
echo "Summary:"
echo "  Users with interactions: ${#USER_IDS[@]}"
echo "  Content items: ${#ALL_IDS[@]}"
echo "  Total interactions: $TOTAL"
echo ""
echo "Test recommendations:"
echo "  curl -s http://localhost:5006/api/recommendations -H 'X-User-Id: ${USER_IDS[premium@streamvault.com]:-<user-id>}' | jq"
echo "  curl -s http://localhost:5006/api/recommendations/home -H 'X-User-Id: ${USER_IDS[premium@streamvault.com]:-<user-id>}' | jq"
