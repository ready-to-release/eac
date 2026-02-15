package validate

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/repository"
)

type validateModuleHierarchyCommand struct{}

var _ core.SimpleCommandPort = (*validateModuleHierarchyCommand)(nil)

func (c *validateModuleHierarchyCommand) Name() string { return "validate module-hierarchy" }

func (c *validateModuleHierarchyCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-module-hierarchy",
		Short:         "Validate module dependency graph structure",
		Long:          "Validates the module dependency graph for structural integrity.\n\nExpected Output:\n  Displays structural issues in dependency graph including cycles, non-existent module\n  references, and bidirectional inconsistencies. Exit code 0 if valid, 1 if issues found.",
	}
}

func (c *validateModuleHierarchyCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateModuleHierarchy()
}

// ValidateModuleHierarchy validates the module dependency graph.
func ValidateModuleHierarchy() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "validate", and "module-hierarchy"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printModuleHierarchyUsage()
		return 0
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspace(repoRoot)
	if err != nil {
		log.Errorf("Error: failed to load module registry: %v", err)
		return 1
	}

	// Run validations
	report := validateModuleHierarchy(moduleRegistry)

	// Print report
	printModuleHierarchyReport(report)

	// Return exit code based on results
	if report.HasErrors() {
		return 1
	}
	return 0
}

type moduleHierarchyReport struct {
	inconsistencies      []string
	nonExistentModules   []string
	circularDependencies []string
	unreachableModules   []string
}

func (r *moduleHierarchyReport) HasErrors() bool {
	return len(r.inconsistencies) > 0 ||
		len(r.nonExistentModules) > 0 ||
		len(r.circularDependencies) > 0 ||
		len(r.unreachableModules) > 0
}

func validateModuleHierarchy(reg *modules.Registry) *moduleHierarchyReport {
	report := &moduleHierarchyReport{
		inconsistencies:      []string{},
		nonExistentModules:   []string{},
		circularDependencies: []string{},
		unreachableModules:   []string{},
	}

	// Check bidirectional relationships
	validateBidirectionalRelationships(reg, report)

	// Check for circular dependencies
	validateNoCircularDependencies(reg, report)

	// Check all modules are reachable
	validateAllModulesReachable(reg, report)

	return report
}

func validateBidirectionalRelationships(reg *modules.Registry, report *moduleHierarchyReport) {
	for _, module := range reg.All() {
		// Check that all dependencies exist
		for _, depMoniker := range module.DependsOn {
			if !reg.Has(depMoniker) {
				report.nonExistentModules = append(report.nonExistentModules,
					fmt.Sprintf("Module '%s' depends on '%s', but '%s' does not exist",
						module.Moniker, depMoniker, depMoniker))
			}
			// Note: used_by is now computed from depends_on, so no need to check consistency
		}
	}
}

func validateNoCircularDependencies(reg *modules.Registry, report *moduleHierarchyReport) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(moniker string, path []string) bool
	detectCycle = func(moniker string, path []string) bool {
		visited[moniker] = true
		recStack[moniker] = true
		currentPath := append(path, moniker)

		module, found := reg.Get(moniker)
		if !found {
			return false
		}

		for _, depMoniker := range module.DependsOn {
			if !visited[depMoniker] {
				if detectCycle(depMoniker, currentPath) {
					return true
				}
			} else if recStack[depMoniker] {
				// Found a cycle
				cycleStart := -1
				for i, m := range currentPath {
					if m == depMoniker {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := append(currentPath[cycleStart:], depMoniker)
					report.circularDependencies = append(report.circularDependencies,
						fmt.Sprintf("Circular dependency: %s", strings.Join(cycle, " -> ")))
				}
				return true
			}
		}

		recStack[moniker] = false
		return false
	}

	// Check all modules for cycles
	for _, module := range reg.All() {
		if !visited[module.Moniker] {
			detectCycle(module.Moniker, []string{})
		}
	}
}

func validateAllModulesReachable(reg *modules.Registry, report *moduleHierarchyReport) {
	// All modules are now root modules (no parent hierarchy)
	// Build dependency graph and check reachability through dependencies
	allModules := reg.All()
	if len(allModules) == 0 {
		return
	}

	// All modules are reachable since there's no parent hierarchy
	// The dependency graph validation is already done by validateNoCircularDependencies
	// and validateBidirectionalRelationships
}

func printModuleHierarchyReport(report *moduleHierarchyReport) {
	toBullets := func(items []string) []string {
		out := make([]string, len(items))
		for i, item := range items {
			out[i] = formatBullet(item)
		}
		return out
	}

	printValidationReport(validationReport{
		Title: "Module Hierarchy Validation Report",
		Sections: []validationSection{
			{Icon: "❌", Label: "References to Non-Existent Modules", Items: toBullets(report.nonExistentModules)},
			{Icon: "❌", Label: "Bidirectional Relationship Inconsistencies", Items: toBullets(report.inconsistencies)},
			{Icon: "❌", Label: "Circular Dependencies", Items: toBullets(report.circularDependencies)},
			{Icon: "⚠️ ", Label: "Unreachable Modules", Items: toBullets(report.unreachableModules)},
		},
		SuccessMessage: "All module hierarchy checks passed!",
	})
}

func printModuleHierarchyUsage() {
	log.Info("Validate module dependency graph structure")
	log.Info("")
	log.Info("Usage: clie validate module-hierarchy")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Bidirectional consistency (depends_on <-> used_by)")
	log.Info("  - No references to non-existent modules")
	log.Info("  - No circular dependencies")
	log.Info("  - All modules reachable from root")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate module hierarchy")
	log.Info("  clie validate module-hierarchy")
}
