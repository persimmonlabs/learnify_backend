package learning

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for learning domain
type Handler struct {
	service *Service
}

// NewHandler creates a new learning handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all learning routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Course routes
	r.HandleFunc("/api/courses", h.GetCourses).Methods("GET")
	r.HandleFunc("/api/courses/{id}", h.GetCourseDetails).Methods("GET")
	r.HandleFunc("/api/courses/{id}/progress", h.GetProgress).Methods("GET")

	// Exercise routes
	r.HandleFunc("/api/exercises/{id}", h.GetExercise).Methods("GET")
	r.HandleFunc("/api/exercises/{id}/submit", h.SubmitExercise).Methods("POST")

	// Review routes
	r.HandleFunc("/api/submissions/{id}/review", h.RequestReview).Methods("POST")
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// getUserID extracts user ID from JWT context
// In a real implementation, this would extract from JWT middleware context
func getUserID(r *http.Request) string {
	// Placeholder - in real implementation, extract from JWT context
	// For now, check for X-User-ID header (for testing)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		// Default test user
		return "00000000-0000-0000-0000-000000000001"
	}
	return userID
}

// GetCourses handles GET /api/courses
func (h *Handler) GetCourses(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	courses, err := h.service.GetUserCourses(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    courses,
	})
}

// GetCourseDetails handles GET /api/courses/:id
func (h *Handler) GetCourseDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["id"]

	if courseID == "" {
		writeError(w, http.StatusBadRequest, "Course ID is required")
		return
	}

	course, modules, err := h.service.GetCourseDetails(courseID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	response := map[string]interface{}{
		"course":  course,
		"modules": modules,
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// GetExercise handles GET /api/exercises/:id
func (h *Handler) GetExercise(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exerciseID := vars["id"]

	if exerciseID == "" {
		writeError(w, http.StatusBadRequest, "Exercise ID is required")
		return
	}

	exercise, err := h.service.GetExercise(exerciseID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    exercise,
	})
}

// SubmitExerciseRequest represents exercise submission request
type SubmitExerciseRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// SubmitExercise handles POST /api/exercises/:id/submit
func (h *Handler) SubmitExercise(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exerciseID := vars["id"]

	if exerciseID == "" {
		writeError(w, http.StatusBadRequest, "Exercise ID is required")
		return
	}

	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req SubmitExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "Code is required")
		return
	}

	if req.Language == "" {
		writeError(w, http.StatusBadRequest, "Language is required")
		return
	}

	// Submit exercise
	completion, err := h.service.SubmitExercise(userID, exerciseID, req.Code, req.Language)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    completion,
	})
}

// RequestReview handles POST /api/submissions/:id/review
func (h *Handler) RequestReview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	submissionID := vars["id"]

	if submissionID == "" {
		writeError(w, http.StatusBadRequest, "Submission ID is required")
		return
	}

	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Request AI review
	review, err := h.service.RequestReview(submissionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    review,
	})
}

// GetProgress handles GET /api/courses/:id/progress
func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["id"]

	if courseID == "" {
		writeError(w, http.StatusBadRequest, "Course ID is required")
		return
	}

	userID := getUserID(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	progress, err := h.service.GetUserProgress(userID, courseID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Data:    progress,
	})
}
