// Package json provides JSON schema validation utilities.
// This is the canonical location for JSON schema validation, migrated
// from domain.JSONSchemaValidator.
package json

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/ready-to-release/eac/go/core/validation"
	"github.com/xeipuuv/gojsonschema"
)

// arrayPathRegex matches array field paths like "module_analyses.0.field".
var arrayPathRegex = regexp.MustCompile(`^([a-z_]+)\.(\d+)`)

// Validator validates JSON against JSON Schema.
type Validator struct {
	schemaPath string
	schema     *gojsonschema.Schema
}

// NewValidator creates a new JSON schema validator.
// Loads and compiles the schema at creation time so Validate calls are fast.
func NewValidator(schemaPath string) (*Validator, error) {
	fileURL := pathToFileURL(schemaPath)
	schemaLoader := gojsonschema.NewReferenceLoader(fileURL)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return &Validator{
		schemaPath: schemaPath,
		schema:     schema,
	}, nil
}

// Validate validates JSON output against the compiled schema.
func (v *Validator) Validate(output string, _ map[string]interface{}) []validation.ValidationError {
	// Parse JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		return []validation.ValidationError{
			*validation.NewValidationError(validation.ErrInvalidJSON, fmt.Sprintf("Invalid JSON: %v", err), 0),
		}
	}

	// Validate against schema
	documentLoader := gojsonschema.NewGoLoader(jsonData)
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		return []validation.ValidationError{
			*validation.NewValidationError(validation.ErrJSONSchemaViolation, fmt.Sprintf("Schema validation failed: %v", err), 0),
		}
	}

	// Convert schema errors to ValidationErrors with enhanced context
	if !result.Valid() {
		return processSchemaErrors(result.Errors(), jsonData)
	}

	return nil
}

// VerifyImplementation checks if the validator is properly configured.
func (v *Validator) VerifyImplementation() []validation.ValidationError {
	if v.schema == nil {
		return []validation.ValidationError{
			*validation.NewValidationError(validation.ErrJSONSchemaViolation, "schema not loaded", 0),
		}
	}
	return nil
}

// --- internal helpers -------------------------------------------------------

// arrayErrorPattern represents a detected pattern in array validation errors.
type arrayErrorPattern struct {
	arrayPath     string
	affectedItems []int
	errorsByField map[string]int
	sampleIndex   int
	actualFields  map[string]bool
}

// processSchemaErrors converts schema errors to contextual validation errors.
func processSchemaErrors(schemaErrors []gojsonschema.ResultError, jsonData interface{}) []validation.ValidationError {
	patterns := detectArrayPatterns(schemaErrors, jsonData)

	if len(patterns) > 0 {
		return createPatternBasedErrors(patterns, schemaErrors, jsonData)
	}

	var errors []validation.ValidationError
	for _, schemaErr := range schemaErrors {
		errors = append(errors, *enhanceError(schemaErr))
	}
	return errors
}

// detectArrayPatterns finds repetitive error patterns in arrays.
func detectArrayPatterns(schemaErrors []gojsonschema.ResultError, jsonData interface{}) map[string]*arrayErrorPattern {
	patterns := make(map[string]*arrayErrorPattern)

	for _, schemaErr := range schemaErrors {
		field := schemaErr.Field()
		arrayPath, itemIndex, ok := parseArrayPath(field)
		if !ok {
			continue
		}

		if patterns[arrayPath] == nil {
			patterns[arrayPath] = &arrayErrorPattern{
				arrayPath:     arrayPath,
				affectedItems: []int{},
				errorsByField: make(map[string]int),
				sampleIndex:   itemIndex,
				actualFields:  make(map[string]bool),
			}
		}

		pattern := patterns[arrayPath]
		if !slices.Contains(pattern.affectedItems, itemIndex) {
			pattern.affectedItems = append(pattern.affectedItems, itemIndex)
		}
		pattern.errorsByField[schemaErr.Description()]++

		if itemIndex == pattern.sampleIndex {
			extractActualFields(jsonData, arrayPath, itemIndex, pattern)
		}
	}

	// Only return patterns that affect multiple items (threshold: 3)
	filtered := make(map[string]*arrayErrorPattern)
	for path, pattern := range patterns {
		if len(pattern.affectedItems) >= 3 {
			filtered[path] = pattern
		}
	}
	return filtered
}

