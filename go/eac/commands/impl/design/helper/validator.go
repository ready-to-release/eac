// Package design provides validation for Structurizr workspace files
// using Structurizr CLI via Docker
package design

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// ValidationResult represents the outcome of validating a workspace
type ValidationResult struct {
	Module        string              `json:"module"`          // Module name (e.g., "r2r-cli")
	WorkspacePath string              `json:"workspace_path"`  // Path to workspace.dsl file
	Valid         bool                `json:"valid"`           // Overall validation status
	Errors        []ValidationMessage `json:"errors"`          // Validation errors
	Warnings      []ValidationMessage `json:"warnings"`        // Validation warnings
	RawOutput     string              `json:"raw_output"`      // Raw Structurizr CLI output
	ExecutionTime time.Duration       `json:"execution_time"`  // Time taken to validate
	Timestamp     time.Time           `json:"timestamp"`       // When validation occurred
}

// ValidationMessage represents a single error or warning
type ValidationMessage struct {
	Severity string `json:"severity"` // "error" or "warning"
	Message  string `json:"message"`  // The validation message
	Line     int    `json:"line"`     // Line number (if available)
	Column   int    `json:"column"`   // Column number (if available)
}

// ValidationSummary aggregates results for multiple modules
type ValidationSummary struct {
	TotalModules   int                `json:"total_modules"`   // Number of modules validated
	PassedModules  int                `json:"passed_modules"`  // Number that passed
	FailedModules  int                `json:"failed_modules"`  // Number that failed
	TotalErrors    int                `json:"total_errors"`    // Sum of all errors
	TotalWarnings  int                `json:"total_warnings"`  // Sum of all warnings
	Results        []ValidationResult `json:"results"`         // Individual results
	ExecutionTime  time.Duration      `json:"execution_time"`  // Total time
	Timestamp      time.Time          `json:"timestamp"`       // When validation occurred
}

// StructurizrValidator validates workspaces using Structurizr CLI via Docker
type StructurizrValidator interface {
	// ValidateModule validates a single module's workspace
	ValidateModule(moduleName string) (*ValidationResult, error)

	// ValidateAll validates all modules with workspaces
	ValidateAll() (*ValidationSummary, error)

	// IsDockerRunning checks if Docker daemon is available
	IsDockerRunning() bool
}

// StructurizrValidatorImpl is the concrete implementation
type StructurizrValidatorImpl struct {
}

// NewValidator creates a new Structurizr validator
func NewValidator() (StructurizrValidator, error) {
	return &StructurizrValidatorImpl{}, nil
}

// moduleInfo represents a module with a workspace file
type moduleInfo struct {
	Name string
	Path string
}

// listAvailableModules returns all modules that have workspace.dsl files
func listAvailableModules() ([]moduleInfo, error) {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	specsDir := repository.SpecsPath(repoRoot, "")

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return []moduleInfo{}, nil
	}

	// Read all entries in specs directory
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	var modules []moduleInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		moduleName := entry.Name()
		workspacePath := repository.WorkspaceDSLPath(repoRoot, moduleName)

		// Check if workspace.dsl exists
		if _, err := os.Stat(workspacePath); err == nil {
			modules = append(modules, moduleInfo{
				Name: moduleName,
				Path: repository.DesignPath(repoRoot, moduleName),
			})
		}
	}

	return modules, nil
}

// IsDockerRunning checks if Docker daemon is available
func (v *StructurizrValidatorImpl) IsDockerRunning() bool {
	cmd := exec.Command("docker", "ps")
	err := cmd.Run()
	return err == nil
}

// ValidateModule validates a single module's workspace
func (v *StructurizrValidatorImpl) ValidateModule(moduleName string) (*ValidationResult, error) {
	// Check Docker first
	if !v.IsDockerRunning() {
		return nil, fmt.Errorf("Docker is not running. Please start Docker to use validation")
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Construct workspace path: specs/<module>/.design/workspace.dsl
	workspacePath := repository.WorkspaceDSLPath(repoRoot, moduleName)

	// Verify workspace file exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace file not found: %s", workspacePath)
	}

	// Use ValidateWorkspacePath to perform the actual validation
	result, err := v.ValidateWorkspacePath(workspacePath)
	if err != nil {
		return nil, err
	}

	// Set module name
	result.Module = moduleName

	return result, nil
}

