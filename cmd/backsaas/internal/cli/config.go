package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `View, set, and validate CLI configuration settings`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current CLI configuration settings`,
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration key to a specific value`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Create a new configuration file with default settings`,
	RunE:  runConfigInit,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate the current configuration and test connectivity`,
	RunE:  runConfigValidate,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
}

type Config struct {
	GatewayURL  string `yaml:"gateway_url"`
	PlatformURL string `yaml:"platform_url"`
	AuthToken   string `yaml:"auth_token,omitempty"`
	DefaultTenant string `yaml:"default_tenant,omitempty"`
	Verbose     bool   `yaml:"verbose"`
	Format      string `yaml:"format"` // json, table, yaml
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	color.Cyan("⚙️  BackSaas CLI Configuration")
	color.Cyan("=============================")

	fmt.Printf("Config File: %s\n", viper.ConfigFileUsed())
	fmt.Println()

	// Display current configuration
	config := Config{
		GatewayURL:    viper.GetString("gateway_url"),
		PlatformURL:   viper.GetString("platform_url"),
		AuthToken:     viper.GetString("auth_token"),
		DefaultTenant: viper.GetString("default_tenant"),
		Verbose:       viper.GetBool("verbose"),
		Format:        viper.GetString("format"),
	}

	fmt.Printf("Gateway URL:     %s\n", config.GatewayURL)
	fmt.Printf("Platform URL:    %s\n", config.PlatformURL)
	
	if config.AuthToken != "" {
		fmt.Printf("Auth Token:      %s...%s\n", config.AuthToken[:8], config.AuthToken[len(config.AuthToken)-8:])
	} else {
		fmt.Printf("Auth Token:      (not set)\n")
	}
	
	fmt.Printf("Default Tenant:  %s\n", config.DefaultTenant)
	fmt.Printf("Verbose:         %v\n", config.Verbose)
	fmt.Printf("Output Format:   %s\n", config.Format)

	// Show environment variables
	fmt.Println("\nEnvironment Variables:")
	envVars := []string{
		"BACKSAAS_GATEWAY_URL",
		"BACKSAAS_PLATFORM_URL", 
		"BACKSAAS_AUTH_TOKEN",
		"BACKSAAS_DEFAULT_TENANT",
		"BACKSAAS_VERBOSE",
	}

	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value != "" {
			if envVar == "BACKSAAS_AUTH_TOKEN" {
				fmt.Printf("  %s: %s...%s\n", envVar, value[:8], value[len(value)-8:])
			} else {
				fmt.Printf("  %s: %s\n", envVar, value)
			}
		} else {
			fmt.Printf("  %s: (not set)\n", envVar)
		}
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	color.Cyan("⚙️  Setting Configuration")
	color.Cyan("========================")

	// Validate key
	validKeys := map[string]bool{
		"gateway_url":     true,
		"platform_url":    true,
		"auth_token":      true,
		"default_tenant":  true,
		"verbose":         true,
		"format":          true,
	}

	if !validKeys[key] {
		return fmt.Errorf("invalid configuration key: %s", key)
	}

	// Special validation for certain keys
	switch key {
	case "format":
		validFormats := map[string]bool{"json": true, "table": true, "yaml": true}
		if !validFormats[value] {
			return fmt.Errorf("invalid format: %s (valid: json, table, yaml)", value)
		}
	case "verbose":
		if value != "true" && value != "false" {
			return fmt.Errorf("verbose must be 'true' or 'false'")
		}
	}

	// Set the value
	viper.Set(key, value)

	// Write to config file
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Create default config file
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(home, ".backsaas.yaml")
	}

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	color.Green("✅ Configuration updated")
	fmt.Printf("Set %s = %s\n", key, value)
	fmt.Printf("Config file: %s\n", configFile)

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	color.Cyan("🔧 Initializing Configuration")
	color.Cyan("=============================")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configFile := filepath.Join(home, ".backsaas.yaml")

	// Check if config file already exists
	if _, err := os.Stat(configFile); err == nil {
		if !promptConfirm(fmt.Sprintf("Config file %s already exists. Overwrite?", configFile)) {
			fmt.Println("Configuration initialization cancelled.")
			return nil
		}
	}

	// Create default configuration
	defaultConfig := Config{
		GatewayURL:  "http://localhost:8000",
		PlatformURL: "http://localhost:8080",
		Verbose:     false,
		Format:      "table",
	}

	// Write config file
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	color.Green("✅ Configuration initialized")
	fmt.Printf("Config file created: %s\n", configFile)
	fmt.Println()
	fmt.Println("Default settings:")
	fmt.Printf("  Gateway URL:  %s\n", defaultConfig.GatewayURL)
	fmt.Printf("  Platform URL: %s\n", defaultConfig.PlatformURL)
	fmt.Printf("  Format:       %s\n", defaultConfig.Format)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Update URLs if needed: backsaas config set gateway_url <url>")
	fmt.Println("  2. Test connectivity: backsaas config validate")
	fmt.Println("  3. Bootstrap platform: backsaas bootstrap --admin-email=<email>")

	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	color.Cyan("🔍 Validating Configuration")
	color.Cyan("===========================")

	config := Config{
		GatewayURL:  viper.GetString("gateway_url"),
		PlatformURL: viper.GetString("platform_url"),
		AuthToken:   viper.GetString("auth_token"),
		Format:      viper.GetString("format"),
	}

	validations := []struct {
		Name     string
		Function func() error
	}{
		{"Configuration file", validateConfigFile},
		{"Gateway URL", func() error { return validateURL(config.GatewayURL, "Gateway") }},
		{"Platform URL", func() error { return validateURL(config.PlatformURL, "Platform API") }},
		{"Output format", func() error { return validateFormat(config.Format) }},
		{"Service connectivity", validateConnectivity},
	}

	allValid := true
	for _, validation := range validations {
		fmt.Printf("🔍 %s...", validation.Name)
		
		if err := validation.Function(); err != nil {
			color.Red(" ❌ Failed")
			fmt.Printf("   Error: %s\n", err.Error())
			allValid = false
		} else {
			color.Green(" ✅ Valid")
		}
	}

	fmt.Println()
	if allValid {
		color.Green("🎉 Configuration is valid!")
		return nil
	} else {
		color.Red("❌ Configuration validation failed")
		return fmt.Errorf("configuration validation failed")
	}
}

func validateConfigFile() error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return fmt.Errorf("no configuration file found")
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", configFile)
	}

	return nil
}

func validateURL(url, service string) error {
	if url == "" {
		return fmt.Errorf("%s URL is not set", service)
	}

	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("%s URL must start with http:// or https://", service)
	}

	return nil
}

func validateFormat(format string) error {
	validFormats := map[string]bool{"json": true, "table": true, "yaml": true}
	if format == "" {
		return nil // Default will be used
	}
	
	if !validFormats[format] {
		return fmt.Errorf("invalid format '%s' (valid: json, table, yaml)", format)
	}

	return nil
}

func validateConnectivity() error {
	// Reuse health check logic
	services := []ServiceHealth{
		{Name: "Gateway", URL: viper.GetString("gateway_url") + "/health"},
		{Name: "Platform API", URL: viper.GetString("platform_url") + "/health"},
	}

	for i := range services {
		checkService(&services[i])
		if services[i].Status != "✅ Healthy" {
			return fmt.Errorf("%s is not accessible", services[i].Name)
		}
	}

	return nil
}
