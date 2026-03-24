package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
)

var log = logging.C()

type testCommand struct{}

var _ core.SimpleCommandPort = (*testCommand)(nil)

func (c *testCommand) Name() string { return "test" }

func (c *testCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "test",
		Short:         "Test one or more modules by moniker",
		Long: "Test one or more modules by moniker using suite-based filtering.\n\nThis command discovers tests, applies inference rules (e.g., Go tests default to @L1),\nfilters by suite tags, and runs matching tests with consistent summary output.\n\nUse --suite to select which tests to run. The default runs suites not marked\nas extended_suite in config (typically unit + integration).",
		Notes: "Expected Output:\n  - Test execution results with pass/fail status\n  - Detailed test summary table showing modules, packages, and assertions\n  - Test logs written to out/test/<module>/ directory\n  - Exit code 0 if all tests pass, non-zero on failure",
		Examples: []string{
			"eac test eac-cli                     # Test single module",
			"eac test core clie                   # Test multiple modules",
			"eac test                             # Test all modules",
			"eac test eac-cli --suite acceptance  # Run acceptance tests only",
		},
		Args:          "modules",
		Flags: []core.FlagSpec{
			{Name: "suite", Type: "string", Usage: "Filter tests by suite (default: non-extended suites from config)"},
			{Name: "coverage", Type: "bool", Usage: "Enable coverage reporting"},
			{Name: "skip-deps", Type: "bool", Usage: "Skip dependency checks before running tests"},
			{Name: "skip-depm", Type: "bool", Usage: "Skip module dependency build artifact validation"},
			{Name: "list-only", Type: "bool", Usage: "List tests without running them"},
			{Name: "timings", Type: "bool", Usage: "Show detailed timing summary after tests complete"},
			{Name: "debug", Type: "bool", Usage: "Enable debug logs to console (file logging always enabled)"},
			{Name: "tui", Type: "bool", Usage: "Enable TUI console (default for local console mode)"},
			{Name: "no-tui", Type: "bool", Usage: "Disable TUI console (use plain output)"},
			{Name: "sequential", Type: "bool", Usage: "Run tests sequentially instead of parallel"},
			{Name: "turbo", Type: "bool", Usage: "Enable turbo mode for faster testing (increases parallelism)"},
			{Name: "skip-cache", Type: "bool", Usage: "Skip incremental cache, force full test run"},
			{Name: "tui-height", Type: "int", Usage: "Set TUI console height (3-20, default: 6)"},
			{Name: "ascii", Type: "bool", Usage: "Use ASCII-only characters in TUI"},
			{Name: "skip-tui-delay", Type: "bool", Usage: "Skip TUI exit delay (exit immediately when done)"},
		},
	}
}

func (c *testCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Test()
}

// TestConfig holds test execution configuration.
type TestConfig struct {
	Monikers     []string
	SuiteName    string
	Coverage     bool
	SkipDeps     bool // Skip system dependency verification (--skip-deps)
	SkipDepm     bool // Skip dependency module build artifact validation (--skip-depm)
	ListOnly     bool
	ShowTimings  bool
	DebugMode    bool
	UseTUI       bool
	TUIHeight    int
	TUIASCIIMode bool // --ascii flag for ASCII-only TUI
	TUI3Demo     bool // --demo flag for experimental tui3 layout
	SkipTUIDelay bool // --skip-tui-delay flag to exit immediately when done
	Parallel     bool
	Turbo        bool // --turbo flag to increase parallelism
	Roof         int  // --roof flag to limit max parallel capacity
	ForceRetest  bool // --skip-cache flag to bypass incremental testing

	// Cache Configuration
	CacheConfig *cache.Config // Fine-grained cache control
}

// Test is the unified entry point for testing modules.
func Test() int {
	args := os.Args[2:] // Skip program name and "test"

	// Check for subcommands that should be handled separately
	if len(args) > 0 {
		switch args[0] {
		case "suite", "list-suites", "debug":
			// These are handled by their own registered commands
			return 0
		case "--help", "-h":
			printTestUsage()
			return 0
		}
	}

	// Parse arguments
	cfg := parseTestArgs(args)
	if cfg == nil {
		return 1
	}

	// Convert to framework config
	cmdCfg := &cmdframework.CommandConfig{
		Type:           core.ActionTest,
		CommandPath:    "test",
		OutputDir:      paths.OutTestRelPath,
		Monikers:       cfg.Monikers,
		MaxConcurrency: cfg.Roof, // 0 = auto-detect
		Sequential:     !cfg.Parallel,
		Turbo:          cfg.Turbo,
		SkipDeps:       cfg.SkipDeps,
		SkipDepm:       cfg.SkipDepm,
		ForceRebuild:   cfg.ForceRetest,
		UseTUI:         cfg.UseTUI,
		TUIHeight:      cfg.TUIHeight,
		TUIASCIIMode:   cfg.TUIASCIIMode,
		TUI3Demo:       cfg.TUI3Demo,
		SkipTUIDelay:   cfg.SkipTUIDelay,
		DebugMode:      cfg.DebugMode,
		ShowTimings:    cfg.ShowTimings,
		SkipResolve:    true,            // Test command manages its own execution plan
		CacheConfig:    cfg.CacheConfig, // Fine-grained cache control
	}

	testCfg := &TestFrameworkConfig{
		SuiteName:   cfg.SuiteName,
		Coverage:    cfg.Coverage,
		ForceRetest: cfg.ForceRetest,
		Parallel:    cfg.Parallel,
		ListOnly:    cfg.ListOnly,
	}

	return RunTestWithFramework(cmdCfg, testCfg)
}

