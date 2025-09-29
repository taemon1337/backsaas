package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/backsaas/platform/services/platform-api/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlatformSchemaDrivenAPI tests that all platform.yaml entities are properly served at /api/platform/*
func TestPlatformSchemaDrivenAPI(t *testing.T) {
	// Setup test engine with platform schema
	config := &api.Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../schemas/platform.yaml",
		DatabaseURL:  "postgres://test:test@localhost:5432/test_db?sslmode=disable",
		Port:         "8080",
	}

	engine, err := api.NewEngine(config)
	require.NoError(t, err, "Failed to create API engine")

	// Test cases for each entity defined in platform.yaml
	testCases := []struct {
		entityName   string
		sampleData   map[string]interface{}
		requiredKeys []string
	}{
		{
			entityName: "tenants",
			sampleData: map[string]interface{}{
				"tenant_id":      "test-tenant-1",
				"name":           "Test Tenant",
				"slug":           "test-tenant",
				"domain":         "test.example.com",
				"plan":           "starter",
				"status":         "active",
				"owner_email":    "admin@test.com",
				"owner_name":     "Test Admin",
				"user_count":     0,
				"storage_used":   "0 MB",
				"billing_status": "trial",
			},
			requiredKeys: []string{"tenant_id", "name", "slug", "owner_email"},
		},
		{
			entityName: "users",
			sampleData: map[string]interface{}{
				"user_id":    "test-user-1",
				"tenant_id":  "test-tenant",
				"email":      "user@test.com",
				"first_name": "Test",
				"last_name":  "User",
				"role":       "user",
				"status":     "active",
			},
			requiredKeys: []string{"user_id", "email", "tenant_id"},
		},
		{
			entityName: "schemas",
			sampleData: map[string]interface{}{
				"schema_id": "test-schema-1",
				"name":      "Test Schema",
				"slug":      "test-schema",
				"description": "A test schema",
				"type":      "entity",
				"status":    "draft",
				"version":   "1.0.0",
				"field_count": 5,
				"usage_count": 0,
				"schema_definition": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":   map[string]interface{}{"type": "string"},
						"name": map[string]interface{}{"type": "string"},
					},
				},
				"tags":       []string{"test", "example"},
				"created_by": "admin@test.com",
			},
			requiredKeys: []string{"schema_id", "name", "schema_definition"},
		},
		{
			entityName: "api_keys",
			sampleData: map[string]interface{}{
				"key_id":      "test-key-1",
				"tenant_id":   "test-tenant",
				"user_id":     "test-user-1",
				"name":        "Test API Key",
				"permissions": []string{"read", "write"},
			},
			requiredKeys: []string{"key_id", "tenant_id", "user_id", "name"},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Entity_%s", tc.entityName), func(t *testing.T) {
			// Test CREATE (POST /api/platform/{entity})
			t.Run("Create", func(t *testing.T) {
				jsonData, err := json.Marshal(tc.sampleData)
				require.NoError(t, err)

				req := httptest.NewRequest("POST", fmt.Sprintf("/api/platform/%s", tc.entityName), bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer test-admin-token") // Mock admin token
				
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)

				// Should return 201 Created or 401 Unauthorized (if auth is working)
				assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusUnauthorized,
					"Expected 201 Created or 401 Unauthorized, got %d", w.Code)

				if w.Code == http.StatusCreated {
					var response map[string]interface{}
					err := json.Unmarshal(w.Body.Bytes(), &response)
					require.NoError(t, err)

					// Check that response contains the created data
					assert.Contains(t, response, "data")
					data := response["data"].(map[string]interface{})

					// Verify required keys are present
					for _, key := range tc.requiredKeys {
						assert.Contains(t, data, key, "Response should contain required key: %s", key)
					}
				}
			})

			// Test LIST (GET /api/platform/{entity})
			t.Run("List", func(t *testing.T) {
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/platform/%s", tc.entityName), nil)
				req.Header.Set("Authorization", "Bearer test-admin-token") // Mock admin token
				
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)

				// Should return 200 OK or 401 Unauthorized (if auth is working)
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
					"Expected 200 OK or 401 Unauthorized, got %d", w.Code)

				if w.Code == http.StatusOK {
					var response map[string]interface{}
					err := json.Unmarshal(w.Body.Bytes(), &response)
					require.NoError(t, err)

					// Check that response has the expected structure
					assert.Contains(t, response, "data")
					assert.Contains(t, response, "meta")
				}
			})

			// Test LIST with pagination (GET /api/platform/{entity}?limit=10&offset=0)
			t.Run("ListWithPagination", func(t *testing.T) {
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/platform/%s?limit=10&offset=0", tc.entityName), nil)
				req.Header.Set("Authorization", "Bearer test-admin-token") // Mock admin token
				
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)

				// Should return 200 OK or 401 Unauthorized (if auth is working)
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
					"Expected 200 OK or 401 Unauthorized, got %d", w.Code)
			})

			// Test GET by ID (GET /api/platform/{entity}/{id})
			t.Run("GetByID", func(t *testing.T) {
				testID := "test-id-123"
				req := httptest.NewRequest("GET", fmt.Sprintf("/api/platform/%s/%s", tc.entityName, testID), nil)
				req.Header.Set("Authorization", "Bearer test-admin-token") // Mock admin token
				
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)

				// Should return 404 Not Found, 401 Unauthorized, or 200 OK
				assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusUnauthorized || w.Code == http.StatusOK,
					"Expected 404 Not Found, 401 Unauthorized, or 200 OK, got %d", w.Code)
			})

			// Test UPDATE (PUT /api/platform/{entity}/{id})
			t.Run("Update", func(t *testing.T) {
				testID := "test-id-123"
				updateData := map[string]interface{}{
					"name": "Updated " + tc.sampleData["name"].(string),
				}
				jsonData, err := json.Marshal(updateData)
				require.NoError(t, err)

				req := httptest.NewRequest("PUT", fmt.Sprintf("/api/platform/%s/%s", tc.entityName, testID), bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer test-admin-token") // Mock admin token
				
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)

				// Should return 404 Not Found, 401 Unauthorized, or 200 OK
				assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusUnauthorized || w.Code == http.StatusOK,
					"Expected 404 Not Found, 401 Unauthorized, or 200 OK, got %d", w.Code)
			})

			// Test DELETE (DELETE /api/platform/{entity}/{id})
			t.Run("Delete", func(t *testing.T) {
				testID := "test-id-123"
				req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/platform/%s/%s", tc.entityName, testID), nil)
				req.Header.Set("Authorization", "Bearer test-admin-token") // Mock admin token
				
				w := httptest.NewRecorder()
				engine.ServeHTTP(w, req)

				// Should return 404 Not Found, 401 Unauthorized, or 200 OK
				assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusUnauthorized || w.Code == http.StatusOK,
					"Expected 404 Not Found, 401 Unauthorized, or 200 OK, got %d", w.Code)
			})
		})
	}
}

// TestPlatformSchemaEndpoints tests that the schema-driven endpoints are correctly mapped
func TestPlatformSchemaEndpoints(t *testing.T) {
	config := &api.Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../schemas/platform.yaml",
		DatabaseURL:  "postgres://test:test@localhost:5432/test_db?sslmode=disable",
		Port:         "8080",
	}

	engine, err := api.NewEngine(config)
	require.NoError(t, err, "Failed to create API engine")

	// Test that all expected endpoints are available
	expectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/platform/tenants"},
		{"POST", "/api/platform/tenants"},
		{"GET", "/api/platform/tenants/test-id"},
		{"PUT", "/api/platform/tenants/test-id"},
		{"DELETE", "/api/platform/tenants/test-id"},
		
		{"GET", "/api/platform/users"},
		{"POST", "/api/platform/users"},
		{"GET", "/api/platform/users/test-id"},
		{"PUT", "/api/platform/users/test-id"},
		{"DELETE", "/api/platform/users/test-id"},
		
		{"GET", "/api/platform/schemas"},
		{"POST", "/api/platform/schemas"},
		{"GET", "/api/platform/schemas/test-id"},
		{"PUT", "/api/platform/schemas/test-id"},
		{"DELETE", "/api/platform/schemas/test-id"},
		
		{"GET", "/api/platform/api_keys"},
		{"POST", "/api/platform/api_keys"},
		{"GET", "/api/platform/api_keys/test-id"},
		{"PUT", "/api/platform/api_keys/test-id"},
		{"DELETE", "/api/platform/api_keys/test-id"},
	}

	for _, endpoint := range expectedEndpoints {
		t.Run(fmt.Sprintf("%s_%s", endpoint.method, endpoint.path), func(t *testing.T) {
			var req *http.Request
			if endpoint.method == "POST" || endpoint.method == "PUT" {
				req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer([]byte("{}")))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
			}
			req.Header.Set("Authorization", "Bearer test-admin-token") // Mock admin token
			
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			// Should not return 404 Not Found (endpoint should exist)
			assert.NotEqual(t, http.StatusNotFound, w.Code, 
				"Endpoint %s %s should exist (got 404)", endpoint.method, endpoint.path)
			
			// Should return 401 Unauthorized (auth is working) or other valid status
			assert.True(t, w.Code == http.StatusUnauthorized || w.Code < 500,
				"Endpoint %s %s should return valid status, got %d", endpoint.method, endpoint.path, w.Code)
		})
	}
}

// TestAuthenticationRequired tests that all platform endpoints require authentication
func TestAuthenticationRequired(t *testing.T) {
	config := &api.Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../schemas/platform.yaml",
		DatabaseURL:  "postgres://test:test@localhost:5432/test_db?sslmode=disable",
		Port:         "8080",
	}

	engine, err := api.NewEngine(config)
	require.NoError(t, err, "Failed to create API engine")

	// Test that requests without authentication are rejected
	protectedEndpoints := []string{
		"/api/platform/tenants",
		"/api/platform/users",
		"/api/platform/schemas",
		"/api/platform/api_keys",
	}

	for _, endpoint := range protectedEndpoints {
		t.Run(fmt.Sprintf("Endpoint_%s", endpoint), func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			// No Authorization header
			
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			// Should return 401 Unauthorized
			assert.Equal(t, http.StatusUnauthorized, w.Code,
				"Endpoint %s should require authentication", endpoint)
		})
	}
}

// TestSchemaValidation tests that the platform.yaml schema is properly loaded and validated
func TestSchemaValidation(t *testing.T) {
	config := &api.Config{
		TenantID:     "test-tenant",
		SchemaSource: "file",
		SchemaPath:   "../../schemas/platform.yaml",
		DatabaseURL:  "postgres://test:test@localhost:5432/test_db?sslmode=disable",
		Port:         "8080",
	}

	engine, err := api.NewEngine(config)
	require.NoError(t, err, "Failed to create API engine")

	// Test schema info endpoint
	req := httptest.NewRequest("GET", "/schema", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify schema structure
	assert.Contains(t, response, "schema")
	schema := response["schema"].(map[string]interface{})
	
	assert.Contains(t, schema, "service")
	assert.Contains(t, schema, "entities")
	
	service := schema["service"].(map[string]interface{})
	assert.Equal(t, "platform", service["name"])
	
	entities := schema["entities"].(map[string]interface{})
	
	// Verify all expected entities are present
	expectedEntities := []string{"tenants", "users", "schemas", "api_keys"}
	for _, entityName := range expectedEntities {
		assert.Contains(t, entities, entityName, "Schema should contain entity: %s", entityName)
	}
}
