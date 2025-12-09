// Command: show build-summary
// Description: Generate pretty build summary for a module
// Short: Generate pretty build summary for a module
// Long: The show build-summary command generates a formatted build summary with module-specific metrics and diagnostics.
// Long: This command is designed to be used in GitHub Actions workflows to create consistent, attractive build summaries.
// Long: The output is formatted as Markdown and can be redirected to $GITHUB_STEP_SUMMARY.
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

func init() {
	registry.Register(ShowBuildSummary)
}

// ShowBuildSummary generates a pretty build summary
func ShowBuildSummary() int {
	args := os.Args[3:] // Skip program name, "show", and "build-summary"

	if len(args) == 0 {
		log.Errorf("Usage: show build-summary <module> [--status=success|failure] [--run-id=<id>]")
		return 1
	}

	module := args[0]
	status := "success"
	runID := ""

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 9 && arg[:9] == "--status=" {
			status = arg[9:]
		} else if len(arg) > 9 && arg[:9] == "--run-id=" {
			runID = arg[9:]
		}
	}

	return generateBuildSummary(module, status, runID)
}

func generateBuildSummary(moduleName, status, runID string) int {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	// Get module contract
	module, ok := cfg.Modules.GetModule(moduleName)
	if !ok {
		log.Errorf("module not found: %s", moduleName)
		return 1
	}

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
		summary += f.StatusSection(fmt.Sprintf("%s built successfully", getModuleTypeDescription(module.Type, cfg)))
	} else {
		summary += f.StatusSection(fmt.Sprintf("%s build failed", getModuleTypeDescription(module.Type, cfg)))
	}

	// Build output metrics (module-type specific)
	if status == "success" {
		summary += buildMetricsSection(f, module, cfg)
	} else {
		summary += buildDiagnosticsSection(f, module)
	}

	// Artifacts section
	summary += artifactsSection(f, module)

	// Configuration details (collapsible)
	summary += buildConfigSection(f, module, cfg)

	return summary
}

func getModuleTypeDescription(moduleType string, cfg *config.EACConfig) string {
	if cfg.ModuleTypes != nil {
		typeDef := cfg.ModuleTypes.Get(moduleType)
		if typeDef != nil {
			return typeDef.Description
		}
	}
	return moduleType
}

func buildMetricsSection(f *SummaryFormatter, module *config.Module, cfg *config.EACConfig) string {
	outputDir := filepath.Join("out", "build", module.Moniker)

	// Get artifact definitions from module type contract
	if cfg.ModuleTypes == nil {
		return f.Section(Emoji("metrics")+" Build Output", "Module types not loaded")
	}
	moduleType := cfg.ModuleTypes.Get(module.Type)
	if moduleType == nil || moduleType.Build == nil || len(moduleType.Build.Artifacts) == 0 {
		return f.Section(Emoji("metrics")+" Build Output", "No artifacts defined in contract")
	}

	// Verify artifacts using contract
	resolver := config.NewArtifactResolverWithMetadata(
		module.Moniker,
		outputDir,
		module.Metadata,
	)
	results := resolver.VerifyArtifacts(moduleType.Build.Artifacts)

	// Show what actually exists
	return formatArtifactMetrics(f, results, module.Type)
}

// formatArtifactMetrics formats artifact verification results intelligently based on types
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

// formatArtifactDetails returns size/count info for an artifact
func formatArtifactDetails(r config.ArtifactVerificationResult) string {
	info, err := os.Stat(r.Path)
	if err != nil {
		return "-"
	}

	switch r.Artifact.Type {
	case config.ArtifactTypeDirectory:
		// Show file count and total size for directories
		fileCount, _ := GetFileCount(r.Path, "**/*")
		dirSize, _ := GetDirectorySize(r.Path)
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

func buildDiagnosticsSection(f *SummaryFormatter, module *config.Module) string {
	var diagnostics string

	// Read actual build log from the correct output directory
	// Build logs are output to out/build/{module}/build.log
	logPath := filepath.Join("out", "build", module.Moniker, "build.log")
	logContent := readLogTail(logPath, 50) // Last 50 lines

	if logContent != "" {
		diagnostics += f.Section(Emoji("diagnostics")+" Build Log (last 50 lines)", f.CodeBlock("", logContent))
	} else {
		diagnostics += f.Section(Emoji("diagnostics")+" Diagnostics", fmt.Sprintf("Build failed - no log file found at %s", logPath))
	}

	// Show timing data if available
	timingPath := filepath.Join("out", "build", module.Moniker, "build-timing.txt")
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

	// Module type
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Module Type"), module.Type)

	// Build dependencies
	if cfg.ModuleTypes != nil {
		buildDeps := cfg.ModuleTypes.GetBuildDeps(module.Type)
		if len(buildDeps) > 0 {
			configDetails += fmt.Sprintf("- %s: %s\n", Bold("Dependencies"), formatSlice(buildDeps))
		}
	}

	// Output directory
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Output"), Code(fmt.Sprintf("%s/%s", paths.OutBuildRelPath, module.Moniker)))

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
