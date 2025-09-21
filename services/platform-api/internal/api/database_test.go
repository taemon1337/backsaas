package api

import (
	"database/sql"
	"testing"

	"github.com/backsaas/platform/services/platform-api/internal/schema"
	_ "github.com/lib/pq"
)

func TestDatabaseOperations(t *testing.T) {
	// Skip if no database connection available
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skip("Database not accessible for testing")
	}

	// Create test schema
	testSchema := &schema.Schema{
		Version: 1,
		Service: schema.ServiceConfig{
			Name:        "test-service",
			Description: "Test service for database operations",
		},
		Entities: map[string]*schema.Entity{
			"users": {
				Key: "user_id",
				Schema: schema.EntitySchema{
					Type:     "object",
					Required: []string{"user_id", "email", "name"},
					Properties: map[string]*schema.PropertyDefinition{
						"user_id": {
							Type: "string",
						},
						"email": {
							Type:   "string",
							Format: "email",
						},
						"name": {
							Type:      "string",
							MinLength: 1,
							MaxLength: 100,
						},
						"age": {
							Type:    "integer",
							Minimum: 0,
							Maximum: 150,
						},
						"active": {
							Type:    "boolean",
							Default: true,
						},
					},
				},
			},
		},
	}

	// Create database operations
	dbOps := NewDatabaseOperations(db, "test-tenant")

	t.Run("EnsureTablesExist", func(t *testing.T) {
		// Clean up any existing table
		db.Exec("DROP TABLE IF EXISTS users")

		err := dbOps.EnsureTablesExist(testSchema)
		if err != nil {
			t.Fatalf("Failed to create tables: %v", err)
		}

		// Verify table exists
		var exists bool
		err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check table existence: %v", err)
		}
		if !exists {
			t.Fatal("Table 'users' was not created")
		}
	})

	t.Run("ValidateEntityData", func(t *testing.T) {
		entity := testSchema.Entities["users"]

		// Valid data
		validData := map[string]interface{}{
			"user_id": "user123",
			"email":   "test@example.com",
			"name":    "Test User",
			"age":     25,
			"active":  true,
		}

		err := dbOps.ValidateEntityData(entity, validData)
		if err != nil {
			t.Errorf("Valid data should not produce error: %v", err)
		}

		// Missing required field
		invalidData := map[string]interface{}{
			"user_id": "user123",
			"name":    "Test User",
			// missing email
		}

		err = dbOps.ValidateEntityData(entity, invalidData)
		if err == nil {
			t.Error("Missing required field should produce error")
		}

		// Invalid type
		invalidTypeData := map[string]interface{}{
			"user_id": "user123",
			"email":   "test@example.com",
			"name":    "Test User",
			"age":     "not a number", // should be integer
		}

		err = dbOps.ValidateEntityData(entity, invalidTypeData)
		if err == nil {
			t.Error("Invalid type should produce error")
		}
	})

	t.Run("CRUD Operations", func(t *testing.T) {
		entity := testSchema.Entities["users"]

		// Clean up
		db.Exec("DELETE FROM users WHERE tenant_id = 'test-tenant'")

		// Test Insert
		testData := map[string]interface{}{
			"user_id": "user123",
			"email":   "test@example.com",
			"name":    "Test User",
			"age":     25,
			"active":  true,
		}

		result, err := dbOps.InsertEntity("users", entity, testData)
		if err != nil {
			t.Fatalf("Failed to insert entity: %v", err)
		}

		if result["user_id"] != "user123" {
			t.Errorf("Expected user_id 'user123', got %v", result["user_id"])
		}
		if result["tenant_id"] != "test-tenant" {
			t.Errorf("Expected tenant_id 'test-tenant', got %v", result["tenant_id"])
		}

		// Test Get
		getResult, err := dbOps.GetEntity("users", entity, "user123")
		if err != nil {
			t.Fatalf("Failed to get entity: %v", err)
		}

		if getResult["email"] != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %v", getResult["email"])
		}

		// Test Update
		updateData := map[string]interface{}{
			"name": "Updated User",
			"age":  30,
		}

		updateResult, err := dbOps.UpdateEntity("users", entity, "user123", updateData)
		if err != nil {
			t.Fatalf("Failed to update entity: %v", err)
		}

		if updateResult["name"] != "Updated User" {
			t.Errorf("Expected name 'Updated User', got %v", updateResult["name"])
		}

		// Test Query
		filters := map[string]interface{}{
			"active": true,
		}

		queryResults, err := dbOps.QueryEntities("users", entity, filters, 10, 0, "name ASC")
		if err != nil {
			t.Fatalf("Failed to query entities: %v", err)
		}

		if len(queryResults) != 1 {
			t.Errorf("Expected 1 result, got %d", len(queryResults))
		}

		// Test Delete
		err = dbOps.DeleteEntity("users", entity, "user123")
		if err != nil {
			t.Fatalf("Failed to delete entity: %v", err)
		}

		// Verify deletion
		_, err = dbOps.GetEntity("users", entity, "user123")
		if err == nil {
			t.Error("Entity should be deleted")
		}
	})

	t.Run("Multi-tenant Isolation", func(t *testing.T) {
		entity := testSchema.Entities["users"]

		// Clean up
		db.Exec("DELETE FROM users")

		// Create operations for two different tenants
		dbOps1 := NewDatabaseOperations(db, "tenant1")
		dbOps2 := NewDatabaseOperations(db, "tenant2")

		// Insert data for tenant1
		testData1 := map[string]interface{}{
			"user_id": "user1",
			"email":   "user1@tenant1.com",
			"name":    "Tenant 1 User",
		}

		_, err := dbOps1.InsertEntity("users", entity, testData1)
		if err != nil {
			t.Fatalf("Failed to insert entity for tenant1: %v", err)
		}

		// Insert data for tenant2
		testData2 := map[string]interface{}{
			"user_id": "user1", // Same ID but different tenant
			"email":   "user1@tenant2.com",
			"name":    "Tenant 2 User",
		}

		_, err = dbOps2.InsertEntity("users", entity, testData2)
		if err != nil {
			t.Fatalf("Failed to insert entity for tenant2: %v", err)
		}

		// Verify tenant1 can only see its data
		result1, err := dbOps1.GetEntity("users", entity, "user1")
		if err != nil {
			t.Fatalf("Failed to get entity for tenant1: %v", err)
		}
		if result1["email"] != "user1@tenant1.com" {
			t.Errorf("Tenant1 got wrong data: %v", result1["email"])
		}

		// Verify tenant2 can only see its data
		result2, err := dbOps2.GetEntity("users", entity, "user1")
		if err != nil {
			t.Fatalf("Failed to get entity for tenant2: %v", err)
		}
		if result2["email"] != "user1@tenant2.com" {
			t.Errorf("Tenant2 got wrong data: %v", result2["email"])
		}

		// Verify tenant1 cannot see tenant2's data in queries
		queryResults1, err := dbOps1.QueryEntities("users", entity, map[string]interface{}{}, 10, 0, "")
		if err != nil {
			t.Fatalf("Failed to query entities for tenant1: %v", err)
		}
		if len(queryResults1) != 1 {
			t.Errorf("Tenant1 should see only 1 record, got %d", len(queryResults1))
		}
	})

	// Clean up
	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS users")
	})
}

