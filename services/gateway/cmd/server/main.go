package main

import (
	"flag"
	"log"
	"os"

	"github.com/backsaas/platform/services/gateway/internal/gateway"
)

func main() {
	// Command line flags
	var (
		port        = flag.String("port", getEnv("PORT", "8000"), "Gateway port")
		redisURL    = flag.String("redis-url", getEnv("REDIS_URL", "redis://localhost:6379"), "Redis URL for rate limiting and caching")
		jwtSecret   = flag.String("jwt-secret", getEnv("JWT_SECRET", ""), "JWT secret for token validation")
		configPath  = flag.String("config", getEnv("GATEWAY_CONFIG", "config/gateway.yaml"), "Gateway configuration file")
		environment = flag.String("env", getEnv("ENVIRONMENT", "development"), "Environment (development, staging, production)")
	)
	flag.Parse()

	// Validate required configuration
	if *jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	// Create gateway configuration
	config := &gateway.Config{
		Port:        *port,
		RedisURL:    *redisURL,
		JWTSecret:   *jwtSecret,
		ConfigPath:  *configPath,
		Environment: *environment,
	}

	// Create and start gateway
	gw, err := gateway.NewGateway(config)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	log.Printf("Starting BackSaas API Gateway on port %s", *port)
	log.Printf("Environment: %s", *environment)
	log.Printf("Config: %s", *configPath)

	if err := gw.Start(); err != nil {
		log.Fatalf("Gateway failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
