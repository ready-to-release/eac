// Command: show risk-report
// Short: Display aggregated risk assessment report
// Long: The show risk-report command displays an aggregated view of risk assessments
// Long: across all modules with assessment-results. It shows control status, evidence
// Long: links, and overall risk posture.
// Long:
// Long: The report aggregates findings from out/risk/<module>/assessment-results.json
// Long: files and provides a summary view with optional detailed breakdowns.
// Flag.format: type=string, default=text, completion=text,json,markdown, usage=Output format: text, json, or markdown
// Flag.module: type=string, shorthand=m, usage=Show report for specific module only
// Flag.detail: type=bool, default=false, usage=Show detailed findings breakdown
package riskreport

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

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
}

// ModuleReport holds report data for a single module.
type ModuleReport struct {
	Module       string               `json:"module"`
	Satisfied    int                  `json:"satisfied"`
	NotSatisfied int                  `json:"not_satisfied"`
	Total        int                  `json:"total"`
	RiskScore    *scoring.RiskScore   `json:"risk_score,omitempty"`
	Findings     []FindingReport      `json:"findings,omitempty"`
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
		log.Error("Run risk assessment first:")
		log.Error("  risk assess <module>")
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

	// Output based on format
	switch config.Format {
	case "json":
		outputJSON(report)
	case "markdown":
		outputMarkdown(config, report)
	default:
		outputText(config, report)
	}

	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	args := os.Args[2:] // Skip program name and "show risk-report"

	config := &Config{
		Format: "text",
	}

	// Get workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot

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
			Module: moduleName,
		}

		// Process findings
		if len(ar.AssessmentResults.Results) > 0 {
			result := ar.AssessmentResults.Results[0]

			for _, finding := range result.Findings {
				moduleReport.Total++
				report.TotalControls++

				if finding.Target.Status.State == oscal.StateSatisfied {
					moduleReport.Satisfied++
					report.TotalSatisfied++
				} else {
					moduleReport.NotSatisfied++
					report.TotalNotSatisfied++
				}

				// Add finding detail if requested
				if config.Detail {
					fr := FindingReport{
						ControlID: finding.Target.TargetID,
						Status:    finding.Target.Status.State,
					}

					// Extract test coverage from props
					for _, prop := range finding.Props {
						if prop.Name == "test-coverage" {
							fr.TestCoverage = prop.Value
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

// outputText outputs the report in text format.
func outputText(config *Config, report *AggregatedReport) {
	log.Info("")
	log.Info("═══════════════════════════════════════════════════════════════")
	log.Info("                      RISK ASSESSMENT REPORT")
	log.Info("═══════════════════════════════════════════════════════════════")
	log.Info("")

	// Summary
	log.Infof("  Modules Assessed:    %d", report.TotalModules)
	log.Infof("  Total Controls:      %d", report.TotalControls)
	log.Infof("  ✓ Satisfied:         %d", report.TotalSatisfied)
	log.Infof("  ✗ Not Satisfied:     %d", report.TotalNotSatisfied)
	log.Info("")

	// Overall risk
	color := scoring.FormatRiskBandColor(report.OverallRiskBand)
	reset := scoring.ResetColor()
	log.Infof("  Overall Risk Band:   %s%s%s", color, report.OverallRiskBand, reset)
	log.Info("")

	// Per-module breakdown
	log.Info("───────────────────────────────────────────────────────────────")
	log.Info("  MODULE BREAKDOWN")
	log.Info("───────────────────────────────────────────────────────────────")
	log.Info("")

	for _, module := range report.Modules {
		riskStr := scoring.FormatRiskScore(module.RiskScore)
		log.Infof("  %-25s %d/%d satisfied    Risk: %s",
			module.Module, module.Satisfied, module.Total, riskStr)

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
				log.Infof("    %s %-15s [%s]", status, finding.ControlID, coverage)
			}
			log.Info("")
		}
	}

	log.Info("")
	log.Info("───────────────────────────────────────────────────────────────")
	log.Info("  RECOMMENDATIONS")
	log.Info("───────────────────────────────────────────────────────────────")
	log.Info("")

	// Generate recommendations based on findings
	if report.TotalNotSatisfied == 0 {
		log.Info("  ✓ All controls are satisfied. Continue monitoring.")
	} else {
		log.Infof("  ! %d control(s) need attention:", report.TotalNotSatisfied)
		log.Info("")

		for _, module := range report.Modules {
			if module.NotSatisfied > 0 {
				log.Infof("    %s:", module.Module)
				for _, finding := range module.Findings {
					if finding.Status != oscal.StateSatisfied {
						log.Infof("      - Address %s (add tests with @control(%s) tags)",
							finding.ControlID, finding.ControlID)
					}
				}
			}
		}
	}

	log.Info("")
	log.Info("═══════════════════════════════════════════════════════════════")
	log.Info("")
}

// outputJSON outputs the report in JSON format.
func outputJSON(report *AggregatedReport) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Errorf("Error marshaling report: %v", err)
		return
	}
	fmt.Println(string(data))
}

// outputMarkdown outputs the report in markdown format.
func outputMarkdown(config *Config, report *AggregatedReport) {
	fmt.Println("# Risk Assessment Report")
	fmt.Println()
	fmt.Println("## Summary")
	fmt.Println()
	fmt.Printf("| Metric | Value |\n")
	fmt.Printf("|--------|-------|\n")
	fmt.Printf("| Modules Assessed | %d |\n", report.TotalModules)
	fmt.Printf("| Total Controls | %d |\n", report.TotalControls)
	fmt.Printf("| Satisfied | %d |\n", report.TotalSatisfied)
	fmt.Printf("| Not Satisfied | %d |\n", report.TotalNotSatisfied)
	fmt.Printf("| **Overall Risk** | **%s** |\n", report.OverallRiskBand)
	fmt.Println()

	fmt.Println("## Module Breakdown")
	fmt.Println()
	fmt.Printf("| Module | Satisfied | Total | Risk Score | Risk Band |\n")
	fmt.Printf("|--------|-----------|-------|------------|----------|\n")

	for _, module := range report.Modules {
		fmt.Printf("| %s | %d | %d | %d | %s |\n",
			module.Module, module.Satisfied, module.Total,
			module.RiskScore.Score, module.RiskScore.Band)
	}
	fmt.Println()

	if config.Detail {
		fmt.Println("## Detailed Findings")
		fmt.Println()

		for _, module := range report.Modules {
			fmt.Printf("### %s\n\n", module.Module)
			fmt.Printf("| Control | Status | Test Coverage |\n")
			fmt.Printf("|---------|--------|---------------|\n")

			for _, finding := range module.Findings {
				status := "✓ Satisfied"
				if finding.Status != oscal.StateSatisfied {
					status = "✗ Not Satisfied"
				}
				coverage := finding.TestCoverage
				if coverage == "" {
					coverage = "None"
				}
				fmt.Printf("| %s | %s | %s |\n", finding.ControlID, status, coverage)
			}
			fmt.Println()
		}
	}

	if report.TotalNotSatisfied > 0 {
		fmt.Println("## Recommendations")
		fmt.Println()
		fmt.Printf("%d control(s) need attention:\n\n", report.TotalNotSatisfied)

		for _, module := range report.Modules {
			if module.NotSatisfied > 0 {
				fmt.Printf("**%s**:\n", module.Module)
				for _, finding := range module.Findings {
					if finding.Status != oscal.StateSatisfied {
						fmt.Printf("- Address `%s`: Add tests with `@control(%s)` tags\n",
							finding.ControlID, finding.ControlID)
					}
				}
				fmt.Println()
			}
		}
	}
}

// showHelp displays help information.
func showHelp() {
	help := `Usage: show risk-report [flags]

Display aggregated risk assessment report

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
  show risk-report --format markdown > report.md
`
	log.Info(help)
}
