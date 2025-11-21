-- Migration 001: Identity Layer
-- Core user identity, archetypes, and runtime variables

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users Table
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name VARCHAR(100) NOT NULL,
  avatar_url TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  last_login TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

COMMENT ON TABLE users IS 'Core user accounts';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';

-- User Archetypes Table
CREATE TABLE user_archetypes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  meta_category VARCHAR(50) NOT NULL,
  domain TEXT NOT NULL,
  skill_level VARCHAR(20) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CHECK (meta_category IN ('Digital', 'Economic', 'Aesthetic', 'Biological', 'Cognitive')),
  CHECK (skill_level IN ('novice', 'analyst', 'quant'))
);

CREATE INDEX idx_user_archetypes_user_id ON user_archetypes(user_id);
CREATE INDEX idx_user_archetypes_meta_category ON user_archetypes(meta_category);

COMMENT ON TABLE user_archetypes IS 'User-selected archetypes from onboarding';
COMMENT ON COLUMN user_archetypes.meta_category IS 'One of the 5 universal domains';
COMMENT ON COLUMN user_archetypes.domain IS 'User passion/obsession (e.g., "Stock Trading")';

-- User Variables Table (Key-Value for flexibility)
CREATE TABLE user_variables (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  variable_key VARCHAR(50) NOT NULL,
  variable_value TEXT NOT NULL,
  archetype_id UUID REFERENCES user_archetypes(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(user_id, variable_key)
);

CREATE INDEX idx_user_variables_user_id ON user_variables(user_id);
CREATE INDEX idx_user_variables_archetype_id ON user_variables(archetype_id);
CREATE INDEX idx_user_variables_key ON user_variables(variable_key);

COMMENT ON TABLE user_variables IS 'Runtime variable injection system (ENTITY, STATE, FLOW, LOGIC, INTERFACE)';
COMMENT ON COLUMN user_variables.variable_key IS 'Variable name without braces (e.g., ENTITY, STATE)';
COMMENT ON COLUMN user_variables.variable_value IS 'User-specific value (e.g., "Trading Bot", "Portfolio")';

-- Insert migration record
INSERT INTO schema_migrations (version, description)
VALUES ('001', 'Create identity tables: users, user_archetypes, user_variables');
