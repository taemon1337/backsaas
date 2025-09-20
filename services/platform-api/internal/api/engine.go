package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/backsaas/platform/services/platform-api/internal/schema"
)

// Engine represents the generic schema-driven API engine
type Engine struct {
	schema   *schema.Schema
	db       *sql.DB
	tenantID string
	router   *gin.Engine
}

// Config holds configuration for the API engine
type Config struct {
	TenantID     string
	SchemaSource string // "file" or "registry"
	SchemaPath   string // file path or tenant ID for registry
	DatabaseURL  string
	Port         string
}

// NewEngine creates a new API engine instance
func NewEngine(config *Config) (*Engine, error) {
	// Load schema
	var schemaObj *schema.Schema
	var err error
	
	loader := schema.NewLoader(".")
	
	switch config.SchemaSource {
	case "file":
		schemaObj, err = loader.LoadFromFile(config.SchemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load schema from file: %w", err)
		}
	case "registry":
		schemaObj, err = loader.LoadFromRegistry(config.SchemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load schema from registry: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid schema source: %s", config.SchemaSource)
	}
	
	// Connect to database
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Create engine
	engine := &Engine{
		schema:   schemaObj,
		db:       db,
		tenantID: config.TenantID,
	}
	
	// Setup router
	if err := engine.setupRouter(); err != nil {
		return nil, fmt.Errorf("failed to setup router: %w", err)
	}
	
	log.Printf("API Engine initialized for tenant: %s, service: %s", 
		config.TenantID, schemaObj.Service.Name)
	
	return engine, nil
}

// setupRouter configures the HTTP router based on the schema
func (e *Engine) setupRouter() error {
	e.router = gin.Default()
	
	// Add middleware
	e.router.Use(e.tenantMiddleware())
	e.router.Use(e.loggingMiddleware())
	
	// Health check endpoint
	e.router.GET("/health", e.healthCheck)
	
	// Schema info endpoint
	e.router.GET("/schema", e.getSchema)
	
	// Generate CRUD endpoints for each entity
	api := e.router.Group("/api")
	for entityName, entity := range e.schema.Entities {
		e.setupEntityRoutes(api, entityName, entity)
	}
	
	return nil
}

// setupEntityRoutes creates CRUD routes for an entity
func (e *Engine) setupEntityRoutes(group *gin.RouterGroup, entityName string, entity *schema.Entity) {
	entityGroup := group.Group("/" + entityName)
	
	// GET /api/{entity} - List entities
	entityGroup.GET("", e.listEntities(entityName, entity))
	
	// POST /api/{entity} - Create entity
	entityGroup.POST("", e.createEntity(entityName, entity))
	
	// GET /api/{entity}/{id} - Get entity by ID
	entityGroup.GET("/:id", e.getEntity(entityName, entity))
	
	// PUT /api/{entity}/{id} - Update entity
	entityGroup.PUT("/:id", e.updateEntity(entityName, entity))
	
	// DELETE /api/{entity}/{id} - Delete entity
	entityGroup.DELETE("/:id", e.deleteEntity(entityName, entity))
}

// tenantMiddleware adds tenant context to requests
func (e *Engine) tenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("tenant_id", e.tenantID)
		c.Set("schema", e.schema)
		c.Next()
	}
}

// loggingMiddleware logs requests
func (e *Engine) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	})
}

// healthCheck returns the health status
func (e *Engine) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"tenant_id": e.tenantID,
		"service":   e.schema.Service.Name,
		"version":   e.schema.Version,
	})
}

// getSchema returns the current schema
func (e *Engine) getSchema(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"schema": e.schema,
	})
}

// listEntities handles GET /api/{entity}
func (e *Engine) listEntities(entityName string, entity *schema.Entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute before_read hooks
		if err := e.executeHooks("before_read", entityName, nil, c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Build query with tenant scoping
		query := fmt.Sprintf("SELECT * FROM %s WHERE tenant_id = $1", entityName)
		
		// Add filters from query parameters
		// TODO: Implement filtering, pagination, sorting
		
		rows, err := e.db.Query(query, e.tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()
		
		// Convert rows to JSON
		results, err := e.rowsToJSON(rows, entity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process results"})
			return
		}
		
		// Execute after_read hooks
		if err := e.executeHooks("after_read", entityName, results, c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data": results,
			"meta": gin.H{
				"count": len(results),
			},
		})
	}
}

