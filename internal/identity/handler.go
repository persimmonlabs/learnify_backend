package identity

import (
	"context"
	"encoding/json"
	"net/http"
)

// Handler handles HTTP requests for identity domain
type Handler struct {
	service *Service
}

// NewHandler creates a new identity handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// UpdateProfileRequest represents profile update payload
type UpdateProfileRequest struct {
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// OnboardingRequest represents onboarding completion payload
type OnboardingRequest struct {
	MetaCategory string            `json:"meta_category"`
	Domain       string            `json:"domain"`
	SkillLevel   string            `json:"skill_level"`
	Variables    map[string]string `json:"variables"`
}

// contextKey is a custom type for context keys
type contextKey string

const userIDKey contextKey = "user_id"

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

// Register handles POST /api/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	authResp, err := h.service.Register(&req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid email format" ||
			err.Error() == "password must be at least 8 characters" {
			status = http.StatusBadRequest
		} else if err.Error() == "email already registered" {
			status = http.StatusConflict
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, authResp)
}

// Login handles POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	authResp, err := h.service.Login(&req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid email or password" {
			status = http.StatusUnauthorized
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, authResp)
}

// GetProfile handles GET /api/users/me
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey)
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetProfile(userID.(string))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// UpdateProfile handles PATCH /api/users/me
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey)
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}

	err := h.service.UpdateProfile(userID.(string), updates)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "profile updated successfully"})
}

// CompleteOnboarding handles POST /api/onboarding/complete
func (h *Handler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey)
	if userID == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req OnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.MetaCategory == "" || req.Domain == "" || req.SkillLevel == "" {
		respondError(w, http.StatusBadRequest, "meta_category, domain, and skill_level are required")
		return
	}

	err := h.service.CompleteOnboarding(
		userID.(string),
		req.MetaCategory,
		req.Domain,
		req.SkillLevel,
		req.Variables,
	)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "onboarding completed successfully"})
}

// WithUserContext adds user ID to request context
func WithUserContext(userID string, r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}
