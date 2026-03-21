// Package design provides validation for Structurizr workspace files
// using Structurizr CLI via Docker
package design

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/adapters/docker"
	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

var (
	reAtLineNumber = regexp.MustCompile(`(?i)at line (\d+)`)
	reLineNumber   = regexp.MustCompile(`(?i)line (\d+)`)
)

// FileValidationResult represents the outcome of validating a single DSL file.
type FileValidationResult struct {
	FileName      string              `json:"file_name"`      // File name (e.g., "workspace.dsl")
	FilePath      string              `json:"file_path"`      // Full path to file
	Valid         bool                `json:"valid"`          // Validation status for this file
	Errors        []ValidationMessage `json:"errors"`         // Validation errors
	Warnings      []ValidationMessage `json:"warnings"`       // Validation warnings
	RawOutput     string              `json:"raw_output"`     // Raw Structurizr CLI output
	ExecutionTime time.Duration       `json:"execution_time"` // Time taken to validate
}

// ValidationResult represents the outcome of validating a module's workspace(s).
type ValidationResult struct {
	Module        string                 `json:"module"`          // Module name (e.g., "clie")
	WorkspacePath string                 `json:"workspace_path"`  // Path to workspace.dsl file (backward compat)
	Valid         bool                   `json:"valid"`           // Overall validation status
	Errors        []ValidationMessage    `json:"errors"`          // Validation errors (backward compat)
	Warnings      []ValidationMessage    `json:"warnings"`        // Validation warnings (backward compat)
	RawOutput     string                 `json:"raw_output"`      // Raw Structurizr CLI output (backward compat)
	ExecutionTime time.Duration          `json:"execution_time"`  // Time taken to validate
	Timestamp     time.Time              `json:"timestamp"`       // When validation occurred
	Files         []FileValidationResult `json:"files,omitempty"` // Individual file results (multi-file mode)
	TotalFiles    int                    `json:"total_files"`     // Number of files validated
}

// ValidationMessage represents a single error or warning.
type ValidationMessage struct {
	Severity string `json:"severity"` // "error" or "warning"
	Message  string `json:"message"`  // The validation message
	Line     int    `json:"line"`     // Line number (if available)
	Column   int    `json:"column"`   // Column number (if available)
}

// ValidationSummary aggregates results for multiple modules.
type ValidationSummary struct {
	TotalModules  int                `json:"total_modules"`  // Number of modules validated
	PassedModules int                `json:"passed_modules"` // Number that passed
	FailedModules int                `json:"failed_modules"` // Number that failed
	TotalErrors   int                `json:"total_errors"`   // Sum of all errors
	TotalWarnings int                `json:"total_warnings"` // Sum of all warnings
	Results       []ValidationResult `json:"results"`        // Individual results
	ExecutionTime time.Duration      `json:"execution_time"` // Total time
	Timestamp     time.Time          `json:"timestamp"`      // When validation occurred
}

// StructurizrValidator validates workspaces using Structurizr CLI via Docker.
type StructurizrValidator interface {
	// ValidateModule validates a single module's workspace(s)
	// Validates all DSL files in the module's .design folder (except _*.dsl fragments)
	ValidateModule(moduleName string) (*ValidationResult, error)

	// ValidateModuleFile validates a specific DSL file within a module
	ValidateModuleFile(moduleName, fileName string) (*ValidationResult, error)

	// ValidateAll validates all modules with workspaces
	ValidateAll() (*ValidationSummary, error)

	// IsDockerRunning checks if Docker daemon is available
	IsDockerRunning() bool
}

// StructurizrValidatorImpl is the concrete implementation.
type StructurizrValidatorImpl struct{}

// NewValidator creates a new Structurizr validator.
// Returns a mock validator if CLIE_MOCK_DOCKER is set.
func NewValidator() (StructurizrValidator, error) {
	// Check for mock mode
	if os.Getenv(environments.EnvCLIEMockDocker) != "" {
		return NewMockValidator(), nil
	}
	return &StructurizrValidatorImpl{}, nil
}

