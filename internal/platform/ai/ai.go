package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client wraps AI service clients (OpenAI, Anthropic, etc.)
type Client struct {
	provider   string
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// New creates a new AI client
func New(provider, apiKey, model string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if model == "" {
		model = "gpt-4" // default model
	}

	baseURL := "https://api.openai.com/v1"
	if provider == "anthropic" {
		baseURL = "https://api.anthropic.com/v1"
	} else if provider == "openrouter" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	return &Client{
		provider: provider,
		apiKey:   apiKey,
		model:    model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: baseURL,
	}, nil
}

// ValidateDomain validates user domain input using LLM
func (c *Client) ValidateDomain(domain string, metaCategory string) (*DomainValidation, error) {
	prompt := fmt.Sprintf(`You are a domain validation expert. Determine if the following domain is valid and real for learning purposes.

Domain: %s
Meta Category: %s

Evaluate if this is:
1. A real, legitimate subject/domain that can be learned
2. Specific enough to create a learning curriculum
3. Not offensive, illegal, or inappropriate

Respond in JSON format:
{
  "is_valid": true/false,
  "reason": "explanation why it's valid or invalid"
}`, domain, metaCategory)

	response, err := c.complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to validate domain: %w", err)
	}

	var validation DomainValidation
	if err := json.Unmarshal([]byte(response), &validation); err != nil {
		// Fallback parsing if JSON is not perfect
		validation = DomainValidation{
			IsValid: strings.Contains(strings.ToLower(response), "true"),
			Reason:  response,
		}
	}

	return &validation, nil
}

// ExtractVariables extracts the 5 universal variables from domain
func (c *Client) ExtractVariables(domain string) (*Variables, error) {
	prompt := fmt.Sprintf(`You are an expert at analyzing learning domains. Extract the 5 universal variables from this domain:

Domain: %s

Extract:
1. ENTITY - Core subject/concept (what is being learned)
2. STATE - Current proficiency level or status
3. FLOW - Learning progression or methodology
4. LOGIC - Reasoning patterns and problem-solving approach
5. INTERFACE - How learner interacts with the material

Respond in JSON format:
{
  "entity": "description",
  "state": "description",
  "flow": "description",
  "logic": "description",
  "interface": "description"
}`, domain)

	response, err := c.complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract variables: %w", err)
	}

	var variables Variables
	if err := json.Unmarshal([]byte(response), &variables); err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	return &variables, nil
}

// GenerateCurriculum generates personalized curriculum
func (c *Client) GenerateCurriculum(archetype, domain string, variables *Variables) (*Curriculum, error) {
	prompt := fmt.Sprintf(`You are an expert curriculum designer. Create a personalized learning curriculum.

Learner Archetype: %s
Domain: %s
Variables:
- Entity: %s
- State: %s
- Flow: %s
- Logic: %s
- Interface: %s

Create a structured curriculum with 5-8 modules. Each module should build on previous ones.

Respond in JSON format:
{
  "title": "curriculum title",
  "description": "brief overview",
  "modules": [
    {
      "number": 1,
      "title": "module title",
      "description": "what will be learned"
    }
  ]
}`, archetype, domain, variables.Entity, variables.State, variables.Flow, variables.Logic, variables.Interface)

	response, err := c.complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate curriculum: %w", err)
	}

	var curriculum Curriculum
	if err := json.Unmarshal([]byte(response), &curriculum); err != nil {
		return nil, fmt.Errorf("failed to parse curriculum: %w", err)
	}

	return &curriculum, nil
}

// ReviewCode performs AI Senior Review on submitted code
func (c *Client) ReviewCode(code, language, context string) (*ArchitectureReview, error) {
	prompt := fmt.Sprintf(`You are a senior software architect. Review this code submission.

Language: %s
Context: %s

Code:
%s

Score the following categories from 1-10:
1. CODE SENSE - Code quality, readability, structure
2. EFFICIENCY - Performance, optimization, resource usage
3. EDGE CASES - Error handling, boundary conditions
4. TASTE - Design patterns, best practices, elegance

Respond in JSON format:
{
  "code_sense": 8,
  "efficiency": 7,
  "edge_cases": 6,
  "taste": 9,
  "feedback": {
    "code_sense": "detailed feedback",
    "efficiency": "detailed feedback",
    "edge_cases": "detailed feedback",
    "taste": "detailed feedback"
  }
}`, language, context, code)

	response, err := c.complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to review code: %w", err)
	}

	var review ArchitectureReview
	if err := json.Unmarshal([]byte(response), &review); err != nil {
		return nil, fmt.Errorf("failed to parse review: %w", err)
	}

	// Calculate overall score
	review.OverallScore = (review.CodeSense + review.Efficiency + review.EdgeCases + review.Taste) / 4

	return &review, nil
}

// complete sends a completion request to the AI API
func (c *Client) complete(prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return result.Choices[0].Message.Content, nil
}

// DomainValidation represents domain validation result
type DomainValidation struct {
	IsValid bool   `json:"is_valid"`
	Reason  string `json:"reason"`
}

// Variables represents the 5 universal variables
type Variables struct {
	Entity    string `json:"entity"`
	State     string `json:"state"`
	Flow      string `json:"flow"`
	Logic     string `json:"logic"`
	Interface string `json:"interface"`
}

// Curriculum represents generated course structure
type Curriculum struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Modules     []Module `json:"modules"`
}

// Module represents a course module
type Module struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ArchitectureReview represents AI code review
type ArchitectureReview struct {
	OverallScore int               `json:"overall_score"`
	CodeSense    int               `json:"code_sense"`
	Efficiency   int               `json:"efficiency"`
	EdgeCases    int               `json:"edge_cases"`
	Taste        int               `json:"taste"`
	Feedback     map[string]string `json:"feedback"`
}
