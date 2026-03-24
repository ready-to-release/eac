package validate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/tool"
)

type validateGoTidyCommand struct{}

var _ core.SimpleCommandPort = (*validateGoTidyCommand)(nil)

func (c *validateGoTidyCommand) Name() string { return "validate go-tidy" }

func (c *validateGoTidyCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-go-tidy",
		Short:         "Validate Go module dependencies are tidy",
		Long: "Validates that all Go modules have tidy dependencies by running 'go mod tidy -diff'.",
		Notes: "Expected Output:\n  Shows pass/fail status for 'go mod tidy' check on all Go modules.\n  Displays diff output for untidy modules. Exit code 0 if all tidy, 1 if any untidy.",
	}
}

func (c *validateGoTidyCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateGoTidy()
}

// ValidateGoTidy validates that all Go modules have tidy dependencies.
func ValidateGoTidy() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "validate", and "go-tidy"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printGoTidyUsage()
		return 0
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		log.Errorf("Error: failed to load module contracts: %v", err)
		return 1
	}

	// Discover Go modules
	var goModules []string
	for _, module := range moduleReport.Registry.All() {
		if module.HasComponent("go") {
			goRoot := module.GetComponentRoot("go")
			if goRoot != "" {
				modulePath := filepath.Join(repoRoot, goRoot)
				goModules = append(goModules, modulePath)
			}
		}
	}

	if len(goModules) == 0 {
		log.Info("No Go modules found in repository")
		return 0
	}

	// Run validations
	report := validateGoModuleTidy(goModules, repoRoot)

	// Print report
	printGoTidyReport(report)

	// Return exit code based on results
	if report.HasErrors() {
		return 1
	}
	return 0
}

type goTidyReport struct {
	totalModules  int
	tidyModules   int
	untidyModules map[string]string // module path -> diff output
	repoRoot      string
}

func (r *goTidyReport) HasErrors() bool {
	return len(r.untidyModules) > 0
}

func validateGoModuleTidy(goModules []string, repoRoot string) *goTidyReport {
	ts := tool.GlobalToolSystem()

	report := &goTidyReport{
		totalModules:  len(goModules),
		tidyModules:   0,
		untidyModules: make(map[string]string),
		repoRoot:      repoRoot,
	}

	for _, modulePath := range goModules {
		// Run go mod tidy -diff
		output, exitCode, err := ts.RunToolCombined(context.Background(), "go", modulePath, "mod", "tidy", "-diff")
		if err != nil {
			report.untidyModules[modulePath] = err.Error()
			continue
		}

		// If command exited non-zero or has output, module is not tidy
		if exitCode != 0 || strings.TrimSpace(string(output)) != "" {
			report.untidyModules[modulePath] = string(output)
		} else {
			report.tidyModules++
		}
	}

	return report
}

func printGoTidyReport(report *goTidyReport) {
	// Build untidy module items - each item includes a leading blank line
	// and optional diff detail, matching the original output format.
	var untidyItems []string
	for modulePath, diff := range report.untidyModules {
		relPath, relErr := filepath.Rel(report.repoRoot, modulePath)
		if relErr != nil {
			relPath = modulePath
		}
		item := fmt.Sprintf("\n  %s", formatBullet(relPath))
		if strings.TrimSpace(diff) != "" {
			item += fmt.Sprintf("\n    Diff:\n%s", indentLines(diff, "    "))
		}
		untidyItems = append(untidyItems, item)
	}

	printValidationReportWithSummary(
		validationReport{
			Title: "Go Module Tidy Validation Report",
			Sections: []validationSection{
				{
					Icon:    "❌",
					Label:   "Modules with untidy dependencies:",
					Items:   untidyItems,
					FixHint: "To fix, run: eac update go-tidy",
				},
			},
			SuccessMessage: "All Go modules have tidy dependencies!",
		},
		[]string{
			fmt.Sprintf("Total Go modules: %d", report.totalModules),
			fmt.Sprintf("Tidy modules: %d", report.tidyModules),
			fmt.Sprintf("Untidy modules: %d", len(report.untidyModules)),
		},
	)
}

func indentLines(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

func printGoTidyUsage() {
	log.Info("Validate Go module dependencies are tidy")
	log.Info("")
	log.Info("Usage: eac validate go-tidy")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Runs 'go mod tidy -diff' on all Go modules")
	log.Info("  - Ensures go.mod and go.sum are synchronized")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate all Go modules")
	log.Info("  eac validate go-tidy")
}
