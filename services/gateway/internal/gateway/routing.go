package gateway

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// RouteMatcher handles route matching logic
type RouteMatcher struct {
	routes []RouteConfig
	
	// Compiled regex patterns for performance
	pathPatterns map[int]*regexp.Regexp
}

// NewRouteMatcher creates a new route matcher
func NewRouteMatcher(routes []RouteConfig) (*RouteMatcher, error) {
	matcher := &RouteMatcher{
		routes:       routes,
		pathPatterns: make(map[int]*regexp.Regexp),
	}
	
	// Pre-compile regex patterns
	for i, route := range routes {
		if route.PathPrefix != "" && strings.Contains(route.PathPrefix, "*") {
			// Convert glob pattern to regex
			pattern := strings.ReplaceAll(route.PathPrefix, "*", ".*")
			regex, err := regexp.Compile("^" + pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid path pattern for route %d: %w", i, err)
			}
			matcher.pathPatterns[i] = regex
		}
	}
	
	return matcher, nil
}

// Match finds the best matching route for a request
func (rm *RouteMatcher) Match(req *http.Request) (*RouteConfig, error) {
	var bestMatch *RouteConfig
	var bestScore int
	
	for i, route := range rm.routes {
		if !route.Enabled {
			continue
		}
		
		score := rm.calculateMatchScore(req, &route, i)
		if score > bestScore {
			bestScore = score
			bestMatch = &route
		}
	}
	
	if bestMatch == nil {
		return nil, fmt.Errorf("no matching route found for %s %s", req.Method, req.URL.Path)
	}
	
	return bestMatch, nil
}

// calculateMatchScore calculates how well a route matches the request
func (rm *RouteMatcher) calculateMatchScore(req *http.Request, route *RouteConfig, routeIndex int) int {
	score := 0
	
	// Host matching (highest priority)
	if route.Host != "" {
		if rm.matchHost(req.Host, route.Host) {
			score += 1000
		} else {
			return 0 // Host mismatch is disqualifying
		}
	}
	
	// Tenant ID matching (high priority)
	if route.TenantID != "" {
		tenantID := rm.extractTenantID(req)
		if tenantID == route.TenantID {
			score += 500
		} else {
			return 0 // Tenant mismatch is disqualifying
		}
	}
	
	// Path prefix matching
	if route.PathPrefix != "" {
		if rm.matchPath(req.URL.Path, route.PathPrefix, routeIndex) {
			// Longer prefixes get higher scores
			score += len(route.PathPrefix) * 10
		} else {
			return 0 // Path mismatch is disqualifying
		}
	}
	
	// Header matching
	headerMatches := 0
	for headerName, expectedValue := range route.Headers {
		actualValue := req.Header.Get(headerName)
		if rm.matchHeaderValue(actualValue, expectedValue) {
			headerMatches++
			score += 50
		} else {
			return 0 // Header mismatch is disqualifying
		}
	}
	
	// Bonus for exact matches
	if route.PathPrefix != "" && req.URL.Path == route.PathPrefix {
		score += 100
	}
	
	return score
}

// matchHost checks if request host matches route host pattern
func (rm *RouteMatcher) matchHost(requestHost, routeHost string) bool {
	// Remove port from request host for comparison
	if colonIndex := strings.Index(requestHost, ":"); colonIndex != -1 {
		requestHost = requestHost[:colonIndex]
	}
	
	// Exact match
	if requestHost == routeHost {
		return true
	}
	
	// Wildcard subdomain matching (*.example.com)
	if strings.HasPrefix(routeHost, "*.") {
		domain := routeHost[2:]
		return strings.HasSuffix(requestHost, "."+domain) || requestHost == domain
	}
	
	return false
}

// matchPath checks if request path matches route path pattern
func (rm *RouteMatcher) matchPath(requestPath, routePath string, routeIndex int) bool {
	// Exact prefix match
	if strings.HasPrefix(requestPath, routePath) {
		return true
	}
	
	// Regex pattern match (if compiled)
	if pattern, exists := rm.pathPatterns[routeIndex]; exists {
		return pattern.MatchString(requestPath)
	}
	
	return false
}

