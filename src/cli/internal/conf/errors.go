package conf

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ConfigErrorType represents different types of configuration errors
type ConfigErrorType int

const (
	ErrorTypeRepositoryNotFound ConfigErrorType = iota
	ErrorTypeConfigFileNotFound
	ErrorTypeConfigFilePermission
	ErrorTypeYAMLParseError
	ErrorTypeYAMLUnmarshalError
	ErrorTypeValidationError
)

// ConfigError provides structured error information for configuration operations
type ConfigError struct {
	Type       ConfigErrorType
	Message    string
	FilePath   string
	Suggestion string
	Underlying error
}

func (ce *ConfigError) Error() string {
	var parts []string

	// Add the main error message
	parts = append(parts, ce.Message)

	// Add file path if available
	if ce.FilePath != "" {
		parts = append(parts, fmt.Sprintf("File: %s", ce.FilePath))
	}

	// Add underlying error if available
	if ce.Underlying != nil {
		parts = append(parts, fmt.Sprintf("Details: %v", ce.Underlying))
	}

	// Add suggestion if available
	if ce.Suggestion != "" {
		parts = append(parts, fmt.Sprintf("Suggestion: %s", ce.Suggestion))
	}

	// Join with newlines for proper error formatting
	return strings.Join(parts, "\n")
}

// NewRepositoryNotFoundError creates an error for when repository root cannot be found
func NewRepositoryNotFoundError(currentDir string) *ConfigError {
	return &ConfigError{
		Type:       ErrorTypeRepositoryNotFound,
		Message:    "Git repository not found",
		Suggestion: fmt.Sprintf("Navigate to a Git repository or run 'git init' in %s to initialize one", currentDir),
	}
}

// NewConfigFileNotFoundError creates an error for missing configuration files
func NewConfigFileNotFoundError(fileName, repoRoot string) *ConfigError {
	configPath := filepath.Join(repoRoot, fileName)
	return &ConfigError{
		Type:       ErrorTypeConfigFileNotFound,
		Message:    fmt.Sprintf("Configuration file %s not found", fileName),
		FilePath:   configPath,
		Suggestion: fmt.Sprintf("Create a configuration file at %s or run 'r2r init' to generate one", configPath),
	}
}

// NewConfigFilePermissionError creates an error for permission issues with config files
func NewConfigFilePermissionError(filePath string, underlying error) *ConfigError {
	return &ConfigError{
		Type:       ErrorTypeConfigFilePermission,
		Message:    "Cannot access configuration file due to permission restrictions",
		FilePath:   filePath,
		Underlying: underlying,
		Suggestion: "Check file permissions and ensure the file is readable",
	}
}

// NewYAMLParseError creates an error for YAML parsing failures
func NewYAMLParseError(filePath string, underlying error) *ConfigError {
	suggestion := "Check YAML syntax - common issues include incorrect indentation, missing colons, or invalid characters"

	// Try to provide more specific suggestions based on the error
	if underlying != nil {
		errorMsg := underlying.Error()
		if strings.Contains(errorMsg, "indent") || strings.Contains(errorMsg, "indentation") {
			suggestion = "Check YAML indentation - use spaces (not tabs) and maintain consistent indentation levels"
		} else if strings.Contains(errorMsg, "mapping") {
			suggestion = "Check YAML structure - ensure proper key:value pairs and list formatting"
		} else if strings.Contains(errorMsg, "found character") {
			suggestion = "Check for invalid characters or missing quotes around string values"
		}
	}

	return &ConfigError{
		Type:       ErrorTypeYAMLParseError,
		Message:    "Failed to parse YAML configuration file",
		FilePath:   filePath,
		Underlying: underlying,
		Suggestion: suggestion,
	}
}

// NewYAMLUnmarshalError creates an error for YAML unmarshaling failures
func NewYAMLUnmarshalError(filePath string, underlying error) *ConfigError {
	return &ConfigError{
		Type:       ErrorTypeYAMLUnmarshalError,
		Message:    "Failed to process configuration structure",
		FilePath:   filePath,
		Underlying: underlying,
		Suggestion: "Verify that the YAML structure matches the expected r2r-cli.yml schema",
	}
}

// NewValidationError creates an error for configuration validation failures
func NewValidationError(filePath string, underlying error) *ConfigError {
	suggestion := "Fix the validation errors listed above and ensure all required fields are present"

	// If it's our custom ValidationError, provide more specific guidance
	if validationErr, ok := underlying.(*ValidationError); ok {
		if len(validationErr.Errors) == 1 {
			suggestion = "Fix the validation error above"
		} else {
			suggestion = fmt.Sprintf("Fix all %d validation errors listed above", len(validationErr.Errors))
		}
	}

	return &ConfigError{
		Type:       ErrorTypeValidationError,
		Message:    "Configuration validation failed",
		FilePath:   filePath,
		Underlying: underlying,
		Suggestion: suggestion,
	}
}

// WrapConfigError wraps a generic error with configuration context
func WrapConfigError(err error, filePath string) error {
	// If it's already a ConfigError, return as-is
	if _, ok := err.(*ConfigError); ok {
		return err
	}

	// If it's our ValidationError, wrap it appropriately
	if _, ok := err.(*ValidationError); ok {
		return NewValidationError(filePath, err)
	}

	// For other errors, try to determine the type based on error message
	errorMsg := err.Error()

	if strings.Contains(errorMsg, "permission denied") || strings.Contains(errorMsg, "access is denied") {
		return NewConfigFilePermissionError(filePath, err)
	}

	if strings.Contains(errorMsg, "yaml") || strings.Contains(errorMsg, "unmarshal") {
		return NewYAMLParseError(filePath, err)
	}

	// Default to a generic config error
	return &ConfigError{
		Type:       ErrorTypeYAMLParseError, // Default assumption
		Message:    "Configuration processing failed",
		FilePath:   filePath,
		Underlying: err,
		Suggestion: "Check the configuration file format and syntax",
	}
}
