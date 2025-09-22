package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/backsaas/platform/coverage-reporter/internal/collector"
	"github.com/backsaas/platform/coverage-reporter/internal/config"
	"github.com/backsaas/platform/coverage-reporter/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	storage   storage.Storage
	collector *collector.Collector
	config    *config.Config
	upgrader  websocket.Upgrader
}

// New creates a new API server
func New(storage storage.Storage, collector *collector.Collector, config *config.Config) *Server {
	return &Server{
		storage:   storage,
		collector: collector,
		config:    config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	router := s.setupRoutes()
	return router.Run(addr)
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Serve static files (dashboard)
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*.html")

	// Dashboard routes
	router.GET("/", s.handleDashboard)
	router.GET("/service/:name", s.handleServiceDetail)

	// API routes
	api := router.Group("/api")
	{
		api.GET("/summary", s.handleSummary)
		api.GET("/services", s.handleServices)
		api.GET("/services/:name", s.handleServiceCoverage)
		api.GET("/services/:name/history", s.handleServiceHistory)
		api.POST("/collect", s.handleCollectAll)
		api.POST("/collect/:name", s.handleCollectService)
		api.GET("/status", s.handleStatus)
		api.GET("/ws", s.handleWebSocket)
		
		// Integration test endpoints
		api.POST("/integration-tests", s.HandleIntegrationTestReport)
		api.GET("/integration-tests", s.HandleGetIntegrationTests)
	}

	return router
}

// Dashboard handlers
func (s *Server) handleDashboard(c *gin.Context) {
	summary, err := s.storage.GetSummary()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	allCoverage, err := s.storage.GetAllCoverage()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"summary":  summary,
		"services": allCoverage,
		"config":   s.config,
	})
}

func (s *Server) handleServiceDetail(c *gin.Context) {
	serviceName := c.Param("name")

	coverage, err := s.storage.GetCoverage(serviceName)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Service not found: " + serviceName,
		})
		return
	}

	history, err := s.storage.GetCoverageHistory(serviceName, 20)
	if err != nil {
		history = []*storage.CoverageData{}
	}

	c.HTML(http.StatusOK, "service.html", gin.H{
		"service":  serviceName,
		"coverage": coverage,
		"history":  history,
	})
}

// API handlers
func (s *Server) handleSummary(c *gin.Context) {
	summary, err := s.storage.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (s *Server) handleServices(c *gin.Context) {
	allCoverage, err := s.storage.GetAllCoverage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allCoverage)
}

func (s *Server) handleServiceCoverage(c *gin.Context) {
	serviceName := c.Param("name")

	coverage, err := s.storage.GetCoverage(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, coverage)
}

func (s *Server) handleServiceHistory(c *gin.Context) {
	serviceName := c.Param("name")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	history, err := s.storage.GetCoverageHistory(serviceName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (s *Server) handleCollectAll(c *gin.Context) {
	go func() {
		if err := s.collector.CollectAll(); err != nil {
			// Log error but don't fail the request
			gin.DefaultWriter.Write([]byte("Collection error: " + err.Error() + "\n"))
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Coverage collection started for all services",
	})
}

func (s *Server) handleCollectService(c *gin.Context) {
	serviceName := c.Param("name")

	if s.collector.IsCollecting(serviceName) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Collection already in progress for service: " + serviceName,
		})
		return
	}

	go func() {
		if err := s.collector.CollectService(serviceName); err != nil {
			// Log error but don't fail the request
			gin.DefaultWriter.Write([]byte("Collection error: " + err.Error() + "\n"))
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Coverage collection started for service: " + serviceName,
	})
}

func (s *Server) handleStatus(c *gin.Context) {
	services := s.collector.GetServices()
	status := make(map[string]interface{})

	for _, service := range services {
		status[service.Name] = map[string]interface{}{
			"collecting": s.collector.IsCollecting(service.Name),
			"priority":   service.Priority,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now(),
		"services":  status,
		"uptime":    time.Since(time.Now()), // This would be calculated from start time
	})
}

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Send periodic updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			summary, err := s.storage.GetSummary()
			if err != nil {
				continue
			}

			if err := conn.WriteJSON(summary); err != nil {
				return
			}
		}
	}
}
