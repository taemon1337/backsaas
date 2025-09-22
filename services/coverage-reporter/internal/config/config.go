package config

import (
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	StorageType   string
	StorageConfig map[string]string
	Services      []ServiceConfig
	RefreshRate   int // seconds
}

// ServiceConfig defines configuration for a service to monitor
type ServiceConfig struct {
	Name        string
	Path        string
	TestCommand string
	CoverageDir string
	Priority    string // high, medium, low
}

// Load loads configuration from environment variables and defaults
func Load() *Config {
	return &Config{
		StorageType: getEnv("STORAGE_TYPE", "memory"),
		StorageConfig: map[string]string{
			"path": getEnv("STORAGE_PATH", "/tmp/coverage-data"),
		},
		RefreshRate: 30, // 30 seconds
		Services: []ServiceConfig{
			{
				Name:        "platform-api",
				Path:        "/workspace/services/platform-api",
				TestCommand: "", // No test command - we'll collect existing coverage files
				CoverageDir: "/workspace/services/platform-api",
				Priority:    "high",
			},
			{
				Name:        "gateway",
				Path:        "/workspace/services/gateway",
				TestCommand: "", // No test command - we'll collect existing coverage files
				CoverageDir: "/workspace/services/gateway",
				Priority:    "high",
			},
			{
				Name:        "api",
				Path:        "/workspace/services/api",
				TestCommand: "", // No test command - we'll collect existing coverage files
				CoverageDir: "/workspace/services/api",
				Priority:    "high",
			},
			{
				Name:        "cli",
				Path:        "/workspace/cmd/backsaas",
				TestCommand: "", // No test command - we'll collect existing coverage files
				CoverageDir: "/workspace/cmd/backsaas",
				Priority:    "medium",
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetServiceByName returns a service config by name
func (c *Config) GetServiceByName(name string) *ServiceConfig {
	for _, svc := range c.Services {
		if strings.EqualFold(svc.Name, name) {
			return &svc
		}
	}
	return nil
}
