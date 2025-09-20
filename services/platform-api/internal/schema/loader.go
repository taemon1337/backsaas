package schema

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Schema represents a complete BackSaas schema definition
type Schema struct {
	Version     int                    `yaml:"version"`
	Service     ServiceConfig          `yaml:"service"`
	Entities    map[string]*Entity     `yaml:"entities"`
	Functions   map[string]*Function   `yaml:"platform_functions,omitempty"`
	GoFunctions map[string]*GoFunction `yaml:"go_function_registry,omitempty"`
	AccessRules *AccessRules           `yaml:"access_rules,omitempty"`
	Events      map[string]*Event      `yaml:"events,omitempty"`
	Indexes     map[string][]Index     `yaml:"indexes,omitempty"`
}

// ServiceConfig defines service-level configuration
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// Entity represents a data entity definition
type Entity struct {
	Key    string                 `yaml:"key"`
	Schema EntitySchema           `yaml:"schema"`
	Access *EntityAccess          `yaml:"access,omitempty"`
}

// EntitySchema defines the JSON schema for an entity
type EntitySchema struct {
	Type       string                            `yaml:"type"`
	Required   []string                          `yaml:"required,omitempty"`
	Properties map[string]*PropertyDefinition   `yaml:"properties"`
}

// PropertyDefinition defines a single property
type PropertyDefinition struct {
	Type        string      `yaml:"type"`
	Format      string      `yaml:"format,omitempty"`
	Pattern     string      `yaml:"pattern,omitempty"`
	Enum        []string    `yaml:"enum,omitempty"`
	Items       interface{} `yaml:"items,omitempty"`
	MaxLength   int         `yaml:"maxLength,omitempty"`
	MinLength   int         `yaml:"minLength,omitempty"`
	Minimum     int         `yaml:"minimum,omitempty"`
	Maximum     int         `yaml:"maximum,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
}

// EntityAccess defines access control rules for an entity
type EntityAccess struct {
	Read   []AccessRule `yaml:"read,omitempty"`
	Write  []AccessRule `yaml:"write,omitempty"`
	Delete []AccessRule `yaml:"delete,omitempty"`
}

// AccessRule defines a single access control rule
type AccessRule struct {
	Role string `yaml:"role,omitempty"`
	Rule string `yaml:"rule,omitempty"`
}

// Function represents a platform function definition
type Function struct {
	Entity    string                 `yaml:"entity"`
	Type      string                 `yaml:"type"`
	Trigger   string                 `yaml:"trigger"`
	Field     string                 `yaml:"field,omitempty"`
	Function  string                 `yaml:"function,omitempty"`
	Config    map[string]interface{} `yaml:"config,omitempty"`
	Functions []FunctionCall         `yaml:"functions,omitempty"`
	Events    []EventCall            `yaml:"events,omitempty"`
	Async     bool                   `yaml:"async,omitempty"`
}

// GoFunction represents a Go function registry entry
type GoFunction struct {
	Package     string                        `yaml:"package"`
	Function    string                        `yaml:"function"`
	Description string                        `yaml:"description"`
	Params      map[string]*ParamDefinition   `yaml:"params"`
	Returns     *ReturnDefinition             `yaml:"returns"`
}

// ParamDefinition defines function parameters
type ParamDefinition struct {
	Type     string      `yaml:"type"`
	Required bool        `yaml:"required"`
	Default  interface{} `yaml:"default,omitempty"`
}

// ReturnDefinition defines function return type
type ReturnDefinition struct {
	Type string `yaml:"type"`
}

// FunctionCall represents a function call in a hook
type FunctionCall struct {
	Function string                 `yaml:"function"`
	Config   map[string]interface{} `yaml:"config"`
}

// EventCall represents an event to publish
type EventCall struct {
	Event string                 `yaml:"event"`
	Data  map[string]interface{} `yaml:"data"`
}

// AccessRules defines global access rules
type AccessRules struct {
	Rules map[string]string   `yaml:"rules"`
	Roles map[string]*Role    `yaml:"roles"`
}

// Role defines a role with inheritance
type Role struct {
	Description string   `yaml:"description"`
	Inherits    []string `yaml:"inherits,omitempty"`
}

// Event defines an event schema
type Event struct {
	Fields []string `yaml:"fields"`
}

// Index defines a database index
type Index struct {
	Fields []string `yaml:"fields"`
	Unique bool     `yaml:"unique,omitempty"`
}

// Loader handles schema loading from various sources
type Loader struct {
	basePath string
}

// NewLoader creates a new schema loader
func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath: basePath,
	}
}

// LoadFromFile loads a schema from a YAML file
func (l *Loader) LoadFromFile(filename string) (*Schema, error) {
	fullPath := filepath.Join(l.basePath, filename)
	
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", fullPath, err)
	}
	
	return l.LoadFromBytes(data)
}

// LoadFromBytes loads a schema from YAML bytes
func (l *Loader) LoadFromBytes(data []byte) (*Schema, error) {
	var schema Schema
	
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema YAML: %w", err)
	}
	
	// Validate schema
	if err := l.validateSchema(&schema); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}
	
	return &schema, nil
}

// LoadFromRegistry loads a schema from the schema registry (for tenant schemas)
func (l *Loader) LoadFromRegistry(tenantID string) (*Schema, error) {
	// TODO: Implement registry loading
	// This would query the platform database for the tenant's schema
	return nil, fmt.Errorf("registry loading not implemented yet")
}

// validateSchema performs basic schema validation
func (l *Loader) validateSchema(schema *Schema) error {
	if schema.Version == 0 {
		return fmt.Errorf("schema version is required")
	}
	
	if schema.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	
	if len(schema.Entities) == 0 {
		return fmt.Errorf("at least one entity is required")
	}
	
	// Validate entities
	for entityName, entity := range schema.Entities {
		if err := l.validateEntity(entityName, entity); err != nil {
			return fmt.Errorf("entity %s validation failed: %w", entityName, err)
		}
	}
	
	// Validate functions reference valid entities
	for functionName, function := range schema.Functions {
		if function.Entity != "" {
			if _, exists := schema.Entities[function.Entity]; !exists {
				return fmt.Errorf("function %s references unknown entity %s", functionName, function.Entity)
			}
		}
	}
	
	return nil
}

// validateEntity validates a single entity definition
func (l *Loader) validateEntity(name string, entity *Entity) error {
	if entity.Key == "" {
		return fmt.Errorf("entity key is required")
	}
	
	if entity.Schema.Type != "object" {
		return fmt.Errorf("entity schema type must be 'object'")
	}
	
	if len(entity.Schema.Properties) == 0 {
		return fmt.Errorf("entity must have at least one property")
	}
	
	// Validate key field exists in properties
	if _, exists := entity.Schema.Properties[entity.Key]; !exists {
		return fmt.Errorf("key field %s not found in properties", entity.Key)
	}
	
	// Validate required fields exist in properties
	for _, required := range entity.Schema.Required {
		if _, exists := entity.Schema.Properties[required]; !exists {
			return fmt.Errorf("required field %s not found in properties", required)
		}
	}
	
	return nil
}
