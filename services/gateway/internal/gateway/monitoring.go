package gateway

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MonitoringMiddleware handles monitoring, logging, and metrics
type MonitoringMiddleware struct {
	config  *MonitoringConfig
	metrics *Metrics
}

// Metrics holds various gateway metrics
type Metrics struct {
	mu sync.RWMutex
	
	// Request metrics
	TotalRequests     int64
	RequestsByStatus  map[int]int64
	RequestsByRoute   map[string]int64
	RequestsByTenant  map[string]int64
	
	// Response time metrics
	ResponseTimes     []time.Duration
	AverageResponseTime time.Duration
	
	// Error metrics
	TotalErrors       int64
	ErrorsByType      map[string]int64
	
	// Rate limiting metrics
	RateLimitHits     int64
	RateLimitByTenant map[string]int64
	
	// Backend metrics
	BackendRequests   map[string]int64
	BackendErrors     map[string]int64
	BackendResponseTimes map[string]time.Duration
	
	// System metrics
	StartTime         time.Time
	LastRequestTime   time.Time
}

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware(config *MonitoringConfig) (*MonitoringMiddleware, error) {
	return &MonitoringMiddleware{
		config: config,
		metrics: &Metrics{
			RequestsByStatus:    make(map[int]int64),
			RequestsByRoute:     make(map[string]int64),
			RequestsByTenant:    make(map[string]int64),
			ErrorsByType:        make(map[string]int64),
			RateLimitByTenant:   make(map[string]int64),
			BackendRequests:     make(map[string]int64),
			BackendErrors:       make(map[string]int64),
			BackendResponseTimes: make(map[string]time.Duration),
			StartTime:           time.Now(),
		},
	}, nil
}

// RequestLogger returns middleware for request logging
func (m *MonitoringMiddleware) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.LogRequests {
			c.Next()
			return
		}
		
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// Process request
		c.Next()
		
		// Calculate latency
		latency := time.Since(start)
		
		// Get client info
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		
		// Get user context if available
		userID := ""
		tenantID := ""
		if uid, exists := c.Get("user_id"); exists {
			userID = uid.(string)
		}
		if tid, exists := c.Get("tenant_id"); exists {
			tenantID = tid.(string)
		}
		
		// Get route info if available
		routeDescription := ""
		if route, exists := c.Get("route"); exists {
			routeDescription = route.(*RouteConfig).Description
		}
		
		// Build log entry
		logEntry := map[string]interface{}{
			"timestamp":    start.Format(time.RFC3339),
			"method":       method,
			"path":         path,
			"query":        raw,
			"status":       statusCode,
			"latency_ms":   latency.Milliseconds(),
			"client_ip":    clientIP,
			"user_agent":   c.Request.UserAgent(),
			"user_id":      userID,
			"tenant_id":    tenantID,
			"route":        routeDescription,
			"request_id":   c.GetHeader("X-Request-ID"),
		}
		
		// Add error info if present
		if len(c.Errors) > 0 {
			logEntry["errors"] = c.Errors.String()
		}
		
		// Log based on format
		if m.config.LogFormat == "json" {
			m.logJSON(logEntry)
		} else {
			m.logText(logEntry)
		}
		
		// Update metrics
		m.updateMetrics(statusCode, latency, tenantID, routeDescription)
	}
}

// Metrics returns middleware for metrics collection
func (m *MonitoringMiddleware) Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		// Update metrics after request processing
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		tenantID := ""
		if tid, exists := c.Get("tenant_id"); exists {
			tenantID = tid.(string)
		}
		
		routeDescription := ""
		if route, exists := c.Get("route"); exists {
			routeDescription = route.(*RouteConfig).Description
		}
		
		m.updateMetrics(statusCode, latency, tenantID, routeDescription)
	}
}

// MetricsHandler returns HTTP handler for metrics endpoint
func (m *MonitoringMiddleware) MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.metrics.mu.RLock()
		defer m.metrics.mu.RUnlock()
		
		// Calculate uptime
		uptime := time.Since(m.metrics.StartTime)
		
		// Calculate average response time
		avgResponseTime := m.calculateAverageResponseTime()
		
		// Build metrics response
		metrics := gin.H{
			"gateway": gin.H{
				"uptime_seconds":        uptime.Seconds(),
				"start_time":           m.metrics.StartTime.Format(time.RFC3339),
				"last_request_time":    m.metrics.LastRequestTime.Format(time.RFC3339),
			},
			"requests": gin.H{
				"total":                m.metrics.TotalRequests,
				"by_status":           m.metrics.RequestsByStatus,
				"by_route":            m.metrics.RequestsByRoute,
				"by_tenant":           m.metrics.RequestsByTenant,
				"average_response_time_ms": avgResponseTime.Milliseconds(),
			},
			"errors": gin.H{
				"total":     m.metrics.TotalErrors,
				"by_type":   m.metrics.ErrorsByType,
			},
			"rate_limiting": gin.H{
				"total_hits":  m.metrics.RateLimitHits,
				"by_tenant":   m.metrics.RateLimitByTenant,
			},
			"backends": gin.H{
				"requests":       m.metrics.BackendRequests,
				"errors":         m.metrics.BackendErrors,
				"response_times": m.formatBackendResponseTimes(),
			},
		}
		
		c.JSON(http.StatusOK, metrics)
	}
}

