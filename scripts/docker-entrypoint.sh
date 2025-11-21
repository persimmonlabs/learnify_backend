#!/bin/sh
# Docker entrypoint for Railway deployment
# Runs migrations before starting the application

set -e

echo "🚀 Learnify API Starting..."

# Check if DATABASE_URL is set
if [ -z "${DATABASE_URL}" ]; then
  echo "❌ ERROR: DATABASE_URL environment variable is not set"
  exit 1
fi

echo "📊 Database URL configured"

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

# Run migrations
echo "🔄 Running database migrations..."

# Create schema_migrations table if it doesn't exist
psql "${DATABASE_URL}" <<EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(255) PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
EOF

# Run migrations in order
for migration_file in /root/migrations/*.sql; do
  # Skip if no migration files found
  if [ ! -f "$migration_file" ]; then
    echo "⚠️  No migration files found"
    break
  fi

  # Extract version from filename (e.g., 001_create_tables.sql -> 001)
  filename=$(basename "$migration_file")

  # Skip README and non-numbered files
  if [ "$filename" = "README.md" ] || [ "$filename" = "seed_test_data.sql" ]; then
    echo "   ⏭️  Skipping $filename"
    continue
  fi

  version="${filename%%_*}"

  # Check if migration already applied
  already_applied=$(psql "${DATABASE_URL}" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version='${version}';" | tr -d ' ')

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

# Start the application
echo "🚀 Starting Learnify API server..."
exec ./main
