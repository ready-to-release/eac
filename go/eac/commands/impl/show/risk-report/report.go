// Command: show risk-report
// Short: Display aggregated risk assessment report
// Long: The show risk-report command displays an aggregated view of risk assessments
// Long: across all modules with assessment-results. It shows control status, evidence
// Long: links, and overall risk posture.
// Long:
// Long: The report aggregates findings from out/risk/<module>/assessment-results.json
// Long: files and provides a summary view with optional detailed breakdowns.
// Long:
// Long: Reports are automatically written to out/risk/reports/ with a timestamp.
// Flag.format: type=string, default=text, completion=text,json,markdown, usage=Output format: text, json, or markdown
// Flag.module: type=string, shorthand=m, usage=Show report for specific module only
// Flag.detail: type=bool, default=false, usage=Show detailed findings breakdown
package riskreport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(ShowRiskReport)
}

// Config holds configuration for show risk-report command.
type Config struct {
	Format        string // "text", "json", "markdown"
	Module        string // Optional: filter to specific module
	Detail        bool   // Show detailed breakdown
	WorkspaceRoot string
	OutputDir     string // Directory to write report files
}

// ModuleReport holds report data for a single module.
type ModuleReport struct {
	Module             string               `json:"module"`
	AssessmentFile     string               `json:"assessment_file"`
	Satisfied          int                  `json:"satisfied"`
	NotSatisfied       int                  `json:"not_satisfied"`
	Total              int                  `json:"total"`
	RiskScore          *scoring.RiskScore   `json:"risk_score,omitempty"`
	Findings           []FindingReport      `json:"findings,omitempty"`
}

// FindingReport holds finding information for reports.
type FindingReport struct {
	ControlID string `json:"control_id"`
	Status    string `json:"status"`
	TestCoverage string `json:"test_coverage,omitempty"`
}

// AggregatedReport holds the complete aggregated report.
type AggregatedReport struct {
	TotalModules      int             `json:"total_modules"`
	TotalControls     int             `json:"total_controls"`
	TotalSatisfied    int             `json:"total_satisfied"`
	TotalNotSatisfied int             `json:"total_not_satisfied"`
	OverallRiskBand   scoring.RiskBand `json:"overall_risk_band"`
	Modules           []ModuleReport  `json:"modules"`
}

// ShowRiskReport is the entry point for the show risk-report command.
func ShowRiskReport() int {
	config, err := parseConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showHelp()
			return 0
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	// Load assessment results
	arMap, err := oscal.DiscoverAssessmentResults(config.WorkspaceRoot)
	if err != nil {
		log.Errorf("Error discovering assessment results: %v", err)
		return 1
	}

	if len(arMap) == 0 {
		log.Error("No assessment results found")
		log.Error("")
		log.Error("You need to run risk assessment first.")
		log.Error("")
		log.Error("Steps:")
		log.Error("  1. Create a risk profile:")
		log.Error("     create risk-profile docs/security-assessment.md")
		log.Error("")
		log.Error("  2. Run risk assessment:")
		log.Error("     create risk-assess --profile specs/.risk-controls/risk-profile.json")
		log.Error("")
		log.Error("  3. View the report:")
		log.Error("     show risk-report")
		return 1
	}

	// Filter by module if specified
	if config.Module != "" {
		if _, ok := arMap[config.Module]; !ok {
			log.Errorf("No assessment results found for module: %s", config.Module)
			return 1
		}
		// Keep only the specified module
		for k := range arMap {
			if k != config.Module {
				delete(arMap, k)
			}
		}
	}

	// Build aggregated report
	report := buildAggregatedReport(config, arMap)

	// Generate output for stdout
	var stdoutOutput string
	switch config.Format {
	case "json":
		stdoutOutput = generateJSON(report)
	case "markdown":
		stdoutOutput = generateMarkdown(config, report)
	default:
		stdoutOutput = generateText(config, report)
	}

	// Generate output for file (use markdown for text format)
	var fileOutput string
	switch config.Format {
	case "json":
		fileOutput = generateJSON(report)
	default:
		// Use markdown for file output (cleaner, no color codes)
		fileOutput = generateMarkdown(config, report)
	}

	// Write to file
	if err := writeReportToFile(config, fileOutput); err != nil {
		log.Warnf("Failed to write report to file: %v", err)
	}

	// Output to stdout
	fmt.Print(stdoutOutput)

	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	args := os.Args[3:] // Skip program name, "show", and "risk-report"

	config := &Config{
		Format: "text",
	}

	// Get workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot
	config.OutputDir = filepath.Join(workspaceRoot, "out", "risk", "reports")

	// Parse flags
	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			return nil, fmt.Errorf("help requested")

		case arg == "--format":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--format requires a value")
			}
			config.Format = args[i+1]
			i += 2

		case arg == "--module" || arg == "-m":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--module requires a value")
			}
			config.Module = args[i+1]
			i += 2

		case arg == "--detail":
			config.Detail = true
			i++

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			i++
		}
	}

	// Validate format
	switch config.Format {
	case "text", "json", "markdown":
		// Valid
	default:
		return nil, fmt.Errorf("invalid format: %s (valid: text, json, markdown)", config.Format)
	}

	return config, nil
}

