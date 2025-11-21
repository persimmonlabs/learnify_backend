package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status
type Status string

const (
	StatusUp   Status = "UP"
	StatusDown Status = "DOWN"
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Response represents the full health check response
type Response struct {
	Status    Status        `json:"status"`
	Version   string        `json:"version"`
	Uptime    string        `json:"uptime"`
	Timestamp time.Time     `json:"timestamp"`
	Checks    []HealthCheck `json:"checks,omitempty"`
}

// Config holds health check configuration
type Config struct {
	Version   string
	StartTime time.Time
	DB        *sql.DB
}

// Handler manages health check endpoints
type Handler struct {
	config Config
	mu     sync.RWMutex
}

// NewHandler creates a new health check handler
func NewHandler(cfg Config) *Handler {
	if cfg.StartTime.IsZero() {
		cfg.StartTime = time.Now()
	}
	return &Handler{
		config: cfg,
	}
}

// Liveness returns a simple liveness check (always returns 200 if server is running)
// This is used by orchestrators to determine if the pod/container should be restarted
func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	response := Response{
		Status:    StatusUp,
		Version:   h.config.Version,
		Uptime:    time.Since(h.config.StartTime).String(),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Readiness performs deep health checks to determine if the service is ready to accept traffic
// This checks database connections and other critical dependencies
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := h.performHealthChecks(ctx)

	// Determine overall status
	overallStatus := StatusUp
	for _, check := range checks {
		if check.Status == StatusDown {
			overallStatus = StatusDown
			break
		}
	}

	response := Response{
		Status:    overallStatus,
		Version:   h.config.Version,
		Uptime:    time.Since(h.config.StartTime).String(),
		Timestamp: time.Now(),
		Checks:    checks,
	}

	statusCode := http.StatusOK
	if overallStatus == StatusDown {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// performHealthChecks runs all configured health checks
func (h *Handler) performHealthChecks(ctx context.Context) []HealthCheck {
	var checks []HealthCheck
	var wg sync.WaitGroup

	// Database check
	if h.config.DB != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			check := h.checkDatabase(ctx)
			h.mu.Lock()
			checks = append(checks, check)
			h.mu.Unlock()
		}()
	}

	// Memory check
	wg.Add(1)
	go func() {
		defer wg.Done()
		check := h.checkMemory()
		h.mu.Lock()
		checks = append(checks, check)
		h.mu.Unlock()
	}()

	wg.Wait()
	return checks
}

// checkDatabase verifies database connectivity
func (h *Handler) checkDatabase(ctx context.Context) HealthCheck {
	check := HealthCheck{
		Name:   "database",
		Status: StatusUp,
	}

	if h.config.DB == nil {
		check.Status = StatusDown
		check.Error = "database not configured"
		return check
	}

	if err := h.config.DB.PingContext(ctx); err != nil {
		check.Status = StatusDown
		check.Error = err.Error()
		return check
	}

	// Check connection pool stats
	stats := h.config.DB.Stats()
	if stats.OpenConnections == 0 {
		check.Status = StatusDown
		check.Error = "no open database connections"
	}

	return check
}

// checkMemory verifies memory usage is within acceptable limits
func (h *Handler) checkMemory() HealthCheck {
	check := HealthCheck{
		Name:   "memory",
		Status: StatusUp,
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Alert if allocated memory exceeds 1GB (configurable threshold)
	const maxMemoryMB = 1024
	allocMB := m.Alloc / 1024 / 1024

	if allocMB > maxMemoryMB {
		check.Status = StatusDown
		check.Error = "memory usage exceeds threshold"
	}

	return check
}

// AddCustomCheck allows adding custom health checks (for future extensibility)
type CheckFunc func(ctx context.Context) HealthCheck

var customChecks []CheckFunc

// RegisterCheck registers a custom health check function
func RegisterCheck(fn CheckFunc) {
	customChecks = append(customChecks, fn)
}
