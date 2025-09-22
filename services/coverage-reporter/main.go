package main

import (
	"log"
	"os"

	"github.com/backsaas/platform/coverage-reporter/internal/api"
	"github.com/backsaas/platform/coverage-reporter/internal/collector"
	"github.com/backsaas/platform/coverage-reporter/internal/config"
	"github.com/backsaas/platform/coverage-reporter/internal/storage"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize storage
	store, err := storage.New(cfg.StorageType, cfg.StorageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize coverage collector
	collector := collector.New(store, cfg.Services)

	// Initialize API server
	server := api.New(store, collector, cfg)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("🚀 Coverage Reporter starting on port %s", port)
	log.Printf("📊 Dashboard available at: http://localhost:%s", port)
	log.Printf("🔍 API available at: http://localhost:%s/api", port)

	if err := server.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
