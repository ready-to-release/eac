// Command: show environments
// Description: Show all environment contracts in a human-readable table
package show

import (
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(ShowEnvironments)
}

func ShowEnvironments() int {
	// Load environment contract using central config
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load config: %v", err)
		return 1
	}

	envConfig := cfg.Environments

	// Validate contract
	if err := envConfig.Validate(); err != nil {
		log.Errorf("Warning: environment contract validation failed: %v", err)
	}

	// Display header
	log.Info("# Environment Contracts\n")
	log.Infof("**Total Environments**: %d  \n", len(envConfig.Environments))

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

	log.Info(tb.Build())
	log.Info("")

	// Display summary by level
	log.Info("## Summary by Level\n")
	l0Envs := envConfig.GetEnvironmentsByLevel("L0")
	l1Envs := envConfig.GetEnvironmentsByLevel("L1")
	l2Envs := envConfig.GetEnvironmentsByLevel("L2")
	l3Envs := envConfig.GetEnvironmentsByLevel("L3")
	l4Envs := envConfig.GetEnvironmentsByLevel("L4")

	log.Infof("- **L0 (Very Fast Unit)**: %d environments", len(l0Envs))
	log.Infof("- **L1 (Fast Unit)**: %d environments", len(l1Envs))
	log.Infof("- **L2 (Local/Docker)**: %d environments", len(l2Envs))
	log.Infof("- **L3 (PLTE)**: %d environments", len(l3Envs))
	log.Infof("- **L4 (Production)**: %d environments", len(l4Envs))
	log.Info("")

	// Display summary by type
	log.Info("## Summary by Type\n")
	typeCounts := make(map[string]int)
	for _, env := range envConfig.Environments {
		typeCounts[env.Type]++
	}

	for envType, count := range typeCounts {
		log.Infof("- **%s**: %d environments", envType, count)
	}
	log.Info("")

	return 0
}
