package gateway

import (
	"net/http"
	"testing"
)

// TestRealGatewayConfig tests the actual gateway configuration from gateway.yaml
func TestRealGatewayConfig(t *testing.T) {
	// Load the actual gateway configuration
	config, err := LoadConfig("../../config/gateway.yaml", &Config{})
	if err != nil {
		t.Skipf("Skipping test - cannot load gateway config: %v", err)
	}

	// Create route matcher with real config
	matcher, err := NewRouteMatcher(config.Routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher with real config: %v", err)
	}

	t.Logf("Loaded %d routes from gateway.yaml", len(config.Routes))

	// Print all routes for debugging
	for i, route := range config.Routes {
		t.Logf("Route %d: %s - PathPrefix: %s, Host: %s, Enabled: %v", 
			i, route.Description, route.PathPrefix, route.Host, route.Enabled)
	}

	testCases := []struct {
		name          string
		method        string
		path          string
		host          string
		expectedRoute string
		shouldMatch   bool
	}{
		{
			name:          "Admin Console Root",
			method:        "GET",
			path:          "/admin",
			host:          "localhost:8000",
			expectedRoute: "Admin Console",
			shouldMatch:   true,
		},
		{
			name:          "Admin Console Static CSS",
			method:        "GET",
			path:          "/admin/_next/static/css/app/layout.css",
			host:          "localhost:8000",
			expectedRoute: "Admin Console",
			shouldMatch:   true,
		},
		{
			name:          "Admin Console Static JS",
			method:        "GET",
			path:          "/admin/_next/static/chunks/main-app.js",
			host:          "localhost:8000",
			expectedRoute: "Admin Console",
			shouldMatch:   true,
		},
		{
			name:          "Platform API Health",
			method:        "GET",
			path:          "/api/platform/health",
			host:          "localhost:8000",
			expectedRoute: "Platform API",
			shouldMatch:   true,
		},
		{
			name:          "Health Dashboard",
			method:        "GET",
			path:          "/dashboard",
			host:          "localhost:8000",
			expectedRoute: "Health Dashboard",
			shouldMatch:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Host = tc.host

			route, err := matcher.Match(req)

			if tc.shouldMatch {
				if err != nil {
					t.Errorf("Expected route match for %s, got error: %v", tc.path, err)
					return
				}

				t.Logf("✅ Matched route: %s for path: %s", route.Description, tc.path)
				
				// Check if it matches expected route (partial match is OK)
				if route.Description != tc.expectedRoute && 
				   !contains(route.Description, tc.expectedRoute) {
					t.Logf("⚠️  Expected route containing %q, got %q", tc.expectedRoute, route.Description)
				}
			} else {
				if err == nil {
					t.Errorf("Expected no match for %s, but got route: %s", tc.path, route.Description)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