// moduleInfo represents a module with workspace file(s).
type moduleInfo struct {
	Name  string   // Module name/moniker
	Path  string   // Path to .design directory
	Files []string // List of DSL files (full paths)
}

// listAvailableModules returns all modules that have DSL files in their .design folder
// Uses the new multi-file discovery (excludes _*.dsl fragments).
func listAvailableModules() ([]moduleInfo, error) {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := filepath.Join(repoRoot, cfg.Repository.Paths.SpecsRoot)

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

		// Use new multi-file discovery
		files, err := paths.WorkspaceDSLFiles(repoRoot, moduleName)
		if err != nil {
			continue // Skip modules with errors
		}

		// Only include modules with at least one DSL file
		if len(files) > 0 {
			modules = append(modules, moduleInfo{
				Name:  moduleName,
				Path:  paths.DesignPath(repoRoot, moduleName),
				Files: files,
			})
		}
	}

	return modules, nil
}

// IsDockerRunning checks if Docker daemon is available.
func (v *StructurizrValidatorImpl) IsDockerRunning() bool {
	return dockerutil.IsDockerAvailable()
}

// ValidateModule validates all DSL files in a module's .design folder
// Files starting with "_" are skipped (they are fragments for !include).
func (v *StructurizrValidatorImpl) ValidateModule(moduleName string) (*ValidationResult, error) {
	// Check Docker first
	if !v.IsDockerRunning() {
		return nil, fmt.Errorf("Docker is not running. Please start Docker to use validation")
	}

	files, err := discoverModuleDSLFiles(moduleName)
	if err != nil {
		return nil, err
	}

	// Start timing
	startTime := time.Now()

	// Initialize result
	result := &ValidationResult{
		Module:     moduleName,
		Valid:      true,
		Errors:     []ValidationMessage{},
		Warnings:   []ValidationMessage{},
		Files:      make([]FileValidationResult, 0, len(files)),
		TotalFiles: len(files),
		Timestamp:  time.Now(),
	}

	// Validate each file and aggregate results
	for _, filePath := range files {
		fileResult := v.validateSingleFile(filePath)
		mergeFileResult(result, fileResult, filePath)
	}

	result.ExecutionTime = time.Since(startTime)

	return result, nil
}

// discoverModuleDSLFiles locates all DSL files for a module, returning an error
// if the repository root cannot be resolved or no DSL files are found.
func discoverModuleDSLFiles(moduleName string) ([]string, error) {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	files, err := paths.WorkspaceDSLFiles(repoRoot, moduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to list DSL files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no DSL files found in %s", paths.DesignPath(repoRoot, moduleName))
	}

	return files, nil
}

// validateSingleFile runs Docker validation on one DSL file and parses the output
// into a FileValidationResult.
func (v *StructurizrValidatorImpl) validateSingleFile(filePath string) FileValidationResult {
	fileName := filepath.Base(filePath)

	fileStartTime := time.Now()
	rawOutput, err := v.executeDockerValidation(filePath)
	fileExecutionTime := time.Since(fileStartTime)

	fileResult := FileValidationResult{
		FileName:      fileName,
		FilePath:      filePath,
		Valid:         true,
		Errors:        []ValidationMessage{},
		Warnings:      []ValidationMessage{},
		RawOutput:     rawOutput,
		ExecutionTime: fileExecutionTime,
	}

	if err != nil {
		fileResult.Valid = false
		fileResult.Errors = append(fileResult.Errors, ValidationMessage{
			Severity: "error",
			Message:  fmt.Sprintf("Validation execution failed: %v", err),
		})
	} else {
		parsed := v.parseValidationOutput(rawOutput)
		fileResult.Valid = parsed.Valid
		fileResult.Errors = parsed.Errors
		fileResult.Warnings = parsed.Warnings
	}

	return fileResult
}

