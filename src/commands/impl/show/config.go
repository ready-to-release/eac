// Command: show config
// Description: Display all EAC configuration in human-readable format
// Short: Display all loaded configurations with defaults applied
// Long: The show config command displays all EAC repository configurations
// Long: loaded with their defaults applied. This includes modules, module types,
// Long: environments, testing tags, and test suites.
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/commands/internal/render"
	"github.com/ready-to-release/eac/src/core/config"
	"github.com/ready-to-release/eac/src/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(ShowConfig)
}

func ShowConfig() int {
	// Check for help flag
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			printShowConfigUsage()
			return 0
		}
	}

	// Load all configs with defaults applied
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Errorf("failed to load configuration: %v", err)
		return 1
	}

	// Display summary table
	log.Info("# EAC Configuration Summary")
	log.Info("")

	summaryTb := render.NewTableBuilder().
		WithHeaders("Config", "Status", "Items")

	// Modules
	if cfg.Modules != nil {
		summaryTb.AddRow("modules", "✓ loaded", len(cfg.Modules.Modules))
	} else {
		summaryTb.AddRow("modules", "✗ not loaded", "-")
	}

	// Module Types
	if cfg.ModuleTypes != nil {
		summaryTb.AddRow("module_types", "✓ loaded", len(cfg.ModuleTypes.Types))
	} else {
		summaryTb.AddRow("module_types", "✗ not loaded", "-")
	}

	// Environments
	if cfg.Environments != nil {
		summaryTb.AddRow("environments", "✓ loaded", len(cfg.Environments.Environments))
	} else {
		summaryTb.AddRow("environments", "✗ not loaded", "-")
	}

	// Testing Tags
	if cfg.TestingTags != nil {
		summaryTb.AddRow("testing_tags", "✓ loaded", len(cfg.TestingTags.Tags))
	} else {
		summaryTb.AddRow("testing_tags", "✗ not loaded", "-")
	}

	// Test Suites
	if cfg.TestSuites != nil {
		summaryTb.AddRow("test_suites", "✓ loaded", len(cfg.TestSuites.Suites))
	} else {
		summaryTb.AddRow("test_suites", "✗ not loaded", "-")
	}

	log.Info(summaryTb.Build())
	log.Info("")

	// Display detailed tables for each config

	// Modules
	if cfg.Modules != nil && len(cfg.Modules.Modules) > 0 {
		log.Info("## Modules")
		log.Info("")
		modTb := render.NewTableBuilder().
			WithHeaders("Moniker", "Type", "Root")
		for _, mod := range cfg.Modules.Modules {
			modTb.AddRow(mod.Moniker, mod.Type, mod.Files.Root)
		}
		log.Info(modTb.Build())
		log.Info("")
	}

	// Module Types
	if cfg.ModuleTypes != nil && len(cfg.ModuleTypes.Types) > 0 {
		log.Info("## Module Types")
		log.Info("")
		typeTb := render.NewTableBuilder().
			WithHeaders("Type", "Build Deps", "Capabilities")
		for _, t := range cfg.ModuleTypes.Types {
			deps := "-"
			if len(t.BuildDeps) > 0 {
				deps = fmt.Sprintf("%v", t.BuildDeps)
			}
			caps := "-"
			if len(t.Capabilities) > 0 {
				caps = fmt.Sprintf("%v", t.Capabilities)
			}
			typeTb.AddRow(t.Name, deps, caps)
		}
		log.Info(typeTb.Build())
		log.Info("")
	}

	// Environments
	if cfg.Environments != nil && len(cfg.Environments.Environments) > 0 {
		log.Info("## Environments")
		log.Info("")
		envTb := render.NewTableBuilder().
			WithHeaders("Name", "Type", "Description")
		for _, env := range cfg.Environments.Environments {
			desc := env.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			envTb.AddRow(env.Name, env.Type, desc)
		}
		log.Info(envTb.Build())
		log.Info("")
	}

	// Testing Tags - group by type
	if cfg.TestingTags != nil && len(cfg.TestingTags.Tags) > 0 {
		log.Info("## Testing Tags")
		log.Info("")
		// Group tags by type
		tagsByType := make(map[string][]config.TagDefinition)
		for _, tag := range cfg.TestingTags.Tags {
			tagsByType[tag.Type] = append(tagsByType[tag.Type], tag)
		}
		for _, tagType := range cfg.TestingTags.Types {
			tags := tagsByType[tagType.Type]
			if len(tags) == 0 {
				continue
			}
			log.Infof("### %s", tagType.Type)
			log.Info("")
			tagTb := render.NewTableBuilder().
				WithHeaders("Tag", "Description")
			for _, tag := range tags {
				desc := tag.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				tagTb.AddRow(tag.Tag, desc)
			}
			log.Info(tagTb.Build())
			log.Info("")
		}
	}

	// Test Suites
	if cfg.TestSuites != nil && len(cfg.TestSuites.Suites) > 0 {
		log.Info("## Test Suites")
		log.Info("")
		suiteTb := render.NewTableBuilder().
			WithHeaders("Moniker", "Name", "Description")
		for _, suite := range cfg.TestSuites.Suites {
			desc := suite.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			suiteTb.AddRow(suite.Moniker, suite.Name, desc)
		}
		log.Info(suiteTb.Build())
		log.Info("")
	}

	log.Info("Use 'r2r get config' for full YAML output with all fields.")

	return 0
}

func printShowConfigUsage() {
	log.Info("Display all EAC configuration in human-readable format")
	log.Info("")
	log.Info("Usage: r2r show config")
	log.Info("")
	log.Info("This command displays a summary of all loaded configurations:")
	log.Info("  - modules: Module contracts with defaults applied")
	log.Info("  - module_types: Module type definitions")
	log.Info("  - environments: Environment contracts")
	log.Info("  - testing_tags: Testing tag definitions")
	log.Info("  - test_suites: Test suite configurations")
	log.Info("")
	log.Info("For full structured output, use 'r2r get config'.")
}
