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
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

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
		log.Errorf("Error: %v", err)
		return 1
	}

	// Find assessment files
	assessmentFiles, err := findAssessmentFiles(config)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	log.Infof("Processing %d assessment file(s)...", len(assessmentFiles))

	totalRisks := 0
	totalCreated := 0
	totalSkipped := 0
	totalFailed := 0

	// Process each assessment file
	for _, assessmentFile := range assessmentFiles {
		log.Infof("\nProcessing: %s", assessmentFile)

		// Parse assessment
		risks, err := parseAssessment(config, assessmentFile)
		if err != nil {
			log.Errorf("  Error parsing assessment: %v", err)
			totalFailed++
			continue
		}

		log.Infof("  Found %d risk(s)", len(risks))
		totalRisks += len(risks)

		// Generate control for each risk
		for _, risk := range risks {
			log.Infof("  Processing %s: %s...", risk.RiskID, risk.Description)

			err := generateControl(config, risk)
			if err != nil {
				log.Errorf("    Error: %v", err)
				totalFailed++
				continue
			}

			totalCreated++
		}
	}

	// Summary
	log.Info("\n═══════════════════════════════════════════════════════════")
	log.Info("  Summary")
	log.Info("═══════════════════════════════════════════════════════════\n")
	log.Infof("  Assessments processed: %d", len(assessmentFiles))
	log.Infof("  Risks identified: %d", totalRisks)
	log.Infof("  Controls created: %d", totalCreated)
	if totalSkipped > 0 {
		log.Infof("  Controls skipped: %d", totalSkipped)
	}
	if totalFailed > 0 {
		log.Infof("  Failures: %d", totalFailed)
	}
	log.Info("")

	if config.Debug {
		log.Info("Debug logs saved to: out/logs/risks/")
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
	log.Info(help)
}
