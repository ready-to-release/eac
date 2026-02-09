// Command: lint
// Short: Lint one or more modules by moniker
// Long: Lint one or more modules by moniker using appropriate linters per module type.
// Long:
// Long: This command runs linters on modules in parallel (by default).
// Long: If no monikers are specified, all modules in the repository are linted.
// Long:
// Long: Expected Output:
// Long:   - Lint logs written to 'out/lint/<module>/lint.log' (one per module)
// Long:   - Structured results at 'out/lint/<module>/lint.json' (linter-specific format)
// Long:   - UoW manifests at 'out/lint/<module>/<component>/uow.manifest.json' (with timing data)
// Long:   - Failed lints are clearly marked with error details
// Long:   - Exit code 0 indicates no lint issues found
// Long:   - Non-zero exit code indicates lint issues found or errors
// Long:
// Long: Example:
// Long:   lint                           # Lint all modules
// Long:   lint eac              # Lint a single module
// Long:   lint --fix                     # Lint with auto-fix
// Long:   lint --skip-cache              # Force full lint, ignore incremental state
// Flag.fix: type=bool, usage=Auto-fix issues where possible
// Flag.config: type=string, usage=Override lint config file path
// Flag.skip-cache: type=bool, usage=Skip incremental cache, force full lint
// Flag.debug: type=bool, usage=Enable debug logs to console
// Flag.tui: type=bool, usage=Enable TUI console (default for local)
// Flag.no-tui: type=bool, usage=Disable TUI console
// Flag.tui-height: type=int, usage=Set TUI console height (3-20, default: 6)
// Flag.ascii: type=bool, usage=Use ASCII-only characters in TUI
// Flag.skip-tui-delay: type=bool, usage=Skip TUI exit delay (exit immediately when done)
// Flag.sequential: type=bool, usage=Run lints sequentially instead of in parallel
// Flag.turbo: type=bool, usage=Enable turbo mode for faster linting (increases parallelism)
// Flag.skip-deps: type=bool, usage=Skip system dependency verification
// Flag.timings: type=bool, usage=Show detailed timing summary
// Flag.dry-run: type=bool, usage=Show what would be linted without running linters
// Args: modules
package lint

import (
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"

	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

var log = logging.C()

func init() {
	registry.Register(Lint)
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
