-- Up
CREATE TABLE IF NOT EXISTS content_features (
    content_id      VARCHAR(36) PRIMARY KEY,
    title           VARCHAR(500) NOT NULL DEFAULT '',
    genres          TEXT[] NOT NULL DEFAULT '{}',
    tags            TEXT[] NOT NULL DEFAULT '{}',
    director        VARCHAR(200) NOT NULL DEFAULT '',
    release_year    INTEGER NOT NULL DEFAULT 0,
    avg_rating      DOUBLE PRECISION NOT NULL DEFAULT 0,
    feature_vector  JSONB NOT NULL DEFAULT '{}',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Down
DROP TABLE IF EXISTS content_features;
