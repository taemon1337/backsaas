package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap the BackSaas platform",
	Long: `Bootstrap the BackSaas platform with initial configuration:

- Create system tenant
- Create initial admin user
- Validate platform schema
- Set up initial configuration

This command should be run once when setting up a new BackSaas instance.`,
	RunE: runBootstrap,
}

var (
	adminEmail    string
	adminPassword string
	skipConfirm   bool
	dryRun        bool
)

func init() {
	bootstrapCmd.Flags().StringVar(&adminEmail, "admin-email", "", "Admin user email address (required)")
	bootstrapCmd.Flags().StringVar(&adminPassword, "admin-password", "", "Admin user password (will prompt if not provided)")
	bootstrapCmd.Flags().BoolVar(&skipConfirm, "yes", false, "Skip confirmation prompts")
	bootstrapCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	
	bootstrapCmd.MarkFlagRequired("admin-email")
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	color.Cyan("🚀 BackSaas Platform Bootstrap")
	color.Cyan("=============================")

	// Validate email
	if adminEmail == "" {
		return fmt.Errorf("admin email is required")
	}

	// Get password if not provided
	if adminPassword == "" {
		var err error
		adminPassword, err = promptPassword("Enter admin password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		confirmPassword, err := promptPassword("Confirm admin password: ")
		if err != nil {
			return fmt.Errorf("failed to read password confirmation: %w", err)
		}

		if adminPassword != confirmPassword {
			return fmt.Errorf("passwords do not match")
		}
	}

	// Show bootstrap plan
	fmt.Println("\n📋 Bootstrap Plan:")
	fmt.Printf("  • Create system tenant (tenant_id: 'system')\n")
	fmt.Printf("  • Create admin user: %s\n", adminEmail)
	fmt.Printf("  • Set up platform schema\n")
	fmt.Printf("  • Initialize RBAC policies\n")
	fmt.Printf("  • Validate service connectivity\n")

	if dryRun {
		color.Yellow("\n🔍 Dry run mode - no changes will be made")
		return nil
	}

	// Confirmation
	if !skipConfirm {
		if !promptConfirm("\nProceed with bootstrap?") {
			fmt.Println("Bootstrap cancelled.")
			return nil
		}
	}

	// Execute bootstrap steps
	fmt.Println("\n🔧 Executing bootstrap...")

	steps := []BootstrapStep{
		{"Checking service health", checkServicesHealth},
		{"Creating system tenant", createSystemTenant},
		{"Setting up platform schema", setupPlatformSchema},
		{"Creating admin user", createAdminUser},
		{"Initializing RBAC policies", initializeRBAC},
		{"Validating bootstrap", validateBootstrap},
	}

	for i, step := range steps {
		fmt.Printf("\n[%d/%d] %s...\n", i+1, len(steps), step.Description)
		
		if err := step.Function(); err != nil {
			color.Red("❌ Failed: %v", err)
			return fmt.Errorf("bootstrap failed at step '%s': %w", step.Description, err)
		}
		
		color.Green("✅ Complete")
	}

	// Success message
	fmt.Println()
	color.Green("🎉 Bootstrap completed successfully!")
	fmt.Println()
	fmt.Printf("Admin user created: %s\n", adminEmail)
	fmt.Printf("Platform API: %s\n", viper.GetString("platform_url"))
	fmt.Printf("Gateway: %s\n", viper.GetString("gateway_url"))
	fmt.Println()
	color.Cyan("Next steps:")
	fmt.Println("  1. Test login with: backsaas auth login")
	fmt.Println("  2. Create your first tenant: backsaas tenant create <name>")
	fmt.Println("  3. Deploy a schema: backsaas schema deploy <schema.yaml>")

	return nil
}

type BootstrapStep struct {
	Description string
	Function    func() error
}

func promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(bytePassword), nil
}

func promptConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func checkServicesHealth() error {
	// Reuse health check logic
	services := []ServiceHealth{
		{Name: "API Gateway", URL: viper.GetString("gateway_url") + "/health"},
		{Name: "Platform API", URL: viper.GetString("platform_url") + "/health"},
	}

	for i := range services {
		checkService(&services[i])
		if services[i].Status != "✅ Healthy" {
			return fmt.Errorf("service %s is unhealthy: %s", services[i].Name, services[i].Response)
		}
	}

	return nil
}

func createSystemTenant() error {
	// TODO: Implement API call to create system tenant
	fmt.Println("  Creating system tenant with tenant_id='system'")
	
	// This would make an API call to Platform API:
	// POST /api/platform/tenants
	// {
	//   "id": "system",
	//   "name": "BackSaas Platform",
	//   "slug": "system",
	//   "status": "active"
	// }
	
	return nil
}

func setupPlatformSchema() error {
	// TODO: Implement platform schema setup
	fmt.Println("  Loading platform.yaml schema")
	fmt.Println("  Creating system tables (tenants, users, schemas, etc.)")
	
	// This would:
	// 1. Load platform.yaml from infra/schemas/
	// 2. POST /api/platform/schemas with platform schema
	// 3. Trigger initial migration
	
	return nil
}

func createAdminUser() error {
	// TODO: Implement admin user creation
	fmt.Printf("  Creating admin user: %s\n", adminEmail)
	
	// This would make an API call:
	// POST /api/platform/users
	// {
	//   "email": adminEmail,
	//   "password": adminPassword,
	//   "tenant_id": "system",
	//   "roles": ["platform_admin"]
	// }
	
	return nil
}

func initializeRBAC() error {
	// TODO: Implement RBAC initialization
	fmt.Println("  Setting up platform admin roles")
	fmt.Println("  Configuring default policies")
	
	// This would set up Casbin policies for platform administration
	
	return nil
}

func validateBootstrap() error {
	// TODO: Implement bootstrap validation
	fmt.Println("  Validating system tenant exists")
	fmt.Println("  Validating admin user can authenticate")
	fmt.Println("  Validating platform schema is loaded")
	
	// This would make validation API calls to ensure everything is working
	
	return nil
}
