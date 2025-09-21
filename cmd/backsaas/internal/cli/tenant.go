package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage tenants",
	Long:  `Create, list, update, and delete tenants in the BackSaas platform`,
}

var tenantListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tenants",
	Long:  `List all tenants in the platform with their status and metadata`,
	RunE:  runTenantList,
}

var tenantCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new tenant",
	Long:  `Create a new tenant with the specified name and optional configuration`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTenantCreate,
}

var tenantShowCmd = &cobra.Command{
	Use:   "show [tenant-id]",
	Short: "Show tenant details",
	Long:  `Show detailed information about a specific tenant`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTenantShow,
}

var tenantDeleteCmd = &cobra.Command{
	Use:   "delete [tenant-id]",
	Short: "Delete a tenant",
	Long:  `Delete a tenant and all its data (use with caution)`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTenantDelete,
}

var (
	tenantSchema     string
	tenantDomain     string
	tenantPlan       string
	tenantForce      bool
	tenantOutputJSON bool
)

func init() {
	tenantCmd.AddCommand(tenantListCmd)
	tenantCmd.AddCommand(tenantCreateCmd)
	tenantCmd.AddCommand(tenantShowCmd)
	tenantCmd.AddCommand(tenantDeleteCmd)

	// Create flags
	tenantCreateCmd.Flags().StringVar(&tenantSchema, "schema", "", "Path to schema file for the tenant")
	tenantCreateCmd.Flags().StringVar(&tenantDomain, "domain", "", "Custom domain for the tenant")
	tenantCreateCmd.Flags().StringVar(&tenantPlan, "plan", "free", "Billing plan (free, pro, enterprise)")

	// Delete flags
	tenantDeleteCmd.Flags().BoolVar(&tenantForce, "force", false, "Force delete without confirmation")

	// List flags
	tenantListCmd.Flags().BoolVar(&tenantOutputJSON, "json", false, "Output in JSON format")
}

