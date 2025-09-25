package api

import (
	"database/sql"
	"os"
	"testing"

	"github.com/backsaas/platform/services/platform-api/internal/schema"
	_ "github.com/lib/pq"
)

func TestDatabaseOperations(t *testing.T) {
	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/backsaas?sslmode=disable"
	}

	// Skip if no database connection available
	db, err := sql.Open("postgres", dbURL)
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
		// Clean up any existing table and its constraints
		db.Exec("DROP TABLE IF EXISTS users CASCADE")

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

		// Clean up - make sure we're working with the test table
		db.Exec("DROP TABLE IF EXISTS users CASCADE")
		dbOps.EnsureTablesExist(testSchema)

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

		// Clean up - make sure we're working with the test table
		db.Exec("DROP TABLE IF EXISTS users CASCADE")
		dbOps.EnsureTablesExist(testSchema)

		// Create operations for two different tenants
		dbOps1 := NewDatabaseOperations(db, "tenant1")
		dbOps2 := NewDatabaseOperations(db, "tenant2")

		// Insert data for tenant1
		testData1 := map[string]interface{}{
			"user_id": "tenant1-user1",
			"email":   "user1@tenant1.com",
			"name":    "Tenant 1 User",
		}

		_, err := dbOps1.InsertEntity("users", entity, testData1)
		if err != nil {
			t.Fatalf("Failed to insert entity for tenant1: %v", err)
		}

		// Insert data for tenant2
		testData2 := map[string]interface{}{
			"user_id": "tenant2-user1", // Different ID for different tenant
			"email":   "user1@tenant2.com",
			"name":    "Tenant 2 User",
		}

		_, err = dbOps2.InsertEntity("users", entity, testData2)
		if err != nil {
			t.Fatalf("Failed to insert entity for tenant2: %v", err)
		}

		// Verify tenant1 can only see its data
		result1, err := dbOps1.GetEntity("users", entity, "tenant1-user1")
		if err != nil {
			t.Fatalf("Failed to get entity for tenant1: %v", err)
		}
		if result1["email"] != "user1@tenant1.com" {
			t.Errorf("Tenant1 got wrong data: %v", result1["email"])
		}

		// Verify tenant2 can only see its data
		result2, err := dbOps2.GetEntity("users", entity, "tenant2-user1")
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

