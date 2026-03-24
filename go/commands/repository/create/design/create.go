// Usage: create design <module>
package design

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	design "github.com/ready-to-release/eac/go/commands/repository/design"
	designInternal "github.com/ready-to-release/eac/go/commands/repository/design/helper"
	"github.com/ready-to-release/eac/go/clibase/flags"
	eacConfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type createDesignCommand struct{}

var _ core.SimpleCommandPort = (*createDesignCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&createDesignCommand{},
	}
}

func (c *createDesignCommand) Name() string { return "create design" }

func (c *createDesignCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "create-design",
		Short:         "Generate workspace.dsl for a module using AI",
		Long: "Generates Structurizr DSL workspace files for a module by analyzing its source code using AI.\nThe AI analyzes the source code in go/eac/<module>/ and creates comprehensive architecture\ndocumentation including system context views (external systems and actors), container views\n(major components), and component views (internal structure). All generated workspaces are\nautomatically validated against Structurizr CLI to ensure correct syntax before saving to\nspecs/<module>/.design/workspace.dsl (or custom path via --output). Use --debug to save AI\nprompts and outputs.",
		Notes: "Expected Output:\n- Structurizr DSL workspace file at output path\n- System context, container, and component views\n- Validation results if Docker available\n- Debug logs in out/commands.log if --debug enabled",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save intermediate outputs (prompts, raw AI responses, validation results) to out/commands.log for debugging"},
			{Name: "force", Shorthand: "f", Type: "bool", DefaultValue: "false", Usage: "Overwrite existing workspace.dsl file if it exists"},
			{Name: "output", Shorthand: "o", Type: "string", Usage: "Custom output path for workspace.dsl (default: specs/<module>/.design/workspace.dsl)"},
			{Name: "prompt", Type: "string", Usage: "Custom AI prompt file path"},
			{Name: "skip-validation", Type: "bool", DefaultValue: "false", Usage: "Skip Docker validation (useful when Docker is unavailable)"},
		},
	}
}

func (c *createDesignCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return CreateDesign()
}
// Intent: Create a Structurizr DSL workspace from natural language description using AI
//
// CreateDesign orchestrates the architecture design generation workflow.
var log = logging.C()

// commandFlags defines valid flags for the create design command

func CreateDesign() int {
	return createDesign(defaultDeps())
}

func createDesign(deps *Deps) int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(config.TemplateRoot, "commands", nil, config.Debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	// Create output handler
	out := design.NewOutput()

	// Validate module exists
	if err := validateModuleExists(config, out); err != nil {
		out.Errorf("Error: %v", err)
		return 1
	}

	// Check Docker availability early (before expensive AI generation) unless skipped
	if !config.SkipValidation {
		if err := checkDockerAvailability(config, out); err != nil {
			out.Errorf("Error: %v", err)
			return 1
		}
	} else {
		log.Warnf("Docker validation skipped via --skip-validation flag for module %q", config.Module)
		out.Progress("⚠️  Skipping Docker validation")
	}

	// Load EAC config for path resolution (uses helper with clean fallback for tests)
	cfg := eacConfig.LoadOrNil(config.TemplateRoot)

	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		if cfg == nil {
			// Fallback to default path if config loading fails (e.g., in tests)
			outputPath = filepath.Join(config.TemplateRoot, "specs", config.Module, ".design", "workspace.dsl")
		} else {
			outputPath = filepath.Join(paths.SpecsPath(config.TemplateRoot, config.Module), ".design", "workspace.dsl")
		}
	}

	// Check if output exists (unless --force)
	if !config.Force && fileExists(outputPath) {
		out.Errorf("Error: workspace already exists at %s", outputPath)
		out.Error("Use --force to overwrite")
		return 1
	}

	// Load contract and build prompt
	fullPrompt, err := loadAndBuildPrompt(config, out)
	if err != nil {
		out.Errorf("Error: %v", err)
		return 1
	}

	// Generate and validate with retry
	cleanedOutput, err := generateAndValidate(config, fullPrompt, out)
	if err != nil {
		out.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Write output and report success
	if err := writeOutputAndReportSuccess(config, outputPath, cleanedOutput, out); err != nil {
		out.Errorf("\n❌ Error: %v", err)
		return 1
	}

	return 0
}

