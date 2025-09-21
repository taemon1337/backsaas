package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"github.com/backsaas/platform/api/internal/types"
)

// HashPassword securely hashes a password using bcrypt
func HashPassword(ctx context.Context, execCtx *types.ExecutionContext, password string) (string, error) {
	// Use bcrypt with cost 12 for good security/performance balance
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		execCtx.Logger.Error("Failed to hash password", err, map[string]interface{}{
			"tenant_id": execCtx.TenantID,
		})
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	execCtx.Logger.Info("Password hashed successfully", map[string]interface{}{
		"tenant_id": execCtx.TenantID,
	})
	
	return string(hash), nil
}

// GenerateAPIKey generates a cryptographically secure API key
func GenerateAPIKey(ctx context.Context, execCtx *types.ExecutionContext, prefix string, length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		execCtx.Logger.Error("Failed to generate random bytes", err, map[string]interface{}{
			"tenant_id": execCtx.TenantID,
		})
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Encode to base64 and make URL-safe
	key := base64.URLEncoding.EncodeToString(bytes)
	
	// Add prefix
	if prefix == "" {
		prefix = "bks"
	}
	
	fullKey := fmt.Sprintf("%s_%s", prefix, key)
	
	execCtx.Logger.Info("API key generated", map[string]interface{}{
		"tenant_id": execCtx.TenantID,
		"prefix":    prefix,
		"length":    length,
	})
	
	return fullKey, nil
}

// GenerateSlug creates a URL-safe slug from text
func GenerateSlug(ctx context.Context, execCtx *types.ExecutionContext, text string, maxLength int, reservedWords []string, checkUniqueness bool) (string, error) {
	// Convert to lowercase
	slug := strings.ToLower(text)
	
	// Replace spaces and special characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Truncate if too long
	if maxLength > 0 && len(slug) > maxLength {
		slug = slug[:maxLength]
		slug = strings.TrimRight(slug, "-")
	}
	
	// Check reserved words
	for _, reserved := range reservedWords {
		if slug == reserved {
			return "", fmt.Errorf("slug '%s' is reserved", slug)
		}
	}
	
	// Check uniqueness if requested
	if checkUniqueness {
		existing, err := execCtx.DataService.FindMany(ctx, execCtx.Entity, map[string]interface{}{
			"slug": slug,
		})
		if err != nil {
			return "", fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		
		if len(existing) > 0 {
			return "", fmt.Errorf("slug '%s' already exists", slug)
		}
	}
	
	execCtx.Logger.Info("Slug generated", map[string]interface{}{
		"original_text": text,
		"slug":         slug,
		"tenant_id":    execCtx.TenantID,
	})
	
	return slug, nil
}
