// Command: validate go-tidy
// Description: Validate Go module dependencies are tidy
// HasSideEffects: false
package validate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(ValidateGoTidy)
}

// ValidateGoTidy validates that all Go modules have tidy dependencies
func ValidateGoTidy() int {
	args := os.Args[2:] // Skip program name and "validate"

	// Check if this is being called as a subcommand
	if len(args) > 0 && args[0] == "go-tidy" {
		args = args[1:] // Skip the subcommand name
	}

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printGoTidyUsage()
		return 0
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get repository root: %v\n", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// Discover Go modules
	var goModules []string
	for _, module := range moduleReport.Registry.All() {
		if isGoModuleType(module.Type) {
			modulePath := filepath.Join(repoRoot, module.Files.Root)
			goModules = append(goModules, modulePath)
		}
	}

	if len(goModules) == 0 {
		fmt.Println("No Go modules found in repository")
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

func isGoModuleType(moduleType string) bool {
	goModuleTypes := []string{
		"go-cli",
		"go-commands",
		"go-library",
		"go-mcp",
		"go-tests",
	}
	for _, t := range goModuleTypes {
		if moduleType == t {
			return true
		}
	}
	return false
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
	fmt.Println("=== Go Module Tidy Validation Report ===")
	fmt.Println()
	fmt.Printf("Total Go modules: %d\n", report.totalModules)
	fmt.Printf("Tidy modules: %d\n", report.tidyModules)
	fmt.Printf("Untidy modules: %d\n", len(report.untidyModules))
	fmt.Println()

	if len(report.untidyModules) > 0 {
		fmt.Printf("❌ Modules with untidy dependencies:\n")
		for modulePath, diff := range report.untidyModules {
			relPath, _ := filepath.Rel(report.repoRoot, modulePath)
			fmt.Printf("\n  • %s\n", relPath)
			if strings.TrimSpace(diff) != "" {
				fmt.Printf("    Diff:\n%s\n", indentLines(diff, "    "))
			}
		}
		fmt.Println()
		fmt.Println("To fix, run: go mod tidy")
		fmt.Println()
	} else {
		fmt.Println("✅ All Go modules have tidy dependencies!")
		fmt.Println()
	}
}

func indentLines(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

func printGoTidyUsage() {
	fmt.Println("Validate Go module dependencies are tidy")
	fmt.Println()
	fmt.Println("Usage: r2r validate go-tidy")
	fmt.Println()
	fmt.Println("Checks:")
	fmt.Println("  - Runs 'go mod tidy -diff' on all Go modules")
	fmt.Println("  - Ensures go.mod and go.sum are synchronized")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Validate all Go modules")
	fmt.Println("  r2r validate go-tidy")
}