// buildAggregatedReport builds the aggregated report from assessment results.
func buildAggregatedReport(config *Config, arMap map[string]string) *AggregatedReport {
	report := &AggregatedReport{
		Modules: make([]ModuleReport, 0, len(arMap)),
	}

	// Get sorted module names
	moduleNames := make([]string, 0, len(arMap))
	for name := range arMap {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)

	// Process each module
	for _, moduleName := range moduleNames {
		arPath := arMap[moduleName]
		ar, err := oscal.LoadAssessmentResults(arPath)
		if err != nil {
			log.Warnf("Failed to load %s: %v", arPath, err)
			continue
		}

		moduleReport := ModuleReport{
			Module:         moduleName,
			AssessmentFile: filepath.Base(arPath),
		}

		// Process findings
		if len(ar.Results) > 0 {
			result := ar.Results[0]

			if result.Findings != nil {
				for _, finding := range *result.Findings {
					moduleReport.Total++
					report.TotalControls++

					if finding.Target.Status.State == oscal.StateSatisfied {
						moduleReport.Satisfied++
						report.TotalSatisfied++
					} else {
						moduleReport.NotSatisfied++
						report.TotalNotSatisfied++
					}

					// Always collect finding details (needed for recommendations)
					fr := FindingReport{
						ControlID: finding.Target.TargetId,
						Status:    finding.Target.Status.State,
					}

					// Extract test coverage from props
					if finding.Props != nil {
						for _, prop := range *finding.Props {
							if prop.Name == "test-coverage" {
								fr.TestCoverage = prop.Value
							}
						}
					}

					moduleReport.Findings = append(moduleReport.Findings, fr)
				}
			}
		}

		// Calculate risk score for module
		likelihood := 1
		if moduleReport.NotSatisfied > 0 {
			likelihood = scoring.CalculateBaseLikelihood(0, 0, moduleReport.NotSatisfied, 0)
		}
		impact := scoring.GetDefaultImpact("service")
		moduleReport.RiskScore = scoring.ComputeRiskScore(moduleName, likelihood, impact, "")

		report.Modules = append(report.Modules, moduleReport)
		report.TotalModules++
	}

	// Calculate overall risk band
	report.OverallRiskBand = calculateOverallRiskBand(report)

	return report
}

// calculateOverallRiskBand determines the overall risk band based on all modules.
func calculateOverallRiskBand(report *AggregatedReport) scoring.RiskBand {
	if report.TotalControls == 0 {
		return scoring.RiskLow
	}

	// Use the worst risk band from all modules
	worstBand := scoring.RiskLow
	for _, module := range report.Modules {
		if module.RiskScore != nil {
			switch module.RiskScore.Band {
			case scoring.RiskCritical:
				return scoring.RiskCritical // Return immediately for critical
			case scoring.RiskHigh:
				worstBand = scoring.RiskHigh
			case scoring.RiskMedium:
				if worstBand != scoring.RiskHigh {
					worstBand = scoring.RiskMedium
				}
			}
		}
	}

	return worstBand
}

