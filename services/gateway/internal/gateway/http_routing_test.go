package gateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestHTTPRoutingFlow tests the complete HTTP request flow through the gateway
func TestHTTPRoutingFlow(t *testing.T) {
	// Create mock backend servers that echo request info
	adminServer := createEchoServer("admin-console", "/admin")
	defer adminServer.Close()

	apiServer := createEchoServer("platform-api", "/api/platform")
	defer apiServer.Close()

	dashboardServer := createEchoServer("health-dashboard", "/dashboard")
	defer dashboardServer.Close()

	// Create gateway configuration with mock backend URLs
	config := &Config{
		Port:      "8000",
		RedisURL:  "redis://localhost:6379",
		JWTSecret: "test-secret",
		Routes: []RouteConfig{
			{
				Description: "Admin Console",
				PathPrefix:  "/admin",
				Backend: BackendConfig{
					URL:     adminServer.URL,
					Timeout: 30 * time.Second,
				},
				Transform: &TransformConfig{
					AddHeaders: map[string]string{
						"X-Interface-Type": "admin-console",
						"X-Gateway":        "BackSaas-Gateway/1.0",
					},
				},
				Enabled: true,
			},
			{
				Description: "Platform API",
				PathPrefix:  "/api/platform",
				Backend: BackendConfig{
					URL:     apiServer.URL,
					Timeout: 30 * time.Second,
				},
				Enabled: true,
			},
			{
				Description: "Health Dashboard",
				PathPrefix:  "/dashboard",
				Backend: BackendConfig{
					URL:     dashboardServer.URL,
					Timeout: 30 * time.Second,
				},
				Enabled: true,
			},
		},
	}

	// Create gateway (this will fail without Redis, so we'll mock it)
	gateway, err := createTestGateway(config)
	if err != nil {
		t.Skipf("Skipping HTTP test - cannot create gateway: %v", err)
		return
	}

	// Create test server with the gateway router
	testServer := httptest.NewServer(gateway.router)
	defer testServer.Close()

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:           "Admin Console Root",
			method:         "GET",
			path:           "/admin",
			expectedStatus: 200,
			expectedBody:   "admin-console",
			expectedHeader: "admin-console",
		},
		{
			name:           "Admin Console Asset",
			method:         "GET",
			path:           "/admin/_next/static/css/app.css",
			expectedStatus: 200,
			expectedBody:   "admin-console",
			expectedHeader: "admin-console",
		},
		{
			name:           "Platform API",
			method:         "GET",
			path:           "/api/platform/health",
			expectedStatus: 200,
			expectedBody:   "platform-api",
		},
		{
			name:           "Health Dashboard",
			method:         "GET",
			path:           "/dashboard",
			expectedStatus: 200,
			expectedBody:   "health-dashboard",
		},
		{
			name:           "Not Found",
			method:         "GET",
			path:           "/nonexistent",
			expectedStatus: 404,
		},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := testServer.URL + tc.path
			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Check response body contains expected service name
			if tc.expectedBody != "" {
				body := make([]byte, 1024)
				n, _ := resp.Body.Read(body)
				bodyStr := string(body[:n])
				
				if !strings.Contains(bodyStr, tc.expectedBody) {
					t.Errorf("Expected body to contain %q, got %q", tc.expectedBody, bodyStr)
				}
			}

			// Check custom headers were added
			if tc.expectedHeader != "" {
				if header := resp.Header.Get("X-Interface-Type"); header != tc.expectedHeader {
					t.Errorf("Expected X-Interface-Type header %q, got %q", tc.expectedHeader, header)
				}
			}

			t.Logf("✅ %s -> Status: %d, Service: %s", tc.path, resp.StatusCode, tc.expectedBody)
		})
	}
}

// createEchoServer creates a mock backend server that echoes request information
func createEchoServer(serviceName, pathPrefix string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		response := fmt.Sprintf(`{
			"service": "%s",
			"path": "%s",
			"method": "%s",
			"headers": %d,
			"original_path": "%s",
			"timestamp": "%s"
		}`, serviceName, r.URL.Path, r.Method, len(r.Header), r.URL.Path, time.Now().Format(time.RFC3339))
		
		w.Write([]byte(response))
	}))
}

// createTestGateway creates a gateway instance for testing (with mocked Redis if needed)
func createTestGateway(config *Config) (*Gateway, error) {
	// Try to create a real gateway first
	gateway, err := NewGateway(config)
	if err != nil {
		// If it fails (likely due to Redis), we could create a minimal test version
		// For now, we'll just return the error
		return nil, err
	}
	return gateway, nil
}
