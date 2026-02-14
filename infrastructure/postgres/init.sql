-- StreamVault Database Initialization

-- Create additional databases for Phase 3 services
-- (Tables are managed by each service's own migrations)
CREATE DATABASE streamvault_subscriptions;
CREATE DATABASE streamvault_recommendations;

GRANT ALL PRIVILEGES ON DATABASE streamvault_subscriptions TO streamvault;
GRANT ALL PRIVILEGES ON DATABASE streamvault_recommendations TO streamvault;

-- ============================================================
-- streamvault_users tables (connected DB via POSTGRES_DB env)
-- ============================================================

-- User tablosu
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'User',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Profil tablosu
CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(20) NOT NULL,
    icon VARCHAR(20) NOT NULL DEFAULT 'Avatar1',
    is_kids BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);

-- Watchlist tablosu
CREATE TABLE IF NOT EXISTS watchlist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    content_id VARCHAR(50) NOT NULL,
    content_type VARCHAR(10) NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note TEXT,
    CONSTRAINT uq_watchlist_unique UNIQUE (profile_id, content_id)
);

CREATE INDEX IF NOT EXISTS idx_watchlist_profile_id ON watchlist_items(profile_id);

-- Refresh token tablosu
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
