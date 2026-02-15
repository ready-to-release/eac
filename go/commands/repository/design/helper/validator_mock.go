package design

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MockStructurizrValidator implements StructurizrValidator for testing.
// Use the builder methods to configure behavior.
type MockStructurizrValidator struct {
	dockerRunning    bool
	moduleResults    map[string]*ValidationResult
	fileResults      map[string]map[string]*ValidationResult // module -> file -> result
	allSummary       *ValidationSummary
	fixturesDir      string

	// Error injection
	ValidateModuleError     error
	ValidateModuleFileError error
	ValidateAllError        error
}

// NewMockValidator creates a new mock Structurizr validator.
func NewMockValidator() *MockStructurizrValidator {
	return &MockStructurizrValidator{
		dockerRunning: true,
		moduleResults: make(map[string]*ValidationResult),
		fileResults:   make(map[string]map[string]*ValidationResult),
	}
}

// Builder methods for configuring the mock

// WithDockerRunning sets whether Docker is running.
func (m *MockStructurizrValidator) WithDockerRunning(running bool) *MockStructurizrValidator {
	m.dockerRunning = running
	return m
}

// WithValidResult adds a valid validation result for a module.
func (m *MockStructurizrValidator) WithValidResult(module string) *MockStructurizrValidator {
	m.moduleResults[module] = &ValidationResult{
		Module:        module,
		WorkspacePath: filepath.Join("specs", module, ".design", "workspace.dsl"),
		Valid:         true,
		Errors:        []ValidationMessage{},
		Warnings:      []ValidationMessage{},
		RawOutput:     "Workspace validation successful",
		ExecutionTime: 50 * time.Millisecond,
		Timestamp:     time.Now(),
		TotalFiles:    1,
	}
	return m
}

// WithInvalidResult adds an invalid validation result for a module.
func (m *MockStructurizrValidator) WithInvalidResult(module string, errors ...string) *MockStructurizrValidator {
	validationErrors := make([]ValidationMessage, len(errors))
	for i, err := range errors {
		validationErrors[i] = ValidationMessage{
			Severity: "error",
			Message:  err,
			Line:     i + 1,
		}
	}

	m.moduleResults[module] = &ValidationResult{
		Module:        module,
		WorkspacePath: filepath.Join("specs", module, ".design", "workspace.dsl"),
		Valid:         false,
		Errors:        validationErrors,
		Warnings:      []ValidationMessage{},
		RawOutput:     fmt.Sprintf("Validation failed with %d errors", len(errors)),
		ExecutionTime: 50 * time.Millisecond,
		Timestamp:     time.Now(),
		TotalFiles:    1,
	}
	return m
}

// WithFileResult adds a validation result for a specific file in a module.
func (m *MockStructurizrValidator) WithFileResult(module, fileName string, valid bool, errors ...string) *MockStructurizrValidator {
	if m.fileResults[module] == nil {
		m.fileResults[module] = make(map[string]*ValidationResult)
	}

	validationErrors := make([]ValidationMessage, len(errors))
	for i, err := range errors {
		validationErrors[i] = ValidationMessage{
			Severity: "error",
			Message:  err,
			Line:     i + 1,
		}
	}

	m.fileResults[module][fileName] = &ValidationResult{
		Module:        module,
		WorkspacePath: filepath.Join("specs", module, ".design", fileName),
		Valid:         valid,
		Errors:        validationErrors,
		Warnings:      []ValidationMessage{},
		RawOutput:     fmt.Sprintf("File validation: %s", fileName),
		ExecutionTime: 50 * time.Millisecond,
		Timestamp:     time.Now(),
		TotalFiles:    1,
	}
	return m
}

// WithFixturesDir sets the directory to load mock responses from.
func (m *MockStructurizrValidator) WithFixturesDir(dir string) *MockStructurizrValidator {
	m.fixturesDir = dir
	return m
}

// Interface implementation

