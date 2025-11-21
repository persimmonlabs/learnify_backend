package identity

import (
	"backend/internal/platform/ai"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CourseGenerator defines the interface for generating courses
type CourseGenerator interface {
	GenerateCourse(userID, archetypeID string, variables map[string]string) error
}

// Service handles identity business logic
type Service struct {
	repo            *Repository
	jwtSecret       string
	aiClient        *ai.Client
	courseGenerator CourseGenerator
}

// NewService creates a new identity service
func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// WithAIClient adds AI client to the service
func (s *Service) WithAIClient(aiClient *ai.Client) *Service {
	s.aiClient = aiClient
	return s
}

// WithCourseGenerator adds course generator to the service
func (s *Service) WithCourseGenerator(generator CourseGenerator) *Service {
	s.courseGenerator = generator
	return s
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Custom JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Register creates a new user account
func (s *Service) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate email format
	if !emailRegex.MatchString(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// Validate password complexity
	if err := validatePasswordComplexity(req.Password); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now()
	user := &User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		AvatarURL:    "",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLogin:    now,
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash in response
	user.PasswordHash = ""

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login authenticates a user
func (s *Service) Login(req *LoginRequest) (*AuthResponse, error) {
	// Find user by email
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Update last login
	user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()
	err = s.repo.UpdateUser(user)
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("warning: failed to update last login: %v\n", err)
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Don't return password hash in response
	user.PasswordHash = ""

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetProfile retrieves user profile
func (s *Service) GetProfile(userID string) (*User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Don't return password hash
	user.PasswordHash = ""
	return user, nil
}

// UpdateProfile updates user profile
func (s *Service) UpdateProfile(userID string, updates map[string]interface{}) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	if avatarURL, ok := updates["avatar_url"].(string); ok {
		user.AvatarURL = avatarURL
	}

	user.UpdatedAt = time.Now()

	err = s.repo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// CompleteOnboarding saves onboarding results
func (s *Service) CompleteOnboarding(userID, metaCategory, domain, skillLevel string, variables map[string]string) error {
	// Validate user exists
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Create archetype
	now := time.Now()
	archetype := &UserArchetype{
		ID:           uuid.New().String(),
		UserID:       userID,
		MetaCategory: metaCategory,
		Domain:       domain,
		SkillLevel:   skillLevel,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = s.repo.CreateArchetype(archetype)
	if err != nil {
		return fmt.Errorf("failed to create archetype: %w", err)
	}

	// Create variables
	if len(variables) > 0 {
		userVariables := make([]UserVariable, 0, len(variables))
		for key, value := range variables {
			userVariables = append(userVariables, UserVariable{
				ID:            uuid.New().String(),
				UserID:        userID,
				VariableKey:   key,
				VariableValue: value,
				ArchetypeID:   archetype.ID,
				CreatedAt:     now,
			})
		}

		err = s.repo.CreateVariables(userVariables)
		if err != nil {
			return fmt.Errorf("failed to create variables: %w", err)
		}
	}

	// Trigger curriculum generation
	if s.courseGenerator != nil {
		// Use AI client to enhance variables if available
		if s.aiClient != nil && len(variables) == 0 {
			// Extract variables from domain using AI
			aiVars, err := s.aiClient.ExtractVariables(domain)
			if err == nil && aiVars != nil {
				variables = map[string]string{
					"ENTITY":    aiVars.Entity,
					"STATE":     aiVars.State,
					"FLOW":      aiVars.Flow,
					"LOGIC":     aiVars.Logic,
					"INTERFACE": aiVars.Interface,
				}

				// Store AI-extracted variables
				userVariables := make([]UserVariable, 0, len(variables))
				for key, value := range variables {
					userVariables = append(userVariables, UserVariable{
						ID:            uuid.New().String(),
						UserID:        userID,
						VariableKey:   key,
						VariableValue: value,
						ArchetypeID:   archetype.ID,
						CreatedAt:     time.Now(),
					})
				}
				_ = s.repo.CreateVariables(userVariables)
			}
		}

		// Generate initial course
		if err := s.courseGenerator.GenerateCourse(userID, archetype.ID, variables); err != nil {
			// Log error but don't fail onboarding
			fmt.Printf("Warning: Failed to generate course: %v\n", err)
		}
	}

	return nil
}

// generateToken creates a JWT token for the user
func (s *Service) generateToken(userID, email string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetArchetype retrieves user's archetype as interface{} for social domain
func (s *Service) GetArchetype(userID string) (interface{}, error) {
	archetype, err := s.repo.GetArchetypeByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype: %w", err)
	}
	if archetype == nil {
		return nil, nil
	}
	return archetype, nil
}

// validatePasswordComplexity checks password meets security requirements
func validatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return errors.New("password is too long (max 128 characters)")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' ||
			 char == '%' || char == '^' || char == '&' || char == '*' ||
			 char == '(' || char == ')' || char == '-' || char == '_' ||
			 char == '=' || char == '+' || char == '[' || char == ']' ||
			 char == '{' || char == '}' || char == '|' || char == ';' ||
			 char == ':' || char == '\'' || char == '"' || char == '<' ||
			 char == '>' || char == ',' || char == '.' || char == '?' ||
			 char == '/' || char == '~' || char == '`':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	// Check for common weak passwords
	commonWeakPasswords := []string{
		"password", "Password1!", "Welcome1!", "Admin123!",
		"12345678", "Qwerty123!", "Abc12345!", "Password123!",
	}

	for _, weak := range commonWeakPasswords {
		if password == weak {
			return errors.New("password is too common, please choose a stronger password")
		}
	}

	return nil
}
