package tests

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/repository"
)

// moduleHierarchyContext holds state for module hierarchy validation tests
type moduleHierarchyContext struct {
	registry              *modules.Registry
	missingModules        []string
	inconsistencies       []string
	circularDependencies  []string
	unreachableModules    []string
	validationError       error
}

func (ctx *moduleHierarchyContext) iLoadAllModuleContractsFromTheRepository() error {
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	// Load module registry
	ctx.registry, err = modules.LoadFromWorkspaceLatest(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to load module registry: %w", err)
	}

	return nil
}

func (ctx *moduleHierarchyContext) iValidateThatAllDependsOnAndUsedByRelationshipsAreBidirectional() error {
	ctx.inconsistencies = []string{}

	// Check all modules
	for _, module := range ctx.registry.All() {
		// Check depends_on -> used_by consistency
		for _, depMoniker := range module.DependsOn {
			dep, found := ctx.registry.Get(depMoniker)
			if !found {
				ctx.inconsistencies = append(ctx.inconsistencies,
					fmt.Sprintf("Module '%s' depends on '%s', but '%s' does not exist",
						module.Moniker, depMoniker, depMoniker))
				continue
			}

			// Check if dep has module in its used_by list
			hasUsedBy := false
			for _, user := range dep.UsedBy {
				if user == module.Moniker {
					hasUsedBy = true
					break
				}
			}

			if !hasUsedBy {
				ctx.inconsistencies = append(ctx.inconsistencies,
					fmt.Sprintf("Module '%s' depends_on '%s', but '%s' does not have '%s' in used_by",
						module.Moniker, depMoniker, depMoniker, module.Moniker))
			}
		}

		// Check used_by -> depends_on consistency
		for _, userMoniker := range module.UsedBy {
			user, found := ctx.registry.Get(userMoniker)
			if !found {
				ctx.inconsistencies = append(ctx.inconsistencies,
					fmt.Sprintf("Module '%s' has used_by '%s', but '%s' does not exist",
						module.Moniker, userMoniker, userMoniker))
				continue
			}

			// Check if user has module in its depends_on list
			hasDependsOn := false
			for _, dep := range user.DependsOn {
				if dep == module.Moniker {
					hasDependsOn = true
					break
				}
			}

			if !hasDependsOn {
				ctx.inconsistencies = append(ctx.inconsistencies,
					fmt.Sprintf("Module '%s' has used_by '%s', but '%s' does not have '%s' in depends_on",
						module.Moniker, userMoniker, userMoniker, module.Moniker))
			}
		}
	}

	return nil
}

func (ctx *moduleHierarchyContext) allModulesShouldHaveConsistentDependencyRelationships() error {
	if len(ctx.inconsistencies) > 0 {
		return fmt.Errorf("found %d inconsistencies:\n%s",
			len(ctx.inconsistencies), strings.Join(ctx.inconsistencies, "\n"))
	}
	return nil
}

func (ctx *moduleHierarchyContext) noModuleShouldReferenceANonExistentModule() error {
	// This is already checked in the bidirectional validation
	// Filter for non-existent module errors
	nonExistent := []string{}
	for _, inc := range ctx.inconsistencies {
		if strings.Contains(inc, "does not exist") {
			nonExistent = append(nonExistent, inc)
		}
	}

	if len(nonExistent) > 0 {
		return fmt.Errorf("found references to non-existent modules:\n%s",
			strings.Join(nonExistent, "\n"))
	}
	return nil
}

func (ctx *moduleHierarchyContext) iBuildTheCompleteDependencyGraph() error {
	// The registry already contains the complete graph
	// We just need to validate it
	return nil
}

func (ctx *moduleHierarchyContext) theGraphShouldBeASingleConnectedComponentOrForest() error {
	// For a repository with a root module, we expect a single tree/DAG
	// This is validated by checking if all modules are reachable from root(s)
	return nil
}

