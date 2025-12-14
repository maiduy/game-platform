package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestCounter counts the number of HTTP requests
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration measures the duration of HTTP requests
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// ResponseSize measures the size of HTTP responses
	ResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status"},
	)

	// ActiveRequests measures the number of active HTTP requests
	ActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_requests",
			Help: "Number of active HTTP requests",
		},
	)

	// DatabaseQueryDuration measures the duration of database queries
	DatabaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query", "operation"},
	)

	// CacheHits counts the number of cache hits
	CacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	// CacheMisses counts the number of cache misses
	CacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)

// Initialize registers all metrics with the default registry
func Initialize() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ResponseSize)
	prometheus.MustRegister(ActiveRequests)
	prometheus.MustRegister(DatabaseQueryDuration)
	prometheus.MustRegister(CacheHits)
	prometheus.MustRegister(CacheMisses)
}

// MetricsMiddleware returns a middleware that collects metrics for HTTP requests
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Increment active requests
		ActiveRequests.Inc()

		// Process request
		c.Next()

		// Decrement active requests
		ActiveRequests.Dec()

		// Record metrics
		status := fmt.Sprintf("%d", c.Writer.Status())
		duration := time.Since(start).Seconds()
		responseSize := float64(c.Writer.Size())

		RequestCounter.WithLabelValues(method, path, status).Inc()
		RequestDuration.WithLabelValues(method, path, status).Observe(duration)
		ResponseSize.WithLabelValues(method, path, status).Observe(responseSize)
	}
}

// RegisterMetricsEndpoint registers the metrics endpoint with the given router
func RegisterMetricsEndpoint(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// ObserveDatabaseQuery records the duration of a database query
func ObserveDatabaseQuery(query string, operation string, duration time.Duration) {
	DatabaseQueryDuration.WithLabelValues(query, operation).Observe(duration.Seconds())
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	CacheHits.Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	CacheMisses.Inc()
}

// HealthHandler returns a handler for the health endpoint
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
