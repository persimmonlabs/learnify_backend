package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// HealthMonitor continuously monitors database health
type HealthMonitor struct {
	db                *DB
	interval          time.Duration
	alertThresholds   HealthThresholds
	metrics           *HealthMetrics
	stopChan          chan struct{}
	wg                sync.WaitGroup
	alertCallbacks    []AlertCallback
	mu                sync.RWMutex
}

// HealthThresholds defines alert thresholds
type HealthThresholds struct {
	MaxIdleConnPct        float64       // Alert if idle connections exceed this percentage
	MinOpenConns          int           // Alert if open connections drop below this
	MaxConnectionWaitTime time.Duration // Alert if waiting for connection exceeds this
	PingTimeout           time.Duration // Timeout for ping operations
	QueryTimeout          time.Duration // Timeout for test queries
}

// HealthMetrics holds current health metrics
type HealthMetrics struct {
	Timestamp           time.Time     `json:"timestamp"`
	Healthy             bool          `json:"healthy"`
	OpenConnections     int           `json:"open_connections"`
	InUse               int           `json:"in_use"`
	Idle                int           `json:"idle"`
	WaitCount           int64         `json:"wait_count"`
	WaitDuration        time.Duration `json:"wait_duration"`
	MaxIdleClosed       int64         `json:"max_idle_closed"`
	MaxLifetimeClosed   int64         `json:"max_lifetime_closed"`
	MaxIdleTimeClosed   int64         `json:"max_idle_time_closed"`
	PingLatency         time.Duration `json:"ping_latency"`
	QueryLatency        time.Duration `json:"query_latency"`
	LastError           string        `json:"last_error,omitempty"`
	mu                  sync.RWMutex
}

// AlertCallback is called when health alerts are triggered
type AlertCallback func(alert HealthAlert)

// HealthAlert represents a health alert
type HealthAlert struct {
	Severity  string    `json:"severity"` // "warning", "critical"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Metrics   HealthMetrics `json:"metrics"`
}

// DefaultHealthThresholds returns recommended health monitoring thresholds
func DefaultHealthThresholds() HealthThresholds {
	return HealthThresholds{
		MaxIdleConnPct:        80.0,
		MinOpenConns:          1,
		MaxConnectionWaitTime: 1 * time.Second,
		PingTimeout:           2 * time.Second,
		QueryTimeout:          3 * time.Second,
	}
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(db *DB, interval time.Duration, thresholds HealthThresholds) *HealthMonitor {
	return &HealthMonitor{
		db:              db,
		interval:        interval,
		alertThresholds: thresholds,
		metrics:         &HealthMetrics{},
		stopChan:        make(chan struct{}),
		alertCallbacks:  make([]AlertCallback, 0),
	}
}

// RegisterAlertCallback registers a callback for health alerts
func (hm *HealthMonitor) RegisterAlertCallback(callback AlertCallback) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.alertCallbacks = append(hm.alertCallbacks, callback)
}

// Start begins health monitoring
func (hm *HealthMonitor) Start() {
	hm.wg.Add(1)
	go hm.monitor()
	log.Printf("Health monitor started with %v interval", hm.interval)
}

// Stop stops health monitoring
func (hm *HealthMonitor) Stop() {
	close(hm.stopChan)
	hm.wg.Wait()
	log.Println("Health monitor stopped")
}

// GetMetrics returns current health metrics
func (hm *HealthMonitor) GetMetrics() HealthMetrics {
	hm.metrics.mu.RLock()
	defer hm.metrics.mu.RUnlock()
	return *hm.metrics
}

// monitor runs the health check loop
func (hm *HealthMonitor) monitor() {
	defer hm.wg.Done()

	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	// Run initial check
	hm.performHealthCheck()

	for {
		select {
		case <-ticker.C:
			hm.performHealthCheck()
		case <-hm.stopChan:
			return
		}
	}
}

// performHealthCheck executes a comprehensive health check
func (hm *HealthMonitor) performHealthCheck() {
	metrics := &HealthMetrics{
		Timestamp: time.Now(),
		Healthy:   true,
	}

	// Collect pool statistics
	stats := hm.db.Stats()
	metrics.OpenConnections = stats.OpenConnections
	metrics.InUse = stats.InUse
	metrics.Idle = stats.Idle
	metrics.WaitCount = stats.WaitCount
	metrics.WaitDuration = stats.WaitDuration
	metrics.MaxIdleClosed = stats.MaxIdleClosed
	metrics.MaxLifetimeClosed = stats.MaxLifetimeClosed
	metrics.MaxIdleTimeClosed = stats.MaxIdleTimeClosed

	// Test ping with latency measurement
	pingStart := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), hm.alertThresholds.PingTimeout)
	if err := hm.db.PingContext(ctx); err != nil {
		metrics.Healthy = false
		metrics.LastError = fmt.Sprintf("ping failed: %v", err)
		hm.triggerAlert("critical", metrics.LastError, metrics)
	}
	cancel()
	metrics.PingLatency = time.Since(pingStart)

	// Test simple query with latency measurement
	queryStart := time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), hm.alertThresholds.QueryTimeout)
	var result int
	if err := hm.db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		metrics.Healthy = false
		if metrics.LastError == "" {
			metrics.LastError = fmt.Sprintf("query test failed: %v", err)
		}
		hm.triggerAlert("critical", fmt.Sprintf("query test failed: %v", err), metrics)
	}
	cancel()
	metrics.QueryLatency = time.Since(queryStart)

	// Check thresholds
	hm.checkThresholds(metrics)

	// Update stored metrics
	hm.metrics.mu.Lock()
	*hm.metrics = *metrics
	hm.metrics.mu.Unlock()
}

