package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProxyMiddleware handles request proxying to backend services
type ProxyMiddleware struct {
	client *http.Client
}

// NewProxyMiddleware creates a new proxy middleware
func NewProxyMiddleware() (*ProxyMiddleware, error) {
	return &ProxyMiddleware{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}, nil
}

// ProxyRequest proxies the request to the backend service
func (p *ProxyMiddleware) ProxyRequest(c *gin.Context, route *RouteConfig) {
	// Select backend URL (for load balancing)
	backendURL := p.selectBackendURL(route)
	
	// Parse backend URL
	target, err := url.Parse(backendURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid backend URL",
		})
		return
	}
	
	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Customize the proxy director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		p.modifyRequest(req, route, c)
	}
	
	// Customize error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		p.handleProxyError(c, err, route)
	}
	
	// Customize response modifier
	proxy.ModifyResponse = func(resp *http.Response) error {
		return p.modifyResponse(resp, route, c)
	}
	
	// Set timeout for this specific route
	if route.Backend.Timeout > 0 {
		p.client.Timeout = route.Backend.Timeout
	}
	
	// Proxy the request
	proxy.ServeHTTP(c.Writer, c.Request)
}

// selectBackendURL selects a backend URL for load balancing
func (p *ProxyMiddleware) selectBackendURL(route *RouteConfig) string {
	// If only one URL, return it
	if route.Backend.URL != "" && len(route.Backend.URLs) == 0 {
		return route.Backend.URL
	}
	
	// If multiple URLs, implement load balancing
	urls := route.Backend.URLs
	if route.Backend.URL != "" {
		urls = append([]string{route.Backend.URL}, urls...)
	}
	
	if len(urls) == 0 {
		return ""
	}
	
	// Simple round-robin for now
	// In production, you'd want more sophisticated load balancing
	index := int(time.Now().UnixNano()) % len(urls)
	return urls[index]
}

// modifyRequest modifies the outgoing request
func (p *ProxyMiddleware) modifyRequest(req *http.Request, route *RouteConfig, c *gin.Context) {
	// Add standard headers
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", p.getScheme(c))
	req.Header.Set("X-Forwarded-Host", c.Request.Host)
	
	// Add user context headers if available
	if userID, exists := c.Get("user_id"); exists {
		req.Header.Set("X-User-ID", userID.(string))
	}
	
	if tenantID, exists := c.Get("tenant_id"); exists {
		req.Header.Set("X-Tenant-ID", tenantID.(string))
	}
	
	if userRoles, exists := c.Get("user_roles"); exists {
		roles := userRoles.([]string)
		req.Header.Set("X-User-Roles", strings.Join(roles, ","))
	}
	
	// Add request ID for tracing
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	} else {
		// Generate request ID if not present
		requestID = p.generateRequestID()
		req.Header.Set("X-Request-ID", requestID)
		c.Header("X-Request-ID", requestID)
	}
	
	// Remove hop-by-hop headers
	p.removeHopByHopHeaders(req.Header)
	
	// Apply route-specific transformations (already applied by transform middleware)
	// This is where you could add additional backend-specific modifications
}

// modifyResponse modifies the response from backend
func (p *ProxyMiddleware) modifyResponse(resp *http.Response, route *RouteConfig, c *gin.Context) error {
	// Add security headers
	resp.Header.Set("X-Content-Type-Options", "nosniff")
	resp.Header.Set("X-Frame-Options", "DENY")
	resp.Header.Set("X-XSS-Protection", "1; mode=block")
	
	// Add gateway identification
	resp.Header.Set("X-Gateway", "BackSaas-Gateway/1.0")
	
	// Remove backend-specific headers that shouldn't be exposed
	resp.Header.Del("Server")
	resp.Header.Del("X-Powered-By")
	
	// Remove hop-by-hop headers
	p.removeHopByHopHeaders(resp.Header)
	
	return nil
}

// handleProxyError handles errors during proxying
func (p *ProxyMiddleware) handleProxyError(c *gin.Context, err error, route *RouteConfig) {
	// Log the error (in production, use structured logging)
	fmt.Printf("Proxy error for route %s: %v\n", route.Description, err)
	
	// Determine error type and response
	var statusCode int
	var message string
	
	if strings.Contains(err.Error(), "timeout") {
		statusCode = http.StatusGatewayTimeout
		message = "Backend service timeout"
	} else if strings.Contains(err.Error(), "connection refused") {
		statusCode = http.StatusBadGateway
		message = "Backend service unavailable"
	} else {
		statusCode = http.StatusBadGateway
		message = "Backend service error"
	}
	
	// Check if we should retry with another backend
	if p.shouldRetry(route, err) {
		// Implement retry logic here
		// For now, just return error
	}
	
	c.JSON(statusCode, gin.H{
		"error":   "Gateway Error",
		"message": message,
		"code":    "BACKEND_ERROR",
	})
}

// shouldRetry determines if request should be retried
func (p *ProxyMiddleware) shouldRetry(route *RouteConfig, err error) bool {
	// Implement retry logic based on error type and route configuration
	// For now, return false
	return false
}

// removeHopByHopHeaders removes headers that shouldn't be forwarded
func (p *ProxyMiddleware) removeHopByHopHeaders(headers http.Header) {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}
	
	for _, header := range hopByHopHeaders {
		headers.Del(header)
	}
}

// getScheme determines the request scheme
func (p *ProxyMiddleware) getScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	
	// Check X-Forwarded-Proto header (from load balancer)
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	
	return "http"
}

// generateRequestID generates a unique request ID
func (p *ProxyMiddleware) generateRequestID() string {
	// Simple request ID generation
	// In production, use a more robust method (UUID, etc.)
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// HealthCheckBackend checks if a backend is healthy
func (p *ProxyMiddleware) HealthCheckBackend(backendURL, healthPath string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(backendURL + healthPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// ProxyWebSocket handles WebSocket proxying (for real-time features)
func (p *ProxyMiddleware) ProxyWebSocket(c *gin.Context, route *RouteConfig) {
	// WebSocket proxy implementation would go here
	// This is more complex and would require gorilla/websocket or similar
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "WebSocket proxying not yet implemented",
	})
}
