// Command: show config
// Description: Display all EAC configuration in human-readable format
// Short: Display all loaded configurations with defaults applied
// Long: The show config command displays all EAC repository configurations
// Long: loaded with their defaults applied. This includes modules, module types,
// Long: environments, testing tags, and test suites.
// Long:
// Long: Expected Output:
// Long: - Human-readable display of all configurations with summary table showing status and counts
// Long: - Detailed tables for modules (moniker, type, root), module types, environments
// Long: - Testing tags grouped by type, test suites with descriptions
// Long: - Each section clearly formatted with markdown headers and tables
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
)

func init() {
	registry.Register(ShowConfig)
}

func ShowConfig() int {
	args := os.Args[2:]

	// Validate flags
	commandFlags := []flags.FlagDefinition{
		{Name: "--help", Shorthand: "-h", HasValue: false},
	}
	if err := flags.ValidateFlags(args, commandFlags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Check for help flag
	if flags.HasFlag(args, "--help", "-h") {
		printShowConfigUsage()
		return 0
	}

	// Load all configs with defaults applied
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		return 1
	}

	// Display summary table
	fmt.Println("# EAC Configuration Summary")
	fmt.Println("")

	summaryTb := render.NewTableBuilder().
		WithHeaders("Config", "Status", "Items")

	// Modules
	if cfg.Repository != nil {
		summaryTb.AddRow("modules", "✓ loaded", len(cfg.Repository.Modules))
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

	fmt.Println(summaryTb.Build())
	fmt.Println("")

	// Display detailed tables for each config

	// Modules
	if cfg.Repository != nil && len(cfg.Repository.Modules) > 0 {
		fmt.Println("## Modules")
		fmt.Println("")
		modTb := render.NewTableBuilder().
			WithHeaders("Moniker", "Type", "Root")
		for _, mod := range cfg.Repository.Modules {
			modTb.AddRow(mod.Moniker, mod.Type, mod.Files.Root)
		}
		fmt.Println(modTb.Build())
		fmt.Println("")
	}

	// Module Types
	if cfg.ModuleTypes != nil && len(cfg.ModuleTypes.Types) > 0 {
		fmt.Println("## Module Types")
		fmt.Println("")
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
		fmt.Println(typeTb.Build())
		fmt.Println("")
	}

	// Environments
	if cfg.Environments != nil && len(cfg.Environments.Environments) > 0 {
		fmt.Println("## Environments")
		fmt.Println("")
		envTb := render.NewTableBuilder().
			WithHeaders("Name", "Type", "Description")
		for _, env := range cfg.Environments.Environments {
			desc := env.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			envTb.AddRow(env.Name, env.Type, desc)
		}
		fmt.Println(envTb.Build())
		fmt.Println("")
	}

	// Testing Tags - group by type
	if cfg.TestingTags != nil && len(cfg.TestingTags.Tags) > 0 {
		fmt.Println("## Testing Tags")
		fmt.Println("")
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
			fmt.Printf("### %s\n", tagType.Type)
			fmt.Println("")
			tagTb := render.NewTableBuilder().
				WithHeaders("Tag", "Description")
			for _, tag := range tags {
				desc := tag.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				tagTb.AddRow(tag.Tag, desc)
			}
			fmt.Println(tagTb.Build())
			fmt.Println("")
		}
	}

	// Test Suites
	if cfg.TestSuites != nil && len(cfg.TestSuites.Suites) > 0 {
		fmt.Println("## Test Suites")
		fmt.Println("")
		suiteTb := render.NewTableBuilder().
			WithHeaders("Moniker", "Name", "Description")
		for _, suite := range cfg.TestSuites.Suites {
			desc := suite.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			suiteTb.AddRow(suite.Moniker, suite.Name, desc)
		}
		fmt.Println(suiteTb.Build())
		fmt.Println("")
	}

	fmt.Println("Use 'r2r get config' for full YAML output with all fields.")

	return 0
}

func printShowConfigUsage() {
	fmt.Println("Display all EAC configuration in human-readable format")
	fmt.Println("")
	fmt.Println("Usage: r2r show config")
	fmt.Println("")
	fmt.Println("This command displays a summary of all loaded configurations:")
	fmt.Println("  - modules: Module contracts with defaults applied")
	fmt.Println("  - module_types: Module type definitions")
	fmt.Println("  - environments: Environment contracts")
	fmt.Println("  - testing_tags: Testing tag definitions")
	fmt.Println("  - test_suites: Test suite configurations")
	fmt.Println("")
	fmt.Println("For full structured output, use 'r2r get config'.")
}
