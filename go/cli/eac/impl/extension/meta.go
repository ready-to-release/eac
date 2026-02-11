package extension

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/tool"
	"gopkg.in/yaml.v3"
)

type extensionMetaCommand struct{}

var _ core.SimpleCommandPort = (*extensionMetaCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&extensionMetaCommand{},
	}
}

func (c *extensionMetaCommand) Name() string { return "extension-meta" }

func (c *extensionMetaCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "extension-meta",
		Short:         "Output extension metadata for clie CLI",
		Long:          "Outputs YAML-formatted metadata describing the extension's capabilities,\ncommands, requirements, and configuration for integration with the clie CLI.\nThis command is used by clie CLI to discover and configure extensions.\n\nExpected Output:\n  - YAML-formatted metadata describing extension capabilities\n  - Commands, requirements, and configuration for clie CLI integration",
	}
}

func (c *extensionMetaCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ExtensionMeta()
}

var log = logging.C()

// Metadata defines the extension metadata structure for clie CLI.
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

// EnvVar defines an environment variable request from the extension.
type EnvVar struct {
	Name     string `yaml:"name"`
	Value    string `yaml:"value,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}

// Volume defines a volume mount request from the extension.
type Volume struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
	Type   string `yaml:"type"`
}

// Command defines a single command structure.
type Command struct {
	Description string   `yaml:"description"`
	Parameters  []string `yaml:"parameters,omitempty"`
}

// Requirements defines extension requirements.
type Requirements struct {
	CLIEVersion      string `yaml:"clie-version"`
	ContainerRuntime string `yaml:"container-runtime"`
	MinimumMemory    string `yaml:"minimum-memory"`
	MinimumCPU       string `yaml:"minimum-cpu"`
}

// ExtensionMetadata defines extension metadata.
type ExtensionMetadata struct {
	Author        string   `yaml:"author"`
	Repository    string   `yaml:"repository"`
	Documentation string   `yaml:"documentation"`
	License       string   `yaml:"license"`
	Tags          []string `yaml:"tags"`
}

// ExtensionMeta outputs extension metadata in YAML format for clie CLI.
func ExtensionMeta() int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Get command registry
	reg := registry.Global()

	// Build commands map from registry
	commands := make(map[string]Command)

	for _, cmd := range reg.All() {
		cmdName := cmd.Name()
		meta := cmd.Metadata()

		// Skip the extension-meta command itself
		if cmdName == "extension-meta" {
			continue
		}

		// Extract parameter names from flags
		params := make([]string, 0, len(meta.Flags))
		for _, flag := range meta.Flags {
			params = append(params, flag.Name)
		}

		commands[cmdName] = Command{
			Description: meta.Short,
			Parameters:  params,
		}
	}

	// Build metadata structure
	metadata := Metadata{
		Name:          "eac",
		Version:       "1.0.0",
		Description:   "Everything as Code - Repository management tooling for CLIE CLI",
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
			CLIEVersion:      ">=1.0.0",
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
			{
				Name:   "docker-socket",
				Target: "/var/run/docker.sock",
				Type:   "bind",
			},
		},
		ExpectedHostImages: []string{
			"mkdocs-render-oci:local",
			tool.GetToolImageWithDefault("structurizr-lite", "structurizr/lite:latest"),
		},
		Env: []EnvVar{
			{Name: "GODOG_SUITE_TAGS"},
			{Name: "GODOG_FORMAT"},
			{Name: "GODOG_OUTPUT_DIR"},
			{Name: "GODOG_REPORT_FORMAT"},
			{Name: "GODOG_REPORT_NAME"},
			{Name: "GODOG_PATHS"},
			{Name: "CLIE_TEST_RUN_ID"},
			{Name: "ANTHROPIC_API_KEY"},
		},
		Metadata: ExtensionMetadata{
			Author:        "Ready to Release Team",
			Repository:    "https://github.com/ready-to-release/eac",
			Documentation: "https://github.com/ready-to-release/eac/tree/main/docs",
			License:       "MIT",
			Tags:          []string{"go", "repository", "modular-monorepo", "tooling", "everything-as-code"},
		},
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&metadata)
	if err != nil {
		log.Errorf("Error marshaling metadata: %v", err)
		return 1
	}

	fmt.Print(string(yamlData))
	return 0
}

// PrintAvailableCommands prints a user-friendly list of available commands.
func PrintAvailableCommands() {
	fmt.Println("EAC - Everything as Code")
	fmt.Println("Repository management tooling for CLIE CLI")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  extension-meta    Display extension metadata (YAML)")
	fmt.Println("  help              Display detailed command help")
	fmt.Println()
	fmt.Println("Use 'help' to see all available commands with descriptions.")
	fmt.Println("Use 'help <command>' for detailed information about a specific command.")
}
