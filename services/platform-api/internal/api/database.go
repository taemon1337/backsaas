package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/backsaas/platform/services/platform-api/internal/schema"
)

// DatabaseOperations handles all database-related operations
type DatabaseOperations struct {
	db       *sql.DB
	tenantID string
}

// NewDatabaseOperations creates a new database operations handler
func NewDatabaseOperations(db *sql.DB, tenantID string) *DatabaseOperations {
	return &DatabaseOperations{
		db:       db,
		tenantID: tenantID,
	}
}

// EnsureTablesExist creates tables for all entities in the schema if they don't exist
func (d *DatabaseOperations) EnsureTablesExist(schemaObj *schema.Schema) error {
	for entityName, entity := range schemaObj.Entities {
		if err := d.createTableIfNotExists(entityName, entity); err != nil {
			return fmt.Errorf("failed to create table for entity %s: %w", entityName, err)
		}
	}
	return nil
}

// createTableIfNotExists creates a table for an entity if it doesn't exist
func (d *DatabaseOperations) createTableIfNotExists(entityName string, entity *schema.Entity) error {
	// Build CREATE TABLE statement
	sqlQuery := d.buildCreateTableSQL(entityName, entity)
	
	// Execute the statement
	_, err := d.db.Exec(sqlQuery)
	if err != nil {
		return fmt.Errorf("failed to execute CREATE TABLE: %w", err)
	}
	
	return nil
}