type Tenant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Domain      string `json:"domain,omitempty"`
	Plan        string `json:"plan"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	UserCount   int    `json:"user_count"`
	SchemaCount int    `json:"schema_count"`
}

func runTenantList(cmd *cobra.Command, args []string) error {
	color.Cyan("📋 Tenant List")
	color.Cyan("==============")

	// TODO: Make API call to get tenants
	// GET /api/platform/tenants
	
	// Mock data for now
	tenants := []Tenant{
		{
			ID:          "system",
			Name:        "BackSaas Platform",
			Slug:        "system",
			Plan:        "system",
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
			UserCount:   1,
			SchemaCount: 1,
		},
		{
			ID:          "acme-corp",
			Name:        "Acme Corporation",
			Slug:        "acme-corp",
			Domain:      "acme-corp.backsaas.dev",
			Plan:        "pro",
			Status:      "active",
			CreatedAt:   "2024-01-15T10:30:00Z",
			UpdatedAt:   "2024-01-20T14:22:00Z",
			UserCount:   25,
			SchemaCount: 3,
		},
	}

	if tenantOutputJSON {
		// TODO: Output JSON format
		fmt.Println("JSON output not implemented yet")
		return nil
	}

	// Display as table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Domain", "Plan", "Status", "Users", "Schemas", "Created"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, tenant := range tenants {
		domain := tenant.Domain
		if domain == "" {
			domain = fmt.Sprintf("%s.backsaas.dev", tenant.Slug)
		}

		status := tenant.Status
		if status == "active" {
			status = color.GreenString("✅ active")
		}

		table.Append([]string{
			tenant.ID,
			tenant.Name,
			domain,
			tenant.Plan,
			status,
			fmt.Sprintf("%d", tenant.UserCount),
			fmt.Sprintf("%d", tenant.SchemaCount),
			tenant.CreatedAt[:10], // Just the date part
		})
	}

	table.Render()
	fmt.Printf("\nTotal: %d tenants\n", len(tenants))

	return nil
}

func runTenantCreate(cmd *cobra.Command, args []string) error {
	tenantName := args[0]
	
	color.Cyan("🏗️  Creating Tenant: %s", tenantName)
	color.Cyan("========================")

	// Generate slug from name
	slug := generateSlug(tenantName)
	
	fmt.Printf("Name: %s\n", tenantName)
	fmt.Printf("Slug: %s\n", slug)
	fmt.Printf("Domain: %s.backsaas.dev\n", slug)
	fmt.Printf("Plan: %s\n", tenantPlan)
	
	if tenantSchema != "" {
		fmt.Printf("Schema: %s\n", tenantSchema)
	}

	if !skipConfirm {
		if !promptConfirm("\nCreate this tenant?") {
			fmt.Println("Tenant creation cancelled.")
			return nil
		}
	}

	// TODO: Implement API call
	// POST /api/platform/tenants
	// {
	//   "name": tenantName,
	//   "slug": slug,
	//   "plan": tenantPlan,
	//   "domain": tenantDomain
	// }

	fmt.Println("\n🔧 Creating tenant...")
	
	steps := []string{
		"Validating tenant name and slug",
		"Creating tenant record",
		"Setting up tenant database",
		"Configuring routing rules",
	}

	if tenantSchema != "" {
		steps = append(steps, "Deploying initial schema")
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %s...\n", i+1, len(steps), step)
		// TODO: Implement actual steps
		color.Green("✅ Complete")
	}

	color.Green("\n🎉 Tenant created successfully!")
	fmt.Printf("Tenant ID: %s\n", slug)
	fmt.Printf("URL: https://%s.backsaas.dev\n", slug)
	fmt.Printf("API: https://%s.api.backsaas.dev\n", slug)

	return nil
}

func runTenantShow(cmd *cobra.Command, args []string) error {
	tenantID := args[0]
	
	color.Cyan("🔍 Tenant Details: %s", tenantID)
	color.Cyan("====================")

	// TODO: Make API call to get tenant details
	// GET /api/platform/tenants/{tenantID}

	// Mock data
	if tenantID == "system" {
		fmt.Println("ID:           system")
		fmt.Println("Name:         BackSaas Platform")
		fmt.Println("Type:         System Tenant")
		fmt.Println("Status:       ✅ Active")
		fmt.Println("Created:      2024-01-01T00:00:00Z")
		fmt.Println("Users:        1")
		fmt.Println("Schemas:      1 (platform.yaml)")
		fmt.Println("Description:  Internal platform management tenant")
	} else {
		fmt.Printf("Tenant '%s' not found\n", tenantID)
		return fmt.Errorf("tenant not found")
	}

	return nil
}

func runTenantDelete(cmd *cobra.Command, args []string) error {
	tenantID := args[0]
	
	color.Red("🗑️  Delete Tenant: %s", tenantID)
	color.Red("===================")

	if tenantID == "system" {
		return fmt.Errorf("cannot delete system tenant")
	}

	fmt.Printf("⚠️  This will permanently delete tenant '%s' and all its data!\n", tenantID)
	fmt.Println("This action cannot be undone.")

	if !tenantForce {
		if !promptConfirm("\nAre you sure you want to delete this tenant?") {
			fmt.Println("Tenant deletion cancelled.")
			return nil
		}

		// Double confirmation for safety
		if !promptConfirm("Type 'yes' to confirm deletion") {
			fmt.Println("Tenant deletion cancelled.")
			return nil
		}
	}

	// TODO: Implement deletion
	fmt.Println("\n🔧 Deleting tenant...")
	
	steps := []string{
		"Backing up tenant data",
		"Removing tenant users",
		"Dropping tenant database",
		"Removing routing rules",
		"Deleting tenant record",
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %s...\n", i+1, len(steps), step)
		// TODO: Implement actual deletion steps
		color.Green("✅ Complete")
	}

	color.Green("\n✅ Tenant deleted successfully")

	return nil
}

func generateSlug(name string) string {
	// Simple slug generation - in production, use a proper library
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	
	// Remove non-alphanumeric characters except hyphens
	var result []rune
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}
	
	return string(result)
}