// TestFieldMappingConsistency tests that column order is consistent between table creation and data retrieval
func TestFieldMappingConsistency(t *testing.T) {
	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/backsaas?sslmode=disable"
	}

	// Skip if no database connection available
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skip("Database not accessible for testing")
	}

	// Create a complex schema with many fields to test field ordering
	testSchema := &schema.Schema{
		Version: 1,
		Service: schema.ServiceConfig{
			Name:        "field-mapping-test",
			Description: "Test service for field mapping consistency",
		},
		Entities: map[string]*schema.Entity{
			"contacts": {
				Key: "contact_id",
				Schema: schema.EntitySchema{
					Type:     "object",
					Required: []string{"contact_id", "email", "first_name", "last_name"},
					Properties: map[string]*schema.PropertyDefinition{
						"contact_id": {
							Type: "string",
						},
						"email": {
							Type:   "string",
							Format: "email",
						},
						"first_name": {
							Type:      "string",
							MinLength: 1,
							MaxLength: 50,
						},
						"last_name": {
							Type:      "string",
							MinLength: 1,
							MaxLength: 50,
						},
						"phone": {
							Type:    "string",
							Pattern: "^[+]?[1-9]?[0-9]{7,15}$",
						},
						"company": {
							Type:      "string",
							MaxLength: 100,
						},
						"status": {
							Type:    "string",
							Default: "lead",
						},
						"tags": {
							Type: "array",
							Items: &schema.PropertyDefinition{
								Type: "string",
							},
						},
						"metadata": {
							Type: "object",
						},
					},
				},
			},
		},
	}

	// Create database operations
	dbOps := NewDatabaseOperations(db, "field-mapping-test")

	t.Run("ColumnOrderConsistency", func(t *testing.T) {
		// Clean up any existing table
		db.Exec("DROP TABLE IF EXISTS contacts")

		// Create table
		err := dbOps.EnsureTablesExist(testSchema)
		if err != nil {
			t.Fatalf("Failed to create tables: %v", err)
		}

		entity := testSchema.Entities["contacts"]

		// Test data with all fields
		testData := map[string]interface{}{
			"contact_id": "contact-123",
			"email":      "john@example.com",
			"first_name": "John",
			"last_name":  "Doe",
			"phone":      "+1234567890",
			"company":    "Acme Corp",
			"status":     "lead",
			"tags":       []string{"important", "new"},
			"metadata":   map[string]interface{}{"source": "website"},
		}

		// Insert the data
		result, err := dbOps.InsertEntity("contacts", entity, testData)
		if err != nil {
			t.Fatalf("Failed to insert entity: %v", err)
		}

		// Verify all fields are correctly mapped
		expectedMappings := map[string]interface{}{
			"contact_id": "contact-123",
			"email":      "john@example.com",
			"first_name": "John",
			"last_name":  "Doe",
			"phone":      "+1234567890",
			"company":    "Acme Corp",
			"status":     "lead",
			"tenant_id":  "field-mapping-test",
		}

		for field, expectedValue := range expectedMappings {
			if result[field] != expectedValue {
				t.Errorf("Field mapping error for '%s': expected '%v', got '%v'", field, expectedValue, result[field])
			}
		}

		// Verify arrays and objects are handled correctly
		if result["tags"] == nil {
			t.Error("Tags field should not be nil")
		}
		if result["metadata"] == nil {
			t.Error("Metadata field should not be nil")
		}

		// Test retrieval to ensure consistency
		retrieved, err := dbOps.GetEntity("contacts", entity, "contact-123")
		if err != nil {
			t.Fatalf("Failed to retrieve entity: %v", err)
		}

		// Verify retrieved data matches inserted data
		for field, expectedValue := range expectedMappings {
			if retrieved[field] != expectedValue {
				t.Errorf("Retrieved field mapping error for '%s': expected '%v', got '%v'", field, expectedValue, retrieved[field])
			}
		}
	})

	t.Run("ColumnOrderDeterministic", func(t *testing.T) {
		entity := testSchema.Entities["contacts"]

		// Get column order multiple times to ensure it's deterministic
		columns1 := dbOps.getEntityColumns(entity)
		columns2 := dbOps.getEntityColumns(entity)
		columns3 := dbOps.getEntityColumns(entity)

		// Verify all calls return the same order
		if len(columns1) != len(columns2) || len(columns2) != len(columns3) {
			t.Fatal("Column count is not consistent across calls")
		}

		for i := range columns1 {
			if columns1[i] != columns2[i] || columns2[i] != columns3[i] {
				t.Errorf("Column order is not deterministic at position %d: %s, %s, %s", i, columns1[i], columns2[i], columns3[i])
			}
		}

		// Verify expected column structure
		expectedStart := []string{"contact_id", "tenant_id"}
		expectedEnd := []string{"created_at", "updated_at"}

		// Check that it starts with key and tenant_id
		for i, expected := range expectedStart {
			if columns1[i] != expected {
				t.Errorf("Expected column %d to be '%s', got '%s'", i, expected, columns1[i])
			}
		}

		// Check that it ends with audit columns
		startOfEnd := len(columns1) - len(expectedEnd)
		for i, expected := range expectedEnd {
			if columns1[startOfEnd+i] != expected {
				t.Errorf("Expected column %d to be '%s', got '%s'", startOfEnd+i, expected, columns1[startOfEnd+i])
			}
		}

		// Verify property columns are in sorted order
		propertyStart := len(expectedStart)
		propertyEnd := len(columns1) - len(expectedEnd)
		propertyColumns := columns1[propertyStart:propertyEnd]

		// Check if property columns are sorted
		for i := 1; i < len(propertyColumns); i++ {
			if propertyColumns[i-1] > propertyColumns[i] {
				t.Errorf("Property columns are not sorted: '%s' should come after '%s'", propertyColumns[i-1], propertyColumns[i])
			}
		}
	})

	t.Run("MultipleInsertionsConsistency", func(t *testing.T) {
		entity := testSchema.Entities["contacts"]

		// Clean up
		db.Exec("DELETE FROM contacts WHERE tenant_id = 'field-mapping-test'")

		// Insert multiple records with different field combinations
		testCases := []map[string]interface{}{
			{
				"contact_id": "contact-1",
				"email":      "user1@example.com",
				"first_name": "Alice",
				"last_name":  "Smith",
				"status":     "prospect",
			},
			{
				"contact_id": "contact-2",
				"email":      "user2@example.com",
				"first_name": "Bob",
				"last_name":  "Johnson",
				"phone":      "+9876543210",
				"company":    "Tech Corp",
				"status":     "customer",
			},
			{
				"contact_id": "contact-3",
				"email":      "user3@example.com",
				"first_name": "Charlie",
				"last_name":  "Brown",
				"phone":      "+5555555555",
				"company":    "StartupXYZ",
				"status":     "lead",
				"tags":       []string{"vip", "urgent"},
				"metadata":   map[string]interface{}{"priority": "high"},
			},
		}

		var insertedResults []map[string]interface{}

		// Insert all test cases
		for i, testData := range testCases {
			result, err := dbOps.InsertEntity("contacts", entity, testData)
			if err != nil {
				t.Fatalf("Failed to insert entity %d: %v", i, err)
			}
			insertedResults = append(insertedResults, result)

			// Verify the inserted data immediately
			for field, expectedValue := range testData {
				retrievedValue := result[field]
				
				// Handle array comparison specially since slices can't be compared directly
				if field == "tags" {
					expectedArray, expectedOk := expectedValue.([]string)
					if expectedOk {
						// Convert retrieved value to comparable format
						if retrievedArray, ok := retrievedValue.([]interface{}); ok {
							if len(expectedArray) != len(retrievedArray) {
								t.Errorf("Insert %d field mapping error for '%s': expected length %d, got length %d", i, field, len(expectedArray), len(retrievedArray))
								continue
							}
							for j, expected := range expectedArray {
								if retrieved, ok := retrievedArray[j].(string); !ok || retrieved != expected {
									t.Errorf("Insert %d field mapping error for '%s'[%d]: expected '%v', got '%v'", i, field, j, expected, retrievedArray[j])
								}
							}
							continue
						}
					}
				}
				
				// Handle object comparison specially
				if field == "metadata" {
					// For objects, just check they're both non-nil or both nil
					if (expectedValue == nil) != (retrievedValue == nil) {
						t.Errorf("Insert %d field mapping error for '%s': expected nil=%v, got nil=%v", i, field, expectedValue == nil, retrievedValue == nil)
					}
					continue
				}
				
				// Regular comparison for simple types
				if retrievedValue != expectedValue {
					t.Errorf("Insert %d field mapping error for '%s': expected '%v', got '%v'", i, field, expectedValue, retrievedValue)
				}
			}
		}

		// Retrieve all records and verify consistency
		queryResults, err := dbOps.QueryEntities("contacts", entity, map[string]interface{}{}, 10, 0, "contact_id ASC")
		if err != nil {
			t.Fatalf("Failed to query entities: %v", err)
		}

		if len(queryResults) != len(testCases) {
			t.Fatalf("Expected %d results, got %d", len(testCases), len(queryResults))
		}

		// Verify each retrieved record matches the original data
		for i, retrieved := range queryResults {
			original := testCases[i]
			for field, expectedValue := range original {
				retrievedValue := retrieved[field]
				
				// Handle array comparison specially since slices can't be compared directly
				if field == "tags" {
					expectedArray, expectedOk := expectedValue.([]string)
					if expectedOk {
						// Convert retrieved value to comparable format
						if retrievedArray, ok := retrievedValue.([]interface{}); ok {
							if len(expectedArray) != len(retrievedArray) {
								t.Errorf("Query result %d field mapping error for '%s': expected length %d, got length %d", i, field, len(expectedArray), len(retrievedArray))
								continue
							}
							for j, expected := range expectedArray {
								if retrieved, ok := retrievedArray[j].(string); !ok || retrieved != expected {
									t.Errorf("Query result %d field mapping error for '%s'[%d]: expected '%v', got '%v'", i, field, j, expected, retrievedArray[j])
								}
							}
							continue
						}
					}
				}
				
				// Handle object comparison specially
				if field == "metadata" {
					// For objects, just check they're both non-nil or both nil
					if (expectedValue == nil) != (retrievedValue == nil) {
						t.Errorf("Query result %d field mapping error for '%s': expected nil=%v, got nil=%v", i, field, expectedValue == nil, retrievedValue == nil)
					}
					continue
				}
				
				// Regular comparison for simple types
				if retrievedValue != expectedValue {
					t.Errorf("Query result %d field mapping error for '%s': expected '%v', got '%v'", i, field, expectedValue, retrievedValue)
				}
			}
		}
	})

	// Clean up
	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS contacts")
	})
}