// buildCreateTableSQL generates a CREATE TABLE SQL statement from entity schema
func (d *DatabaseOperations) buildCreateTableSQL(entityName string, entity *schema.Entity) string {
	var columns []string
	
	// Add primary key column (entity key)
	keyColumn := fmt.Sprintf("%s VARCHAR(255) PRIMARY KEY", entity.Key)
	columns = append(columns, keyColumn)
	
	// Add tenant_id column for multi-tenancy only if not already defined in schema
	if _, exists := entity.Schema.Properties["tenant_id"]; !exists {
		columns = append(columns, "tenant_id VARCHAR(255) NOT NULL")
	}
	
	// Add columns for each property in sorted order for consistency
	var propNames []string
	for propName := range entity.Schema.Properties {
		if propName != entity.Key {
			propNames = append(propNames, propName)
		}
	}
	
	// Sort property names to ensure consistent ordering
	for i := 0; i < len(propNames); i++ {
		for j := i + 1; j < len(propNames); j++ {
			if propNames[i] > propNames[j] {
				propNames[i], propNames[j] = propNames[j], propNames[i]
			}
		}
	}
	
	for _, propName := range propNames {
		propDef := entity.Schema.Properties[propName]
		columnDef := d.propertyToColumnDefinition(propName, propDef)
		columns = append(columns, columnDef)
	}
	
	// Add audit columns only if not already defined in schema
	if _, exists := entity.Schema.Properties["created_at"]; !exists {
		columns = append(columns, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
	}
	if _, exists := entity.Schema.Properties["updated_at"]; !exists {
		columns = append(columns, "updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
	}
	
	// Build the complete SQL
	sqlQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			%s
		)`,
		entityName,
		strings.Join(columns, ",\n\t\t\t"),
	)
	
	return sqlQuery
}

// propertyToColumnDefinition converts a schema property to a SQL column definition
func (d *DatabaseOperations) propertyToColumnDefinition(propName string, propDef *schema.PropertyDefinition) string {
	var sqlType string
	var constraints []string
	
	// Map JSON schema types to PostgreSQL types
	switch propDef.Type {
	case "string":
		if propDef.Format == "email" {
			sqlType = "VARCHAR(255)"
		} else if propDef.Format == "uri" {
			sqlType = "TEXT"
		} else if propDef.MaxLength > 0 {
			sqlType = fmt.Sprintf("VARCHAR(%d)", propDef.MaxLength)
		} else {
			sqlType = "TEXT"
		}
	case "integer":
		sqlType = "INTEGER"
	case "number":
		sqlType = "DECIMAL"
	case "boolean":
		sqlType = "BOOLEAN"
	case "array":
		sqlType = "JSONB"
	case "object":
		sqlType = "JSONB"
	default:
		sqlType = "TEXT"
	}
	
	// Add constraints
	if propDef.Default != nil {
		constraints = append(constraints, fmt.Sprintf("DEFAULT '%v'", propDef.Default))
	}
	
	// Build column definition
	columnDef := fmt.Sprintf("%s %s", propName, sqlType)
	if len(constraints) > 0 {
		columnDef += " " + strings.Join(constraints, " ")
	}
	
	return columnDef
}

// InsertEntity inserts a new entity into the database
func (d *DatabaseOperations) InsertEntity(entityName string, entity *schema.Entity, data map[string]interface{}) (map[string]interface{}, error) {
	// Create a copy of the data to avoid mutating the original map
	insertData := make(map[string]interface{})
	for k, v := range data {
		insertData[k] = v
	}
	
	// Add audit fields - truncate to microsecond precision to match PostgreSQL
	now := time.Now().Truncate(time.Microsecond)
	
	// Add audit fields only if not already provided and not defined in schema
	if _, exists := insertData["created_at"]; !exists && entity.Schema.Properties["created_at"] == nil {
		insertData["created_at"] = now
	}
	if _, exists := insertData["updated_at"]; !exists {
		insertData["updated_at"] = now
	}
	
	// Add tenant_id only if not already provided and not defined in schema
	if _, exists := insertData["tenant_id"]; !exists {
		insertData["tenant_id"] = d.tenantID
	} else {
		// If tenant_id is provided, ensure it matches the current tenant for security
		insertData["tenant_id"] = d.tenantID
	}
	
	// Generate ID if not provided
	if insertData[entity.Key] == nil {
		insertData[entity.Key] = d.generateID()
	}
	
	// Serialize complex types (arrays, objects) to JSON for PostgreSQL
	for key, value := range insertData {
		if propDef, exists := entity.Schema.Properties[key]; exists {
			if propDef.Type == "array" || propDef.Type == "object" {
				if value != nil {
					jsonBytes, err := json.Marshal(value)
					if err != nil {
						return nil, fmt.Errorf("failed to serialize %s: %w", key, err)
					}
					insertData[key] = string(jsonBytes)
				}
			}
		}
	}
	
	// Get expected column order to ensure consistent insertion
	expectedColumns := d.getEntityColumns(entity)
	columns := make([]string, 0, len(expectedColumns))
	placeholders := make([]string, 0, len(expectedColumns))
	values := make([]interface{}, 0, len(expectedColumns))
	
	i := 1
	for _, col := range expectedColumns {
		if value, exists := insertData[col]; exists {
			columns = append(columns, col)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			values = append(values, value)
			i++
		}
	}
	
	sqlQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		entityName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	
	// Execute the insert
	row := d.db.QueryRow(sqlQuery, values...)
	
	// Convert result back to map
	result, err := d.rowToMap(row, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to insert entity: %w", err)
	}
	
	return result, nil
}

// UpdateEntity updates an existing entity in the database
func (d *DatabaseOperations) UpdateEntity(entityName string, entity *schema.Entity, id string, data map[string]interface{}) (map[string]interface{}, error) {
	// Create a copy to avoid mutating the original
	updateData := make(map[string]interface{})
	for k, v := range data {
		updateData[k] = v
	}
	
	// Add audit fields - truncate to microsecond precision to match PostgreSQL
	// Only add updated_at if it's not already in the data
	if _, exists := updateData["updated_at"]; !exists {
		updateData["updated_at"] = time.Now().Truncate(time.Microsecond)
	}
	
	// Remove key and tenant_id from update data
	delete(updateData, entity.Key)
	delete(updateData, "tenant_id")
	delete(updateData, "created_at") // Don't allow updating created_at
	
	if len(updateData) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	
	// Serialize complex types (arrays, objects) to JSON for PostgreSQL
	for key, value := range updateData {
		if propDef, exists := entity.Schema.Properties[key]; exists {
			if propDef.Type == "array" || propDef.Type == "object" {
				if value != nil {
					jsonBytes, err := json.Marshal(value)
					if err != nil {
						return nil, fmt.Errorf("failed to serialize %s: %w", key, err)
					}
					updateData[key] = string(jsonBytes)
				}
			}
		}
	}
	
	// Build UPDATE statement
	setParts := make([]string, 0, len(updateData))
	values := make([]interface{}, 0, len(updateData)+2)
	
	i := 1
	for key, value := range updateData {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, i))
		values = append(values, value)
		i++
	}
	
	// Add WHERE clause parameters
	values = append(values, id, d.tenantID)
	
	sqlQuery := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d AND tenant_id = $%d RETURNING *",
		entityName,
		strings.Join(setParts, ", "),
		entity.Key,
		i,
		i+1,
	)
	
	// Execute the update
	row := d.db.QueryRow(sqlQuery, values...)
	
	// Convert result back to map
	result, err := d.rowToMap(row, entity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}
	
	return result, nil
}

// QueryEntities retrieves entities with optional filtering, pagination, and sorting
func (d *DatabaseOperations) QueryEntities(entityName string, entity *schema.Entity, filters map[string]interface{}, limit, offset int, orderBy string) ([]map[string]interface{}, error) {
	// Build base query
	query := fmt.Sprintf("SELECT * FROM %s WHERE tenant_id = $1", entityName)
	args := []interface{}{d.tenantID}
	argIndex := 2
	
	// Add filters
	for key, value := range filters {
		query += fmt.Sprintf(" AND %s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}
	
	// Add ordering
	if orderBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", orderBy)
	} else {
		query += " ORDER BY created_at DESC"
	}
	
	// Add pagination
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}
	
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}
	
	// Execute query
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query entities: %w", err)
	}
	defer rows.Close()
	
	// Convert rows to maps
	results, err := d.rowsToMaps(rows, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to process query results: %w", err)
	}
	
	return results, nil
}

// GetEntity retrieves a single entity by ID
func (d *DatabaseOperations) GetEntity(entityName string, entity *schema.Entity, id string) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 AND tenant_id = $2", entityName, entity.Key)
	
	row := d.db.QueryRow(query, id, d.tenantID)
	
	result, err := d.rowToMap(row, entity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}
	
	return result, nil
}

// DeleteEntity deletes an entity by ID
func (d *DatabaseOperations) DeleteEntity(entityName string, entity *schema.Entity, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1 AND tenant_id = $2", entityName, entity.Key)
	
	result, err := d.db.Exec(query, id, d.tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("entity not found")
	}
	
	return nil
}

// rowToMap converts a single database row to a map
func (d *DatabaseOperations) rowToMap(row *sql.Row, entity *schema.Entity) (map[string]interface{}, error) {
	// Get column names
	columns := d.getEntityColumns(entity)
	
	// Create slice of interface{} to hold values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	
	// Scan the row
	err := row.Scan(valuePtrs...)
	if err != nil {
		return nil, err
	}
	
	// Convert to map
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		
		// Convert []byte to string for text fields
		if b, ok := val.([]byte); ok {
			val = string(b)
		}
		
		// Normalize timestamp fields to match Go's timezone format
		if t, ok := val.(time.Time); ok && (col == "created_at" || col == "updated_at") {
			// Ensure timezone is in UTC and truncate to microsecond precision
			val = t.UTC().Truncate(time.Microsecond)
		}
		
		// Deserialize JSON fields back to Go types
		if propDef, exists := entity.Schema.Properties[col]; exists {
			if propDef.Type == "array" || propDef.Type == "object" {
				if strVal, ok := val.(string); ok && strVal != "" {
					var jsonVal interface{}
					if err := json.Unmarshal([]byte(strVal), &jsonVal); err == nil {
						val = jsonVal
					}
				}
			}
		}
		
		result[col] = val
	}
	
	return result, nil
}

// rowsToMaps converts multiple database rows to maps
func (d *DatabaseOperations) rowsToMaps(rows *sql.Rows, entity *schema.Entity) ([]map[string]interface{}, error) {
	// Get column names from the result set
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	var results []map[string]interface{}
	
	for rows.Next() {
		// Create slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		// Scan the row
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}
		
		// Convert to map
		result := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			
			// Convert []byte to string for text fields
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			
			// Normalize timestamp fields to match Go's timezone format
			if t, ok := val.(time.Time); ok && (col == "created_at" || col == "updated_at") {
				// Ensure timezone is in UTC and truncate to microsecond precision
				val = t.UTC().Truncate(time.Microsecond)
			}
			
			// Deserialize JSON fields back to Go types
			if propDef, exists := entity.Schema.Properties[col]; exists {
				if propDef.Type == "array" || propDef.Type == "object" {
					if strVal, ok := val.(string); ok && strVal != "" {
						var jsonVal interface{}
						if err := json.Unmarshal([]byte(strVal), &jsonVal); err == nil {
							val = jsonVal
						}
					}
				}
			}
			
			result[col] = val
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// getEntityColumns returns the expected column names for an entity in the same order as table creation
func (d *DatabaseOperations) getEntityColumns(entity *schema.Entity) []string {
	columns := []string{entity.Key}
	
	// Add tenant_id only if not already defined in schema properties
	if _, exists := entity.Schema.Properties["tenant_id"]; !exists {
		columns = append(columns, "tenant_id")
	}
	
	// Get property names in sorted order for consistency
	var propNames []string
	for propName := range entity.Schema.Properties {
		if propName != entity.Key {
			propNames = append(propNames, propName)
		}
	}
	
	// Sort property names to ensure consistent ordering
	// This matches the order used in buildCreateTableSQL
	for i := 0; i < len(propNames); i++ {
		for j := i + 1; j < len(propNames); j++ {
			if propNames[i] > propNames[j] {
				propNames[i], propNames[j] = propNames[j], propNames[i]
			}
		}
	}
	
	columns = append(columns, propNames...)
	
	// Add audit columns only if not already defined in schema
	if _, exists := entity.Schema.Properties["created_at"]; !exists {
		columns = append(columns, "created_at")
	}
	if _, exists := entity.Schema.Properties["updated_at"]; !exists {
		columns = append(columns, "updated_at")
	}
	
	return columns
}

// generateID generates a unique ID for an entity
func (d *DatabaseOperations) generateID() string {
	// Simple UUID-like ID generation
	// In production, you might want to use a proper UUID library
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// ValidateEntityData validates entity data against the schema
func (d *DatabaseOperations) ValidateEntityData(entity *schema.Entity, data map[string]interface{}) error {
	// Check required fields
	for _, requiredField := range entity.Schema.Required {
		if _, exists := data[requiredField]; !exists {
			return fmt.Errorf("required field '%s' is missing", requiredField)
		}
	}
	
	// Validate each property
	for propName, value := range data {
		// Skip system fields
		if propName == "tenant_id" || propName == "created_at" || propName == "updated_at" {
			continue
		}
		
		propDef, exists := entity.Schema.Properties[propName]
		if !exists {
			return fmt.Errorf("unknown property '%s'", propName)
		}
		
		if err := d.validateProperty(propName, value, propDef); err != nil {
			return err
		}
	}
	
	return nil
}

// validateProperty validates a single property value against its definition
func (d *DatabaseOperations) validateProperty(propName string, value interface{}, propDef *schema.PropertyDefinition) error {
	if value == nil {
		return nil // Allow null values for optional fields
	}
	
	// Type validation
	switch propDef.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("property '%s' must be a string", propName)
		}
		
		str := value.(string)
		
		// Length validation
		if propDef.MinLength > 0 && len(str) < propDef.MinLength {
			return fmt.Errorf("property '%s' must be at least %d characters", propName, propDef.MinLength)
		}
		if propDef.MaxLength > 0 && len(str) > propDef.MaxLength {
			return fmt.Errorf("property '%s' must be at most %d characters", propName, propDef.MaxLength)
		}
		
		// Enum validation
		if len(propDef.Enum) > 0 {
			valid := false
			for _, enumValue := range propDef.Enum {
				if str == enumValue {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("property '%s' must be one of: %v", propName, propDef.Enum)
			}
		}
		
	case "integer":
		// Check if it's a number that can be converted to int
		switch v := value.(type) {
		case int, int32, int64:
			// Already an integer
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("property '%s' must be an integer", propName)
			}
		default:
			return fmt.Errorf("property '%s' must be an integer", propName)
		}
		
	case "number":
		if !isNumeric(value) {
			return fmt.Errorf("property '%s' must be a number", propName)
		}
		
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("property '%s' must be a boolean", propName)
		}
		
	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			return fmt.Errorf("property '%s' must be an array", propName)
		}
		
	case "object":
		if reflect.TypeOf(value).Kind() != reflect.Map {
			return fmt.Errorf("property '%s' must be an object", propName)
		}
	}
	
	return nil
}

// isNumeric checks if a value is numeric
func isNumeric(value interface{}) bool {
	switch value.(type) {
	case int, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}
