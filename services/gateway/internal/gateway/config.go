package gateway

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the gateway configuration
type Config struct {
	Port        string `yaml:"port"`
	RedisURL    string `yaml:"redis_url"`
	JWTSecret   string `yaml:"jwt_secret"`
	ConfigPath  string `yaml:"-"`
	Environment string `yaml:"environment"`

	// Loaded from config file
	Routes      []RouteConfig      `yaml:"routes"`
	RateLimit   RateLimitConfig    `yaml:"rate_limit"`
	Auth        AuthConfig         `yaml:"auth"`
	Monitoring  MonitoringConfig   `yaml:"monitoring"`
	Cors        CorsConfig         `yaml:"cors"`
}

// RouteConfig defines how to route requests to backend services
type RouteConfig struct {
	// Matching criteria
	Host       string            `yaml:"host"`        // Host header matching
	PathPrefix string            `yaml:"path_prefix"` // Path prefix matching
	TenantID   string            `yaml:"tenant_id"`   // Tenant identifier
	Headers    map[string]string `yaml:"headers"`     // Header matching

	// Backend configuration
	Backend     BackendConfig `yaml:"backend"`
	Auth        *AuthConfig   `yaml:"auth,omitempty"`        // Override auth for this route
	RateLimit   *RateLimitConfig `yaml:"rate_limit,omitempty"` // Override rate limit
	Transform   *TransformConfig `yaml:"transform,omitempty"`  // Request/response transformation
	Enabled     bool          `yaml:"enabled"`
	Description string        `yaml:"description"`
}

// BackendConfig defines the backend service
type BackendConfig struct {
	URL             string        `yaml:"url"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	HealthCheckPath string        `yaml:"health_check_path"`
	LoadBalancing   string        `yaml:"load_balancing"` // round_robin, least_connections, etc.
	
	// Multiple backend URLs for load balancing
	URLs []string `yaml:"urls,omitempty"`
}

// RateLimitConfig defines rate limiting rules
type RateLimitConfig struct {
	Enabled     bool          `yaml:"enabled"`
	RequestsPerMinute int     `yaml:"requests_per_minute"`
	BurstSize   int           `yaml:"burst_size"`
	KeyStrategy string        `yaml:"key_strategy"` // ip, user, tenant, custom
	CustomKey   string        `yaml:"custom_key,omitempty"`
	
	// Different limits for different user types
	Limits map[string]RateLimit `yaml:"limits,omitempty"`
}

type RateLimit struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	BurstSize         int `yaml:"burst_size"`
}

// AuthConfig defines authentication and authorization
type AuthConfig struct {
	Enabled      bool     `yaml:"enabled"`
	Required     bool     `yaml:"required"`
	JWTSecret    string   `yaml:"jwt_secret,omitempty"`
	
	// Token sources
	HeaderName   string   `yaml:"header_name"`   // Default: "Authorization"
	CookieName   string   `yaml:"cookie_name,omitempty"`
	QueryParam   string   `yaml:"query_param,omitempty"`
	
	// Authorization
	RequiredRoles []string `yaml:"required_roles,omitempty"`
	RequiredScopes []string `yaml:"required_scopes,omitempty"`
	
	// Bypass patterns
	BypassPaths  []string `yaml:"bypass_paths,omitempty"`
}

// TransformConfig defines request/response transformations
type TransformConfig struct {
	// Request transformations
	AddHeaders     map[string]string `yaml:"add_headers,omitempty"`
	RemoveHeaders  []string          `yaml:"remove_headers,omitempty"`
	RewritePath    string            `yaml:"rewrite_path,omitempty"`
	
	// Response transformations
	AddResponseHeaders    map[string]string `yaml:"add_response_headers,omitempty"`
	RemoveResponseHeaders []string          `yaml:"remove_response_headers,omitempty"`
}

// MonitoringConfig defines monitoring and observability
type MonitoringConfig struct {
	Enabled     bool   `yaml:"enabled"`
	MetricsPath string `yaml:"metrics_path"`
	HealthPath  string `yaml:"health_path"`
	
	// Logging
	LogLevel    string `yaml:"log_level"`
	LogFormat   string `yaml:"log_format"` // json, text
	LogRequests bool   `yaml:"log_requests"`
	
	// Tracing
	TracingEnabled bool   `yaml:"tracing_enabled"`
	TracingService string `yaml:"tracing_service"`
}

// CorsConfig defines CORS settings
type CorsConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// LoadConfig loads configuration from file and merges with runtime config
func LoadConfig(configPath string, runtimeConfig *Config) (*Config, error) {
	// Start with runtime config
	config := *runtimeConfig
	
	// Load from file if it exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
			
			var fileConfig Config
			if err := yaml.Unmarshal(data, &fileConfig); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
			
			// Merge file config with runtime config
			mergeConfigs(&config, &fileConfig)
		}
	}
	
	// Set defaults
	setDefaults(&config)
	
	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// mergeConfigs merges file config into runtime config
func mergeConfigs(runtime, file *Config) {
	if file.Port != "" {
		runtime.Port = file.Port
	}
	if file.RedisURL != "" {
		runtime.RedisURL = file.RedisURL
	}
	if file.Environment != "" {
		runtime.Environment = file.Environment
	}
	
	// Merge complex structures
	runtime.Routes = file.Routes
	runtime.RateLimit = file.RateLimit
	runtime.Auth = file.Auth
	runtime.Monitoring = file.Monitoring
	runtime.Cors = file.Cors
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	// Default auth config
	if config.Auth.HeaderName == "" {
		config.Auth.HeaderName = "Authorization"
	}
	
	// Default monitoring config
	if config.Monitoring.MetricsPath == "" {
		config.Monitoring.MetricsPath = "/metrics"
	}
	if config.Monitoring.HealthPath == "" {
		config.Monitoring.HealthPath = "/health"
	}
	if config.Monitoring.LogLevel == "" {
		config.Monitoring.LogLevel = "info"
	}
	if config.Monitoring.LogFormat == "" {
		config.Monitoring.LogFormat = "json"
	}
	
	// Default CORS config
	if config.Cors.Enabled && len(config.Cors.AllowedMethods) == 0 {
		config.Cors.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if config.Cors.Enabled && len(config.Cors.AllowedHeaders) == 0 {
		config.Cors.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Tenant-ID"}
	}
	
	// Default backend timeouts
	for i := range config.Routes {
		if config.Routes[i].Backend.Timeout == 0 {
			config.Routes[i].Backend.Timeout = 30 * time.Second
		}
		if config.Routes[i].Backend.MaxRetries == 0 {
			config.Routes[i].Backend.MaxRetries = 3
		}
		if config.Routes[i].Backend.HealthCheckPath == "" {
			config.Routes[i].Backend.HealthCheckPath = "/health"
		}
	}
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.Port == "" {
		return fmt.Errorf("port is required")
	}
	
	if config.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	
	// Validate routes
	for i, route := range config.Routes {
		if route.Backend.URL == "" && len(route.Backend.URLs) == 0 {
			return fmt.Errorf("route %d: backend URL is required", i)
		}
		
		if route.PathPrefix == "" && route.Host == "" && route.TenantID == "" {
			return fmt.Errorf("route %d: at least one matching criteria is required", i)
		}
	}
	
	return nil
}
