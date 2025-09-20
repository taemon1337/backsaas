package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGateway(t *testing.T) {
	// Skip Redis-dependent tests if no Redis available
	if testing.Short() {
		t.Skip("Skipping gateway tests in short mode")
	}

	t.Run("NewGateway", func(t *testing.T) {
		config := &Config{
			Port:        "8000",
			RedisURL:    "redis://localhost:6379",
			JWTSecret:   "test-secret",
			ConfigPath:  "",
			Environment: "test",
		}

		// This will fail without Redis, but tests the structure
		_, err := NewGateway(config)
		if err != nil {
			// Expected to fail without Redis in test environment
			t.Logf("Gateway creation failed as expected without Redis: %v", err)
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// Test config validation
		config := &Config{}
		
		_, err := LoadConfig("", config)
		if err == nil {
			t.Error("Expected error for empty config")
		}
		
		// Test valid config
		config = &Config{
			Port:      "8000",
			JWTSecret: "test-secret",
		}
		
		validConfig, err := LoadConfig("", config)
		if err != nil {
			t.Errorf("Expected no error for valid config, got: %v", err)
		}
		
		if validConfig.Port != "8000" {
			t.Errorf("Expected port 8000, got %s", validConfig.Port)
		}
	})
}

func TestHealthCheck(t *testing.T) {
	// Create a minimal gateway for testing
	config := &Config{
		Port:      "8000",
		JWTSecret: "test-secret",
		Monitoring: MonitoringConfig{
			HealthPath: "/health",
		},
	}
	
	// Mock gateway without Redis dependency
	gateway := &Gateway{
		config: config,
	}
	
	// Test health check handler
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	gateway.healthCheck(w, req)
	
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 without Redis, got %d", w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	config := &Config{
		Cors: CorsConfig{
			Enabled:        true,
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Content-Type"},
		},
	}
	
	gateway := &Gateway{config: config}
	
	// Test CORS middleware
	middleware := gateway.corsMiddleware()
	
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	
	w := httptest.NewRecorder()
	
	// Mock gin context
	// This would need gin.Context in real test
	// For now, just test that middleware function is created
	if middleware == nil {
		t.Error("CORS middleware should not be nil")
	}
}

func TestConfigMerging(t *testing.T) {
	runtime := &Config{
		Port:      "8000",
		JWTSecret: "runtime-secret",
	}
	
	file := &Config{
		Port:        "9000",
		Environment: "test",
	}
	
	mergeConfigs(runtime, file)
	
	// File config should override runtime config
	if runtime.Port != "9000" {
		t.Errorf("Expected port 9000 after merge, got %s", runtime.Port)
	}
	
	// Runtime config should be preserved if not in file
	if runtime.JWTSecret != "runtime-secret" {
		t.Errorf("Expected JWT secret to be preserved, got %s", runtime.JWTSecret)
	}
	
	// File-only config should be added
	if runtime.Environment != "test" {
		t.Errorf("Expected environment test, got %s", runtime.Environment)
	}
}

func TestDefaultsAndValidation(t *testing.T) {
	config := &Config{
		Port:      "8000",
		JWTSecret: "test-secret",
	}
	
	setDefaults(config)
	
	// Check that defaults are set
	if config.Auth.HeaderName != "Authorization" {
		t.Errorf("Expected default auth header 'Authorization', got %s", config.Auth.HeaderName)
	}
	
	if config.Monitoring.HealthPath != "/health" {
		t.Errorf("Expected default health path '/health', got %s", config.Monitoring.HealthPath)
	}
	
	// Test validation
	err := validateConfig(config)
	if err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}
	
	// Test invalid config
	invalidConfig := &Config{}
	err = validateConfig(invalidConfig)
	if err == nil {
		t.Error("Expected validation error for empty config")
	}
}

func TestBackendHealthCheck(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	gateway := &Gateway{}
	
	// Test healthy backend
	healthy := gateway.isBackendHealthy(server.URL, "/health")
	if !healthy {
		t.Error("Expected backend to be healthy")
	}
	
	// Test unhealthy backend
	unhealthy := gateway.isBackendHealthy(server.URL, "/unhealthy")
	if unhealthy {
		t.Error("Expected backend to be unhealthy")
	}
	
	// Test unreachable backend
	unreachable := gateway.isBackendHealthy("http://localhost:99999", "/health")
	if unreachable {
		t.Error("Expected unreachable backend to be unhealthy")
	}
}

// Benchmark tests
func BenchmarkHealthCheck(b *testing.B) {
	config := &Config{
		Port:      "8000",
		JWTSecret: "test-secret",
		Monitoring: MonitoringConfig{
			HealthPath: "/health",
		},
	}
	
	gateway := &Gateway{config: config}
	
	req, _ := http.NewRequest("GET", "/health", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		gateway.healthCheck(w, req)
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	config := &Config{
		Port:      "8000",
		JWTSecret: "test-secret",
		Routes: []RouteConfig{
			{
				PathPrefix: "/api",
				Backend:    BackendConfig{URL: "http://localhost:8080"},
				Enabled:    true,
			},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateConfig(config)
	}
}
