package learning

import (
	"time"
)

// BlueprintModule represents universal blueprint template
type BlueprintModule struct {
	ID                 string
	ModuleNumber       int
	TitleTemplate      string
	DescriptionTemplate string
	Difficulty         string
	EstimatedHours     int
	LearningObjectives interface{}
	VariableSchema     interface{}
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// GeneratedCourse represents user-specific course instance
type GeneratedCourse struct {
	ID               string
	UserID           string
	ArchetypeID      string
	Title            string
	Description      string
	MetaCategory     string
	InjectedVariables interface{}
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// GeneratedModule represents module instance with injected variables
type GeneratedModule struct {
	ID                string
	CourseID          string
	BlueprintModuleID string
	ModuleNumber      int
	Title             string
	Description       string
	Content           interface{}
	Status            string
	UnlockedAt        *time.Time
	CreatedAt         time.Time
}

// Exercise represents a coding challenge
type Exercise struct {
	ID             string
	ModuleID       string
	ExerciseNumber int
	Title          string
	Description    string
	Language       string
	StarterCode    string
	SolutionCode   string
	TestCases      interface{}
	Difficulty     string
	Points         int
	Hints          interface{}
	CreatedAt      time.Time
}

// UserProgress represents overall course progress
type UserProgress struct {
	ID                 string
	UserID             string
	CourseID           string
	CurrentModuleID    string
	ProgressPercentage int
	TimeSpentMinutes   int
	LastActivity       time.Time
	StartedAt          time.Time
	CompletedAt        *time.Time
}

// ModuleCompletion represents exercise submission
type ModuleCompletion struct {
	ID               string
	UserID           string
	ModuleID         string
	ExerciseID       string
	SubmittedCode    string
	Language         string
	TestResults      interface{}
	Passed           bool
	Score            int
	Attempts         int
	HintsUsed        int
	TimeSpentMinutes int
	SubmittedAt      time.Time
}

// ArchitectureReview represents AI Senior Review
type ArchitectureReview struct {
	ID              string
	UserID          string
	ModuleID        string
	SubmissionID    string
	OverallScore    int
	CodeSenseScore  int
	EfficiencyScore int
	EdgeCasesScore  int
	TasteScore      int
	Feedback        interface{}
	ReviewedAt      time.Time
}
