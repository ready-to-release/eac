// Package design provides export functionality for Structurizr workspace files
// using Structurizr CLI via Docker
package design

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// ExportedView represents a single exported view from a workspace
type ExportedView struct {
	ViewKey    string // View key from DSL (e.g., "SystemContext", "Containers")
	SVGPath    string // Path to exported SVG file
	Module     string // Module name the view belongs to
	DSLHash    string // First 8 chars of SHA256 hash of workspace.dsl content
}

// ExportResult represents the result of exporting a module's workspace
type ExportResult struct {
	Module        string         // Module name
	WorkspacePath string         // Path to workspace.dsl
	DSLHash       string         // First 8 chars of SHA256 hash of workspace.dsl content
	Views         []ExportedView // Exported views
	ExportDir     string         // Directory where SVGs were exported
	ExecutionTime time.Duration  // Time taken to export
	Error         error          // Any error that occurred
}

// ExportSummary aggregates export results for multiple modules
type ExportSummary struct {
	TotalModules   int            // Number of modules processed
	SuccessModules int            // Number of modules successfully exported
	FailedModules  int            // Number of modules that failed
	TotalViews     int            // Total number of views exported
	Results        []ExportResult // Individual results
	ExecutionTime  time.Duration  // Total time
}

// StructurizrExporter exports workspace views to SVG
type StructurizrExporter interface {
	// ExportModule exports all views from a module's workspace.dsl to SVG
	ExportModule(moduleName string) (*ExportResult, error)

	// ExportAll exports all modules with workspaces
	ExportAll() (*ExportSummary, error)

	// IsDockerRunning checks if Docker daemon is available
	IsDockerRunning() bool
}

// StructurizrExporterImpl is the concrete implementation
type StructurizrExporterImpl struct {
	OutputDir string // Output directory for exported SVGs (defaults to docs/assets/cache/structurizr)
}

// NewExporter creates a new Structurizr exporter
func NewExporter() (StructurizrExporter, error) {
	return &StructurizrExporterImpl{}, nil
}

// NewExporterWithOutput creates a new Structurizr exporter with custom output directory
func NewExporterWithOutput(outputDir string) (StructurizrExporter, error) {
	return &StructurizrExporterImpl{OutputDir: outputDir}, nil
}

// IsDockerRunning checks if Docker daemon is available
func (e *StructurizrExporterImpl) IsDockerRunning() bool {
	return dockerutil.IsDockerAvailable()
}

