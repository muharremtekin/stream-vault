-- StreamVault Database Initialization

-- Create additional databases for Phase 3 services
-- (Tables are managed by each service's own migrations)
CREATE DATABASE streamvault_subscriptions;
CREATE DATABASE streamvault_recommendations;

GRANT ALL PRIVILEGES ON DATABASE streamvault_subscriptions TO streamvault;
GRANT ALL PRIVILEGES ON DATABASE streamvault_recommendations TO streamvault;

-- Tables for streamvault_users are managed by EF Core migrations (User Service)
