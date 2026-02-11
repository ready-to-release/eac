// Usage: update ai-summary [module...]
package aisummary

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type updateAISummaryCommand struct{}

var _ core.SimpleCommandPort = (*updateAISummaryCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updateAISummaryCommand{},
	}
}

func (c *updateAISummaryCommand) Name() string { return "update ai-summary" }

func (c *updateAISummaryCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-ai-summary",
		Short:         "Generate AI-powered analysis summaries for modules",
		Long:          "Analyzes modules using AI to generate comprehensive status summaries.\nThe analysis covers architecture (DSL), specifications (Gherkin),\nsource code, and documentation. Each analysis type produces a\nseparate status file in the output directory.\n\nAnalysis types:\n  - ai-dsl: Analyzes Structurizr DSL architecture files\n  - ai-specs: Analyzes Gherkin BDD specifications\n  - ai-source: Analyzes source code (depends on dsl and specs)\n  - ai-docs: Analyzes documentation files\n\nExpected Output:\n  - out/ai-summary/<module>/dsl-status.md\n  - out/ai-summary/<module>/specs-status.md\n  - out/ai-summary/<module>/source-status.md\n  - out/ai-summary/<module>/docs-status.md\n\nExample:\n  update ai-summary                  # Analyze all modules\n  update ai-summary eac          # Analyze single module\n  update ai-summary --type=dsl core  # Analyze only DSL for core",
		Flags: []core.FlagSpec{
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Enable debug logging"},
			{Name: "type", Shorthand: "t", Type: "string", Usage: "Specific analysis type (dsl, specs, source, docs)"},
			{Name: "skip-cache", Type: "bool", Usage: "Force regeneration even if cached"},
			{Name: "dry-run", Type: "bool", Usage: "Show what would be analyzed without executing"},
			{Name: "tui", Type: "bool", Usage: "Enable TUI console"},
			{Name: "no-tui", Type: "bool", Usage: "Disable TUI console"},
			{Name: "turbo", Type: "bool", Usage: "Enable turbo mode (+2 parallel workers)"},
		},
	}
}

func (c *updateAISummaryCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdateAISummary()
}

var log = logging.C()
// UpdateAISummary is the entry point for the ai-summary command.
func UpdateAISummary() int {
	args := os.Args[3:] // Skip program name, "update", and "ai-summary"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printUsage()
		return 0
	}

	// Detect execution environment
	env := environment.Detect()

	// Parse flags
	config, err := parseConfig(args, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		printUsage()
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("Error: failed to load module contracts: %v", err)
		return 1
	}

	// Expand monikers to all modules if none specified
	monikers := config.Monikers
	if len(monikers) == 0 {
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Validate specified monikers
	for _, moniker := range monikers {
		if _, exists := moduleReport.Registry.Get(moniker); !exists {
			log.Errorf("Error: module not found: %s", moniker)
			return 1
		}
	}

	// Create command config for framework
	cmdCfg := &cmdframework.CommandConfig{
		Type:           core.ActionAISummary,
		CommandPath:    "update ai-summary",
		OutputDir:      paths.OutAISummaryRelPath,
		Monikers:       monikers,
		SkipDeps:       true, // AI analysis doesn't need system deps
		SkipDepm:       true, // No module dependencies
		ForceRebuild:   config.SkipCache,
		MaxConcurrency: 2, // AI tasks are token-limited, not CPU/RAM-limited
		Turbo:          config.Turbo,
		DryRun:         config.DryRun,
		UseTUI:         config.UseTUI,
		TUIHeight:      6,
		DebugMode:      config.Debug,
	}

	// Store analysis type filter
	cmdCfg.AnalysisType = config.AnalysisType

	return RunAISummaryWithFramework(cmdCfg)
}

// Config holds configuration for the ai-summary command.
type Config struct {
	Monikers     []string
	AnalysisType string // Empty = all types, or specific type (dsl, specs, source, docs)
	Debug        bool
	SkipCache    bool
	DryRun       bool
	UseTUI       bool
	Turbo        bool
}

// parseConfig parses command line arguments into Config.
func parseConfig(args []string, env *environment.Env) (*Config, error) {
	config := &Config{
		UseTUI: env.IsLocalConsole, // TUI on by default for local
	}

	var positionalArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--debug" || arg == "-d":
			config.Debug = true

		case arg == "--skip-cache":
			config.SkipCache = true

		case arg == "--dry-run":
			config.DryRun = true

		case arg == "--tui":
			config.UseTUI = true

		case arg == "--no-tui":
			config.UseTUI = false

		case arg == "--turbo":
			config.Turbo = true

		case strings.HasPrefix(arg, "--type="):
			config.AnalysisType = strings.TrimPrefix(arg, "--type=")
			if !isValidAnalysisType(config.AnalysisType) {
				return nil, fmt.Errorf("invalid analysis type: %s (valid: dsl, specs, source, docs)", config.AnalysisType)
			}

		case arg == "--type" || arg == "-t":
			if i+1 < len(args) {
				config.AnalysisType = args[i+1]
				i++
				if !isValidAnalysisType(config.AnalysisType) {
					return nil, fmt.Errorf("invalid analysis type: %s (valid: dsl, specs, source, docs)", config.AnalysisType)
				}
			}

		case !strings.HasPrefix(arg, "-"):
			positionalArgs = append(positionalArgs, arg)

		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	config.Monikers = positionalArgs
	return config, nil
}

// isValidAnalysisType checks if the analysis type is valid.
func isValidAnalysisType(t string) bool {
	switch t {
	case "", "dsl", "specs", "source", "docs":
		return true
	default:
		return false
	}
}

// parseDebugFlag checks for debug flag in args (compatibility with shared flags).
func parseDebugFlag(args []string) bool {
	return flags.ParseDebugFlag(args)
}

func printUsage() {
	log.Info("Generate AI-powered analysis summaries for modules")
	log.Info("")
	log.Info("Usage: update ai-summary [flags] [module...]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  module...     Module monikers to analyze (analyzes all if none specified)")
	log.Info("")
	log.Info("Flags:")
	log.Info("  -d, --debug       Enable debug logging")
	log.Info("  -t, --type TYPE   Specific analysis type (dsl, specs, source, docs)")
	log.Info("  --skip-cache      Force regeneration even if cached")
	log.Info("  --dry-run         Show what would be analyzed without executing")
	log.Info("  --tui             Enable TUI console (default for local)")
	log.Info("  --no-tui          Disable TUI console")
	log.Info("  --turbo           Enable turbo mode (+2 parallel workers)")
	log.Info("  -h, --help        Show this help message")
	log.Info("")
	log.Info("Analysis Types:")
	log.Info("  dsl       Analyze Structurizr DSL architecture files")
	log.Info("  specs     Analyze Gherkin BDD specifications")
	log.Info("  source    Analyze source code (depends on dsl and specs)")
	log.Info("  docs      Analyze documentation files")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/ai-summary/<module>/dsl-status.md")
	log.Info("  out/ai-summary/<module>/specs-status.md")
	log.Info("  out/ai-summary/<module>/source-status.md")
	log.Info("  out/ai-summary/<module>/docs-status.md")
	log.Info("")
	log.Info("Examples:")
	log.Info("  update ai-summary                   # Analyze all modules")
	log.Info("  update ai-summary eac           # Analyze single module")
	log.Info("  update ai-summary --type=dsl core   # Analyze only DSL for core")
}
