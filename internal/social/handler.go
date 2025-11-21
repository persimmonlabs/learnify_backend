package social

import (
	"backend/internal/platform/middleware"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for social domain
type Handler struct {
	service *Service
}

// NewHandler creates a new social handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// FollowUser handles POST /api/users/:id/follow
func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	followingID := vars["id"]

	if followingID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Extract current user from JWT context
	followerID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Follow user
	if err := h.service.FollowUser(followerID, followingID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully followed user",
	})
}

// UnfollowUser handles DELETE /api/users/:id/follow
func (h *Handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	followingID := vars["id"]

	if followingID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Extract current user from JWT context
	followerID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Unfollow user
	if err := h.service.UnfollowUser(followerID, followingID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully unfollowed user",
	})
}

// GetActivityFeed handles GET /api/feed
func (h *Handler) GetActivityFeed(w http.ResponseWriter, r *http.Request) {
	// Extract current user from JWT context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse limit from query params
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	// Get activity feed
	activities, err := h.service.GetActivityFeed(userID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return ticker data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
	})
}

// GetRecommendations handles GET /api/recommendations
func (h *Handler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	// Extract current user from JWT context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get recommendations grouped by type
	recommendations, err := h.service.GetRecommendations(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return Netflix-style rows
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
		"sections": map[string]string{
			"collaborative_filtering": "Because You Completed",
			"skill_adjacency":        "Next Level Skills",
			"social_signal":          "Friends Are Learning",
			"trending":               "Trending Now",
		},
	})
}

// GetTrendingCourses handles GET /api/trending
func (h *Handler) GetTrendingCourses(w http.ResponseWriter, r *http.Request) {
	// Get trending courses
	courses, err := h.service.GetTrendingCourses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trending": courses,
		"count":    len(courses),
	})
}

// GetUserProfile handles GET /api/users/:id/profile
func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get complete user profile data from all domains
	profileData, err := h.service.GetUserProfileData(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return Living Resume data
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profileData)
}

// GetAchievements handles GET /api/users/me/achievements
func (h *Handler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	// Extract current user from JWT context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user achievements
	achievements, err := h.service.CheckAchievements(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"achievements": achievements,
		"count":        len(achievements),
	})
}

// GetFollowers handles GET /api/users/:id/followers
func (h *Handler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get followers
	followers, err := h.service.GetFollowers(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"followers": followers,
		"count":     len(followers),
	})
}

// GetFollowing handles GET /api/users/:id/following
func (h *Handler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	vars := mux.Vars(r)
	userID := vars["id"]

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get following
	following, err := h.service.GetFollowing(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"following": following,
		"count":     len(following),
	})
}

// RefreshRecommendations handles POST /api/recommendations/refresh
func (h *Handler) RefreshRecommendations(w http.ResponseWriter, r *http.Request) {
	// Extract current user from JWT context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate new recommendations
	if err := h.service.GenerateRecommendations(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Recommendations refreshed successfully",
	})
}

// RefreshTrending handles POST /api/trending/refresh (admin only)
func (h *Handler) RefreshTrending(w http.ResponseWriter, r *http.Request) {
	// Check admin authorization from context
	isAdmin, ok := r.Context().Value("is_admin").(bool)
	if !ok || !isAdmin {
		http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
		return
	}

	// Refresh trending cache
	if err := h.service.RefreshTrendingCache(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Trending cache refreshed successfully",
	})
}

// RegisterRoutes registers all social routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Follow/Unfollow
	r.HandleFunc("/api/users/{id}/follow", h.FollowUser).Methods("POST")
	r.HandleFunc("/api/users/{id}/follow", h.UnfollowUser).Methods("DELETE")
	r.HandleFunc("/api/users/{id}/followers", h.GetFollowers).Methods("GET")
	r.HandleFunc("/api/users/{id}/following", h.GetFollowing).Methods("GET")

	// Activity Feed
	r.HandleFunc("/api/feed", h.GetActivityFeed).Methods("GET")

	// Recommendations
	r.HandleFunc("/api/recommendations", h.GetRecommendations).Methods("GET")
	r.HandleFunc("/api/recommendations/refresh", h.RefreshRecommendations).Methods("POST")

	// Trending
	r.HandleFunc("/api/trending", h.GetTrendingCourses).Methods("GET")
	r.HandleFunc("/api/trending/refresh", h.RefreshTrending).Methods("POST")

	// Profile
	r.HandleFunc("/api/users/{id}/profile", h.GetUserProfile).Methods("GET")
	r.HandleFunc("/api/users/me/achievements", h.GetAchievements).Methods("GET")
}
