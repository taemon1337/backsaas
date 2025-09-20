package schema

import (
	"os"
	"testing"
)

func TestSchemaLoader(t *testing.T) {
	// Create loader with testdata path
	loader := NewLoader("../../testdata")
	
	t.Run("LoadTestSchema", func(t *testing.T) {
		schema, err := loader.LoadFromFile("test-schema.yaml")
		if err != nil {
			t.Fatalf("Failed to load test schema: %v", err)
		}
		
		// Validate basic schema properties
		if schema.Version != 1 {
			t.Errorf("Expected version 1, got %d", schema.Version)
		}
		
		if schema.Service.Name != "test-api" {
			t.Errorf("Expected service name 'test-api', got '%s'", schema.Service.Name)
		}
		
		// Validate entities
		if len(schema.Entities) != 2 {
			t.Errorf("Expected 2 entities, got %d", len(schema.Entities))
		}
		
		// Check users entity
		users, exists := schema.Entities["users"]
		if !exists {
			t.Fatal("Users entity not found")
		}
		
		if users.Key != "id" {
			t.Errorf("Expected users key 'id', got '%s'", users.Key)
		}
		
		if len(users.Schema.Required) != 3 {
			t.Errorf("Expected 3 required fields for users, got %d", len(users.Schema.Required))
		}
		
		// Check products entity
		products, exists := schema.Entities["products"]
		if !exists {
			t.Fatal("Products entity not found")
		}
		
		if products.Key != "id" {
			t.Errorf("Expected products key 'id', got '%s'", products.Key)
		}
		
		// Validate Go functions
		if len(schema.GoFunctions) != 3 {
			t.Errorf("Expected 3 Go functions, got %d", len(schema.GoFunctions))
		}
		
		validateEmail, exists := schema.GoFunctions["validate_email"]
		if !exists {
			t.Fatal("validate_email function not found")
		}
		
		if validateEmail.Package != "validation" {
			t.Errorf("Expected package 'validation', got '%s'", validateEmail.Package)
		}
		
		// Validate platform functions
		if len(schema.Functions) != 3 {
			t.Errorf("Expected 3 platform functions, got %d", len(schema.Functions))
		}
		
		validateUserEmail, exists := schema.Functions["validate_user_email"]
		if !exists {
			t.Fatal("validate_user_email function not found")
		}
		
		if validateUserEmail.Entity != "users" {
			t.Errorf("Expected entity 'users', got '%s'", validateUserEmail.Entity)
		}
		
		if validateUserEmail.Type != "validation" {
			t.Errorf("Expected type 'validation', got '%s'", validateUserEmail.Type)
		}
	})
	
	t.Run("LoadPlatformSchema", func(t *testing.T) {
		// Test loading the actual platform schema - skip if not found
		platformSchemaPath := "../../../services/api/system/schema/platform.yaml"
		
		// Check if platform schema exists, skip if not
		if _, err := os.Stat(platformSchemaPath); os.IsNotExist(err) {
			t.Skip("Platform schema not found, skipping test")
			return
		}
		
		loader := NewLoader("../../../services/api/system/schema")
		schema, err := loader.LoadFromFile("platform.yaml")
		if err != nil {
			t.Fatalf("Failed to load platform schema: %v", err)
		}
		
		// Validate platform schema
		if schema.Service.Name != "backsaas-platform" {
			t.Errorf("Expected service name 'backsaas-platform', got '%s'", schema.Service.Name)
		}
		
		// Check for key platform entities
		expectedEntities := []string{"users", "tenants", "schemas", "functions"}
		for _, entityName := range expectedEntities {
			if _, exists := schema.Entities[entityName]; !exists {
				t.Errorf("Expected entity '%s' not found in platform schema", entityName)
			}
		}
		
		// Validate platform has Go function registry
		if len(schema.GoFunctions) == 0 {
			t.Error("Platform schema should have Go function registry")
		}
		
		// Validate platform functions
		if len(schema.Functions) == 0 {
			t.Error("Platform schema should have platform functions")
		}
	})
	
	t.Run("InvalidSchema", func(t *testing.T) {
		// Test invalid YAML
		_, err := loader.LoadFromBytes([]byte("invalid: yaml: content:"))
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
		
		// Test schema without version
		invalidSchema := `
service:
  name: "test"
entities:
  test:
    key: id
    schema:
      type: object
      properties:
        id: { type: string }
`
		_, err = loader.LoadFromBytes([]byte(invalidSchema))
		if err == nil {
			t.Error("Expected error for schema without version")
		}
		
		// Test schema without entities
		invalidSchema2 := `
version: 1
service:
  name: "test"
entities: {}
`
		_, err = loader.LoadFromBytes([]byte(invalidSchema2))
		if err == nil {
			t.Error("Expected error for schema without entities")
		}
	})
	
	t.Run("EntityValidation", func(t *testing.T) {
		// Test entity without key
		invalidEntity := `
version: 1
service:
  name: "test"
entities:
  test:
    schema:
      type: object
      properties:
        id: { type: string }
`
		_, err := loader.LoadFromBytes([]byte(invalidEntity))
		if err == nil {
			t.Error("Expected error for entity without key")
		}
		
		// Test entity with non-object schema
		invalidEntity2 := `
version: 1
service:
  name: "test"
entities:
  test:
    key: id
    schema:
      type: string
`
		_, err = loader.LoadFromBytes([]byte(invalidEntity2))
		if err == nil {
			t.Error("Expected error for entity with non-object schema")
		}
		
		// Test entity where key field doesn't exist in properties
		invalidEntity3 := `
version: 1
service:
  name: "test"
entities:
  test:
    key: missing_key
    schema:
      type: object
      properties:
        id: { type: string }
`
		_, err = loader.LoadFromBytes([]byte(invalidEntity3))
		if err == nil {
			t.Error("Expected error for entity where key field doesn't exist")
		}
	})
	
	t.Run("FunctionValidation", func(t *testing.T) {
		// Test function referencing non-existent entity
		invalidFunction := `
version: 1
service:
  name: "test"
entities:
  users:
    key: id
    schema:
      type: object
      properties:
        id: { type: string }
platform_functions:
  test_function:
    entity: "non_existent_entity"
    type: "validation"
    trigger: "before_create"
`
		_, err := loader.LoadFromBytes([]byte(invalidFunction))
		if err == nil {
			t.Error("Expected error for function referencing non-existent entity")
		}
	})
}

