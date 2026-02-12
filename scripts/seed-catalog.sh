#!/bin/bash
# StreamVault - Catalog Seed Data Script
# Loads sample movies and series into the catalog service

set -e

API_URL="${API_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@streamvault.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Admin123!}"

echo "=== StreamVault Catalog Seeder ==="

# Login as admin to get token
echo "[1/3] Logging in as admin..."
TOKEN=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$ADMIN_EMAIL\", \"password\": \"$ADMIN_PASSWORD\"}" \
  | jq -r '.accessToken')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to login as admin. Make sure admin user exists."
  exit 1
fi

echo "   Token acquired."

# Seed movies
echo "[2/3] Seeding movies..."

declare -a MOVIES=(
  '{"title":"Interstellar","description":"Dünya'\''nın geleceği tehlikede. Bir grup kaşif, insanlık için yeni bir yuva bulmak üzere galaksiler arası yolculuğa çıkar.","releaseYear":2014,"durationMinutes":169,"maturityRating":"PG13","genres":["bilim-kurgu","dram","macera"],"director":"Christopher Nolan","cast":[{"name":"Matthew McConaughey","role":"Cooper"}],"thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
  '{"title":"The Dark Knight","description":"Batman, Gotham'\''ın en tehlikeli suçlusu Joker ile yüzleşmek zorunda kalır.","releaseYear":2008,"durationMinutes":152,"maturityRating":"PG13","genres":["aksiyon","dram","suç"],"director":"Christopher Nolan","cast":[{"name":"Christian Bale","role":"Bruce Wayne"}],"thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
  '{"title":"Inception","description":"Bir hırsız, insanların rüyalarına girerek fikirlerini çalma yeteneğine sahiptir.","releaseYear":2010,"durationMinutes":148,"maturityRating":"PG13","genres":["bilim-kurgu","aksiyon","gerilim"],"director":"Christopher Nolan","cast":[{"name":"Leonardo DiCaprio","role":"Dom Cobb"}],"thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
  '{"title":"Pulp Fiction","description":"Los Angeles'\''ın yeraltı dünyasında birbirine bağlanan hikayeler.","releaseYear":1994,"durationMinutes":154,"maturityRating":"R","genres":["suç","dram"],"director":"Quentin Tarantino","cast":[{"name":"John Travolta","role":"Vincent Vega"}],"thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
  '{"title":"The Matrix","description":"Bir bilgisayar programcısı, gerçekliğin aslında makineler tarafından yaratılmış bir simülasyon olduğunu keşfeder.","releaseYear":1999,"durationMinutes":136,"maturityRating":"R","genres":["bilim-kurgu","aksiyon"],"director":"Lana Wachowski","cast":[{"name":"Keanu Reeves","role":"Neo"}],"thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
)

for movie in "${MOVIES[@]}"; do
  TITLE=$(echo "$movie" | jq -r '.title')
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/catalog/movies" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$movie")
  echo "   $TITLE -> HTTP $STATUS"
done

# Seed series
echo "[3/3] Seeding series..."

declare -a SERIES=(
  '{"title":"Breaking Bad","description":"Lise kimya öğretmeni kanser teşhisi sonrası metamfetamin üretmeye başlar.","releaseYear":2008,"maturityRating":"R","genres":["dram","gerilim","suç"],"creator":"Vince Gilligan","thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
  '{"title":"Stranger Things","description":"Küçük bir kasabada kaybolan çocuk, gizli deneyler ve doğaüstü güçleri ortaya çıkarır.","releaseYear":2016,"maturityRating":"PG13","genres":["bilim-kurgu","korku","dram"],"creator":"Duffer Brothers","thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
  '{"title":"The Witcher","description":"Mutasyona uğramış canavar avcısı Geralt, kaderini bulmaya çalışır.","releaseYear":2019,"maturityRating":"R","genres":["fantastik","aksiyon","macera"],"creator":"Lauren Schmidt Hissrich","thumbnailUrl":"https://placeholder.co/300x450","bannerUrl":"https://placeholder.co/1920x600"}'
)

for series in "${SERIES[@]}"; do
  TITLE=$(echo "$series" | jq -r '.title')
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/catalog/series" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$series")
  echo "   $TITLE -> HTTP $STATUS"
done

echo ""
echo "=== Seed completed ==="
