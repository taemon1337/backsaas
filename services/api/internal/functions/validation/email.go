package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/backsaas/platform/api/internal/functions"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates email format and domain restrictions
func ValidateEmail(ctx context.Context, execCtx *functions.ExecutionContext, email string, allowedDomains []string) (bool, error) {
	// Basic format validation
	if !emailRegex.MatchString(email) {
		return false, fmt.Errorf("invalid email format")
	}
	
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))
	
	// Check domain restrictions if provided
	if len(allowedDomains) > 0 {
		domain := strings.Split(email, "@")[1]
		domainAllowed := false
		
		for _, allowedDomain := range allowedDomains {
			if domain == allowedDomain {
				domainAllowed = true
				break
			}
		}
		
		if !domainAllowed {
			return false, fmt.Errorf("email domain %s not allowed", domain)
		}
	}
	
	// Check for uniqueness in tenant data
	existing, err := execCtx.DataService.FindMany(ctx, "users", map[string]interface{}{
		"email": email,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	
	if len(existing) > 0 {
		return false, fmt.Errorf("email already exists")
	}
	
	execCtx.Logger.Info("Email validation passed", map[string]interface{}{
		"email":    email,
		"tenant_id": execCtx.TenantID,
	})
	
	return true, nil
}

// ValidatePassword validates password strength requirements
func ValidatePassword(ctx context.Context, execCtx *functions.ExecutionContext, password string, minLength int, requireUppercase, requireNumbers, requireSymbols bool) (bool, error) {
	if len(password) < minLength {
		return false, fmt.Errorf("password must be at least %d characters", minLength)
	}
	
	if requireUppercase {
		hasUpper := false
		for _, char := range password {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			return false, fmt.Errorf("password must contain at least one uppercase letter")
		}
	}
	
	if requireNumbers {
		hasNumber := false
		for _, char := range password {
			if char >= '0' && char <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			return false, fmt.Errorf("password must contain at least one number")
		}
	}
	
	if requireSymbols {
		hasSymbol := false
		symbols := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		for _, char := range password {
			if strings.ContainsRune(symbols, char) {
				hasSymbol = true
				break
			}
		}
		if !hasSymbol {
			return false, fmt.Errorf("password must contain at least one symbol")
		}
	}
	
	execCtx.Logger.Info("Password validation passed", map[string]interface{}{
		"tenant_id": execCtx.TenantID,
		"user_id":   execCtx.UserID,
	})
	
	return true, nil
}

// ValidatePhone validates phone number format
func ValidatePhone(ctx context.Context, execCtx *functions.ExecutionContext, phone, countryCode string) (bool, error) {
	// Remove all non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	
	// Basic length validation
	if len(digits) < 10 || len(digits) > 15 {
		return false, fmt.Errorf("phone number must be 10-15 digits")
	}
	
	// Country-specific validation could be added here
	if countryCode != "" {
		// TODO: Add country-specific phone validation
		execCtx.Logger.Info("Country-specific phone validation", map[string]interface{}{
			"country_code": countryCode,
			"phone":        phone,
		})
	}
	
	return true, nil
}
