// Command: validate module-files
// Description: Validate module file ownership
// HasSideEffects: false
package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/git"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(ValidateModuleFiles)
}

// ValidateModuleFiles validates file ownership in modules
func ValidateModuleFiles() int {
	args := os.Args[2:] // Skip program name and "validate"

	// Check if this is being called as a subcommand
	if len(args) > 0 && args[0] == "module-files" {
		args = args[1:] // Skip the subcommand name
	}

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printModuleFilesUsage()
		return 0
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get repository root: %v\n", err)
		return 1
	}

	// Open git repository
	repo, err := git.Open(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open git repository: %v\n", err)
		return 1
	}

	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(repo, true, false, false, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get repository files with modules: %v\n", err)
		return 1
	}

	// Run validations
	report := validateModuleFiles(files)

	// Print report
	printModuleFilesReport(report)

	// Return exit code based on results
	if report.HasErrors() {
		return 1
	}
	return 0
}

type moduleFilesReport struct {
	unorderedFiles      []string
	multiOwnershipFiles map[string][]string // file -> modules
}

func (r *moduleFilesReport) HasErrors() bool {
	return len(r.unorderedFiles) > 0 || len(r.multiOwnershipFiles) > 0
}

func validateModuleFiles(files []repository.RepositoryFileWithModule) *moduleFilesReport {
	report := &moduleFilesReport{
		unorderedFiles:      []string{},
		multiOwnershipFiles: make(map[string][]string),
	}

	for _, file := range files {
		// Check for unordered files
		for _, moduleName := range file.Modules {
			if moduleName == "unordered" {
				report.unorderedFiles = append(report.unorderedFiles, file.Name)
				break
			}
		}

		// Check for multi-module ownership
		if len(file.Modules) > 1 {
			report.multiOwnershipFiles[file.Name] = file.Modules
		}
	}

	return report
}

func printModuleFilesReport(report *moduleFilesReport) {
	fmt.Println("=== Module File Ownership Validation Report ===")
	fmt.Println()

	hasIssues := false

	// Unordered files
	if len(report.unorderedFiles) > 0 {
		hasIssues = true
		fmt.Printf("❌ Files in Unordered Module (%d):\n", len(report.unorderedFiles))
		fmt.Println("   These files should be claimed by a proper module:")
		for _, filePath := range report.unorderedFiles {
			fmt.Printf("  • %s\n", filePath)
		}
		fmt.Println()
		fmt.Println("   Fix: Create or update module contracts to claim these files.")
		fmt.Println()
	}

	// Multi-ownership files
	if len(report.multiOwnershipFiles) > 0 {
		hasIssues = true
		fmt.Printf("❌ Files with Multi-Module Ownership (%d):\n", len(report.multiOwnershipFiles))
		fmt.Println("   Each file should belong to exactly one module:")
		for filePath, modules := range report.multiOwnershipFiles {
			fmt.Printf("  • %s\n", filePath)
			fmt.Printf("    Claimed by: %s\n", strings.Join(modules, ", "))
		}
		fmt.Println()
		fmt.Println("   Fix: Adjust module contract glob patterns to prevent overlap.")
		fmt.Println()
	}

	if !hasIssues {
		fmt.Println("✅ All module file ownership checks passed!")
		fmt.Println()
	}
}

func printModuleFilesUsage() {
	fmt.Println("Validate module file ownership")
	fmt.Println()
	fmt.Println("Usage: r2r validate module-files")
	fmt.Println()
	fmt.Println("Checks:")
	fmt.Println("  - No files belong to the 'unordered' catch-all module")
	fmt.Println("  - Each file belongs to exactly one module")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Validate module file ownership")
	fmt.Println("  r2r validate module-files")
}
