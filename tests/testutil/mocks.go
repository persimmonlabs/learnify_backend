package testutil

import (
	"backend/internal/platform/ai"
	"errors"
)

// MockAIClient is a mock implementation of the AI client
type MockAIClient struct {
	ShouldFail        bool
	ExtractVarsResult *ai.Variables
	CurriculumResult  *ai.Curriculum
	ReviewResult      *ai.CodeReview
}

// ExtractVariables mocks variable extraction
func (m *MockAIClient) ExtractVariables(domain string) (*ai.Variables, error) {
	if m.ShouldFail {
		return nil, errors.New("mock AI failure")
	}
	if m.ExtractVarsResult != nil {
		return m.ExtractVarsResult, nil
	}
	return &ai.Variables{
		Entity:    "E-commerce",
		State:     "Cart",
		Flow:      "Checkout",
		Logic:     "Validation",
		Interface: "REST API",
	}, nil
}

// GenerateCurriculum mocks curriculum generation
func (m *MockAIClient) GenerateCurriculum(archetypeID, domain string, variables *ai.Variables) (*ai.Curriculum, error) {
	if m.ShouldFail {
		return nil, errors.New("mock AI failure")
	}
	if m.CurriculumResult != nil {
		return m.CurriculumResult, nil
	}
	return &ai.Curriculum{
		Title:       "Generated Course Title",
		Description: "AI-generated course description",
		Modules: []ai.Module{
			{
				Title:       "Module 1",
				Description: "First module",
				Lessons:     []string{"Lesson 1", "Lesson 2"},
			},
		},
	}, nil
}

// ReviewCode mocks code review
func (m *MockAIClient) ReviewCode(code, language, context string) (*ai.CodeReview, error) {
	if m.ShouldFail {
		return nil, errors.New("mock AI failure")
	}
	if m.ReviewResult != nil {
		return m.ReviewResult, nil
	}
	return &ai.CodeReview{
		OverallScore: 85,
		CodeSense:    88,
		Efficiency:   82,
		EdgeCases:    80,
		Taste:        90,
		Feedback: map[string]interface{}{
			"strengths": []string{"Good structure"},
			"improvements": []string{"Add error handling"},
		},
	}, nil
}

// MockCourseGenerator mocks course generation
type MockCourseGenerator struct {
	ShouldFail bool
	Called     bool
}

// GenerateCourse mocks course generation
func (m *MockCourseGenerator) GenerateCourse(userID, archetypeID string, variables map[string]string) error {
	m.Called = true
	if m.ShouldFail {
		return errors.New("mock course generation failure")
	}
	return nil
}
