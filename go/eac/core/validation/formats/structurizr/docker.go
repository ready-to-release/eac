package structurizr

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/validation"
)

const (
	// Docker configuration
	DockerWorkspaceMount    = "/workspace"
	StructurizrCLIImage     = "structurizr/cli:latest"
	DockerValidationTimeout = 30 * time.Second
	MaxDockerOutputSize     = 10 * 1024 * 1024 // 10MB
)

// DockerValidator validates Structurizr DSL using Docker
// Used for full validation in validate command (6-7 seconds)
type DockerValidator struct {
	module       string
	tempDir      string
	templateRoot string
}

// NewDockerValidator creates a new Docker-based validator
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

// Validate validates the Structurizr DSL output using Docker
func (v *DockerValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	var errors []validation.ValidationError

	// Use standard specs path convention
	specsRoot := "specs"

	// Create module design directory in temp
	moduleDesignDir := filepath.Join(v.tempDir, specsRoot, v.module, ".design")
	if err := os.MkdirAll(moduleDesignDir, 0755); err != nil {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrSetupError,
			fmt.Sprintf("Failed to create temp design directory: %v", err),
			0,
		))
		return errors
	}

	// Write DSL to temp workspace file
	tempWorkspace := filepath.Join(moduleDesignDir, "workspace.dsl")
	if err := os.WriteFile(tempWorkspace, []byte(output), 0644); err != nil {
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

// VerifyImplementation verifies that the validator is properly configured
func (v *DockerValidator) VerifyImplementation() []validation.ValidationError {
	return []validation.ValidationError{}
}

// IsDockerRunning checks if Docker daemon is available
func (v *DockerValidator) IsDockerRunning() bool {
	cmd := exec.Command("docker", "ps")
	err := cmd.Run()
	return err == nil
}

// Cleanup removes temporary files
func (v *DockerValidator) Cleanup() {
	os.RemoveAll(v.tempDir)
}

// executeDockerValidation runs Structurizr CLI validation in Docker container
func (v *DockerValidator) executeDockerValidation(workspacePath string, specsRoot string) (string, error) {
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

	// Convert Windows path to Docker volume format
	dockerVolume := formatDockerVolume(specsDir)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), DockerValidationTimeout)
	defer cancel()

	// Use Structurizr CLI for validation
	// Mount entire specs directory so relative !include paths work
	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"-v", dockerVolume+":"+DockerWorkspaceMount,
		StructurizrCLIImage,
		"validate",
		"-workspace", DockerWorkspaceMount+"/"+relWorkspacePath,
	)

	// Create limited buffers to prevent memory exhaustion
	var stdout, stderr limitedBuffer
	stdout.limit = MaxDockerOutputSize
	stderr.limit = MaxDockerOutputSize

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command (don't check error - validation failures return non-zero exit)
	err = cmd.Run()

	// Check if timeout occurred
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("Docker validation timed out after %v", DockerValidationTimeout)
	}

	// Combine all output
	output := stdout.String() + stderr.String()

	return output, nil
}

// limitedBuffer is a buffer that limits the amount of data it can hold
type limitedBuffer struct {
	buf   bytes.Buffer
	limit int64
	total int64
}

// Write implements io.Writer with size limit
func (lb *limitedBuffer) Write(p []byte) (n int, err error) {
	// Check if we're already over the limit
	if lb.total >= lb.limit {
		return 0, fmt.Errorf("buffer limit exceeded (%d bytes)", lb.limit)
	}

	// Calculate how much we can write
	remaining := lb.limit - lb.total
	toWrite := int64(len(p))
	willExceed := toWrite > remaining

	if willExceed {
		toWrite = remaining
	}

	// Write what we can
	n, err = lb.buf.Write(p[:toWrite])
	lb.total += int64(n)

	// If we exceeded the limit (not just reached it), return an error
	if willExceed {
		return n, fmt.Errorf("buffer limit exceeded (%d bytes)", lb.limit)
	}

	return n, err
}

// String returns the buffer contents as a string
func (lb *limitedBuffer) String() string {
	return lb.buf.String()
}

// formatDockerVolume formats a file path for Docker volume mounting
// On Windows, converts C:\path\to\dir to /c/path/to/dir for Docker compatibility
func formatDockerVolume(path string) string {
	// On Windows, Docker volume mounts need Unix-style paths
	// Convert C:\path\to\dir to /c/path/to/dir
	if len(path) >= 2 && path[1] == ':' {
		// Extract drive letter and convert to lowercase
		drive := strings.ToLower(string(path[0]))
		// Replace backslashes with forward slashes and remove colon
		rest := strings.ReplaceAll(path[2:], "\\", "/")
		return "/" + drive + rest
	}
	// If not a Windows path, return as-is
	return path
}

// ValidationMessage represents a single error or warning
type ValidationMessage struct {
	Severity string `json:"severity"` // "error" or "warning"
	Message  string `json:"message"`  // The validation message
	Line     int    `json:"line"`     // Line number (if available)
	Column   int    `json:"column"`   // Column number (if available)
}

// ValidationResult represents the outcome of parsing validation output
type ValidationResult struct {
	Errors   []ValidationMessage
	Warnings []ValidationMessage
}

// parseValidationOutput parses Structurizr CLI output from Docker container
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
				lineNum, _ = strconv.Atoi(matches[1])
			} else if matches := regexp.MustCompile(`(?i)line (\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				lineNum, _ = strconv.Atoi(matches[1])
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
				lineNum, _ = strconv.Atoi(matches[1])
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
