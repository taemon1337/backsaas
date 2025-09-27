package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User represents a regular platform user (not admin)
type User struct {
	ID        string    `json:"id" db:"id"`
	FirstName string    `json:"firstName" db:"first_name"`
	LastName  string    `json:"lastName" db:"last_name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"` // Never include in JSON responses
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Tenant represents a tenant that a user can own or belong to
type Tenant struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	Template    string    `json:"template" db:"template"`
	OwnerID     string    `json:"ownerId" db:"owner_id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// UserClaims represents JWT claims for regular users
type UserClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CreateTenantRequest represents tenant creation data
type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	Template    string `json:"template" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token   string   `json:"token"`
	User    User     `json:"user"`
	Tenants []Tenant `json:"tenants"`
}

// UserAuthService handles user authentication and tenant management
type UserAuthService struct {
	jwtSecret []byte
	users     map[string]*User   // In-memory user store for demo
	tenants   map[string]*Tenant // In-memory tenant store for demo
	userTenants map[string][]string // User ID -> Tenant IDs
}

// NewUserAuthService creates a new user auth service
func NewUserAuthService() *UserAuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "backsaas-dev-secret-change-in-production"
	}

	// Create demo user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	demoUser := &User{
		ID:        "demo-user-1",
		FirstName: "Demo",
		LastName:  "User",
		Email:     "demo@backsaas.dev",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create demo tenant
	demoTenant := &Tenant{
		ID:          "demo-tenant-1",
		Name:        "Demo Company",
		Slug:        "demo-company",
		Description: "Demo tenant for testing",
		Template:    "crm",
		OwnerID:     "demo-user-1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return &UserAuthService{
		jwtSecret: []byte(jwtSecret),
		users: map[string]*User{
			"demo@backsaas.dev": demoUser,
		},
		tenants: map[string]*Tenant{
			"demo-tenant-1": demoTenant,
		},
		userTenants: map[string][]string{
			"demo-user-1": {"demo-tenant-1"},
		},
	}
}

// generateID creates a random ID
func (s *UserAuthService) generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Register handles user registration
func (s *UserAuthService) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	if _, exists := s.users[req.Email]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := &User{
		ID:        s.generateID(),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store user
	s.users[req.Email] = user
	s.userTenants[user.ID] = []string{} // Initialize empty tenant list

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, AuthResponse{
		Token:   token,
		User:    *user,
		Tenants: []Tenant{}, // New user has no tenants initially
	})
}

// Login handles user login
func (s *UserAuthService) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	user, exists := s.users[req.Email]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Get user's tenants
	tenants := s.getUserTenants(user.ID)

	// Return response
	c.JSON(http.StatusOK, AuthResponse{
		Token:   token,
		User:    *user,
		Tenants: tenants,
	})
}

// CreateTenant handles tenant creation
func (s *UserAuthService) CreateTenant(c *gin.Context) {
	// Get user from JWT token
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	currentUser := user.(*User)

	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if slug is available
	for _, tenant := range s.tenants {
		if tenant.Slug == req.Slug {
			c.JSON(http.StatusConflict, gin.H{"error": "Tenant slug is already taken"})
			return
		}
	}

	// Create tenant
	tenant := &Tenant{
		ID:          s.generateID(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Template:    req.Template,
		OwnerID:     currentUser.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store tenant
	s.tenants[tenant.ID] = tenant
	
	// Add tenant to user's tenant list
	if s.userTenants[currentUser.ID] == nil {
		s.userTenants[currentUser.ID] = []string{}
	}
	s.userTenants[currentUser.ID] = append(s.userTenants[currentUser.ID], tenant.ID)

	c.JSON(http.StatusCreated, tenant)
}

// CheckSlugAvailability checks if a tenant slug is available
func (s *UserAuthService) CheckSlugAvailability(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug parameter is required"})
		return
	}

	// Check if slug is taken
	available := true
	for _, tenant := range s.tenants {
		if tenant.Slug == slug {
			available = false
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{"available": available})
}

// generateToken creates a JWT token for the user
func (s *UserAuthService) generateToken(user *User) (string, error) {
	claims := UserClaims{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// getUserTenants gets all tenants for a user
func (s *UserAuthService) getUserTenants(userID string) []Tenant {
	tenantIDs := s.userTenants[userID]
	if tenantIDs == nil {
		return []Tenant{}
	}

	tenants := make([]Tenant, 0, len(tenantIDs))
	for _, tenantID := range tenantIDs {
		if tenant, exists := s.tenants[tenantID]; exists {
			tenants = append(tenants, *tenant)
		}
	}

	return tenants
}

// AuthMiddleware validates JWT tokens for user endpoints
func (s *UserAuthService) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Parse token
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Find user
		user, exists := s.users[claims.Email]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	}
}
