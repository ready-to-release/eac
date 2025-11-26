// Command Extension Entry Point
//
// This extension wraps src-commands for R2R CLI integration.
// It provides dynamic command discovery and metadata generation.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"gopkg.in/yaml.v3"

	// Import all command packages to register them via init()
	_ "github.com/ready-to-release/eac/src/commands/impl/build"
	_ "github.com/ready-to-release/eac/src/commands/impl/commit"
	_ "github.com/ready-to-release/eac/src/commands/impl/completion"
	_ "github.com/ready-to-release/eac/src/commands/impl/describe"
	_ "github.com/ready-to-release/eac/src/commands/impl/design"
	_ "github.com/ready-to-release/eac/src/commands/impl/docs"
	_ "github.com/ready-to-release/eac/src/commands/impl/get"
	_ "github.com/ready-to-release/eac/src/commands/impl/help"
	_ "github.com/ready-to-release/eac/src/commands/impl/init"
	_ "github.com/ready-to-release/eac/src/commands/impl/list"
	_ "github.com/ready-to-release/eac/src/commands/impl/pipeline"
	_ "github.com/ready-to-release/eac/src/commands/impl/release"
	_ "github.com/ready-to-release/eac/src/commands/impl/show"
	_ "github.com/ready-to-release/eac/src/commands/impl/specs"
	_ "github.com/ready-to-release/eac/src/commands/impl/templates"
	_ "github.com/ready-to-release/eac/src/commands/impl/test"
	_ "github.com/ready-to-release/eac/src/commands/impl/validate"
	_ "github.com/ready-to-release/eac/src/commands/impl/work"
)

// Metadata defines the extension metadata structure
type Metadata struct {
	Name               string             `yaml:"name"`
	Version            string             `yaml:"version"`
	Description        string             `yaml:"description"`
	SchemaVersion      string             `yaml:"schema-version"`
	Capabilities       []string           `yaml:"capabilities"`
	Commands           map[string]Command `yaml:"commands"`
	Requirements       Requirements       `yaml:"requirements"`
	Volumes            []Volume           `yaml:"volumes,omitempty"`
	ExpectedHostImages []string           `yaml:"expected-host-images,omitempty"`
	Env                []EnvVar           `yaml:"env,omitempty"`
	Metadata           ExtensionMetadata  `yaml:"metadata"`
}

// EnvVar defines an environment variable request from the extension
type EnvVar struct {
	Name     string `yaml:"name"`                        // Environment variable name
	Value    string `yaml:"value,omitempty"`             // If empty, pass through from host
	Required bool   `yaml:"required,omitempty"`          // If true, fail at runtime when not set
}

// Volume defines a volume mount request from the extension
type Volume struct {
	Name   string `yaml:"name"`   // Logical name for the volume (used in Docker volume naming)
	Target string `yaml:"target"` // Container path to mount
	Type   string `yaml:"type"`   // Volume type: "cache" for named volumes, "bind" for bind mounts
}

// Command defines a single command structure
type Command struct {
	Description string   `yaml:"description"`
	Parameters  []string `yaml:"parameters,omitempty"`
}

// Requirements defines extension requirements
type Requirements struct {
	R2RCLIVersion    string `yaml:"r2r-version"`
	ContainerRuntime string `yaml:"container-runtime"`
	MinimumMemory    string `yaml:"minimum-memory"`
	MinimumCPU       string `yaml:"minimum-cpu"`
}

// ExtensionMetadata defines extension metadata
type ExtensionMetadata struct {
	Author        string   `yaml:"author"`
	Repository    string   `yaml:"repository"`
	Documentation string   `yaml:"documentation"`
	License       string   `yaml:"license"`
	Tags          []string `yaml:"tags"`
}

func main() {
	// Handle special commands
	if len(os.Args) < 2 {
		printAvailableCommands()
		os.Exit(0)
	}

	command := os.Args[1]

	// Handle extension-meta command
	if command == "extension-meta" {
		outputMetadata()
		os.Exit(0)
	}

	// Handle help command - delegate to src-commands
	if command == "help" {
		exitCode := dispatchToCommands(os.Args[1:])
		os.Exit(exitCode)
	}

	// Pass through all other commands to src-commands dispatcher
	exitCode := dispatchToCommands(os.Args[1:])
	os.Exit(exitCode)
}

