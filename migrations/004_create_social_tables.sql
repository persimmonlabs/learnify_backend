-- Migration 004: Social Layer
-- Strava for Brains - Passive social learning network

-- User Relationships (Follow Graph)
CREATE TABLE user_relationships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  follower_id UUID REFERENCES users(id) ON DELETE CASCADE,
  following_id UUID REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(follower_id, following_id),
  CHECK (follower_id != following_id)
);

CREATE INDEX idx_user_relationships_follower ON user_relationships(follower_id);
CREATE INDEX idx_user_relationships_following ON user_relationships(following_id);
CREATE INDEX idx_user_relationships_created_at ON user_relationships(created_at);

COMMENT ON TABLE user_relationships IS 'Social graph for activity feed';

-- Activity Feed (Ticker Data)
CREATE TABLE activity_feed (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  activity_type VARCHAR(50) NOT NULL,
  reference_type VARCHAR(50),
  reference_id UUID,
  metadata JSONB,
  visibility VARCHAR(20) DEFAULT 'friends',
  created_at TIMESTAMP DEFAULT NOW(),
  CHECK (activity_type IN (
    'module_completed',
    'course_completed',
    'exercise_solved',
    'achievement_earned',
    'review_passed',
    'optimization_achieved'
  )),
  CHECK (visibility IN ('public', 'friends', 'private'))
);

CREATE INDEX idx_activity_feed_user_id ON activity_feed(user_id);
CREATE INDEX idx_activity_feed_type ON activity_feed(activity_type);
CREATE INDEX idx_activity_feed_created_at ON activity_feed(created_at DESC);
CREATE INDEX idx_activity_feed_visibility ON activity_feed(visibility);

COMMENT ON TABLE activity_feed IS 'Data-driven activity ticker (no fluff)';
COMMENT ON COLUMN activity_feed.metadata IS 'Additional context (scores, course name, etc.)';

-- Achievements (Badges)
CREATE TABLE achievements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  badge_icon VARCHAR(50),
  criteria JSONB NOT NULL,
  rarity VARCHAR(20) DEFAULT 'common',
  created_at TIMESTAMP DEFAULT NOW(),
  CHECK (rarity IN ('common', 'rare', 'epic', 'legendary'))
);

CREATE INDEX idx_achievements_rarity ON achievements(rarity);

COMMENT ON TABLE achievements IS 'Achievement definitions (e.g., "First Module", "Perfect Score")';
COMMENT ON COLUMN achievements.criteria IS 'Conditions for unlocking';

-- User Achievements
CREATE TABLE user_achievements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  achievement_id UUID REFERENCES achievements(id) ON DELETE CASCADE,
  unlocked_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(user_id, achievement_id)
);

CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);
CREATE INDEX idx_user_achievements_achievement_id ON user_achievements(achievement_id);
CREATE INDEX idx_user_achievements_unlocked_at ON user_achievements(unlocked_at);

COMMENT ON TABLE user_achievements IS 'Earned badges for Living Resume';

-- Insert migration record
INSERT INTO schema_migrations (version, description)
VALUES ('004', 'Create social tables: user_relationships, activity_feed, achievements, user_achievements');
