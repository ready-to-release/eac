// Command: get units
// Short: List units of work for a specific framework
// Long: Returns structured data about units of work for build, test, lint, or scan.
// Long: Each unit includes its ID, component, tool, cache status, and dependencies.
// Long:
// Long: Usage: get units <build|test|lint|scan> [flags]
// Long:
// Long: The cache status shows whether a unit is:
// Long:   - up_to_date: Cached and doesn't need execution
// Long:   - stale: Needs execution (source changed, previous failure, etc.)
// Long:   - new: Never executed before
// Long:
// Long: Filter Examples:
// Long:   get units build --module eac-cli    # Build units for specific module
// Long:   get units test --cached                  # Only cached test units
// Long:   get units scan --stale                   # Only stale scan units
// Long:   get units lint --container               # Only container-based lint units
// Flag.module: type=string, default="", usage=Filter to specific module
// Flag.component: type=string, default="", usage=Filter to specific component
// Flag.cached: type=bool, default=false, usage=Only show cached (up-to-date) units
// Flag.stale: type=bool, default=false, usage=Only show stale (needs execution) units
// Flag.container: type=bool, default=false, usage=Only show container-based units
// Flag.host: type=bool, default=false, usage=Only show host-installed units
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--as-toml: Output as TOML
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(GetUnits)
}

// unitFilters holds the parsed filter flags.
type unitFilters struct {
	Module    string
	Component string
	Cached    bool
	Stale     bool
	Container bool
	Host      bool
}

// parseUnitFilters extracts filter flags from command arguments.
func parseUnitFilters(args []string) unitFilters {
	filters := unitFilters{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--module":
			if i+1 < len(args) {
				filters.Module = args[i+1]
				i++
			}
		case "--component":
			if i+1 < len(args) {
				filters.Component = args[i+1]
				i++
			}
		case "--cached":
			filters.Cached = true
		case "--stale":
			filters.Stale = true
		case "--container":
			filters.Container = true
		case "--host":
			filters.Host = true
		}
	}
	return filters
}

// toReportFilters converts unitFilters to reports.UnitFilters.
func (f unitFilters) toReportFilters() reports.UnitFilters {
	return reports.UnitFilters{
		Module:    f.Module,
		Component: f.Component,
		Cached:    f.Cached,
		Stale:     f.Stale,
		Container: f.Container,
		Host:      f.Host,
	}
}

// GetUnits returns all units of work for a framework.
func GetUnits() int {
	// Parse framework argument (should be after "get units")
	args := os.Args[3:] // Skip program name, "get", "units"
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: get units <build|test|lint|scan> [flags]\n")
		return 1
	}

	frameworkStr := args[0]
	if frameworkStr[0] == '-' {
		fmt.Fprintf(os.Stderr, "Usage: get units <build|test|lint|scan> [flags]\n")
		return 1
	}

	var framework reports.Framework
	switch frameworkStr {
	case "build":
		framework = reports.FrameworkBuild
	case "test":
		framework = reports.FrameworkTest
	case "lint":
		framework = reports.FrameworkLint
	case "scan":
		framework = reports.FrameworkScan
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid framework '%s'. Use: build, test, lint, or scan\n", frameworkStr)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse filter flags (from remaining args)
	filters := parseUnitFilters(args[1:])

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		report, err := reports.GetUnits(workspaceRoot, framework)
		if err != nil {
			return nil, err
		}

		// Apply filters
		filtered := reports.FilterUnits(report.Units, filters.toReportFilters())

		// Return units with skipped info if any components were skipped
		if len(report.Skipped) > 0 {
			return struct {
				Units   []*reports.UnitInfo         `json:"units" yaml:"units"`
				Skipped []*reports.SkippedComponent `json:"skipped" yaml:"skipped"`
			}{
				Units:   filtered,
				Skipped: report.Skipped,
			}, nil
		}

		return filtered, nil
	})
}
