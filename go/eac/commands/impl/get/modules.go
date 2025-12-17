// Command: get modules
// Description: Get all module contracts in the repository
// Flag.calver: type=bool, default=false, usage=Filter to only CalVer versioned modules
// Flag.semver: type=bool, default=false, usage=Filter to only SemVer versioned modules
// Flag.with-ci: type=bool, default=false, usage=Filter to modules that have a CI workflow
// Flag.with-release: type=bool, default=false, usage=Filter to modules that have a release workflow
// Flag.bundle: type=bool, default=false, usage=Filter to bundle modules (CalVer with release but no CI)
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --as-<name>: Output using custom renderer (e.g., --as-summary, --as-count)
// Long:
// Long: Expected Output:
// Long: YAML list of all module contracts, each containing:
// Long:   - moniker: Unique module identifier
// Long:   - type: Module type (e.g., go, container, typescript, static)
// Long:   - root: Root path relative to repository
// Long:   - depends_on: List of dependency module monikers
// Long:   - Additional metadata (books, files, etc.)
// Long:
// Long: Filter Examples:
// Long:   get modules --calver --with-ci     # CalVer modules that auto-release on CI pass
// Long:   get modules --calver --bundle      # CalVer bundle modules (release when deps change)
// Long:   get modules --semver               # Traditional versioned modules
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetModules)
}

// moduleFilters holds the parsed filter flags
type moduleFilters struct {
	CalVer      bool
	SemVer      bool
	WithCI      bool
	WithRelease bool
	Bundle      bool
}

// parseModuleFilters extracts filter flags from command arguments
func parseModuleFilters(args []string) moduleFilters {
	filters := moduleFilters{}
	for _, arg := range args {
		switch arg {
		case "--calver":
			filters.CalVer = true
		case "--semver":
			filters.SemVer = true
		case "--with-ci":
			filters.WithCI = true
		case "--with-release":
			filters.WithRelease = true
		case "--bundle":
			filters.Bundle = true
		}
	}
	return filters
}

// filterModules applies the filters to the module list
func filterModules(mods []*modules.ModuleContract, filters moduleFilters) []*modules.ModuleContract {
	// If no filters are set, return all modules
	if !filters.CalVer && !filters.SemVer && !filters.WithCI && !filters.WithRelease && !filters.Bundle {
		return mods
	}

	var result []*modules.ModuleContract
	for _, mod := range mods {
		// Check versioning scheme
		scheme := ""
		if mod.Versioning != nil {
			scheme = strings.ToLower(mod.Versioning.Scheme)
		}

		isCalVer := scheme == "calver"
		isSemVer := scheme == "semver"

		// Check workflow presence (Files and Workflows are struct types, not pointers)
		hasCI := mod.Files.Workflows.CI != ""
		hasRelease := mod.Files.Workflows.Release != ""

		// Bundle mode: CalVer + has release + no CI
		isBundle := isCalVer && hasRelease && !hasCI

		// Apply filters
		if filters.CalVer && !isCalVer {
			continue
		}
		if filters.SemVer && !isSemVer {
			continue
		}
		if filters.WithCI && !hasCI {
			continue
		}
		if filters.WithRelease && !hasRelease {
			continue
		}
		if filters.Bundle && !isBundle {
			continue
		}

		result = append(result, mod)
	}

	return result
}

func GetModules() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse filter flags
	filters := parseModuleFilters(os.Args[1:])

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		report, err := reports.GetModuleContracts(workspaceRoot)
		if err != nil {
			return nil, err
		}

		// Apply filters
		filtered := filterModules(report.Modules, filters)
		return filtered, nil
	})
}
