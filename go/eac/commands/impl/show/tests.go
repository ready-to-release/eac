// Command: show tests
// Description: Show all tests in the repository in a human-readable table
// Short: Display all test assertions with metadata
// Long: The show tests command displays all test assertions in the repository with their metadata.
// Long: Shows test names, modules, packages, tags (type, level, verification, dependencies), with indicators for inferred values.
// Long:
// Long: Expected Output:
// Long: - Table with test assertions showing: Module, Package, Assertion, Type, Level, Verify, Deps
// Long: - Module overview table with per-module statistics (assertions, packages, level breakdown, types)
// Long: - Summary sections showing counts by type and by level
// Long: - Legend indicating inferred tags with * and ~ symbols
package show

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

func init() {
	registry.Register(ShowTests)
}

func ShowTests() int {
	// Get repository root
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get current directory: %v\n", err)
		return 1
	}

	repoRoot, err := testdata.FindRepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get all tests with metadata and aggregations
	data, err := testdata.GetAllTests(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get tests: %v\n", err)
		return 1
	}

	// Display header
	fmt.Println("# Test Repository Overview")
	fmt.Println("")
	fmt.Printf("**Total Assertions**: %d  \n", data.TotalCount)

	// Build module overview table with OS filtering info
	fmt.Println("## Module Overview")
	fmt.Println("")
	moduleOverview := buildModuleOverview(data.Tests, data.OSFilteredCount, data.CurrentOS)
	fmt.Println(moduleOverview)
	fmt.Println("")

	// Build assertions table with inferred indicators
	fmt.Println("## All Assertions")
	fmt.Println("")
	fmt.Println("Legend: `*` = inferred tag, `~` = inferred from module type")
	fmt.Println("")

	tb := render.NewTableBuilder().
		WithHeaders("#", "Module", "Package", "Assertion", "Type", "Level", "Verify", "Deps")

	for i, entry := range data.Tests {
		levelStr := formatTagsWithInferred(entry.Level, entry.InferredTags)
		verifyStr := formatTagsWithInferred(entry.Verification, entry.InferredTags)
		// Check both InferredDeps (from module type) and InferredTags (from inference rules)
		allInferredDeps := append(entry.InferredDeps, entry.InferredTags...)
		depsStr := formatDepsWithInferred(entry.SystemDeps, allInferredDeps)

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			entry.Module,
			entry.Package,
			entry.TestName,
			entry.Type,
			levelStr,
			verifyStr,
			depsStr,
		)
	}

	fmt.Println(tb.Build())
	fmt.Println("")

	// Display summary by type
	fmt.Println("## Summary")
	fmt.Println("")
	fmt.Println("### By Type")
	fmt.Println("")
	for testType, count := range data.ByType {
		fmt.Printf("- **%s**: %d assertions\n", testType, count)
	}
	fmt.Println("")

	// Display summary by level (ordered)
	fmt.Println("### By Level")
	fmt.Println("")
	for _, level := range []string{"@L0", "@L1", "@L2", "@L3", "@L4"} {
		if count, ok := data.ByLevel[level]; ok {
			fmt.Printf("- **%s**: %d assertions\n", level, count)
		}
	}
	fmt.Println("")

	return 0
}