// DesignConfig holds configuration for design create command.
type DesignConfig struct {
	Module         string // Module name (e.g., "clie", "commands")
	SourcePath     string // Path to source code (e.g., "go/cli/eac")
	OutputPath     string // Custom output path (empty = default to specs/<module>/.design/workspace.dsl)
	PromptPath     string // Custom AI prompt file path (empty = default prompt)
	Debug          bool
	Force          bool
	SkipValidation bool
	TemplateRoot   string
}

// parseConfig parses command line arguments into DesignConfig.
func parseConfig() (*DesignConfig, error) {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Find and parse arguments
	modulePath, cmdFlags, err := parseCreateCommandArgs(os.Args)
	if err != nil {
		return nil, err
	}

	// Create config with parsed flags
	config := &DesignConfig{
		Debug:          cmdFlags.debug,
		Force:          cmdFlags.force,
		OutputPath:     cmdFlags.outputPath,
		PromptPath:     cmdFlags.promptPath,
		SkipValidation: cmdFlags.skipValidation,
		TemplateRoot:   repoRoot,
	}

	// Use input as-is (moniker)
	config.Module = modulePath

	// Validate module name for security
	if err := designInternal.ValidateModuleName(config.Module); err != nil {
		return nil, fmt.Errorf("invalid module name: %w", err)
	}

	// Load module contracts and validate moniker exists (same as build command)
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load module contracts: %w", err)
	}

	module, exists := moduleReport.Registry.Get(config.Module)
	if !exists {
		return nil, fmt.Errorf("module not found: %s\n\nAvailable modules:\n%s",
			config.Module, formatModuleList(moduleReport))
	}

	// Store validated moniker and source path from contract
	config.Module = module.Moniker
	// Get the buildable package root (go or typescript) for source analysis
	buildableRoot := module.Components.GetBuildableRoot()
	if buildableRoot == "" {
		// Fallback to first available package root
		for _, root := range module.GetComponentRoots() {
			buildableRoot = root
			break
		}
	}
	config.SourcePath = filepath.Join(repoRoot, buildableRoot)

	return config, nil
}

// createFlags holds parsed command-line flags.
type createFlags struct {
	debug          bool
	force          bool
	outputPath     string
	promptPath     string
	skipValidation bool
}

// parseCreateCommandArgs parses args for the create subcommand
// Returns: module path, flags, error.
func parseCreateCommandArgs(args []string) (string, *createFlags, error) {
	// Find command position
	cmdPos := -1
	for i, arg := range args {
		if arg == "create" {
			cmdPos = i
			break
		}
	}

	if cmdPos == -1 {
		return "", nil, fmt.Errorf("invalid command structure")
	}

	// Skip "create" and "design" tokens - start parsing from the module name
	// Args structure: [... "create" "design" <module> <flags>...]
	startPos := cmdPos + 1
	if startPos < len(args) && args[startPos] == "design" {
		startPos++ // Skip "design" subcommand token
	}

	// Parse flags and positional arguments
	cmdFlags := &createFlags{}
	var positionalArgs []string

	// Extract args from startPos onward for flag parsing
	argsToParse := args[startPos:]

	// Use shared flags package for debug flag
	cmdFlags.debug = flags.ParseDebugFlag(argsToParse)
	cmdFlags.force = flags.HasFlag(argsToParse, "--force", "")
	cmdFlags.skipValidation = flags.HasFlag(argsToParse, "--skip-validation", "")

	// Parse value flags and positional args
	for i := 0; i < len(argsToParse); i++ {
		arg := argsToParse[i]

		if arg == "--output" || arg == "-o" {
			// Next argument is the output path
			if i+1 < len(argsToParse) {
				cmdFlags.outputPath = argsToParse[i+1]
				i++ // Skip next arg
			}
		} else if arg == "--prompt" {
			// Next argument is the prompt path
			if i+1 < len(argsToParse) {
				cmdFlags.promptPath = argsToParse[i+1]
				i++ // Skip next arg
			}
		} else if !strings.HasPrefix(arg, "-") {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Validate we have a module name
	if len(positionalArgs) == 0 {
		return "", nil, fmt.Errorf("module name required\n\nUsage: create design <module>\nExample: create design clie")
	}

	return positionalArgs[0], cmdFlags, nil
}

// formatModuleList returns a formatted list of available modules.
func formatModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		// Get first package root for display
		var displayRoot string
		for _, root := range mod.GetComponentRoots() {
			displayRoot = root
			break
		}
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, displayRoot))
	}
	return sb.String()
}

