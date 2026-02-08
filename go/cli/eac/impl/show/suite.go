// Command: show suite
// Short: Display test suite details with all tests
// Long: The show suite command displays detailed information about a specific test suite.
// Long: Shows selection criteria, production tests with their metadata, and summary statistics.
// Long:
// Long: Expected Output:
// Long: - Suite header with name, moniker, description, and test counts
// Long: - Selection criteria section showing AnyOf/RequireAll/Exclude tag rules
// Long: - Formatted table of production tests with columns: #, Moniker, Test Name, Type, Module, Level, Verification, System Deps
// Long: - Statistics section with counts by type, by module, and aggregated dependencies
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	contractsreports "github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
	testing "github.com/ready-to-release/eac/go/core/testing"
)

// ShowSuite displays detailed information about a test suite in markdown table format
//
// Command: show suite <suite-moniker>
// Example: show suite unit
//
// Output format:
// - Suite header with metadata
// - Markdown table with one row per test
// - Columns: Test Name, Type, Module, Tags.
func ShowSuite() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

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
		fmt.Fprintln(os.Stderr, "Error: suite moniker required")
		fmt.Fprintln(os.Stderr, "Usage: show suite <suite-moniker>")
		fmt.Fprintln(os.Stderr, "Available suites:")
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "Available suites:")
		for _, moniker := range testing.ListSuites() {
			fmt.Fprintf(os.Stderr, "  - %s\n", moniker)
		}
		return 1
	}

	// Load module registry
	moduleReport, err := contractsreports.GetModuleContracts(repoRoot)
	var moduleRegistry *modules.Registry
	if err == nil {
		moduleRegistry = moduleReport.Registry
	}

	// Build file-to-module mapping
	fileModuleMap, err := testdata.BuildFileModuleMap(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load file-module mapping: %v\n", err)
		fileModuleMap = make(map[string]string)
	}

	// Generate suite report using canonical data generator
	report, err := testing.GenerateSuiteReport(suite, repoRoot, moduleRegistry, fileModuleMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: generating suite report: %v\n", err)
		return 1
	}

	// Display validation errors if any
	if len(report.ValidationErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  WARNING: %d tests have validation errors:\n", len(report.ValidationErrors))
		if len(report.FrameworkTests) > 0 {
			fmt.Fprintf(os.Stderr, "          (%d framework tests excluded from validation)\n", len(report.FrameworkTests))
		}
		fmt.Fprintln(os.Stderr, "")
		for testName, errors := range report.ValidationErrors {
			fmt.Fprintf(os.Stderr, "  - %s:\n", testName)
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "    • %s\n", err)
			}
		}
		fmt.Fprintln(os.Stderr, "")
	} else if len(report.FrameworkTests) > 0 {
		fmt.Fprintf(os.Stderr, "\n✓ All tests pass validation (%d framework tests excluded from display)\n", len(report.FrameworkTests))
	}

	// Display suite information
	fmt.Printf("# Test Suite: %s\n\n", report.SuiteName)
	fmt.Printf("**Moniker**: `%s`  \n", report.SuiteMoniker)
	fmt.Printf("**Description**: %s  \n", report.Description)
	fmt.Printf("**Production Tests**: %d  \n", len(report.ProductionTests))
	fmt.Printf("**Framework Tests**: %d (excluded from display)  \n", len(report.FrameworkTests))
	fmt.Printf("**Total Discovered**: %d  \n", report.TotalDiscovered)
	fmt.Println("")

	// Display selection criteria
	fmt.Println("## Selection Criteria")
	fmt.Println("")
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
		fmt.Println("")
	}

	// Display tests in markdown table using TableBuilder
	fmt.Println("## Production Tests")
	fmt.Println("")

	tb := render.NewTableBuilder().
		WithHeaders("#", "Moniker", "Test Name", "Type", "Module", "Level", "Verification", "System Deps")

	for i, entry := range report.ProductionTests {
		// Format tag columns
		levelStr := strings.Join(entry.Level, ", ")
		verificationStr := strings.Join(entry.Verification, ", ")
		systemDepsStr := strings.Join(entry.SystemDeps, ", ")

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			entry.Moniker,
			entry.TestName,
			entry.Type,
			entry.Module,
			levelStr,
			verificationStr,
			systemDepsStr,
		)
	}

	fmt.Println(tb.Build())
	fmt.Println("")

	// Display summary statistics
	fmt.Println("## Statistics")
	fmt.Println("")

	// Count by type
	typeCounts := make(map[string]int)
	for _, entry := range report.ProductionTests {
		typeCounts[entry.Type]++
	}

	fmt.Println("**By Type**:")
	for testType, count := range typeCounts {
		fmt.Printf("  - %s: %d\n", testType, count)
	}
	fmt.Println("")

	// Count by module
	moduleCounts := make(map[string]int)
	for _, entry := range report.ProductionTests {
		moduleCounts[entry.Module]++
	}

	fmt.Println("**By Module**:")
	for module, count := range moduleCounts {
		fmt.Printf("  - %s: %d\n", module, count)
	}
	fmt.Println("")

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
		fmt.Println("**Dependencies**:")
		if len(systemDeps) > 0 {
			fmt.Printf("  - System: %s\n", strings.Join(systemDeps, ", "))
		}
		if len(moduleDeps) > 0 {
			fmt.Printf("  - Module: %s\n", strings.Join(moduleDeps, ", "))
		}
		fmt.Println("")
	}

	return 0
}
