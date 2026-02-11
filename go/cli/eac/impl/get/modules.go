package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getModulesCommand struct{}

var _ core.SimpleCommandPort = (*getModulesCommand)(nil)

func (c *getModulesCommand) Name() string { return "get modules" }

func (c *getModulesCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-modules",
		Short:         "Get module contracts with optional filtering",
		Long:          "Expected Output:\nYAML list of all module contracts, each containing:\n  - moniker: Unique module identifier\n  - type: Module type (e.g., go, container, typescript, static)\n  - root: Root path relative to repository\n  - depends_on: List of dependency module monikers\n  - Additional metadata (books, files, etc.)\n\nFilter Examples:\n  get modules --calver --with-ci     # CalVer modules that auto-release on CI pass\n  get modules --calver --bundle      # CalVer bundle modules (release when deps change)\n  get modules --semver               # Traditional versioned modules",
		Flags: []core.FlagSpec{
			{Name: "calver", Type: "bool", DefaultValue: "false", Usage: "Filter to only CalVer versioned modules"},
			{Name: "semver", Type: "bool", DefaultValue: "false", Usage: "Filter to only SemVer versioned modules"},
			{Name: "with-ci", Type: "bool", DefaultValue: "false", Usage: "Filter to modules that have a CI workflow"},
			{Name: "with-release", Type: "bool", DefaultValue: "false", Usage: "Filter to modules that have a release workflow"},
			{Name: "bundle", Type: "bool", DefaultValue: "false", Usage: "Filter to bundle modules (CalVer with release but no CI)"},
		},
	}
}

func (c *getModulesCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetModules()
}

// moduleFilters holds the parsed filter flags.
type moduleFilters struct {
	CalVer      bool
	SemVer      bool
	WithCI      bool
	WithRelease bool
	Bundle      bool
}

// parseModuleFilters extracts filter flags from command arguments.
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

// filterModules applies the filters to the module list.
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

		// Check workflow presence using helper methods
		hasCI := mod.GetCIWorkflowPath() != ""
		hasRelease := mod.GetReleaseWorkflowPath() != ""

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
