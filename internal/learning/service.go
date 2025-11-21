package learning

import (
	"backend/internal/platform/ai"
	"fmt"
	"strings"
)

// Service handles learning business logic
type Service struct {
	repo     *Repository
	aiClient *ai.Client
}

// NewService creates a new learning service
func NewService(repo *Repository, aiClient *ai.Client) *Service {
	return &Service{
		repo:     repo,
		aiClient: aiClient,
	}
}

// GenerateCourse creates personalized course from blueprint
func (s *Service) GenerateCourse(userID, archetypeID string, variables map[string]string) (*GeneratedCourse, error) {
	// 1. Fetch blueprint modules
	blueprints, err := s.repo.GetBlueprintModules()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blueprint modules: %w", err)
	}

	if len(blueprints) == 0 {
		return nil, fmt.Errorf("no blueprint modules found")
	}

	// 2. Extract variables for template injection
	entity := variables["ENTITY"]
	state := variables["STATE"]
	flow := variables["FLOW"]
	logic := variables["LOGIC"]
	iface := variables["INTERFACE"]

	if entity == "" {
		return nil, fmt.Errorf("ENTITY variable is required")
	}

	// 3. Create course title from first blueprint template
	courseTitle := s.injectVariables(blueprints[0].TitleTemplate, variables)
	courseDescription := fmt.Sprintf("Learn to build a %s system from first principles", entity)

	// 4. Use AI to enhance course description if available
	if s.aiClient != nil {
		aiVars := &ai.Variables{
			Entity:    entity,
			State:     state,
			Flow:      flow,
			Logic:     logic,
			Interface: iface,
		}
		curriculum, err := s.aiClient.GenerateCurriculum(archetypeID, entity, aiVars)
		if err == nil && curriculum != nil {
			courseDescription = curriculum.Description
		}
	}

	// 5. Create course instance
	course := &GeneratedCourse{
		UserID:            userID,
		ArchetypeID:       archetypeID,
		Title:             courseTitle,
		Description:       courseDescription,
		MetaCategory:      "Digital", // Default, should be determined by archetype
		InjectedVariables: variables,
		Status:            "active",
	}

	if err := s.repo.CreateGeneratedCourse(course); err != nil {
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	// 6. Create module instances with injected variables
	var modules []GeneratedModule
	for _, blueprint := range blueprints {
		module := GeneratedModule{
			CourseID:          course.ID,
			BlueprintModuleID: blueprint.ID,
			ModuleNumber:      blueprint.ModuleNumber,
			Title:             s.injectVariables(blueprint.TitleTemplate, variables),
			Description:       s.injectVariables(blueprint.DescriptionTemplate, variables),
			Status:            "locked",
		}

		// Unlock first module
		if blueprint.ModuleNumber == 1 {
			module.Status = "active"
		}

		// Generate module content using AI
		if s.aiClient != nil {
			content := map[string]interface{}{
				"lessons": []string{
					fmt.Sprintf("Introduction to %s", module.Title),
					fmt.Sprintf("Core concepts of %s", entity),
					fmt.Sprintf("Implementation patterns"),
				},
				"exercises": []string{},
			}
			module.Content = content
		}

		modules = append(modules, module)
	}

	if err := s.repo.CreateGeneratedModules(modules); err != nil {
		return nil, fmt.Errorf("failed to create modules: %w", err)
	}

	return course, nil
}

// injectVariables replaces template placeholders with actual values
func (s *Service) injectVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// GetUserCourses retrieves all courses for user
func (s *Service) GetUserCourses(userID string) ([]GeneratedCourse, error) {
	courses, err := s.repo.GetUserCourses(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user courses: %w", err)
	}
	return courses, nil
}

// GetCourseDetails retrieves course with modules
func (s *Service) GetCourseDetails(courseID string) (*GeneratedCourse, []GeneratedModule, error) {
	course, err := s.repo.GetCourseByID(courseID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get course: %w", err)
	}

	modules, err := s.repo.GetCourseModules(courseID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get course modules: %w", err)
	}

	return course, modules, nil
}

// GetExercise retrieves exercise details
func (s *Service) GetExercise(exerciseID string) (*Exercise, error) {
	exercise, err := s.repo.GetExerciseByID(exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}
	return exercise, nil
}

// TestCase represents a single test case
type TestCase struct {
	Input          interface{} `json:"input"`
	ExpectedOutput interface{} `json:"expected_output"`
	IsHidden       bool        `json:"is_hidden"`
}

// TestResult represents the result of a test case execution
type TestResult struct {
	TestCase       TestCase    `json:"test_case"`
	ActualOutput   interface{} `json:"actual_output"`
	Passed         bool        `json:"passed"`
	Error          string      `json:"error,omitempty"`
	ExecutionTime  int         `json:"execution_time_ms"`
}

