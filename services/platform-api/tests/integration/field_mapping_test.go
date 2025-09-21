package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/backsaas/platform/services/platform-api/internal/api"
	"github.com/backsaas/platform/services/platform-api/internal/schema"
	_ "github.com/lib/pq"
)

// TestFieldMappingIntegration tests field mapping with real database operations
func TestFieldMappingIntegration(t *testing.T) {
	// Get database URL from environment or use default for Docker network
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

	// Create CRM-like schema that was causing field mapping issues
	crmSchema := &schema.Schema{
		Version: 1,
		Service: schema.ServiceConfig{
			Name:        "integration-test-crm",
			Description: "Integration test CRM for field mapping verification",
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
	dbOps := api.NewDatabaseOperations(db, "integration-test")

	t.Run("CRMFieldMappingIntegration", func(t *testing.T) {
		// Clean up any existing table
		db.Exec("DROP TABLE IF EXISTS contacts")

		// Create table using our fixed implementation
		err := dbOps.EnsureTablesExist(crmSchema)
		if err != nil {
			t.Fatalf("Failed to create CRM tables: %v", err)
		}

		entity := crmSchema.Entities["contacts"]

		// Test the exact data structure that was failing before the fix
		testData := map[string]interface{}{
			"contact_id": "integration-contact-1",
			"email":      "integration@example.com",
			"first_name": "Integration",
			"last_name":  "Test",
			"phone":      "+1234567890",
			"company":    "Test Corp",
			"status":     "lead",
		}

		// Insert the data
		result, err := dbOps.InsertEntity("contacts", entity, testData)
		if err != nil {
			t.Fatalf("Failed to insert integration contact: %v", err)
		}

		// Verify the exact field mappings that were problematic before the fix
		fieldMappings := map[string]interface{}{
			"contact_id": "integration-contact-1",
			"email":      "integration@example.com", // This was ending up in 'status' field before fix
			"first_name": "Integration",              // This was correct
			"last_name":  "Test",                     // This was ending up in 'tags' field before fix
			"phone":      "+1234567890",              // This was ending up in wrong field before fix
			"company":    "Test Corp",                // This was ending up in wrong field before fix
			"status":     "lead",                     // This was ending up in 'phone' field before fix
		}

		for field, expectedValue := range fieldMappings {
			if result[field] != expectedValue {
				t.Errorf("Integration field mapping error for '%s': expected '%v', got '%v'", field, expectedValue, result[field])
			}
		}

		// Test retrieval to ensure consistency
		retrieved, err := dbOps.GetEntity("contacts", entity, "integration-contact-1")
		if err != nil {
			t.Fatalf("Failed to retrieve integration contact: %v", err)
		}

		// Verify retrieved data is correct
		for field, expectedValue := range fieldMappings {
			if retrieved[field] != expectedValue {
				t.Errorf("Integration retrieved field mapping error for '%s': expected '%v', got '%v'", field, expectedValue, retrieved[field])
			}
		}

		// Test multiple insertions to verify consistency
		for i := 2; i <= 5; i++ {
			testData := map[string]interface{}{
				"contact_id": fmt.Sprintf("integration-contact-%d", i),
				"email":      fmt.Sprintf("test%d@example.com", i),
				"first_name": fmt.Sprintf("Test%d", i),
				"last_name":  "User",
				"phone":      fmt.Sprintf("+123456789%d", i),
				"company":    fmt.Sprintf("Company%d", i),
				"status":     "prospect",
			}

			result, err := dbOps.InsertEntity("contacts", entity, testData)
			if err != nil {
				t.Fatalf("Failed to insert contact %d: %v", i, err)
			}

			// Verify each insertion has correct field mappings
			for field, expectedValue := range testData {
				if result[field] != expectedValue {
					t.Errorf("Contact %d field mapping error for '%s': expected '%v', got '%v'", i, field, expectedValue, result[field])
				}
			}
		}

		// Query all contacts to verify consistency
		queryResults, err := dbOps.QueryEntities("contacts", entity, map[string]interface{}{}, 10, 0, "contact_id ASC")
		if err != nil {
			t.Fatalf("Failed to query contacts: %v", err)
		}

		if len(queryResults) != 5 {
			t.Errorf("Expected 5 contacts, got %d", len(queryResults))
		}

		// Verify each queried result has correct field mappings
		for i, contact := range queryResults {
			expectedID := fmt.Sprintf("integration-contact-%d", i+1)
			if contact["contact_id"] != expectedID {
				t.Errorf("Query result %d has wrong contact_id: expected '%s', got '%v'", i, expectedID, contact["contact_id"])
			}
		}
	})

	// Clean up
	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS contacts")
	})
}
