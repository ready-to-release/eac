// Command: show environments
// Short: Display all environment configurations
// Long: The show environments command displays all environment contracts defined in environments.yml.
// Long: Shows environment details including moniker, name, level, type, and system dependencies.
// Long:
// Long: Expected Output:
// Long: - Table with environment definitions showing: Moniker, Name, Level, Type, System Dependencies
// Long: - Summary by level (L0-L4) showing count of environments at each level
// Long: - Summary by type showing count of environments for each type
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(ShowEnvironments)
}

func ShowEnvironments() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Load environment contract using central config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		return 1
	}

	envConfig := cfg.Environments

	// Validate contract
	if err := envConfig.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: environment contract validation failed: %v\n", err)
	}

	// Display header
	fmt.Println("# Environment Contracts")
	fmt.Println("")
	fmt.Printf("**Total Environments**: %d  \n", len(envConfig.Environments))

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("Moniker", "Name", "Level", "Type", "System Dependencies")

	for _, env := range envConfig.Environments {
		systemDepsStr := strings.Join(env.SystemDeps, ", ")

		tb.AddRow(
			env.Moniker,
			env.Name,
			env.Level,
			env.Type,
			systemDepsStr,
		)
	}

	fmt.Println(tb.Build())
	fmt.Println("")

	// Display summary by level
	fmt.Println("## Summary by Level")
	fmt.Println("")
	l0Envs := envConfig.GetEnvironmentsByLevel("L0")
	l1Envs := envConfig.GetEnvironmentsByLevel("L1")
	l2Envs := envConfig.GetEnvironmentsByLevel("L2")
	l3Envs := envConfig.GetEnvironmentsByLevel("L3")
	l4Envs := envConfig.GetEnvironmentsByLevel("L4")

	fmt.Printf("- **L0 (Very Fast Unit)**: %d environments\n", len(l0Envs))
	fmt.Printf("- **L1 (Fast Unit)**: %d environments\n", len(l1Envs))
	fmt.Printf("- **L2 (Local/Docker)**: %d environments\n", len(l2Envs))
	fmt.Printf("- **L3 (PLTE)**: %d environments\n", len(l3Envs))
	fmt.Printf("- **L4 (Production)**: %d environments\n", len(l4Envs))
	fmt.Println("")

	// Display summary by type
	fmt.Println("## Summary by Type")
	fmt.Println("")
	typeCounts := make(map[string]int)
	for _, env := range envConfig.Environments {
		typeCounts[env.Type]++
	}

	for envType, count := range typeCounts {
		fmt.Printf("- **%s**: %d environments\n", envType, count)
	}
	fmt.Println("")

	return 0
}
