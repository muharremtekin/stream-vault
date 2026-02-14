-- Up
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id           VARCHAR(36) PRIMARY KEY,
    genre_weights     JSONB NOT NULL DEFAULT '{}',
    avg_rating        DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_interactions INTEGER NOT NULL DEFAULT 0,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Down
DROP TABLE IF EXISTS user_profiles;
