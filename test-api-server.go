package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Simplified schema structures for testing
type PropertyDefinition struct {
	Type      string      `json:"type"`
	Format    string      `json:"format,omitempty"`
	MinLength int         `json:"minLength,omitempty"`
	MaxLength int         `json:"maxLength,omitempty"`
	Default   interface{} `json:"default,omitempty"`
}

type EntitySchema struct {
	Type       string                          `json:"type"`
	Required   []string                        `json:"required,omitempty"`
	Properties map[string]*PropertyDefinition  `json:"properties"`
}

type Entity struct {
	Key    string       `json:"key"`
	Schema EntitySchema `json:"schema"`
}

type Schema struct {
	Version  int                `json:"version"`
	Entities map[string]*Entity `json:"entities"`
}

// Simplified DatabaseOperations using our fixed implementation
type DatabaseOperations struct {
	db       *sql.DB
	tenantID string
}

func NewDatabaseOperations(db *sql.DB, tenantID string) *DatabaseOperations {
	return &DatabaseOperations{
		db:       db,
		tenantID: tenantID,
	}
}

func (d *DatabaseOperations) getEntityColumns(entity *Entity) []string {
	// This is our key fix - deterministic column ordering
	columns := []string{entity.Key, "tenant_id"}
	
	// Add property columns in sorted order (this was the fix!)
	var propertyNames []string
	for name := range entity.Schema.Properties {
		if name != entity.Key { // Skip the key as it's already added
			propertyNames = append(propertyNames, name)
		}
	}
	
	// Sort property names to ensure deterministic order
	for i := 0; i < len(propertyNames); i++ {
		for j := i + 1; j < len(propertyNames); j++ {
			if propertyNames[i] > propertyNames[j] {
				propertyNames[i], propertyNames[j] = propertyNames[j], propertyNames[i]
			}
		}
	}
	
	columns = append(columns, propertyNames...)
	columns = append(columns, "created_at", "updated_at")
	
	return columns
}

func (d *DatabaseOperations) createTableIfNotExists(tableName string, entity *Entity) error {
	columns := d.getEntityColumns(entity)
	
	var columnDefs []string
	for _, col := range columns {
		if col == entity.Key {
			columnDefs = append(columnDefs, fmt.Sprintf("%s VARCHAR(255) PRIMARY KEY", col))
		} else if col == "tenant_id" {
			columnDefs = append(columnDefs, "tenant_id VARCHAR(255) NOT NULL")
		} else if col == "created_at" || col == "updated_at" {
			columnDefs = append(columnDefs, fmt.Sprintf("%s TIMESTAMP DEFAULT CURRENT_TIMESTAMP", col))
		} else {
			// Simplified property to column mapping
			prop := entity.Schema.Properties[col]
			if prop != nil {
				switch prop.Type {
				case "string":
					if prop.MaxLength > 0 {
						columnDefs = append(columnDefs, fmt.Sprintf("%s VARCHAR(%d)", col, prop.MaxLength))
					} else {
						columnDefs = append(columnDefs, fmt.Sprintf("%s VARCHAR(255)", col))
					}
				case "integer":
					columnDefs = append(columnDefs, fmt.Sprintf("%s INTEGER", col))
				case "boolean":
					if prop.Default != nil {
						columnDefs = append(columnDefs, fmt.Sprintf("%s BOOLEAN DEFAULT '%v'", col, prop.Default))
					} else {
						columnDefs = append(columnDefs, fmt.Sprintf("%s BOOLEAN", col))
					}
				default:
					columnDefs = append(columnDefs, fmt.Sprintf("%s TEXT", col))
				}
			}
		}
	}
	
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, 
		fmt.Sprintf("%s", columnDefs[0]))
	for i := 1; i < len(columnDefs); i++ {
		createSQL += ", " + columnDefs[i]
	}
	createSQL += ")"
	
	_, err := d.db.Exec(createSQL)
	return err
}

func (d *DatabaseOperations) insertEntity(tableName string, entity *Entity, data map[string]interface{}) (map[string]interface{}, error) {
	columns := d.getEntityColumns(entity)
	
	// Build INSERT statement with deterministic column order
	var insertCols []string
	var placeholders []string
	var values []interface{}
	
	for _, col := range columns {
		if col == "created_at" || col == "updated_at" {
			continue // Skip audit columns in INSERT
		}
		
		insertCols = append(insertCols, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)+1))
		
		if col == "tenant_id" {
			values = append(values, d.tenantID)
		} else if val, exists := data[col]; exists {
			values = append(values, val)
		} else {
			values = append(values, nil)
		}
	}
	
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
		tableName,
		fmt.Sprintf("%s", insertCols[0]) + func() string {
			result := ""
			for i := 1; i < len(insertCols); i++ {
				result += ", " + insertCols[i]
			}
			return result
		}(),
		fmt.Sprintf("%s", placeholders[0]) + func() string {
			result := ""
			for i := 1; i < len(placeholders); i++ {
				result += ", " + placeholders[i]
			}
			return result
		}(),
		fmt.Sprintf("%s", columns[0]) + func() string {
			result := ""
			for i := 1; i < len(columns); i++ {
				result += ", " + columns[i]
			}
			return result
		}())
	
	// Prepare result scanning
	scanDests := make([]interface{}, len(columns))
	result := make(map[string]interface{})
	
	for i := range columns {
		var val interface{}
		scanDests[i] = &val
	}
	
	err := d.db.QueryRow(insertSQL, values...).Scan(scanDests...)
	if err != nil {
		return nil, err
	}
	
	// Map scanned values back to result
	for i, col := range columns {
		if scanDests[i] != nil {
			val := *(scanDests[i].(*interface{}))
			result[col] = val
		}
	}
	
	return result, nil
}

