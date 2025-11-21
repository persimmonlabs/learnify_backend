package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"backend/config"
	"backend/internal/identity"
	"backend/internal/learning"
	"backend/internal/platform/ai"
	"backend/internal/platform/database"
	"backend/internal/platform/health"
	"backend/internal/platform/logger"
	"backend/internal/platform/metrics"
	"backend/internal/platform/middleware"
	"backend/internal/platform/server"
	"backend/internal/social"
)

const (
	shutdownTimeout = 15 * time.Second
)

func main() {
	log.Println("Learnify API starting...")

	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Logger
	appLogger := logger.New(cfg.Server.Env)
	if appLogger == nil {
		log.Fatal("Failed to initialize logger")
	}
	appLogger.Info("Logger initialized", "env", cfg.Server.Env)

	// 3. Connect to Database
	// Convert port string to int with proper error handling
	dbPort := 5432 // default port
	if cfg.Database.Port != "" {
		parsedPort, err := strconv.Atoi(cfg.Database.Port)
		if err != nil {
			appLogger.Warn("Invalid database port in config, using default",
				"config_port", cfg.Database.Port,
				"default_port", dbPort,
				"error", err)
		} else {
			dbPort = parsedPort
		}
	}

	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     dbPort,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		log.Fatalf("Database connection failed: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.Error("Failed to close database connection", "error", err)
		}
	}()

	// Verify database connection
	if err := db.Ping(); err != nil {
		appLogger.Error("Database ping failed", "error", err)
		log.Fatalf("Database not responding: %v", err)
	}
	appLogger.Info("Database connected successfully")

	// 4. Initialize AI Client
	aiClient, err := ai.New(cfg.AI.Provider, cfg.AI.APIKey, cfg.AI.Model)
	if err != nil {
		appLogger.Error("Failed to initialize AI client", "error", err)
		log.Fatalf("AI client initialization failed: %v", err)
	}
	appLogger.Info("AI client initialized", "provider", cfg.AI.Provider, "model", cfg.AI.Model)

	// 5. Initialize Repositories
	identityRepo := identity.NewRepository(db.DB)
	learningRepo := learning.NewRepository(db.DB)
	socialRepo := social.NewRepository(db.DB)
	appLogger.Info("Repositories initialized")

	// 6. Initialize Services
	identityService := identity.NewService(identityRepo, cfg.JWT.Secret, cfg.JWT.Expiration)
	learningService := learning.NewService(learningRepo, aiClient)
	socialService := social.NewService(socialRepo)
	appLogger.Info("Services initialized")

	// 7. Initialize Handlers
	identityHandler := identity.NewHandler(identityService)
	learningHandler := learning.NewHandler(learningService)
	socialHandler := social.NewHandler(socialService)
	appLogger.Info("Handlers initialized")

	// 8. Setup Health Check Handler
	healthHandler := health.NewHandler(health.Config{
		Version:   "1.0.0",
		StartTime: time.Now(),
		DB:        db.DB,
	})
	appLogger.Info("Health check handler initialized")

	// 9. Start Background Metric Collectors
	metrics.StartDatabaseMetricsCollector(db.DB, 15*time.Second)
	metrics.StartPerformanceMetricsCollector(10*time.Second)
	appLogger.Info("Metrics collectors started")

	// 10. Setup Router
	router := mux.NewRouter()

	// Health check endpoints (no auth required)
	router.HandleFunc("/health", healthHandler.Liveness).Methods("GET")
	router.HandleFunc("/health/ready", healthHandler.Readiness).Methods("GET")

	// Metrics endpoint (no auth required, can be restricted by firewall/network policy)
	router.Handle("/metrics", metrics.Handler()).Methods("GET")

	// Create API subrouter
	api := router.PathPrefix("/api").Subrouter()

	// Configure security middleware
	rateLimitConfig := middleware.DefaultRateLimiterConfig()
	securityHeadersConfig := middleware.DefaultSecurityHeadersConfig()
	sizeLimitConfig := middleware.DefaultSizeLimitConfig()

	// Auth middleware for protected routes
	authMiddleware := middleware.Auth(cfg.JWT.Secret)

	// Public routes - Authentication (with rate limiting and size limits)
	authRouter := api.PathPrefix("/auth").Subrouter()
	authRouter.Use(middleware.RateLimitAuth(rateLimitConfig))
	authRouter.Use(middleware.RequestSizeLimit(sizeLimitConfig))
	authRouter.HandleFunc("/register", identityHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", identityHandler.Login).Methods("POST")

	// Protected routes - Identity/User Management
	api.Handle("/users/me", authMiddleware(http.HandlerFunc(identityHandler.GetProfile))).Methods("GET")
	api.Handle("/users/me", authMiddleware(http.HandlerFunc(identityHandler.UpdateProfile))).Methods("PATCH")
	api.Handle("/onboarding/complete", authMiddleware(http.HandlerFunc(identityHandler.CompleteOnboarding))).Methods("POST")

	// Protected routes - Learning/Courses
	api.Handle("/courses", authMiddleware(http.HandlerFunc(learningHandler.GetCourses))).Methods("GET")
	api.Handle("/courses/{id}", authMiddleware(http.HandlerFunc(learningHandler.GetCourseDetails))).Methods("GET")
	api.Handle("/courses/{id}/progress", authMiddleware(http.HandlerFunc(learningHandler.GetProgress))).Methods("GET")

	// Protected routes - Exercises
	api.Handle("/exercises/{id}", authMiddleware(http.HandlerFunc(learningHandler.GetExercise))).Methods("GET")
	api.Handle("/exercises/{id}/submit", authMiddleware(http.HandlerFunc(learningHandler.SubmitExercise))).Methods("POST")
	api.Handle("/submissions/{id}/review", authMiddleware(http.HandlerFunc(learningHandler.RequestReview))).Methods("POST")

	// Protected routes - Social/Activity Feed
	api.Handle("/feed", authMiddleware(http.HandlerFunc(socialHandler.GetActivityFeed))).Methods("GET")
	api.Handle("/users/{id}/follow", authMiddleware(http.HandlerFunc(socialHandler.FollowUser))).Methods("POST")
	api.Handle("/users/{id}/follow", authMiddleware(http.HandlerFunc(socialHandler.UnfollowUser))).Methods("DELETE")
	api.Handle("/recommendations", authMiddleware(http.HandlerFunc(socialHandler.GetRecommendations))).Methods("GET")
	api.Handle("/users/{id}/profile", authMiddleware(http.HandlerFunc(socialHandler.GetUserProfile))).Methods("GET")
	api.Handle("/users/me/achievements", authMiddleware(http.HandlerFunc(socialHandler.GetAchievements))).Methods("GET")

	// Public routes - Trending (no auth required)
	api.HandleFunc("/trending", socialHandler.GetTrendingCourses).Methods("GET")

	appLogger.Info("Routes registered")

	// 11. Apply Global Middleware (order matters!)
	// The middleware chain is applied in reverse order (last applied = first executed)

	// Parse CORS allowed origins from config
	var corsMiddleware func(http.Handler) http.Handler
	if cfg.CORS.AllowedOrigins == "*" {
		corsMiddleware = middleware.CORS() // Development: allow all origins
		appLogger.Info("CORS configured (permissive)", "origins", "*")
	} else {
		// Production: strict CORS with specific origins
		origins := strings.Split(cfg.CORS.AllowedOrigins, ",")
		// Trim whitespace from each origin
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		corsMiddleware = middleware.CORSStrict(origins)
		appLogger.Info("CORS configured (strict)", "origins", origins)
	}

	// Apply middleware chain (executed in reverse order)
	// Execution order: Recovery -> RequestID -> Logging -> Security -> Metrics -> SizeLimit -> RateLimit -> CORS
	handler := corsMiddleware(router)                                     // Last: CORS headers
	handler = middleware.RateLimitAPI(rateLimitConfig)(handler)           // Sixth: Rate limiting
	handler = middleware.RequestSizeLimit(sizeLimitConfig)(handler)       // Fifth: Size limits
	handler = middleware.Metrics()(handler)                                // Fourth: Collect metrics
	handler = middleware.SecurityHeaders(securityHeadersConfig)(handler) // Third: Security headers
	handler = middleware.LoggingSimple()(handler)                         // Second: Log with request ID
	handler = middleware.RequestID()(handler)                              // Early: Generate request ID
	handler = middleware.Recovery()(handler)                              // First: Panic recovery (catches everything)
	appLogger.Info("Middleware applied (recovery, request-id, logging, security, metrics, size limits, rate limiting, CORS)")

	// 12. Create and Start Server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	serverConfig := server.Config{
		Addr: addr,
	}
	srv := server.New(serverConfig, handler)

	// Handle graceful shutdown
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		appLogger.Info("Starting HTTP server", "address", addr)
		serverErrors <- srv.Start()
	}()

	// Listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until either server error or shutdown signal
	select {
	case err := <-serverErrors:
		appLogger.Error("Server error", "error", err)
		log.Fatalf("Server failed to start: %v", err)

	case sig := <-shutdown:
		appLogger.Info("Shutdown signal received", "signal", sig)

		// Give outstanding requests time to complete
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown with timeout
		appLogger.Info("Attempting graceful shutdown", "timeout", shutdownTimeout)
		if err := srv.ShutdownWithContext(ctx); err != nil {
			appLogger.Error("Graceful shutdown failed, forcing close", "error", err)
			// Force close if graceful shutdown fails or times out
			if closeErr := srv.Close(); closeErr != nil {
				appLogger.Error("Force close failed", "error", closeErr)
			}
		} else {
			appLogger.Info("Server shutdown complete")
		}
	}
}