// ExportModule exports all views from a module's workspace.dsl to SVG
func (e *StructurizrExporterImpl) ExportModule(moduleName string) (*ExportResult, error) {
	startTime := time.Now()

	// Check Docker first
	if !e.IsDockerRunning() {
		return nil, fmt.Errorf("Docker is not running. Please start Docker to export diagrams")
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Find workspace.dsl file
	workspacePath := paths.WorkspaceDSLPath(repoRoot, moduleName)
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace.dsl not found for module %s at %s", moduleName, workspacePath)
	}

	// Read and hash the workspace.dsl content
	content, err := os.ReadFile(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace.dsl: %w", err)
	}
	dslHash := HashDSLContent(string(content))

	// Determine output directory
	outputDir := e.OutputDir
	if outputDir == "" {
		outputDir = paths.StructurizrCachePath(repoRoot)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create temp directory for export (Structurizr CLI exports with view key names)
	tempDir, err := os.MkdirTemp("", "structurizr-export-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Execute Docker export
	if err := e.executeDockerExport(workspacePath, tempDir, repoRoot); err != nil {
		return &ExportResult{
			Module:        moduleName,
			WorkspacePath: workspacePath,
			DSLHash:       dslHash,
			ExportDir:     outputDir,
			ExecutionTime: time.Since(startTime),
			Error:         err,
		}, nil
	}

	// Find exported SVG files and copy to final location with proper naming
	views, err := e.collectExportedViews(tempDir, outputDir, moduleName, dslHash)
	if err != nil {
		return &ExportResult{
			Module:        moduleName,
			WorkspacePath: workspacePath,
			DSLHash:       dslHash,
			ExportDir:     outputDir,
			ExecutionTime: time.Since(startTime),
			Error:         fmt.Errorf("failed to collect exported views: %w", err),
		}, nil
	}

	return &ExportResult{
		Module:        moduleName,
		WorkspacePath: workspacePath,
		DSLHash:       dslHash,
		Views:         views,
		ExportDir:     outputDir,
		ExecutionTime: time.Since(startTime),
	}, nil
}

// ExportAll exports all modules with workspaces
func (e *StructurizrExporterImpl) ExportAll() (*ExportSummary, error) {
	startTime := time.Now()

	// Get list of modules with workspaces
	modules, err := listAvailableModules()
	if err != nil {
		return nil, fmt.Errorf("failed to list available modules: %w", err)
	}

	summary := &ExportSummary{
		TotalModules: len(modules),
		Results:      make([]ExportResult, 0, len(modules)),
	}

	// Export each module
	for _, module := range modules {
		result, err := e.ExportModule(module.Name)
		if err != nil {
			summary.FailedModules++
			summary.Results = append(summary.Results, ExportResult{
				Module: module.Name,
				Error:  err,
			})
			continue
		}

		if result.Error != nil {
			summary.FailedModules++
		} else {
			summary.SuccessModules++
			summary.TotalViews += len(result.Views)
		}
		summary.Results = append(summary.Results, *result)
	}

	summary.ExecutionTime = time.Since(startTime)
	return summary, nil
}

// PlantUMLImage is the Docker image for rendering PlantUML to SVG
const PlantUMLImage = "plantuml/plantuml:latest"

// executeDockerExport runs a two-step export process:
// 1. Structurizr CLI exports to PlantUML format
// 2. PlantUML renders the .puml files to SVG
func (e *StructurizrExporterImpl) executeDockerExport(workspacePath, outputDir, repoRoot string) error {
	// Get absolute paths
	absWorkspacePath, err := filepath.Abs(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute workspace path: %w", err)
	}
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute output path: %w", err)
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: repoRoot})
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	specsDir := filepath.Join(repoRoot, cfg.Repository.Paths.SpecsRoot)

	// Calculate relative path from specs dir to workspace file
	relWorkspacePath, err := filepath.Rel(specsDir, absWorkspacePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	relWorkspacePath = filepath.ToSlash(relWorkspacePath)

	// Format paths for Docker
	dockerSpecsVolume := dockerutil.FormatDockerVolume(specsDir)
	dockerOutputVolume := dockerutil.FormatDockerVolume(absOutputDir)

	// Step 1: Export to PlantUML using Structurizr CLI
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	args := []string{
		"run", "--rm",
		"-v", dockerSpecsVolume + ":" + DockerWorkspaceMount + ":ro",
		"-v", dockerOutputVolume + ":/output",
		StructurizrCLIImage,
		"export",
		"-workspace", DockerWorkspaceMount + "/" + relWorkspacePath,
		"-format", "plantuml/c4plantuml",
		"-output", "/output",
	}

	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("structurizr export timed out after 2 minutes")
	}

	if err != nil {
		return fmt.Errorf("structurizr export failed: %w\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	// Step 2: Render PlantUML files to SVG
	// Find all .puml files in output directory
	pumlFiles, err := filepath.Glob(filepath.Join(absOutputDir, "*.puml"))
	if err != nil {
		return fmt.Errorf("failed to find puml files: %w", err)
	}

	if len(pumlFiles) == 0 {
		// No views to render is not an error
		return nil
	}

	// Render each .puml file to SVG using PlantUML
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel2()

	renderArgs := []string{
		"run", "--rm",
		"-v", dockerOutputVolume + ":/data",
		PlantUMLImage,
		"-tsvg", "/data",
	}

	renderCmd := exec.CommandContext(ctx2, "docker", renderArgs...)

	var renderStdout, renderStderr bytes.Buffer
	renderCmd.Stdout = &renderStdout
	renderCmd.Stderr = &renderStderr

	err = renderCmd.Run()

	if ctx2.Err() == context.DeadlineExceeded {
		return fmt.Errorf("plantuml render timed out after 3 minutes")
	}

	if err != nil {
		return fmt.Errorf("plantuml render failed: %w\nstdout: %s\nstderr: %s",
			err, renderStdout.String(), renderStderr.String())
	}

	// Cleanup: remove .puml files (we only need the SVGs)
	for _, pumlFile := range pumlFiles {
		os.Remove(pumlFile)
	}

	return nil
}

// collectExportedViews finds exported SVG files and copies them to final location
func (e *StructurizrExporterImpl) collectExportedViews(tempDir, outputDir, moduleName, dslHash string) ([]ExportedView, error) {
	var views []ExportedView

	// Read temp directory for exported SVGs
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".svg") {
			continue
		}

		// Extract view key from filename (e.g., "structurizr-SystemContext.svg" -> "SystemContext")
		viewKey := extractViewKey(entry.Name())
		if viewKey == "" {
			continue
		}

		// Read source file
		srcPath := filepath.Join(tempDir, entry.Name())
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read exported SVG %s: %w", entry.Name(), err)
		}

		// Build final filename: {module}_{viewKey}_{hash}.svg
		finalName := fmt.Sprintf("%s_%s_%s.svg", moduleName, viewKey, dslHash)
		finalPath := filepath.Join(outputDir, finalName)

		// Write to final location
		if err := os.WriteFile(finalPath, content, 0644); err != nil {
			return nil, fmt.Errorf("failed to write SVG to %s: %w", finalPath, err)
		}

		views = append(views, ExportedView{
			ViewKey: viewKey,
			SVGPath: finalPath,
			Module:  moduleName,
			DSLHash: dslHash,
		})
	}

	return views, nil
}

