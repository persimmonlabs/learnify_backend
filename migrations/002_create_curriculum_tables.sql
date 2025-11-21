-- Migration 002: Curriculum Layer
-- Blueprint templates and generated course instances

-- Blueprint Modules (Template Layer)
CREATE TABLE blueprint_modules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  module_number INT NOT NULL,
  title_template VARCHAR(255) NOT NULL,
  description_template TEXT,
  difficulty VARCHAR(20) NOT NULL,
  estimated_hours INT,
  learning_objectives JSONB,
  variable_schema JSONB NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CHECK (module_number BETWEEN 1 AND 7),
  CHECK (difficulty IN ('beginner', 'intermediate', 'advanced'))
);

CREATE INDEX idx_blueprint_modules_number ON blueprint_modules(module_number);
CREATE INDEX idx_blueprint_modules_difficulty ON blueprint_modules(difficulty);

COMMENT ON TABLE blueprint_modules IS 'Universal Blueprint templates (7 modules)';
COMMENT ON COLUMN blueprint_modules.title_template IS 'Template with {ENTITY} placeholders';
COMMENT ON COLUMN blueprint_modules.variable_schema IS 'Required variables for this module';

-- Generated Courses (Instance Layer)
CREATE TABLE generated_courses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  archetype_id UUID REFERENCES user_archetypes(id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  meta_category VARCHAR(50) NOT NULL,
  injected_variables JSONB NOT NULL,
  status VARCHAR(20) DEFAULT 'active',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CHECK (meta_category IN ('Digital', 'Economic', 'Aesthetic', 'Biological', 'Cognitive')),
  CHECK (status IN ('active', 'paused', 'completed', 'archived'))
);

CREATE INDEX idx_generated_courses_user_id ON generated_courses(user_id);
CREATE INDEX idx_generated_courses_archetype_id ON generated_courses(archetype_id);
CREATE INDEX idx_generated_courses_meta_category ON generated_courses(meta_category);
CREATE INDEX idx_generated_courses_status ON generated_courses(status);

COMMENT ON TABLE generated_courses IS 'User-specific course instances with injected variables';
COMMENT ON COLUMN generated_courses.injected_variables IS 'User runtime variables (e.g., {ENTITY: "Trading Bot"})';

-- Generated Modules (Blueprint → Instance)
CREATE TABLE generated_modules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id UUID REFERENCES generated_courses(id) ON DELETE CASCADE,
  blueprint_module_id UUID REFERENCES blueprint_modules(id),
  module_number INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  content JSONB,
  status VARCHAR(20) DEFAULT 'locked',
  unlocked_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  CHECK (module_number BETWEEN 1 AND 7),
  CHECK (status IN ('locked', 'active', 'completed'))
);

CREATE INDEX idx_generated_modules_course_id ON generated_modules(course_id);
CREATE INDEX idx_generated_modules_blueprint_id ON generated_modules(blueprint_module_id);
CREATE INDEX idx_generated_modules_status ON generated_modules(status);

COMMENT ON TABLE generated_modules IS 'Course-specific module instances (variables already injected)';
COMMENT ON COLUMN generated_modules.content IS 'Lessons, exercises, visualizations for this module';

-- Exercises
CREATE TABLE exercises (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  module_id UUID REFERENCES generated_modules(id) ON DELETE CASCADE,
  exercise_number INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  language VARCHAR(20) NOT NULL,
  starter_code TEXT,
  solution_code TEXT,
  test_cases JSONB NOT NULL,
  difficulty VARCHAR(20) NOT NULL,
  points INT DEFAULT 100,
  hints JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  CHECK (language IN ('python', 'go', 'java', 'javascript')),
  CHECK (difficulty IN ('easy', 'medium', 'hard'))
);

CREATE INDEX idx_exercises_module_id ON exercises(module_id);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty);
CREATE INDEX idx_exercises_language ON exercises(language);

COMMENT ON TABLE exercises IS 'Coding challenges for each module';
COMMENT ON COLUMN exercises.test_cases IS 'Array of {input, expected_output, is_hidden}';
COMMENT ON COLUMN exercises.hints IS 'Array of {text, penalty_points}';

-- Course Tags (for filtering/recommendations)
CREATE TABLE course_tags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id UUID REFERENCES generated_courses(id) ON DELETE CASCADE,
  tag VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(course_id, tag)
);

CREATE INDEX idx_course_tags_course_id ON course_tags(course_id);
CREATE INDEX idx_course_tags_tag ON course_tags(tag);

COMMENT ON TABLE course_tags IS 'Skills/topics for recommendation engine';

-- Insert migration record
INSERT INTO schema_migrations (version, description)
VALUES ('002', 'Create curriculum tables: blueprint_modules, generated_courses, generated_modules, exercises, course_tags');
