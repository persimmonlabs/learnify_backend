#!/bin/sh
# Docker entrypoint for Railway deployment
# Runs migrations before starting the application

set -e

echo "🚀 Learnify API Starting..."

# Construct DATABASE_URL if not set (Railway provides individual components)
if [ -z "${DATABASE_URL}" ]; then
  echo "📋 DATABASE_URL not set, constructing from components..."

  # Check if individual components are set
  if [ -z "${DATABASE_HOST}" ] || [ -z "${DATABASE_NAME}" ] || [ -z "${DATABASE_USER}" ] || [ -z "${DATABASE_PASSWORD}" ]; then
    echo "❌ ERROR: Missing database configuration variables"
    echo "   Required: DATABASE_HOST, DATABASE_NAME, DATABASE_USER, DATABASE_PASSWORD"
    exit 1
  fi

  # Default port if not set
  DATABASE_PORT="${DATABASE_PORT:-5432}"

  # Construct PostgreSQL connection URL
  export DATABASE_URL="postgresql://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}"
  echo "✅ DATABASE_URL constructed from components"
else
  echo "✅ DATABASE_URL configured"
fi

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

# Check if schema_migrations table exists and has correct schema
echo "   📋 Checking migration tracking table..."
table_exists=$(psql "${DATABASE_URL}" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations');" 2>/dev/null | tr -d ' ')

if [ "$table_exists" = "t" ]; then
  # Check if table has correct columns (description and checksum)
  has_description=$(psql "${DATABASE_URL}" -t -c "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'schema_migrations' AND column_name = 'description');" 2>/dev/null | tr -d ' ')

  if [ "$has_description" = "f" ]; then
    echo "   ⚠️  schema_migrations has incorrect schema, recreating..."
    # Backup existing migrations first
    psql "${DATABASE_URL}" -t -c "SELECT version FROM schema_migrations;" > /tmp/applied_migrations.txt 2>/dev/null || true
    # Drop and recreate
    psql "${DATABASE_URL}" -c "DROP TABLE IF EXISTS schema_migrations CASCADE;" > /dev/null 2>&1
    psql "${DATABASE_URL}" -f "/root/migrations/000_migration_tracker.sql"
    # Restore migration history
    while read -r version; do
      version=$(echo "$version" | tr -d ' ')
      if [ -n "$version" ]; then
        psql "${DATABASE_URL}" -c "INSERT INTO schema_migrations (version) VALUES ('${version}') ON CONFLICT DO NOTHING;" > /dev/null 2>&1 || true
      fi
    done < /tmp/applied_migrations.txt
    echo "   ✅ Migration tracking table recreated with correct schema"
  else
    echo "   ✅ Migration tracking table has correct schema"
  fi
else
  # Table doesn't exist, create it
  echo "   📋 Initializing migration tracking table..."
  if psql "${DATABASE_URL}" -f "/root/migrations/000_migration_tracker.sql"; then
    echo "   ✅ Migration tracking initialized"
  else
    echo "   ❌ Failed to initialize migration tracking"
    exit 1
  fi
fi

# Run remaining migrations in order, checking if already applied
for migration_file in /root/migrations/*.sql; do
  # Skip if no migration files found
  if [ ! -f "$migration_file" ]; then
    echo "⚠️  No migration files found"
    break
  fi

  # Extract filename and version
  filename=$(basename "$migration_file")
  version="${filename%%_*}"

  # Skip special files
  if [ "$filename" = "README.md" ] || [ "$filename" = "seed_test_data.sql" ] || [ "$filename" = "000_migration_tracker.sql" ]; then
    continue
  fi

  # Check if migration already applied
  already_applied=$(psql "${DATABASE_URL}" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version='${version}';" 2>/dev/null | tr -d ' ')

  if [ "$already_applied" -gt 0 ]; then
    echo "   ⏭️  Migration ${filename} already applied, skipping..."
    continue
  fi

  echo "   📝 Applying migration: ${filename}"

  # Run migration (it will insert its own tracking record)
  if psql "${DATABASE_URL}" -f "$migration_file"; then
    echo "   ✅ Migration ${filename} completed"
  else
    echo "   ❌ Migration ${filename} failed"
    exit 1
  fi
done

echo "🎉 All migrations completed successfully!"

# Start the application
echo "🚀 Starting Learnify API server..."
exec ./main
