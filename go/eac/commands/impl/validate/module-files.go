// Command: validate module-files
// Description: Validate module file ownership
// Short: Validate module file ownership
// Long: Validates that all files have proper module ownership and no files are unordered.
// Long:
// Long: Expected Output:
// Long:   Displays files without proper module ownership (unordered or multi-module files).
// Long:   Shows file paths and claiming modules. Exit code 0 if all files properly owned, 1 if issues found.
package validate

import (
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/repository"
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
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	// Open git repository
	repo, err := git.Open(repoRoot)
	if err != nil {
		log.Errorf("Error: failed to open git repository: %v", err)
		return 1
	}

	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(repo, true, false, false)
	if err != nil {
		log.Errorf("Error: failed to get repository files with modules: %v", err)
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
	log.Info("=== Module File Ownership Validation Report ===")
	log.Info("")

	hasIssues := false

	// Unordered files
	if len(report.unorderedFiles) > 0 {
		hasIssues = true
		log.Infof("❌ Files in Unordered Module (%d):", len(report.unorderedFiles))
		log.Info("   These files should be claimed by a proper module:")
		for _, filePath := range report.unorderedFiles {
			log.Infof("  • %s", filePath)
		}
		log.Info("")
		log.Info("   Fix: Create or update module contracts to claim these files.")
		log.Info("")
	}

	// Multi-ownership files
	if len(report.multiOwnershipFiles) > 0 {
		hasIssues = true
		log.Infof("❌ Files with Multi-Module Ownership (%d):", len(report.multiOwnershipFiles))
		log.Info("   Each file should belong to exactly one module:")
		for filePath, modules := range report.multiOwnershipFiles {
			log.Infof("  • %s", filePath)
			log.Infof("    Claimed by: %s", strings.Join(modules, ", "))
		}
		log.Info("")
		log.Info("   Fix: Adjust module contract glob patterns to prevent overlap.")
		log.Info("")
	}

	if !hasIssues {
		log.Info("✅ All module file ownership checks passed!")
		log.Info("")
	}
}

func printModuleFilesUsage() {
	log.Info("Validate module file ownership")
	log.Info("")
	log.Info("Usage: r2r validate module-files")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - No files belong to the 'unordered' catch-all module")
	log.Info("  - Each file belongs to exactly one module")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate module file ownership")
	log.Info("  r2r validate module-files")
}