// checkThresholds evaluates metrics against configured thresholds
func (hm *HealthMonitor) checkThresholds(metrics *HealthMetrics) {
	// Check idle connection percentage
	if metrics.OpenConnections > 0 {
		idlePct := float64(metrics.Idle) / float64(metrics.OpenConnections) * 100
		if idlePct > hm.alertThresholds.MaxIdleConnPct {
			hm.triggerAlert("warning",
				fmt.Sprintf("High idle connection percentage: %.1f%% (threshold: %.1f%%)",
					idlePct, hm.alertThresholds.MaxIdleConnPct),
				metrics)
		}
	}

	// Check minimum open connections
	if metrics.OpenConnections < hm.alertThresholds.MinOpenConns {
		hm.triggerAlert("critical",
			fmt.Sprintf("Low open connections: %d (minimum: %d)",
				metrics.OpenConnections, hm.alertThresholds.MinOpenConns),
			metrics)
	}

	// Check connection wait time
	if metrics.WaitCount > 0 {
		avgWaitTime := metrics.WaitDuration / time.Duration(metrics.WaitCount)
		if avgWaitTime > hm.alertThresholds.MaxConnectionWaitTime {
			hm.triggerAlert("warning",
				fmt.Sprintf("High connection wait time: %v (threshold: %v)",
					avgWaitTime, hm.alertThresholds.MaxConnectionWaitTime),
				metrics)
		}
	}

	// Check ping latency
	if metrics.PingLatency > hm.alertThresholds.PingTimeout/2 {
		hm.triggerAlert("warning",
			fmt.Sprintf("High ping latency: %v", metrics.PingLatency),
			metrics)
	}

	// Check query latency
	if metrics.QueryLatency > hm.alertThresholds.QueryTimeout/2 {
		hm.triggerAlert("warning",
			fmt.Sprintf("High query latency: %v", metrics.QueryLatency),
			metrics)
	}
}

// triggerAlert sends alerts to registered callbacks
func (hm *HealthMonitor) triggerAlert(severity, message string, metrics *HealthMetrics) {
	alert := HealthAlert{
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Metrics:   *metrics,
	}

	log.Printf("[%s] Database health alert: %s", severity, message)

	hm.mu.RLock()
	callbacks := make([]AlertCallback, len(hm.alertCallbacks))
	copy(callbacks, hm.alertCallbacks)
	hm.mu.RUnlock()

	for _, callback := range callbacks {
		go callback(alert)
	}
}

// RecycleStaleConnections closes and recreates stale connections
func (hm *HealthMonitor) RecycleStaleConnections(ctx context.Context) error {
	// Force close idle connections by setting MaxIdleConns to 0 temporarily
	stats := hm.db.Stats()
	currentMaxIdle := stats.MaxIdleClosed

	hm.db.SetMaxIdleConns(0)
	time.Sleep(100 * time.Millisecond)

	// Restore original setting
	hm.db.SetMaxIdleConns(int(currentMaxIdle))

	// Verify connection is still healthy
	if err := hm.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to verify connection after recycling: %w", err)
	}

	log.Println("Successfully recycled stale database connections")
	return nil
}

// GetConnectionPoolStatus returns a human-readable pool status
func (hm *HealthMonitor) GetConnectionPoolStatus() string {
	metrics := hm.GetMetrics()

	status := fmt.Sprintf(`Database Connection Pool Status:
  Healthy: %v
  Open Connections: %d
  In Use: %d
  Idle: %d
  Wait Count: %d
  Wait Duration: %v
  Ping Latency: %v
  Query Latency: %v`,
		metrics.Healthy,
		metrics.OpenConnections,
		metrics.InUse,
		metrics.Idle,
		metrics.WaitCount,
		metrics.WaitDuration,
		metrics.PingLatency,
		metrics.QueryLatency,
	)

	if metrics.LastError != "" {
		status += fmt.Sprintf("\n  Last Error: %s", metrics.LastError)
	}

	return status
}
