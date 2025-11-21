package identity

import (
	"time"
)

// User represents a user account
type User struct {
	ID              string
	Email           string
	PasswordHash    string
	Name            string
	AvatarURL       string
	PrivacySettings *PrivacySettings `json:"privacy_settings,omitempty"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLogin       time.Time
}

// PrivacySettings represents user privacy preferences
type PrivacySettings struct {
	ProfileVisibility     string `json:"profile_visibility"`      // public, friends, private
	ActivityVisibility    string `json:"activity_visibility"`     // public, friends, private
	ProgressVisibility    string `json:"progress_visibility"`     // public, friends, private
	AllowFollowers        bool   `json:"allow_followers"`
	ShowInLeaderboards    bool   `json:"show_in_leaderboards"`
	ShowCompletedCourses  bool   `json:"show_completed_courses"`
}

// UserArchetype represents user's selected archetype
type UserArchetype struct {
	ID           string
	UserID       string
	MetaCategory string
	Domain       string
	SkillLevel   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserVariable represents runtime variable
type UserVariable struct {
	ID            string
	UserID        string
	VariableKey   string
	VariableValue string
	ArchetypeID   string
	CreatedAt     time.Time
}

// RegisterRequest represents user registration payload
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
