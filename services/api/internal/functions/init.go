package functions

import (
	"reflect"
	"time"

	"github.com/backsaas/platform/api/internal/functions/validation"
	"github.com/backsaas/platform/api/internal/functions/security"
	"github.com/backsaas/platform/api/internal/functions/communication"
)

// InitializeRegistry creates and populates the function registry with all available functions
func InitializeRegistry() *FunctionRegistry {
	registry := NewFunctionRegistry()
	
	// Register validation functions
	registerValidationFunctions(registry)
	
	// Register security functions
	registerSecurityFunctions(registry)
	
	// Register communication functions
	registerCommunicationFunctions(registry)
	
	// Register utility functions
	registerUtilityFunctions(registry)
	
	return registry
}

// registerValidationFunctions registers all validation functions
func registerValidationFunctions(registry *FunctionRegistry) {
	// validate_email
	registry.Register(&FunctionDefinition{
		Name:        "validate_email",
		Package:     "validation",
		Function:    "ValidateEmail",
		Description: "Validate email format and domain restrictions",
		Params: map[string]ParamDefinition{
			"email": {
				Type:     "string",
				Required: true,
			},
			"allowed_domains": {
				Type:     "[]string",
				Required: false,
				Default:  []string{},
			},
		},
		Returns: ReturnDefinition{
			Type: "bool",
		},
		Handler: reflect.ValueOf(validation.ValidateEmail),
	})
	
	// validate_password
	registry.Register(&FunctionDefinition{
		Name:        "validate_password",
		Package:     "validation",
		Function:    "ValidatePassword",
		Description: "Validate password strength requirements",
		Params: map[string]ParamDefinition{
			"password": {
				Type:     "string",
				Required: true,
			},
			"min_length": {
				Type:     "int",
				Required: false,
				Default:  8,
			},
			"require_uppercase": {
				Type:     "bool",
				Required: false,
				Default:  true,
			},
			"require_numbers": {
				Type:     "bool",
				Required: false,
				Default:  true,
			},
			"require_symbols": {
				Type:     "bool",
				Required: false,
				Default:  false,
			},
		},
		Returns: ReturnDefinition{
			Type: "bool",
		},
		Handler: reflect.ValueOf(validation.ValidatePassword),
	})
	
	// validate_phone
	registry.Register(&FunctionDefinition{
		Name:        "validate_phone",
		Package:     "validation",
		Function:    "ValidatePhone",
		Description: "Validate phone number format",
		Params: map[string]ParamDefinition{
			"phone": {
				Type:     "string",
				Required: true,
			},
			"country_code": {
				Type:     "string",
				Required: false,
				Default:  "",
			},
		},
		Returns: ReturnDefinition{
			Type: "bool",
		},
		Handler: reflect.ValueOf(validation.ValidatePhone),
	})
}

// registerSecurityFunctions registers all security functions
func registerSecurityFunctions(registry *FunctionRegistry) {
	// hash_password
	registry.Register(&FunctionDefinition{
		Name:        "hash_password",
		Package:     "security",
		Function:    "HashPassword",
		Description: "Hash password using bcrypt",
		Params: map[string]ParamDefinition{
			"password": {
				Type:     "string",
				Required: true,
			},
		},
		Returns: ReturnDefinition{
			Type: "string",
		},
		Handler: reflect.ValueOf(security.HashPassword),
	})
	
	// generate_api_key
	registry.Register(&FunctionDefinition{
		Name:        "generate_api_key",
		Package:     "security",
		Function:    "GenerateAPIKey",
		Description: "Generate cryptographically secure API key",
		Params: map[string]ParamDefinition{
			"prefix": {
				Type:     "string",
				Required: false,
				Default:  "bks",
			},
			"length": {
				Type:     "int",
				Required: false,
				Default:  32,
			},
		},
		Returns: ReturnDefinition{
			Type: "string",
		},
		Handler: reflect.ValueOf(security.GenerateAPIKey),
	})
	
	// generate_slug
	registry.Register(&FunctionDefinition{
		Name:        "generate_slug",
		Package:     "security",
		Function:    "GenerateSlug",
		Description: "Generate URL-safe slug from text",
		Params: map[string]ParamDefinition{
			"text": {
				Type:     "string",
				Required: true,
			},
			"max_length": {
				Type:     "int",
				Required: false,
				Default:  50,
			},
			"reserved_words": {
				Type:     "[]string",
				Required: false,
				Default:  []string{},
			},
			"check_uniqueness": {
				Type:     "bool",
				Required: false,
				Default:  false,
			},
		},
		Returns: ReturnDefinition{
			Type: "string",
		},
		Handler: reflect.ValueOf(security.GenerateSlug),
	})
}

// registerCommunicationFunctions registers all communication functions
func registerCommunicationFunctions(registry *FunctionRegistry) {
	// send_email
	registry.Register(&FunctionDefinition{
		Name:        "send_email",
		Package:     "communication",
		Function:    "SendEmail",
		Description: "Send templated email",
		Params: map[string]ParamDefinition{
			"template": {
				Type:     "string",
				Required: true,
			},
			"to": {
				Type:     "string",
				Required: true,
			},
			"data": {
				Type:     "map[string]interface{}",
				Required: true,
			},
		},
		Returns: ReturnDefinition{
			Type: "error",
		},
		Handler: reflect.ValueOf(communication.SendEmail),
	})
	
	// send_webhook
	registry.Register(&FunctionDefinition{
		Name:        "send_webhook",
		Package:     "communication",
		Function:    "SendWebhook",
		Description: "Send HTTP webhook",
		Params: map[string]ParamDefinition{
			"url": {
				Type:     "string",
				Required: true,
			},
			"payload": {
				Type:     "map[string]interface{}",
				Required: true,
			},
			"timeout": {
				Type:     "time.Duration",
				Required: false,
				Default:  30 * time.Second,
			},
		},
		Returns: ReturnDefinition{
			Type: "error",
		},
		Handler: reflect.ValueOf(communication.SendWebhook),
	})
}

// registerUtilityFunctions registers all utility functions
func registerUtilityFunctions(registry *FunctionRegistry) {
	// TODO: Add utility functions like format_currency, parse_date, calculate_age
}
