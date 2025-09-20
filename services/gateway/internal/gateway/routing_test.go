package gateway

import (
	"net/http"
	"testing"
)

func TestRouteMatcher(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Platform API",
			PathPrefix:  "/platform",
			Backend:     BackendConfig{URL: "http://localhost:8080"},
			Enabled:     true,
		},
		{
			Description: "Tenant API by Host",
			Host:        "*.api.example.com",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://localhost:8081"},
			Enabled:     true,
		},
		{
			Description: "Tenant API by Path",
			PathPrefix:  "/tenant/*/api",
			Backend:     BackendConfig{URL: "http://localhost:8082"},
			Enabled:     true,
		},
		{
			Description: "Auth API",
			PathPrefix:  "/auth",
			Backend:     BackendConfig{URL: "http://localhost:8083"},
			Enabled:     true,
		},
		{
			Description: "Disabled Route",
			PathPrefix:  "/disabled",
			Backend:     BackendConfig{URL: "http://localhost:8084"},
			Enabled:     false,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	t.Run("ExactPathMatch", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/platform/users", nil)
		req.Host = "api.example.com"

		route, err := matcher.Match(req)
		if err != nil {
			t.Fatalf("Expected route match, got error: %v", err)
		}

		if route.Description != "Platform API" {
			t.Errorf("Expected Platform API route, got %s", route.Description)
		}
	})

	t.Run("HostMatch", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Host = "tenant1.api.example.com"

		route, err := matcher.Match(req)
		if err != nil {
			t.Fatalf("Expected route match, got error: %v", err)
		}

		if route.Description != "Tenant API by Host" {
			t.Errorf("Expected Tenant API by Host route, got %s", route.Description)
		}
	})

	t.Run("WildcardPathMatch", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tenant/acme-corp/api/users", nil)
		req.Host = "api.example.com"

		route, err := matcher.Match(req)
		if err != nil {
			t.Fatalf("Expected route match, got error: %v", err)
		}

		if route.Description != "Tenant API by Path" {
			t.Errorf("Expected Tenant API by Path route, got %s", route.Description)
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		req.Host = "api.example.com"

		_, err := matcher.Match(req)
		if err == nil {
			t.Error("Expected no match error")
		}
	})

	t.Run("DisabledRoute", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/disabled", nil)
		req.Host = "api.example.com"

		_, err := matcher.Match(req)
		if err == nil {
			t.Error("Expected no match for disabled route")
		}
	})
}

func TestHostMatching(t *testing.T) {
	matcher := &RouteMatcher{}

	testCases := []struct {
		requestHost string
		routeHost   string
		expected    bool
	}{
		{"api.example.com", "api.example.com", true},
		{"api.example.com:8080", "api.example.com", true},
		{"tenant1.api.example.com", "*.api.example.com", true},
		{"api.example.com", "*.api.example.com", true},
		{"subdomain.tenant1.api.example.com", "*.api.example.com", false},
		{"different.com", "api.example.com", false},
		{"api.different.com", "*.api.example.com", false},
	}

	for _, tc := range testCases {
		result := matcher.matchHost(tc.requestHost, tc.routeHost)
		if result != tc.expected {
			t.Errorf("matchHost(%s, %s) = %v, expected %v", 
				tc.requestHost, tc.routeHost, result, tc.expected)
		}
	}
}

