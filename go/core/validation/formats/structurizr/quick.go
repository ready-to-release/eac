package structurizr

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/validation"
)

// QuickValidator performs fast syntax validation without Docker
// Used for AI generation to provide rapid feedback (< 100ms, no Docker dependency).
type QuickValidator struct {
	identifierPattern   *regexp.Regexp
	relationshipPattern *regexp.Regexp
	contract            *domain.Contract
}

// NewQuickValidator creates a new quick validator.
func NewQuickValidator() *QuickValidator {
	return &QuickValidator{
		// Valid identifier: alphanumeric, underscore, dash (not just numbers)
		identifierPattern: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
		// Relationship pattern: identifier -> identifier "description"
		relationshipPattern: regexp.MustCompile(`(\w+)\s*->\s*(\w+)`),
		contract:            nil,
	}
}

// NewQuickValidatorWithContract creates a new quick validator with contract.
func NewQuickValidatorWithContract(contract *domain.Contract) *QuickValidator {
	v := NewQuickValidator()
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

// Validate performs quick syntax validation.
func (v *QuickValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	var errors []validation.ValidationError
	formatter := validation.NewErrorFormatter()

	lines := strings.Split(output, "\n")

	// Rule 1: Must start with "workspace"
	if err := v.validateWorkspaceStart(output, formatter); err != nil {
		errors = append(errors, *err)
	}

	// Rule 2: Check balanced braces
	if braceErrors := v.validateBalancedBraces(lines, formatter); len(braceErrors) > 0 {
		errors = append(errors, braceErrors...)
	}

	// Rule 3: Check for invalid identifiers
	if identifierErrors := v.validateIdentifiers(lines, formatter); len(identifierErrors) > 0 {
		errors = append(errors, identifierErrors...)
	}

	// Rule 4: Check for parent-child relationships (common error)
	if relationshipErrors := v.validateRelationships(lines); len(relationshipErrors) > 0 {
		errors = append(errors, relationshipErrors...)
	}

	return errors
}

// validateWorkspaceStart checks if the DSL starts with "workspace".
func (v *QuickValidator) validateWorkspaceStart(output string, formatter *validation.ErrorFormatter) *validation.ValidationError {
	trimmed := strings.TrimSpace(output)
	if !strings.HasPrefix(trimmed, "workspace") {
		// Show what AI actually generated
		lines := strings.Split(trimmed, "\n")
		firstLine := lines[0]
		if len(firstLine) > 50 {
			firstLine = firstLine[:50] + "..."
		}

		return formatter.FormatEnhancedError(
			validation.ErrMissingWorkspace,
			"DSL must start with 'workspace' keyword",
			fmt.Sprintf("AI generated output starts with:\n%s", firstLine),
			"Valid Structurizr DSL structure",
			`workspace {
  model {
    user = person "User"
    system = softwareSystem "System"
  }
  views {
    systemContext system {
      include *
    }
  }
}`,
			"Start the DSL with 'workspace {' then add 'model {' and 'views {' sections inside.",
			1,
		)
	}
	return nil
}

// validateBalancedBraces checks if braces are balanced.
func (v *QuickValidator) validateBalancedBraces(lines []string, formatter *validation.ErrorFormatter) []validation.ValidationError {
	var errors []validation.ValidationError
	braceCount := 0
	var braceStack []int // Track line numbers where braces open

	for i, line := range lines {
		lineNum := i + 1
		for _, char := range line {
			switch char {
			case '{':
				braceCount++
				braceStack = append(braceStack, lineNum)
			case '}':
				braceCount--
				if braceCount < 0 {
					errors = append(errors, *formatter.FormatEnhancedError(
						validation.ErrUnmatchedBrace,
						fmt.Sprintf("Unmatched closing brace '}' at line %d", lineNum),
						strings.TrimSpace(line),
						"Balanced braces",
						`workspace {     ← opening brace
  model {        ← opening brace
    ...
  }              ← closing brace
}                ← closing brace`,
						"Remove the extra closing brace or add a matching opening brace.",
						lineNum,
					))
					return errors // Stop on first unmatched brace
				}
				if len(braceStack) > 0 {
					braceStack = braceStack[:len(braceStack)-1]
				}
			}
		}
	}

	if braceCount > 0 {
		// Unclosed braces - create a single comprehensive error
		var openLines []string
		for _, lineNum := range braceStack {
			openLines = append(openLines, fmt.Sprintf("line %d", lineNum))
		}

		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrUnclosedBrace,
			fmt.Sprintf("%d unclosed brace(s) '{' - missing closing braces", len(braceStack)),
			fmt.Sprintf("Opening braces at: %s", strings.Join(openLines, ", ")),
			"Each opening brace must have a closing brace",
			`workspace {
  model {
    person = person "User"
  }              ← closes model
}                ← closes workspace`,
			fmt.Sprintf("Add %d closing brace(s) '}' at the appropriate locations.", len(braceStack)),
			0,
		))
	}

	return errors
}

// validateIdentifiers checks for invalid identifier names.
func (v *QuickValidator) validateIdentifiers(lines []string, formatter *validation.ErrorFormatter) []validation.ValidationError {
	var errors []validation.ValidationError

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
							// Suggest a corrected identifier
							suggested := identifier
							if len(identifier) > 0 && (identifier[0] >= '0' && identifier[0] <= '9') {
								suggested = "elem_" + identifier
							}

							errors = append(errors, *formatter.FormatEnhancedError(
								validation.ErrInvalidIdentifier,
								fmt.Sprintf("Invalid identifier '%s' at line %d", identifier, lineNum),
								fmt.Sprintf("%s %s = ...", keyword, identifier),
								"Valid identifier starting with letter, containing only alphanumeric, underscore, or dash",
								`Valid examples:
  user = person "User"
  apiSystem = softwareSystem "API"
  webApp = container "Web Application"
  db_server = softwareSystem "Database"

Invalid examples:
  123system = softwareSystem "System"    ✗ (starts with number)
  my-system! = softwareSystem "System"   ✗ (contains special char)
  system space = softwareSystem "Sys"    ✗ (contains space)`,
								fmt.Sprintf("Change '%s' to '%s' or another valid identifier that starts with a letter.", identifier, suggested),
								lineNum,
							))
						}
					}
				}
			}
		}
	}

	return errors
}

// validateRelationships checks for invalid relationships (e.g., parent-child).
func (v *QuickValidator) validateRelationships(lines []string) []validation.ValidationError {
	var errors []validation.ValidationError

	// Build a map of element hierarchies to detect parent-child relationships
	// This is a simplified check - full validation requires parsing the entire structure

	// Track elements and their potential parents
	elementHierarchy := make(map[string]string) // element -> parent
	var currentContext []string                 // Stack of current context (for nesting)

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
					errors = append(errors, *validation.NewValidationError(
						validation.ErrParentChildRelationship,
						fmt.Sprintf("Relationships cannot be added between parents and children: %s -> %s", source, target),
						lineNum,
					))
				}

				// Check if source is a parent of target
				if parent, exists := elementHierarchy[target]; exists && parent == source {
					errors = append(errors, *validation.NewValidationError(
						validation.ErrParentChildRelationship,
						fmt.Sprintf("Relationships cannot be added between parents and children: %s -> %s", source, target),
						lineNum,
					))
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

// VerifyImplementation verifies the validator is ready.
func (v *QuickValidator) VerifyImplementation() []validation.ValidationError {
	return []validation.ValidationError{}
}
