// Command: show suite
// Description: Display detailed information about a test suite
// HasSideEffects: false
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	contractsreports "github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
	testing "github.com/ready-to-release/eac/src/core/testing"
)

func init() {
	registry.Register(ShowSuite)
}

// ShowSuite displays detailed information about a test suite in markdown table format
//
// Command: show suite <suite-moniker>
// Example: show suite commit
//
// Output format:
// - Suite header with metadata
// - Markdown table with one row per test
// - Columns: Test Name, Type, Module, Tags
func ShowSuite() int {
	// Parse arguments - expect suite moniker after "show suite"
	args := os.Args[1:]

	// Find where "show suite" ends
	suiteIdx := -1
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "show" && args[i+1] == "suite" {
			suiteIdx = i + 2
			break
		}
	}

	if suiteIdx == -1 || suiteIdx >= len(args) {
		fmt.Fprintf(os.Stderr, "Error: suite moniker required\n\n")
		fmt.Fprintf(os.Stderr, "Usage: show suite <suite-moniker>\n\n")
		fmt.Fprintf(os.Stderr, "Available suites:\n")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
		}
		return 1
	}

	suiteMoniker := args[suiteIdx]

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a git repository: %v\n", err)
		return 1
	}

	// Get suite
	suite, err := testing.GetSuite(suiteMoniker)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		fmt.Fprintf(os.Stderr, "Available suites:\n")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
		}
		return 1
	}

	// Load module registry
	moduleReport, err := contractsreports.GetModuleContracts(repoRoot, "0.1.0")
	var moduleRegistry *modules.Registry
	if err == nil {
		moduleRegistry = moduleReport.Registry
	}

	// Build file-to-module mapping
	fileModuleMap, err := buildFileModuleMap(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load file-module mapping: %v\n", err)
		fileModuleMap = make(map[string]string)
	}

	// Generate suite report using canonical data generator
	report, err := testing.GenerateSuiteReport(suite, repoRoot, moduleRegistry, fileModuleMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating suite report: %v\n", err)
		return 1
	}

	// Display validation errors if any
	if len(report.ValidationErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: %d tests have validation errors:\n", len(report.ValidationErrors))
		if len(report.FrameworkTests) > 0 {
			fmt.Fprintf(os.Stderr, "          (%d framework tests excluded from validation)\n", len(report.FrameworkTests))
		}
		fmt.Fprintf(os.Stderr, "\n")
		for testName, errors := range report.ValidationErrors {
			fmt.Fprintf(os.Stderr, "  - %s:\n", testName)
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "    • %s\n", err)
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
	} else if len(report.FrameworkTests) > 0 {
		fmt.Fprintf(os.Stderr, "\n✓ All tests pass validation (%d framework tests excluded from display)\n\n", len(report.FrameworkTests))
	}

	// Display suite information
	fmt.Printf("# Test Suite: %s\n\n", report.SuiteName)
	fmt.Printf("**Moniker**: `%s`  \n", report.SuiteMoniker)
	fmt.Printf("**Description**: %s  \n", report.Description)
	fmt.Printf("**Production Tests**: %d  \n", len(report.ProductionTests))
	fmt.Printf("**Framework Tests**: %d (excluded from display)  \n", len(report.FrameworkTests))
	fmt.Printf("**Total Discovered**: %d  \n", report.TotalDiscovered)
	fmt.Printf("\n")

	// Display selection criteria
	fmt.Printf("## Selection Criteria\n\n")
	for i, selector := range report.Selectors {
		fmt.Printf("**Selector %d**:\n", i+1)
		if len(selector.AnyOfTags) > 0 {
			fmt.Printf("  - **AnyOf**: %s\n", strings.Join(selector.AnyOfTags, ", "))
		}
		if len(selector.RequireTags) > 0 {
			fmt.Printf("  - **RequireAll**: %s\n", strings.Join(selector.RequireTags, ", "))
		}
		if len(selector.ExcludeTags) > 0 {
			fmt.Printf("  - **Exclude**: %s\n", strings.Join(selector.ExcludeTags, ", "))
		}
		fmt.Printf("\n")
	}

	// Display tests in markdown table using TableBuilder
	fmt.Printf("## Production Tests\n\n")

	tb := render.NewTableBuilder().
		WithHeaders("#", "Test Name", "Type", "Module", "Level", "Verification", "System Deps", "Module Deps", "Module Type")

	for i, entry := range report.ProductionTests {
		// Format tag columns
		levelStr := strings.Join(entry.Level, ", ")
		verificationStr := strings.Join(entry.Verification, ", ")
		systemDepsStr := strings.Join(entry.SystemDeps, ", ")
		moduleDepsStr := strings.Join(entry.ModuleDeps, ", ")
		moduleTypesStr := strings.Join(entry.ModuleTypes, ", ")

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			entry.TestName,
			entry.Type,
			entry.Module,
			levelStr,
			verificationStr,
			systemDepsStr,
			moduleDepsStr,
			moduleTypesStr,
		)
	}

	fmt.Println(tb.Build())
	fmt.Printf("\n")

	// Display summary statistics
	fmt.Printf("## Statistics\n\n")

	// Count by type
	typeCounts := make(map[string]int)
	for _, entry := range report.ProductionTests {
		typeCounts[entry.Type]++
	}

	fmt.Printf("**By Type**:\n")
	for testType, count := range typeCounts {
		fmt.Printf("  - %s: %d\n", testType, count)
	}
	fmt.Printf("\n")

	// Count by module
	moduleCounts := make(map[string]int)
	for _, entry := range report.ProductionTests {
		moduleCounts[entry.Module]++
	}

	fmt.Printf("**By Module**:\n")
	for module, count := range moduleCounts {
		fmt.Printf("  - %s: %d\n", module, count)
	}
	fmt.Printf("\n")

	// Extract and display dependencies
	allSystemDeps := make(map[string]bool)
	allModuleDeps := make(map[string]bool)
	for _, entry := range report.ProductionTests {
		for _, dep := range entry.SystemDeps {
			allSystemDeps[dep] = true
		}
		for _, dep := range entry.ModuleDeps {
			allModuleDeps[dep] = true
		}
	}

	systemDeps := []string{}
	for dep := range allSystemDeps {
		systemDeps = append(systemDeps, dep)
	}
	moduleDeps := []string{}
	for dep := range allModuleDeps {
		moduleDeps = append(moduleDeps, dep)
	}

	if len(systemDeps) > 0 || len(moduleDeps) > 0 {
		fmt.Printf("**Dependencies**:\n")
		if len(systemDeps) > 0 {
			fmt.Printf("  - System: %s\n", strings.Join(systemDeps, ", "))
		}
		if len(moduleDeps) > 0 {
			fmt.Printf("  - Module: %s\n", strings.Join(moduleDeps, ", "))
		}
		fmt.Printf("\n")
	}

	return 0
}

// buildFileModuleMap creates a map of file paths to module names
func buildFileModuleMap(repoRoot string) (map[string]string, error) {
	fileMap := make(map[string]string)

	// Get all files with module ownership
	files, err := repository.GetRepositoryFilesWithModules(true, false, false, repoRoot, "0.1.0")
	if err != nil {
		return nil, err
	}

	// Build the mapping
	for _, file := range files {
		if len(file.Modules) > 0 {
			fileMap[file.Name] = file.Modules[0] // Use first module
		}
	}

	return fileMap, nil
}
