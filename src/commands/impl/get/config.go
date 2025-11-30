// Command: get config
// Description: Get all EAC configuration in structured format
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
// HasSideEffects: false
package get

import (
	"fmt"
	"os"

	get "github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/config"
)

func init() {
	registry.Register(GetConfig)
}

// ConfigOutput represents the structured output of all configs
type ConfigOutput struct {
	Modules      interface{} `yaml:"modules" json:"modules"`
	ModuleTypes  interface{} `yaml:"module_types" json:"module_types"`
	Environments interface{} `yaml:"environments" json:"environments"`
	TestingTags  interface{} `yaml:"testing_tags" json:"testing_tags"`
	TestSuites   interface{} `yaml:"test_suites" json:"test_suites"`
}

func GetConfig() int {
	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		// Load all configs with defaults applied
		cfg, err := config.Load(config.DefaultLoadOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}

		// Build output structure
		output := ConfigOutput{}

		if cfg.Modules != nil {
			output.Modules = cfg.Modules
		}

		if cfg.ModuleTypes != nil {
			output.ModuleTypes = cfg.ModuleTypes
		}

		if cfg.Environments != nil {
			output.Environments = cfg.Environments
		}

		if cfg.TestingTags != nil {
			output.TestingTags = cfg.TestingTags
		}

		if cfg.TestSuites != nil {
			output.TestSuites = cfg.TestSuites
		}

		return output, nil
	})
}

// printConfigUsage prints help for the get config command
func printConfigUsage() {
	fmt.Println("Get all EAC configuration in structured format")
	fmt.Println()
	fmt.Println("Usage: r2r get config [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --as-yaml    Output as YAML (default)")
	fmt.Println("  --as-json    Output as JSON")
	fmt.Println("  --as-toml    Output as TOML")
	fmt.Println("  -h, --help   Show this help message")
	fmt.Println()
	fmt.Println("Output includes all loaded configurations with defaults applied:")
	fmt.Println("  - modules: Module contracts")
	fmt.Println("  - module_types: Module type definitions")
	fmt.Println("  - environments: Environment contracts")
	fmt.Println("  - testing_tags: Testing tag definitions")
	fmt.Println("  - test_suites: Test suite configurations")
}

func checkConfigHelp() bool {
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			printConfigUsage()
			return true
		}
	}
	return false
}
