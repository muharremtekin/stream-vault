-- Up
CREATE TABLE IF NOT EXISTS content_similarity (
    content_id_a    VARCHAR(36) NOT NULL,
    content_id_b    VARCHAR(36) NOT NULL,
    similarity_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (content_id_a, content_id_b)
);

CREATE INDEX IF NOT EXISTS idx_content_similarity_a ON content_similarity (content_id_a, similarity_score DESC);
CREATE INDEX IF NOT EXISTS idx_content_similarity_b ON content_similarity (content_id_b, similarity_score DESC);

-- Down
DROP INDEX IF EXISTS idx_content_similarity_b;
DROP INDEX IF EXISTS idx_content_similarity_a;
DROP TABLE IF EXISTS content_similarity;
