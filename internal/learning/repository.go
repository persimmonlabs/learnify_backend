package learning

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository handles learning data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new learning repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetBlueprintModules retrieves all blueprint templates
func (r *Repository) GetBlueprintModules() ([]BlueprintModule, error) {
	query := `
		SELECT id, module_number, title_template, description_template,
			   difficulty, estimated_hours, learning_objectives, variable_schema,
			   created_at, updated_at
		FROM blueprint_modules
		ORDER BY module_number ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query blueprint modules: %w", err)
	}
	defer rows.Close()

	var modules []BlueprintModule
	for rows.Next() {
		var module BlueprintModule
		var objectives, schema []byte

		err := rows.Scan(
			&module.ID,
			&module.ModuleNumber,
			&module.TitleTemplate,
			&module.DescriptionTemplate,
			&module.Difficulty,
			&module.EstimatedHours,
			&objectives,
			&schema,
			&module.CreatedAt,
			&module.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blueprint module: %w", err)
		}

		// Unmarshal JSONB fields
		if len(objectives) > 0 {
			if err := json.Unmarshal(objectives, &module.LearningObjectives); err != nil {
				return nil, fmt.Errorf("failed to unmarshal learning_objectives: %w", err)
			}
		}
		if len(schema) > 0 {
			if err := json.Unmarshal(schema, &module.VariableSchema); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variable_schema: %w", err)
			}
		}

		modules = append(modules, module)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating blueprint modules: %w", err)
	}

	return modules, nil
}

// CreateGeneratedCourse creates a new course instance
func (r *Repository) CreateGeneratedCourse(course *GeneratedCourse) error {
	if course.ID == "" {
		course.ID = uuid.New().String()
	}

	// Marshal JSONB field
	variablesJSON, err := json.Marshal(course.InjectedVariables)
	if err != nil {
		return fmt.Errorf("failed to marshal injected_variables: %w", err)
	}

	query := `
		INSERT INTO generated_courses
			(id, user_id, archetype_id, title, description, meta_category,
			 injected_variables, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	course.CreatedAt = now
	course.UpdatedAt = now

	_, err = r.db.Exec(query,
		course.ID,
		course.UserID,
		course.ArchetypeID,
		course.Title,
		course.Description,
		course.MetaCategory,
		variablesJSON,
		course.Status,
		course.CreatedAt,
		course.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create generated course: %w", err)
	}

	return nil
}

// GetCourseByID retrieves course by ID
func (r *Repository) GetCourseByID(courseID string) (*GeneratedCourse, error) {
	query := `
		SELECT id, user_id, archetype_id, title, description, meta_category,
			   injected_variables, status, created_at, updated_at
		FROM generated_courses
		WHERE id = $1
	`

	var course GeneratedCourse
	var variablesJSON []byte

	err := r.db.QueryRow(query, courseID).Scan(
		&course.ID,
		&course.UserID,
		&course.ArchetypeID,
		&course.Title,
		&course.Description,
		&course.MetaCategory,
		&variablesJSON,
		&course.Status,
		&course.CreatedAt,
		&course.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("course not found: %s", courseID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	// Unmarshal JSONB field
	if len(variablesJSON) > 0 {
		if err := json.Unmarshal(variablesJSON, &course.InjectedVariables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal injected_variables: %w", err)
		}
	}

	return &course, nil
}

// GetUserCourses retrieves all courses for a user
func (r *Repository) GetUserCourses(userID string) ([]GeneratedCourse, error) {
	query := `
		SELECT id, user_id, archetype_id, title, description, meta_category,
			   injected_variables, status, created_at, updated_at
		FROM generated_courses
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user courses: %w", err)
	}
	defer rows.Close()

	var courses []GeneratedCourse
	for rows.Next() {
		var course GeneratedCourse
		var variablesJSON []byte

		err := rows.Scan(
			&course.ID,
			&course.UserID,
			&course.ArchetypeID,
			&course.Title,
			&course.Description,
			&course.MetaCategory,
			&variablesJSON,
			&course.Status,
			&course.CreatedAt,
			&course.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}

		// Unmarshal JSONB field
		if len(variablesJSON) > 0 {
			if err := json.Unmarshal(variablesJSON, &course.InjectedVariables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal injected_variables: %w", err)
			}
		}

		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating courses: %w", err)
	}

	return courses, nil
}

// CreateGeneratedModules creates module instances (batch insert)
func (r *Repository) CreateGeneratedModules(modules []GeneratedModule) error {
	if len(modules) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO generated_modules
			(id, course_id, blueprint_module_id, module_number, title,
			 description, content, status, unlocked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, module := range modules {
		if module.ID == "" {
			modules[i].ID = uuid.New().String()
			module.ID = modules[i].ID
		}

		// Marshal JSONB content
		var contentJSON []byte
		if module.Content != nil {
			contentJSON, err = json.Marshal(module.Content)
			if err != nil {
				return fmt.Errorf("failed to marshal content for module %d: %w", i, err)
			}
		}

		now := time.Now()
		modules[i].CreatedAt = now

		_, err = stmt.Exec(
			module.ID,
			module.CourseID,
			module.BlueprintModuleID,
			module.ModuleNumber,
			module.Title,
			module.Description,
			contentJSON,
			module.Status,
			module.UnlockedAt,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert module %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetCourseModules retrieves modules for a course
func (r *Repository) GetCourseModules(courseID string) ([]GeneratedModule, error) {
	query := `
		SELECT id, course_id, blueprint_module_id, module_number, title,
			   description, content, status, unlocked_at, created_at
		FROM generated_modules
		WHERE course_id = $1
		ORDER BY module_number ASC
	`

	rows, err := r.db.Query(query, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query course modules: %w", err)
	}
	defer rows.Close()

	var modules []GeneratedModule
	for rows.Next() {
		var module GeneratedModule
		var contentJSON []byte
		var unlockedAt sql.NullTime

		err := rows.Scan(
			&module.ID,
			&module.CourseID,
			&module.BlueprintModuleID,
			&module.ModuleNumber,
			&module.Title,
			&module.Description,
			&contentJSON,
			&module.Status,
			&unlockedAt,
			&module.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}

		// Unmarshal JSONB content
		if len(contentJSON) > 0 {
			if err := json.Unmarshal(contentJSON, &module.Content); err != nil {
				return nil, fmt.Errorf("failed to unmarshal content: %w", err)
			}
		}

		if unlockedAt.Valid {
			module.UnlockedAt = &unlockedAt.Time
		}

		modules = append(modules, module)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating modules: %w", err)
	}

	return modules, nil
}

// CreateExercise creates a coding challenge
func (r *Repository) CreateExercise(exercise *Exercise) error {
	if exercise.ID == "" {
		exercise.ID = uuid.New().String()
	}

	// Marshal JSONB fields
	testCasesJSON, err := json.Marshal(exercise.TestCases)
	if err != nil {
		return fmt.Errorf("failed to marshal test_cases: %w", err)
	}

	hintsJSON, err := json.Marshal(exercise.Hints)
	if err != nil {
		return fmt.Errorf("failed to marshal hints: %w", err)
	}

	query := `
		INSERT INTO exercises
			(id, module_id, exercise_number, title, description, language,
			 starter_code, solution_code, test_cases, difficulty, points, hints, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now()
	exercise.CreatedAt = now

	_, err = r.db.Exec(query,
		exercise.ID,
		exercise.ModuleID,
		exercise.ExerciseNumber,
		exercise.Title,
		exercise.Description,
		exercise.Language,
		exercise.StarterCode,
		exercise.SolutionCode,
		testCasesJSON,
		exercise.Difficulty,
		exercise.Points,
		hintsJSON,
		exercise.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create exercise: %w", err)
	}

	return nil
}

// GetExerciseByID retrieves exercise by ID
func (r *Repository) GetExerciseByID(exerciseID string) (*Exercise, error) {
	query := `
		SELECT id, module_id, exercise_number, title, description, language,
			   starter_code, solution_code, test_cases, difficulty, points, hints, created_at
		FROM exercises
		WHERE id = $1
	`

	var exercise Exercise
	var testCasesJSON, hintsJSON []byte

	err := r.db.QueryRow(query, exerciseID).Scan(
		&exercise.ID,
		&exercise.ModuleID,
		&exercise.ExerciseNumber,
		&exercise.Title,
		&exercise.Description,
		&exercise.Language,
		&exercise.StarterCode,
		&exercise.SolutionCode,
		&testCasesJSON,
		&exercise.Difficulty,
		&exercise.Points,
		&hintsJSON,
		&exercise.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("exercise not found: %s", exerciseID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	// Unmarshal JSONB fields
	if len(testCasesJSON) > 0 {
		if err := json.Unmarshal(testCasesJSON, &exercise.TestCases); err != nil {
			return nil, fmt.Errorf("failed to unmarshal test_cases: %w", err)
		}
	}
	if len(hintsJSON) > 0 {
		if err := json.Unmarshal(hintsJSON, &exercise.Hints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal hints: %w", err)
		}
	}

	return &exercise, nil
}

// SubmitExercise saves exercise submission
func (r *Repository) SubmitExercise(completion *ModuleCompletion) error {
	if completion.ID == "" {
		completion.ID = uuid.New().String()
	}

	// Marshal JSONB test results
	testResultsJSON, err := json.Marshal(completion.TestResults)
	if err != nil {
		return fmt.Errorf("failed to marshal test_results: %w", err)
	}

	query := `
		INSERT INTO module_completions
			(id, user_id, module_id, exercise_id, submitted_code, language,
			 test_results, passed, score, attempts, hints_used, time_spent_minutes, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now()
	completion.SubmittedAt = now

	_, err = r.db.Exec(query,
		completion.ID,
		completion.UserID,
		completion.ModuleID,
		completion.ExerciseID,
		completion.SubmittedCode,
		completion.Language,
		testResultsJSON,
		completion.Passed,
		completion.Score,
		completion.Attempts,
		completion.HintsUsed,
		completion.TimeSpentMinutes,
		completion.SubmittedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to submit exercise: %w", err)
	}

	return nil
}

// GetUserProgress retrieves user's course progress
func (r *Repository) GetUserProgress(userID, courseID string) (*UserProgress, error) {
	query := `
		SELECT id, user_id, course_id, current_module_id, progress_percentage,
			   time_spent_minutes, last_activity, started_at, completed_at
		FROM user_progress
		WHERE user_id = $1 AND course_id = $2
	`

	var progress UserProgress
	var currentModuleID sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRow(query, userID, courseID).Scan(
		&progress.ID,
		&progress.UserID,
		&progress.CourseID,
		&currentModuleID,
		&progress.ProgressPercentage,
		&progress.TimeSpentMinutes,
		&progress.LastActivity,
		&progress.StartedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("progress not found for user %s and course %s", userID, courseID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user progress: %w", err)
	}

	if currentModuleID.Valid {
		progress.CurrentModuleID = currentModuleID.String
	}
	if completedAt.Valid {
		progress.CompletedAt = &completedAt.Time
	}

	return &progress, nil
}

// UpdateUserProgress updates course progress (or creates if not exists)
func (r *Repository) UpdateUserProgress(progress *UserProgress) error {
	// Try to update first
	updateQuery := `
		UPDATE user_progress
		SET current_module_id = $1,
			progress_percentage = $2,
			time_spent_minutes = $3,
			last_activity = $4,
			completed_at = $5
		WHERE user_id = $6 AND course_id = $7
	`

	result, err := r.db.Exec(updateQuery,
		progress.CurrentModuleID,
		progress.ProgressPercentage,
		progress.TimeSpentMinutes,
		time.Now(),
		progress.CompletedAt,
		progress.UserID,
		progress.CourseID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user progress: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// If no rows updated, insert new record
	if rowsAffected == 0 {
		if progress.ID == "" {
			progress.ID = uuid.New().String()
		}

		insertQuery := `
			INSERT INTO user_progress
				(id, user_id, course_id, current_module_id, progress_percentage,
				 time_spent_minutes, last_activity, started_at, completed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		now := time.Now()
		progress.StartedAt = now
		progress.LastActivity = now

		_, err = r.db.Exec(insertQuery,
			progress.ID,
			progress.UserID,
			progress.CourseID,
			progress.CurrentModuleID,
			progress.ProgressPercentage,
			progress.TimeSpentMinutes,
			progress.LastActivity,
			progress.StartedAt,
			progress.CompletedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to insert user progress: %w", err)
		}
	}

	return nil
}

// CreateArchitectureReview saves AI review
func (r *Repository) CreateArchitectureReview(review *ArchitectureReview) error {
	if review.ID == "" {
		review.ID = uuid.New().String()
	}

	// Marshal JSONB feedback
	feedbackJSON, err := json.Marshal(review.Feedback)
	if err != nil {
		return fmt.Errorf("failed to marshal feedback: %w", err)
	}

	query := `
		INSERT INTO architecture_reviews
			(id, user_id, module_id, submission_id, overall_score,
			 code_sense_score, efficiency_score, edge_cases_score, taste_score,
			 feedback, reviewed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	review.ReviewedAt = now

	_, err = r.db.Exec(query,
		review.ID,
		review.UserID,
		review.ModuleID,
		review.SubmissionID,
		review.OverallScore,
		review.CodeSenseScore,
		review.EfficiencyScore,
		review.EdgeCasesScore,
		review.TasteScore,
		feedbackJSON,
		review.ReviewedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create architecture review: %w", err)
	}

	return nil
}
