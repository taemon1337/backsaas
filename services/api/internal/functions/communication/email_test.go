package communication

import (
	"context"
	"testing"
	"time"

	"github.com/backsaas/platform/api/internal/types"
)

// Mock logger for testing
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Data    map[string]interface{}
}

func (m *MockLogger) Info(message string, data map[string]interface{}) {
	m.logs = append(m.logs, LogEntry{
		Level:   "info",
		Message: message,
		Data:    data,
	})
}

func (m *MockLogger) Error(message string, err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err != nil {
		data["error"] = err.Error()
	}
	m.logs = append(m.logs, LogEntry{
		Level:   "error",
		Message: message,
		Data:    data,
	})
}

func (m *MockLogger) Warn(message string, data map[string]interface{}) {
	m.logs = append(m.logs, LogEntry{
		Level:   "warn",
		Message: message,
		Data:    data,
	})
}

func (m *MockLogger) Debug(message string, data map[string]interface{}) {
	m.logs = append(m.logs, LogEntry{
		Level:   "debug",
		Message: message,
		Data:    data,
	})
}

func TestSendEmail(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		to          string
		data        map[string]interface{}
		expectError bool
		description string
	}{
		{
			name:     "ValidEmail",
			template: "welcome",
			to:       "test@example.com",
			data: map[string]interface{}{
				"user_name": "Test User",
			},
			expectError: false,
			description: "Should send email successfully with valid parameters",
		},
		{
			name:        "EmptyTemplate",
			template:    "",
			to:          "test@example.com",
			data:        map[string]interface{}{},
			expectError: true,
			description: "Should fail when template is empty",
		},
		{
			name:        "EmptyRecipient",
			template:    "welcome",
			to:          "",
			data:        map[string]interface{}{},
			expectError: true,
			description: "Should fail when recipient email is empty",
		},
		{
			name:     "ValidEmailWithData",
			template: "password_reset",
			to:       "user@example.com",
			data: map[string]interface{}{
				"user_name":  "John Doe",
				"reset_link": "https://example.com/reset/token123",
			},
			expectError: false,
			description: "Should send email with template data successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock logger and event service
			mockLogger := &MockLogger{}
			mockEventService := &MockEventService{}
			
			// Create execution context
			execCtx := &types.ExecutionContext{
				TenantID:     "test-tenant-123",
				Logger:       mockLogger,
				EventService: mockEventService,
			}

			// Execute SendEmail
			err := SendEmail(context.Background(), execCtx, tt.template, tt.to, tt.data)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}

			// For successful cases, verify logging
			if !tt.expectError {
				if len(mockLogger.logs) == 0 {
					t.Errorf("%s: expected log entry but got none", tt.description)
				} else {
					log := mockLogger.logs[0]
					if log.Level != "info" {
						t.Errorf("%s: expected info log level, got %s", tt.description, log.Level)
					}
					if log.Message != "Email sent" {
						t.Errorf("%s: expected 'Email sent' message, got %s", tt.description, log.Message)
					}
					
					// Verify log data contains expected fields
					if log.Data["template"] != tt.template {
						t.Errorf("%s: expected template %s in log, got %v", tt.description, tt.template, log.Data["template"])
					}
					if log.Data["to"] != tt.to {
						t.Errorf("%s: expected recipient %s in log, got %v", tt.description, tt.to, log.Data["to"])
					}
					if log.Data["tenant_id"] != "test-tenant-123" {
						t.Errorf("%s: expected tenant_id in log", tt.description)
					}
				}
			}
		})
	}
}

