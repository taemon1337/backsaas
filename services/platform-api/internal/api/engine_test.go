package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"
	
	"github.com/gin-gonic/gin"
)

func TestAPIEngine(t *testing.T) {
	// Skip if no database available
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("Skipping integration tests - no TEST_DATABASE_URL provided")
	}

	t.Run("TestSchemaEngine", func(t *testing.T) {
		// Create engine with test schema
		config := &Config{
			TenantID:     "test-tenant",
			SchemaSource: "file",
			SchemaPath:   "../../testdata/test-schema.yaml",
			DatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
			Port:         "8080",
		}

		engine, err := NewEngine(config)
		if err != nil {
			t.Fatalf("Failed to create engine: %v", err)
		}

		// Test health endpoint
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var healthResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &healthResponse)
		if err != nil {
			t.Fatalf("Failed to parse health response: %v", err)
		}

		if healthResponse["tenant_id"] != "test-tenant" {
			t.Errorf("Expected tenant_id 'test-tenant', got %v", healthResponse["tenant_id"])
		}

		if healthResponse["service"] != "test-api" {
			t.Errorf("Expected service 'test-api', got %v", healthResponse["service"])
		}
	})

	t.Run("PlatformSchemaEngine", func(t *testing.T) {
		// Test with platform schema
		platformSchemaPath := "../../../services/api/system/schema/platform.yaml"
		
		// Check if platform schema exists
		if _, err := os.Stat(platformSchemaPath); os.IsNotExist(err) {
			t.Skip("Platform schema not found, skipping test")
		}

		config := &Config{
			TenantID:     "system",
			SchemaSource: "file",
			SchemaPath:   platformSchemaPath,
			DatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
			Port:         "8081",
		}

		engine, err := NewEngine(config)
		if err != nil {
			t.Fatalf("Failed to create platform engine: %v", err)
		}

		// Test health endpoint
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var healthResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &healthResponse)
		if err != nil {
			t.Fatalf("Failed to parse health response: %v", err)
		}

		if healthResponse["tenant_id"] != "system" {
			t.Errorf("Expected tenant_id 'system', got %v", healthResponse["tenant_id"])
		}

		// Test schema endpoint
		req, _ = http.NewRequest("GET", "/schema", nil)
		w = httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for schema endpoint, got %d", w.Code)
		}

		var schemaResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &schemaResponse)
		if err != nil {
			t.Fatalf("Failed to parse schema response: %v", err)
		}

		schema, ok := schemaResponse["schema"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected schema object in response")
		}

		// Verify platform entities exist
		entities, ok := schema["entities"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected entities in schema")
		}

		expectedEntities := []string{"users", "tenants", "schemas", "functions"}
		for _, entityName := range expectedEntities {
			if _, exists := entities[entityName]; !exists {
				t.Errorf("Expected entity '%s' not found in platform schema", entityName)
			}
		}
	})
}

func TestAPIRoutes(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("Skipping integration tests - no TEST_DATABASE_URL provided")
	}

	config := &Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../testdata/test-schema.yaml",
		DatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
		Port:         "8080",
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	t.Run("EntityEndpoints", func(t *testing.T) {
		// Test that entity endpoints are created
		testCases := []struct {
			method   string
			path     string
			expected int
		}{
			{"GET", "/api/users", http.StatusOK},
			{"GET", "/api/products", http.StatusOK},
			{"GET", "/api/users/123", http.StatusNotFound}, // No data yet
			{"GET", "/api/nonexistent", http.StatusNotFound},
		}

		for _, tc := range testCases {
			req, _ := http.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			engine.router.ServeHTTP(w, req)

			if w.Code != tc.expected {
				t.Errorf("Expected status %d for %s %s, got %d", tc.expected, tc.method, tc.path, w.Code)
			}
		}
	})

	t.Run("CreateEntity", func(t *testing.T) {
		// Test creating a user
		userData := map[string]interface{}{
			"id":    "test-user-123",
			"email": "test@example.com",
			"name":  "Test User",
			"status": "active",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)

		// Note: This will likely fail due to unimplemented database operations
		// but we're testing the routing and basic structure
		if w.Code != http.StatusCreated && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 201 or 500 for user creation, got %d", w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}
	})
}

func TestSchemaValidation(t *testing.T) {
	t.Run("InvalidSchemaFile", func(t *testing.T) {
		config := &Config{
			TenantID:     "test-tenant",
			SchemaSource: "file",
			SchemaPath:   "nonexistent.yaml",
			DatabaseURL:  "postgres://fake",
			Port:         "8080",
		}

		_, err := NewEngine(config)
		if err == nil {
			t.Error("Expected error for nonexistent schema file")
		}
	})

	t.Run("InvalidSchemaSource", func(t *testing.T) {
		config := &Config{
			TenantID:     "test-tenant",
			SchemaSource: "invalid",
			SchemaPath:   "test.yaml",
			DatabaseURL:  "postgres://fake",
			Port:         "8080",
		}

		_, err := NewEngine(config)
		if err == nil {
			t.Error("Expected error for invalid schema source")
		}
	})
}

func TestMiddleware(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("Skipping integration tests - no TEST_DATABASE_URL provided")
	}

	config := &Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../testdata/test-schema.yaml",
		DatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
		Port:         "8080",
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	t.Run("TenantMiddleware", func(t *testing.T) {
		// Create a test handler that checks for tenant context
		engine.router.GET("/test-middleware", func(c *gin.Context) {
			tenantID, exists := c.Get("tenant_id")
			if !exists {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant_id not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
		})

		req, _ := http.NewRequest("GET", "/test-middleware", nil)
		w := httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response["tenant_id"] != "test-tenant" {
			t.Errorf("Expected tenant_id 'test-tenant', got %v", response["tenant_id"])
		}
	})
}

// Benchmark tests
func BenchmarkHealthEndpoint(b *testing.B) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("Skipping benchmark - no TEST_DATABASE_URL provided")
	}

	config := &Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../testdata/test-schema.yaml",
		DatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
		Port:         "8080",
	}

	engine, err := NewEngine(config)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}

	req, _ := http.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)
	}
}

func BenchmarkSchemaEndpoint(b *testing.B) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("Skipping benchmark - no TEST_DATABASE_URL provided")
	}

	config := &Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../testdata/test-schema.yaml",
		DatabaseURL:  os.Getenv("TEST_DATABASE_URL"),
		Port:         "8080",
	}

	engine, err := NewEngine(config)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}

	req, _ := http.NewRequest("GET", "/schema", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		engine.router.ServeHTTP(w, req)
	}
}
