// Package design provides validation for Structurizr workspace files
// using the Structurizr CLI via Docker
package design

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationResult represents the outcome of validating a workspace
type ValidationResult struct {
	Module        string              `json:"module"`          // Module name (e.g., "src-cli")
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
	client *Client
}

// NewValidator creates a new Structurizr validator
func NewValidator() (StructurizrValidator, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &StructurizrValidatorImpl{
		client: client,
	}, nil
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

	// Validate module exists
	if err := v.client.ValidateModule(moduleName); err != nil {
		return nil, err
	}

	// Get module info to find workspace path
	moduleInfo, err := v.client.GetModuleInfo(moduleName)
	if err != nil {
		return nil, err
	}

	// moduleInfo.Path is the design directory (e.g., specs/src-cli/design)
	// Append workspace.dsl to get the full file path
	workspacePath := filepath.Join(moduleInfo.Path, "workspace.dsl")

	// Execute validation via Docker
	startTime := time.Now()
	rawOutput, err := v.executeDockerValidation(workspacePath)
	executionTime := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("validation execution failed: %w", err)
	}

	// Parse output
	result := v.parseValidationOutput(rawOutput)
	result.Module = moduleName
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
	modules, err := v.client.ListModules()
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

// executeDockerValidation runs Structurizr CLI validation in Docker container
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

	// First, create container and capture its ID
	// Using --name with timestamp to avoid conflicts
	containerName := fmt.Sprintf("structurizr-validation-%d", time.Now().UnixNano())

	cmd := exec.Command("docker", "run",
		"--name", containerName,
		"-v", dockerVolume+":/workspace",
		"structurizr/cli",
		"validate",
		"-workspace", "/workspace/"+workspaceFile,
	)

	// Capture output from docker run
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute with timeout
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start Docker: %w", err)
	}

	// Wait with timeout (30 seconds)
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var execErr error
	select {
	case execErr = <-done:
		// Command completed
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		// Clean up container
		exec.Command("docker", "rm", "-f", containerName).Run()
		return "", fmt.Errorf("validation timed out after 30 seconds")
	}

	// Now get logs from the container
	logsCmd := exec.Command("docker", "logs", containerName)
	var logsOut, logsErr bytes.Buffer
	logsCmd.Stdout = &logsOut
	logsCmd.Stderr = &logsErr
	logsCmd.Run()

	// Combine all output sources
	output := stdout.String() + stderr.String() + logsOut.String() + logsErr.String()

	// Clean up container
	exec.Command("docker", "rm", "-f", containerName).Run()

	// If output is still empty and command succeeded, that's unusual
	if output == "" && execErr == nil {
		output = "Validation completed with no output from Structurizr CLI (checked both stdout/stderr and container logs)"
	}

	if execErr != nil {
		// Non-zero exit code - validation may have failed
		if output == "" {
			output = fmt.Sprintf("Validation failed with exit code but no output captured: %v", execErr)
		}
		return output, nil
	}

	return output, nil
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

		// Check for error patterns
		// Structurizr CLI outputs errors like:
		// "- <identifier> is not a valid identifier (expected: ...)"
		// or "ERROR: ..."
		if strings.Contains(line, "ERROR") ||
			(strings.HasPrefix(line, "- ") && strings.Contains(line, "is not")) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationMessage{
				Severity: "error",
				Message:  line,
			})
		}

		// Check for warning patterns
		if strings.Contains(line, "WARNING") || strings.Contains(line, "warning") {
			result.Warnings = append(result.Warnings, ValidationMessage{
				Severity: "warning",
				Message:  line,
			})
		}

		// Extract line numbers if present (e.g., "Line 15: error message")
		if matches := regexp.MustCompile(`Line (\d+)`).FindStringSubmatch(line); len(matches) > 1 {
			if lineNum, err := strconv.Atoi(matches[1]); err == nil {
				if len(result.Errors) > 0 {
					result.Errors[len(result.Errors)-1].Line = lineNum
				} else if len(result.Warnings) > 0 {
					result.Warnings[len(result.Warnings)-1].Line = lineNum
				}
			}
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