// TestCRMSchemaFieldMapping tests the exact CRM schema structure used in the test server
func TestCRMSchemaFieldMapping(t *testing.T) {
	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/backsaas?sslmode=disable"
	}

	// Skip if no database connection available
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skip("Database not accessible for testing")
	}

	// Create the exact CRM schema structure that was causing issues
	crmSchema := &schema.Schema{
		Version: 1,
		Service: schema.ServiceConfig{
			Name:        "sample-crm",
			Description: "Sample CRM system for testing database operations",
		},
		Entities: map[string]*schema.Entity{
			"contacts": {
				Key: "contact_id",
				Schema: schema.EntitySchema{
					Type:     "object",
					Required: []string{"contact_id", "email", "first_name", "last_name"},
					Properties: map[string]*schema.PropertyDefinition{
						"contact_id": {
							Type: "string",
						},
						"email": {
							Type:   "string",
							Format: "email",
						},
						"first_name": {
							Type:      "string",
							MinLength: 1,
							MaxLength: 50,
						},
						"last_name": {
							Type:      "string",
							MinLength: 1,
							MaxLength: 50,
						},
						"phone": {
							Type:    "string",
							Pattern: "^[+]?[1-9]?[0-9]{7,15}$",
						},
						"company": {
							Type:      "string",
							MaxLength: 100,
						},
						"status": {
							Type:    "string",
							Default: "lead",
						},
						"tags": {
							Type: "array",
							Items: &schema.PropertyDefinition{
								Type: "string",
							},
						},
						"metadata": {
							Type: "object",
						},
					},
				},
			},
		},
	}

	// Create database operations
	dbOps := NewDatabaseOperations(db, "crm-test")

	t.Run("CRMContactsFieldMapping", func(t *testing.T) {
		// Clean up any existing table
		db.Exec("DROP TABLE IF EXISTS contacts")

		// Create table
		err := dbOps.EnsureTablesExist(crmSchema)
		if err != nil {
			t.Fatalf("Failed to create CRM tables: %v", err)
		}

		entity := crmSchema.Entities["contacts"]

		// Test the exact data structure that was failing
		testData := map[string]interface{}{
			"contact_id": "contact-1",
			"email":      "john@example.com",
			"first_name": "John",
			"last_name":  "Doe",
			"phone":      "+1234567890",
			"company":    "Acme Corp",
			"status":     "lead",
		}

		// Insert the data
		result, err := dbOps.InsertEntity("contacts", entity, testData)
		if err != nil {
			t.Fatalf("Failed to insert CRM contact: %v", err)
		}

		// Verify the exact field mappings that were problematic
		fieldMappings := map[string]interface{}{
			"contact_id": "contact-1",
			"email":      "john@example.com", // This was ending up in 'status' field
			"first_name": "John",             // This was correct
			"last_name":  "Doe",              // This was ending up in 'tags' field
			"phone":      "+1234567890",      // This was ending up in wrong field
			"company":    "Acme Corp",        // This was ending up in wrong field
			"status":     "lead",             // This was ending up in 'phone' field
		}

		for field, expectedValue := range fieldMappings {
			if result[field] != expectedValue {
				t.Errorf("CRM field mapping error for '%s': expected '%v', got '%v'", field, expectedValue, result[field])
			}
		}

		// Test retrieval
		retrieved, err := dbOps.GetEntity("contacts", entity, "contact-1")
		if err != nil {
			t.Fatalf("Failed to retrieve CRM contact: %v", err)
		}

		// Verify retrieved data is correct
		for field, expectedValue := range fieldMappings {
			if retrieved[field] != expectedValue {
				t.Errorf("CRM retrieved field mapping error for '%s': expected '%v', got '%v'", field, expectedValue, retrieved[field])
			}
		}
	})

	// Clean up
	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS contacts")
	})
}
