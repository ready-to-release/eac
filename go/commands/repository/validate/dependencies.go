package validate

import (
	"context"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/repository/gomod"
)

type validateDependenciesCommand struct{}

var _ core.SimpleCommandPort = (*validateDependenciesCommand)(nil)

func (c *validateDependenciesCommand) Name() string { return "validate dependencies" }

func (c *validateDependenciesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-dependencies",
		Short:         "Validate module dependencies from go.mod files against contracts",
		Long: "Validate module dependencies from go.mod files against domain.\n\nThis command checks that actual dependencies in go.mod files match the dependencies\ndeclared in module domain. It helps ensure consistency between contract definitions\nand actual implementation dependencies.\n\nValidation failures indicate mismatches that should be resolved by either updating\nthe contract or fixing the go.mod file.",
		Notes: "Expected Output:\n  Displays mismatches between module contracts and actual go.mod dependencies.\n  Shows modules with extra or missing dependencies. Exit code 0 if all match, 1 if discrepancies found.",
		Examples: []string{
			"eac validate dependencies",
		},
	}
}

func (c *validateDependenciesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateDependencies()
}

func ValidateDependencies() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	baseModulePath := "github.com/ready-to-release/eac"

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		log.Errorf("Error loading contracts: %v", err)
		return 1
	}

	// Build dependency graph from go.mod files
	excludeDirs := []string{"out", "vendor", ".git", "node_modules", "tools", "templates"}
	graph, err := gomod.BuildFromDirectory(workspaceRoot, moduleRegistry, baseModulePath, excludeDirs)
	if err != nil {
		log.Errorf("Error building dependency graph: %v", err)
		return 1
	}

	// Validate
	validator := gomod.NewValidator(graph, moduleRegistry)
	report := validator.Validate()

	// Print report
	log.Info(validator.FormatReport(report))

	// Return appropriate exit code
	if report.HasDiscrepancies() {
		return 1
	}

	return 0
}