// ValidateWorkspacePath validates a workspace file at an absolute path
// This method allows validation without requiring specific directory structure
func (v *StructurizrValidatorImpl) ValidateWorkspacePath(workspacePath string) (*ValidationResult, error) {
	// Check Docker first
	if !v.IsDockerRunning() {
		return nil, fmt.Errorf("Docker is not running. Please start Docker to use validation")
	}

	// Verify workspace file exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace file not found: %s", workspacePath)
	}

	// Execute validation via Docker
	startTime := time.Now()
	rawOutput, err := v.executeDockerValidation(workspacePath)
	executionTime := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("validation execution failed: %w", err)
	}

	// Parse output
	result := v.parseValidationOutput(rawOutput)
	result.WorkspacePath = workspacePath
	result.ExecutionTime = executionTime
	result.Timestamp = time.Now()

	return result, nil
}

// ValidateAll validates all modules with workspaces
func (v *StructurizrValidatorImpl) ValidateAll() (*ValidationSummary, error) {
	// Check Docker first
	if !v.IsDockerRunning() {
		return nil, fmt.Errorf("Docker is not running. Please start Docker to use validation")
	}

	// Get all modules
	modules, err := listAvailableModules()
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules with workspace files found")
	}

	// Validate each module
	startTime := time.Now()
	summary := &ValidationSummary{
		TotalModules:  len(modules),
		PassedModules: 0,
		FailedModules: 0,
		TotalErrors:   0,
		TotalWarnings: 0,
		Results:       make([]ValidationResult, 0, len(modules)),
		Timestamp:     time.Now(),
	}

	for _, module := range modules {
		result, err := v.ValidateModule(module.Name)
		if err != nil {
			// Create error result
			result = &ValidationResult{
				Module:        module.Name,
				WorkspacePath: module.Path,
				Valid:         false,
				Errors: []ValidationMessage{
					{
						Severity: "error",
						Message:  fmt.Sprintf("Failed to validate: %v", err),
					},
				},
				Warnings:  []ValidationMessage{},
				Timestamp: time.Now(),
			}
		}

		summary.Results = append(summary.Results, *result)

		if result.Valid {
			summary.PassedModules++
		} else {
			summary.FailedModules++
		}

		summary.TotalErrors += len(result.Errors)
		summary.TotalWarnings += len(result.Warnings)
	}

	summary.ExecutionTime = time.Since(startTime)

	return summary, nil
}

// executeDockerValidation runs Structurizr Lite validation in Docker container
// Uses Lite instead of CLI to ensure same validation rules as serve command
func (v *StructurizrValidatorImpl) executeDockerValidation(workspacePath string) (string, error) {
	// Get absolute path for volume mount
	absWorkspacePath, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get directory containing workspace.dsl
	workspaceDir := filepath.Dir(absWorkspacePath)
	workspaceFile := filepath.Base(absWorkspacePath)

	// Convert Windows path to Docker volume format
	dockerVolume := formatDockerVolume(workspaceDir)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), DockerValidationTimeout)
	defer cancel()

	// Use Structurizr CLI for validation
	// Run it directly without -d flag so we can capture output immediately
	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"-v", dockerVolume+":"+DockerWorkspaceMount,
		StructurizrCLIImage,
		"validate",
		"-workspace", DockerWorkspaceMount+"/"+workspaceFile,
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

// parseValidationOutput parses Structurizr CLI output from Docker container
func (v *StructurizrValidatorImpl) parseValidationOutput(raw string) *ValidationResult {
	result := &ValidationResult{
		RawOutput: raw,
		Valid:     true,
		Errors:    []ValidationMessage{},
		Warnings:  []ValidationMessage{},
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
			result.Valid = false

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

// MarshalJSON customizes JSON encoding for time.Duration
func (r ValidationResult) MarshalJSON() ([]byte, error) {
	type Alias ValidationResult
	return json.Marshal(&struct {
		ExecutionTime string `json:"execution_time"`
		*Alias
	}{
		ExecutionTime: r.ExecutionTime.String(),
		Alias:         (*Alias)(&r),
	})
}

// MarshalJSON customizes JSON encoding for time.Duration in summary
func (s ValidationSummary) MarshalJSON() ([]byte, error) {
	type Alias ValidationSummary
	return json.Marshal(&struct {
		ExecutionTime string `json:"execution_time"`
		*Alias
	}{
		ExecutionTime: s.ExecutionTime.String(),
		Alias:         (*Alias)(&s),
	})
}
