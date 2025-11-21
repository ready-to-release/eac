// Intent: Provide fast in-process DSL validation to catch common errors
// without Docker overhead
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Simple regex-based validation rules
//   - Clear error messages with line numbers
//   - Single-purpose validation functions
//
// Easy to change:
//   - Add new validation rules easily
//   - Rules are independent and composable
//   - No external dependencies
//
// Hard to break:
//   - Each rule has clear success criteria
//   - Comprehensive error reporting
//   - Fails fast on critical errors
//   - Doesn't replace full validation, just catches common cases

package create

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts"
)

// QuickDSLValidator performs fast syntax validation without Docker
type QuickDSLValidator struct {
	identifierPattern   *regexp.Regexp
	relationshipPattern *regexp.Regexp
	contract            *contracts.Contract
}

// NewQuickDSLValidator creates a new quick validator
func NewQuickDSLValidator() *QuickDSLValidator {
	return &QuickDSLValidator{
		// Valid identifier: alphanumeric, underscore, dash (not just numbers)
		identifierPattern: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
		// Relationship pattern: identifier -> identifier "description"
		relationshipPattern: regexp.MustCompile(`(\w+)\s*->\s*(\w+)`),
		contract:            nil,
	}
}

// NewQuickDSLValidatorWithContract creates a new quick validator with contract
func NewQuickDSLValidatorWithContract(contract *contracts.Contract) *QuickDSLValidator {
	v := NewQuickDSLValidator()
	v.contract = contract

	// Update identifier pattern from contract if available
	if contract != nil && contract.RawData != nil {
		if patternVal, ok := contract.RawData["identifier_pattern"].(string); ok && patternVal != "" {
			if compiled, err := regexp.Compile(patternVal); err == nil {
				v.identifierPattern = compiled
			}
		}
	}

	return v
}

// Validate performs quick syntax validation
func (v *QuickDSLValidator) Validate(output string, context map[string]interface{}) []contracts.ValidationError {
	var errors []contracts.ValidationError

	lines := strings.Split(output, "\n")

	// Rule 1: Must start with "workspace"
	if err := v.validateWorkspaceStart(output); err != nil {
		errors = append(errors, *err)
	}

	// Rule 2: Check balanced braces
	if braceErrors := v.validateBalancedBraces(lines); len(braceErrors) > 0 {
		errors = append(errors, braceErrors...)
	}

	// Rule 3: Check for invalid identifiers
	if identifierErrors := v.validateIdentifiers(lines); len(identifierErrors) > 0 {
		errors = append(errors, identifierErrors...)
	}

	// Rule 4: Check for parent-child relationships (common error)
	if relationshipErrors := v.validateRelationships(lines); len(relationshipErrors) > 0 {
		errors = append(errors, relationshipErrors...)
	}

	return errors
}

// validateWorkspaceStart checks if the DSL starts with "workspace"
func (v *QuickDSLValidator) validateWorkspaceStart(output string) *contracts.ValidationError {
	trimmed := strings.TrimSpace(output)
	if !strings.HasPrefix(trimmed, "workspace") {
		return &contracts.ValidationError{
			Code:     "missing_workspace",
			Message:  "DSL must start with 'workspace' keyword",
			Severity: "error",
			Line:     1,
		}
	}
	return nil
}

// validateBalancedBraces checks if braces are balanced
func (v *QuickDSLValidator) validateBalancedBraces(lines []string) []contracts.ValidationError {
	var errors []contracts.ValidationError
	braceCount := 0
	var braceStack []int // Track line numbers where braces open

	for i, line := range lines {
		lineNum := i + 1
		for _, char := range line {
			if char == '{' {
				braceCount++
				braceStack = append(braceStack, lineNum)
			} else if char == '}' {
				braceCount--
				if braceCount < 0 {
					errors = append(errors, contracts.ValidationError{
						Code:     "unmatched_brace",
						Message:  "Unmatched closing brace '}'",
						Severity: "error",
						Line:     lineNum,
					})
					return errors // Stop on first unmatched brace
				}
				if len(braceStack) > 0 {
					braceStack = braceStack[:len(braceStack)-1]
				}
			}
		}
	}

	if braceCount > 0 {
		// Unclosed braces
		for _, lineNum := range braceStack {
			errors = append(errors, contracts.ValidationError{
				Code:     "unclosed_brace",
				Message:  fmt.Sprintf("Unclosed brace '{' (opened at line %d)", lineNum),
				Severity: "error",
				Line:     lineNum,
			})
		}
	}

	return errors
}

