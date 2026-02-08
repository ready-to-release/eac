// Command: show config
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
// Flag.verbose: type=bool, default=false, usage=Show all config source files with layers and value counts
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
)

func ShowConfig() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	args := os.Args[2:]

	// Check for help flag
	if flags.HasFlag(args, "--help", "-h") {
		printShowConfigUsage()
		return 0
	}

	// Check for verbose flag
	verbose := flags.HasFlag(args, "--verbose", "-v")

	// Load all configs with defaults applied
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		return 1
	}

	// If verbose, show source files first
	if verbose {
		printConfigSources(cfg)
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

	// Component Types
	if cfg.ComponentTypes != nil {
		summaryTb.AddRow("component_types", "✓ loaded", len(cfg.ComponentTypes.ComponentTypes))
	} else {
		summaryTb.AddRow("component_types", "✗ not loaded", "-")
	}

	// Environments
	if cfg.Environments != nil {
		summaryTb.AddRow("environments", "✓ loaded", len(cfg.Environments.Environments))
	} else {
		summaryTb.AddRow("environments", "✗ not loaded", "-")
	}

	// Testing (unified config from eac-testing contract)
	if cfg.Testing != nil {
		summaryTb.AddRow("testing", "✓ loaded", fmt.Sprintf("%d tags, %d suites", len(cfg.Testing.ListTags()), len(cfg.Testing.ListSuites())))
	} else {
		summaryTb.AddRow("testing", "✗ not loaded", "-")
	}

	fmt.Println(summaryTb.Build())
	fmt.Println("")

	// Display detailed tables for each config

	// Repository Settings
	if cfg.Repository != nil {
		fmt.Println("## Repository Settings")
		fmt.Println("")
		settingsTb := render.NewTableBuilder().
			WithHeaders("Setting", "Value")
		settingsTb.AddRow("type", valueOrDefault(cfg.Repository.Repository.Type, "-"))
		settingsTb.AddRow("trunk_branch", valueOrDefault(cfg.Repository.Repository.TrunkBranch, "main"))
		settingsTb.AddRow("max_branch_age_days", cfg.Repository.Repository.MaxBranchAgeDays)
		settingsTb.AddRow("optimize_git_ls_in_ci", cfg.Repository.Repository.OptimizeGitLsInCI)
		if cfg.Repository.Repository.Parallelism.CI > 0 || cfg.Repository.Repository.Parallelism.Devbox > 0 {
			settingsTb.AddRow("parallelism.ci", cfg.Repository.Repository.Parallelism.CI)
			settingsTb.AddRow("parallelism.devbox", cfg.Repository.Repository.Parallelism.Devbox)
		}
		fmt.Println(settingsTb.Build())
		fmt.Println("")
	}

	// Modules
	if cfg.Repository != nil && len(cfg.Repository.Modules) > 0 {
		fmt.Println("## Modules")
		fmt.Println("")
		modTb := render.NewTableBuilder().
			WithHeaders("Moniker", "Type", "Root")
		for _, mod := range cfg.Repository.Modules {
			// Get first package root for display
			var displayRoot string
			for _, pkgName := range mod.GetEnabledComponents() {
				root := mod.Components.GetComponentRoot(pkgName)
				if root != "" {
					displayRoot = root
					break
				}
			}
			modTb.AddRow(mod.Moniker, mod.GetComponentTypesDisplay(), displayRoot)
		}
		fmt.Println(modTb.Build())
		fmt.Println("")
	}

	// Package Types
	if cfg.ComponentTypes != nil && len(cfg.ComponentTypes.ComponentTypes) > 0 {
		fmt.Println("## Package Types")
		fmt.Println("")
		typeTb := render.NewTableBuilder().
			WithHeaders("Type", "Pool")
		for name, t := range cfg.ComponentTypes.ComponentTypes {
			pool := t.GetPool()
			typeTb.AddRow(name, pool)
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

	// Testing - Tags and Suites (from unified Testing config)
	if cfg.Testing != nil {
		// Testing Tags - display by type
		tags := cfg.Testing.ListTags()
		if len(tags) > 0 {
			fmt.Println("## Testing Tags")
			fmt.Println("")
			tagTb := render.NewTableBuilder().
				WithHeaders("Tag", "Type", "Description")
			for _, tagName := range tags {
				tag, ok := cfg.Testing.GetTag(tagName)
				if !ok {
					continue
				}
				desc := tag.Description()
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				tagTb.AddRow(tag.Tag(), tag.Type(), desc)
			}
			fmt.Println(tagTb.Build())
			fmt.Println("")
		}

		// Test Suites
		suites := cfg.Testing.ListSuites()
		if len(suites) > 0 {
			fmt.Println("## Test Suites")
			fmt.Println("")
			suiteTb := render.NewTableBuilder().
				WithHeaders("Moniker", "Name", "Description")
			for _, moniker := range suites {
				suite, ok := cfg.Testing.GetSuite(moniker)
				if !ok {
					continue
				}
				desc := suite.Description()
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				suiteTb.AddRow(suite.Moniker(), suite.Name(), desc)
			}
			fmt.Println(suiteTb.Build())
			fmt.Println("")
		}
	}

	fmt.Println("Use 'clie get config' for full YAML output with all fields.")

	return 0
}

func printShowConfigUsage() {
	fmt.Println("Display all EAC configuration in human-readable format")
	fmt.Println("")
	fmt.Println("Usage: clie show config [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  --verbose, -v  Show all config source files with layers and value counts")
	fmt.Println("")
	fmt.Println("This command displays a summary of all loaded configurations:")
	fmt.Println("  - modules: Module contracts with defaults applied")
	fmt.Println("  - module_types: Module type definitions")
	fmt.Println("  - environments: Environment contracts")
	fmt.Println("  - testing: Testing tags and test suites (from eac-testing contract)")
	fmt.Println("")
	fmt.Println("With --verbose, shows configuration source files:")
	fmt.Println("  - Contract defaults from contracts/core/0.1.0/defaults/")
	fmt.Println("  - User configuration from .eac/")
	fmt.Println("  - File existence status and value counts")
	fmt.Println("")
	fmt.Println("For full structured output, use 'clie get config'.")
}

// valueOrDefault returns the value if non-empty, otherwise the default.
func valueOrDefault(value, defaultVal string) string {
	if value == "" {
		return defaultVal
	}
	return value
}

// printConfigSources displays all configuration source files with their layers.
func printConfigSources(cfg *config.EACConfig) {
	fmt.Println("# Configuration Sources")
	fmt.Println("")
	fmt.Println("Configuration files are loaded in layer order: contract → user → personal")
	fmt.Println("Later layers override earlier ones.")
	fmt.Println("")

	loadedFiles := cfg.GetLoadedFiles()

	for _, lc := range loadedFiles {
		fmt.Printf("## %s\n", lc.Name)
		fmt.Println("")

		tb := render.NewTableBuilder().
			WithHeaders("Layer", "Path", "Status", "Values")

		for _, file := range lc.Files {
			status := "✗ not found"
			values := "-"
			if file.Exists {
				status = "✓ loaded"
				values = fmt.Sprintf("%d", file.Values)
			}
			tb.AddRow(string(file.Layer), file.Path, status, values)
		}

		fmt.Println(tb.Build())
		fmt.Println("")
	}
}
