package gateway

import (
	"net/http"
	"testing"
)

// TestCurrentGatewayConfig tests the current gateway configuration manually
func TestCurrentGatewayConfig(t *testing.T) {
	// This matches our current gateway.yaml configuration
	routes := []RouteConfig{
		{
			Description: "Platform API",
			PathPrefix:  "/api/platform",
			Backend:     BackendConfig{URL: "http://platform-api:8080"},
			Enabled:     true,
		},
		{
			Description: "Admin Console (All /admin traffic)",
			PathPrefix:  "/admin",
			Backend:     BackendConfig{URL: "http://admin-console:3000"},
			Transform: &TransformConfig{
				AddHeaders: map[string]string{
					"X-Interface-Type": "admin-console",
				},
			},
			Enabled: true,
		},
		{
			Description: "Health Dashboard",
			PathPrefix:  "/dashboard",
			Backend:     BackendConfig{URL: "http://health-dashboard:8090"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	t.Logf("Testing with %d routes", len(routes))

	testCases := []struct {
		name          string
		path          string
		expectedRoute string
		shouldMatch   bool
	}{
		// Admin Console tests
		{
			name:          "Admin Console Root",
			path:          "/admin",
			expectedRoute: "Admin Console (All /admin traffic)",
			shouldMatch:   true,
		},
		{
			name:          "Admin Console Page",
			path:          "/admin/dashboard",
			expectedRoute: "Admin Console (All /admin traffic)",
			shouldMatch:   true,
		},
		{
			name:          "Admin Console CSS Asset",
			path:          "/admin/_next/static/css/app/layout.css",
			expectedRoute: "Admin Console (All /admin traffic)",
			shouldMatch:   true,
		},
		{
			name:          "Admin Console JS Asset",
			path:          "/admin/_next/static/chunks/main-app.js",
			expectedRoute: "Admin Console (All /admin traffic)",
			shouldMatch:   true,
		},
		{
			name:          "Admin Console Webpack Asset",
			path:          "/admin/_next/static/chunks/webpack.js",
			expectedRoute: "Admin Console (All /admin traffic)",
			shouldMatch:   true,
		},
		// Platform API tests
		{
			name:          "Platform API Health",
			path:          "/api/platform/health",
			expectedRoute: "Platform API",
			shouldMatch:   true,
		},
		{
			name:          "Platform API Users",
			path:          "/api/platform/users",
			expectedRoute: "Platform API",
			shouldMatch:   true,
		},
		// Health Dashboard tests
		{
			name:          "Health Dashboard Root",
			path:          "/dashboard",
			expectedRoute: "Health Dashboard",
			shouldMatch:   true,
		},
		{
			name:          "Health Dashboard Health",
			path:          "/dashboard/health",
			expectedRoute: "Health Dashboard",
			shouldMatch:   true,
		},
		// No match tests
		{
			name:        "No Match - Root",
			path:        "/",
			shouldMatch: false,
		},
		{
			name:        "No Match - Random",
			path:        "/random/path",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Host = "localhost:8000"

			route, err := matcher.Match(req)

			if tc.shouldMatch {
				if err != nil {
					t.Errorf("❌ Expected route match for %s, got error: %v", tc.path, err)
					return
				}

				if route.Description != tc.expectedRoute {
					t.Errorf("❌ Expected route %q, got %q for path %s", 
						tc.expectedRoute, route.Description, tc.path)
					return
				}

				t.Logf("✅ SUCCESS: %s -> %s", tc.path, route.Description)
			} else {
				if err == nil {
					t.Errorf("❌ Expected no match for %s, but got route: %s", tc.path, route.Description)
					return
				}
				t.Logf("✅ SUCCESS: %s -> No match (expected)", tc.path)
			}
		})
	}
}

// TestRouteOrdering tests that route ordering works correctly
func TestRouteOrdering(t *testing.T) {
	// Test with different route orders to ensure most specific matches first
	routes := []RouteConfig{
		{
			Description: "Specific Admin Route",
			PathPrefix:  "/admin/api",
			Backend:     BackendConfig{URL: "http://admin-api:8080"},
			Enabled:     true,
		},
		{
			Description: "General Admin Route", 
			PathPrefix:  "/admin",
			Backend:     BackendConfig{URL: "http://admin-console:3000"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	// Test that more specific route wins
	req, _ := http.NewRequest("GET", "/admin/api/users", nil)
	req.Host = "localhost:8000"

	route, err := matcher.Match(req)
	if err != nil {
		t.Fatalf("Expected route match, got error: %v", err)
	}

	// Should match the more specific route due to longer path prefix
	if route.Description != "Specific Admin Route" {
		t.Errorf("Expected 'Specific Admin Route', got '%s'", route.Description)
	}

	t.Logf("✅ Route ordering works: /admin/api/users -> %s", route.Description)
}
