package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func TestConfigCommands(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "ConfigShowCommand",
			cmd:         configShowCmd,
			args:        []string{},
			expectError: false,
			description: "Should execute config show command without error",
		},
		{
			name:        "ConfigSetCommand_ValidArgs",
			cmd:         configSetCmd,
			args:        []string{"gateway_url", "http://localhost:8000"},
			expectError: false,
			description: "Should execute config set with valid key-value pair",
		},
		{
			name:        "ConfigSetCommand_InvalidArgs",
			cmd:         configSetCmd,
			args:        []string{"gateway_url"},
			expectError: true,
			description: "Should fail when only key provided without value",
		},
		{
			name:        "ConfigInitCommand",
			cmd:         configInitCmd,
			args:        []string{},
			expectError: false,
			description: "Should execute config init command without error",
		},
		{
			name:        "ConfigValidateCommand",
			cmd:         configValidateCmd,
			args:        []string{},
			expectError: false,
			description: "Should execute config validate command without error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temporary config for testing
			setupTestConfig(t)
			defer cleanupTestConfig(t)

			// Capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Reset command args
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

			// Basic output validation for successful commands
			if !tt.expectError && len(output) == 0 {
				t.Errorf("%s: expected output but got none", tt.description)
			}
		})
	}
}

func TestConfigStruct(t *testing.T) {
	t.Run("ConfigYAMLMarshaling", func(t *testing.T) {
		config := Config{
			GatewayURL:    "http://localhost:8000",
			PlatformURL:   "http://localhost:8080",
			AuthToken:     "test-token-123456789",
			DefaultTenant: "test-tenant",
			Verbose:       true,
			Format:        "json",
		}

		// Test YAML marshaling
		yamlData, err := yaml.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config to YAML: %v", err)
		}

		// Test YAML unmarshaling
		var unmarshaledConfig Config
		err = yaml.Unmarshal(yamlData, &unmarshaledConfig)
		if err != nil {
			t.Fatalf("Failed to unmarshal config from YAML: %v", err)
		}

		// Verify data integrity
		if unmarshaledConfig.GatewayURL != config.GatewayURL {
			t.Errorf("Expected GatewayURL %s, got %s", config.GatewayURL, unmarshaledConfig.GatewayURL)
		}
		if unmarshaledConfig.PlatformURL != config.PlatformURL {
			t.Errorf("Expected PlatformURL %s, got %s", config.PlatformURL, unmarshaledConfig.PlatformURL)
		}
		if unmarshaledConfig.Verbose != config.Verbose {
			t.Errorf("Expected Verbose %v, got %v", config.Verbose, unmarshaledConfig.Verbose)
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		testCases := []struct {
			name   string
			config Config
			valid  bool
		}{
			{
				name: "ValidConfig",
				config: Config{
					GatewayURL:  "http://localhost:8000",
					PlatformURL: "http://localhost:8080",
					Format:      "json",
				},
				valid: true,
			},
			{
				name: "EmptyURLs",
				config: Config{
					GatewayURL:  "",
					PlatformURL: "",
					Format:      "json",
				},
				valid: false,
			},
			{
				name: "InvalidFormat",
				config: Config{
					GatewayURL:  "http://localhost:8000",
					PlatformURL: "http://localhost:8080",
					Format:      "invalid",
				},
				valid: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				valid := validateConfig(tc.config)
				if valid != tc.valid {
					t.Errorf("Expected validation result %v, got %v", tc.valid, valid)
				}
			})
		}
	})
}

func TestRunConfigShow(t *testing.T) {
	t.Run("ConfigShowExecution", func(t *testing.T) {
		// Setup test config
		setupTestConfig(t)
		defer cleanupTestConfig(t)

		// Set some test values
		viper.Set("gateway_url", "http://test-gateway:8000")
		viper.Set("platform_url", "http://test-platform:8080")
		viper.Set("verbose", true)
		viper.Set("format", "json")

		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute config show
		err := runConfigShow(configShowCmd, []string{})

		// Restore output
		w.Close()
		os.Stdout = old

		// Read output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify execution
		if err != nil {
			t.Errorf("runConfigShow failed: %v", err)
		}

		// Check for expected output elements
		expectedStrings := []string{
			"BackSaas CLI Configuration",
			"Gateway URL:",
			"Platform URL:",
			"http://test-gateway:8000",
			"http://test-platform:8080",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected '%s' in output, but not found", expected)
			}
		}
	})
}

