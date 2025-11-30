// Command: create risk-controls
// Short: Create risk control specifications from assessment reports
// Args: files
// Flag.force: type=bool, default=false, shorthand=f, usage=Overwrite existing risk control files
// Flag.allow-orphans: type=bool, default=false, usage=Allow overwrite even if it creates orphaned tags (requires --force)
// Flag.output: type=string, default=specs/risk-controls/, shorthand=o, usage=Output directory for controls
// Flag.prompt: type=string, shorthand=p, usage=Custom AI prompt file
// Flag.debug: type=bool, default=false, shorthand=D, usage=Save intermediate outputs to out/logs/risks/

package create

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register(CreateRiskControls)
}

// CreateRiskControls is the entry point for the create risk-controls command
func CreateRiskControls() int {
	return Run()
}

// Run executes the risks create command
func Run() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showHelp()
			return 0
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Find assessment files
	assessmentFiles, err := findAssessmentFiles(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("Processing %d assessment file(s)...\n", len(assessmentFiles))

	totalRisks := 0
	totalCreated := 0
	totalSkipped := 0
	totalFailed := 0

	// Process each assessment file
	for _, assessmentFile := range assessmentFiles {
		fmt.Printf("\nProcessing: %s\n", assessmentFile)

		// Parse assessment
		risks, err := parseAssessment(config, assessmentFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error parsing assessment: %v\n", err)
			totalFailed++
			continue
		}

		fmt.Printf("  Found %d risk(s)\n", len(risks))
		totalRisks += len(risks)

		// Generate control for each risk
		for _, risk := range risks {
			fmt.Printf("  Processing %s: %s...\n", risk.RiskID, risk.Description)

			err := generateControl(config, risk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "    Error: %v\n", err)
				totalFailed++
				continue
			}

			totalCreated++
		}
	}

	// Summary
	fmt.Printf("\n═══════════════════════════════════════════════════════════\n")
	fmt.Printf("  Summary\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")
	fmt.Printf("  Assessments processed: %d\n", len(assessmentFiles))
	fmt.Printf("  Risks identified: %d\n", totalRisks)
	fmt.Printf("  Controls created: %d\n", totalCreated)
	if totalSkipped > 0 {
		fmt.Printf("  Controls skipped: %d\n", totalSkipped)
	}
	if totalFailed > 0 {
		fmt.Printf("  Failures: %d\n", totalFailed)
	}
	fmt.Printf("\n")

	if config.Debug {
		fmt.Printf("Debug logs saved to: out/logs/risks/\n")
	}

	if totalFailed > 0 {
		return 1
	}

	return 0
}

// showHelp displays help information
func showHelp() {
	help := `Usage: risks create <assessment-file-or-folder> [flags]

Create risk control specifications from assessment reports

Arguments:
  assessment-file-or-folder    Path to assessment file or folder containing assessments

Flags:
  -f, --force                  Overwrite existing risk control files
      --allow-orphans          Allow overwrite even if it creates orphaned tags (requires --force)
  -o, --output <dir>           Output directory for controls (default: specs/risk-controls/)
  -p, --prompt <path>          Custom AI prompt file
  -D, --debug                  Save intermediate outputs to out/logs/risks/
  -h, --help                   Show this help message

Examples:
  risks create .docs/reference/assessment.md              # Create from single assessment
  risks create .docs/assessments/                         # Create from folder
  risks create assessment.md --force                      # Overwrite existing
  risks create assessment.md -o specs/custom/             # Custom output directory
`
	fmt.Print(help)
}
