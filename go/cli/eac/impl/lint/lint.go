package lint

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"

	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

var log = logging.C()
type lintCommand struct{}

var _ core.SimpleCommandPort = (*lintCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&lintCommand{},
	}
}

func (c *lintCommand) Name() string { return "lint" }

func (c *lintCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "lint",
		Short:         "Lint one or more modules by moniker",
		Long:          "Lint one or more modules by moniker using appropriate linters per module type.\n\nThis command runs linters on modules in parallel (by default).\nIf no monikers are specified, all modules in the repository are linted.\n\nExpected Output:\n  - Lint logs written to 'out/lint/<module>/lint.log' (one per module)\n  - Structured results at 'out/lint/<module>/lint.json' (linter-specific format)\n  - UoW manifests at 'out/lint/<module>/<component>/uow.manifest.json' (with timing data)\n  - Failed lints are clearly marked with error details\n  - Exit code 0 indicates no lint issues found\n  - Non-zero exit code indicates lint issues found or errors\n\nExample:\n  lint                           # Lint all modules\n  lint eac              # Lint a single module\n  lint --fix                     # Lint with auto-fix\n  lint --skip-cache              # Force full lint, ignore incremental state",
		Args:          "modules",
		Flags: []core.FlagSpec{
			{Name: "fix", Type: "bool", Usage: "Auto-fix issues where possible"},
			{Name: "config", Type: "string", Usage: "Override lint config file path"},
			{Name: "skip-cache", Type: "bool", Usage: "Skip incremental cache, force full lint"},
			{Name: "debug", Type: "bool", Usage: "Enable debug logs to console"},
			{Name: "tui", Type: "bool", Usage: "Enable TUI console (default for local)"},
			{Name: "no-tui", Type: "bool", Usage: "Disable TUI console"},
			{Name: "tui-height", Type: "int", Usage: "Set TUI console height (3-20, default: 6)"},
			{Name: "ascii", Type: "bool", Usage: "Use ASCII-only characters in TUI"},
			{Name: "skip-tui-delay", Type: "bool", Usage: "Skip TUI exit delay (exit immediately when done)"},
			{Name: "sequential", Type: "bool", Usage: "Run lints sequentially instead of in parallel"},
			{Name: "turbo", Type: "bool", Usage: "Enable turbo mode for faster linting (increases parallelism)"},
			{Name: "skip-deps", Type: "bool", Usage: "Skip system dependency verification"},
			{Name: "timings", Type: "bool", Usage: "Show detailed timing summary"},
			{Name: "dry-run", Type: "bool", Usage: "Show what would be linted without running linters"},
		},
	}
}

func (c *lintCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Lint()
}

// Lint command entry point - lints one or more modules.
func Lint() int {
	args := os.Args[2:] // Skip program name and "lint"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printLintUsage()
		return 0
	}

	// Detect execution environment
	env := environment.Detect()

	// Parse shared flags using flag sets
	shared, err := flags.ParseSharedFlagsWithEnv(flags.LintConfig(), args, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		printLintUsage()
		return 1
	}

	// Parse lint-specific flags from remaining args
	lintFlags, unknownArgs, err := ParseLintSpecificFlags(shared.Remaining)
	if err != nil {
		log.Errorf("Error: %v", err)
		printLintUsage()
		return 1
	}

	// Check for unknown flags
	for _, arg := range unknownArgs {
		if strings.HasPrefix(arg, "--") {
			log.Errorf("Error: unknown flag: %s", arg)
			return 1
		}
	}

	// Determine sequential mode: --roof 1 means sequential
	sequential := shared.MaxConcurrency == 1

	// Validate --turbo and sequential mutual exclusivity
	if shared.Turbo && sequential {
		log.Errorf("Error: --turbo and --roof 1 (sequential) cannot be used together")
		return 1
	}

	// Create command config for framework
	cmdCfg := &cmdframework.CommandConfig{
		Type:           core.ActionLint,
		CommandPath:    "lint",
		OutputDir:      paths.OutLintRelPath,
		Monikers:       shared.Monikers,
		SkipDeps:       shared.SkipDeps,
		Sequential:     sequential,
		Turbo:          shared.Turbo,
		MaxConcurrency: shared.MaxConcurrency,
		ForceRebuild:   shared.CacheConfig.ShouldSkipState(), // Use ForceRebuild for skip-cache flag
		DryRun:         shared.DryRun,
		UseTUI:         shared.UseTUI,
		TUIHeight:      shared.TUIHeight,
		TUIASCIIMode:   shared.TUIASCIIMode,
		TUI3Demo:       shared.TUI3Demo,
		SkipTUIDelay:   shared.SkipTUIDelay,
		ShowTimings:    shared.ShowTimings,
		DebugMode:      shared.Debug,
		CacheConfig:    shared.CacheConfig,
	}

	// Create lint-specific config
	lintCfg := &LintConfig{
		Fix:       lintFlags.Fix,
		Config:    lintFlags.ConfigPath,
		ForceLint: shared.CacheConfig.ShouldSkipState(),
	}

	return RunLintWithFramework(cmdCfg, lintCfg)
}

func printLintUsage() {
	log.Info("Lint one or more modules by moniker")
	log.Info("")
	log.Info("Usage: lint [flags] [module1] [module2] ...")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  module1, module2, ...     Module monikers to lint (lints all if none)")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --fix                     Auto-fix issues where possible")
	log.Info("  --config PATH             Override lint config file path")
	log.Info("  --skip-cache              Skip incremental cache, force full lint")
	log.Info("  --dry-run                 Show what would be linted without running linters")
	log.Info("  --turbo                   Enable turbo mode for faster linting (increases parallelism)")
	log.Info("  --sequential              Run lints sequentially (default: parallel)")
	log.Info("  --skip-deps               Skip system dependency verification")
	log.Info("  --timings                 Show detailed timing summary")
	log.Info("  --debug                   Enable debug logs to console")
	log.Info("  --tui                     Enable TUI console")
	log.Info("  --no-tui                  Disable TUI console")
	log.Info(fmt.Sprintf("  --tui-height N            Set TUI height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --ascii                   Use ASCII-only characters in TUI")
	log.Info("  --skip-tui-delay          Skip TUI exit delay (exit immediately when done)")
	log.Info("  -h, --help                Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  lint                      # Lint all modules")
	log.Info("  lint eac         # Lint a single module")
	log.Info("  lint --fix                # Lint with auto-fix")
	log.Info("  lint --skip-cache         # Force full lint")
}
