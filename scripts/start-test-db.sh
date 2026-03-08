#!/usr/bin/env bash
# Start a PostgreSQL container for integration testing.
# Usage: ./scripts/start-test-db.sh
#
# Sets GOLLERY_TEST_POSTGRES_DSN for test consumption.
# Run integration tests with:
#   source <(./scripts/start-test-db.sh)
#   cd backend && go test -tags integration ./internal/analytics/postgres/

set -euo pipefail

CONTAINER_NAME="gollery-test-postgres"
PG_PORT="${PG_PORT:-5433}"
PG_USER="gollery_test"
PG_PASS="gollery_test"
PG_DB="gollery_test"

# Stop any existing container.
docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

# Start PostgreSQL.
docker run -d \
  --name "$CONTAINER_NAME" \
  -e POSTGRES_USER="$PG_USER" \
  -e POSTGRES_PASSWORD="$PG_PASS" \
  -e POSTGRES_DB="$PG_DB" \
  -p "${PG_PORT}:5432" \
  postgres:16-alpine \
  >/dev/null

# Wait for PostgreSQL to be ready.
for i in $(seq 1 30); do
  if docker exec "$CONTAINER_NAME" pg_isready -U "$PG_USER" -d "$PG_DB" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

DSN="postgres://${PG_USER}:${PG_PASS}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable"

echo "export GOLLERY_TEST_POSTGRES_DSN='${DSN}'"
echo "# PostgreSQL ready at ${DSN}" >&2
echo "# Stop with: docker rm -f ${CONTAINER_NAME}" >&2
