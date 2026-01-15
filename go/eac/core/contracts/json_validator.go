package contracts

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// JSONSchemaValidator validates JSON against a schema
type JSONSchemaValidator struct {
	SchemaPath string
	Schema     *gojsonschema.Schema
}

// arrayErrorPattern represents a detected pattern in array validation errors
type arrayErrorPattern struct {
	arrayPath     string
	affectedItems []int
	errorsByField map[string]int
	sampleIndex   int
	actualFields  map[string]bool
}

// NewJSONSchemaValidator creates a JSON schema validator
func NewJSONSchemaValidator(schemaPath string) (*JSONSchemaValidator, error) {
	// Convert file path to proper file:// URL
	fileURL := pathToFileURL(schemaPath)
	schemaLoader := gojsonschema.NewReferenceLoader(fileURL)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return &JSONSchemaValidator{
		SchemaPath: schemaPath,
		Schema:     schema,
	}, nil
}

// Validate validates JSON content against schema (implements Validator interface)
func (v *JSONSchemaValidator) Validate(output string, context map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Parse JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		return []ValidationError{
			*NewValidationError(ErrInvalidJSON, fmt.Sprintf("Invalid JSON: %v", err), 0),
		}
	}

	// Validate against schema
	documentLoader := gojsonschema.NewGoLoader(jsonData)
	result, err := v.Schema.Validate(documentLoader)
	if err != nil {
		return []ValidationError{
			*NewValidationError(ErrJSONSchemaViolation, fmt.Sprintf("Schema validation failed: %v", err), 0),
		}
	}

	// Convert schema errors to ValidationErrors with enhanced context
	if !result.Valid() {
		errors = v.processSchemaErrors(result.Errors(), jsonData)
	}

	return errors
}

// processSchemaErrors converts schema errors to contextual validation errors
func (v *JSONSchemaValidator) processSchemaErrors(schemaErrors []gojsonschema.ResultError, jsonData interface{}) []ValidationError {
	// Detect array error patterns
	patterns := v.detectArrayPatterns(schemaErrors, jsonData)

	// If we found patterns, create contextual error messages
	if len(patterns) > 0 {
		return v.createPatternBasedErrors(patterns, schemaErrors, jsonData)
	}

	// Fallback: convert errors individually with basic enhancement
	var errors []ValidationError
	for _, schemaErr := range schemaErrors {
		errors = append(errors, *v.enhanceError(schemaErr, jsonData))
	}
	return errors
}

// detectArrayPatterns finds repetitive error patterns in arrays
func (v *JSONSchemaValidator) detectArrayPatterns(schemaErrors []gojsonschema.ResultError, jsonData interface{}) map[string]*arrayErrorPattern {
	patterns := make(map[string]*arrayErrorPattern)

	for _, schemaErr := range schemaErrors {
		field := schemaErr.Field()

		// Check if this is an array error (e.g., "module_analyses.0.field")
		arrayPath, itemIndex, ok := v.parseArrayPath(field)
		if !ok {
			continue
		}

		// Create or update pattern for this array
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

		// Track affected items
		if !contains(pattern.affectedItems, itemIndex) {
			pattern.affectedItems = append(pattern.affectedItems, itemIndex)
		}

		// Track error types by field
		desc := schemaErr.Description()
		pattern.errorsByField[desc]++

		// Extract actual fields from JSON for this array item
		if itemIndex == pattern.sampleIndex {
			v.extractActualFields(jsonData, arrayPath, itemIndex, pattern)
		}
	}

	// Only return patterns that affect multiple items (threshold: 3)
	filteredPatterns := make(map[string]*arrayErrorPattern)
	for path, pattern := range patterns {
		if len(pattern.affectedItems) >= 3 {
			filteredPatterns[path] = pattern
		}
	}

	return filteredPatterns
}

// parseArrayPath extracts array path and index from field like "module_analyses.0.field"
func (v *JSONSchemaValidator) parseArrayPath(field string) (arrayPath string, itemIndex int, ok bool) {
	// Match pattern: path.number or path.number.field
	re := regexp.MustCompile(`^([a-z_]+)\.(\d+)`)
	matches := re.FindStringSubmatch(field)
	if len(matches) < 3 {
		return "", 0, false
	}

	arrayPath = matches[1]
	fmt.Sscanf(matches[2], "%d", &itemIndex)
	return arrayPath, itemIndex, true
}

// extractActualFields finds what fields were actually present in the JSON
func (v *JSONSchemaValidator) extractActualFields(jsonData interface{}, arrayPath string, itemIndex int, pattern *arrayErrorPattern) {
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

	// Record all fields present in this item
	for fieldName := range item {
		pattern.actualFields[fieldName] = true
	}
}

