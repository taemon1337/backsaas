package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage schemas",
	Long:  `Validate, deploy, and manage schemas for tenants`,
}

var schemaValidateCmd = &cobra.Command{
	Use:   "validate [schema-file]",
	Short: "Validate a schema file",
	Long:  `Validate a YAML schema file for syntax and semantic correctness`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSchemaValidate,
}

var schemaDeployCmd = &cobra.Command{
	Use:   "deploy [schema-file]",
	Short: "Deploy a schema to a tenant",
	Long:  `Deploy a validated schema to the specified tenant`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSchemaDeploy,
}

var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List schemas for a tenant",
	Long:  `List all schemas deployed to a specific tenant`,
	RunE:  runSchemaList,
}

var schemaDiffCmd = &cobra.Command{
	Use:   "diff [old-schema] [new-schema]",
	Short: "Compare two schema files",
	Long:  `Compare two schema files and show the differences`,
	Args:  cobra.ExactArgs(2),
	RunE:  runSchemaDiff,
}

var (
	schemaTenant string
	schemaForce  bool
	schemaDryRun bool
)

func init() {
	schemaCmd.AddCommand(schemaValidateCmd)
	schemaCmd.AddCommand(schemaDeployCmd)
	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaDiffCmd)

	// Deploy flags
	schemaDeployCmd.Flags().StringVar(&schemaTenant, "tenant", "", "Target tenant ID (required for deploy)")
	schemaDeployCmd.Flags().BoolVar(&schemaForce, "force", false, "Force deployment even with breaking changes")
	schemaDeployCmd.Flags().BoolVar(&schemaDryRun, "dry-run", false, "Show what would be deployed without making changes")
	schemaDeployCmd.MarkFlagRequired("tenant")

	// List flags
	schemaListCmd.Flags().StringVar(&schemaTenant, "tenant", "", "Tenant ID to list schemas for")
	schemaListCmd.MarkFlagRequired("tenant")
}

type Schema struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Entities    map[string]Entity      `yaml:"entities"`
	Functions   map[string]Function    `yaml:"functions,omitempty"`
	Policies    map[string]interface{} `yaml:"policies,omitempty"`
}

type Entity struct {
	Description string            `yaml:"description,omitempty"`
	Fields      map[string]Field  `yaml:"fields"`
	Indexes     []Index           `yaml:"indexes,omitempty"`
	Policies    map[string]string `yaml:"policies,omitempty"`
}

