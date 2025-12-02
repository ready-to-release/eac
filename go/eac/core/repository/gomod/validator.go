package gomod

import (
	"fmt"
	"sort"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// Validator validates go.mod dependencies against module contracts
type Validator struct {
	graph    *DependencyGraph
	registry *modules.Registry
}

// NewValidator creates a new validator
func NewValidator(graph *DependencyGraph, registry *modules.Registry) *Validator {
	return &Validator{
		graph:    graph,
		registry: registry,
	}
}

// Validate compares the dependency graph against module contracts
func (v *Validator) Validate() *ValidationReport {
	report := &ValidationReport{
		Discrepancies: []Discrepancy{},
		Summary: ValidationSummary{
			TotalModules:        0,
			Matching:            0,
			WithDiscrepancies:   0,
			ModulesWithoutGoMod: 0,
		},
	}

	// Get all modules from registry
	allModules := v.registry.All()
	goModModules := make(map[string]bool)

	// Track which modules have go.mod files
	for moniker := range v.graph.Modules {
		goModModules[moniker] = true
	}

	// Validate each module that has a go.mod file
	for _, module := range allModules {
		moniker := module.Moniker

		// Check if module has a go.mod
		if !goModModules[moniker] {
			continue // Skip modules without go.mod files
		}

		report.Summary.TotalModules++

		// Get contract dependencies
		contractDeps := module.DependsOn
		sort.Strings(contractDeps)

		// Get actual dependencies from go.mod
		actualDeps := v.graph.GetDependencies(moniker)
		sort.Strings(actualDeps)

		// Validate that all actual dependencies are valid modules (exist in registry)
		var invalidDeps []string
		for _, dep := range actualDeps {
			if !v.registry.Has(dep) {
				invalidDeps = append(invalidDeps, dep)
			}
		}

		// Create discrepancy record
		discrepancy := Discrepancy{
			Moniker:              moniker,
			ContractDependencies: contractDeps,
			ActualDependencies:   actualDeps,
			Invalid:              invalidDeps,
		}

		// Only flag as discrepancy if there are invalid dependencies
		if len(invalidDeps) == 0 {
			discrepancy.Status = "✅ VALID"
			report.Summary.Matching++
		} else {
			discrepancy.Status = "❌ INVALID"
			report.Summary.WithDiscrepancies++
		}

		report.Discrepancies = append(report.Discrepancies, discrepancy)
	}

	// Count modules without go.mod
	for _, module := range allModules {
		if !goModModules[module.Moniker] {
			report.Summary.ModulesWithoutGoMod++
		}
	}

	return report
}

// compareDependencies compares two sorted dependency lists
// Returns: (missing from actual, extra in actual)
func compareDependencies(contract, actual []string) ([]string, []string) {
	contractSet := make(map[string]bool)
	actualSet := make(map[string]bool)

	for _, dep := range contract {
		contractSet[dep] = true
	}

	for _, dep := range actual {
		actualSet[dep] = true
	}

	// Find missing: in contract but not in actual
	var missing []string
	for dep := range contractSet {
		if !actualSet[dep] {
			missing = append(missing, dep)
		}
	}
	sort.Strings(missing)

	// Find extra: in actual but not in contract
	var extra []string
	for dep := range actualSet {
		if !contractSet[dep] {
			extra = append(extra, dep)
		}
	}
	sort.Strings(extra)

	return missing, extra
}

// FormatReport formats the validation report as a human-readable string
func (v *Validator) FormatReport(report *ValidationReport) string {
	var output string

	output += "=== Module Dependency Validation Report ===\n\n"

	// Summary
	output += fmt.Sprintf("Summary:\n")
	output += fmt.Sprintf("  Total Modules with go.mod: %d\n", report.Summary.TotalModules)
	output += fmt.Sprintf("  Matching: %d\n", report.Summary.Matching)
	output += fmt.Sprintf("  With Discrepancies: %d\n", report.Summary.WithDiscrepancies)
	output += fmt.Sprintf("  Modules without go.mod: %d\n", report.Summary.ModulesWithoutGoMod)
	output += "\n"

	// Discrepancies
	if len(report.Discrepancies) == 0 {
		output += "No modules with go.mod files found.\n"
		return output
	}

	output += "Module Details:\n\n"

	for _, disc := range report.Discrepancies {
		output += fmt.Sprintf("%s %s\n", disc.Status, disc.Moniker)

		if len(disc.ActualDependencies) > 0 {
			output += fmt.Sprintf("  Dependencies: %v\n", disc.ActualDependencies)
		} else {
			output += "  Dependencies: (none)\n"
		}

		if len(disc.Invalid) > 0 {
			output += fmt.Sprintf("  ❌ Invalid (not registered modules): %v\n", disc.Invalid)
		}

		output += "\n"
	}

	// Final verdict
	if report.Summary.WithDiscrepancies == 0 {
		output += "✅ All module dependencies are valid!\n"
	} else {
		output += fmt.Sprintf("❌ %d module(s) have invalid dependencies.\n", report.Summary.WithDiscrepancies)
	}

	return output
}

// ValidateAndReport performs validation and returns a formatted report
func ValidateAndReport(rootPath string, registry *modules.Registry, baseModulePath string, excludeDirs []string) (string, error) {
	// Build dependency graph from go.mod files
	graph, err := BuildFromDirectory(rootPath, registry, baseModulePath, excludeDirs)
	if err != nil {
		return "", fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Create validator
	validator := NewValidator(graph, registry)

	// Run validation
	report := validator.Validate()

	// Format and return report
	return validator.FormatReport(report), nil
}

// GetDiscrepanciesByStatus returns discrepancies filtered by status
func (r *ValidationReport) GetDiscrepanciesByStatus(status string) []Discrepancy {
	var filtered []Discrepancy
	for _, disc := range r.Discrepancies {
		if disc.Status == status {
			filtered = append(filtered, disc)
		}
	}
	return filtered
}

// HasDiscrepancies returns true if any discrepancies were found
func (r *ValidationReport) HasDiscrepancies() bool {
	return r.Summary.WithDiscrepancies > 0
}

// AllMatch returns true if all modules match their contracts
func (r *ValidationReport) AllMatch() bool {
	return r.Summary.WithDiscrepancies == 0 && r.Summary.TotalModules > 0
}
