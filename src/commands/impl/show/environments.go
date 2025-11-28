// Command: show environments
// Description: Show all environment contracts in a human-readable table
// HasSideEffects: false
package show

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/config"
)

func init() {
	registry.Register(ShowEnvironments)
}

func ShowEnvironments() int {
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

	// Display contract metadata
	fmt.Printf("# Environment Contracts\n\n")
	fmt.Printf("**Version**: %s  \n", envConfig.Metadata.Version)
	fmt.Printf("**Description**: %s  \n", envConfig.Metadata.Description)
	fmt.Printf("**Total Environments**: %d  \n\n", len(envConfig.Environments))

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("Moniker", "Name", "Level", "Type", "System Dependencies", "Env Tags")

	for _, env := range envConfig.Environments {
		systemDepsStr := strings.Join(env.SystemDeps, ", ")
		envTagsStr := strings.Join(env.EnvTags, ", ")

		tb.AddRow(
			env.Moniker,
			env.Name,
			env.Level,
			env.Type,
			systemDepsStr,
			envTagsStr,
		)
	}

	fmt.Println(tb.Build())
	fmt.Printf("\n")

	// Display summary by level
	fmt.Printf("## Summary by Level\n\n")
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
	fmt.Printf("\n")

	// Display summary by type
	fmt.Printf("## Summary by Type\n\n")
	typeCounts := make(map[string]int)
	for _, env := range envConfig.Environments {
		typeCounts[env.Type]++
	}

	for envType, count := range typeCounts {
		fmt.Printf("- **%s**: %d environments\n", envType, count)
	}
	fmt.Printf("\n")

	return 0
}
