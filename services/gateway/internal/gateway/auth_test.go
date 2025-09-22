package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func TestNewAuthMiddleware(t *testing.T) {
	config := &AuthConfig{
		Required:   true, 
		JWTSecret:  "test",
		HeaderName: "Authorization",
	}
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	
	middleware, err := NewAuthMiddleware(config, redisClient)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if middleware == nil {
		t.Error("Expected middleware, got nil")
	}
}

func TestAuthHandler_NotRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &AuthConfig{Required: false}
	middleware := &AuthMiddleware{}
	
	router := gin.New()
	router.Use(middleware.Handler(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestAuthHandler_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &AuthConfig{
		Required:   true, 
		HeaderName: "Authorization",
		JWTSecret:  "test-secret",
	}
	middleware := &AuthMiddleware{}
	
	router := gin.New()
	router.Use(middleware.Handler(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code != 401 {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestShouldBypass(t *testing.T) {
	middleware := &AuthMiddleware{}
	
	tests := []struct {
		path     string
		bypaths  []string
		expected bool
	}{
		{"/health", []string{"/health"}, true},
		{"/api", []string{"/health"}, false},
		{"/health", []string{}, false},
	}
	
	for _, tt := range tests {
		result := middleware.shouldBypass(tt.path, tt.bypaths)
		if result != tt.expected {
			t.Errorf("Path %s: expected %v, got %v", tt.path, tt.expected, result)
		}
	}
}

// Helper function to create test JWT
func createTestJWT(secret, userID, tenantID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}
