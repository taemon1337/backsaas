package gateway

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	redisClient *redis.Client
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(config *AuthConfig, redisClient *redis.Client) (*AuthMiddleware, error) {
	return &AuthMiddleware{
		redisClient: redisClient,
	}, nil
}

// Handler returns the auth middleware handler
func (a *AuthMiddleware) Handler(config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if not required
		if !config.Required {
			c.Next()
			return
		}
		
		// Check if path is in bypass list
		if a.shouldBypass(c.Request.URL.Path, config.BypassPaths) {
			c.Next()
			return
		}
		
		// Extract token
		token, err := a.extractToken(c, config)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"message": err.Error(),
			})
			c.Abort()
			return
		}
		
		// Validate token
		claims, err := a.validateToken(token, config.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}
		
		// Check authorization (roles, scopes)
		if err := a.checkAuthorization(claims, config); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"message": err.Error(),
			})
			c.Abort()
			return
		}
		
		// Store user info in context
		a.setUserContext(c, claims)
		
		c.Next()
	}
}

// extractToken extracts JWT token from request
func (a *AuthMiddleware) extractToken(c *gin.Context, config *AuthConfig) (string, error) {
	// Try header first
	if config.HeaderName != "" {
		authHeader := c.GetHeader(config.HeaderName)
		if authHeader != "" {
			// Handle "Bearer <token>" format
			if strings.HasPrefix(authHeader, "Bearer ") {
				return strings.TrimPrefix(authHeader, "Bearer "), nil
			}
			return authHeader, nil
		}
	}
	
	// Try cookie
	if config.CookieName != "" {
		if cookie, err := c.Cookie(config.CookieName); err == nil {
			return cookie, nil
		}
	}
	
	// Try query parameter
	if config.QueryParam != "" {
		if token := c.Query(config.QueryParam); token != "" {
			return token, nil
		}
	}
	
	return "", fmt.Errorf("no authentication token found")
}

// validateToken validates JWT token and returns claims
func (a *AuthMiddleware) validateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	
	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}
	
	return claims, nil
}

// checkAuthorization checks if user has required roles/scopes
func (a *AuthMiddleware) checkAuthorization(claims jwt.MapClaims, config *AuthConfig) error {
	// Check required roles
	if len(config.RequiredRoles) > 0 {
		userRoles, ok := claims["roles"].([]interface{})
		if !ok {
			return fmt.Errorf("no roles found in token")
		}
		
		hasRequiredRole := false
		for _, requiredRole := range config.RequiredRoles {
			for _, userRole := range userRoles {
				if userRole.(string) == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}
		
		if !hasRequiredRole {
			return fmt.Errorf("insufficient role permissions")
		}
	}
	
	// Check required scopes
	if len(config.RequiredScopes) > 0 {
		userScopes, ok := claims["scopes"].([]interface{})
		if !ok {
			return fmt.Errorf("no scopes found in token")
		}
		
		hasRequiredScope := false
		for _, requiredScope := range config.RequiredScopes {
			for _, userScope := range userScopes {
				if userScope.(string) == requiredScope {
					hasRequiredScope = true
					break
				}
			}
			if hasRequiredScope {
				break
			}
		}
		
		if !hasRequiredScope {
			return fmt.Errorf("insufficient scope permissions")
		}
	}
	
	return nil
}

// shouldBypass checks if path should bypass authentication
func (a *AuthMiddleware) shouldBypass(path string, bypassPaths []string) bool {
	for _, bypassPath := range bypassPaths {
		if strings.HasPrefix(path, bypassPath) {
			return true
		}
	}
	return false
}

// setUserContext stores user information in request context
func (a *AuthMiddleware) setUserContext(c *gin.Context, claims jwt.MapClaims) {
	// Extract common user fields
	if userID, ok := claims["sub"].(string); ok {
		c.Set("user_id", userID)
	}
	
	if email, ok := claims["email"].(string); ok {
		c.Set("user_email", email)
	}
	
	if tenantID, ok := claims["tenant_id"].(string); ok {
		c.Set("tenant_id", tenantID)
	}
	
	if roles, ok := claims["roles"].([]interface{}); ok {
		roleStrings := make([]string, len(roles))
		for i, role := range roles {
			roleStrings[i] = role.(string)
		}
		c.Set("user_roles", roleStrings)
	}
	
	if scopes, ok := claims["scopes"].([]interface{}); ok {
		scopeStrings := make([]string, len(scopes))
		for i, scope := range scopes {
			scopeStrings[i] = scope.(string)
		}
		c.Set("user_scopes", scopeStrings)
	}
	
	// Store full claims for advanced use cases
	c.Set("jwt_claims", claims)
}

// GenerateToken generates a JWT token (utility function for testing)
func GenerateToken(userID, email, tenantID string, roles, scopes []string, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":       userID,
		"email":     email,
		"tenant_id": tenantID,
		"roles":     roles,
		"scopes":    scopes,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(expiry).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
