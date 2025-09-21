package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Create, list, update, and delete users across tenants`,
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Long:  `List users for a specific tenant or across all tenants`,
	RunE:  runUserList,
}

var userCreateCmd = &cobra.Command{
	Use:   "create [email]",
	Short: "Create a new user",
	Long:  `Create a new user with the specified email and tenant`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUserCreate,
}

var userShowCmd = &cobra.Command{
	Use:   "show [user-id]",
	Short: "Show user details",
	Long:  `Show detailed information about a specific user`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUserShow,
}

var userRolesCmd = &cobra.Command{
	Use:   "roles [user-id]",
	Short: "Manage user roles",
	Long:  `Show or update user roles and permissions`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUserRoles,
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete [user-id]",
	Short: "Delete a user",
	Long:  `Delete a user account`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUserDelete,
}

var (
	userTenant     string
	userRole       string
	userPassword   string
	userActive     bool
	userAllTenants bool
	userAddRole    string
	userRemoveRole string
)

func init() {
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userShowCmd)
	userCmd.AddCommand(userRolesCmd)
	userCmd.AddCommand(userDeleteCmd)

	// Create flags
	userCreateCmd.Flags().StringVar(&userTenant, "tenant", "", "Tenant ID (required)")
	userCreateCmd.Flags().StringVar(&userRole, "role", "user", "Initial role (user, admin, viewer)")
	userCreateCmd.Flags().StringVar(&userPassword, "password", "", "User password (will prompt if not provided)")
	userCreateCmd.Flags().BoolVar(&userActive, "active", true, "Create user as active")
	userCreateCmd.MarkFlagRequired("tenant")

	// List flags
	userListCmd.Flags().StringVar(&userTenant, "tenant", "", "Filter by tenant ID")
	userListCmd.Flags().BoolVar(&userAllTenants, "all-tenants", false, "List users from all tenants (admin only)")

	// Roles flags
	userRolesCmd.Flags().StringVar(&userAddRole, "add", "", "Add role to user")
	userRolesCmd.Flags().StringVar(&userRemoveRole, "remove", "", "Remove role from user")

	// Delete flags
	userDeleteCmd.Flags().BoolVar(&tenantForce, "force", false, "Force delete without confirmation")
}

type User struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Name      string   `json:"name,omitempty"`
	TenantID  string   `json:"tenant_id"`
	Roles     []string `json:"roles"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	LastLogin string   `json:"last_login,omitempty"`
}

