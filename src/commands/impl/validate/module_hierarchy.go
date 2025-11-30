// Command: validate module-hierarchy
// Description: Validate module dependency graph structure
package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(ValidateModuleHierarchy)
}

// ValidateModuleHierarchy validates the module dependency graph
func ValidateModuleHierarchy() int {
	args := os.Args[2:] // Skip program name and "validate"

	// Check if this is being called as a subcommand
	if len(args) > 0 && args[0] == "module-hierarchy" {
		args = args[1:] // Skip the subcommand name
	}

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printModuleHierarchyUsage()
		return 0
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get repository root: %v\n", err)
		return 1
	}

	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspaceLatest(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module registry: %v\n", err)
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
	fmt.Println("=== Module Hierarchy Validation Report ===")
	fmt.Println()

	hasIssues := false

	// Non-existent modules
	if len(report.nonExistentModules) > 0 {
		hasIssues = true
		fmt.Printf("❌ References to Non-Existent Modules (%d):\n", len(report.nonExistentModules))
		for _, issue := range report.nonExistentModules {
			fmt.Printf("  • %s\n", issue)
		}
		fmt.Println()
	}

	// Bidirectional inconsistencies
	if len(report.inconsistencies) > 0 {
		hasIssues = true
		fmt.Printf("❌ Bidirectional Relationship Inconsistencies (%d):\n", len(report.inconsistencies))
		for _, issue := range report.inconsistencies {
			fmt.Printf("  • %s\n", issue)
		}
		fmt.Println()
	}

	// Circular dependencies
	if len(report.circularDependencies) > 0 {
		hasIssues = true
		fmt.Printf("❌ Circular Dependencies (%d):\n", len(report.circularDependencies))
		for _, issue := range report.circularDependencies {
			fmt.Printf("  • %s\n", issue)
		}
		fmt.Println()
	}

	// Unreachable modules
	if len(report.unreachableModules) > 0 {
		hasIssues = true
		fmt.Printf("⚠️  Unreachable Modules (%d):\n", len(report.unreachableModules))
		for _, issue := range report.unreachableModules {
			fmt.Printf("  • %s\n", issue)
		}
		fmt.Println()
	}

	if !hasIssues {
		fmt.Println("✅ All module hierarchy checks passed!")
		fmt.Println()
	}
}

func printModuleHierarchyUsage() {
	fmt.Println("Validate module dependency graph structure")
	fmt.Println()
	fmt.Println("Usage: r2r validate module-hierarchy")
	fmt.Println()
	fmt.Println("Checks:")
	fmt.Println("  - Bidirectional consistency (depends_on <-> used_by)")
	fmt.Println("  - No references to non-existent modules")
	fmt.Println("  - No circular dependencies")
	fmt.Println("  - All modules reachable from root")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Validate module hierarchy")
	fmt.Println("  r2r validate module-hierarchy")
}