// mergeFileResult incorporates a single file's validation result into the overall
// module ValidationResult, including backward-compatibility fields for the first file.
func mergeFileResult(result *ValidationResult, fileResult FileValidationResult, filePath string) {
	if !fileResult.Valid {
		result.Valid = false
	}
	result.Errors = append(result.Errors, fileResult.Errors...)
	result.Warnings = append(result.Warnings, fileResult.Warnings...)
	result.Files = append(result.Files, fileResult)

	// For backward compatibility, set WorkspacePath and RawOutput from first file
	if len(result.Files) == 1 {
		result.WorkspacePath = filePath
		result.RawOutput = fileResult.RawOutput
	}
}

// ValidateModuleFile validates a specific DSL file within a module.
func (v *StructurizrValidatorImpl) ValidateModuleFile(moduleName, fileName string) (*ValidationResult, error) {
	// Check Docker first
	if !v.IsDockerRunning() {
		return nil, fmt.Errorf("Docker is not running. Please start Docker to use validation")
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Ensure fileName has .dsl extension
	if !strings.HasSuffix(fileName, ".dsl") {
		fileName = fileName + ".dsl"
	}

	// Construct file path
	filePath := filepath.Join(paths.DesignPath(repoRoot, moduleName), fileName)

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("DSL file not found: %s", filePath)
	}

	// Validate using ValidateWorkspacePath
	result, err := v.ValidateWorkspacePath(filePath)
	if err != nil {
		return nil, err
	}

	// Set module name and file info
	result.Module = moduleName
	result.TotalFiles = 1
	result.Files = []FileValidationResult{
		{
			FileName:      fileName,
			FilePath:      filePath,
			Valid:         result.Valid,
			Errors:        result.Errors,
			Warnings:      result.Warnings,
			RawOutput:     result.RawOutput,
			ExecutionTime: result.ExecutionTime,
		},
	}

	return result, nil
}

// ValidateWorkspacePath validates a workspace file at an absolute path
// This method allows validation without requiring specific directory structure.
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
	result.TotalFiles = 1

	return result, nil
}

// ValidateAll validates all modules with workspaces.
func (v *StructurizrValidatorImpl) ValidateAll() (*ValidationSummary, error) {
	// Check Docker first
	if !v.IsDockerRunning() {
		return nil, fmt.Errorf("Docker is not running. Please start Docker to use validation")
	}

	modules, err := listAvailableModules()
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules with workspace files found")
	}

	startTime := time.Now()
	summary := &ValidationSummary{
		TotalModules:  len(modules),
		Results:       make([]ValidationResult, 0, len(modules)),
		Timestamp:     time.Now(),
	}

	for _, module := range modules {
		result := v.validateModuleOrError(module)
		accumulateSummary(summary, result)
	}

	summary.ExecutionTime = time.Since(startTime)

	return summary, nil
}