// printAvailableCommands prints a user-friendly list of available commands
func printAvailableCommands() {
	fmt.Println("R2R CLI Command Extension - Repository Management Tooling")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  extension-meta    Display extension metadata (YAML)")
	fmt.Println("  help              Display detailed command help")
	fmt.Println()
	fmt.Println("Command Categories:")

	// Get command registry
	cmdRegistry := registry.GetCommandRegistry()

	// Group commands by category (first word)
	categories := make(map[string][]string)
	for _, reg := range cmdRegistry {
		parts := strings.Split(reg.ActualCommand, " ")
		category := parts[0]
		categories[category] = append(categories[category], reg.ActualCommand)
	}

	// Print categories
	categoryOrder := []string{"build", "design", "docs", "get", "show", "list", "describe",
		"pipeline", "release", "specs", "templates", "test", "validate", "work"}

	for _, category := range categoryOrder {
		if cmds, ok := categories[category]; ok {
			fmt.Printf("  %s: %d commands\n", category, len(cmds))
		}
	}

	fmt.Println()
	fmt.Println("Use 'help' to see all available commands with descriptions.")
	fmt.Println("Use 'help <command>' for detailed information about a specific command.")
	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Println("  r2r run cmd show modules")
	fmt.Println("  r2r run cmd test module src-core")
	fmt.Println("  r2r run cmd get dependencies")
}

// outputMetadata generates and outputs extension metadata from command registry
func outputMetadata() {
	// Get command registry
	cmdRegistry := registry.GetCommandRegistry()

	// Build commands map from registry
	commands := make(map[string]Command)

	for _, reg := range cmdRegistry {
		// Use Short description if available, fall back to Description
		description := reg.Short
		if description == "" {
			description = reg.Description
		}

		// Extract parameter names from flags
		params := make([]string, 0, len(reg.Flags))
		for _, flag := range reg.Flags {
			params = append(params, flag.Name)
		}

		commands[reg.ActualCommand] = Command{
			Description: description,
			Parameters:  params,
		}
	}

	// Build metadata structure
	metadata := Metadata{
		Name:          "eac",
		Version:       "1.0.0",
		Description:   "Everything as Code - Repository management tooling for R2R CLI",
		SchemaVersion: "0.1.0",
		Capabilities: []string{
			"repository-management",
			"module-discovery",
			"dependency-analysis",
			"testing",
			"documentation",
			"specifications",
			"architecture-design",
			"git-worktree",
			"build-automation",
			"pipeline-execution",
		},
		Commands: commands,
		Requirements: Requirements{
			R2RCLIVersion:    ">=1.0.0",
			ContainerRuntime: "docker",
			MinimumMemory:    "256Mi",
			MinimumCPU:       "0.1",
		},
		Volumes: []Volume{
			{
				Name:   "go-build",
				Target: "/root/.cache/go-build",
				Type:   "cache",
			},
			{
				Name:   "go-mod",
				Target: "/go/pkg/mod",
				Type:   "cache",
			},
		},
		ExpectedHostImages: []string{
			"cli-mkdocs:latest",       // docs serve command
			"structurizr/lite:latest", // design serve command
		},
		Env: []EnvVar{
			{Name: "GODOG_SUITE_TAGS"},    // Test suite tag filter for godog scenarios
			{Name: "GODOG_FORMAT"},        // Godog output format
			{Name: "GODOG_OUTPUT_DIR"},    // Godog report output directory
			{Name: "GODOG_REPORT_FORMAT"}, // Godog report format (cucumber/junit)
			{Name: "GODOG_REPORT_NAME"},   // Godog report file name
			{Name: "GODOG_PATHS"},         // Godog feature file paths
			{Name: "R2R_TEST_RUN_ID"},     // Test run identifier
		},
		Metadata: ExtensionMetadata{
			Author:        "Ready to Release Team",
			Repository:    "https://github.com/ready-to-release/eac",
			Documentation: "https://github.com/ready-to-release/eac/tree/main/containers/ext-eac",
			License:       "MIT",
			Tags:          []string{"go", "repository", "modular-monorepo", "tooling", "everything-as-code"},
		},
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&metadata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(yamlData))
}

// dispatchToCommands delegates command execution to src-commands
// This function replicates the logic from src/commands/main.go
func dispatchToCommands(args []string) int {
	// Import the commands map
	commands := registry.GetCommands()

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no command specified\n")
		return 1
	}

	// Try longest match first for nested commands
	var cmdFunc registry.CommandFunc
	var exists bool

	for argCount := len(args); argCount >= 1; argCount-- {
		testPath := strings.Join(args[:argCount], " ")
		if fn, found := commands[testPath]; found {
			cmdFunc = fn
			exists = true
			// Update os.Args to reflect the matched command structure
			// This allows command implementations to parse remaining args correctly
			newArgs := []string{os.Args[0]}
			newArgs = append(newArgs, args[:argCount]...)
			newArgs = append(newArgs, args[argCount:]...)
			os.Args = newArgs
			break
		}
	}

	if !exists {
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", strings.Join(args, " "))
		fmt.Fprintf(os.Stderr, "Use 'help' to see available commands.\n")
		return 1
	}

	// Execute the command
	return cmdFunc()
}