// parseTestArgs parses command line arguments into TestConfig.
func parseTestArgs(args []string) *TestConfig {
	// Detect execution environment for TUI defaults
	env := environment.Detect()

	// Parse test-specific flags FIRST (before shared flags).
	// This ensures --suite <value> is consumed before the shared parser
	// separates the value into positional args (same pattern as scan.go:78).
	testFlags, remainingAfterTest, err := ParseTestSpecificFlags(args)
	if err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}

	// Parse shared flags from remaining args
	shared, err := flags.ParseSharedFlagsWithEnv(flags.TestConfig(), remainingAfterTest, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}

	// Check for unknown flags
	for _, arg := range shared.Remaining {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			log.Errorf("unknown flag: %s", arg)
			log.Errorf("Valid flags: --suite, --coverage, --skip-deps, --skip-depm, --list-only, --timings, --debug, --tui, --no-tui, --tui-height, --turbo, --roof, --skip-cache, --dry-run")
			return nil
		}
	}

	// Determine parallel mode: --roof 1 means sequential
	parallel := shared.MaxConcurrency != 1

	// Validate --turbo and sequential mutual exclusivity
	if shared.Turbo && !parallel {
		log.Errorf("Error: --turbo and --roof 1 (sequential) cannot be used together")
		return nil
	}

	cfg := &TestConfig{
		// From shared flags
		Monikers:     shared.Monikers,
		SkipDeps:     shared.SkipDeps,
		SkipDepm:     shared.SkipDepm,
		ShowTimings:  shared.ShowTimings,
		DebugMode:    shared.Debug,
		UseTUI:       shared.UseTUI,
		TUIHeight:    shared.TUIHeight,
		TUIASCIIMode: shared.TUIASCIIMode,
		TUI3Demo:     shared.TUI3Demo,
		SkipTUIDelay: shared.SkipTUIDelay,
		Turbo:        shared.Turbo,
		Roof:         shared.MaxConcurrency,
		ForceRetest:  shared.CacheConfig.ShouldSkipState(),
		Parallel:     parallel,
		CacheConfig:  shared.CacheConfig,

		// From test-specific flags
		SuiteName: testFlags.SuiteName,
		Coverage:  testFlags.Coverage,
		ListOnly:  testFlags.ListOnly,
	}

	return cfg
}

// newCommand creates a new exec.Cmd (abstraction for testing).
// Kept as exec.Command for test mock injection.
var newCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func printTestUsage() {
	log.Info("Test one or more modules by moniker")
	log.Info("")
	log.Info("Usage: clie eac test [module1] [module2] ... [options]")
	log.Info("")
	log.Info("Options:")
	log.Info("  --suite <name>         Filter tests by suite (default: unit+integration)")
	log.Info("  --coverage             Generate coverage reports (coverage.out, coverage.json)")
	log.Info("  --skip-deps            Skip system dependency verification before running tests")
	log.Info("  --skip-depm            Skip module dependency build artifact validation")
	log.Info("  --list-only            List tests that would run without executing them")
	log.Info("  --timings              Show detailed timing summary")
	log.Info("  --debug                Enable debug logs to console (file logging always enabled)")
	log.Info("  --no-tui               Disable TUI console (TUI is default for local console)")
	log.Info(fmt.Sprintf("  --tui-height N         Set TUI console height (3-20, default: %d)", tui.DefaultHeight))
	log.Info("  --ascii                Use ASCII-only characters in TUI")
	log.Info("  --skip-tui-delay       Skip TUI exit delay (exit immediately when done)")
	log.Info("  --turbo                Enable turbo mode for faster testing (increases parallelism)")
	log.Info("  --roof N               Limit max parallel capacity to N (default: auto-detect from CPU/RAM)")
	log.Info("  --sequential           Run tests sequentially instead of in parallel")
	log.Info("  --skip-cache           Skip incremental cache, force full test run")
	log.Info("")
	log.Info("Available suites:")
	log.Info("  unit                     L0-L1 tests (fast unit tests)")
	log.Info("  integration              L2 tests (Docker-based emulated tests)")
	log.Info("  acceptance               L3 tests (production-like tests)")
	log.Info("  production-verification  L4+PIV tests (production smoke)")
	log.Info("")
	log.Info("Composite suites (use + to combine):")
	log.Info("  unit+integration+acceptance           # Run multiple suites together")
	log.Info("")
	log.Info("Examples:")
	log.Info("  clie eac test                          # Test all modules (default suites)")
	log.Info("  clie eac test eac-cli             # Test single module")
	log.Info("  clie eac test clie core         # Test multiple modules")
	log.Info("  clie eac test --suite acceptance       # Run acceptance suite")
	log.Info("  clie eac test eac-cli --no-tui    # Disable TUI display")
}