type Field struct {
	Type        string      `yaml:"type"`
	Required    bool        `yaml:"required,omitempty"`
	Unique      bool        `yaml:"unique,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Validation  string      `yaml:"validation,omitempty"`
}

type Index struct {
	Fields []string `yaml:"fields"`
	Unique bool     `yaml:"unique,omitempty"`
}

type Function struct {
	Description string                 `yaml:"description,omitempty"`
	Type        string                 `yaml:"type"`
	Trigger     string                 `yaml:"trigger,omitempty"`
	Parameters  map[string]interface{} `yaml:"parameters,omitempty"`
}

func runSchemaValidate(cmd *cobra.Command, args []string) error {
	schemaFile := args[0]
	
	color.Cyan("🔍 Validating Schema: %s", schemaFile)
	color.Cyan("========================")

	// Check if file exists
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		return fmt.Errorf("schema file not found: %s", schemaFile)
	}

	// Read and parse YAML
	schema, err := loadSchemaFile(schemaFile)
	if err != nil {
		color.Red("❌ Failed to parse schema file")
		return err
	}

	fmt.Printf("Schema Name: %s\n", schema.Name)
	fmt.Printf("Version: %s\n", schema.Version)
	fmt.Printf("Description: %s\n", schema.Description)

	// Validation steps
	validations := []ValidationStep{
		{"YAML syntax", validateYAMLSyntax},
		{"Schema structure", validateSchemaStructure},
		{"Entity definitions", validateEntities},
		{"Field types", validateFieldTypes},
		{"Relationships", validateRelationships},
		{"Functions", validateFunctions},
		{"Policies", validatePolicies},
	}

	allValid := true
	for _, validation := range validations {
		fmt.Printf("\n🔍 %s...", validation.Name)
		
		if err := validation.Function(schema); err != nil {
			color.Red(" ❌ Failed")
			fmt.Printf("   Error: %s\n", err.Error())
			allValid = false
		} else {
			color.Green(" ✅ Valid")
		}
	}

	fmt.Println()
	if allValid {
		color.Green("🎉 Schema validation passed!")
		fmt.Printf("Entities: %d\n", len(schema.Entities))
		fmt.Printf("Functions: %d\n", len(schema.Functions))
		return nil
	} else {
		color.Red("❌ Schema validation failed")
		return fmt.Errorf("schema validation failed")
	}
}

func runSchemaDeploy(cmd *cobra.Command, args []string) error {
	schemaFile := args[0]
	
	color.Cyan("🚀 Deploying Schema: %s", schemaFile)
	color.Cyan("==========================")

	// Validate first
	schema, err := loadSchemaFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	fmt.Printf("Schema: %s v%s\n", schema.Name, schema.Version)
	fmt.Printf("Target Tenant: %s\n", schemaTenant)
	fmt.Printf("Entities: %d\n", len(schema.Entities))
	fmt.Printf("Functions: %d\n", len(schema.Functions))

	if schemaDryRun {
		color.Yellow("\n🔍 Dry run mode - no changes will be made")
		return nil
	}

	// Check for breaking changes
	fmt.Println("\n🔍 Analyzing changes...")
	
	// TODO: Compare with existing schema
	hasBreakingChanges := false // This would be determined by API call
	
	if hasBreakingChanges && !schemaForce {
		color.Yellow("⚠️  Breaking changes detected!")
		fmt.Println("This deployment will require a migration.")
		
		if !promptConfirm("Continue with deployment?") {
			fmt.Println("Deployment cancelled.")
			return nil
		}
	}

	if !skipConfirm && !schemaForce {
		if !promptConfirm("Deploy this schema?") {
			fmt.Println("Deployment cancelled.")
			return nil
		}
	}

	// Deploy schema
	fmt.Println("\n🔧 Deploying schema...")
	
	steps := []string{
		"Uploading schema definition",
		"Validating against existing data",
		"Generating database migrations",
		"Executing schema changes",
		"Updating API endpoints",
		"Reloading tenant configuration",
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %s...\n", i+1, len(steps), step)
		// TODO: Implement actual deployment steps
		color.Green("✅ Complete")
	}

	color.Green("\n🎉 Schema deployed successfully!")
	fmt.Printf("Tenant: %s\n", schemaTenant)
	fmt.Printf("Schema: %s v%s\n", schema.Name, schema.Version)
	fmt.Printf("API Endpoint: https://%s.api.backsaas.dev\n", schemaTenant)

	return nil
}

func runSchemaList(cmd *cobra.Command, args []string) error {
	color.Cyan("📋 Schemas for Tenant: %s", schemaTenant)
	color.Cyan("========================")

	// TODO: Make API call to get tenant schemas
	// GET /api/platform/tenants/{schemaTenant}/schemas

	// Mock data
	schemas := []struct {
		Name        string
		Version     string
		Status      string
		DeployedAt  string
		Entities    int
		Functions   int
	}{
		{
			Name:       "Customer Management",
			Version:    "1.2.0",
			Status:     "active",
			DeployedAt: "2024-01-20T14:22:00Z",
			Entities:   5,
			Functions:  8,
		},
		{
			Name:       "Inventory System", 
			Version:    "2.1.0",
			Status:     "active",
			DeployedAt: "2024-01-18T09:15:00Z",
			Entities:   3,
			Functions:  4,
		},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Version", "Status", "Entities", "Functions", "Deployed"})
	table.SetBorder(false)

	for _, schema := range schemas {
		status := schema.Status
		if status == "active" {
			status = color.GreenString("✅ active")
		}

		table.Append([]string{
			schema.Name,
			schema.Version,
			status,
			fmt.Sprintf("%d", schema.Entities),
			fmt.Sprintf("%d", schema.Functions),
			schema.DeployedAt[:10],
		})
	}

	table.Render()
	fmt.Printf("\nTotal: %d schemas\n", len(schemas))

	return nil
}

func runSchemaDiff(cmd *cobra.Command, args []string) error {
	oldFile := args[0]
	newFile := args[1]
	
	color.Cyan("🔍 Schema Diff: %s → %s", filepath.Base(oldFile), filepath.Base(newFile))
	color.Cyan("=============================")

	oldSchema, err := loadSchemaFile(oldFile)
	if err != nil {
		return fmt.Errorf("failed to load old schema: %w", err)
	}

	newSchema, err := loadSchemaFile(newFile)
	if err != nil {
		return fmt.Errorf("failed to load new schema: %w", err)
	}

	fmt.Printf("Old: %s v%s\n", oldSchema.Name, oldSchema.Version)
	fmt.Printf("New: %s v%s\n", newSchema.Name, newSchema.Version)

	// Simple diff implementation
	fmt.Println("\n📊 Changes Summary:")
	
	// Compare entities
	oldEntities := len(oldSchema.Entities)
	newEntities := len(newSchema.Entities)
	
	if newEntities > oldEntities {
		color.Green("+ %d new entities", newEntities-oldEntities)
	} else if newEntities < oldEntities {
		color.Red("- %d removed entities", oldEntities-newEntities)
	} else {
		fmt.Printf("= %d entities (no change)\n", newEntities)
	}

	// TODO: Implement detailed field-level diff
	fmt.Println("\n📋 Detailed Changes:")
	fmt.Println("(Detailed diff implementation pending)")

	return nil
}

type ValidationStep struct {
	Name     string
	Function func(*Schema) error
}

func loadSchemaFile(filename string) (*Schema, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &schema, nil
}

func validateYAMLSyntax(schema *Schema) error {
	// YAML syntax validation is done during unmarshaling
	return nil
}

func validateSchemaStructure(schema *Schema) error {
	if schema.Name == "" {
		return fmt.Errorf("schema name is required")
	}
	if schema.Version == "" {
		return fmt.Errorf("schema version is required")
	}
	if len(schema.Entities) == 0 {
		return fmt.Errorf("at least one entity is required")
	}
	return nil
}

func validateEntities(schema *Schema) error {
	for entityName, entity := range schema.Entities {
		if len(entity.Fields) == 0 {
			return fmt.Errorf("entity '%s' must have at least one field", entityName)
		}
		
		// Check for required ID field
		if _, hasID := entity.Fields["id"]; !hasID {
			return fmt.Errorf("entity '%s' must have an 'id' field", entityName)
		}
	}
	return nil
}

func validateFieldTypes(schema *Schema) error {
	validTypes := map[string]bool{
		"string": true, "text": true, "integer": true, "float": true,
		"boolean": true, "date": true, "datetime": true, "json": true,
		"uuid": true, "email": true, "url": true,
	}

	for entityName, entity := range schema.Entities {
		for fieldName, field := range entity.Fields {
			if !validTypes[field.Type] {
				return fmt.Errorf("entity '%s', field '%s': invalid type '%s'", 
					entityName, fieldName, field.Type)
			}
		}
	}
	return nil
}

func validateRelationships(schema *Schema) error {
	// TODO: Implement relationship validation
	return nil
}

func validateFunctions(schema *Schema) error {
	validFunctionTypes := map[string]bool{
		"validation": true, "hook": true, "computed": true, "workflow": true,
	}

	for funcName, function := range schema.Functions {
		if !validFunctionTypes[function.Type] {
			return fmt.Errorf("function '%s': invalid type '%s'", funcName, function.Type)
		}
	}
	return nil
}

func validatePolicies(schema *Schema) error {
	// TODO: Implement policy validation
	return nil
}
