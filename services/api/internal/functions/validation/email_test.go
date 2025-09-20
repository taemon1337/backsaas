package validation

import (
	"context"
	"testing"

	"github.com/backsaas/platform/api/internal/functions"
)

// MockDataService for testing
type MockDataService struct {
	existingEmails []string
}

func (m *MockDataService) FindByID(ctx context.Context, entity, id string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *MockDataService) FindMany(ctx context.Context, entity string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	if entity == "users" {
		if email, ok := filters["email"].(string); ok {
			for _, existing := range m.existingEmails {
				if existing == email {
					return []map[string]interface{}{
						{"id": "existing-user", "email": email},
					}, nil
				}
			}
		}
	}
	return []map[string]interface{}{}, nil
}

func (m *MockDataService) Create(ctx context.Context, entity string, data map[string]interface{}) (map[string]interface{}, error) {
	return data, nil
}

func (m *MockDataService) Update(ctx context.Context, entity, id string, data map[string]interface{}) (map[string]interface{}, error) {
	return data, nil
}

func (m *MockDataService) Delete(ctx context.Context, entity, id string) error {
	return nil
}

func (m *MockDataService) Count(ctx context.Context, entity string, filters map[string]interface{}) (int64, error) {
	return 0, nil
}

// MockEventService for testing
type MockEventService struct {
	publishedEvents []string
}

func (m *MockEventService) Publish(ctx context.Context, event string, data map[string]interface{}) error {
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *MockEventService) Schedule(ctx context.Context, event string, data map[string]interface{}, delay string) error {
	return nil
}

// MockLogger for testing
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Info(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "INFO: "+msg)
}

func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "WARN: "+msg)
}

func (m *MockLogger) Error(msg string, err error, fields map[string]interface{}) {
	m.logs = append(m.logs, "ERROR: "+msg)
}

func createTestContext() *functions.ExecutionContext {
	return &functions.ExecutionContext{
		TenantID:     "test-tenant",
		UserID:       "test-user",
		RequestID:    "test-request",
		Entity:       "users",
		Operation:    "create",
		DataService:  &MockDataService{},
		EventService: &MockEventService{},
		Logger:       &MockLogger{},
	}
}