// SubmitExercise handles code submission
func (s *Service) SubmitExercise(userID, exerciseID, code, language string) (*ModuleCompletion, error) {
	// 1. Fetch exercise details
	exercise, err := s.repo.GetExerciseByID(exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	// 2. Parse test cases from JSONB
	testCases, ok := exercise.TestCases.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid test cases format")
	}

	// 3. Run test cases
	var testResults []TestResult
	passedCount := 0
	totalCount := len(testCases)

	for _, tc := range testCases {
		tcMap, ok := tc.(map[string]interface{})
		if !ok {
			continue
		}

		testCase := TestCase{
			Input:          tcMap["input"],
			ExpectedOutput: tcMap["expected_output"],
			IsHidden:       tcMap["is_hidden"] != nil && tcMap["is_hidden"].(bool),
		}

		// Execute test case (simplified - real implementation would run code)
		result := s.executeTestCase(code, language, testCase, exercise.SolutionCode)
		testResults = append(testResults, result)

		if result.Passed {
			passedCount++
		}
	}

	// 4. Calculate score
	score := 0
	passed := false
	if totalCount > 0 {
		score = (passedCount * 100) / totalCount
		passed = passedCount == totalCount
	}

	// 5. Create submission record
	completion := &ModuleCompletion{
		UserID:           userID,
		ModuleID:         exercise.ModuleID,
		ExerciseID:       exerciseID,
		SubmittedCode:    code,
		Language:         language,
		TestResults:      testResults,
		Passed:           passed,
		Score:            score,
		Attempts:         1,
		HintsUsed:        0,
		TimeSpentMinutes: 0,
	}

	if err := s.repo.SubmitExercise(completion); err != nil {
		return nil, fmt.Errorf("failed to save submission: %w", err)
	}

	// 6. Update user progress
	if passed {
		// Fetch course from module
		modules, err := s.repo.GetCourseModules(exercise.ModuleID)
		if err == nil && len(modules) > 0 {
			courseID := modules[0].CourseID

			// Get current progress
			progress, err := s.repo.GetUserProgress(userID, courseID)
			if err != nil {
				// Create new progress if doesn't exist
				progress = &UserProgress{
					UserID:             userID,
					CourseID:           courseID,
					CurrentModuleID:    exercise.ModuleID,
					ProgressPercentage: 10, // Increment by 10% per module (7 modules = ~70%)
					TimeSpentMinutes:   0,
				}
			} else {
				// Update existing progress
				progress.ProgressPercentage += 10
				if progress.ProgressPercentage > 100 {
					progress.ProgressPercentage = 100
				}
			}

			_ = s.repo.UpdateUserProgress(progress)
		}
	}

	return completion, nil
}

// executeTestCase runs a test case against submitted code
// This is a simplified implementation - real version would execute code in sandbox
func (s *Service) executeTestCase(code, language string, testCase TestCase, solutionCode string) TestResult {
	result := TestResult{
		TestCase:      testCase,
		Passed:        false,
		ExecutionTime: 100, // Mock execution time
	}

	// Simplified validation: check if code contains solution patterns
	// Real implementation would execute code in a sandbox
	if strings.TrimSpace(code) == "" {
		result.Error = "Code cannot be empty"
		return result
	}

	// Basic validation: code should have similar length to solution (very naive)
	if len(code) < len(solutionCode)/3 {
		result.Error = "Solution appears incomplete"
		return result
	}

	// Mock: 80% chance of passing if code is reasonable length
	if len(code) >= len(solutionCode)/2 {
		result.Passed = true
		result.ActualOutput = testCase.ExpectedOutput
	} else {
		result.ActualOutput = nil
		result.Error = "Output does not match expected result"
	}

	return result
}

// RequestReview triggers AI Senior Review
func (s *Service) RequestReview(submissionID string) (*ArchitectureReview, error) {
	// 1. Fetch submission
	// Note: We'd need a GetSubmissionByID method in repository
	// For now, we'll create a placeholder review

	if s.aiClient == nil {
		return nil, fmt.Errorf("AI client not configured")
	}

	// 2. Mock submission data (in real implementation, fetch from DB)
	submittedCode := "// Code would be fetched from submission"
	language := "go"
	context := "Module 1: System fundamentals"

	// 3. Call AI for review
	aiReview, err := s.aiClient.ReviewCode(submittedCode, language, context)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI review: %w", err)
	}

	// 4. Create architecture review record
	review := &ArchitectureReview{
		UserID:          "", // Would be from submission
		ModuleID:        "", // Would be from submission
		SubmissionID:    submissionID,
		OverallScore:    aiReview.OverallScore,
		CodeSenseScore:  aiReview.CodeSense,
		EfficiencyScore: aiReview.Efficiency,
		EdgeCasesScore:  aiReview.EdgeCases,
		TasteScore:      aiReview.Taste,
		Feedback:        aiReview.Feedback,
	}

	if err := s.repo.CreateArchitectureReview(review); err != nil {
		return nil, fmt.Errorf("failed to save review: %w", err)
	}

	return review, nil
}

// GetUserProgress retrieves learning progress
func (s *Service) GetUserProgress(userID, courseID string) (*UserProgress, error) {
	progress, err := s.repo.GetUserProgress(userID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}
	return progress, nil
}

// GetUserCoursesInterface retrieves all courses as interface{} for social domain
func (s *Service) GetUserCoursesInterface(userID string) ([]interface{}, error) {
	courses, err := s.GetUserCourses(userID)
	if err != nil {
		return nil, err
	}

	// Convert to []interface{} for cross-domain compatibility
	result := make([]interface{}, len(courses))
	for i, course := range courses {
		result[i] = course
	}
	return result, nil
}
