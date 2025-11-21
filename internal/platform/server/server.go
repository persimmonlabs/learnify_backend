package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server wraps the HTTP server
type Server struct {
	httpServer *http.Server
	router     http.Handler
}

// Config holds server configuration
type Config struct {
	Addr              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	MaxHeaderBytes    int
}

// New creates a new HTTP server instance
func New(cfg Config, router http.Handler) *Server {
	// Set defaults if not provided
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 120 * time.Second
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = 5 * time.Second
	}
	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = 1 << 20 // 1MB
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	return &Server{
		httpServer: srv,
		router:     router,
	}
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	if s.httpServer == nil {
		return fmt.Errorf("server not initialized")
	}

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}

// ShutdownWithContext gracefully stops the server with custom context
func (s *Server) ShutdownWithContext(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}

// Addr returns the server address
func (s *Server) Addr() string {
	if s.httpServer != nil {
		return s.httpServer.Addr
	}
	return ""
}
