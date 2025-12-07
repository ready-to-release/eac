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

	// Get module type to determine what metrics to show
	switch module.Type {
	case "mkdocs-site":
		return mkdocsSiteMetrics(f, outputDir)
	case "mkdocs-pdf":
		return mkdocsPdfMetrics(f, outputDir)
	case "go-library", "go-commands", "go-cli":
		return goModuleMetrics(f, outputDir)
	case "r2r-extension":
		return extensionMetrics(f, module, cfg)
	default:
		return genericMetrics(f, outputDir)
	}
}

func mkdocsSiteMetrics(f *SummaryFormatter, outputDir string) string {
	siteDir := filepath.Join(outputDir, "site")

	htmlCount, _ := GetFileCount(siteDir, "**/*.html")
	totalFiles, _ := GetFileCount(siteDir, "**/*")
	siteSize, _ := GetDirectorySize(siteDir)

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"HTML Pages", fmt.Sprintf("%d", htmlCount)},
		{"Total Files", fmt.Sprintf("%d", totalFiles)},
		{"Site Size", siteSize},
	}

	return f.Section(Emoji("metrics")+" Build Output", f.Table(headers, rows))
}

func mkdocsPdfMetrics(f *SummaryFormatter, outputDir string) string {
	pdfCount, _ := GetFileCount(outputDir, "*.pdf")

	// List individual PDFs with sizes
	headers := []string{"Book", "Size"}
	var rows [][]string

	pdfs, _ := filepath.Glob(filepath.Join(outputDir, "*.pdf"))
	for _, pdf := range pdfs {
		info, err := os.Stat(pdf)
		if err == nil {
			rows = append(rows, []string{
				filepath.Base(pdf),
				formatBytes(info.Size()),
			})
		}
	}

	summary := fmt.Sprintf("%s PDF Books: %d\n\n", Emoji("build"), pdfCount)
	if len(rows) > 0 {
		summary += f.Table(headers, rows)
	}

	return summary
}

func goModuleMetrics(f *SummaryFormatter, outputDir string) string {
	// Check for build marker
	markerPath := filepath.Join(outputDir, "build-complete.marker")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		return f.Section(Emoji("metrics")+" Build Output", "Build artifacts not found")
	}

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Status", Emoji("success") + " Complete"},
		{"Output", Code(outputDir)},
	}

	return f.Section(Emoji("metrics")+" Build Output", f.Table(headers, rows))
}

func extensionMetrics(f *SummaryFormatter, module *config.Module, cfg *config.EACConfig) string {
	// Get docker_build config from module type
	dockerBuild := cfg.ModuleTypes.GetDockerBuildConfig(module.Type)
	if dockerBuild == nil {
		return ""
	}

	headers := []string{"Property", "Value"}
	rows := [][]string{
		{"Container", dockerBuild.Container},
		{"Platforms", formatSlice(dockerBuild.Platforms)},
		{"Tags", fmt.Sprintf("%d", len(dockerBuild.Tags))},
	}

	if dockerBuild.SBOM {
		rows = append(rows, []string{"SBOM", Emoji("success")})
	}
	if dockerBuild.Provenance {
		rows = append(rows, []string{"Provenance", Emoji("success")})
	}

	return f.Section(Emoji("build")+" Container Image", f.Table(headers, rows))
}

func genericMetrics(f *SummaryFormatter, outputDir string) string {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return f.Section(Emoji("metrics")+" Build Output", "No build artifacts found")
	}

	dirSize, _ := GetDirectorySize(outputDir)
	return f.Section(Emoji("metrics")+" Build Output", fmt.Sprintf("Output directory: %s (%s)", Code(outputDir), dirSize))
}

func buildDiagnosticsSection(f *SummaryFormatter, module *config.Module) string {
	var diagnostics string

	// Read actual build log
	logPath := filepath.Join("out", "logs", fmt.Sprintf("%s-build.log", module.Moniker))
	logContent := readLogTail(logPath, 50) // Last 50 lines

	if logContent != "" {
		diagnostics += f.Section(Emoji("diagnostics")+" Build Log (last 50 lines)", f.CodeBlock("", logContent))
	} else {
		diagnostics += f.Section(Emoji("diagnostics")+" Diagnostics", "Build failed - no log file found")
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
	configDetails += fmt.Sprintf("- %s: %s\n", Bold("Output"), Code(fmt.Sprintf("out/build/%s", module.Moniker)))

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