// validateModuleOrError validates a single module, returning a synthetic error
// result if validation itself fails to execute.
func (v *StructurizrValidatorImpl) validateModuleOrError(module moduleInfo) *ValidationResult {
	result, err := v.ValidateModule(module.Name)
	if err != nil {
		return &ValidationResult{
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
	return result
}

// accumulateSummary adds a single module's validation result to the running summary totals.
func accumulateSummary(summary *ValidationSummary, result *ValidationResult) {
	summary.Results = append(summary.Results, *result)

	if result.Valid {
		summary.PassedModules++
	} else {
		summary.FailedModules++
	}

	summary.TotalErrors += len(result.Errors)
	summary.TotalWarnings += len(result.Warnings)
}

// executeDockerValidation runs Structurizr Lite validation in Docker container
// Uses Lite instead of CLI to ensure same validation rules as serve command.
func (v *StructurizrValidatorImpl) executeDockerValidation(workspacePath string) (string, error) {
	specsDir, relWorkspacePath, err := resolveValidationPaths(workspacePath)
	if err != nil {
		return "", err
	}

	return runDockerValidationCommand(specsDir, relWorkspacePath)
}

// resolveValidationPaths resolves the workspace path into a Docker volume mount
// and a relative workspace path suitable for the Structurizr CLI validate command.
func resolveValidationPaths(workspacePath string) (string, string, error) {
	absWorkspacePath, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get repository root to mount the entire specs directory
	// This allows !include paths like ../../repository/.design/_styles.dsl to work
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return "", "", fmt.Errorf("failed to get repository root: %w", err)
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := filepath.Join(repoRoot, cfg.Repository.Paths.SpecsRoot)

	relWorkspacePath, err := filepath.Rel(specsDir, absWorkspacePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to get relative path: %w", err)
	}
	// Convert to Unix-style path for Docker
	relWorkspacePath = filepath.ToSlash(relWorkspacePath)

	return specsDir, relWorkspacePath, nil
}

// runDockerValidationCommand executes the Structurizr CLI validate command in Docker
// and returns the combined stdout+stderr output.
func runDockerValidationCommand(specsDir, relWorkspacePath string) (string, error) {
	ctx := context.Background()

	runConfig := &docker.RunConfig{
		Image:   GetStructurizrCLIImage(),
		Command: []string{"validate", "-workspace", DockerWorkspaceMount + "/" + relWorkspacePath},
		Mounts: []docker.MountConfig{
			{Source: specsDir, Target: DockerWorkspaceMount},
		},
		Timeout:       DockerValidationTimeout,
		ContainerName: fmt.Sprintf("structurizr-validate-%d", time.Now().UnixNano()),
	}

	result, err := docker.RunContainer(ctx, runConfig)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("Docker validation timed out after %v", DockerValidationTimeout)
		}
		return "", fmt.Errorf("failed to run validation container: %w", err)
	}

	// Validation failures return non-zero exit; we parse stdout/stderr
	return result.Stdout + result.Stderr, nil
}

// parseValidationOutput parses Structurizr CLI output from Docker container.
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
		if line == "" {
			continue
		}

		if isValidationError(line) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationMessage{
				Severity: "error",
				Message:  line,
				Line:     extractLineNumber(line),
			})
		}

		if isValidationWarning(line) {
			result.Warnings = append(result.Warnings, ValidationMessage{
				Severity: "warning",
				Message:  line,
				Line:     extractLineNumber(line),
			})
		}
	}

	return result
}

// isValidationError checks whether a Structurizr CLI output line represents an error.
// Matches patterns like "ERROR: ...", "Exception: ...", "cannot be added", etc.
func isValidationError(line string) bool {
	return strings.Contains(line, "ERROR") ||
		strings.Contains(line, "Exception") ||
		strings.Contains(line, "cannot be added") ||
		strings.Contains(line, "cannot be removed") ||
		strings.Contains(line, "is not a valid") ||
		(strings.HasPrefix(line, "- ") && strings.Contains(line, "is not"))
}

// isValidationWarning checks whether a Structurizr CLI output line represents a warning.
func isValidationWarning(line string) bool {
	return strings.Contains(line, "WARNING") || strings.Contains(line, "warning")
}

// extractLineNumber extracts a line number from Structurizr CLI output text.
// Looks for patterns like "at line 62" or "line 62". Returns 0 if no line number is found.
func extractLineNumber(line string) int {
	if matches := reAtLineNumber.FindStringSubmatch(line); len(matches) > 1 {
		n, _ := strconv.Atoi(matches[1]) //nolint:errcheck // default 0 on parse error
		return n
	}
	if matches := reLineNumber.FindStringSubmatch(line); len(matches) > 1 {
		n, _ := strconv.Atoi(matches[1]) //nolint:errcheck // default 0 on parse error
		return n
	}
	return 0
}

// MarshalJSON customizes JSON encoding for time.Duration.
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

// MarshalJSON customizes JSON encoding for time.Duration in summary.
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

// MarshalJSON customizes JSON encoding for time.Duration in file results.
func (f FileValidationResult) MarshalJSON() ([]byte, error) {
	type Alias FileValidationResult
	return json.Marshal(&struct {
		ExecutionTime string `json:"execution_time"`
		*Alias
	}{
		ExecutionTime: f.ExecutionTime.String(),
		Alias:         (*Alias)(&f),
	})
}
