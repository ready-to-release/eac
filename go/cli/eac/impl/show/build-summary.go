// Command: show build-summary
// Short: Generate pretty build summary for a module
// Long: The show build-summary command generates a formatted build summary with module-specific metrics and diagnostics.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive build summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
// Long:
// Long: Expected Output:
// Long: - Markdown-formatted build summary with emojis and styling, suitable for GitHub Actions $GITHUB_STEP_SUMMARY
// Long: - Success: includes status section, build output metrics table, artifacts section, and collapsible configuration
// Long: - Failure: includes status section, diagnostics with last 50 lines of build log, timing data, and configuration
// Flag.run-id: type=string, usage=GitHub Actions run ID for linking to workflow
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	registry.Register(ShowBuildSummary)
}

// ShowBuildSummary generates a pretty build summary.
func ShowBuildSummary() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "show", and "build-summary"

	if len(args) == 0 {
		log.Errorf("Usage: show build-summary <module> [--run-id=<id>]")
		return 1
	}

	module := args[0]
	runID := flags.GetFlagValue(args, "--run-id")

	return generateBuildSummary(module, runID)
}

func generateBuildSummary(moduleName, runID string) int {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Get module contract
	module, ok := cfg.Repository.GetModule(moduleName)
	if !ok {
		log.Errorf("module not found: %s", moduleName)
		return 1
	}

	// Derive status from build manifest
	status := deriveBuildStatus(cfg, moduleName)

	// Create formatter
	formatter := NewSummaryFormatter(moduleName, status)

	// Generate summary
	summary := buildSummaryContent(formatter, module, status, cfg)

	// Add footer with duration
	duration := time.Since(startTime)
	summary += formatter.Footer(duration)

	// Output to stdout (caller redirects to GITHUB_STEP_SUMMARY)
	fmt.Print(summary)

	return 0
}

func buildSummaryContent(f *SummaryFormatter, module *config.Module, status string, cfg *config.EACConfig) string {
	var summary string

	// Header
	summary += f.Header("build")

	// Status section
	if status == "success" {
		summary += f.StatusSection(fmt.Sprintf("%s built successfully", module.GetComponentTypesDisplay()))
	} else {
		summary += f.StatusSection(fmt.Sprintf("%s build failed", module.GetComponentTypesDisplay()))
	}

	// Build output metrics (module-type specific)
	if status == "success" {
		summary += buildMetricsSection(f, module, cfg)
	} else {
		summary += buildDiagnosticsSection(f, module, cfg)
	}

	// Artifacts section
	summary += artifactsSection(f, module)

	// Configuration details (collapsible)
	summary += buildConfigSection(f, module, cfg)

	return summary
}

func buildMetricsSection(f *SummaryFormatter, module *config.Module, cfg *config.EACConfig) string {
	outputDir := cfg.Repository.BuildOutputPath(module.Moniker)

	// Collect artifact definitions from all packages
	var artifacts []config.Artifact
	for _, pkg := range module.Components {
		if pkg != nil && pkg.Build != nil {
			for _, ma := range pkg.Build.Artifacts {
				artifacts = append(artifacts, config.Artifact{
					ID:          ma.ID,
					Type:        ma.Type,
					Pattern:     ma.Pattern,
					Compression: ma.Compression,
					DeriveFrom:  ma.DeriveFrom,
				})
			}
		}
	}

	if len(artifacts) == 0 {
		return f.Section(Emoji("metrics")+" Build Output", "No artifacts defined for module")
	}

	// Verify artifacts using module config
	resolver := config.NewArtifactResolverWithMetadata(
		module.Moniker,
		outputDir,
		module.Metadata,
	)
	results := resolver.VerifyArtifacts(artifacts)

	// Show what actually exists
	return formatArtifactMetrics(f, results, module.GetComponentTypesDisplay())
}

// formatArtifactMetrics formats artifact verification results intelligently based on types.
func formatArtifactMetrics(f *SummaryFormatter, results []config.ArtifactVerificationResult, moduleType string) string {
	if len(results) == 0 {
		return f.Section(Emoji("metrics")+" Build Output", "No artifacts defined")
	}

	// Count artifacts by type and status
	var foundArtifacts []config.ArtifactVerificationResult
	var missingArtifacts []config.ArtifactVerificationResult

	for _, r := range results {
		if r.Exists && r.Error == nil {
			foundArtifacts = append(foundArtifacts, r)
		} else {
			missingArtifacts = append(missingArtifacts, r)
		}
	}

	// If nothing found, show error
	if len(foundArtifacts) == 0 {
		return f.Section(Emoji("metrics")+" Build Output", "Build artifacts not found")
	}

	// Format based on artifact types present
	headers := []string{"Artifact", "Type", "Details"}
	var rows [][]string

	for _, r := range foundArtifacts {
		details := formatArtifactDetails(r)
		rows = append(rows, []string{
			Emoji("success") + " " + Code(r.Pattern),
			r.Artifact.Type,
			details,
		})
	}

	return f.Section(Emoji("metrics")+" Build Output", f.Table(headers, rows))
}

