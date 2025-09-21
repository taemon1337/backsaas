package types

import (
	"context"
)

// ExecutionContext provides secure context for function execution
type ExecutionContext struct {
	TenantID     string
	UserID       string
	RequestID    string
	Entity       string
	Operation    string
	Logger       Logger
	DataService  DataService
	EventService EventService
	Data         map[string]interface{}
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
}

// EventService interface for publishing events
type EventService interface {
	Publish(event string, data map[string]interface{}) error
}

// DataService interface for data operations
type DataService interface {
	FindByID(ctx context.Context, entity, id string) (map[string]interface{}, error)
	FindMany(ctx context.Context, entity string, filters map[string]interface{}) ([]map[string]interface{}, error)
	Create(ctx context.Context, entity string, data map[string]interface{}) (map[string]interface{}, error)
	Update(ctx context.Context, entity, id string, data map[string]interface{}) (map[string]interface{}, error)
	Delete(ctx context.Context, entity, id string) error
	List(ctx context.Context, entity string, filters map[string]interface{}) ([]map[string]interface{}, error)
}