// matchHeaderValue checks if header value matches expected pattern
func (rm *RouteMatcher) matchHeaderValue(actualValue, expectedValue string) bool {
	// Exact match
	if actualValue == expectedValue {
		return true
	}
	
	// Wildcard matching
	if strings.Contains(expectedValue, "*") {
		pattern := strings.ReplaceAll(expectedValue, "*", ".*")
		if regex, err := regexp.Compile("^" + pattern + "$"); err == nil {
			return regex.MatchString(actualValue)
		}
	}
	
	return false
}

// extractTenantID extracts tenant ID from various sources
func (rm *RouteMatcher) extractTenantID(req *http.Request) string {
	// Try header first
	if tenantID := req.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	
	// Try subdomain extraction (tenant.api.example.com)
	host := req.Host
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}
	
	parts := strings.Split(host, ".")
	if len(parts) >= 3 && parts[1] == "api" {
		return parts[0]
	}
	
	// Try path extraction (/tenant/api/...)
	pathParts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(pathParts) >= 2 && pathParts[1] == "api" {
		return pathParts[0]
	}
	
	// Try query parameter
	if tenantID := req.URL.Query().Get("tenant_id"); tenantID != "" {
		return tenantID
	}
	
	return ""
}

// GetRoutesByTenant returns all routes for a specific tenant
func (rm *RouteMatcher) GetRoutesByTenant(tenantID string) []RouteConfig {
	var routes []RouteConfig
	
	for _, route := range rm.routes {
		if route.TenantID == tenantID && route.Enabled {
			routes = append(routes, route)
		}
	}
	
	return routes
}

// GetRoutesByHost returns all routes for a specific host
func (rm *RouteMatcher) GetRoutesByHost(host string) []RouteConfig {
	var routes []RouteConfig
	
	for _, route := range rm.routes {
		if rm.matchHost(host, route.Host) && route.Enabled {
			routes = append(routes, route)
		}
	}
	
	return routes
}

// ValidateRoutes validates route configuration for conflicts
func (rm *RouteMatcher) ValidateRoutes() []string {
	var warnings []string
	
	// Check for duplicate routes
	for i, route1 := range rm.routes {
		for j, route2 := range rm.routes {
			if i >= j {
				continue
			}
			
			if rm.routesConflict(&route1, &route2) {
				warnings = append(warnings, fmt.Sprintf(
					"Routes %d and %d may conflict: both match similar patterns",
					i, j,
				))
			}
		}
	}
	
	// Check for unreachable routes
	for i, route := range rm.routes {
		if rm.isRouteUnreachable(&route, i) {
			warnings = append(warnings, fmt.Sprintf(
				"Route %d may be unreachable due to more specific routes defined earlier",
				i,
			))
		}
	}
	
	return warnings
}

// routesConflict checks if two routes might conflict
func (rm *RouteMatcher) routesConflict(route1, route2 *RouteConfig) bool {
	// Same host and overlapping paths
	if route1.Host == route2.Host {
		if strings.HasPrefix(route1.PathPrefix, route2.PathPrefix) ||
			strings.HasPrefix(route2.PathPrefix, route1.PathPrefix) {
			return true
		}
	}
	
	// Same tenant ID
	if route1.TenantID != "" && route1.TenantID == route2.TenantID {
		return true
	}
	
	return false
}

// isRouteUnreachable checks if a route is unreachable
func (rm *RouteMatcher) isRouteUnreachable(route *RouteConfig, routeIndex int) bool {
	// Check if any earlier route would always match first
	for i := 0; i < routeIndex; i++ {
		earlierRoute := &rm.routes[i]
		if !earlierRoute.Enabled {
			continue
		}
		
		// If earlier route is more general and would match everything this route matches
		if rm.routeSubsumes(earlierRoute, route) {
			return true
		}
	}
	
	return false
}

// routeSubsumes checks if route1 would match everything that route2 matches
func (rm *RouteMatcher) routeSubsumes(route1, route2 *RouteConfig) bool {
	// If route1 has no host restriction but route2 does, route1 is more general
	if route1.Host == "" && route2.Host != "" {
		// Check if path is also more general
		if route1.PathPrefix == "" || 
			(route2.PathPrefix != "" && strings.HasPrefix(route2.PathPrefix, route1.PathPrefix)) {
			return true
		}
	}
	
	return false
}
