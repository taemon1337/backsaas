package main

import (
	"flag"
	"log"
	"os"

	"github.com/backsaas/platform/services/platform-api/internal/api"
)

func main() {
	// Command line flags
	var (
		tenantID     = flag.String("tenant-id", "", "Tenant ID (required)")
		schemaSource = flag.String("schema-source", "file", "Schema source: 'file' or 'registry'")
		schemaPath   = flag.String("schema-path", "", "Schema file path or tenant ID for registry")
		databaseURL  = flag.String("database-url", "", "Database connection URL")
		port         = flag.String("port", "8080", "Server port")
	)
	flag.Parse()

	// Environment variable fallbacks
	if *tenantID == "" {
		*tenantID = os.Getenv("TENANT_ID")
	}
	if *schemaSource == "" {
		*schemaSource = os.Getenv("SCHEMA_SOURCE")
		if *schemaSource == "" {
			*schemaSource = "file"
		}
	}
	if *schemaPath == "" {
		*schemaPath = os.Getenv("SCHEMA_PATH")
	}
	if *databaseURL == "" {
		*databaseURL = os.Getenv("DATABASE_URL")
	}
	if *port == "" {
		*port = os.Getenv("PORT")
		if *port == "" {
			*port = "8080"
		}
	}

	// Validate required parameters
	if *tenantID == "" {
		log.Fatal("tenant-id is required (use flag or TENANT_ID env var)")
	}
	if *schemaPath == "" {
		log.Fatal("schema-path is required (use flag or SCHEMA_PATH env var)")
	}
	if *databaseURL == "" {
		log.Fatal("database-url is required (use flag or DATABASE_URL env var)")
	}

	// Create API engine configuration
	config := &api.Config{
		TenantID:     *tenantID,
		SchemaSource: *schemaSource,
		SchemaPath:   *schemaPath,
		DatabaseURL:  *databaseURL,
		Port:         *port,
	}

	// Create and start API engine
	engine, err := api.NewEngine(config)
	if err != nil {
		log.Fatalf("Failed to create API engine: %v", err)
	}

	log.Printf("Starting BackSaas API server...")
	log.Printf("  Tenant ID: %s", *tenantID)
	log.Printf("  Schema Source: %s", *schemaSource)
	log.Printf("  Schema Path: %s", *schemaPath)
	log.Printf("  Port: %s", *port)

	if err := engine.Start(*port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
