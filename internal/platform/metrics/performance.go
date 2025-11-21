package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Memory Metrics
	memoryAllocBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_alloc_bytes",
			Help: "Current bytes allocated and in use",
		},
	)

	memoryTotalAllocBytes = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "memory_total_alloc_bytes",
			Help: "Cumulative bytes allocated",
		},
	)

	memorySysBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_sys_bytes",
			Help: "Total bytes obtained from system",
		},
	)

	// Goroutine Metrics
	goroutinesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_count",
			Help: "Number of goroutines currently running",
		},
	)

	// GC Metrics
	gcPauseDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gc_pause_duration_seconds",
			Help:    "GC pause duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	gcRunsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gc_runs_total",
			Help: "Total number of GC runs",
		},
	)

	// SLI/SLO Metrics
	sliLatencyBucket = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "sli_latency_seconds",
			Help: "Request latency for SLI tracking",
			Buckets: []float64{
				0.001, // 1ms
				0.005, // 5ms
				0.010, // 10ms
				0.050, // 50ms
				0.100, // 100ms
				0.250, // 250ms
				0.500, // 500ms
				1.000, // 1s
				2.500, // 2.5s
				5.000, // 5s
			},
		},
		[]string{"endpoint", "method"},
	)

	sliAvailability = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sli_availability_total",
			Help: "Request count for availability SLI (5xx errors vs total)",
		},
		[]string{"status_class"}, // 2xx, 3xx, 4xx, 5xx
	)
)

func init() {
	// Register performance metrics
	prometheus.MustRegister(
		memoryAllocBytes,
		memoryTotalAllocBytes,
		memorySysBytes,
		goroutinesCount,
		gcPauseDuration,
		gcRunsTotal,
		sliLatencyBucket,
		sliAvailability,
	)
}

// CollectRuntimeMetrics collects Go runtime metrics
func CollectRuntimeMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryAllocBytes.Set(float64(m.Alloc))
	memoryTotalAllocBytes.Add(float64(m.TotalAlloc))
	memorySysBytes.Set(float64(m.Sys))
	goroutinesCount.Set(float64(runtime.NumGoroutine()))

	// GC metrics
	if m.NumGC > 0 {
		// Record most recent GC pause
		gcPauseDuration.Observe(float64(m.PauseNs[(m.NumGC+255)%256]) / 1e9)
		gcRunsTotal.Add(float64(m.NumGC))
	}
}

// StartPerformanceMetricsCollector starts a background goroutine to collect runtime metrics
func StartPerformanceMetricsCollector(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			CollectRuntimeMetrics()
		}
	}()
}

// RecordSLILatency records request latency for SLI tracking
func RecordSLILatency(endpoint, method string, duration time.Duration) {
	sliLatencyBucket.WithLabelValues(endpoint, method).Observe(duration.Seconds())
}

// RecordSLIAvailability records availability SLI (based on status code)
func RecordSLIAvailability(statusCode int) {
	var statusClass string
	switch {
	case statusCode >= 200 && statusCode < 300:
		statusClass = "2xx"
	case statusCode >= 300 && statusCode < 400:
		statusClass = "3xx"
	case statusCode >= 400 && statusCode < 500:
		statusClass = "4xx"
	case statusCode >= 500:
		statusClass = "5xx"
	default:
		statusClass = "unknown"
	}

	sliAvailability.WithLabelValues(statusClass).Inc()
}

// Helper functions for SLO calculations

// CalculateAvailabilitySLO calculates availability percentage (successful requests / total requests)
// Target SLO example: 99.9% availability means < 0.1% 5xx errors
func CalculateAvailabilitySLO(successful, total float64) float64 {
	if total == 0 {
		return 100.0
	}
	return (successful / total) * 100.0
}

// CalculateLatencySLO calculates percentage of requests under target latency
// Target SLO example: 95% of requests complete in < 500ms
func CalculateLatencySLO(underTarget, total float64) float64 {
	if total == 0 {
		return 100.0
	}
	return (underTarget / total) * 100.0
}
