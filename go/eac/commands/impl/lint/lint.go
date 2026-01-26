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
// Long:   - Lint manifest at 'out/lint/<module>/lint.manifest.json' (with timing data)
// Long:   - Failed lints are clearly marked with error details
// Long:   - Exit code 0 indicates no lint issues found
// Long:   - Non-zero exit code indicates lint issues found or errors
// Long:
// Long: Example:
// Long:   lint                           # Lint all modules
// Long:   lint eac-commands              # Lint a single module
// Long:   lint --fix                     # Lint with auto-fix
// Long:   lint --relint                  # Force full lint, ignore incremental state
// Flag.fix: type=bool, usage=Auto-fix issues where possible
// Flag.config: type=string, usage=Override lint config file path
// Flag.relint: type=bool, usage=Force full lint, ignoring incremental state
// Flag.debug: type=bool, usage=Enable debug logs to console
// Flag.tui: type=bool, usage=Enable TUI console (default for local)
// Flag.no-tui: type=bool, usage=Disable TUI console
// Flag.tui-height: type=int, usage=Set TUI console height (3-20, default: 6)
// Flag.sequential: type=bool, usage=Run lints sequentially instead of in parallel
// Flag.turbo: type=bool, usage=Enable turbo mode for faster linting (increases parallelism)
// Flag.skip-deps: type=bool, usage=Skip system dependency verification
// Flag.timings: type=bool, usage=Show detailed timing summary
// Args: modules
package lint

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	// Import linters package to trigger handler registration via init().
	_ "github.com/ready-to-release/eac/go/eac/commands/impl/update/lint/linters"

	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
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

	// Parse module monikers and flags
	var monikers []string
	fix := false
	configPath := ""
	skipDeps := false
	sequential := false
	turbo := false
	forceRelint := false
	debugMode := false
	showTimings := false
	useTUI := env.ShouldUseTUI()
	tuiExplicitlySet := false
	tuiHeight := tui.DefaultHeight

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--fix":
			fix = true
		case "--config":
			if i+1 >= len(args) {
				log.Errorf("Error: --config requires a value")
				return 1
			}
			i++
			configPath = args[i]
		case "--skip-deps":
			skipDeps = true
		case "--sequential":
			sequential = true
		case "--turbo":
			turbo = true
		case "--relint":
			forceRelint = true
		case "--debug":
			debugMode = true
		case "--timings":
			showTimings = true
		case "--tui":
			useTUI = true
			tuiExplicitlySet = true
		case "--no-tui":
			useTUI = false
			tuiExplicitlySet = true
		case "--tui-height":
			if i+1 >= len(args) {
				log.Errorf("Error: --tui-height requires a value")
				return 1
			}
			i++
			var err error
			tuiHeight, err = strconv.Atoi(args[i])
			if err != nil || tuiHeight < 3 || tuiHeight > 20 {
				log.Errorf("Error: --tui-height must be a number between 3 and 20")
				return 1
			}
		default:
			if strings.HasPrefix(arg, "--config=") {
				configPath = strings.TrimPrefix(arg, "--config=")
			} else if strings.HasPrefix(arg, "--tui-height=") {
				heightStr := strings.TrimPrefix(arg, "--tui-height=")
				var err error
				tuiHeight, err = strconv.Atoi(heightStr)
				if err != nil || tuiHeight < 3 || tuiHeight > 20 {
					log.Errorf("Error: --tui-height must be a number between 3 and 20")
					return 1
				}
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf("Error: unknown flag: %s", arg)
				return 1
			} else {
				monikers = append(monikers, arg)
			}
		}
	}

	// Validate TUI usage
	if err := env.ValidateTUI(tuiExplicitlySet, useTUI); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Validate --turbo and --sequential mutual exclusivity
	if turbo && sequential {
		log.Errorf("Error: --turbo and --sequential cannot be used together")
		return 1
	}

	// Create command config for framework
	cmdCfg := &cmdframework.CommandConfig{
		Type:         cmdframework.CommandTypeLint,
		ActionVerb:   "Linting",
		OutputDir:    paths.OutLintRelPath,
		LogFileName:  "lint.log",
		Monikers:     monikers,
		SkipDeps:     skipDeps,
		Sequential:   sequential,
		Turbo:        turbo,
		ForceRebuild: forceRelint, // Use ForceRebuild for relint flag
		Layered:      false,       // Lint runs in parallel (no dependency ordering needed)
		UseTUI:       useTUI,
		TUIHeight:    tuiHeight,
		ShowTimings:  showTimings,
		DebugMode:    debugMode,
	}

	// Create lint-specific config
	lintCfg := &LintConfig{
		Fix:       fix,
		Config:    configPath,
		ForceLint: forceRelint,
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
	log.Info("  --relint                  Force full lint, ignore incremental state")
	log.Info("  --turbo                   Enable turbo mode for faster linting (increases parallelism)")
	log.Info("  --sequential              Run lints sequentially (default: parallel)")
	log.Info("  --skip-deps               Skip system dependency verification")
	log.Info("  --timings                 Show detailed timing summary")
	log.Info("  --debug                   Enable debug logs to console")
	log.Info("  --tui                     Enable TUI console")
	log.Info("  --no-tui                  Disable TUI console")
	log.Info(fmt.Sprintf("  --tui-height N            Set TUI height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  -h, --help                Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  lint                      # Lint all modules")
	log.Info("  lint eac-commands         # Lint a single module")
	log.Info("  lint --fix                # Lint with auto-fix")
	log.Info("  lint --relint             # Force full lint")
}
