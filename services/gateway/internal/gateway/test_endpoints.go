package gateway

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TestEndpointResponse represents the response from test endpoints
type TestEndpointResponse struct {
	Service     string            `json:"service"`
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query_params"`
	Host        string            `json:"host"`
	RemoteAddr  string            `json:"remote_addr"`
	Timestamp   string            `json:"timestamp"`
	Route       *RouteInfo        `json:"route,omitempty"`
}

// RouteInfo contains information about the matched route
type RouteInfo struct {
	Description string `json:"description"`
	PathPrefix  string `json:"path_prefix"`
	BackendURL  string `json:"backend_url"`
}

// setupTestEndpoints adds test and debugging endpoints to the gateway
func (g *Gateway) setupTestEndpoints() {
	// Echo endpoint - reflects request information back to client
	g.router.GET("/echo", g.handleEcho)
	g.router.POST("/echo", g.handleEcho)
	g.router.PUT("/echo", g.handleEcho)
	g.router.DELETE("/echo", g.handleEcho)
	
	// Route test endpoint - shows which route would match a given path
	g.router.GET("/test/route", g.handleRouteTest)
	
	// Gateway health endpoint
	g.router.GET("/test/health", g.handleGatewayHealth)
	
	// Route list endpoint - shows all configured routes
	g.router.GET("/test/routes", g.handleRouteList)
	
	// Debug: log that endpoints were registered
	log.Printf("Test endpoints registered: /echo, /test/route, /test/health, /test/routes")
}

// handleEcho reflects request information back to the client
func (g *Gateway) handleEcho(c *gin.Context) {
	// Extract headers
	headers := make(map[string]string)
	for name, values := range c.Request.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	// Extract query parameters
	queryParams := make(map[string]string)
	for name, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[name] = values[0]
		}
	}

	// Try to match the route (for debugging)
	var routeInfo *RouteInfo
	if g.matcher != nil {
		if route, err := g.matcher.Match(c.Request); err == nil {
			routeInfo = &RouteInfo{
				Description: route.Description,
				PathPrefix:  route.PathPrefix,
				BackendURL:  route.Backend.URL,
			}
		}
	}

	response := TestEndpointResponse{
		Service:     "gateway-echo",
		Path:        c.Request.URL.Path,
		Method:      c.Request.Method,
		Headers:     headers,
		QueryParams: queryParams,
		Host:        c.Request.Host,
		RemoteAddr:  c.ClientIP(),
		Timestamp:   time.Now().Format(time.RFC3339),
		Route:       routeInfo,
	}

	c.Header("X-Gateway-Echo", "true")
	c.JSON(http.StatusOK, response)
}

// handleRouteTest tests which route would match a given path
func (g *Gateway) handleRouteTest(c *gin.Context) {
	testPath := c.Query("path")
	if testPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	// Create a test request
	testReq, err := http.NewRequest("GET", testPath, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	testReq.Host = c.Request.Host

	// Try to match the route
	if g.matcher == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "route matcher not initialized"})
		return
	}

	route, err := g.matcher.Match(testReq)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"path":  testPath,
			"match": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":  testPath,
		"match": true,
		"route": RouteInfo{
			Description: route.Description,
			PathPrefix:  route.PathPrefix,
			BackendURL:  route.Backend.URL,
		},
	})
}

// handleGatewayHealth returns gateway health information
func (g *Gateway) handleGatewayHealth(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"routes":    len(g.config.Routes),
	}

	// Test Redis connection if available
	if g.redisClient != nil {
		ctx := c.Request.Context()
		if err := g.redisClient.Ping(ctx).Err(); err != nil {
			health["redis"] = "unhealthy: " + err.Error()
		} else {
			health["redis"] = "healthy"
		}
	}

	c.JSON(http.StatusOK, health)
}

// handleRouteList returns all configured routes
func (g *Gateway) handleRouteList(c *gin.Context) {
	routes := make([]RouteInfo, len(g.config.Routes))
	for i, route := range g.config.Routes {
		routes[i] = RouteInfo{
			Description: route.Description,
			PathPrefix:  route.PathPrefix,
			BackendURL:  route.Backend.URL,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"routes": routes,
		"count":  len(routes),
	})
}
