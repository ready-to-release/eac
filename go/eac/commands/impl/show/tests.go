// Command: show tests
// Description: Show all tests in the repository in a human-readable table
package show

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/internal/testdata"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(ShowTests)
}

func ShowTests() int {
	// Get repository root
	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("failed to get current directory: %v", err)
		return 1
	}

	repoRoot, err := testdata.FindRepoRoot(cwd)
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Get all tests with metadata and aggregations
	data, err := testdata.GetAllTests(repoRoot)
	if err != nil {
		log.Errorf("failed to get tests: %v", err)
		return 1
	}

	// Display header
	log.Info("# All Tests\n")
	log.Infof("**Total Tests**: %d  \n", data.TotalCount)

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("#", "Moniker", "Type", "Module", "Level", "Verification", "System Deps")

	for i, entry := range data.Tests {
		levelStr := strings.Join(entry.Level, ", ")
		verificationStr := strings.Join(entry.Verification, ", ")
		systemDepsStr := strings.Join(entry.SystemDeps, ", ")

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			entry.Moniker,
			entry.Type,
			entry.Module,
			levelStr,
			verificationStr,
			systemDepsStr,
		)
	}

	log.Info(tb.Build())
	log.Info("")

	// Display summary by type
	log.Info("## Summary\n")
	log.Info("### By Type\n")
	for testType, count := range data.ByType {
		log.Infof("- **%s**: %d tests", testType, count)
	}
	log.Info("")

	// Display summary by level (ordered)
	log.Info("### By Level\n")
	for _, level := range []string{"@L0", "@L1", "@L2", "@L3", "@L4"} {
		if count, ok := data.ByLevel[level]; ok {
			log.Infof("- **%s**: %d tests", level, count)
		}
	}
	log.Info("")

	// Display summary by module (sorted)
	log.Info("### By Module\n")
	modules := make([]string, 0, len(data.ByModule))
	for module := range data.ByModule {
		modules = append(modules, module)
	}
	sort.Strings(modules)
	for _, module := range modules {
		log.Infof("- **%s**: %d tests", module, data.ByModule[module])
	}
	log.Info("")

	return 0
}
