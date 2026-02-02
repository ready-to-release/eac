package structurizr

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	container "github.com/ready-to-release/eac/contracts/docker-adapter/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/validation"
)

const (
	// Docker configuration.
	DockerWorkspaceMount    = "/workspace"
	StructurizrCLIImage     = "structurizr/cli:latest"
	DockerValidationTimeout = 30 * time.Second
)

// ContainerProvider is a function that returns a ContainerPort.
type ContainerProvider func() container.ContainerPort

// defaultContainerProvider is set by the application to provide the container runtime.
var defaultContainerProvider ContainerProvider

// SetContainerProvider sets the container provider for structurizr validation.
func SetContainerProvider(provider ContainerProvider) {
	defaultContainerProvider = provider
}

// DockerValidator validates Structurizr DSL using Docker
// Used for full validation in validate command (6-7 seconds).
type DockerValidator struct {
	module       string
	tempDir      string
	templateRoot string
	container    container.ContainerPort
}

// NewDockerValidator creates a new Docker-based validator.
func NewDockerValidator(module, templateRoot string) (*DockerValidator, error) {
	tempDir, err := os.MkdirTemp("", "design-validation-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &DockerValidator{
		module:       module,
		tempDir:      tempDir,
		templateRoot: templateRoot,
	}, nil
}

// Validate validates the Structurizr DSL output using Docker.
func (v *DockerValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	var errors []validation.ValidationError

	// Use standard specs path convention
	specsRoot := "specs"

	// Create module design directory in temp
	moduleDesignDir := filepath.Join(v.tempDir, specsRoot, v.module, ".design")
	if err := os.MkdirAll(moduleDesignDir, 0o755); err != nil { //nolint:gosec // G301: Temp directory for validation
		errors = append(errors, *validation.NewValidationError(
			validation.ErrSetupError,
			fmt.Sprintf("Failed to create temp design directory: %v", err),
			0,
		))
		return errors
	}

	// Write DSL to temp workspace file
	tempWorkspace := filepath.Join(moduleDesignDir, "workspace.dsl")
	if err := os.WriteFile(tempWorkspace, []byte(output), 0o644); err != nil {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrSetupError,
			fmt.Sprintf("Failed to write temp workspace: %v", err),
			0,
		))
		return errors
	}

	// Run Structurizr validation using Docker
	rawOutput, err := v.executeDockerValidation(tempWorkspace, specsRoot)
	if err != nil {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrValidationError,
			fmt.Sprintf("Validation execution failed: %v", err),
			0,
		))
		return errors
	}

	// Parse output and convert to validation errors
	parsed := v.parseValidationOutput(rawOutput)
	for _, msg := range parsed.Errors {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrDSLValidation,
			msg.Message,
			msg.Line,
		))
	}

	return errors
}

// VerifyImplementation verifies that the validator is properly configured.
func (v *DockerValidator) VerifyImplementation() []validation.ValidationError {
	return []validation.ValidationError{}
}

// IsDockerRunning checks if Docker daemon is available.
func (v *DockerValidator) IsDockerRunning() bool {
	c := v.getContainer()
	if c == nil {
		return false
	}
	return c.IsAvailable()
}

// Cleanup removes temporary files.
func (v *DockerValidator) Cleanup() {
	os.RemoveAll(v.tempDir)
}

// getContainer returns the container port, using the provider if needed.
func (v *DockerValidator) getContainer() container.ContainerPort {
	if v.container != nil {
		return v.container
	}
	if defaultContainerProvider != nil {
		return defaultContainerProvider()
	}
	return nil
}

// executeDockerValidation runs Structurizr CLI validation in Docker container.
func (v *DockerValidator) executeDockerValidation(workspacePath, specsRoot string) (string, error) {
	c := v.getContainer()
	if c == nil {
		return "", fmt.Errorf("container runtime not configured")
	}

	// Get absolute path for volume mount
	absWorkspacePath, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Mount the entire specs directory from the temp directory
	// This allows !include paths like ../../repository/.design/_styles.dsl to work
	specsDir := filepath.Join(v.tempDir, specsRoot)

	// Calculate relative path from specs dir to workspace file
	relWorkspacePath, err := filepath.Rel(specsDir, absWorkspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}
	// Convert to Unix-style path for Docker
	relWorkspacePath = filepath.ToSlash(relWorkspacePath)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), DockerValidationTimeout)
	defer cancel()

	// Build container config
	var stdout, stderr bytes.Buffer
	config := &container.ContainerConfig{
		Image:   StructurizrCLIImage,
		Command: []string{"validate", "-workspace", DockerWorkspaceMount + "/" + relWorkspacePath},
		Mounts: []container.MountConfig{
			{
				Source: specsDir,
				Target: DockerWorkspaceMount,
			},
		},
		Timeout: DockerValidationTimeout,
	}

	// Execute container
	result, err := c.Execute(ctx, config)
	if err != nil {
		// Check if timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("docker validation timed out after %v", DockerValidationTimeout)
		}
		return "", err
	}

	// Combine all output
	stdout.Write(result.Stdout)
	stderr.Write(result.Stderr)
	output := stdout.String() + stderr.String()

	return output, nil
}

// ValidationMessage represents a single error or warning.
type ValidationMessage struct {
	Severity string `json:"severity"` // "error" or "warning"
	Message  string `json:"message"`  // The validation message
	Line     int    `json:"line"`     // Line number (if available)
	Column   int    `json:"column"`   // Column number (if available)
}

// ValidationResult represents the outcome of parsing validation output.
type ValidationResult struct {
	Errors   []ValidationMessage
	Warnings []ValidationMessage
}

// parseValidationOutput parses Structurizr CLI output from Docker container.
func (v *DockerValidator) parseValidationOutput(raw string) *ValidationResult {
	result := &ValidationResult{
		Errors:   []ValidationMessage{},
		Warnings: []ValidationMessage{},
	}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for error patterns from Structurizr CLI:
		// - "Relationships cannot be added..."
		// - "ERROR: ..."
		// - "Exception: ..."
		// - "- <identifier> is not a valid identifier (expected: ...)"
		isError := strings.Contains(line, "ERROR") ||
			strings.Contains(line, "Exception") ||
			strings.Contains(line, "cannot be added") ||
			strings.Contains(line, "cannot be removed") ||
			strings.Contains(line, "is not a valid") ||
			(strings.HasPrefix(line, "- ") && strings.Contains(line, "is not"))

		if isError {
			// Extract line number from patterns like "at line 62", "line 62", or "Line 15:"
			lineNum := 0
			if matches := regexp.MustCompile(`(?i)at line (\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				if n, err := strconv.Atoi(matches[1]); err == nil {
					lineNum = n
				}
			} else if matches := regexp.MustCompile(`(?i)line (\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				if n, err := strconv.Atoi(matches[1]); err == nil {
					lineNum = n
				}
			}

			result.Errors = append(result.Errors, ValidationMessage{
				Severity: "error",
				Message:  line,
				Line:     lineNum,
			})
		}

		// Check for warning patterns
		if strings.Contains(line, "WARNING") || strings.Contains(line, "warning") {
			// Extract line number from warnings too
			lineNum := 0
			if matches := regexp.MustCompile(`(?i)line (\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				if n, err := strconv.Atoi(matches[1]); err == nil {
					lineNum = n
				}
			}

			result.Warnings = append(result.Warnings, ValidationMessage{
				Severity: "warning",
				Message:  line,
				Line:     lineNum,
			})
		}
	}

	return result
}