// writeReportToFile writes the report content to a timestamped file.
func writeReportToFile(config *Config, content string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate timestamped filename with milliseconds to prevent collisions
	timestamp := time.Now().Format("2006-01-02T15-04-05.000")
	var ext string
	switch config.Format {
	case "json":
		ext = "json"
	case "markdown":
		ext = "md"
	default:
		ext = "md"
	}
	filename := fmt.Sprintf("%s-risk-report.%s", timestamp, ext)
	filepath := filepath.Join(config.OutputDir, filename)

	// Write file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Infof("Report written to: %s", filepath)
	return nil
}

// generateText generates the report in text format.
func generateText(config *Config, report *AggregatedReport) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("                      RISK ASSESSMENT REPORT\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("\n")

	// Summary
	sb.WriteString(fmt.Sprintf("  Modules Assessed:    %d\n", report.TotalModules))
	sb.WriteString(fmt.Sprintf("  Total Controls:      %d\n", report.TotalControls))
	sb.WriteString(fmt.Sprintf("  ✓ Satisfied:         %d\n", report.TotalSatisfied))
	sb.WriteString(fmt.Sprintf("  ✗ Not Satisfied:     %d\n", report.TotalNotSatisfied))
	sb.WriteString("\n")

	// Overall risk
	color := scoring.FormatRiskBandColor(report.OverallRiskBand)
	reset := scoring.ResetColor()
	sb.WriteString(fmt.Sprintf("  Overall Risk Band:   %s%s%s\n", color, report.OverallRiskBand, reset))
	sb.WriteString("\n")

	// Per-module breakdown
	sb.WriteString("───────────────────────────────────────────────────────────────\n")
	sb.WriteString("  MODULE BREAKDOWN\n")
	sb.WriteString("───────────────────────────────────────────────────────────────\n")
	sb.WriteString("\n")

	for _, module := range report.Modules {
		riskStr := "N/A"
		if module.RiskScore != nil {
			riskStr = scoring.FormatRiskScore(module.RiskScore)
		}
		sb.WriteString(fmt.Sprintf("  %-25s %d/%d satisfied    Risk: %s\n",
			module.Module, module.Satisfied, module.Total, riskStr))
		sb.WriteString(fmt.Sprintf("    Assessment: %s\n", module.AssessmentFile))

		if config.Detail && len(module.Findings) > 0 {
			for _, finding := range module.Findings {
				status := "✓"
				if finding.Status != oscal.StateSatisfied {
					status = "✗"
				}
				coverage := finding.TestCoverage
				if coverage == "" {
					coverage = "no tests"
				}
				sb.WriteString(fmt.Sprintf("    %s %-15s [%s]\n", status, finding.ControlID, coverage))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString("───────────────────────────────────────────────────────────────\n")
	sb.WriteString("  RECOMMENDATIONS\n")
	sb.WriteString("───────────────────────────────────────────────────────────────\n")
	sb.WriteString("\n")

	// Generate recommendations based on findings
	if report.TotalNotSatisfied == 0 {
		sb.WriteString("  ✓ All controls are satisfied. Continue monitoring.\n")
	} else {
		sb.WriteString(fmt.Sprintf("  ! %d control(s) need attention:\n", report.TotalNotSatisfied))
		sb.WriteString("\n")

		for _, module := range report.Modules {
			if module.NotSatisfied > 0 {
				sb.WriteString(fmt.Sprintf("    %s:\n", module.Module))
				for _, finding := range module.Findings {
					if finding.Status != oscal.StateSatisfied {
						sb.WriteString(fmt.Sprintf("      - Address %s (add tests with @control(%s) tags)\n",
							finding.ControlID, finding.ControlID))
					}
				}
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("\n")

	return sb.String()
}

// generateJSON generates the report in JSON format.
func generateJSON(report *AggregatedReport) string {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Errorf("Error marshaling report: %v", err)
		return ""
	}
	return string(data) + "\n"
}

// generateMarkdown generates the report in markdown format.
func generateMarkdown(config *Config, report *AggregatedReport) string {
	var sb strings.Builder

	sb.WriteString("# Risk Assessment Report\n")
	sb.WriteString("\n")
	sb.WriteString("## Summary\n")
	sb.WriteString("\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Modules Assessed | %d |\n", report.TotalModules))
	sb.WriteString(fmt.Sprintf("| Total Controls | %d |\n", report.TotalControls))
	sb.WriteString(fmt.Sprintf("| Satisfied | %d |\n", report.TotalSatisfied))
	sb.WriteString(fmt.Sprintf("| Not Satisfied | %d |\n", report.TotalNotSatisfied))
	sb.WriteString(fmt.Sprintf("| **Overall Risk** | **%s** |\n", report.OverallRiskBand))
	sb.WriteString("\n")

	sb.WriteString("## Module Breakdown\n")
	sb.WriteString("\n")
	sb.WriteString("| Module | Satisfied | Total | Risk Score | Risk Band | Assessment File |\n")
	sb.WriteString("|--------|-----------|-------|------------|----------|------------------|\n")

	for _, module := range report.Modules {
		score := 0
		band := "N/A"
		if module.RiskScore != nil {
			score = module.RiskScore.Score
			band = string(module.RiskScore.Band)
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %s | `%s` |\n",
			module.Module, module.Satisfied, module.Total,
			score, band, module.AssessmentFile))
	}
	sb.WriteString("\n")

	if config.Detail {
		sb.WriteString("## Detailed Findings\n")
		sb.WriteString("\n")

		for _, module := range report.Modules {
			sb.WriteString(fmt.Sprintf("### %s\n\n", module.Module))
			sb.WriteString("| Control | Status | Test Coverage |\n")
			sb.WriteString("|---------|--------|---------------|\n")

			for _, finding := range module.Findings {
				status := "✓ Satisfied"
				if finding.Status != oscal.StateSatisfied {
					status = "✗ Not Satisfied"
				}
				coverage := finding.TestCoverage
				if coverage == "" {
					coverage = "None"
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", finding.ControlID, status, coverage))
			}
			sb.WriteString("\n")
		}
	}

	if report.TotalNotSatisfied > 0 {
		sb.WriteString("## Recommendations\n")
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%d control(s) need attention:\n\n", report.TotalNotSatisfied))

		for _, module := range report.Modules {
			if module.NotSatisfied > 0 {
				sb.WriteString(fmt.Sprintf("**%s**:\n", module.Module))
				for _, finding := range module.Findings {
					if finding.Status != oscal.StateSatisfied {
						sb.WriteString(fmt.Sprintf("- Address `%s`: Add tests with `@control(%s)` tags\n",
							finding.ControlID, finding.ControlID))
					}
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// showHelp displays help information.
func showHelp() {
	help := `Usage: show risk-report [flags]

Display aggregated risk assessment report

The report is automatically written to out/risk/reports/ with a timestamp.

Flags:
      --format <format>    Output format: text, json, or markdown (default: text)
  -m, --module <name>      Show report for specific module only
      --detail             Show detailed findings breakdown
  -h, --help               Show this help message

Examples:
  # Show text report for all modules
  show risk-report

  # Show detailed report
  show risk-report --detail

  # Show report for specific module
  show risk-report --module billing-service

  # Output as JSON
  show risk-report --format json

  # Output as markdown
  show risk-report --format markdown

Output:
  - Console: Report is displayed to stdout
  - File: Report is written to out/risk/reports/YYYY-MM-DDTHH-MM-SS-risk-report.md
           (or .json for JSON format)
`
	log.Info(help)
}