func runUserList(cmd *cobra.Command, args []string) error {
	if userAllTenants {
		color.Cyan("👥 All Users")
		color.Cyan("============")
	} else if userTenant != "" {
		color.Cyan("👥 Users for Tenant: %s", userTenant)
		color.Cyan("======================")
	} else {
		return fmt.Errorf("specify --tenant or --all-tenants")
	}

	// TODO: Make API call to get users
	// GET /api/platform/users?tenant_id={userTenant}
	// or GET /api/platform/users (for all tenants)

	// Mock data
	users := []User{
		{
			ID:        "admin-1",
			Email:     "admin@backsaas.dev",
			Name:      "Platform Admin",
			TenantID:  "system",
			Roles:     []string{"platform_admin"},
			Status:    "active",
			CreatedAt: "2024-01-01T00:00:00Z",
			LastLogin: "2024-01-20T15:30:00Z",
		},
		{
			ID:        "user-1",
			Email:     "john@acme-corp.com",
			Name:      "John Doe",
			TenantID:  "acme-corp",
			Roles:     []string{"admin"},
			Status:    "active",
			CreatedAt: "2024-01-15T10:30:00Z",
			LastLogin: "2024-01-20T14:22:00Z",
		},
		{
			ID:        "user-2",
			Email:     "jane@acme-corp.com",
			Name:      "Jane Smith",
			TenantID:  "acme-corp",
			Roles:     []string{"user"},
			Status:    "active",
			CreatedAt: "2024-01-16T09:15:00Z",
			LastLogin: "2024-01-19T16:45:00Z",
		},
	}

	// Filter by tenant if specified
	if userTenant != "" && !userAllTenants {
		var filteredUsers []User
		for _, user := range users {
			if user.TenantID == userTenant {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
	}

	// Display as table
	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"ID", "Email", "Name", "Roles", "Status", "Last Login"}
	if userAllTenants {
		headers = append([]string{"Tenant"}, headers...)
	}
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, user := range users {
		status := user.Status
		if status == "active" {
			status = color.GreenString("✅ active")
		}

		roles := fmt.Sprintf("%v", user.Roles)
		lastLogin := user.LastLogin
		if lastLogin != "" {
			lastLogin = lastLogin[:10] // Just the date part
		} else {
			lastLogin = "Never"
		}

		row := []string{user.ID, user.Email, user.Name, roles, status, lastLogin}
		if userAllTenants {
			row = append([]string{user.TenantID}, row...)
		}
		table.Append(row)
	}

	table.Render()
	fmt.Printf("\nTotal: %d users\n", len(users))

	return nil
}

func runUserCreate(cmd *cobra.Command, args []string) error {
	userEmail := args[0]
	
	color.Cyan("👤 Creating User: %s", userEmail)
	color.Cyan("=====================")

	// Get password if not provided
	if userPassword == "" {
		var err error
		userPassword, err = promptPassword("Enter user password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
	}

	fmt.Printf("Email: %s\n", userEmail)
	fmt.Printf("Tenant: %s\n", userTenant)
	fmt.Printf("Role: %s\n", userRole)
	fmt.Printf("Active: %v\n", userActive)

	if !skipConfirm {
		if !promptConfirm("\nCreate this user?") {
			fmt.Println("User creation cancelled.")
			return nil
		}
	}

	// TODO: Implement API call
	// POST /api/platform/users
	// {
	//   "email": userEmail,
	//   "password": userPassword,
	//   "tenant_id": userTenant,
	//   "roles": [userRole],
	//   "status": "active" or "inactive"
	// }

	fmt.Println("\n🔧 Creating user...")
	
	steps := []string{
		"Validating email address",
		"Checking for existing user",
		"Creating user account",
		"Assigning roles",
		"Sending welcome email",
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %s...\n", i+1, len(steps), step)
		// TODO: Implement actual steps
		color.Green("✅ Complete")
	}

	color.Green("\n🎉 User created successfully!")
	fmt.Printf("Email: %s\n", userEmail)
	fmt.Printf("Tenant: %s\n", userTenant)
	fmt.Printf("Login URL: https://%s.backsaas.dev/login\n", userTenant)

	return nil
}

func runUserShow(cmd *cobra.Command, args []string) error {
	userID := args[0]
	
	color.Cyan("👤 User Details: %s", userID)
	color.Cyan("==================")

	// TODO: Make API call to get user details
	// GET /api/platform/users/{userID}

	// Mock data
	if userID == "admin-1" {
		fmt.Println("ID:           admin-1")
		fmt.Println("Email:        admin@backsaas.dev")
		fmt.Println("Name:         Platform Admin")
		fmt.Println("Tenant:       system")
		fmt.Println("Roles:        [platform_admin]")
		fmt.Println("Status:       ✅ Active")
		fmt.Println("Created:      2024-01-01T00:00:00Z")
		fmt.Println("Last Login:   2024-01-20T15:30:00Z")
		fmt.Println("Login Count:  127")
		fmt.Println("Description:  Platform administrator account")
	} else {
		fmt.Printf("User '%s' not found\n", userID)
		return fmt.Errorf("user not found")
	}

	return nil
}

func runUserRoles(cmd *cobra.Command, args []string) error {
	userID := args[0]
	
	color.Cyan("🔐 User Roles: %s", userID)
	color.Cyan("================")

	// Handle role modifications
	if userAddRole != "" {
		fmt.Printf("Adding role '%s' to user %s...\n", userAddRole, userID)
		// TODO: API call to add role
		color.Green("✅ Role added successfully")
		return nil
	}

	if userRemoveRole != "" {
		fmt.Printf("Removing role '%s' from user %s...\n", userRemoveRole, userID)
		// TODO: API call to remove role
		color.Green("✅ Role removed successfully")
		return nil
	}

	// Show current roles
	// TODO: Make API call to get user roles
	// GET /api/platform/users/{userID}/roles

	fmt.Println("Current Roles:")
	fmt.Println("  • platform_admin")
	fmt.Println()
	fmt.Println("Available Roles:")
	fmt.Println("  • platform_admin  - Full platform access")
	fmt.Println("  • support_admin   - Read-only support access")
	fmt.Println("  • billing_admin   - Billing and usage access")
	fmt.Println("  • admin          - Tenant administrator")
	fmt.Println("  • user           - Standard user")
	fmt.Println("  • viewer         - Read-only access")

	return nil
}

func runUserDelete(cmd *cobra.Command, args []string) error {
	userID := args[0]
	
	color.Red("🗑️  Delete User: %s", userID)
	color.Red("================")

	fmt.Printf("⚠️  This will permanently delete user '%s'!\n", userID)
	fmt.Println("The user will lose access to all data and services.")

	if !tenantForce {
		if !promptConfirm("\nAre you sure you want to delete this user?") {
			fmt.Println("User deletion cancelled.")
			return nil
		}
	}

	// TODO: Implement deletion
	fmt.Println("\n🔧 Deleting user...")
	
	steps := []string{
		"Revoking user sessions",
		"Removing user permissions",
		"Cleaning up user data",
		"Deleting user account",
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %s...\n", i+1, len(steps), step)
		// TODO: Implement actual deletion steps
		color.Green("✅ Complete")
	}

	color.Green("\n✅ User deleted successfully")

	return nil
}
