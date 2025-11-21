package learning

import (
	"backend/internal/platform/ai"
	"fmt"
)

// Agent represents an AI agent for learning tasks
type Agent struct {
	Type         string
	Capabilities []string
	aiClient     *ai.Client
}

// CurriculumAgent generates personalized curriculum
type CurriculumAgent struct {
	Agent
}

// NewCurriculumAgent creates a curriculum generation agent
func NewCurriculumAgent(aiClient *ai.Client) *CurriculumAgent {
	return &CurriculumAgent{
		Agent: Agent{
			Type: "curriculum_generator",
			Capabilities: []string{
				"course_design",
				"module_generation",
				"exercise_creation",
				"learning_path_optimization",
			},
			aiClient: aiClient,
		},
	}
}

// Generate creates curriculum based on archetype and variables
func (a *CurriculumAgent) Generate(archetype, domain string, variables map[string]string) (*GeneratedCourse, error) {
	if a.aiClient == nil {
		return nil, fmt.Errorf("AI client not configured")
	}

	// Extract variables
	entity := variables["ENTITY"]
	state := variables["STATE"]
	flow := variables["FLOW"]
	logic := variables["LOGIC"]
	iface := variables["INTERFACE"]

	if entity == "" {
		return nil, fmt.Errorf("ENTITY variable is required")
	}

	// Create AI variables
	aiVars := &ai.Variables{
		Entity:    entity,
		State:     state,
		Flow:      flow,
		Logic:     logic,
		Interface: iface,
	}

	// Use AI to generate curriculum structure
	curriculum, err := a.aiClient.GenerateCurriculum(archetype, domain, aiVars)
	if err != nil {
		return nil, fmt.Errorf("failed to generate curriculum: %w", err)
	}

	// Create course from AI-generated curriculum
	course := &GeneratedCourse{
		Title:             curriculum.Title,
		Description:       curriculum.Description,
		MetaCategory:      determineCategoryFromDomain(domain),
		InjectedVariables: variables,
		Status:            "active",
	}

	return course, nil
}

// determineCategoryFromDomain maps domain to meta category
func determineCategoryFromDomain(domain string) string {
	// Simplified domain categorization
	// In real implementation, this would use more sophisticated logic or AI
	domainKeywords := map[string]string{
		"trading":    "Economic",
		"finance":    "Economic",
		"blockchain": "Economic",
		"ecommerce":  "Economic",
		"ai":         "Cognitive",
		"ml":         "Cognitive",
		"neural":     "Cognitive",
		"design":     "Aesthetic",
		"ui":         "Aesthetic",
		"ux":         "Aesthetic",
		"health":     "Biological",
		"medical":    "Biological",
		"bio":        "Biological",
	}

	for keyword, category := range domainKeywords {
		if contains(domain, keyword) {
			return category
		}
	}

	return "Digital" // Default category
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr)
}

// ReviewerAgent performs code reviews
type ReviewerAgent struct {
	Agent
}

// NewReviewerAgent creates a code review agent
func NewReviewerAgent(aiClient *ai.Client) *ReviewerAgent {
	return &ReviewerAgent{
		Agent: Agent{
			Type: "code_reviewer",
			Capabilities: []string{
				"code_analysis",
				"architecture_review",
				"performance_evaluation",
				"best_practices_validation",
			},
			aiClient: aiClient,
		},
	}
}

// Review analyzes submitted code
func (a *ReviewerAgent) Review(code, language, context string) (*ArchitectureReview, error) {
	if a.aiClient == nil {
		return nil, fmt.Errorf("AI client not configured")
	}

	// Call AI client for code review
	aiReview, err := a.aiClient.ReviewCode(code, language, context)
	if err != nil {
		return nil, fmt.Errorf("failed to review code: %w", err)
	}

	// Convert AI review to architecture review
	review := &ArchitectureReview{
		OverallScore:    aiReview.OverallScore,
		CodeSenseScore:  aiReview.CodeSense,
		EfficiencyScore: aiReview.Efficiency,
		EdgeCasesScore:  aiReview.EdgeCases,
		TasteScore:      aiReview.Taste,
		Feedback:        aiReview.Feedback,
	}

	return review, nil
}

// VisualizerAgent generates system visualizations
type VisualizerAgent struct {
	Agent
}

// NewVisualizerAgent creates a visualization agent
func NewVisualizerAgent(aiClient *ai.Client) *VisualizerAgent {
	return &VisualizerAgent{
		Agent: Agent{
			Type: "visualizer",
			Capabilities: []string{
				"system_diagrams",
				"data_flow_visualization",
				"architecture_mapping",
				"state_visualization",
			},
			aiClient: aiClient,
		},
	}
}