// ValidateModule validates a single module's workspace(s).
func (m *MockStructurizrValidator) ValidateModule(moduleName string) (*ValidationResult, error) {
	if m.ValidateModuleError != nil {
		return nil, m.ValidateModuleError
	}

	// Check if we have a pre-configured result
	if result, ok := m.moduleResults[moduleName]; ok {
		return result, nil
	}

	// Try to load from fixtures if configured
	if m.fixturesDir != "" {
		if result := m.loadFixture(moduleName, ""); result != nil {
			return result, nil
		}
	}

	// Check if workspace file exists and parse it for validity
	workspacePath := filepath.Join("specs", moduleName, ".design", "workspace.dsl")
	content, err := os.ReadFile(workspacePath)
	if err != nil {
		// File doesn't exist - this is fine for mock, return valid by default
		return &ValidationResult{
			Module:        moduleName,
			WorkspacePath: workspacePath,
			Valid:         true,
			Errors:        []ValidationMessage{},
			Warnings:      []ValidationMessage{},
			RawOutput:     "Mock validation successful",
			ExecutionTime: 50 * time.Millisecond,
			Timestamp:     time.Now(),
			TotalFiles:    1,
		}, nil
	}

	// Simple validation: check for obviously invalid DSL syntax
	contentStr := string(content)
	var errors []ValidationMessage

	// Check for unmatched braces (simple heuristic)
	if openCount, closeCount := countBraces(contentStr); openCount != closeCount {
		errors = append(errors, ValidationMessage{
			Severity: "error",
			Message:  fmt.Sprintf("Mismatched braces: %d opening, %d closing", openCount, closeCount),
			Line:     1,
		})
	}

	// Check if it starts with a reasonable pattern (not "invalid {")
	if strings.HasPrefix(contentStr, "invalid") {
		errors = append(errors, ValidationMessage{
			Severity: "error",
			Message:  "Invalid DSL syntax",
			Line:     1,
		})
	}

	valid := len(errors) == 0

	return &ValidationResult{
		Module:        moduleName,
		WorkspacePath: workspacePath,
		Valid:         valid,
		Errors:        errors,
		Warnings:      []ValidationMessage{},
		RawOutput:     fmt.Sprintf("Mock validation: %d errors", len(errors)),
		ExecutionTime: 50 * time.Millisecond,
		Timestamp:     time.Now(),
		TotalFiles:    1,
	}, nil
}

// ValidateModuleFile validates a specific DSL file within a module.
func (m *MockStructurizrValidator) ValidateModuleFile(moduleName, fileName string) (*ValidationResult, error) {
	if m.ValidateModuleFileError != nil {
		return nil, m.ValidateModuleFileError
	}

	// Check if we have a pre-configured file result
	if moduleFiles, ok := m.fileResults[moduleName]; ok {
		if result, ok := moduleFiles[fileName]; ok {
			return result, nil
		}
	}

	// Try to load from fixtures if configured
	if m.fixturesDir != "" {
		if result := m.loadFixture(moduleName, fileName); result != nil {
			return result, nil
		}
	}

	// Default: return valid result
	return &ValidationResult{
		Module:        moduleName,
		WorkspacePath: filepath.Join("specs", moduleName, ".design", fileName),
		Valid:         true,
		Errors:        []ValidationMessage{},
		Warnings:      []ValidationMessage{},
		RawOutput:     fmt.Sprintf("Mock file validation successful: %s", fileName),
		ExecutionTime: 50 * time.Millisecond,
		Timestamp:     time.Now(),
		TotalFiles:    1,
	}, nil
}

// ValidateAll validates all modules with workspaces.
func (m *MockStructurizrValidator) ValidateAll() (*ValidationSummary, error) {
	if m.ValidateAllError != nil {
		return nil, m.ValidateAllError
	}

	if m.allSummary != nil {
		return m.allSummary, nil
	}

	// Build summary from individual results
	summary := &ValidationSummary{
		TotalModules:  len(m.moduleResults),
		Results:       make([]ValidationResult, 0, len(m.moduleResults)),
		ExecutionTime: 50 * time.Millisecond,
		Timestamp:     time.Now(),
	}

	for _, result := range m.moduleResults {
		summary.Results = append(summary.Results, *result)
		if result.Valid {
			summary.PassedModules++
		} else {
			summary.FailedModules++
			summary.TotalErrors += len(result.Errors)
		}
		summary.TotalWarnings += len(result.Warnings)
	}

	return summary, nil
}

// IsDockerRunning checks if Docker daemon is available.
func (m *MockStructurizrValidator) IsDockerRunning() bool {
	return m.dockerRunning
}

// loadFixture attempts to load a validation result from a fixture file.
func (m *MockStructurizrValidator) loadFixture(moduleName, fileName string) *ValidationResult {
	var fixturePath string
	if fileName != "" {
		fixturePath = filepath.Join(m.fixturesDir, "structurizr", fmt.Sprintf("%s-%s.txt", moduleName, fileName))
	} else {
		fixturePath = filepath.Join(m.fixturesDir, "structurizr", fmt.Sprintf("%s.txt", moduleName))
	}

	content, err := os.ReadFile(fixturePath)
	if err != nil {
		return nil
	}

	// Parse fixture content - simple format:
	// VALID or INVALID on first line
	// Followed by error messages (if any)
	lines := string(content)
	valid := !strings.HasPrefix(lines, "INVALID")

	return &ValidationResult{
		Module:        moduleName,
		WorkspacePath: filepath.Join("specs", moduleName, ".design", fileName),
		Valid:         valid,
		Errors:        []ValidationMessage{},
		RawOutput:     lines,
		ExecutionTime: 50 * time.Millisecond,
		Timestamp:     time.Now(),
		TotalFiles:    1,
	}
}

// countBraces counts opening and closing braces in a string.
func countBraces(s string) (int, int) {
	open := 0
	close := 0
	for _, ch := range s {
		if ch == '{' {
			open++
		} else if ch == '}' {
			close++
		}
	}
	return open, close
}

