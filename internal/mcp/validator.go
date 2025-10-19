package mcp

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed schema/config-schema.json
var schemaJSON []byte

// Validator validates MCP configurations against the JSON schema.
type Validator struct {
	schema *jsonschema.Schema
}

// NewValidator creates a new validator with the embedded schema.
func NewValidator() *Validator {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	// Load embedded schema
	if err := compiler.AddResource("config-schema.json", bytes.NewReader(schemaJSON)); err != nil {
		// This should never fail with embedded schema
		panic(fmt.Sprintf("failed to add schema resource: %v", err))
	}

	schema, err := compiler.Compile("config-schema.json")
	if err != nil {
		panic(fmt.Sprintf("failed to compile schema: %v", err))
	}

	return &Validator{schema: schema}
}

// Validate validates a config file against the JSON schema.
// Returns detailed schema errors with file/line/property paths.
func (v *Validator) Validate(configPath string) []types.SchemaError {
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return []types.SchemaError{{
			FilePath: configPath,
			Line:     1,
			Column:   1,
			JSONPath: "/",
			Message:  fmt.Sprintf("failed to read file: %v", err),
		}}
	}

	// Parse JSON to get structure
	var configData interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return []types.SchemaError{{
			FilePath: configPath,
			Line:     1,
			Column:   1,
			JSONPath: "/",
			Message:  fmt.Sprintf("invalid JSON: %v", err),
		}}
	}

	// Validate against schema
	if err := v.schema.Validate(configData); err != nil {
		return v.convertValidationErrors(configPath, err)
	}

	return nil
}

// convertValidationErrors converts jsonschema validation errors to SchemaErrors.
func (v *Validator) convertValidationErrors(filePath string, err error) []types.SchemaError {
	var schemaErrors []types.SchemaError

	if validationErr, ok := err.(*jsonschema.ValidationError); ok {
		schemaErrors = append(schemaErrors, v.flattenValidationError(filePath, validationErr)...)
	} else {
		schemaErrors = append(schemaErrors, types.SchemaError{
			FilePath: filePath,
			Line:     1,
			Column:   1,
			JSONPath: "/",
			Message:  err.Error(),
		})
	}

	return schemaErrors
}

// flattenValidationError recursively flattens validation errors.
func (v *Validator) flattenValidationError(filePath string, err *jsonschema.ValidationError) []types.SchemaError {
	var errors []types.SchemaError

	// Add the current error
	errors = append(errors, types.SchemaError{
		FilePath: filePath,
		Line:     1, // TODO: Track line numbers from JSON parser
		Column:   1,
		JSONPath: err.InstanceLocation,
		Message:  err.Message,
	})

	// Add any sub-errors
	for _, cause := range err.Causes {
		errors = append(errors, v.flattenValidationError(filePath, cause)...)
	}

	return errors
}