// validateIdentifiers checks for invalid identifier names
func (v *QuickDSLValidator) validateIdentifiers(lines []string) []contracts.ValidationError {
	var errors []contracts.ValidationError

	// Pattern to extract identifiers from common DSL constructs
	// Matches: person identifier, softwareSystem identifier, container identifier, etc.
	identifierKeywords := []string{"person", "softwareSystem", "container", "component", "group"}

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		for _, keyword := range identifierKeywords {
			if strings.Contains(trimmed, keyword) {
				// Extract identifier after keyword
				parts := strings.Fields(trimmed)
				for j, part := range parts {
					if part == keyword && j+1 < len(parts) {
						identifier := parts[j+1]
						// Remove quotes if present
						identifier = strings.Trim(identifier, `"`)
						// Check if it's a valid identifier (not a string literal)
						if !strings.Contains(parts[j+1], `"`) && !v.identifierPattern.MatchString(identifier) {
							errors = append(errors, contracts.ValidationError{
								Code:     "invalid_identifier",
								Message:  fmt.Sprintf("Invalid identifier '%s' - must start with letter and contain only alphanumeric, underscore, or dash", identifier),
								Severity: "error",
								Line:     lineNum,
							})
						}
					}
				}
			}
		}
	}

	return errors
}

// validateRelationships checks for invalid relationships (e.g., parent-child)
func (v *QuickDSLValidator) validateRelationships(lines []string) []contracts.ValidationError {
	var errors []contracts.ValidationError

	// Build a map of element hierarchies to detect parent-child relationships
	// This is a simplified check - full validation requires parsing the entire structure

	// Track elements and their potential parents
	elementHierarchy := make(map[string]string) // element -> parent
	var currentContext []string // Stack of current context (for nesting)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Track context (simplified - just look for opening/closing braces)
		if strings.Contains(trimmed, "{") {
			// Extract identifier before '{'
			parts := strings.Fields(strings.Split(trimmed, "{")[0])
			if len(parts) >= 2 {
				identifier := parts[1]
				identifier = strings.Trim(identifier, `"`)
				currentContext = append(currentContext, identifier)
			}
		}
		if strings.Contains(trimmed, "}") && len(currentContext) > 0 {
			currentContext = currentContext[:len(currentContext)-1]
		}

		// Check for relationships (->)
		if strings.Contains(trimmed, "->") {
			matches := v.relationshipPattern.FindStringSubmatch(trimmed)
			if len(matches) >= 3 {
				source := matches[1]
				target := matches[2]

				// Check if target is a parent of source
				if parent, exists := elementHierarchy[source]; exists && parent == target {
					errors = append(errors, contracts.ValidationError{
						Code:     "parent_child_relationship",
						Message:  fmt.Sprintf("Relationships cannot be added between parents and children: %s -> %s", source, target),
						Severity: "error",
						Line:     lineNum,
					})
				}

				// Check if source is a parent of target
				if parent, exists := elementHierarchy[target]; exists && parent == source {
					errors = append(errors, contracts.ValidationError{
						Code:     "parent_child_relationship",
						Message:  fmt.Sprintf("Relationships cannot be added between parents and children: %s -> %s", source, target),
						Severity: "error",
						Line:     lineNum,
					})
				}
			}
		}

		// Track element hierarchy (simplified)
		for _, keyword := range []string{"container", "component"} {
			if strings.Contains(trimmed, keyword) && len(currentContext) > 0 {
				parts := strings.Fields(trimmed)
				for j, part := range parts {
					if part == keyword && j+1 < len(parts) {
						identifier := parts[j+1]
						identifier = strings.Trim(identifier, `"`)
						if len(currentContext) > 0 {
							elementHierarchy[identifier] = currentContext[len(currentContext)-1]
						}
					}
				}
			}
		}
	}

	return errors
}

// VerifyImplementation verifies the validator is ready
func (v *QuickDSLValidator) VerifyImplementation() []contracts.ValidationError {
	return []contracts.ValidationError{}
}
