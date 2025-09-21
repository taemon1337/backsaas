package functions

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	
	"github.com/backsaas/platform/api/internal/types"
)

// FunctionRegistry manages all available Go functions
type FunctionRegistry struct {
	functions map[string]*FunctionDefinition
	mu        sync.RWMutex
}

// FunctionDefinition describes a registered Go function
type FunctionDefinition struct {
	Name        string
	Package     string
	Function    string
	Description string
	Params      map[string]ParamDefinition
	Returns     ReturnDefinition
	Handler     reflect.Value
}

// ParamDefinition describes function parameters
type ParamDefinition struct {
	Type     string
	Required bool
	Default  interface{}
}

// ReturnDefinition describes function return type
type ReturnDefinition struct {
	Type string
	// Add a new field to ReturnDefinition
	Nullable bool
}


// DataService provides secure, tenant-scoped data operations
type DataService interface {
	FindByID(ctx context.Context, entity, id string) (map[string]interface{}, error)
	FindMany(ctx context.Context, entity string, filters map[string]interface{}) ([]map[string]interface{}, error)
	Create(ctx context.Context, entity string, data map[string]interface{}) (map[string]interface{}, error)
	Update(ctx context.Context, entity, id string, data map[string]interface{}) (map[string]interface{}, error)
	Delete(ctx context.Context, entity, id string) error
	Count(ctx context.Context, entity string, filters map[string]interface{}) (int64, error)
}

// EventService handles event publishing
type EventService interface {
	Publish(ctx context.Context, event string, data map[string]interface{}) error
	Schedule(ctx context.Context, event string, data map[string]interface{}, delay string) error
}

// Logger provides structured logging
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
}

// NewFunctionRegistry creates a new function registry
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: make(map[string]*FunctionDefinition),
	}
}

// Register adds a function to the registry
func (r *FunctionRegistry) Register(def *FunctionDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.functions[def.Name]; exists {
		return fmt.Errorf("function %s already registered", def.Name)
	}
	
	r.functions[def.Name] = def
	return nil
}

// Get retrieves a function definition
func (r *FunctionRegistry) Get(name string) (*FunctionDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	def, exists := r.functions[name]
	return def, exists
}

// List returns all registered functions
func (r *FunctionRegistry) List() map[string]*FunctionDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]*FunctionDefinition)
	for name, def := range r.functions {
		result[name] = def
	}
	return result
}

// Execute runs a function with the given parameters
func (r *FunctionRegistry) Execute(
	ctx context.Context,
	functionName string,
	params map[string]interface{},
	execCtx *types.ExecutionContext,
) (interface{}, error) {
	
	// Get function definition
	def, exists := r.Get(functionName)
	if !exists {
		return nil, fmt.Errorf("function %s not found", functionName)
	}
	
	// Validate parameters
	if err := r.validateParams(def, params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}
	
	// Prepare function arguments
	args, err := r.prepareArgs(def, params, execCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare arguments: %w", err)
	}
	
	// Execute function
	results := def.Handler.Call(args)
	
	// Handle results
	return r.handleResults(results)
}

// validateParams validates function parameters against definition
func (r *FunctionRegistry) validateParams(def *FunctionDefinition, params map[string]interface{}) error {
	// Check required parameters
	for paramName, paramDef := range def.Params {
		if paramDef.Required {
			if _, exists := params[paramName]; !exists {
				return fmt.Errorf("required parameter %s missing", paramName)
			}
		}
	}
	
	// TODO: Add type validation
	return nil
}

// prepareArgs converts parameters to function arguments
func (r *FunctionRegistry) prepareArgs(
	def *FunctionDefinition,
	params map[string]interface{},
	execCtx *types.ExecutionContext,
) ([]reflect.Value, error) {
	
	funcType := def.Handler.Type()
	numArgs := funcType.NumIn()
	args := make([]reflect.Value, numArgs)
	
	// First argument is always context
	args[0] = reflect.ValueOf(context.Background())
	
	// Second argument is execution context
	if numArgs > 1 {
		args[1] = reflect.ValueOf(execCtx)
	}
	
	// Remaining arguments are function parameters
	argIndex := 2
	for paramName, paramDef := range def.Params {
		if argIndex >= numArgs {
			break
		}
		
		value, exists := params[paramName]
		if !exists {
			// Use default value
			value = paramDef.Default
		}
		
		if value != nil {
			args[argIndex] = reflect.ValueOf(value)
		} else {
			// Use zero value for the type
			args[argIndex] = reflect.Zero(funcType.In(argIndex))
		}
		argIndex++
	}
	
	return args, nil
}

// handleResults processes function return values
func (r *FunctionRegistry) handleResults(results []reflect.Value) (interface{}, error) {
	if len(results) == 0 {
		return nil, nil
	}
	
	// Last result is always error (if present)
	if len(results) > 1 {
		if errVal := results[len(results)-1]; !errVal.IsNil() {
			if err, ok := errVal.Interface().(error); ok {
				return nil, err
			}
		}
	}
	
	// First result is the return value
	if results[0].IsValid() {
		return results[0].Interface(), nil
	}
	
	return nil, nil
}
