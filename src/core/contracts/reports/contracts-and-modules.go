package reports

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// ModuleContractReport contains information about loaded module contracts
type ModuleContractReport struct {
	TotalModules int
	Modules      []*modules.ModuleContract
	Registry     *modules.Registry
}

// GetModuleContracts loads and reports on all module contracts
//
// Parameters:
//   - workspaceRoot: Repository root (if empty, will be detected automatically)
//   - version: Module contract version (e.g., "0.1.0")
//
// Returns:
//   - ModuleContractReport containing all loaded contracts and metadata
//   - Error if module loading fails
//
// Example:
//
//	report, err := reports.GetModuleContracts("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total modules: %d\n", report.TotalModules)
func GetModuleContracts(workspaceRoot string) (*ModuleContractReport, error) {
	// Load all module contracts
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Get all modules sorted by moniker
	allModules := registry.All()
	sort.Slice(allModules, func(i, j int) bool {
		return allModules[i].Moniker < allModules[j].Moniker
	})

	report := &ModuleContractReport{
		TotalModules: len(allModules),
		Modules:      allModules,
		Registry:     registry,
	}

	return report, nil
}

// FormatReport returns a formatted string representation of the module contracts
func (r *ModuleContractReport) FormatReport() string {
	var sb strings.Builder

	sb.WriteString("=== Module Contracts Report ===\n\n")
	sb.WriteString(fmt.Sprintf("✅ Loaded %d module contracts (version: %s)\n\n", r.TotalModules, r.Registry.Version()))

	// List all modules
	sb.WriteString("=== Modules ===\n")
	for i, module := range r.Modules {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, module.Moniker))
		sb.WriteString(fmt.Sprintf("   Name: %s\n", module.Name))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", module.Type))
		sb.WriteString(fmt.Sprintf("   Root: %s\n", module.Files.Root))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", module.Description))

		// Source patterns
		if len(module.Files.Source) > 0 {
			sb.WriteString("   Source patterns:\n")
			for _, pattern := range module.Files.Source {
				sb.WriteString(fmt.Sprintf("     - %s\n", pattern))
			}
		}

		// Dependencies
		if len(module.DependsOn) > 0 {
			sb.WriteString(fmt.Sprintf("   Depends on: %v\n", module.DependsOn))
		}

		// Used by (computed from registry)
		usedBy := r.Registry.GetUsedBy(module.Moniker)
		if len(usedBy) > 0 {
			sb.WriteString(fmt.Sprintf("   Used by: %v\n", usedBy))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatCompact returns a compact one-line-per-module format
func (r *ModuleContractReport) FormatCompact() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Module Contracts (%d modules) ===\n\n", r.TotalModules))

	for _, module := range r.Modules {
		sb.WriteString(fmt.Sprintf("%-30s %-20s %s\n", module.Moniker, module.Type, module.Files.Root))
	}

	return sb.String()
}

// GetModuleByMoniker returns a specific module contract by moniker
func (r *ModuleContractReport) GetModuleByMoniker(moniker string) (*modules.ModuleContract, bool) {
	return r.Registry.Get(moniker)
}

// GetModulesByType returns all modules of a specific type
func (r *ModuleContractReport) GetModulesByType(moduleType string) []*modules.ModuleContract {
	return r.Registry.FilterByType(moduleType)
}

// GetModulesByRoot returns modules with a specific root path
func (r *ModuleContractReport) GetModulesByRoot(root string) []*modules.ModuleContract {
	var result []*modules.ModuleContract
	for _, module := range r.Modules {
		if module.Files.Root == root {
			result = append(result, module)
		}
	}
	return result
}

// GetDependencyGraph returns the dependency relationships
func (r *ModuleContractReport) GetDependencyGraph() map[string][]string {
	return r.Registry.GetDependencyGraph()
}

// GetReverseDependencyGraph returns the reverse dependency relationships
func (r *ModuleContractReport) GetReverseDependencyGraph() map[string][]string {
	return r.Registry.GetReverseDependencyGraph()
}

// GetModulesWithPattern returns modules that use a specific glob pattern
func (r *ModuleContractReport) GetModulesWithPattern(pattern string) []*modules.ModuleContract {
	var result []*modules.ModuleContract
	for _, module := range r.Modules {
		// Check all pattern categories
		allPatterns := append(module.Files.Source, module.Files.Config...)
		allPatterns = append(allPatterns, module.Files.Assets...)
		allPatterns = append(allPatterns, module.Files.Tests...)
		for _, p := range allPatterns {
			if p == pattern {
				result = append(result, module)
				break
			}
		}
	}
	return result
}

// PrintSummary prints a concise summary of the loaded contracts
func (r *ModuleContractReport) PrintSummary() {
	fmt.Println("=== Module Contracts Summary ===")
	fmt.Printf("Total modules:    %d\n", r.TotalModules)
	fmt.Printf("Version:          %s\n", r.Registry.Version())

	// Count by type
	typeCount := make(map[string]int)
	for _, module := range r.Modules {
		typeCount[module.Type]++
	}

	fmt.Println("\nBy type:")
	for typ, count := range typeCount {
		fmt.Printf("  %-20s %d modules\n", typ, count)
	}
}
