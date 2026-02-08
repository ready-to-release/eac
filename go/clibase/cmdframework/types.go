// Package cmdframework provides a shared framework for orchestrated commands.
// It extracts common patterns from build, test, and scan commands into reusable
// phases: init, resolve, verify, execute, and summary.
package cmdframework

import (
	"context"
	"fmt"
	"io"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ErrInformationalExit signals that the command should exit with code 1,
// but all necessary user-facing output has already been written.
// This prevents the framework from logging an additional error message,
// creating a cleaner user experience for expected validation failures.
type ErrInformationalExit struct {
	// Reason is a short internal identifier (not shown to user)
	Reason string
}

func (e ErrInformationalExit) Error() string {
	return e.Reason
}

// NewInformationalExit creates an ErrInformationalExit with the given reason.
func NewInformationalExit(reason string) ErrInformationalExit {
	return ErrInformationalExit{Reason: reason}
}

// CommandConfig holds all configuration for a command execution.
// Commands populate this struct and pass it to Run().
type CommandConfig struct {
	// Identity
	Type        core.ActionType // build, test, scan
	CommandPath string      // Command path for TUI binding (e.g., "build", "test", "work create")
	OutputDir   string      // Relative path: "out/build", "out/test", "out/scan"

	// Inputs
	Monikers []string // Module monikers to process (empty = all)

	// Execution Control
	MaxConcurrency int  // 0 = use config default
	Sequential     bool // Force sequential execution (maxConcurrency=1)
	Turbo          bool // Enable turbo mode (+2 parallel workers)
	DryRun         bool // Skip actual execution

	// Dependency Handling
	IncludeDepm     bool // Include module dependencies in execution
	SkipDeps        bool // Skip system dependency verification
	SkipDepm        bool // Skip module dependency (build artifact) validation
	UseExistingDepm bool // Use pre-built dependencies without rebuilding (CI mode)

	// Resolution Control
	SkipResolve bool // Skip module resolution (command provides its own execution plan)

	// Incremental Detection
	ForceRebuild bool // Bypass incremental change detection

	// UI
	UseTUI       bool // Enable TUI console
	TUIHeight    int  // TUI height (0 = default)
	TUIASCIIMode bool // Use ASCII-only characters in TUI (--ascii flag)
	TUI3Demo     bool // Use experimental tui3 layout (--demo flag)
	SkipTUIDelay bool // Skip TUI exit delay (exit immediately when done)
	DebugMode    bool // Enable debug logging to console
	ShowTimings  bool // Show timing breakdown in summary

	// Cache Configuration
	CacheConfig *cache.Config // Fine-grained cache control (--skip-cache flag)

	// Command-Specific Typed Configuration (replaces Extra map)
	// Each command sets its own config struct; others remain nil.
	SkipMonikers   []string             // Modules to skip during resolution (set by scan)
	UnitSpecsCache []workunit.UnitSpec  // Cached UnitSpecs for TUI tree building

	// Command-specific configs are stored as interface{} to avoid import cycles.
	// Each command package defines its own typed config struct and uses type assertion.
	BuildCmdConfig   interface{} // *build.BuildConfig (set by build command)
	BuildCmdContext  interface{} // *build.buildContext (set by build command)
	BuildWorkSpecs   []workunit.UnitSpec // Cached component work specs for build
	TestCmdConfig    interface{} // *test.TestFrameworkConfig (set by test command)
	ScanCmdConfig    interface{} // *scan.ScanFrameworkConfig (set by scan command)
	ScanCmdWorker    interface{} // scan.ScanWorker (set by scan command)
	MultiScanConfig  interface{} // *scan.MultiScanConfig (set by scan command)
	ScanCmdContext   interface{} // *scan.scanContext (set by scan command)
	LintCmdConfig    interface{} // *lint.LintConfig (set by lint command)
	LintCmdContext   interface{} // *lint.lintContext (set by lint command)
	AnalysisType     string      // AI summary analysis type filter (set by ai-summary command)
}

// ActionVerb returns the present-continuous verb for this command's action type
// (e.g., "Building", "Testing"). Derived from the ActionDescriptor registry.
func (c *CommandConfig) ActionVerb() string {
	if d, ok := core.GetActionDescriptor(c.Type); ok {
		return d.Verb
	}
	return string(c.Type)
}

// LogFileName returns the log file name for this command's action type
// (e.g., "build.log", "test.log"). Derived from the ActionDescriptor registry.
func (c *CommandConfig) LogFileName() string {
	if d, ok := core.GetActionDescriptor(c.Type); ok {
		return d.LogFile
	}
	return string(c.Type) + ".log"
}

// ExecutionContext provides access to loaded config and state during execution.
// It's populated by the framework phases and passed to worker functions.
type ExecutionContext struct {
	// Configuration
	Config *CommandConfig

	// Loaded State (populated by phaseInit)
	WorkspaceRoot string
	EACConfig     *config.EACConfig
	RepoConfig    *config.RepositoryConfig
	Orchestrator  *orchestrator.Orchestrator

	// Module State (populated by phaseResolve)
	ModuleReport   *reports.ModuleContractReport
	ModuleRegistry *modules.Registry
	ScopeMonikers  []string              // Modules in execution scope (requested + expanded deps)
	ComponentTypesDisplay map[string]string // moniker -> component types (comma-separated)

	// Verification State (populated by phaseVerify)
	InitSummary *initsummary.Summary

	// Execution State
	StartTime time.Time
	Results   []orchestrator.WorkResult

	// Component-level results (populated by component-based execution)
	// UnitResults contains raw unit results for detailed display
	UnitResults []orchestrator.UnitResult
	// ModuleResultSets contains unit results grouped by module
	ModuleResultSets []orchestrator.ModuleResultSet

	// Incremental summary builder (receives results as components complete)
	SummaryBuilder *SummaryBuilder

	// TUI Hooks (nil for console-only mode)
	TUIHooks core.TUIHooks

	// OutputBuffer captures stdout/stderr during TUI mode.
	// In console mode, this is a passthrough that does nothing.
	OutputBuffer core.OutputBufferPort

	// Internal
	tuiWriter       io.Writer
	cleanup         []func()
	initTimings     InitTimings
	asyncDepsResult *asyncDepsCheck // Background deps verification (started in phaseInitDeferred)
}

// initOutputBuffer sets up the output buffer based on TUI mode.
// For TUI mode, it creates a capturing buffer. For console mode, a passthrough.
func (ctx *ExecutionContext) initOutputBuffer() {
	if ctx.Config.UseTUI {
		ctx.OutputBuffer = output.NewBuffer()
		if err := ctx.OutputBuffer.Start(); err != nil {
			log.Warnf("Failed to start output buffer: %v", err)
			ctx.OutputBuffer = output.NewPassthrough()
		}
	} else {
		ctx.OutputBuffer = output.NewPassthrough()
	}
}

// stopOutputBuffer stops the output buffer and flushes any captured content.
func (ctx *ExecutionContext) stopOutputBuffer() {
	if ctx.OutputBuffer != nil {
		ctx.OutputBuffer.Stop()
	}
}

// InitTimings tracks duration of initialization phases for boot-style output.
type InitTimings struct {
	ConfigLoad      time.Duration
	ToolInit        time.Duration
	ModuleDiscovery time.Duration
	ExecutionOrder  time.Duration
	ChangeDetection time.Duration
	DepsVerify      time.Duration
}

// CommandWorkerFunc processes a single module and returns an exit code.
// It receives a cancellation context, the execution context, moniker, and a log writer.
// Return 0 for success, non-zero for failure.
type CommandWorkerFunc func(goCtx context.Context, ctx *ExecutionContext, moniker string, logWriter io.Writer) int

// UnitWorkerFunc processes a single work unit and returns an exit code.
// It receives a cancellation context, the execution context, module moniker, component name, and a log writer.
// Return 0 for success, non-zero for failure.
type UnitWorkerFunc func(goCtx context.Context, ctx *ExecutionContext, module, component string, logWriter io.Writer) int

// PhaseHook is called at specific points during command execution.
// Return an error to abort execution.
type PhaseHook func(ctx *ExecutionContext) error

// Hooks allows commands to customize framework behavior at key points.
type Hooks struct {
	// AfterInit is called after config loading and orchestrator setup
	AfterInit PhaseHook

	// AfterResolve is called after module resolution and execution planning
	AfterResolve PhaseHook

	// BeforeExecute is called just before worker execution begins
	BeforeExecute PhaseHook

	// AfterExecute is called after all workers complete, before summary
	AfterExecute PhaseHook

	// CustomSummary generates command-specific summary data (optional)
	// If nil, uses default summary generation
	CustomSummary func(ctx *ExecutionContext) *initsummary.Summary
}


// WriteInit writes a message to the init phase output (TUI or console).
// When TUI is enabled but not yet started, messages also go to console
// so they're visible if the command exits before TUI starts.
func (ctx *ExecutionContext) WriteInit(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if ctx.Orchestrator != nil && ctx.Config.UseTUI {
		ctx.Orchestrator.SendInitLine(msg)
		// Also print to console if TUI hasn't started yet
		// This ensures messages are visible if command exits before TUI starts
		if !ctx.Orchestrator.IsTUIStarted() {
			log.Info(msg)
		}
	} else {
		log.Info(msg)
	}
}

// WriteStatus writes a boot-style status line: "[ ok ] message" or "[    ] message".
// Use ok=true for completed steps, ok=false for in-progress steps.
func (ctx *ExecutionContext) WriteStatus(ok bool, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	var prefix string
	if ok {
		prefix = "[ ok ]"
	} else {
		prefix = "[    ]"
	}
	ctx.WriteInit("%s %s", prefix, msg)
}

// SetChangeDetectionTiming sets the change detection timing for boot-style output.
func (ctx *ExecutionContext) SetChangeDetectionTiming(d time.Duration) {
	ctx.initTimings.ChangeDetection = d
}

// AddCleanup registers a function to be called during cleanup.
func (ctx *ExecutionContext) AddCleanup(fn func()) {
	ctx.cleanup = append(ctx.cleanup, fn)
}

// Cleanup runs all registered cleanup functions in reverse order.
func (ctx *ExecutionContext) Cleanup() {
	for i := len(ctx.cleanup) - 1; i >= 0; i-- {
		ctx.cleanup[i]()
	}
}