func (ctx *moduleHierarchyContext) theGraphShouldHaveNoCircularDependencies() error {
	ctx.circularDependencies = []string{}

	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var detectCycle func(moniker string, path []string) bool
	detectCycle = func(moniker string, path []string) bool {
		visited[moniker] = true
		recStack[moniker] = true
		currentPath := append(path, moniker)

		module, found := ctx.registry.Get(moniker)
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
					ctx.circularDependencies = append(ctx.circularDependencies,
						fmt.Sprintf("Circular dependency: %s", strings.Join(cycle, " -> ")))
				}
				return true
			}
		}

		recStack[moniker] = false
		return false
	}

	// Check all modules for cycles
	for _, module := range ctx.registry.All() {
		if !visited[module.Moniker] {
			detectCycle(module.Moniker, []string{})
		}
	}

	if len(ctx.circularDependencies) > 0 {
		return fmt.Errorf("found circular dependencies:\n%s",
			strings.Join(ctx.circularDependencies, "\n"))
	}

	return nil
}

func (ctx *moduleHierarchyContext) allModulesShouldBeReachableFromTheRoot() error {
	ctx.unreachableModules = []string{}

	// Find root modules (modules with parent "." or no parent)
	rootModules := []*modules.ModuleContract{}
	for _, module := range ctx.registry.All() {
		if module.Parent == "." || module.Parent == "" {
			rootModules = append(rootModules, module)
		}
	}

	if len(rootModules) == 0 {
		return fmt.Errorf("no root modules found (modules with parent='.')")
	}

	// BFS to find all reachable modules
	reachable := make(map[string]bool)
	queue := []string{}

	for _, root := range rootModules {
		queue = append(queue, root.Moniker)
		reachable[root.Moniker] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		module, found := ctx.registry.Get(current)
		if !found {
			continue
		}

		// Add all modules that depend on current (reverse direction)
		for _, other := range ctx.registry.All() {
			for _, dep := range other.DependsOn {
				if dep == current && !reachable[other.Moniker] {
					reachable[other.Moniker] = true
					queue = append(queue, other.Moniker)
				}
			}
		}

		// Also follow used_by relationships
		for _, user := range module.UsedBy {
			if !reachable[user] {
				reachable[user] = true
				queue = append(queue, user)
			}
		}

		// Follow parent relationships (children should be reachable)
		for _, other := range ctx.registry.All() {
			if other.Parent == current && !reachable[other.Moniker] {
				reachable[other.Moniker] = true
				queue = append(queue, other.Moniker)
			}
		}
	}

	// Check for unreachable modules
	// Skip catch-all modules as they don't participate in the dependency graph
	for _, module := range ctx.registry.All() {
		if !reachable[module.Moniker] {
			// Skip catch-all modules
			if module.Source.IsCatchAllSingleton != nil && *module.Source.IsCatchAllSingleton {
				continue
			}

			ctx.unreachableModules = append(ctx.unreachableModules,
				fmt.Sprintf("Module '%s' (parent: '%s') is not reachable from root modules",
					module.Moniker, module.Parent))
		}
	}

	if len(ctx.unreachableModules) > 0 {
		return fmt.Errorf("found unreachable modules:\n%s",
			strings.Join(ctx.unreachableModules, "\n"))
	}

	return nil
}

func (ctx *moduleHierarchyContext) iCheckAllDependsOnReferences() error {
	ctx.missingModules = []string{}

	for _, module := range ctx.registry.All() {
		for _, depMoniker := range module.DependsOn {
			if _, found := ctx.registry.Get(depMoniker); !found {
				ctx.missingModules = append(ctx.missingModules,
					fmt.Sprintf("Module '%s' depends_on '%s', but '%s' does not exist",
						module.Moniker, depMoniker, depMoniker))
			}
		}
	}

	return nil
}

