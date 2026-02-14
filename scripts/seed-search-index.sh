#!/bin/sh
set -e

CATALOG_URL="${CATALOG_URL:-http://localhost:5002}"
ES_URL="${ES_URL:-http://localhost:9200}"
INDEX_NAME="${INDEX_NAME:-streamvault-content}"
PAGE_SIZE=100

echo "Seeding search index from catalog service..."
echo "Catalog URL: ${CATALOG_URL}"
echo "ES URL: ${ES_URL}"
echo "Index: ${INDEX_NAME}"

seed_content() {
    content_type="$1"
    endpoint="$2"

    page=1
    total_indexed=0

    while true; do
        response=$(curl -sf "${CATALOG_URL}${endpoint}?page=${page}&pageSize=${PAGE_SIZE}" 2>/dev/null) || {
            echo "Failed to fetch ${content_type} page ${page}"
            break
        }

        items=$(echo "$response" | jq -r '.items // empty')

        if [ -z "$items" ] || [ "$items" = "[]" ] || [ "$items" = "null" ]; then
            break
        fi

        count=$(echo "$items" | jq 'length')

        bulk_data=$(echo "$items" | jq -c --arg idx "$INDEX_NAME" --arg ct "$content_type" '.[] |
            {"index": {"_index": $idx, "_id": .id}},
            {
                id: .id,
                title: .title,
                original_title: (.originalTitle // ""),
                description: (.description // ""),
                content_type: $ct,
                genres: (.genres // []),
                tags: (.tags // []),
                cast_names: (if .cast then [.cast[].name] else [] end),
                director: (.director // ""),
                release_year: (.releaseYear // 0),
                maturity_rating: (.maturityRating // ""),
                average_rating: (.averageRating // 0),
                rating_count: (.ratingCount // 0),
                view_count: 0,
                video_status: (.videoStatus // ""),
                thumbnail_url: (.thumbnailUrl // ""),
                banner_url: (.bannerUrl // ""),
                duration_seconds: (if $ct == "movie" then ((.durationMinutes // 0) * 60) else 0 end),
                season_count: (.seasonCount // 0),
                episode_count: (.episodeCount // 0),
                created_at: (.createdAt // "2026-01-01T00:00:00Z"),
                updated_at: (.updatedAt // "2026-01-01T00:00:00Z")
            }')

        echo "${bulk_data}" | curl -sf -X POST "${ES_URL}/_bulk" \
            -H "Content-Type: application/x-ndjson" \
            --data-binary @- > /dev/null

        total_indexed=$((total_indexed + count))
        echo "Indexed ${count} ${content_type}(s) from page ${page} (total: ${total_indexed})"

        total_pages=$(echo "$response" | jq -r '.totalPages // 1')
        if [ "$page" -ge "$total_pages" ]; then
            break
        fi
        page=$((page + 1))
    done

    echo "Finished indexing ${content_type}s: ${total_indexed} total"
}

seed_content "movie" "/api/catalog/movies"
seed_content "series" "/api/catalog/series"

echo ""
echo "Refreshing index..."
curl -sf -X POST "${ES_URL}/${INDEX_NAME}/_refresh" > /dev/null
echo "Done. Search index seeded successfully."
