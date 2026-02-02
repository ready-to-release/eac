// Command: validate go-tidy
// Short: Validate Go module dependencies are tidy
// Long: Validates that all Go modules have tidy dependencies by running 'go mod tidy -diff'.
// Long:
// Long: Expected Output:
// Long:   Shows pass/fail status for 'go mod tidy' check on all Go modules.
// Long:   Displays diff output for untidy modules. Exit code 0 if all tidy, 1 if any untidy.
package validate

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(ValidateGoTidy)
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
	report := &goTidyReport{
		totalModules:  len(goModules),
		tidyModules:   0,
		untidyModules: make(map[string]string),
		repoRoot:      repoRoot,
	}

	for _, modulePath := range goModules {
		// Run go mod tidy -diff
		cmd := exec.Command("go", "mod", "tidy", "-diff")
		cmd.Dir = modulePath

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		output := stdout.String() + stderr.String()

		// If command failed or has output, module is not tidy
		if err != nil || strings.TrimSpace(output) != "" {
			report.untidyModules[modulePath] = output
		} else {
			report.tidyModules++
		}
	}

	return report
}

func printGoTidyReport(report *goTidyReport) {
	log.Info("=== Go Module Tidy Validation Report ===")
	log.Info("")
	log.Infof("Total Go modules: %d", report.totalModules)
	log.Infof("Tidy modules: %d", report.tidyModules)
	log.Infof("Untidy modules: %d", len(report.untidyModules))
	log.Info("")

	if len(report.untidyModules) > 0 {
		log.Info("❌ Modules with untidy dependencies:")
		for modulePath, diff := range report.untidyModules {
			relPath, relErr := filepath.Rel(report.repoRoot, modulePath)
			if relErr != nil {
				relPath = modulePath
			}
			log.Infof("\n  • %s", relPath)
			if strings.TrimSpace(diff) != "" {
				log.Infof("    Diff:\n%s", indentLines(diff, "    "))
			}
		}
		log.Info("")
		log.Info("To fix, run: go mod tidy")
		log.Info("")
	} else {
		log.Info("✅ All Go modules have tidy dependencies!")
		log.Info("")
	}
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
	log.Info("Usage: r2r validate go-tidy")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Runs 'go mod tidy -diff' on all Go modules")
	log.Info("  - Ensures go.mod and go.sum are synchronized")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate all Go modules")
	log.Info("  r2r validate go-tidy")
}
