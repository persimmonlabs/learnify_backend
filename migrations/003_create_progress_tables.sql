-- Migration 003: Progress Layer
-- User learning progress and architecture reviews

-- User Progress (Overall Course Progress)
CREATE TABLE user_progress (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  course_id UUID REFERENCES generated_courses(id) ON DELETE CASCADE,
  current_module_id UUID REFERENCES generated_modules(id),
  progress_percentage INT DEFAULT 0,
  time_spent_minutes INT DEFAULT 0,
  last_activity TIMESTAMP DEFAULT NOW(),
  started_at TIMESTAMP DEFAULT NOW(),
  completed_at TIMESTAMP,
  CHECK (progress_percentage BETWEEN 0 AND 100),
  UNIQUE(user_id, course_id)
);

CREATE INDEX idx_user_progress_user_id ON user_progress(user_id);
CREATE INDEX idx_user_progress_course_id ON user_progress(course_id);
CREATE INDEX idx_user_progress_last_activity ON user_progress(last_activity);

COMMENT ON TABLE user_progress IS 'High-level course progress tracking';

-- Module Completions
CREATE TABLE module_completions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  module_id UUID REFERENCES generated_modules(id) ON DELETE CASCADE,
  exercise_id UUID REFERENCES exercises(id),
  submitted_code TEXT,
  language VARCHAR(20),
  test_results JSONB,
  passed BOOLEAN DEFAULT false,
  score INT,
  attempts INT DEFAULT 1,
  hints_used INT DEFAULT 0,
  time_spent_minutes INT DEFAULT 0,
  submitted_at TIMESTAMP DEFAULT NOW(),
  CHECK (language IN ('python', 'go', 'java', 'javascript'))
);

CREATE INDEX idx_module_completions_user_id ON module_completions(user_id);
CREATE INDEX idx_module_completions_module_id ON module_completions(module_id);
CREATE INDEX idx_module_completions_exercise_id ON module_completions(exercise_id);
CREATE INDEX idx_module_completions_passed ON module_completions(passed);
CREATE INDEX idx_module_completions_submitted_at ON module_completions(submitted_at);

COMMENT ON TABLE module_completions IS 'Exercise submissions and results';
COMMENT ON COLUMN module_completions.test_results IS 'Array of test case results';

-- Architecture Reviews (AI Senior Review)
CREATE TABLE architecture_reviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  module_id UUID REFERENCES generated_modules(id) ON DELETE CASCADE,
  submission_id UUID REFERENCES module_completions(id),
  overall_score INT,
  code_sense_score INT,
  efficiency_score INT,
  edge_cases_score INT,
  taste_score INT,
  feedback JSONB,
  reviewed_at TIMESTAMP DEFAULT NOW(),
  CHECK (overall_score BETWEEN 0 AND 100),
  CHECK (code_sense_score BETWEEN 0 AND 100),
  CHECK (efficiency_score BETWEEN 0 AND 100),
  CHECK (edge_cases_score BETWEEN 0 AND 100),
  CHECK (taste_score BETWEEN 0 AND 100)
);

CREATE INDEX idx_architecture_reviews_user_id ON architecture_reviews(user_id);
CREATE INDEX idx_architecture_reviews_module_id ON architecture_reviews(module_id);
CREATE INDEX idx_architecture_reviews_overall_score ON architecture_reviews(overall_score);
CREATE INDEX idx_architecture_reviews_reviewed_at ON architecture_reviews(reviewed_at);

COMMENT ON TABLE architecture_reviews IS 'AI Senior Review critiques (4 categories)';
COMMENT ON COLUMN architecture_reviews.feedback IS 'Detailed critique per category';

-- Insert migration record
INSERT INTO schema_migrations (version, description)
VALUES ('003', 'Create progress tables: user_progress, module_completions, architecture_reviews');