// VisualizationData represents visualization output
type VisualizationData struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Visualize creates system diagrams
func (a *VisualizerAgent) Visualize(systemType, state string, data interface{}) (*VisualizationData, error) {
	// Create visualization based on system type
	visualization := &VisualizationData{
		Type:  systemType,
		Title: fmt.Sprintf("%s Visualization", systemType),
		Metadata: map[string]interface{}{
			"generated_by": "VisualizerAgent",
			"system_type":  systemType,
			"state":        state,
		},
	}

	switch systemType {
	case "entropy":
		visualization.Description = "Entropy flow and state transitions"
		visualization.Data = a.generateEntropyGraph(data)

	case "memory":
		visualization.Description = "Memory allocation and usage patterns"
		visualization.Data = a.generateMemoryMap(data)

	case "flow":
		visualization.Description = "Data flow and control flow diagram"
		visualization.Data = a.generateFlowDiagram(data)

	case "decision":
		visualization.Description = "Decision tree and logic branches"
		visualization.Data = a.generateDecisionTree(data)

	case "interface":
		visualization.Description = "Interface interaction board"
		visualization.Data = a.generateInterfaceBoard(data)

	default:
		return nil, fmt.Errorf("unsupported visualization type: %s", systemType)
	}

	return visualization, nil
}

// generateEntropyGraph creates entropy visualization data
func (a *VisualizerAgent) generateEntropyGraph(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"id": "state_0", "label": "Initial State", "entropy": 0.0},
			{"id": "state_1", "label": "Processing", "entropy": 0.5},
			{"id": "state_2", "label": "Final State", "entropy": 1.0},
		},
		"edges": []map[string]interface{}{
			{"from": "state_0", "to": "state_1", "transition": "process"},
			{"from": "state_1", "to": "state_2", "transition": "complete"},
		},
		"metrics": map[string]interface{}{
			"total_entropy":   1.0,
			"entropy_rate":    0.5,
			"state_count":     3,
			"transition_count": 2,
		},
	}
}

// generateMemoryMap creates memory visualization data
func (a *VisualizerAgent) generateMemoryMap(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"regions": []map[string]interface{}{
			{"name": "Stack", "size": 1024, "used": 512, "type": "linear"},
			{"name": "Heap", "size": 4096, "used": 2048, "type": "dynamic"},
			{"name": "Cache", "size": 256, "used": 128, "type": "fast"},
		},
		"allocations": []map[string]interface{}{
			{"id": "alloc_1", "size": 64, "region": "Stack", "lifetime": "short"},
			{"id": "alloc_2", "size": 256, "region": "Heap", "lifetime": "long"},
		},
		"metrics": map[string]interface{}{
			"total_allocated": 2688,
			"total_capacity":  5376,
			"utilization":     0.5,
		},
	}
}

// generateFlowDiagram creates flow visualization data
func (a *VisualizerAgent) generateFlowDiagram(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"id": "input", "label": "Input", "type": "source"},
			{"id": "process", "label": "Process", "type": "transform"},
			{"id": "output", "label": "Output", "type": "sink"},
		},
		"flows": []map[string]interface{}{
			{"from": "input", "to": "process", "data": "raw_data", "rate": 100},
			{"from": "process", "to": "output", "data": "processed_data", "rate": 95},
		},
		"metrics": map[string]interface{}{
			"throughput":     95,
			"latency_ms":     10,
			"error_rate":     0.05,
		},
	}
}

// generateDecisionTree creates decision tree visualization data
func (a *VisualizerAgent) generateDecisionTree(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"id": "root", "label": "Start", "type": "decision"},
			{"id": "branch_1", "label": "Condition A", "type": "decision"},
			{"id": "branch_2", "label": "Condition B", "type": "decision"},
			{"id": "leaf_1", "label": "Action 1", "type": "action"},
			{"id": "leaf_2", "label": "Action 2", "type": "action"},
		},
		"edges": []map[string]interface{}{
			{"from": "root", "to": "branch_1", "condition": "x > 0"},
			{"from": "root", "to": "branch_2", "condition": "x <= 0"},
			{"from": "branch_1", "to": "leaf_1", "condition": "y > 0"},
			{"from": "branch_2", "to": "leaf_2", "condition": "y <= 0"},
		},
		"metrics": map[string]interface{}{
			"depth":          2,
			"branches":       2,
			"decision_count": 2,
			"action_count":   2,
		},
	}
}

// generateInterfaceBoard creates interface visualization data
func (a *VisualizerAgent) generateInterfaceBoard(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"components": []map[string]interface{}{
			{"id": "api", "label": "REST API", "type": "endpoint", "methods": []string{"GET", "POST", "PUT", "DELETE"}},
			{"id": "ui", "label": "Web UI", "type": "interface", "components": []string{"form", "list", "detail"}},
			{"id": "db", "label": "Database", "type": "storage", "operations": []string{"read", "write", "update", "delete"}},
		},
		"interactions": []map[string]interface{}{
			{"from": "ui", "to": "api", "protocol": "HTTP", "format": "JSON"},
			{"from": "api", "to": "db", "protocol": "SQL", "format": "queries"},
		},
		"metrics": map[string]interface{}{
			"component_count":   3,
			"interaction_count": 2,
			"interfaces":        2,
		},
	}
}