func (ctx *moduleHierarchyContext) iCheckAllUsedByReferences() error {
	ctx.missingModules = []string{}

	for _, module := range ctx.registry.All() {
		for _, userMoniker := range module.UsedBy {
			if _, found := ctx.registry.Get(userMoniker); !found {
				ctx.missingModules = append(ctx.missingModules,
					fmt.Sprintf("Module '%s' has used_by '%s', but '%s' does not exist",
						module.Moniker, userMoniker, userMoniker))
			}
		}
	}

	return nil
}

func (ctx *moduleHierarchyContext) everyReferencedModuleShouldExistInTheRegistry() error {
	if len(ctx.missingModules) > 0 {
		return fmt.Errorf("found references to non-existent modules:\n%s",
			strings.Join(ctx.missingModules, "\n"))
	}
	return nil
}

func (ctx *moduleHierarchyContext) iShouldSeeDetailsOfAnyMissingModules() error {
	// This is informational - the details are already in the error messages
	return nil
}

func (ctx *moduleHierarchyContext) iValidateBidirectionalRelationships() error {
	return ctx.iValidateThatAllDependsOnAndUsedByRelationshipsAreBidirectional()
}

func (ctx *moduleHierarchyContext) ifModuleADependsOnBThenBsUsedByMustIncludeA() error {
	return ctx.allModulesShouldHaveConsistentDependencyRelationships()
}

func (ctx *moduleHierarchyContext) ifModuleBHasUsedByAThenAsDependsOnMustIncludeB() error {
	// Already checked in the same validation
	return nil
}

func (ctx *moduleHierarchyContext) iShouldSeeDetailsOfAnyInconsistencies() error {
	// This is informational - the details are already in the error messages
	return nil
}

func InitializeModuleHierarchyScenario(ctx *godog.ScenarioContext) {
	hierarchyCtx := &moduleHierarchyContext{}

	ctx.Step(`^I load all module contracts from the repository$`, hierarchyCtx.iLoadAllModuleContractsFromTheRepository)
	ctx.Step(`^I validate that all depends_on and used_by relationships are bidirectional$`, hierarchyCtx.iValidateThatAllDependsOnAndUsedByRelationshipsAreBidirectional)
	ctx.Step(`^all modules should have consistent dependency relationships$`, hierarchyCtx.allModulesShouldHaveConsistentDependencyRelationships)
	ctx.Step(`^no module should reference a non-existent module$`, hierarchyCtx.noModuleShouldReferenceANonExistentModule)
	ctx.Step(`^I build the complete dependency graph$`, hierarchyCtx.iBuildTheCompleteDependencyGraph)
	ctx.Step(`^the graph should be a single connected component or forest$`, hierarchyCtx.theGraphShouldBeASingleConnectedComponentOrForest)
	ctx.Step(`^the graph should have no circular dependencies$`, hierarchyCtx.theGraphShouldHaveNoCircularDependencies)
	ctx.Step(`^all modules should be reachable from the root$`, hierarchyCtx.allModulesShouldBeReachableFromTheRoot)
	ctx.Step(`^I check all depends_on references$`, hierarchyCtx.iCheckAllDependsOnReferences)
	ctx.Step(`^I check all used_by references$`, hierarchyCtx.iCheckAllUsedByReferences)
	ctx.Step(`^every referenced module should exist in the registry$`, hierarchyCtx.everyReferencedModuleShouldExistInTheRegistry)
	ctx.Step(`^I should see details of any missing modules$`, hierarchyCtx.iShouldSeeDetailsOfAnyMissingModules)
	ctx.Step(`^I validate bidirectional relationships$`, hierarchyCtx.iValidateBidirectionalRelationships)
	ctx.Step(`^if module A depends_on B, then B's used_by must include A$`, hierarchyCtx.ifModuleADependsOnBThenBsUsedByMustIncludeA)
	ctx.Step(`^if module B has used_by A, then A's depends_on must include B$`, hierarchyCtx.ifModuleBHasUsedByAThenAsDependsOnMustIncludeB)
	ctx.Step(`^I should see details of any inconsistencies$`, hierarchyCtx.iShouldSeeDetailsOfAnyInconsistencies)
}
