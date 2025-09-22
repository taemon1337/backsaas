package security

import (
	"context"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/backsaas/platform/api/internal/types"
)

// MockDataService for testing slug uniqueness
type MockDataService struct {
	existingSlugs []string
}

func (m *MockDataService) FindByID(ctx context.Context, entity, id string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *MockDataService) FindMany(ctx context.Context, entity string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	if slug, ok := filters["slug"].(string); ok {
		for _, existing := range m.existingSlugs {
			if existing == slug {
				return []map[string]interface{}{
					{"id": "existing-item", "slug": slug},
				}, nil
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

func (m *MockDataService) List(ctx context.Context, entity string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

// MockEventService for testing
type MockEventService struct{}

func (m *MockEventService) Publish(event string, data map[string]interface{}) error {
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

func createTestContext() *types.ExecutionContext {
	return &types.ExecutionContext{
		TenantID:     "test-tenant",
		UserID:       "test-user",
		RequestID:    "test-request",
		Entity:       "tenants",
		Operation:    "create",
		DataService:  &MockDataService{},
		EventService: &MockEventService{},
		Logger:       &MockLogger{},
	}
}

func TestHashPassword(t *testing.T) {
	ctx := context.Background()
	execCtx := createTestContext()

	t.Run("ValidPasswordHashing", func(t *testing.T) {
		password := "TestPassword123!"
		hash, err := HashPassword(ctx, execCtx, password)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if hash == "" {
			t.Error("Expected non-empty hash")
		}
		
		if hash == password {
			t.Error("Hash should not equal original password")
		}
		
		// Verify the hash can be used to validate the password
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err != nil {
			t.Errorf("Hash verification failed: %v", err)
		}
	})

	t.Run("DifferentPasswordsDifferentHashes", func(t *testing.T) {
		password1 := "Password1"
		password2 := "Password2"
		
		hash1, err1 := HashPassword(ctx, execCtx, password1)
		hash2, err2 := HashPassword(ctx, execCtx, password2)
		
		if err1 != nil || err2 != nil {
			t.Errorf("Expected no errors, got: %v, %v", err1, err2)
		}
		
		if hash1 == hash2 {
			t.Error("Different passwords should produce different hashes")
		}
	})

	t.Run("SamePasswordDifferentHashes", func(t *testing.T) {
		password := "SamePassword123!"
		
		hash1, err1 := HashPassword(ctx, execCtx, password)
		hash2, err2 := HashPassword(ctx, execCtx, password)
		
		if err1 != nil || err2 != nil {
			t.Errorf("Expected no errors, got: %v, %v", err1, err2)
		}
		
		// bcrypt should produce different hashes for the same password (due to salt)
		if hash1 == hash2 {
			t.Error("Same password should produce different hashes due to salt")
		}
		
		// But both should validate the original password
		err1 = bcrypt.CompareHashAndPassword([]byte(hash1), []byte(password))
		err2 = bcrypt.CompareHashAndPassword([]byte(hash2), []byte(password))
		
		if err1 != nil || err2 != nil {
			t.Error("Both hashes should validate the original password")
		}
	})
}

func TestGenerateAPIKey(t *testing.T) {
	ctx := context.Background()
	execCtx := createTestContext()

	t.Run("DefaultAPIKey", func(t *testing.T) {
		key, err := GenerateAPIKey(ctx, execCtx, "bks", 32)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if !strings.HasPrefix(key, "bks_") {
			t.Errorf("Expected key to start with 'bks_', got: %s", key)
		}
		
		if len(key) < 10 {
			t.Errorf("Expected key to be reasonably long, got length: %d", len(key))
		}
	})

	t.Run("CustomPrefix", func(t *testing.T) {
		key, err := GenerateAPIKey(ctx, execCtx, "custom", 16)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if !strings.HasPrefix(key, "custom_") {
			t.Errorf("Expected key to start with 'custom_', got: %s", key)
		}
	})

	t.Run("UniqueKeys", func(t *testing.T) {
		key1, err1 := GenerateAPIKey(ctx, execCtx, "test", 32)
		key2, err2 := GenerateAPIKey(ctx, execCtx, "test", 32)
		
		if err1 != nil || err2 != nil {
			t.Errorf("Expected no errors, got: %v, %v", err1, err2)
		}
		
		if key1 == key2 {
			t.Error("Generated keys should be unique")
		}
	})

	t.Run("EmptyPrefix", func(t *testing.T) {
		key, err := GenerateAPIKey(ctx, execCtx, "", 32)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if !strings.HasPrefix(key, "bks_") {
			t.Errorf("Expected default prefix 'bks_' for empty prefix, got: %s", key)
		}
	})
}

func TestGenerateSlug(t *testing.T) {
	ctx := context.Background()
	execCtx := createTestContext()

	t.Run("BasicSlugGeneration", func(t *testing.T) {
		slug, err := GenerateSlug(ctx, execCtx, "Hello World", 50, []string{}, false)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		expected := "hello-world"
		if slug != expected {
			t.Errorf("Expected slug '%s', got '%s'", expected, slug)
		}
	})

	t.Run("ComplexTextSlug", func(t *testing.T) {
		slug, err := GenerateSlug(ctx, execCtx, "My Awesome Product! (2023)", 50, []string{}, false)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		expected := "my-awesome-product-2023"
		if slug != expected {
			t.Errorf("Expected slug '%s', got '%s'", expected, slug)
		}
	})

	t.Run("SlugTruncation", func(t *testing.T) {
		longText := "This is a very long text that should be truncated to fit within the maximum length limit"
		slug, err := GenerateSlug(ctx, execCtx, longText, 20, []string{}, false)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if len(slug) > 20 {
			t.Errorf("Expected slug length <= 20, got length: %d, slug: %s", len(slug), slug)
		}
		
		// Should not end with hyphen after truncation
		if strings.HasSuffix(slug, "-") {
			t.Errorf("Slug should not end with hyphen: %s", slug)
		}
	})

	t.Run("ReservedWords", func(t *testing.T) {
		reservedWords := []string{"admin", "api", "www"}
		
		_, err := GenerateSlug(ctx, execCtx, "admin", 50, reservedWords, false)
		if err == nil {
			t.Error("Expected error for reserved word 'admin'")
		}
		
		_, err = GenerateSlug(ctx, execCtx, "API", 50, reservedWords, false)
		if err == nil {
			t.Error("Expected error for reserved word 'api' (case insensitive)")
		}
		
		// Non-reserved word should work
		slug, err := GenerateSlug(ctx, execCtx, "dashboard", 50, reservedWords, false)
		if err != nil {
			t.Errorf("Expected no error for non-reserved word, got: %v", err)
		}
		if slug != "dashboard" {
			t.Errorf("Expected slug 'dashboard', got '%s'", slug)
		}
	})

	t.Run("UniquenessCheck", func(t *testing.T) {
		// Mock existing slugs
		mockData := &MockDataService{
			existingSlugs: []string{"existing-slug"},
		}
		execCtx.DataService = mockData
		
		// Test existing slug
		_, err := GenerateSlug(ctx, execCtx, "Existing Slug", 50, []string{}, true)
		if err == nil {
			t.Error("Expected error for existing slug")
		}
		
		// Test new slug
		slug, err := GenerateSlug(ctx, execCtx, "New Slug", 50, []string{}, true)
		if err != nil {
			t.Errorf("Expected no error for new slug, got: %v", err)
		}
		if slug != "new-slug" {
			t.Errorf("Expected slug 'new-slug', got '%s'", slug)
		}
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"hello@world.com", "hello-world-com"},
			{"user#123", "user-123"},
			{"test_underscore", "test-underscore"},
			{"multiple   spaces", "multiple-spaces"},
			{"--leading-trailing--", "leading-trailing"},
			{"CamelCaseText", "camelcasetext"},
		}
		
		for _, tc := range testCases {
			slug, err := GenerateSlug(ctx, execCtx, tc.input, 50, []string{}, false)
			if err != nil {
				t.Errorf("Expected no error for input '%s', got: %v", tc.input, err)
			}
			if slug != tc.expected {
				t.Errorf("For input '%s', expected slug '%s', got '%s'", tc.input, tc.expected, slug)
			}
		}
	})

	t.Run("EmptyInput", func(t *testing.T) {
		slug, err := GenerateSlug(ctx, execCtx, "", 50, []string{}, false)
		if err != nil {
			t.Errorf("Expected no error for empty input, got: %v", err)
		}
		if slug != "" {
			t.Errorf("Expected empty slug for empty input, got '%s'", slug)
		}
	})

	t.Run("OnlySpecialCharacters", func(t *testing.T) {
		slug, err := GenerateSlug(ctx, execCtx, "!@#$%^&*()", 50, []string{}, false)
		if err != nil {
			t.Errorf("Expected no error for special characters only, got: %v", err)
		}
		if slug != "" {
			t.Errorf("Expected empty slug for special characters only, got '%s'", slug)
		}
	})
}
