package validate

import (
	"context"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type validateModuleFilesCommand struct{}

var _ core.SimpleCommandPort = (*validateModuleFilesCommand)(nil)

func (c *validateModuleFilesCommand) Name() string { return "validate module-files" }

func (c *validateModuleFilesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-module-files",
		Short:         "Validate module file ownership",
		Long:          "Validates that all files have proper module ownership and no files are unordered.\n\nExpected Output:\n  Displays files without proper module ownership (unordered or multi-module files).\n  Shows file paths and claiming modules. Exit code 0 if all files properly owned, 1 if issues found.",
	}
}

func (c *validateModuleFilesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateModuleFiles()
}

// ValidateModuleFiles validates file ownership in modules.
func ValidateModuleFiles() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "validate", and "module-files"

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
	gitMgr := git.NewManager(logging.C().Zap())
	repo, err := gitMgr.Open(repoRoot)
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
	// Build unordered file items
	unorderedItems := make([]string, len(report.unorderedFiles))
	for i, filePath := range report.unorderedFiles {
		unorderedItems[i] = formatBullet(filePath)
	}

	// Build multi-ownership items
	var multiItems []string
	for filePath, mods := range report.multiOwnershipFiles {
		multiItems = append(multiItems, formatBulletWithDetail(filePath, "Claimed by: "+strings.Join(mods, ", ")))
	}

	printValidationReport(validationReport{
		Title: "Module File Ownership Validation Report",
		Sections: []validationSection{
			{
				Icon:        "❌",
				Label:       "Files in Unordered Module",
				Items:       unorderedItems,
				Description: "These files should be claimed by a proper module:",
				FixHint:     "Fix: Create or update module contracts to claim these files.",
			},
			{
				Icon:        "❌",
				Label:       "Files with Multi-Module Ownership",
				Items:       multiItems,
				Description: "Each file should belong to exactly one module:",
				FixHint:     "Fix: Adjust module contract glob patterns to prevent overlap.",
			},
		},
		SuccessMessage: "All module file ownership checks passed!",
	})
}

func printModuleFilesUsage() {
	log.Info("Validate module file ownership")
	log.Info("")
	log.Info("Usage: clie validate module-files")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - No files belong to the 'unordered' catch-all module")
	log.Info("  - Each file belongs to exactly one module")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate module file ownership")
	log.Info("  clie validate module-files")
}
