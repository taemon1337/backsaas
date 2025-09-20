package communication

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"bytes"
	"encoding/json"

	"github.com/backsaas/platform/api/internal/functions"
)

// EmailService interface for sending emails
type EmailService interface {
	SendTemplatedEmail(template, to string, data map[string]interface{}) error
}

// WebhookClient interface for HTTP requests
type WebhookClient interface {
	Post(url string, payload interface{}, timeout time.Duration) error
}

// SendEmail sends a templated email
func SendEmail(ctx context.Context, execCtx *functions.ExecutionContext, template, to string, data map[string]interface{}) error {
	// Validate email address
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}
	
	// Validate template
	if template == "" {
		return fmt.Errorf("email template is required")
	}
	
	// Add tenant context to email data
	emailData := make(map[string]interface{})
	for k, v := range data {
		emailData[k] = v
	}
	emailData["tenant_id"] = execCtx.TenantID
	
	// TODO: Integrate with actual email service (SendGrid, AWS SES, etc.)
	execCtx.Logger.Info("Email sent", map[string]interface{}{
		"template":  template,
		"to":        to,
		"tenant_id": execCtx.TenantID,
		"data":      emailData,
	})
	
	// Publish email sent event
	err := execCtx.EventService.Publish(ctx, "email.sent", map[string]interface{}{
		"template":  template,
		"to":        to,
		"tenant_id": execCtx.TenantID,
		"sent_at":   time.Now().UTC(),
	})
	if err != nil {
		execCtx.Logger.Error("Failed to publish email.sent event", err, map[string]interface{}{
			"template": template,
			"to":       to,
		})
	}
	
	return nil
}

// SendWebhook sends an HTTP webhook
func SendWebhook(ctx context.Context, execCtx *functions.ExecutionContext, url string, payload map[string]interface{}, timeout time.Duration) error {
	// Validate URL
	if url == "" {
		return fmt.Errorf("webhook URL is required")
	}
	
	// Add tenant context to payload
	webhookPayload := make(map[string]interface{})
	for k, v := range payload {
		webhookPayload[k] = v
	}
	webhookPayload["tenant_id"] = execCtx.TenantID
	webhookPayload["timestamp"] = time.Now().UTC()
	
	// Set default timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}
	
	// Marshal payload to JSON
	jsonPayload, err := json.Marshal(webhookPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "BackSaas-Webhook/1.0")
	
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		execCtx.Logger.Error("Webhook request failed", err, map[string]interface{}{
			"url":       url,
			"tenant_id": execCtx.TenantID,
		})
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode >= 400 {
		execCtx.Logger.Error("Webhook returned error status", nil, map[string]interface{}{
			"url":         url,
			"status_code": resp.StatusCode,
			"tenant_id":   execCtx.TenantID,
		})
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	
	execCtx.Logger.Info("Webhook sent successfully", map[string]interface{}{
		"url":         url,
		"status_code": resp.StatusCode,
		"tenant_id":   execCtx.TenantID,
	})
	
	// Publish webhook sent event
	err = execCtx.EventService.Publish(ctx, "webhook.sent", map[string]interface{}{
		"url":         url,
		"status_code": resp.StatusCode,
		"tenant_id":   execCtx.TenantID,
		"sent_at":     time.Now().UTC(),
	})
	if err != nil {
		execCtx.Logger.Error("Failed to publish webhook.sent event", err, map[string]interface{}{
			"url": url,
		})
	}
	
	return nil
}