func TestSchemaProperties(t *testing.T) {
	loader := NewLoader("../../testdata")
	schema, err := loader.LoadFromFile("test-schema.yaml")
	if err != nil {
		t.Fatalf("Failed to load test schema: %v", err)
	}
	
	t.Run("PropertyValidation", func(t *testing.T) {
		users := schema.Entities["users"]
		
		// Check email property
		emailProp, exists := users.Schema.Properties["email"]
		if !exists {
			t.Fatal("Email property not found")
		}
		
		if emailProp.Type != "string" {
			t.Errorf("Expected email type 'string', got '%s'", emailProp.Type)
		}
		
		if emailProp.Format != "email" {
			t.Errorf("Expected email format 'email', got '%s'", emailProp.Format)
		}
		
		// Check status property with enum
		statusProp, exists := users.Schema.Properties["status"]
		if !exists {
			t.Fatal("Status property not found")
		}
		
		expectedEnums := []string{"active", "inactive", "pending"}
		if len(statusProp.Enum) != len(expectedEnums) {
			t.Errorf("Expected %d enum values, got %d", len(expectedEnums), len(statusProp.Enum))
		}
		
		for i, expected := range expectedEnums {
			if i >= len(statusProp.Enum) || statusProp.Enum[i] != expected {
				t.Errorf("Expected enum value '%s' at index %d", expected, i)
			}
		}
	})
	
	t.Run("AccessRules", func(t *testing.T) {
		users := schema.Entities["users"]
		
		if users.Access == nil {
			t.Fatal("Users entity should have access rules")
		}
		
		// Check read access
		if len(users.Access.Read) != 2 {
			t.Errorf("Expected 2 read access rules, got %d", len(users.Access.Read))
		}
		
		// Check for admin role
		hasAdminRole := false
		hasSelfRule := false
		
		for _, rule := range users.Access.Read {
			if rule.Role == "admin" {
				hasAdminRole = true
			}
			if rule.Rule == "self" {
				hasSelfRule = true
			}
		}
		
		if !hasAdminRole {
			t.Error("Expected admin role in read access")
		}
		
		if !hasSelfRule {
			t.Error("Expected self rule in read access")
		}
	})
	
	t.Run("Indexes", func(t *testing.T) {
		if schema.Indexes == nil {
			t.Fatal("Schema should have indexes")
		}
		
		userIndexes, exists := schema.Indexes["users"]
		if !exists {
			t.Fatal("Users indexes not found")
		}
		
		if len(userIndexes) != 2 {
			t.Errorf("Expected 2 user indexes, got %d", len(userIndexes))
		}
		
		// Check email unique index
		emailIndex := userIndexes[0]
		if len(emailIndex.Fields) != 1 || emailIndex.Fields[0] != "email" {
			t.Error("Expected email index with single field 'email'")
		}
		
		if !emailIndex.Unique {
			t.Error("Expected email index to be unique")
		}
	})
}
