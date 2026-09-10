#!/bin/sh
set -eu

# psql performs identifier-safe substitution for :"app_user" in init.sql.
psql \
  --set=ON_ERROR_STOP=1 \
  --set=app_user="$POSTGRES_USER" \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --file /opt/streamvault/init.sql