func TestPathMatching(t *testing.T) {
	routes := []RouteConfig{
		{PathPrefix: "/api/v1"},
		{PathPrefix: "/tenant/*/api"},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	testCases := []struct {
		requestPath string
		routeIndex  int
		expected    bool
	}{
		{"/api/v1/users", 0, true},
		{"/api/v1", 0, true},
		{"/api/v2", 0, false},
		{"/tenant/acme/api/users", 1, true},
		{"/tenant/xyz/api", 1, true},
		{"/tenant/api", 1, false},
		{"/different/path", 0, false},
	}

	for _, tc := range testCases {
		route := &routes[tc.routeIndex]
		result := matcher.matchPath(tc.requestPath, route.PathPrefix, tc.routeIndex)
		if result != tc.expected {
			t.Errorf("matchPath(%s, %s) = %v, expected %v", 
				tc.requestPath, route.PathPrefix, result, tc.expected)
		}
	}
}

func TestTenantIDExtraction(t *testing.T) {
	matcher := &RouteMatcher{}

	testCases := []struct {
		description string
		setupReq    func() *http.Request
		expected    string
	}{
		{
			description: "Header extraction",
			setupReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "/api/users", nil)
				req.Header.Set("X-Tenant-ID", "header-tenant")
				return req
			},
			expected: "header-tenant",
		},
		{
			description: "Subdomain extraction",
			setupReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "/api/users", nil)
				req.Host = "acme-corp.api.example.com"
				return req
			},
			expected: "acme-corp",
		},
		{
			description: "Path extraction",
			setupReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "/tenant-path/api/users", nil)
				req.Host = "api.example.com"
				return req
			},
			expected: "tenant-path",
		},
		{
			description: "Query parameter extraction",
			setupReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "/api/users?tenant_id=query-tenant", nil)
				req.Host = "api.example.com"
				return req
			},
			expected: "query-tenant",
		},
		{
			description: "No tenant ID",
			setupReq: func() *http.Request {
				req, _ := http.NewRequest("GET", "/api/users", nil)
				req.Host = "api.example.com"
				return req
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			req := tc.setupReq()
			result := matcher.extractTenantID(req)
			if result != tc.expected {
				t.Errorf("Expected tenant ID %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestRouteScoring(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Specific Host and Path",
			Host:        "api.example.com",
			PathPrefix:  "/api/v1",
			Backend:     BackendConfig{URL: "http://localhost:8080"},
			Enabled:     true,
		},
		{
			Description: "Wildcard Host",
			Host:        "*.api.example.com",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://localhost:8081"},
			Enabled:     true,
		},
		{
			Description: "Path Only",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://localhost:8082"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	// Request that matches all routes
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	req.Host = "api.example.com"

	route, err := matcher.Match(req)
	if err != nil {
		t.Fatalf("Expected route match, got error: %v", err)
	}

	// Should match the most specific route (host + longer path)
	if route.Description != "Specific Host and Path" {
		t.Errorf("Expected most specific route, got %s", route.Description)
	}
}

func TestRouteValidation(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Valid Route",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://localhost:8080"},
			Enabled:     true,
		},
		{
			Description: "Conflicting Route",
			PathPrefix:  "/api/v1",
			Backend:     BackendConfig{URL: "http://localhost:8081"},
			Enabled:     true,
		},
		{
			Description: "Broad Route (may be unreachable)",
			PathPrefix:  "/",
			Backend:     BackendConfig{URL: "http://localhost:8082"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	warnings := matcher.ValidateRoutes()
	if len(warnings) == 0 {
		t.Error("Expected validation warnings for conflicting routes")
	}

	t.Logf("Validation warnings: %v", warnings)
}

func TestGetRoutesByTenant(t *testing.T) {
	routes := []RouteConfig{
		{
			Description: "Tenant A Route",
			TenantID:    "tenant-a",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://localhost:8080"},
			Enabled:     true,
		},
		{
			Description: "Tenant B Route",
			TenantID:    "tenant-b",
			PathPrefix:  "/api",
			Backend:     BackendConfig{URL: "http://localhost:8081"},
			Enabled:     true,
		},
		{
			Description: "Global Route",
			PathPrefix:  "/global",
			Backend:     BackendConfig{URL: "http://localhost:8082"},
			Enabled:     true,
		},
	}

	matcher, err := NewRouteMatcher(routes)
	if err != nil {
		t.Fatalf("Failed to create route matcher: %v", err)
	}

	tenantARoutes := matcher.GetRoutesByTenant("tenant-a")
	if len(tenantARoutes) != 1 {
		t.Errorf("Expected 1 route for tenant-a, got %d", len(tenantARoutes))
	}

	if tenantARoutes[0].Description != "Tenant A Route" {
		t.Errorf("Expected Tenant A Route, got %s", tenantARoutes[0].Description)
	}
}

// Benchmark tests
func BenchmarkRouteMatching(b *testing.B) {
	routes := []RouteConfig{
		{PathPrefix: "/api/v1", Backend: BackendConfig{URL: "http://localhost:8080"}, Enabled: true},
		{PathPrefix: "/api/v2", Backend: BackendConfig{URL: "http://localhost:8081"}, Enabled: true},
		{PathPrefix: "/auth", Backend: BackendConfig{URL: "http://localhost:8082"}, Enabled: true},
		{Host: "*.api.example.com", PathPrefix: "/api", Backend: BackendConfig{URL: "http://localhost:8083"}, Enabled: true},
	}

	matcher, _ := NewRouteMatcher(routes)
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	req.Host = "tenant.api.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.Match(req)
	}
}

func BenchmarkTenantIDExtraction(b *testing.B) {
	matcher := &RouteMatcher{}
	req, _ := http.NewRequest("GET", "/tenant/acme-corp/api/users", nil)
	req.Host = "tenant1.api.example.com"
	req.Header.Set("X-Tenant-ID", "header-tenant")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.extractTenantID(req)
	}
}