// formatArtifactDetails returns size/count info for an artifact.
func formatArtifactDetails(r config.ArtifactVerificationResult) string {
	info, err := os.Stat(r.Path)
	if err != nil {
		return "-"
	}

	switch r.Artifact.Type {
	case config.ArtifactTypeDirectory:
		// Show file count and total size for directories
		fileCount, fcErr := GetFileCount(r.Path, "**/*")
		if fcErr != nil {
			fileCount = 0
		}
		dirSize, dsErr := GetDirectorySize(r.Path)
		if dsErr != nil {
			dirSize = "-"
		}
		if fileCount > 0 {
			return fmt.Sprintf("%d files, %s", fileCount, dirSize)
		}
		return dirSize

	case config.ArtifactTypeFile, config.ArtifactTypeExecutable:
		// Show file size
		return formatBytes(info.Size())

	case config.ArtifactTypeImage:
		// Docker images - just confirm
		return "Built"

	default:
		return "-"
	}
}

func buildDiagnosticsSection(f *SummaryFormatter, module *config.Module, cfg *config.EACConfig) string {
	var diagnostics string

	// Search for component-level build logs in out/build/<module>/<component>/build.log
	moduleDir := cfg.Repository.BuildOutputPath(module.Moniker)
	var foundLogs []string

	entries, err := os.ReadDir(moduleDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			logPath := filepath.Join(moduleDir, entry.Name(), "build.log")
			if _, err := os.Stat(logPath); err == nil {
				foundLogs = append(foundLogs, logPath)
			}
		}
	}

	if len(foundLogs) > 0 {
		// Show logs from all components (most relevant for diagnosing failures)
		for _, logPath := range foundLogs {
			component := filepath.Base(filepath.Dir(logPath))
			logContent := readLogTail(logPath, 30) // Last 30 lines per component
			if logContent != "" {
				diagnostics += f.Section(
					fmt.Sprintf("%s Build Log: %s (last 30 lines)", Emoji("diagnostics"), component),
					f.CodeBlock("", logContent),
				)
			}
		}
	} else {
		diagnostics += f.Section(Emoji("diagnostics")+" Diagnostics", fmt.Sprintf("Build failed - no log files found in %s", moduleDir))
	}

	// Show timing data if available (check both module and component level)
	timingPath := cfg.Repository.BuildTimingPath(module.Moniker)
	if timing, err := os.ReadFile(timingPath); err == nil {
		diagnostics += f.Section(Emoji("time")+" Timing", string(timing))
	}

	return diagnostics
}

func artifactsSection(f *SummaryFormatter, module *config.Module) string {
	artifactName := fmt.Sprintf("build-artifacts-%s", module.Moniker)
	return f.Section(Emoji("artifact")+" Artifacts", fmt.Sprintf("- %s\n", Code(artifactName)))
}

func buildConfigSection(f *SummaryFormatter, module *config.Module, cfg *config.EACConfig) string {
	var configDetails string

	// Component types
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Component Types"), module.GetComponentTypesDisplay())

	// Build dependencies from package types
	if cfg.ComponentTypes != nil {
		enabledPackages := module.Components.GetEnabled()
		buildDeps := cfg.ComponentTypes.GetBuildRequirements(enabledPackages)
		if len(buildDeps) > 0 {
			configDetails += fmt.Sprintf("- %s: %s\n", Bold("Dependencies"), formatSlice(buildDeps))
		}
	}

	// Output directory
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Output"), Code(cfg.Repository.BuildOutputPath(module.Moniker)))

	return f.CollapsibleSection(Emoji("config")+" Build Configuration", configDetails)
}

func formatSlice(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	if len(items) == 1 {
		return Code(items[0])
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += Code(item)
	}
	return result
}

// deriveBuildStatus determines build status from UoW manifests.
// Status is derived as:
// - "success" if UoW manifests exist and have artifacts
// - "failure" if manifests are missing or have no artifacts.
func deriveBuildStatus(cfg *config.EACConfig, moduleName string) string {
	// Get workspace root for absolute path
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		// Can't find workspace root - assume failure
		return "failure"
	}

	// Use UoW adapter to get manifest for display
	manifest, err := implinternal.GetModuleManifestFromUoWs(
		workspaceRoot, workunit.ContextBuild, moduleName, "")
	if err != nil || manifest == nil {
		// No manifest = failure
		return "failure"
	}

	// Check if manifest has artifacts
	if len(manifest.Artifacts) == 0 {
		return "failure"
	}

	return "success"
}
