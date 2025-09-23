package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Gateway represents the API gateway
type Gateway struct {
	config      *Config
	router      *gin.Engine
	redisClient *redis.Client
	
	// Middleware components
	auth        *AuthMiddleware
	rateLimit   *RateLimitMiddleware
	proxy       *ProxyMiddleware
	monitoring  *MonitoringMiddleware
	
	// Route matcher
	matcher     *RouteMatcher
}

// NewGateway creates a new gateway instance
func NewGateway(config *Config) (*Gateway, error) {
	// Load full configuration
	fullConfig, err := LoadConfig(config.ConfigPath, config)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	// Setup Redis client
	redisClient, err := setupRedis(fullConfig.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Redis: %w", err)
	}
	
	// Create gateway
	gateway := &Gateway{
		config:      fullConfig,
		redisClient: redisClient,
	}
	
	// Initialize components
	if err := gateway.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}
	
	// Setup router
	if err := gateway.setupRouter(); err != nil {
		return nil, fmt.Errorf("failed to setup router: %w", err)
	}
	
	return gateway, nil
}

// initializeComponents initializes all gateway components
func (g *Gateway) initializeComponents() error {
	var err error
	
	// Initialize auth middleware
	g.auth, err = NewAuthMiddleware(&g.config.Auth, g.redisClient)
	if err != nil {
		return fmt.Errorf("failed to initialize auth middleware: %w", err)
	}
	
	// Initialize rate limit middleware
	g.rateLimit, err = NewRateLimitMiddleware(&g.config.RateLimit, g.redisClient)
	if err != nil {
		return fmt.Errorf("failed to initialize rate limit middleware: %w", err)
	}
	
	// Initialize proxy middleware
	g.proxy, err = NewProxyMiddleware()
	if err != nil {
		return fmt.Errorf("failed to initialize proxy middleware: %w", err)
	}
	
	// Initialize monitoring middleware
	g.monitoring, err = NewMonitoringMiddleware(&g.config.Monitoring)
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring middleware: %w", err)
	}
	
	// Initialize route matcher
	g.matcher, err = NewRouteMatcher(g.config.Routes)
	if err != nil {
		return fmt.Errorf("failed to initialize route matcher: %w", err)
	}
	
	return nil
}

// setupRouter configures the HTTP router
func (g *Gateway) setupRouter() error {
	// Set gin mode based on environment
	if g.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	g.router = gin.New()
	
	// Add recovery middleware
	g.router.Use(gin.Recovery())
	
	// Add monitoring middleware (logging, metrics)
	g.router.Use(g.monitoring.RequestLogger())
	g.router.Use(g.monitoring.Metrics())
	
	// Add CORS middleware if enabled
	if g.config.Cors.Enabled {
		g.router.Use(g.corsMiddleware())
	}
	
	// Add health check endpoint
	g.router.GET(g.config.Monitoring.HealthPath, g.healthCheck)
	
	// Add metrics endpoint
	if g.config.Monitoring.Enabled {
		g.router.GET(g.config.Monitoring.MetricsPath, g.monitoring.MetricsHandler())
	}
	
	// Add test endpoints for debugging and testing (BEFORE NoRoute)
	log.Printf("DEBUG: Environment check - current: '%s', production check: %v", g.config.Environment, g.config.Environment != "production")
	if g.config.Environment != "production" {
		log.Printf("Setting up test endpoints (environment: %s)", g.config.Environment)
		g.setupTestEndpoints()
	} else {
		log.Printf("Skipping test endpoints (production environment: %s)", g.config.Environment)
	}
	
	// WebSocket support is integrated into the main proxy handler for development
	if g.config.Environment == "development" {
		log.Printf("WebSocket support enabled for development (integrated into proxy handler)")
	}
	
	// Add main proxy handler with middleware chain (this catches all unmatched routes)
	g.router.NoRoute(g.proxyHandler())
	
	return nil
}

// proxyHandler creates the main proxy handler with middleware chain
func (g *Gateway) proxyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for WebSocket upgrade first (in development mode)
		if g.config.Environment == "development" && isWebSocketUpgrade(c.Request) {
			g.handleWebSocketUpgrade(c)
			return
		}
		
		// Find matching route
		route, err := g.matcher.Match(c.Request)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No matching route found",
				"path":  c.Request.URL.Path,
			})
			return
		}
		
		// Store route in context for other middleware
		c.Set("route", route)
		
		// Apply route-specific middleware chain
		middlewares := g.buildMiddlewareChain(route)
		
		// Execute middleware chain
		for _, middleware := range middlewares {
			middleware(c)
			if c.IsAborted() {
				return
			}
		}
		
		// If we get here, proxy the request
		g.proxy.ProxyRequest(c, route)
	}
}

// buildMiddlewareChain builds the middleware chain for a route
func (g *Gateway) buildMiddlewareChain(route *RouteConfig) []gin.HandlerFunc {
	var middlewares []gin.HandlerFunc
	
	// 1. Rate limiting (if enabled)
	rateLimitConfig := &g.config.RateLimit
	if route.RateLimit != nil {
		rateLimitConfig = route.RateLimit
	}
	if rateLimitConfig.Enabled {
		middlewares = append(middlewares, g.rateLimit.Handler(rateLimitConfig))
	}
	
	// 2. Authentication (if required)
	authConfig := &g.config.Auth
	if route.Auth != nil {
		authConfig = route.Auth
	}
	if authConfig.Enabled {
		middlewares = append(middlewares, g.auth.Handler(authConfig))
	}
	
	// 3. Request transformation (if configured)
	if route.Transform != nil {
		middlewares = append(middlewares, g.transformRequest(route.Transform))
	}
	
	return middlewares
}