// updateMetrics updates internal metrics
func (m *MonitoringMiddleware) updateMetrics(statusCode int, latency time.Duration, tenantID, route string) {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()
	
	// Update request metrics
	m.metrics.TotalRequests++
	m.metrics.RequestsByStatus[statusCode]++
	m.metrics.LastRequestTime = time.Now()
	
	if route != "" {
		m.metrics.RequestsByRoute[route]++
	}
	
	if tenantID != "" {
		m.metrics.RequestsByTenant[tenantID]++
	}
	
	// Update response time metrics
	m.metrics.ResponseTimes = append(m.metrics.ResponseTimes, latency)
	
	// Keep only last 1000 response times to prevent memory growth
	if len(m.metrics.ResponseTimes) > 1000 {
		m.metrics.ResponseTimes = m.metrics.ResponseTimes[len(m.metrics.ResponseTimes)-1000:]
	}
	
	// Update error metrics
	if statusCode >= 400 {
		m.metrics.TotalErrors++
		
		errorType := "client_error"
		if statusCode >= 500 {
			errorType = "server_error"
		}
		m.metrics.ErrorsByType[errorType]++
	}
}

// RecordRateLimitHit records a rate limit hit
func (m *MonitoringMiddleware) RecordRateLimitHit(tenantID string) {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()
	
	m.metrics.RateLimitHits++
	if tenantID != "" {
		m.metrics.RateLimitByTenant[tenantID]++
	}
}

// RecordBackendRequest records a backend request
func (m *MonitoringMiddleware) RecordBackendRequest(backendURL string, latency time.Duration, success bool) {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()
	
	m.metrics.BackendRequests[backendURL]++
	m.metrics.BackendResponseTimes[backendURL] = latency
	
	if !success {
		m.metrics.BackendErrors[backendURL]++
	}
}

// calculateAverageResponseTime calculates average response time
func (m *MonitoringMiddleware) calculateAverageResponseTime() time.Duration {
	if len(m.metrics.ResponseTimes) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, rt := range m.metrics.ResponseTimes {
		total += rt
	}
	
	return total / time.Duration(len(m.metrics.ResponseTimes))
}

// formatBackendResponseTimes formats backend response times for JSON
func (m *MonitoringMiddleware) formatBackendResponseTimes() map[string]int64 {
	formatted := make(map[string]int64)
	for backend, duration := range m.metrics.BackendResponseTimes {
		formatted[backend] = duration.Milliseconds()
	}
	return formatted
}

// logJSON logs in JSON format
func (m *MonitoringMiddleware) logJSON(entry map[string]interface{}) {
	// In production, use a proper structured logger like logrus or zap
	fmt.Printf(`{"level":"info","type":"request","data":%+v}`+"\n", entry)
}

// logText logs in text format
func (m *MonitoringMiddleware) logText(entry map[string]interface{}) {
	fmt.Printf("[%s] %s %s %d %dms - %s %s\n",
		entry["timestamp"],
		entry["method"],
		entry["path"],
		entry["status"],
		entry["latency_ms"],
		entry["client_ip"],
		entry["user_agent"],
	)
}

// GetMetrics returns current metrics (for external monitoring systems)
func (m *MonitoringMiddleware) GetMetrics() *Metrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()
	
	// Return a copy to prevent concurrent access issues
	metricsCopy := &Metrics{
		TotalRequests:       m.metrics.TotalRequests,
		RequestsByStatus:    make(map[int]int64),
		RequestsByRoute:     make(map[string]int64),
		RequestsByTenant:    make(map[string]int64),
		TotalErrors:         m.metrics.TotalErrors,
		ErrorsByType:        make(map[string]int64),
		RateLimitHits:       m.metrics.RateLimitHits,
		RateLimitByTenant:   make(map[string]int64),
		BackendRequests:     make(map[string]int64),
		BackendErrors:       make(map[string]int64),
		BackendResponseTimes: make(map[string]time.Duration),
		StartTime:           m.metrics.StartTime,
		LastRequestTime:     m.metrics.LastRequestTime,
	}
	
	// Copy maps
	for k, v := range m.metrics.RequestsByStatus {
		metricsCopy.RequestsByStatus[k] = v
	}
	for k, v := range m.metrics.RequestsByRoute {
		metricsCopy.RequestsByRoute[k] = v
	}
	for k, v := range m.metrics.RequestsByTenant {
		metricsCopy.RequestsByTenant[k] = v
	}
	for k, v := range m.metrics.ErrorsByType {
		metricsCopy.ErrorsByType[k] = v
	}
	for k, v := range m.metrics.RateLimitByTenant {
		metricsCopy.RateLimitByTenant[k] = v
	}
	for k, v := range m.metrics.BackendRequests {
		metricsCopy.BackendRequests[k] = v
	}
	for k, v := range m.metrics.BackendErrors {
		metricsCopy.BackendErrors[k] = v
	}
	for k, v := range m.metrics.BackendResponseTimes {
		metricsCopy.BackendResponseTimes[k] = v
	}
	
	return metricsCopy
}

// ResetMetrics resets all metrics (useful for testing)
func (m *MonitoringMiddleware) ResetMetrics() {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()
	
	m.metrics.TotalRequests = 0
	m.metrics.RequestsByStatus = make(map[int]int64)
	m.metrics.RequestsByRoute = make(map[string]int64)
	m.metrics.RequestsByTenant = make(map[string]int64)
	m.metrics.ResponseTimes = []time.Duration{}
	m.metrics.TotalErrors = 0
	m.metrics.ErrorsByType = make(map[string]int64)
	m.metrics.RateLimitHits = 0
	m.metrics.RateLimitByTenant = make(map[string]int64)
	m.metrics.BackendRequests = make(map[string]int64)
	m.metrics.BackendErrors = make(map[string]int64)
	m.metrics.BackendResponseTimes = make(map[string]time.Duration)
	m.metrics.StartTime = time.Now()
}
