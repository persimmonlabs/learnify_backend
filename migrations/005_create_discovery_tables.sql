-- Migration 005: Discovery Layer
-- Netflix-style recommendations and trending courses

-- Recommendations (Cached Suggestions)
CREATE TABLE recommendations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  course_id UUID REFERENCES generated_courses(id) ON DELETE CASCADE,
  recommendation_type VARCHAR(50) NOT NULL,
  match_score INT,
  reason TEXT,
  metadata JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  expires_at TIMESTAMP,
  CHECK (recommendation_type IN (
    'collaborative_filtering',
    'skill_adjacency',
    'social_signal',
    'trending'
  )),
  CHECK (match_score BETWEEN 0 AND 100)
);

CREATE INDEX idx_recommendations_user_id ON recommendations(user_id);
CREATE INDEX idx_recommendations_course_id ON recommendations(course_id);
CREATE INDEX idx_recommendations_type ON recommendations(recommendation_type);
CREATE INDEX idx_recommendations_match_score ON recommendations(match_score DESC);
CREATE INDEX idx_recommendations_expires_at ON recommendations(expires_at);

COMMENT ON TABLE recommendations IS 'Pre-computed course recommendations';
COMMENT ON COLUMN recommendations.reason IS 'Display reason (e.g., "Because you mastered X")';
COMMENT ON COLUMN recommendations.metadata IS 'Friend avatars, trending data, etc.';

-- Trending Courses (Velocity Cache)
CREATE TABLE trending_courses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id UUID REFERENCES generated_courses(id) ON DELETE CASCADE,
  velocity DECIMAL(10, 2),
  signups_24h INT,
  signups_previous_24h INT,
  rank INT,
  meta_category VARCHAR(50),
  calculated_at TIMESTAMP DEFAULT NOW(),
  CHECK (velocity >= 0)
);

CREATE UNIQUE INDEX idx_trending_courses_course_id ON trending_courses(course_id);
CREATE INDEX idx_trending_courses_rank ON trending_courses(rank);
CREATE INDEX idx_trending_courses_meta_category ON trending_courses(meta_category);
CREATE INDEX idx_trending_courses_velocity ON trending_courses(velocity DESC);

COMMENT ON TABLE trending_courses IS 'Hourly-refreshed trending cache';
COMMENT ON COLUMN trending_courses.velocity IS 'Signups ratio (>2.0 = Trending, >5.0 = Hot)';

-- User Course Interactions (for collaborative filtering)
CREATE TABLE user_course_interactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  course_id UUID REFERENCES generated_courses(id) ON DELETE CASCADE,
  interaction_type VARCHAR(50) NOT NULL,
  interaction_value INT,
  created_at TIMESTAMP DEFAULT NOW(),
  CHECK (interaction_type IN (
    'enrolled',
    'completed',
    'viewed',
    'saved',
    'shared'
  ))
);

CREATE INDEX idx_user_course_interactions_user_id ON user_course_interactions(user_id);
CREATE INDEX idx_user_course_interactions_course_id ON user_course_interactions(course_id);
CREATE INDEX idx_user_course_interactions_type ON user_course_interactions(interaction_type);
CREATE INDEX idx_user_course_interactions_created_at ON user_course_interactions(created_at);

COMMENT ON TABLE user_course_interactions IS 'User behavior for recommendation algorithms';

-- Insert migration record
INSERT INTO schema_migrations (version, description)
VALUES ('005', 'Create discovery tables: recommendations, trending_courses, user_course_interactions');
