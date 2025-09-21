package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	gatewayURL  string
	platformURL string
	verbose     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "backsaas",
	Short: "BackSaas Platform Administration CLI",
	Long: `BackSaas CLI - A powerful command-line tool for managing the BackSaas platform.

This CLI provides commands for:
- Managing tenants and users
- Validating and deploying schemas
- Monitoring system health with real-time dashboard
- Bootstrapping the platform
- Debugging and troubleshooting
- Comprehensive end-to-end platform testing

Examples:
  backsaas dashboard                       # Real-time platform monitoring
  backsaas health check                    # Check all services
  backsaas tenant create acme-corp         # Create a new tenant
  backsaas schema validate ./schema.yaml   # Validate a schema file
  backsaas bootstrap --admin-email=admin@example.com  # Bootstrap platform
  backsaas test platform                   # Run comprehensive platform tests
  backsaas test tenant-lifecycle           # Test complete tenant lifecycle`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.backsaas.yaml)")
	rootCmd.PersistentFlags().StringVar(&gatewayURL, "gateway-url", "http://localhost:8000", "API Gateway URL")
	rootCmd.PersistentFlags().StringVar(&platformURL, "platform-url", "http://localhost:8080", "Platform API URL")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("gateway_url", rootCmd.PersistentFlags().Lookup("gateway-url"))
	viper.BindPFlag("platform_url", rootCmd.PersistentFlags().Lookup("platform-url"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add subcommands
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(bootstrapCmd)
	rootCmd.AddCommand(tenantCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(configCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".backsaas" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".backsaas")
	}

	// Environment variables
	viper.SetEnvPrefix("BACKSAAS")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("gateway_url", "http://localhost:8000")
	viper.SetDefault("platform_url", "http://localhost:8080")
	viper.SetDefault("verbose", false)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