// parseArrayPath extracts array path and index from field like "module_analyses.0.field".
func parseArrayPath(field string) (string, int, bool) {
	matches := arrayPathRegex.FindStringSubmatch(field)
	if len(matches) < 3 {
		return "", 0, false
	}
	var idx int
	if _, err := fmt.Sscanf(matches[2], "%d", &idx); err != nil {
		return "", 0, false
	}
	return matches[1], idx, true
}

// extractActualFields finds what fields were actually present in the JSON.
func extractActualFields(jsonData interface{}, arrayPath string, itemIndex int, pattern *arrayErrorPattern) {
	dataMap, ok := jsonData.(map[string]interface{})
	if !ok {
		return
	}
	arrayData, ok := dataMap[arrayPath]
	if !ok {
		return
	}
	arraySlice, ok := arrayData.([]interface{})
	if !ok || itemIndex >= len(arraySlice) {
		return
	}
	item, ok := arraySlice[itemIndex].(map[string]interface{})
	if !ok {
		return
	}
	for fieldName := range item {
		pattern.actualFields[fieldName] = true
	}
}

// createPatternBasedErrors generates contextual errors based on detected patterns.
func createPatternBasedErrors(patterns map[string]*arrayErrorPattern, schemaErrors []gojsonschema.ResultError, jsonData interface{}) []validation.ValidationError {
	var errors []validation.ValidationError
	processedFields := make(map[string]bool)

	for arrayPath, pattern := range patterns {
		msg := formatArrayPatternError(pattern)
		errors = append(errors, *validation.NewValidationError(validation.ErrJSONSchemaViolation, msg, 0))

		for _, schemaErr := range schemaErrors {
			if strings.HasPrefix(schemaErr.Field(), arrayPath+".") {
				processedFields[schemaErr.Field()] = true
			}
		}
	}

	for _, schemaErr := range schemaErrors {
		if !processedFields[schemaErr.Field()] {
			errors = append(errors, *enhanceError(schemaErr))
		}
	}
	_ = jsonData // used indirectly through patterns
	return errors
}

// formatArrayPatternError creates a detailed error message for array patterns.
func formatArrayPatternError(pattern *arrayErrorPattern) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\nARRAY STRUCTURE ERROR: '%s'\n", pattern.arrayPath))
	sb.WriteString(fmt.Sprintf("Problem: %d items in array have schema validation errors\n", len(pattern.affectedItems)))

	sb.WriteString("Common validation errors:\n")
	for errDesc, count := range pattern.errorsByField {
		sb.WriteString(fmt.Sprintf("  - %s (occurs %d times)\n", errDesc, count))
	}

	if len(pattern.actualFields) > 0 {
		sb.WriteString("Fields you provided in the items:\n")
		for field := range pattern.actualFields {
			sb.WriteString(fmt.Sprintf("  - %s\n", field))
		}
	}

	return sb.String()
}

// enhanceError adds context to individual errors.
func enhanceError(schemaErr gojsonschema.ResultError) *validation.ValidationError {
	field := schemaErr.Field()
	desc := schemaErr.Description()
	msg := fmt.Sprintf("%s: %s", field, desc)

	if strings.Contains(strings.ToLower(desc), "invalid type") {
		if strings.Contains(strings.ToLower(desc), "integer") {
			msg += "\n  Hint: Use whole numbers (1, 2, 3), NOT decimals (0.85, 3.5)"
		}
	}
	if strings.Contains(desc, "must be one of") {
		msg += "\n  Hint: Values are case-sensitive and must match exactly"
	}
	if strings.Contains(desc, "required") {
		msg += "\n  Hint: This field cannot be omitted from the JSON"
	}

	return validation.NewValidationError(validation.ErrJSONSchemaViolation, msg, 0)
}

// pathToFileURL converts a file path to a proper file:// URL.
func pathToFileURL(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	urlPath := filepath.ToSlash(absPath)

	if runtime.GOOS == "windows" {
		urlPath = strings.TrimPrefix(urlPath, "/")
		return "file:///" + urlPath
	}
	return "file://" + urlPath
}
