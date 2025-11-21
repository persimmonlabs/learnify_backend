-- Migration Tracker Table
-- Tracks which migrations have been applied

CREATE TABLE IF NOT EXISTS schema_migrations (
  id SERIAL PRIMARY KEY,
  version VARCHAR(20) NOT NULL UNIQUE,
  description TEXT,
  applied_at TIMESTAMP DEFAULT NOW(),
  checksum VARCHAR(64)
);

CREATE INDEX idx_schema_migrations_version ON schema_migrations(version);

COMMENT ON TABLE schema_migrations IS 'Tracks database migration history';
COMMENT ON COLUMN schema_migrations.version IS 'Migration version number (e.g., 001, 002)';
COMMENT ON COLUMN schema_migrations.checksum IS 'SHA-256 checksum of migration file for validation';