func (d *DatabaseOperations) queryEntities(tableName string, entity *Entity) ([]map[string]interface{}, error) {
	columns := d.getEntityColumns(entity)
	
	selectSQL := fmt.Sprintf("SELECT %s FROM %s WHERE tenant_id = $1 ORDER BY %s",
		fmt.Sprintf("%s", columns[0]) + func() string {
			result := ""
			for i := 1; i < len(columns); i++ {
				result += ", " + columns[i]
			}
			return result
		}(),
		tableName,
		entity.Key)
	
	rows, err := d.db.Query(selectSQL, d.tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var results []map[string]interface{}
	for rows.Next() {
		scanDests := make([]interface{}, len(columns))
		for i := range columns {
			var val interface{}
			scanDests[i] = &val
		}
		
		err := rows.Scan(scanDests...)
		if err != nil {
			return nil, err
		}
		
		result := make(map[string]interface{})
		for i, col := range columns {
			if scanDests[i] != nil {
				val := *(scanDests[i].(*interface{}))
				result[col] = val
			}
		}
		results = append(results, result)
	}
	
	return results, nil
}

func main() {
	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/backsaas?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Create CRM schema for testing our field mapping fix
	crmSchema := &Schema{
		Version: 1,
		Entities: map[string]*Entity{
			"contacts": {
				Key: "contact_id",
				Schema: EntitySchema{
					Type:     "object",
					Required: []string{"contact_id", "email", "first_name", "last_name"},
					Properties: map[string]*PropertyDefinition{
						"contact_id": {Type: "string"},
						"email":      {Type: "string", Format: "email"},
						"first_name": {Type: "string", MinLength: 1, MaxLength: 50},
						"last_name":  {Type: "string", MinLength: 1, MaxLength: 50},
						"phone":      {Type: "string"},
						"company":    {Type: "string", MaxLength: 100},
						"status":     {Type: "string", Default: "lead"},
					},
				},
			},
		},
	}

	// Create database operations
	dbOps := NewDatabaseOperations(db, "test-api")

	// Create table
	err = dbOps.createTableIfNotExists("contacts", crmSchema.Entities["contacts"])
	if err != nil {
		log.Fatalf("Failed to create contacts table: %v", err)
	}

	fmt.Println("✅ Created contacts table successfully")

	// Set up Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "test-api",
			"tenant_id": "test-api",
			"database":  "connected",
		})
	})

	// Create contact endpoint
	r.POST("/contacts", func(c *gin.Context) {
		var contactData map[string]interface{}
		if err := c.ShouldBindJSON(&contactData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := dbOps.insertEntity("contacts", crmSchema.Entities["contacts"], contactData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, result)
	})

	// List contacts endpoint
	r.GET("/contacts", func(c *gin.Context) {
		results, err := dbOps.queryEntities("contacts", crmSchema.Entities["contacts"])
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"contacts": results,
			"count":    len(results),
		})
	})

	// Schema endpoint
	r.GET("/schema", func(c *gin.Context) {
		c.JSON(http.StatusOK, crmSchema)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	fmt.Printf("🚀 Test API server starting on port %s\n", port)
	fmt.Printf("📊 Available endpoints:\n")
	fmt.Printf("  GET  /health   - Health check\n")
	fmt.Printf("  GET  /schema   - API schema\n")
	fmt.Printf("  POST /contacts - Create contact\n")
	fmt.Printf("  GET  /contacts - List contacts\n")
	fmt.Printf("\n💡 Test field mapping fix with:\n")
	fmt.Printf("  curl -X POST http://localhost:%s/contacts -H 'Content-Type: application/json' -d '{\"contact_id\":\"test-1\",\"email\":\"test@example.com\",\"first_name\":\"John\",\"last_name\":\"Doe\",\"phone\":\"+1234567890\",\"company\":\"Acme Corp\",\"status\":\"lead\"}'\n", port)

	log.Fatal(r.Run(":" + port))
}
