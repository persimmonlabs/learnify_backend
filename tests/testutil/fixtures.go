package testutil

import (
	"time"

	"github.com/google/uuid"
)

// TestUser represents a test user (avoiding import cycle)
type TestUser struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	AvatarURL    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLogin    time.Time
}

// CreateTestUser creates a test user
func CreateTestUser(email, name string) *TestUser {
	return &TestUser{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: "$2a$10$test.hash.value.here.for.testing",
		Name:         name,
		AvatarURL:    "https://example.com/avatar.jpg",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastLogin:    time.Now(),
	}
}

// TestArchetype represents a test user archetype
type TestArchetype struct {
	ID           string
	UserID       string
	MetaCategory string
	Domain       string
	SkillLevel   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateTestArchetype creates a test user archetype
func CreateTestArchetype(userID string) *TestArchetype {
	return &TestArchetype{
		ID:           uuid.New().String(),
		UserID:       userID,
		MetaCategory: "Digital",
		Domain:       "Web Development",
		SkillLevel:   "beginner",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// TestVariable represents a test user variable
type TestVariable struct {
	ID            string
	UserID        string
	VariableKey   string
	VariableValue string
	ArchetypeID   string
	CreatedAt     time.Time
}

// CreateTestVariables creates test user variables
func CreateTestVariables(userID, archetypeID string) []TestVariable {
	return []TestVariable{
		{
			ID:            uuid.New().String(),
			UserID:        userID,
			VariableKey:   "ENTITY",
			VariableValue: "E-commerce",
			ArchetypeID:   archetypeID,
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New().String(),
			UserID:        userID,
			VariableKey:   "STATE",
			VariableValue: "Cart",
			ArchetypeID:   archetypeID,
			CreatedAt:     time.Now(),
		},
	}
}

// TestBlueprintModule represents a test blueprint module
type TestBlueprintModule struct {
	ID                  string
	ModuleNumber        int
	TitleTemplate       string
	DescriptionTemplate string
	Difficulty          string
	EstimatedHours      int
	LearningObjectives  []string
	VariableSchema      map[string]interface{}
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// CreateTestBlueprintModule creates a test blueprint module
func CreateTestBlueprintModule(moduleNumber int) TestBlueprintModule {
	return TestBlueprintModule{
		ID:                  uuid.New().String(),
		ModuleNumber:        moduleNumber,
		TitleTemplate:       "Module {ENTITY} - Part {NUMBER}",
		DescriptionTemplate: "Learn about {ENTITY} systems",
		Difficulty:          "beginner",
		EstimatedHours:      5,
		LearningObjectives:  []string{"Understand basics", "Build projects"},
		VariableSchema:      map[string]interface{}{"entity": "string"},
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// TestGeneratedCourse represents a test generated course
type TestGeneratedCourse struct {
	ID                string
	UserID            string
	ArchetypeID       string
	Title             string
	Description       string
	MetaCategory      string
	InjectedVariables map[string]string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CreateTestGeneratedCourse creates a test generated course
func CreateTestGeneratedCourse(userID, archetypeID string) *TestGeneratedCourse {
	return &TestGeneratedCourse{
		ID:                uuid.New().String(),
		UserID:            userID,
		ArchetypeID:       archetypeID,
		Title:             "Test Course",
		Description:       "A test course description",
		MetaCategory:      "Digital",
		InjectedVariables: map[string]string{"ENTITY": "E-commerce"},
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// TestGeneratedModule represents a test generated module
type TestGeneratedModule struct {
	ID                string
	CourseID          string
	BlueprintModuleID string
	ModuleNumber      int
	Title             string
	Description       string
	Content           map[string]interface{}
	Status            string
	CreatedAt         time.Time
}

// CreateTestGeneratedModule creates a test generated module
func CreateTestGeneratedModule(courseID string, moduleNumber int) TestGeneratedModule {
	return TestGeneratedModule{
		ID:                uuid.New().String(),
		CourseID:          courseID,
		BlueprintModuleID: uuid.New().String(),
		ModuleNumber:      moduleNumber,
		Title:             "Test Module",
		Description:       "Test module description",
		Content:           map[string]interface{}{"lessons": []string{"Lesson 1"}},
		Status:            "active",
		CreatedAt:         time.Now(),
	}
}

// TestExercise represents a test exercise
type TestExercise struct {
	ID             string
	ModuleID       string
	ExerciseNumber int
	Title          string
	Description    string
	Language       string
	StarterCode    string
	SolutionCode   string
	TestCases      []interface{}
	Difficulty     string
	Points         int
	Hints          []string
	CreatedAt      time.Time
}

// CreateTestExercise creates a test exercise
func CreateTestExercise(moduleID string) *TestExercise {
	return &TestExercise{
		ID:             uuid.New().String(),
		ModuleID:       moduleID,
		ExerciseNumber: 1,
		Title:          "Test Exercise",
		Description:    "Solve this problem",
		Language:       "go",
		StarterCode:    "// Start here",
		SolutionCode:   "// Solution code\nfunc solve() { return 42 }",
		TestCases: []interface{}{
			map[string]interface{}{
				"input":           1,
				"expected_output": 2,
				"is_hidden":       false,
			},
		},
		Difficulty: "easy",
		Points:     10,
		Hints:      []string{"Think about the problem"},
		CreatedAt:  time.Now(),
	}
}

// TestActivity represents a test activity feed item
type TestActivity struct {
	ID            string
	UserID        string
	ActivityType  string
	ReferenceType string
	ReferenceID   string
	Metadata      map[string]interface{}
	Visibility    string
	CreatedAt     time.Time
}

// CreateTestActivity creates a test activity feed item
func CreateTestActivity(userID string) *TestActivity {
	return &TestActivity{
		ID:            uuid.New().String(),
		UserID:        userID,
		ActivityType:  "course_completed",
		ReferenceType: "course",
		ReferenceID:   uuid.New().String(),
		Metadata:      map[string]interface{}{"score": 95},
		Visibility:    "public",
		CreatedAt:     time.Now(),
	}
}

// TestRecommendation represents a test recommendation
type TestRecommendation struct {
	ID                 string
	UserID             string
	CourseID           string
	RecommendationType string
	MatchScore         int
	Reason             string
	Metadata           map[string]interface{}
	ExpiresAt          *time.Time
	CreatedAt          time.Time
}

// CreateTestRecommendation creates a test recommendation
func CreateTestRecommendation(userID, courseID string) *TestRecommendation {
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	return &TestRecommendation{
		ID:                 uuid.New().String(),
		UserID:             userID,
		CourseID:           courseID,
		RecommendationType: "collaborative_filtering",
		MatchScore:         85,
		Reason:             "Users like you enjoyed this",
		Metadata:           map[string]interface{}{"similar_users": 5},
		ExpiresAt:          &expiresAt,
		CreatedAt:          time.Now(),
	}
}

// TestAchievement represents a test achievement
type TestAchievement struct {
	ID          string
	Name        string
	Description string
	Rarity      string
	IconURL     string
	CreatedAt   time.Time
}

// CreateTestAchievement creates a test achievement
func CreateTestAchievement(id, name string) TestAchievement {
	return TestAchievement{
		ID:          id,
		Name:        name,
		Description: "Test achievement",
		Rarity:      "rare",
		IconURL:     "https://example.com/icon.png",
		CreatedAt:   time.Now(),
	}
}
