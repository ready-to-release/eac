// Package cmdframework provides a shared framework for orchestrated commands.
// It extracts common patterns from build, test, and scan commands into reusable
// phases: init, resolve, verify, execute, and summary.
package cmdframework

import (
	"fmt"
	"io"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/domain/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
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

// CommandType defines the type of command for configuration defaults.
type CommandType string

const (
	CommandTypeBuild CommandType = "build"
	CommandTypeTest  CommandType = "test"
	CommandTypeScan  CommandType = "scan"
	CommandTypeLint  CommandType = "lint"
)

// CommandConfig holds all configuration for a command execution.
// Commands populate this struct and pass it to Run().
type CommandConfig struct {
	// Identity
	Type        CommandType // build, test, scan
	CommandPath string      // Command path for TUI binding (e.g., "build", "test", "work create")
	ActionVerb  string      // "Building", "Testing", "Scanning"
	OutputDir   string      // Relative path: "out/build", "out/test", "out/scan"
	LogFileName string      // "build.log", "test.log", "scan.log"

	// Inputs
	Monikers []string // Module monikers to process (empty = all)

	// Execution Control
	MaxConcurrency int  // 0 = use config default
	Layered        bool // Use dependency layers (build) vs parallel (test/scan)
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
	SkipTUIDelay bool // Skip TUI exit delay (exit immediately when done)
	DebugMode    bool // Enable debug logging to console
	ShowTimings  bool // Show timing breakdown in summary

	// Command-Specific Options
	Extra map[string]interface{}
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
	ExecutionPlan  *repository.ExecutionPlan
	ModuleTypes    map[string]string // moniker -> component types (comma-separated)

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

	// Internal
	tuiWriter   io.Writer
	cleanup     []func()
	initTimings InitTimings
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

// WorkerFunc processes a single module and returns an exit code.
// It receives the execution context, moniker, and a log writer for output.
// Return 0 for success, non-zero for failure.
type WorkerFunc func(ctx *ExecutionContext, moniker string, logWriter io.Writer) int

// UnitWorkerFunc processes a single work unit and returns an exit code.
// It receives the execution context, module moniker, component name, and a log writer.
// Return 0 for success, non-zero for failure.
type UnitWorkerFunc func(ctx *ExecutionContext, module, component string, logWriter io.Writer) int

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

// GetExtra retrieves a command-specific option with type assertion.
// Returns the zero value if the key doesn't exist or type doesn't match.
func (c *CommandConfig) GetExtra(key string) interface{} {
	if c.Extra == nil {
		return nil
	}
	return c.Extra[key]
}

// GetExtraString retrieves a string option from Extra.
func (c *CommandConfig) GetExtraString(key string) string {
	if v, ok := c.Extra[key].(string); ok {
		return v
	}
	return ""
}

// GetExtraBool retrieves a bool option from Extra.
func (c *CommandConfig) GetExtraBool(key string) bool {
	if v, ok := c.Extra[key].(bool); ok {
		return v
	}
	return false
}

// GetExtraInt retrieves an int option from Extra.
func (c *CommandConfig) GetExtraInt(key string) int {
	if v, ok := c.Extra[key].(int); ok {
		return v
	}
	return 0
}

// WriteInit writes a message to the init phase output (TUI or console).
func (ctx *ExecutionContext) WriteInit(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if ctx.Orchestrator != nil && ctx.Config.UseTUI {
		ctx.Orchestrator.SendInitLine(msg)
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
