package main

import (
	"log"
	"os"

	"github.com/backsaas/platform/services/platform-api/internal/api"
)

func main() {
	// Configuration
	config := &api.Config{
		TenantID:     getEnv("TENANT_ID", "test-tenant"),
		SchemaSource: "file",
		SchemaPath:   getEnv("SCHEMA_PATH", "./testdata/sample-crm.yaml"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/backsaas_dev?sslmode=disable"),
		Port:         getEnv("PORT", "8081"),
	}

	log.Printf("Starting test server with config:")
	log.Printf("  Tenant ID: %s", config.TenantID)
	log.Printf("  Schema Path: %s", config.SchemaPath)
	log.Printf("  Database URL: %s", maskPassword(config.DatabaseURL))
	log.Printf("  Port: %s", config.Port)

	// Create and start the engine
	engine, err := api.NewEngine(config)
	if err != nil {
		log.Fatalf("Failed to create API engine: %v", err)
	}

	log.Printf("🚀 Test server starting on port %s", config.Port)
	log.Printf("📊 Available endpoints:")
	log.Printf("  GET    /health                 - Health check")
	log.Printf("  GET    /schema                 - Schema information")
	log.Printf("  GET    /api/contacts           - List contacts")
	log.Printf("  POST   /api/contacts           - Create contact")
	log.Printf("  GET    /api/contacts/{id}      - Get contact")
	log.Printf("  PUT    /api/contacts/{id}      - Update contact")
	log.Printf("  DELETE /api/contacts/{id}      - Delete contact")
	log.Printf("  GET    /api/companies          - List companies")
	log.Printf("  POST   /api/companies          - Create company")
	log.Printf("  GET    /api/deals              - List deals")
	log.Printf("  POST   /api/deals              - Create deal")
	log.Printf("")
	log.Printf("💡 Example usage:")
	log.Printf("  curl http://localhost:%s/health", config.Port)
	log.Printf("  curl http://localhost:%s/api/contacts", config.Port)

	if err := engine.Start(config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskPassword(url string) string {
	// Simple password masking for logs
	// In a real implementation, use a proper URL parser
	if len(url) > 20 {
		return url[:20] + "***"
	}
	return "***"
}