// createEntity handles POST /api/{entity}
func (e *Engine) createEntity(entityName string, entity *schema.Entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		
		// Add tenant_id to data
		data["tenant_id"] = e.tenantID
		
		// Validate data against schema
		if err := e.validateEntityData(entity, data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Execute validation functions
		if err := e.executeValidationFunctions(entityName, "before_create", data, c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Execute before_create hooks
		if err := e.executeHooks("before_create", entityName, data, c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Insert into database
		result, err := e.insertEntity(entityName, entity, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entity"})
			return
		}
		
		// Execute after_create hooks (async)
		go func() {
			if err := e.executeHooks("after_create", entityName, result, c); err != nil {
				log.Printf("after_create hook failed: %v", err)
			}
		}()
		
		c.JSON(http.StatusCreated, gin.H{
			"data": result,
		})
	}
}

// getEntity handles GET /api/{entity}/{id}
func (e *Engine) getEntity(entityName string, entity *schema.Entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 AND tenant_id = $2", 
			entityName, entity.Key)
		
		row := e.db.QueryRow(query, id, e.tenantID)
		
		result, err := e.rowToJSON(row, entity)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	}
}

// updateEntity handles PUT /api/{entity}/{id}
func (e *Engine) updateEntity(entityName string, entity *schema.Entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
		
		// Ensure tenant_id and key are not modified
		data["tenant_id"] = e.tenantID
		data[entity.Key] = id
		
		// Validate data against schema
		if err := e.validateEntityData(entity, data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Execute validation functions
		if err := e.executeValidationFunctions(entityName, "before_update", data, c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Execute before_update hooks
		if err := e.executeHooks("before_update", entityName, data, c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Update in database
		result, err := e.updateEntityInDB(entityName, entity, id, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entity"})
			return
		}
		
		// Execute after_update hooks (async)
		go func() {
			if err := e.executeHooks("after_update", entityName, result, c); err != nil {
				log.Printf("after_update hook failed: %v", err)
			}
		}()
		
		c.JSON(http.StatusOK, gin.H{
			"data": result,
		})
	}
}

// deleteEntity handles DELETE /api/{entity}/{id}
func (e *Engine) deleteEntity(entityName string, entity *schema.Entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		
		// Execute before_delete hooks
		if err := e.executeHooks("before_delete", entityName, map[string]interface{}{entity.Key: id}, c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Delete from database
		query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1 AND tenant_id = $2", 
			entityName, entity.Key)
		
		result, err := e.db.Exec(query, id, e.tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete entity"})
			return
		}
		
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entity not found"})
			return
		}
		
		// Execute after_delete hooks (async)
		go func() {
			if err := e.executeHooks("after_delete", entityName, map[string]interface{}{entity.Key: id}, c); err != nil {
				log.Printf("after_delete hook failed: %v", err)
			}
		}()
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Entity deleted successfully",
		})
	}
}

// Start starts the HTTP server
func (e *Engine) Start(port string) error {
	log.Printf("Starting API server on port %s for tenant: %s", port, e.tenantID)
	return e.router.Run(":" + port)
}

// Helper methods will be implemented in separate files
func (e *Engine) executeHooks(trigger, entityName string, data interface{}, c *gin.Context) error {
	// TODO: Implement hook execution
	return nil
}

func (e *Engine) executeValidationFunctions(entityName, trigger string, data map[string]interface{}, c *gin.Context) error {
	// TODO: Implement validation function execution
	return nil
}

func (e *Engine) validateEntityData(entity *schema.Entity, data map[string]interface{}) error {
	// TODO: Implement schema validation
	return nil
}

func (e *Engine) insertEntity(entityName string, entity *schema.Entity, data map[string]interface{}) (map[string]interface{}, error) {
	// TODO: Implement database insert
	return data, nil
}

func (e *Engine) updateEntityInDB(entityName string, entity *schema.Entity, id string, data map[string]interface{}) (map[string]interface{}, error) {
	// TODO: Implement database update
	return data, nil
}

func (e *Engine) rowsToJSON(rows *sql.Rows, entity *schema.Entity) ([]map[string]interface{}, error) {
	// TODO: Implement rows to JSON conversion
	return []map[string]interface{}{}, nil
}

func (e *Engine) rowToJSON(row *sql.Row, entity *schema.Entity) (map[string]interface{}, error) {
	// TODO: Implement row to JSON conversion
	return map[string]interface{}{}, nil
}
