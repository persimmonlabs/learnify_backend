-- Seed Test Data for Learnify Backend
-- Creates sample courses, exercises, and test cases for testing
-- Run after initial migrations: psql -U postgres -d learnify -f seed_test_data.sql

-- Only run if tables are empty (safe for testing)
DO $$
BEGIN
  IF (SELECT COUNT(*) FROM courses) = 0 THEN

    -- Insert sample courses
    INSERT INTO courses (id, title, description, difficulty, estimated_hours, archetype_id, created_at, updated_at)
    VALUES
      (
        gen_random_uuid(),
        'Web Development Fundamentals',
        'Learn the basics of web development including HTML, CSS, JavaScript, and modern frameworks. Build real-world projects from scratch.',
        'beginner',
        40,
        (SELECT id FROM archetypes WHERE name = 'code-craftsperson' LIMIT 1),
        NOW(),
        NOW()
      ),
      (
        gen_random_uuid(),
        'Python for Data Science',
        'Master Python programming and data science libraries including NumPy, Pandas, and Matplotlib. Work with real datasets.',
        'intermediate',
        60,
        (SELECT id FROM archetypes WHERE name = 'system-thinker' LIMIT 1),
        NOW(),
        NOW()
      ),
      (
        gen_random_uuid(),
        'System Design Patterns',
        'Learn architectural patterns and system design principles. Design scalable, resilient distributed systems.',
        'advanced',
        80,
        (SELECT id FROM archetypes WHERE name = 'system-thinker' LIMIT 1),
        NOW(),
        NOW()
      );

    -- Insert exercises for Web Development course
    INSERT INTO exercises (id, course_id, title, description, difficulty, order_index, starter_code, solution_code, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      c.id,
      'Build a Responsive Navigation Bar',
      'Create a mobile-friendly navigation bar using HTML5 and CSS3. Include a hamburger menu for mobile devices.',
      'beginner',
      1,
      '<!DOCTYPE html>
<html>
<head>
  <style>
    /* Your CSS here */
  </style>
</head>
<body>
  <nav>
    <!-- Your navigation structure here -->
  </nav>
</body>
</html>',
      '<!DOCTYPE html>
<html>
<head>
  <style>
    nav { background: #333; color: white; padding: 1rem; }
    nav ul { list-style: none; display: flex; gap: 1rem; }
    @media (max-width: 768px) {
      nav ul { flex-direction: column; }
    }
  </style>
</head>
<body>
  <nav>
    <ul>
      <li><a href="#">Home</a></li>
      <li><a href="#">About</a></li>
      <li><a href="#">Contact</a></li>
    </ul>
  </nav>
</body>
</html>',
      NOW(),
      NOW()
    FROM courses c
    WHERE c.title = 'Web Development Fundamentals';

    INSERT INTO exercises (id, course_id, title, description, difficulty, order_index, starter_code, solution_code, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      c.id,
      'Implement Form Validation',
      'Create a registration form with client-side validation using JavaScript. Validate email, password strength, and required fields.',
      'beginner',
      2,
      'function validateForm(formData) {
  // Implement validation logic
  return { valid: false, errors: [] };
}',
      'function validateForm(formData) {
  const errors = [];
  if (!formData.email.includes("@")) errors.push("Invalid email");
  if (formData.password.length < 8) errors.push("Password too short");
  return { valid: errors.length === 0, errors };
}',
      NOW(),
      NOW()
    FROM courses c
    WHERE c.title = 'Web Development Fundamentals';

    -- Insert exercises for Python course
    INSERT INTO exercises (id, course_id, title, description, difficulty, order_index, starter_code, solution_code, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      c.id,
      'Data Cleaning with Pandas',
      'Clean a messy dataset by handling missing values, removing duplicates, and standardizing formats.',
      'intermediate',
      1,
      'import pandas as pd

def clean_dataset(df):
    # Implement data cleaning logic
    return df',
      'import pandas as pd

def clean_dataset(df):
    df = df.drop_duplicates()
    df = df.fillna(df.mean())
    df["date"] = pd.to_datetime(df["date"])
    return df',
      NOW(),
      NOW()
    FROM courses c
    WHERE c.title = 'Python for Data Science';

    INSERT INTO exercises (id, course_id, title, description, difficulty, order_index, starter_code, solution_code, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      c.id,
      'Statistical Analysis',
      'Perform statistical analysis on a dataset including mean, median, standard deviation, and correlation.',
      'intermediate',
      2,
      'import numpy as np

def analyze_data(data):
    # Return statistical summary
    return {}',
      'import numpy as np

def analyze_data(data):
    return {
        "mean": np.mean(data),
        "median": np.median(data),
        "std": np.std(data),
        "min": np.min(data),
        "max": np.max(data)
    }',
      NOW(),
      NOW()
    FROM courses c
    WHERE c.title = 'Python for Data Science';

    -- Insert exercises for System Design course
    INSERT INTO exercises (id, course_id, title, description, difficulty, order_index, starter_code, solution_code, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      c.id,
      'Design a URL Shortener',
      'Design a distributed URL shortening service like bit.ly. Consider scalability, collision handling, and analytics.',
      'advanced',
      1,
      '// System Design Prompt
// Design a URL shortener that handles:
// - 100M URLs
// - 1000 reads/sec
// - High availability
// - Custom short URLs
// - Click analytics

class URLShortener {
  // Your design here
}',
      '// System Components:
// 1. Load Balancer (NGINX)
// 2. API Servers (Node.js cluster)
// 3. Redis Cache (short URL → long URL)
// 4. PostgreSQL (persistent storage)
// 5. Analytics Service (Kafka + ClickHouse)
// 6. CDN (CloudFlare)

// Algorithm: Base62 encoding of auto-increment ID
// Collision: Check Redis before insertion
// Scaling: Horizontal scaling with consistent hashing',
      NOW(),
      NOW()
    FROM courses c
    WHERE c.title = 'System Design Patterns';

    -- Insert test cases for exercises
    INSERT INTO test_cases (id, exercise_id, name, input, expected_output, is_hidden, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      e.id,
      'Test valid email',
      '{"email": "user@example.com", "password": "SecurePass123"}',
      '{"valid": true, "errors": []}',
      false,
      NOW(),
      NOW()
    FROM exercises e
    WHERE e.title = 'Implement Form Validation';

    INSERT INTO test_cases (id, exercise_id, name, input, expected_output, is_hidden, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      e.id,
      'Test invalid email',
      '{"email": "invalid-email", "password": "SecurePass123"}',
      '{"valid": false, "errors": ["Invalid email"]}',
      false,
      NOW(),
      NOW()
    FROM exercises e
    WHERE e.title = 'Implement Form Validation';

    INSERT INTO test_cases (id, exercise_id, name, input, expected_output, is_hidden, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      e.id,
      'Test short password',
      '{"email": "user@example.com", "password": "short"}',
      '{"valid": false, "errors": ["Password too short"]}',
      false,
      NOW(),
      NOW()
    FROM exercises e
    WHERE e.title = 'Implement Form Validation';

    -- Insert test cases for Python exercise
    INSERT INTO test_cases (id, exercise_id, name, input, expected_output, is_hidden, created_at, updated_at)
    SELECT
      gen_random_uuid(),
      e.id,
      'Test basic statistics',
      '[1, 2, 3, 4, 5]',
      '{"mean": 3.0, "median": 3.0, "std": 1.414, "min": 1, "max": 5}',
      false,
      NOW(),
      NOW()
    FROM exercises e
    WHERE e.title = 'Statistical Analysis';

    RAISE NOTICE 'Test data seeded successfully!';
    RAISE NOTICE 'Created: 3 courses, 5 exercises, 4 test cases';
  ELSE
    RAISE NOTICE 'Test data already exists. Skipping seed.';
  END IF;
END $$;