func TestPropertyToColumnDefinition(t *testing.T) {
	dbOps := &DatabaseOperations{}

	testCases := []struct {
		name     string
		propName string
		propDef  *schema.PropertyDefinition
		expected string
	}{
		{
			name:     "String property",
			propName: "name",
			propDef: &schema.PropertyDefinition{
				Type:      "string",
				MaxLength: 100,
			},
			expected: "name VARCHAR(100)",
		},
		{
			name:     "Email property",
			propName: "email",
			propDef: &schema.PropertyDefinition{
				Type:   "string",
				Format: "email",
			},
			expected: "email VARCHAR(255)",
		},
		{
			name:     "Integer property",
			propName: "age",
			propDef: &schema.PropertyDefinition{
				Type: "integer",
			},
			expected: "age INTEGER",
		},
		{
			name:     "Boolean property with default",
			propName: "active",
			propDef: &schema.PropertyDefinition{
				Type:    "boolean",
				Default: true,
			},
			expected: "active BOOLEAN DEFAULT 'true'",
		},
		{
			name:     "JSON property",
			propName: "metadata",
			propDef: &schema.PropertyDefinition{
				Type: "object",
			},
			expected: "metadata JSONB",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := dbOps.propertyToColumnDefinition(tc.propName, tc.propDef)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