// buildModuleOverview creates a summary table with one row per module
func buildModuleOverview(tests []testing.SuiteTestEntry, osFilteredCount int, currentOS string) string {
	// Aggregate by module
	type moduleStats struct {
		Total    int
		ByLevel  map[string]int
		ByType   map[string]int
		Packages map[string]bool
	}
	modules := make(map[string]*moduleStats)

	// Totals for summary row
	totals := &moduleStats{
		ByLevel:  make(map[string]int),
		ByType:   make(map[string]int),
		Packages: make(map[string]bool),
	}

	for _, test := range tests {
		mod := test.Module
		if mod == "" {
			mod = "(unknown)"
		}
		if _, ok := modules[mod]; !ok {
			modules[mod] = &moduleStats{
				ByLevel:  make(map[string]int),
				ByType:   make(map[string]int),
				Packages: make(map[string]bool),
			}
		}
		stats := modules[mod]
		stats.Total++
		totals.Total++
		for _, level := range test.Level {
			stats.ByLevel[level]++
			totals.ByLevel[level]++
		}
		stats.ByType[test.Type]++
		totals.ByType[test.Type]++
		if test.Package != "" {
			stats.Packages[test.Package] = true
			// Use module:package for unique package count
			totals.Packages[mod+":"+test.Package] = true
		}
	}

	// Sort module names
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build table
	tb := render.NewTableBuilder().
		WithHeaders("Module", "Assertions", "Packages", "L0", "L1", "L2", "L3", "L4", "Types")

	for _, name := range names {
		stats := modules[name]

		// Format types
		types := make([]string, 0)
		for t := range stats.ByType {
			types = append(types, t)
		}
		sort.Strings(types)

		tb.AddRow(
			name,
			fmt.Sprintf("%d", stats.Total),
			fmt.Sprintf("%d", len(stats.Packages)),
			formatCount(stats.ByLevel["@L0"]),
			formatCount(stats.ByLevel["@L1"]),
			formatCount(stats.ByLevel["@L2"]),
			formatCount(stats.ByLevel["@L3"]),
			formatCount(stats.ByLevel["@L4"]),
			strings.Join(types, ", "),
		)
	}

	// Add separator row
	tb.AddRow("---", "---", "---", "---", "---", "---", "---", "---", "---")

	// Format total types
	totalTypes := make([]string, 0)
	for t := range totals.ByType {
		totalTypes = append(totalTypes, t)
	}
	sort.Strings(totalTypes)

	// Add summary row
	tb.AddRow(
		"**TOTAL**",
		fmt.Sprintf("**%d**", totals.Total),
		fmt.Sprintf("**%d**", len(totals.Packages)),
		formatCountBold(totals.ByLevel["@L0"]),
		formatCountBold(totals.ByLevel["@L1"]),
		formatCountBold(totals.ByLevel["@L2"]),
		formatCountBold(totals.ByLevel["@L3"]),
		formatCountBold(totals.ByLevel["@L4"]),
		strings.Join(totalTypes, ", "),
	)

	// Add OS filtered row if any tests were filtered
	if osFilteredCount > 0 {
		tb.AddRow(
			fmt.Sprintf("*filtered (%s)*", currentOS),
			fmt.Sprintf("*-%d*", osFilteredCount),
			"", "", "", "", "", "", "",
		)
	}

	return tb.Build()
}

// formatCount returns count as string, or "-" if zero
func formatCount(count int) string {
	if count == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", count)
}

// formatCountBold returns count as bold string, or "-" if zero
func formatCountBold(count int) string {
	if count == 0 {
		return "-"
	}
	return fmt.Sprintf("**%d**", count)
}

// formatTagsWithInferred formats tags with * suffix for inferred ones
func formatTagsWithInferred(tags []string, inferred []string) string {
	inferredSet := make(map[string]bool)
	for _, t := range inferred {
		inferredSet[t] = true
	}

	parts := make([]string, len(tags))
	for i, tag := range tags {
		if inferredSet[tag] {
			parts[i] = tag + "*"
		} else {
			parts[i] = tag
		}
	}
	return strings.Join(parts, " ")
}

// formatDepsWithInferred formats deps with ~ suffix for inferred from module type
func formatDepsWithInferred(deps []string, inferredDeps []string) string {
	inferredSet := make(map[string]bool)
	for _, d := range inferredDeps {
		inferredSet[d] = true
	}

	parts := make([]string, len(deps))
	for i, dep := range deps {
		// Strip @deps: prefix for display
		short := strings.TrimPrefix(dep, "@deps:")
		if inferredSet[dep] {
			parts[i] = short + "~"
		} else {
			parts[i] = short
		}
	}
	return strings.Join(parts, " ")
}
