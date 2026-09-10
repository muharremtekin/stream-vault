-- StreamVault Database Initialization

-- Create additional databases for Phase 3 services
-- (Tables are managed by each service's own migrations)
CREATE DATABASE streamvault_subscriptions;
CREATE DATABASE streamvault_recommendations;

-- app_user is supplied by init.sh as a psql variable. The :"..." form quotes it
-- as an SQL identifier, so custom POSTGRES_USER values are never concatenated.
GRANT ALL PRIVILEGES ON DATABASE streamvault_subscriptions TO :"app_user";
GRANT ALL PRIVILEGES ON DATABASE streamvault_recommendations TO :"app_user";

-- Tables for streamvault_users are managed by EF Core migrations (User Service)
