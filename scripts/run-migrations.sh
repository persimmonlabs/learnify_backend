#!/bin/sh
# Railway Database Migration Runner
# Automatically runs all pending migrations on deploy

set -e

echo "🚀 Starting database migrations..."

# Wait for database to be ready
echo "⏳ Waiting for database connection..."
MAX_RETRIES=30
RETRY_COUNT=0

until psql "${DATABASE_URL}" -c '\q' 2>/dev/null; do
  RETRY_COUNT=$((RETRY_COUNT + 1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "❌ Database connection failed after ${MAX_RETRIES} attempts"
    exit 1
  fi
  echo "   Attempt ${RETRY_COUNT}/${MAX_RETRIES}: Database not ready, waiting..."
  sleep 2
done

echo "✅ Database connection established"

# Create schema_migrations table if it doesn't exist
echo "📋 Creating migration tracking table..."
psql "${DATABASE_URL}" <<EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(255) PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
EOF

# Run migrations in order
echo "🔄 Running migrations..."

for migration_file in ./migrations/*.sql; do
  if [ ! -f "$migration_file" ]; then
    echo "⚠️  No migration files found"
    continue
  fi

  # Extract version from filename (e.g., 001_create_tables.sql -> 001)
  filename=$(basename "$migration_file")
  version="${filename%%_*}"

  # Check if migration already applied
  already_applied=$(psql "${DATABASE_URL}" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version='${version}';")

  if [ "$already_applied" -gt 0 ]; then
    echo "   ⏭️  Migration ${version} already applied, skipping..."
    continue
  fi

  echo "   📝 Applying migration: ${filename}"

  # Run migration
  if psql "${DATABASE_URL}" -f "$migration_file"; then
    # Record successful migration
    psql "${DATABASE_URL}" -c "INSERT INTO schema_migrations (version) VALUES ('${version}');"
    echo "   ✅ Migration ${version} applied successfully"
  else
    echo "   ❌ Migration ${version} failed"
    exit 1
  fi
done

echo "🎉 All migrations completed successfully!"
exit 0