// createPatternBasedErrors generates contextual errors based on detected patterns
func (v *JSONSchemaValidator) createPatternBasedErrors(patterns map[string]*arrayErrorPattern, schemaErrors []gojsonschema.ResultError, jsonData interface{}) []ValidationError {
	var errors []ValidationError
	processedFields := make(map[string]bool)

	// Create one comprehensive error per array pattern
	for arrayPath, pattern := range patterns {
		msg := v.formatArrayPatternError(pattern, jsonData)
		errors = append(errors, *NewValidationError(
			ErrJSONSchemaViolation,
			msg,
			0,
		))

		// Mark all errors for this array as processed
		for _, schemaErr := range schemaErrors {
			field := schemaErr.Field()
			if strings.HasPrefix(field, arrayPath+".") {
				processedFields[field] = true
			}
		}
	}

	// Add remaining non-pattern errors
	for _, schemaErr := range schemaErrors {
		field := schemaErr.Field()
		if !processedFields[field] {
			errors = append(errors, *v.enhanceError(schemaErr, jsonData))
		}
	}

	return errors
}

// formatArrayPatternError creates a detailed error message for array patterns
func (v *JSONSchemaValidator) formatArrayPatternError(pattern *arrayErrorPattern, jsonData interface{}) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	sb.WriteString(fmt.Sprintf("ARRAY STRUCTURE ERROR: '%s'\n", pattern.arrayPath))
	sb.WriteString(fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))

	// Summary
	sb.WriteString(fmt.Sprintf("Problem: %d items in array have schema validation errors\n", len(pattern.affectedItems)))
	sb.WriteString(fmt.Sprintf("Affected items: %v\n\n", formatItemList(pattern.affectedItems)))

	// Common errors
	sb.WriteString("Common validation errors:\n")
	for errDesc, count := range pattern.errorsByField {
		sb.WriteString(fmt.Sprintf("  • %s (occurs %d times)\n", errDesc, count))
	}
	sb.WriteString("\n")

	// Show what fields were actually provided
	if len(pattern.actualFields) > 0 {
		sb.WriteString("Fields you provided in the items:\n")
		for field := range pattern.actualFields {
			sb.WriteString(fmt.Sprintf("  • %s\n", field))
		}
		sb.WriteString("\n")
	}

	// Guidance
	sb.WriteString("CORRECTION NEEDED:\n")
	sb.WriteString("  1. Review the schema requirements for this array\n")
	sb.WriteString("  2. Ensure ALL items have the required fields\n")
	sb.WriteString("  3. Check field names match the schema exactly (case-sensitive)\n")
	sb.WriteString("  4. Verify field types (integer vs decimal, string vs number)\n")
	sb.WriteString("  5. Remove any fields not defined in the schema\n")
	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return sb.String()
}

// enhanceError adds context to individual errors
func (v *JSONSchemaValidator) enhanceError(schemaErr gojsonschema.ResultError, jsonData interface{}) *ValidationError {
	field := schemaErr.Field()
	desc := schemaErr.Description()

	// Add helpful hints based on error type
	msg := fmt.Sprintf("%s: %s", field, desc)

	// Type mismatch hints
	if strings.Contains(strings.ToLower(desc), "invalid type") {
		if strings.Contains(strings.ToLower(desc), "integer") {
			msg += "\n  → Hint: Use whole numbers (1, 2, 3), NOT decimals (0.85, 3.5)"
		}
	}

	// Enum value hints
	if strings.Contains(desc, "must be one of") {
		msg += "\n  → Hint: Values are case-sensitive and must match exactly"
	}

	// Required field hints
	if strings.Contains(desc, "required") {
		msg += "\n  → Hint: This field cannot be omitted from the JSON"
	}

	return NewValidationError(ErrJSONSchemaViolation, msg, 0)
}

// Helper functions

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func formatItemList(items []int) string {
	if len(items) <= 5 {
		return fmt.Sprintf("%v", items)
	}
	return fmt.Sprintf("[0-%d] (%d total)", items[len(items)-1], len(items))
}

// VerifyImplementation verifies the validator is ready (implements Validator interface)
func (v *JSONSchemaValidator) VerifyImplementation() []ValidationError {
	// No special verification needed for JSON schema validator
	return []ValidationError{}
}

// pathToFileURL converts a file path to a proper file:// URL
// This handles Windows paths correctly by converting backslashes to forward slashes
// and ensuring the proper URL format
func pathToFileURL(path string) string {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't get absolute path, use as-is
		absPath = path
	}

	// Convert backslashes to forward slashes for URL
	urlPath := filepath.ToSlash(absPath)

	// On Windows, paths start with drive letter (e.g., C:/)
	// Need to ensure proper file:// URL format: file:///C:/path
	if runtime.GOOS == "windows" {
		// Remove leading slash if present (ToSlash might add it)
		urlPath = strings.TrimPrefix(urlPath, "/")
		return "file:///" + urlPath
	}

	// On Unix-like systems, path already starts with /
	return "file://" + urlPath
}
