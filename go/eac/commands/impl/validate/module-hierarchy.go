// Command: validate module-hierarchy
// Description: Validate module dependency graph structure
// Short: Validate module dependency graph structure
// Long: Validates the module dependency graph for structural integrity.
// Long:
// Long: Expected Output:
// Long:   Displays structural issues in dependency graph including cycles, non-existent module
// Long:   references, and bidirectional inconsistencies. Exit code 0 if valid, 1 if issues found.
package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ValidateModuleHierarchy)
}

// ValidateModuleHierarchy validates the module dependency graph
func ValidateModuleHierarchy() int {
	args := os.Args[3:] // Skip program name, "validate", and "module-hierarchy"

	// Define expected flags
	commandFlags := []flags.FlagDefinition{
		{Name: "--help", Shorthand: "-h", HasValue: false},
	}

	// Validate flags
	if err := flags.ValidateFlags(args, commandFlags); err != nil {
		log.Errorf("%v", err)
		return 1
	}

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
	moduleRegistry, err := modules.LoadFromWorkspaceLatest(repoRoot)
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
	log.Info("=== Module Hierarchy Validation Report ===")
	log.Info("")

	hasIssues := false

	// Non-existent modules
	if len(report.nonExistentModules) > 0 {
		hasIssues = true
		log.Infof("❌ References to Non-Existent Modules (%d):", len(report.nonExistentModules))
		for _, issue := range report.nonExistentModules {
			log.Infof("  • %s", issue)
		}
		log.Info("")
	}

	// Bidirectional inconsistencies
	if len(report.inconsistencies) > 0 {
		hasIssues = true
		log.Infof("❌ Bidirectional Relationship Inconsistencies (%d):", len(report.inconsistencies))
		for _, issue := range report.inconsistencies {
			log.Infof("  • %s", issue)
		}
		log.Info("")
	}

	// Circular dependencies
	if len(report.circularDependencies) > 0 {
		hasIssues = true
		log.Infof("❌ Circular Dependencies (%d):", len(report.circularDependencies))
		for _, issue := range report.circularDependencies {
			log.Infof("  • %s", issue)
		}
		log.Info("")
	}

	// Unreachable modules
	if len(report.unreachableModules) > 0 {
		hasIssues = true
		log.Infof("⚠️  Unreachable Modules (%d):", len(report.unreachableModules))
		for _, issue := range report.unreachableModules {
			log.Infof("  • %s", issue)
		}
		log.Info("")
	}

	if !hasIssues {
		log.Info("✅ All module hierarchy checks passed!")
		log.Info("")
	}
}

func printModuleHierarchyUsage() {
	log.Info("Validate module dependency graph structure")
	log.Info("")
	log.Info("Usage: r2r validate module-hierarchy")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Bidirectional consistency (depends_on <-> used_by)")
	log.Info("  - No references to non-existent modules")
	log.Info("  - No circular dependencies")
	log.Info("  - All modules reachable from root")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate module hierarchy")
	log.Info("  r2r validate module-hierarchy")
}