func TestSendWebhook(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		payload     map[string]interface{}
		timeout     time.Duration
		expectError bool
		description string
	}{
		{
			name: "ValidWebhook",
			url:  "https://example.com/webhook",
			payload: map[string]interface{}{
				"event": "user.created",
				"data":  "test data",
			},
			timeout:     30 * time.Second,
			expectError: false,
			description: "Should send webhook successfully with valid parameters",
		},
		{
			name:        "EmptyURL",
			url:         "",
			payload:     map[string]interface{}{},
			timeout:     30 * time.Second,
			expectError: true,
			description: "Should fail when webhook URL is empty",
		},
		{
			name: "ValidWebhookWithDefaultTimeout",
			url:  "https://example.com/webhook",
			payload: map[string]interface{}{
				"event": "user.updated",
			},
			timeout:     0, // Should use default timeout
			expectError: false,
			description: "Should use default timeout when timeout is 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock logger and event service
			mockLogger := &MockLogger{}
			mockEventService := &MockEventService{}
			
			// Create execution context
			execCtx := &types.ExecutionContext{
				TenantID:     "test-tenant-456",
				Logger:       mockLogger,
				EventService: mockEventService,
			}

			// Execute SendWebhook
			err := SendWebhook(context.Background(), execCtx, tt.url, tt.payload, tt.timeout)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.description, err)
			}

			// Note: Since SendWebhook creates its own HTTP client and makes real requests,
			// we can't easily test the actual HTTP call without more complex mocking.
			// In a production environment, we'd inject the HTTP client as a dependency.
		})
	}
}

func TestEmailServiceInterface(t *testing.T) {
	t.Run("EmailServiceInterfaceExists", func(t *testing.T) {
		// This test verifies that the EmailService interface exists
		// and can be implemented by a mock
		var _ EmailService = &MockEmailService{}
	})
}

func TestWebhookClientInterface(t *testing.T) {
	t.Run("WebhookClientInterfaceExists", func(t *testing.T) {
		// This test verifies that the WebhookClient interface exists
		// and can be implemented by a mock
		var _ WebhookClient = &MockWebhookClient{}
	})
}

// Mock implementations to verify interfaces
type MockEmailService struct {
	sentEmails []MockEmail
}

type MockEmail struct {
	Template string
	To       string
	Data     map[string]interface{}
}

func (m *MockEmailService) SendTemplatedEmail(template, to string, data map[string]interface{}) error {
	m.sentEmails = append(m.sentEmails, MockEmail{
		Template: template,
		To:       to,
		Data:     data,
	})
	return nil
}

type MockWebhookClient struct {
	requests []MockWebhookRequest
}

type MockWebhookRequest struct {
	URL     string
	Payload interface{}
	Timeout time.Duration
}

func (m *MockWebhookClient) Post(url string, payload interface{}, timeout time.Duration) error {
	m.requests = append(m.requests, MockWebhookRequest{
		URL:     url,
		Payload: payload,
		Timeout: timeout,
	})
	return nil
}

type MockEventService struct {
	events []MockEvent
}

type MockEvent struct {
	Event string
	Data  map[string]interface{}
}

func (m *MockEventService) Publish(event string, data map[string]interface{}) error {
	m.events = append(m.events, MockEvent{
		Event: event,
		Data:  data,
	})
	return nil
}

// Benchmark tests
func BenchmarkSendEmail(b *testing.B) {
	mockLogger := &MockLogger{}
	execCtx := &types.ExecutionContext{
		TenantID: "benchmark-tenant",
		Logger:   mockLogger,
	}

	template := "welcome"
	to := "benchmark@example.com"
	data := map[string]interface{}{
		"user_name": "Benchmark User",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SendEmail(context.Background(), execCtx, template, to, data)
	}
}

func BenchmarkSendWebhook(b *testing.B) {
	mockLogger := &MockLogger{}
	execCtx := &types.ExecutionContext{
		TenantID: "benchmark-tenant",
		Logger:   mockLogger,
	}

	url := "https://example.com/webhook"
	payload := map[string]interface{}{
		"event": "benchmark.test",
		"data":  "benchmark data",
	}
	timeout := 30 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This will make actual HTTP requests in benchmarks
		// In production, we'd want to mock the HTTP client
		SendWebhook(context.Background(), execCtx, url, payload, timeout)
	}
}