// extractViewKey extracts the view key from Structurizr CLI export filename
// Structurizr CLI exports files as "structurizr-<key>.svg" or "structurizr-<key>-<number>.svg"
var viewKeyPattern = regexp.MustCompile(`^structurizr-(.+?)(?:-\d+)?\.svg$`)

func extractViewKey(filename string) string {
	matches := viewKeyPattern.FindStringSubmatch(filename)
	if len(matches) >= 2 {
		return matches[1]
	}
	// Fallback: just remove extension and "structurizr-" prefix
	name := strings.TrimSuffix(filename, ".svg")
	name = strings.TrimPrefix(name, "structurizr-")
	return name
}

// HashDSLContent returns the first 8 characters of SHA256 hash of DSL content
// This is used for cache invalidation - all views from the same DSL share the same hash
func HashDSLContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)[:8]
}

// GetModuleDSLHash returns the DSL hash for a module's workspace.dsl
// Useful for checking cache validity without a full export
func GetModuleDSLHash(moduleName string) (string, error) {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return "", fmt.Errorf("failed to find repository root: %w", err)
	}

	workspacePath := paths.WorkspaceDSLPath(repoRoot, moduleName)
	content, err := os.ReadFile(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to read workspace.dsl: %w", err)
	}

	return HashDSLContent(string(content)), nil
}

// ListModulesWithWorkspaces returns all module names that have workspace.dsl files
func ListModulesWithWorkspaces() ([]string, error) {
	modules, err := listAvailableModules()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(modules))
	for i, m := range modules {
		names[i] = m.Name
	}
	return names, nil
}

// ParseViewKeysFromDSL extracts view keys from workspace.dsl content
// This is a lightweight parser that finds view definitions without running Structurizr
var viewDefinitionPattern = regexp.MustCompile(`(?m)^\s*(systemContext|container|component|dynamic|deployment|custom|filtered)\s+\S+\s+"([^"]+)"`)

func ParseViewKeysFromDSL(content string) []string {
	matches := viewDefinitionPattern.FindAllStringSubmatch(content, -1)
	keys := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			keys = append(keys, match[2])
		}
	}
	return keys
}