// transformRequest applies request transformations
func (g *Gateway) transformRequest(transform *TransformConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add headers
		for key, value := range transform.AddHeaders {
			c.Request.Header.Set(key, value)
		}
		
		// Remove headers
		for _, header := range transform.RemoveHeaders {
			c.Request.Header.Del(header)
		}
		
		// Rewrite path
		if transform.RewritePath != "" {
			c.Request.URL.Path = transform.RewritePath
		}
		
		c.Next()
		
		// Apply response transformations after proxy
		for key, value := range transform.AddResponseHeaders {
			c.Header(key, value)
		}
		
		for _, header := range transform.RemoveResponseHeaders {
			c.Header(header, "")
		}
	}
}

// corsMiddleware handles CORS
func (g *Gateway) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cors := &g.config.Cors
		
		// Set CORS headers
		if len(cors.AllowedOrigins) > 0 {
			origin := c.Request.Header.Get("Origin")
			for _, allowedOrigin := range cors.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					c.Header("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}
		}
		
		if len(cors.AllowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", joinStrings(cors.AllowedMethods, ", "))
		}
		
		if len(cors.AllowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", joinStrings(cors.AllowedHeaders, ", "))
		}
		
		if len(cors.ExposedHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(cors.ExposedHeaders, ", "))
		}
		
		if cors.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		
		if cors.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", cors.MaxAge))
		}
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// healthCheck handles health check requests
func (g *Gateway) healthCheck(c *gin.Context) {
	// Check Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	redisStatus := "ok"
	if g.redisClient != nil {
		if err := g.redisClient.Ping(ctx).Err(); err != nil {
			redisStatus = "error: " + err.Error()
		}
	} else {
		redisStatus = "not configured"
	}
	
	// Check backend health (sample a few routes)
	backendStatus := g.checkBackendHealth()
	
	status := "ok"
	httpStatus := http.StatusOK
	if redisStatus != "ok" || len(backendStatus["unhealthy"].([]string)) > 0 {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}
	
	c.JSON(httpStatus, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"components": gin.H{
			"redis":    redisStatus,
			"backends": backendStatus,
		},
		"routes_configured": len(g.config.Routes),
	})
}

// checkBackendHealth checks the health of configured backends
func (g *Gateway) checkBackendHealth() gin.H {
	healthy := []string{}
	unhealthy := []string{}
	
	for _, route := range g.config.Routes {
		if !route.Enabled {
			continue
		}
		
		// Check primary backend URL
		if route.Backend.URL != "" {
			if g.isBackendHealthy(route.Backend.URL, route.Backend.HealthCheckPath) {
				healthy = append(healthy, route.Backend.URL)
			} else {
				unhealthy = append(unhealthy, route.Backend.URL)
			}
		}
		
		// Check additional URLs for load balancing
		for _, url := range route.Backend.URLs {
			if g.isBackendHealthy(url, route.Backend.HealthCheckPath) {
				healthy = append(healthy, url)
			} else {
				unhealthy = append(unhealthy, url)
			}
		}
	}
	
	return gin.H{
		"healthy":   healthy,
		"unhealthy": unhealthy,
	}
}

// isBackendHealthy checks if a backend is healthy
func (g *Gateway) isBackendHealthy(baseURL, healthPath string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + healthPath)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// Start starts the gateway server
func (g *Gateway) Start() error {
	log.Printf("Gateway starting on port %s", g.config.Port)
	log.Printf("Configured routes: %d", len(g.config.Routes))
	
	return g.router.Run(":" + g.config.Port)
}

// Shutdown gracefully shuts down the gateway
func (g *Gateway) Shutdown(ctx context.Context) error {
	log.Println("Gateway shutting down...")
	
	// Close Redis connection
	if g.redisClient != nil {
		return g.redisClient.Close()
	}
	
	return nil
}

// setupRedis initializes Redis client
func setupRedis(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}
	
	client := redis.NewClient(opts)
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	return client, nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// handleWebSocketUpgrade handles WebSocket upgrade requests by proxying to the appropriate backend
func (g *Gateway) handleWebSocketUpgrade(c *gin.Context) {
	// Find matching route for WebSocket
	route, err := g.matcher.Match(c.Request)
	if err != nil {
		log.Printf("No route found for WebSocket request: %s", c.Request.URL.Path)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Parse backend URL
	backendURL, err := url.Parse(route.Backend.URL)
	if err != nil {
		log.Printf("Invalid backend URL for WebSocket: %s", route.Backend.URL)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Create reverse proxy for WebSocket
	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	
	// Modify the request for WebSocket forwarding
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = backendURL.Host
		
		// Apply path transformation if configured
		if route.Transform != nil && route.Transform.RewritePath != "" {
			if route.Transform.RewritePath == "/" {
				// Strip the path prefix
				req.URL.Path = strings.TrimPrefix(req.URL.Path, route.PathPrefix)
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}
			} else {
				req.URL.Path = route.Transform.RewritePath
			}
		}

		// Add custom headers
		if route.Transform != nil && route.Transform.AddHeaders != nil {
			for key, value := range route.Transform.AddHeaders {
				req.Header.Set(key, value)
			}
		}

		// Preserve WebSocket headers
		req.Header.Set("X-Forwarded-For", c.ClientIP())
		req.Header.Set("X-Forwarded-Proto", "http")
		req.Header.Set("X-Forwarded-Host", req.Host)
	}

	// Handle WebSocket upgrade
	log.Printf("Proxying WebSocket request: %s -> %s", c.Request.URL.Path, backendURL.String())
	proxy.ServeHTTP(c.Writer, c.Request)
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade request
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}