func TestRunConfigSet(t *testing.T) {
	t.Run("ConfigSetValidKey", func(t *testing.T) {
		// Setup test config
		setupTestConfig(t)
		defer cleanupTestConfig(t)

		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute config set
		err := runConfigSet(configSetCmd, []string{"gateway_url", "http://new-gateway:8000"})

		// Restore output
		w.Close()
		os.Stdout = old

		// Read output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify execution
		if err != nil {
			t.Errorf("runConfigSet failed: %v", err)
		}

		// Verify the value was set
		if viper.GetString("gateway_url") != "http://new-gateway:8000" {
			t.Errorf("Expected gateway_url to be set to 'http://new-gateway:8000', got '%s'", viper.GetString("gateway_url"))
		}

		// Check output contains success message
		if !strings.Contains(output, "Configuration updated") {
			t.Error("Expected success message in output")
		}
	})
}

func TestConfigCommandStructure(t *testing.T) {
	t.Run("ConfigCommandHasSubcommands", func(t *testing.T) {
		subcommands := configCmd.Commands()
		expectedSubcommands := []string{"show", "set", "init", "validate"}

		if len(subcommands) != len(expectedSubcommands) {
			t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
		}

		for _, expected := range expectedSubcommands {
			found := false
			for _, cmd := range subcommands {
				if cmd.Use == expected || strings.HasPrefix(cmd.Use, expected+" ") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected subcommand '%s' not found", expected)
			}
		}
	})

	t.Run("ConfigCommandMetadata", func(t *testing.T) {
		if configCmd.Use != "config" {
			t.Errorf("Expected Use 'config', got '%s'", configCmd.Use)
		}
		if configCmd.Short == "" {
			t.Error("Expected non-empty Short description")
		}
		if configCmd.Long == "" {
			t.Error("Expected non-empty Long description")
		}
	})
}

func TestConfigFileOperations(t *testing.T) {
	t.Run("ConfigFileCreation", func(t *testing.T) {
		// Create temporary directory
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		// Create test config
		config := Config{
			GatewayURL:  "http://localhost:8000",
			PlatformURL: "http://localhost:8080",
			Format:      "json",
			Verbose:     false,
		}

		// Write config to file
		data, err := yaml.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		err = os.WriteFile(configFile, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Read and verify config file
		readData, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var readConfig Config
		err = yaml.Unmarshal(readData, &readConfig)
		if err != nil {
			t.Fatalf("Failed to unmarshal config: %v", err)
		}

		// Verify config integrity
		if readConfig.GatewayURL != config.GatewayURL {
			t.Errorf("Config file integrity check failed for GatewayURL")
		}
	})
}

// Helper functions for testing
func setupTestConfig(t *testing.T) {
	// Create temporary config directory
	tmpDir := t.TempDir()
	viper.AddConfigPath(tmpDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set default values
	viper.SetDefault("gateway_url", "http://localhost:8000")
	viper.SetDefault("platform_url", "http://localhost:8080")
	viper.SetDefault("format", "table")
	viper.SetDefault("verbose", false)
}

func cleanupTestConfig(t *testing.T) {
	// Reset viper
	viper.Reset()
}

// validateConfig validates a configuration struct
func validateConfig(config Config) bool {
	// Check required URLs
	if config.GatewayURL == "" || config.PlatformURL == "" {
		return false
	}

	// Check valid format
	validFormats := []string{"json", "table", "yaml"}
	formatValid := false
	for _, format := range validFormats {
		if config.Format == format {
			formatValid = true
			break
		}
	}

	return formatValid
}

// Benchmark tests
func BenchmarkConfigShow(b *testing.B) {
	// Setup
	setupTestConfigForBench()
	defer cleanupTestConfigForBench()

	// Redirect output to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runConfigShow(configShowCmd, []string{})
	}
}

func BenchmarkConfigYAMLMarshal(b *testing.B) {
	config := Config{
		GatewayURL:    "http://localhost:8000",
		PlatformURL:   "http://localhost:8080",
		AuthToken:     "benchmark-token",
		DefaultTenant: "benchmark-tenant",
		Verbose:       true,
		Format:        "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		yaml.Marshal(config)
	}
}

func setupTestConfigForBench() {
	viper.SetDefault("gateway_url", "http://localhost:8000")
	viper.SetDefault("platform_url", "http://localhost:8080")
	viper.SetDefault("format", "table")
	viper.SetDefault("verbose", false)
}

func cleanupTestConfigForBench() {
	viper.Reset()
}
