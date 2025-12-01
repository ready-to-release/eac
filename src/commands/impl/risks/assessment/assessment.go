// Command: create risk-assessment
// Short: Generate risk assessment from code changes
// Flag.scope: type=string, default=staged, shorthand=s, usage=File scope: staged, changed, or all, completion=staged,changed,all
// Flag.destination: type=string, shorthand=d, usage=Output file path (default: .docs/reference/risk-assessment-{timestamp}.md)
// Flag.prompt: type=string, shorthand=p, usage=Custom AI prompt file
// Flag.debug: type=bool, default=false, shorthand=D, usage=Save intermediate outputs to out/logs/risks/

package assessment

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(CreateRiskAssessment)
}

// CreateRiskAssessment is the entry point for the create risk-assessment command
func CreateRiskAssessment() int {
	return Run()
}

// Run executes the risks assessment command
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

	// Get files in scope
	files, err := getFilesInScope(config.Scope, config.WorkspaceRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	if len(files) == 0 {
		log.Errorf("No files found in scope '%s'", config.Scope)
		return 1
	}

	log.Infof("Analyzing %d file(s) in scope '%s'...", len(files), config.Scope)

	// Prepare analysis input
	input, err := prepareAnalysisInput(config, files)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	log.Infof("Loaded %d specification(s)...", len(input.Specifications))

	// Generate report
	log.Info("Generating risk assessment report...")
	report, err := generateReport(config, input)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Ensure output directory exists
	if err := ensureOutputDir(config.Destination); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Write report to file
	destPath := config.Destination
	if !filepath.IsAbs(destPath) {
		destPath = filepath.Join(config.WorkspaceRoot, destPath)
	}

	if err := os.WriteFile(destPath, []byte(report), 0644); err != nil {
		log.Errorf("Error writing report: %v", err)
		return 1
	}

	log.Infof("✓ Risk assessment report generated: %s", config.Destination)

	if config.Debug {
		debugDir := filepath.Join(config.WorkspaceRoot, "out", "logs", "risks")
		log.Infof("Debug logs saved to: %s", debugDir)
	}

	return 0
}

// showHelp displays help information
func showHelp() {
	help := `Usage: risks assessment [flags]

Generate risk assessment from code changes

Flags:
  -s, --scope <scope>          File scope: staged, changed, or all (default: staged)
  -d, --destination <path>     Output file path (default: .docs/reference/risk-assessment-{timestamp}.md)
  -p, --prompt <path>          Custom AI prompt file
  -D, --debug                  Save intermediate outputs to out/logs/risks/
  -h, --help                   Show this help message

Examples:
  risks assessment                           # Analyze staged changes
  risks assessment -s changed                # Analyze changed files
  risks assessment -s all -d report.md       # Analyze all files, custom destination
  risks assessment -D                        # Debug mode with logs
`
	log.Info(help)
}
