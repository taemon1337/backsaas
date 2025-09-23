package gateway

import (
	"net/http"
	"testing"
)

// TestAdminConsoleRouting tests the specific admin console routing scenarios
func TestAdminConsoleRouting(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Admin Console",
			PathPrefix:  "/admin",
			Backend:     BackendConfig{URL: "http://admin-console:3000"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	testCases := []struct {
		path     string
		expected bool
	}{
		{"/admin", true},
		{"/admin/dashboard", true},
		{"/admin/_next/static/css/app.css", true},
		{"/admin/_next/static/js/main.js", true},
		{"/other", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.path, nil)
			req.Host = "localhost:8000"

			route, err := matcher.Match(req)
			
			if tc.expected {
				if err != nil {
					t.Errorf("Expected match for %s, got error: %v", tc.path, err)
				} else {
					t.Logf("✅ Matched: %s -> %s", tc.path, route.Description)
				}
			} else {
				if err == nil {
					t.Errorf("Expected no match for %s, got: %s", tc.path, route.Description)
				}
			}
		})
	}
}
