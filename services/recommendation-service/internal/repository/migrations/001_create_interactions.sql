-- Up
CREATE TABLE IF NOT EXISTS interactions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         VARCHAR(36) NOT NULL,
    content_id      VARCHAR(36) NOT NULL,
    interaction_type VARCHAR(20) NOT NULL,
    rating          DOUBLE PRECISION,
    completion_pct  DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_interactions_user_id ON interactions (user_id);
CREATE INDEX IF NOT EXISTS idx_interactions_content_id ON interactions (content_id);
CREATE INDEX IF NOT EXISTS idx_interactions_user_content ON interactions (user_id, content_id);
CREATE INDEX IF NOT EXISTS idx_interactions_created_at ON interactions (created_at DESC);

-- Down
DROP INDEX IF EXISTS idx_interactions_created_at;
DROP INDEX IF EXISTS idx_interactions_user_content;
DROP INDEX IF EXISTS idx_interactions_content_id;
DROP INDEX IF EXISTS idx_interactions_user_id;
DROP TABLE IF EXISTS interactions;
