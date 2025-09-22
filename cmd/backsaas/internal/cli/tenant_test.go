package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestTenantCommands(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "TenantListCommand",
			cmd:         tenantListCmd,
			args:        []string{},
			expectError: false,
			description: "Should execute tenant list command without error",
		},
		{
			name:        "TenantCreateCommand_ValidArgs",
			cmd:         tenantCreateCmd,
			args:        []string{"test-tenant"},
			expectError: false,
			description: "Should execute tenant create with valid tenant name",
		},
		{
			name:        "TenantCreateCommand_NoArgs",
			cmd:         tenantCreateCmd,
			args:        []string{},
			expectError: true,
			description: "Should fail when no tenant name provided",
		},
		{
			name:        "TenantShowCommand_ValidArgs",
			cmd:         tenantShowCmd,
			args:        []string{"test-tenant-id"},
			expectError: false,
			description: "Should execute tenant show with valid tenant ID",
		},
		{
			name:        "TenantShowCommand_NoArgs",
			cmd:         tenantShowCmd,
			args:        []string{},
			expectError: true,
			description: "Should fail when no tenant ID provided",
		},
		{
			name:        "TenantDeleteCommand_ValidArgs",
			cmd:         tenantDeleteCmd,
			args:        []string{"test-tenant-id"},
			expectError: false,
			description: "Should execute tenant delete with valid tenant ID",
		},
		{
			name:        "TenantDeleteCommand_NoArgs",
			cmd:         tenantDeleteCmd,
			args:        []string{},
			expectError: true,
			description: "Should fail when no tenant ID provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Reset command args and flags
			tt.cmd.SetArgs(tt.args)
			tt.cmd.ResetFlags()

			// Execute command
			err := tt.cmd.Execute()

			// Restore output
			w.Close()
			os.Stdout = old
			
			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}

			// Basic output validation
			if !tt.expectError && len(output) == 0 {
				t.Errorf("%s: expected output but got none", tt.description)
			}
		})
	}
}

func TestTenantStruct(t *testing.T) {
	t.Run("TenantJSONMarshaling", func(t *testing.T) {
		tenant := Tenant{
			ID:          "test-123",
			Name:        "Test Tenant",
			Slug:        "test-tenant",
			Domain:      "test.example.com",
			Plan:        "pro",
			Status:      "active",
			CreatedAt:   "2023-01-01T00:00:00Z",
			UpdatedAt:   "2023-01-02T00:00:00Z",
			UserCount:   5,
			SchemaCount: 3,
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(tenant)
		if err != nil {
			t.Fatalf("Failed to marshal tenant to JSON: %v", err)
		}

		// Test JSON unmarshaling
		var unmarshaledTenant Tenant
		err = json.Unmarshal(jsonData, &unmarshaledTenant)
		if err != nil {
			t.Fatalf("Failed to unmarshal tenant from JSON: %v", err)
		}

		// Verify data integrity
		if unmarshaledTenant.ID != tenant.ID {
			t.Errorf("Expected ID %s, got %s", tenant.ID, unmarshaledTenant.ID)
		}
		if unmarshaledTenant.Name != tenant.Name {
			t.Errorf("Expected Name %s, got %s", tenant.Name, unmarshaledTenant.Name)
		}
		if unmarshaledTenant.UserCount != tenant.UserCount {
			t.Errorf("Expected UserCount %d, got %d", tenant.UserCount, unmarshaledTenant.UserCount)
		}
	})
}

func TestTenantFlags(t *testing.T) {
	t.Run("TenantCreateFlags", func(t *testing.T) {
		// Reset flags to default values
		tenantSchema = ""
		tenantDomain = ""
		tenantPlan = ""

		// Test flag parsing
		tenantCreateCmd.SetArgs([]string{"test-tenant", "--schema", "schema.yaml", "--domain", "test.com", "--plan", "enterprise"})
		
		// Parse flags
		err := tenantCreateCmd.ParseFlags([]string{"--schema", "schema.yaml", "--domain", "test.com", "--plan", "enterprise"})
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		// Verify flag values
		if tenantSchema != "schema.yaml" {
			t.Errorf("Expected schema flag 'schema.yaml', got '%s'", tenantSchema)
		}
		if tenantDomain != "test.com" {
			t.Errorf("Expected domain flag 'test.com', got '%s'", tenantDomain)
		}
		if tenantPlan != "enterprise" {
			t.Errorf("Expected plan flag 'enterprise', got '%s'", tenantPlan)
		}
	})

	t.Run("TenantDeleteFlags", func(t *testing.T) {
		// Reset flag
		tenantForce = false

		// Test force flag
		tenantDeleteCmd.SetArgs([]string{"test-tenant", "--force"})
		err := tenantDeleteCmd.ParseFlags([]string{"--force"})
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		if !tenantForce {
			t.Error("Expected force flag to be true")
		}
	})

	t.Run("TenantListFlags", func(t *testing.T) {
		// Reset flag
		tenantOutputJSON = false

		// Test JSON output flag
		tenantListCmd.SetArgs([]string{"--json"})
		err := tenantListCmd.ParseFlags([]string{"--json"})
		if err != nil {
			t.Fatalf("Failed to parse flags: %v", err)
		}

		if !tenantOutputJSON {
			t.Error("Expected json flag to be true")
		}
	})
}

func TestRunTenantList(t *testing.T) {
	t.Run("TenantListExecution", func(t *testing.T) {
		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute tenant list
		err := runTenantList(tenantListCmd, []string{})

		// Restore output
		w.Close()
		os.Stdout = old

		// Read output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify execution
		if err != nil {
			t.Errorf("runTenantList failed: %v", err)
		}

		// Check for expected output elements
		if !strings.Contains(output, "Tenant List") {
			t.Error("Expected 'Tenant List' in output")
		}

		// Should contain table headers or tenant information
		if !strings.Contains(output, "system") && !strings.Contains(output, "ID") {
			t.Error("Expected tenant data or headers in output")
		}
	})
}

func TestTenantCommandStructure(t *testing.T) {
	t.Run("TenantCommandHasSubcommands", func(t *testing.T) {
		subcommands := tenantCmd.Commands()
		expectedSubcommands := []string{"list", "create", "show", "delete"}

		if len(subcommands) != len(expectedSubcommands) {
			t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
		}

		for _, expected := range expectedSubcommands {
			found := false
			for _, cmd := range subcommands {
				if cmd.Use == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected subcommand '%s' not found", expected)
			}
		}
	})

	t.Run("TenantCommandMetadata", func(t *testing.T) {
		if tenantCmd.Use != "tenant" {
			t.Errorf("Expected Use 'tenant', got '%s'", tenantCmd.Use)
		}
		if tenantCmd.Short == "" {
			t.Error("Expected non-empty Short description")
		}
		if tenantCmd.Long == "" {
			t.Error("Expected non-empty Long description")
		}
	})
}

// Benchmark tests for performance
func BenchmarkTenantList(b *testing.B) {
	// Redirect output to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTenantList(tenantListCmd, []string{})
	}
}

func BenchmarkTenantJSONMarshal(b *testing.B) {
	tenant := Tenant{
		ID:          "benchmark-test",
		Name:        "Benchmark Tenant",
		Slug:        "benchmark-tenant",
		Plan:        "pro",
		Status:      "active",
		UserCount:   100,
		SchemaCount: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(tenant)
	}
}
