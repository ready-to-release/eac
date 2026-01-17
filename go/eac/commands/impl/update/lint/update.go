// Command: update lint
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
// Long:   - Failed lints do not stop execution of remaining modules
// Long:   - Exit code 0 indicates no lint issues found
// Long:   - Non-zero exit code indicates lint issues found or errors
// Long:
// Long: Example:
// Long:   update lint                    # Lint all modules
// Long:   update lint eac-commands       # Lint a single module
// Long:   update lint --fix              # Lint with auto-fix
// Long:   update lint --config my.yml    # Use custom config file
// Flag.fix: type=bool, usage=Auto-fix issues where possible
// Flag.config: type=string, usage=Override lint config file path
// Flag.debug: type=bool, usage=Enable debug logs to console (file logging always enabled)
// Flag.tui: type=bool, usage=Enable TUI console (default for local, errors in CI/container)
// Flag.no-tui: type=bool, usage=Disable TUI console (use plain output)
// Flag.tui-height: type=int, usage=Set TUI console height (3-20, default: 6)
// Flag.sequential: type=bool, usage=Run lints sequentially instead of in parallel
// Flag.skip-deps: type=bool, usage=Skip system dependency verification (golangci-lint, etc.)
// Args: modules
package lint

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	// Import linters package to trigger handler registration via init()
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
	registry.Register(UpdateLint)
}

// UpdateLint command entry point - lints one or more modules
func UpdateLint() int {
	args := os.Args[3:] // Skip program name, "update", and "lint"

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
	debugMode := false
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
				printLintUsage()
				return 1
			}
			i++
			configPath = args[i]
		case "--skip-deps":
			skipDeps = true
		case "--sequential":
			sequential = true
		case "--debug":
			debugMode = true
		case "--tui":
			useTUI = true
			tuiExplicitlySet = true
		case "--no-tui":
			useTUI = false
			tuiExplicitlySet = true
		case "--tui-height":
			if i+1 >= len(args) {
				log.Errorf("Error: --tui-height requires a value")
				printLintUsage()
				return 1
			}
			i++
			var err error
			tuiHeight, err = strconv.Atoi(args[i])
			if err != nil || tuiHeight < 3 || tuiHeight > 20 {
				log.Errorf("Error: --tui-height must be a number between 3 and 20")
				printLintUsage()
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
					printLintUsage()
					return 1
				}
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf("Error: unknown flag: %s", arg)
				printLintUsage()
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

	// Create command config for framework
	cmdCfg := &cmdframework.CommandConfig{
		Type:        cmdframework.CommandTypeLint,
		ActionVerb:  "Linting",
		OutputDir:   paths.OutLintRelPath,
		LogFileName: "lint.log",
		Monikers:    monikers,
		SkipDeps:    skipDeps,
		Sequential:  sequential,
		Layered:     false, // Lint runs in parallel (no dependency ordering needed)
		UseTUI:      useTUI,
		TUIHeight:   tuiHeight,
		DebugMode:   debugMode,
	}

	// Create lint-specific config
	lintCfg := &LintConfig{
		Fix:    fix,
		Config: configPath,
	}

	return RunLintWithFramework(cmdCfg, lintCfg)
}

func printLintUsage() {
	log.Info("Lint one or more modules by moniker")
	log.Info("")
	log.Info("Usage: update lint [flags] [module1] [module2] ...")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  module1, module2, ...     Module monikers to lint (lints all if none specified)")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --fix                     Auto-fix issues where possible")
	log.Info("  --config PATH             Override lint config file path")
	log.Info("  --sequential              Run lints sequentially instead of in parallel")
	log.Info("  --skip-deps               Skip system dependency verification (golangci-lint, etc.)")
	log.Info("  --debug                   Enable debug logs to console (file logging always enabled)")
	log.Info("  --tui                     Enable TUI console (default for local, errors in CI/container)")
	log.Info("  --no-tui                  Disable TUI console (use plain output)")
	log.Info(fmt.Sprintf("  --tui-height N            Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  -h, --help                Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  update lint                      # Lint all modules")
	log.Info("  update lint eac-commands         # Lint a single module")
	log.Info("  update lint --fix                # Lint all modules with auto-fix")
	log.Info("  update lint --config my.yml      # Use custom config file")
}