func TestValidateEmail(t *testing.T) {
	ctx := context.Background()
	execCtx := createTestContext()

	t.Run("ValidEmail", func(t *testing.T) {
		valid, err := ValidateEmail(ctx, execCtx, "test@example.com", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !valid {
			t.Error("Expected email to be valid")
		}
	})

	t.Run("InvalidEmailFormat", func(t *testing.T) {
		valid, err := ValidateEmail(ctx, execCtx, "invalid-email", []string{})
		if err == nil {
			t.Error("Expected error for invalid email format")
		}
		if valid {
			t.Error("Expected email to be invalid")
		}
	})

	t.Run("EmailWithoutAtSymbol", func(t *testing.T) {
		valid, err := ValidateEmail(ctx, execCtx, "invalidemail.com", []string{})
		if err == nil {
			t.Error("Expected error for email without @ symbol")
		}
		if valid {
			t.Error("Expected email to be invalid")
		}
	})

	t.Run("EmailWithoutDomain", func(t *testing.T) {
		valid, err := ValidateEmail(ctx, execCtx, "test@", []string{})
		if err == nil {
			t.Error("Expected error for email without domain")
		}
		if valid {
			t.Error("Expected email to be invalid")
		}
	})

	t.Run("AllowedDomains", func(t *testing.T) {
		allowedDomains := []string{"company.com", "partner.com"}
		
		// Test allowed domain
		valid, err := ValidateEmail(ctx, execCtx, "test@company.com", allowedDomains)
		if err != nil {
			t.Errorf("Expected no error for allowed domain, got: %v", err)
		}
		if !valid {
			t.Error("Expected email with allowed domain to be valid")
		}
		
		// Test disallowed domain
		valid, err = ValidateEmail(ctx, execCtx, "test@other.com", allowedDomains)
		if err == nil {
			t.Error("Expected error for disallowed domain")
		}
		if valid {
			t.Error("Expected email with disallowed domain to be invalid")
		}
	})

	t.Run("EmailUniqueness", func(t *testing.T) {
		// Mock existing email
		mockData := &MockDataService{
			existingEmails: []string{"existing@example.com"},
		}
		execCtx.DataService = mockData
		
		// Test existing email
		valid, err := ValidateEmail(ctx, execCtx, "existing@example.com", []string{})
		if err == nil {
			t.Error("Expected error for existing email")
		}
		if valid {
			t.Error("Expected existing email to be invalid")
		}
		
		// Test new email
		valid, err = ValidateEmail(ctx, execCtx, "new@example.com", []string{})
		if err != nil {
			t.Errorf("Expected no error for new email, got: %v", err)
		}
		if !valid {
			t.Error("Expected new email to be valid")
		}
	})

	t.Run("EmailNormalization", func(t *testing.T) {
		// Email should be normalized to lowercase
		valid, err := ValidateEmail(ctx, execCtx, "Test@Example.COM", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !valid {
			t.Error("Expected normalized email to be valid")
		}
	})
}

func TestValidatePassword(t *testing.T) {
	ctx := context.Background()
	execCtx := createTestContext()

	t.Run("ValidPassword", func(t *testing.T) {
		valid, err := ValidatePassword(ctx, execCtx, "Password123!", 8, true, true, true)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !valid {
			t.Error("Expected password to be valid")
		}
	})

	t.Run("TooShortPassword", func(t *testing.T) {
		valid, err := ValidatePassword(ctx, execCtx, "Pass1!", 8, true, true, true)
		if err == nil {
			t.Error("Expected error for too short password")
		}
		if valid {
			t.Error("Expected short password to be invalid")
		}
	})

	t.Run("NoUppercasePassword", func(t *testing.T) {
		valid, err := ValidatePassword(ctx, execCtx, "password123!", 8, true, true, true)
		if err == nil {
			t.Error("Expected error for password without uppercase")
		}
		if valid {
			t.Error("Expected password without uppercase to be invalid")
		}
	})

	t.Run("NoNumberPassword", func(t *testing.T) {
		valid, err := ValidatePassword(ctx, execCtx, "Password!", 8, true, true, true)
		if err == nil {
			t.Error("Expected error for password without number")
		}
		if valid {
			t.Error("Expected password without number to be invalid")
		}
	})

	t.Run("NoSymbolPassword", func(t *testing.T) {
		valid, err := ValidatePassword(ctx, execCtx, "Password123", 8, true, true, true)
		if err == nil {
			t.Error("Expected error for password without symbol")
		}
		if valid {
			t.Error("Expected password without symbol to be invalid")
		}
	})

	t.Run("FlexibleRequirements", func(t *testing.T) {
		// Test with relaxed requirements
		valid, err := ValidatePassword(ctx, execCtx, "password", 6, false, false, false)
		if err != nil {
			t.Errorf("Expected no error with relaxed requirements, got: %v", err)
		}
		if !valid {
			t.Error("Expected password to be valid with relaxed requirements")
		}
	})
}

func TestValidatePhone(t *testing.T) {
	ctx := context.Background()
	execCtx := createTestContext()

	t.Run("ValidPhone", func(t *testing.T) {
		valid, err := ValidatePhone(ctx, execCtx, "+1-555-123-4567", "US")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !valid {
			t.Error("Expected phone to be valid")
		}
	})

	t.Run("ValidPhoneDigitsOnly", func(t *testing.T) {
		valid, err := ValidatePhone(ctx, execCtx, "15551234567", "")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !valid {
			t.Error("Expected phone to be valid")
		}
	})

	t.Run("TooShortPhone", func(t *testing.T) {
		valid, err := ValidatePhone(ctx, execCtx, "123456789", "")
		if err == nil {
			t.Error("Expected error for too short phone")
		}
		if valid {
			t.Error("Expected short phone to be invalid")
		}
	})

	t.Run("TooLongPhone", func(t *testing.T) {
		valid, err := ValidatePhone(ctx, execCtx, "1234567890123456", "")
		if err == nil {
			t.Error("Expected error for too long phone")
		}
		if valid {
			t.Error("Expected long phone to be invalid")
		}
	})

	t.Run("PhoneWithFormatting", func(t *testing.T) {
		valid, err := ValidatePhone(ctx, execCtx, "(555) 123-4567", "")
		if err != nil {
			t.Errorf("Expected no error for formatted phone, got: %v", err)
		}
		if !valid {
			t.Error("Expected formatted phone to be valid")
		}
	})
}
