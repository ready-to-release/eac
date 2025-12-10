// Command: show suite
// Description: Display detailed information about a test suite
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	contractsreports "github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	testing "github.com/ready-to-release/eac/go/eac/core/testing"
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
		log.Errorf("suite moniker required\n")
		log.Errorf("Usage: show suite <suite-moniker>\n")
		log.Errorf("Available suites:")
		for _, moniker := range testing.ListSuites() {
			log.Errorf("  - %s", moniker)
		}
		return 1
	}

	suiteMoniker := args[suiteIdx]

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot(".")
	if err != nil {
		log.Errorf("not in a git repository: %v", err)
		return 1
	}

	// Get suite
	suite, err := testing.GetSuite(suiteMoniker)
	if err != nil {
		log.Errorf("%v\n", err)
		log.Errorf("Available suites:")
		for _, moniker := range testing.ListSuites() {
			log.Errorf("  - %s", moniker)
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
		log.Errorf("Warning: could not load file-module mapping: %v", err)
		fileModuleMap = make(map[string]string)
	}

	// Generate suite report using canonical data generator
	report, err := testing.GenerateSuiteReport(suite, repoRoot, moduleRegistry, fileModuleMap)
	if err != nil {
		log.Errorf("Error generating suite report: %v", err)
		return 1
	}

	// Display validation errors if any
	if len(report.ValidationErrors) > 0 {
		log.Errorf("\n⚠️  WARNING: %d tests have validation errors:", len(report.ValidationErrors))
		if len(report.FrameworkTests) > 0 {
			log.Errorf("          (%d framework tests excluded from validation)", len(report.FrameworkTests))
		}
		log.Errorf("")
		for testName, errors := range report.ValidationErrors {
			log.Errorf("  - %s:", testName)
			for _, err := range errors {
				log.Errorf("    • %s", err)
			}
		}
		log.Errorf("")
	} else if len(report.FrameworkTests) > 0 {
		log.Errorf("\n✓ All tests pass validation (%d framework tests excluded from display)\n", len(report.FrameworkTests))
	}

	// Display suite information
	log.Infof("# Test Suite: %s\n", report.SuiteName)
	log.Infof("**Moniker**: `%s`  ", report.SuiteMoniker)
	log.Infof("**Description**: %s  ", report.Description)
	log.Infof("**Production Tests**: %d  ", len(report.ProductionTests))
	log.Infof("**Framework Tests**: %d (excluded from display)  ", len(report.FrameworkTests))
	log.Infof("**Total Discovered**: %d  ", report.TotalDiscovered)
	log.Info("")

	// Display selection criteria
	log.Info("## Selection Criteria\n")
	for i, selector := range report.Selectors {
		log.Infof("**Selector %d**:", i+1)
		if len(selector.AnyOfTags) > 0 {
			log.Infof("  - **AnyOf**: %s", strings.Join(selector.AnyOfTags, ", "))
		}
		if len(selector.RequireTags) > 0 {
			log.Infof("  - **RequireAll**: %s", strings.Join(selector.RequireTags, ", "))
		}
		if len(selector.ExcludeTags) > 0 {
			log.Infof("  - **Exclude**: %s", strings.Join(selector.ExcludeTags, ", "))
		}
		log.Info("")
	}

	// Display tests in markdown table using TableBuilder
	log.Info("## Production Tests\n")

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

	log.Info(tb.Build())
	log.Info("")

	// Display summary statistics
	log.Info("## Statistics\n")

	// Count by type
	typeCounts := make(map[string]int)
	for _, entry := range report.ProductionTests {
		typeCounts[entry.Type]++
	}

	log.Info("**By Type**:")
	for testType, count := range typeCounts {
		log.Infof("  - %s: %d", testType, count)
	}
	log.Info("")

	// Count by module
	moduleCounts := make(map[string]int)
	for _, entry := range report.ProductionTests {
		moduleCounts[entry.Module]++
	}

	log.Info("**By Module**:")
	for module, count := range moduleCounts {
		log.Infof("  - %s: %d", module, count)
	}
	log.Info("")

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
		log.Info("**Dependencies**:")
		if len(systemDeps) > 0 {
			log.Infof("  - System: %s", strings.Join(systemDeps, ", "))
		}
		if len(moduleDeps) > 0 {
			log.Infof("  - Module: %s", strings.Join(moduleDeps, ", "))
		}
		log.Info("")
	}

	return 0
}
