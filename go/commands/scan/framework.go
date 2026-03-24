// Package scan provides the scan command implementation using cmdframework.
package scan

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/caching"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/evidence"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register scan component-level execution support
	cmdframework.RegisterUnitProvider(core.ActionScan, ResolveScanUnitSpecs)
	cmdframework.RegisterUnitWorker(core.ActionScan, scanUnitWorker)
	cmdframework.SetUoWCountProvider(getScanUoWCount)
}

// NOTE: Scanner-specific semaphores removed - parallelism is now controlled by
// the orchestrator's weighted semaphore system via component weights in getScanWeight()

// log is declared in scan.go

// ScanFrameworkConfig holds scan-specific configuration for the framework.
type ScanFrameworkConfig struct {
	// Scanner identity
	ScannerType  evidence.ScannerType
	ScannerName  string
	ScannerEmoji string

	// Scanner-specific arguments (passed to worker)
	CustomArgs interface{}

	// Execution state
	ScanOutDir  string
	GitCommit   string
	ScanResults map[string]*ScanModuleResult
}

// ScanModuleResult holds scan results for a single module.
type ScanModuleResult struct {
	Moniker      string
	Success      bool
	EvidencePath string
	ErrorMessage string
	Duration     time.Duration
}

// scanContext holds scan-specific state during execution.
type scanContext struct {
	cachedModules     map[string]bool      // Modules that are up-to-date (aggregated from UoWs for TUI)
	cacheTimes        map[string]time.Time // Module-level cache times for TUI
	moduleInputHashes map[string]string    // Pre-computed hashes for cache consistency
	scanResults       map[string]bool      // module -> passed (true = all scanners passed)
	mu                sync.Mutex           // Protects scanResults map for concurrent access

	// UoW-level cache tracking
	cachedUoWs    map[string]bool      // UoW longname -> cached
	uowCacheTimes map[string]time.Time // UoW longname -> cache time
	tracker       *coreoutput.InMemoryTracker
}

// ScanWorker is the function signature for scanner-specific work.
// It receives the module contract and scan config, returns findings and error.
type ScanWorker func(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, scanCfg *ScanFrameworkConfig, logWriter io.Writer) (interface{}, error)

// RunScanWithFramework executes a scan using the cmdframework.
// This provides parallel execution, TUI support, and consistent output.
func RunScanWithFramework(cmdCfg *cmdframework.CommandConfig, scanCfg *ScanFrameworkConfig, worker ScanWorker) int {
	// Store scan config in typed fields for access in hooks/worker
	cmdCfg.ScanCmdConfig = scanCfg
	cmdCfg.ScanCmdWorker = worker
	scanCfg.ScanResults = make(map[string]*ScanModuleResult)

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:    scanAfterInit,
		AfterExecute: scanAfterExecute,
	}

	// Register deps verifier for scanner tools
	cmdframework.SetDepsVerifier(scanDepsVerifier)

	return cmdframework.Run(cmdCfg, nil, hooks)
}

// scanAfterInit handles scan-specific initialization.
func scanAfterInit(ctx *cmdframework.ExecutionContext) error {
	scanCfg, ok := ctx.Config.ScanCmdConfig.(*ScanFrameworkConfig)
	if !ok {
		return fmt.Errorf("ScanCmdConfig not found or wrong type")
	}

	// Set security output directory from config
	scanCfg.ScanOutDir = ctx.EACConfig.Repository.Paths.Out.Scan

	// Get git commit
	scanCfg.GitCommit = getGitCommit(ctx.WorkspaceRoot)

	// Build init summary
	buildScanInitSummary(ctx, scanCfg)

	return nil
}

// scanAfterExecute handles post-scan tasks.
// Note: UoW manifests are written immediately in writeUoWScanManifest via RecordComplete.
func scanAfterExecute(ctx *cmdframework.ExecutionContext) error {
	// Assert all UoWs have valid manifests (skip in dry-run mode)
	if !ctx.Config.DryRun {
		if err := assertScanManifestsExist(ctx); err != nil {
			return err
		}
	}

	return nil
}

// assertScanManifestsExist verifies that all executed UoWs have valid manifests.
func assertScanManifestsExist(ctx *cmdframework.ExecutionContext) error {
	return cmdframework.AssertManifestsExist(ctx, "scan", ResolveScanUnitSpecs(ctx))
}

// buildScanInitSummary creates the init summary for scan commands.
func buildScanInitSummary(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig) {
	// Use scanner name in command field for display
	summary := initsummary.New(fmt.Sprintf("scan-%s", scanCfg.ScannerType)).
		SetRequest(ctx.Config.Monikers, ctx.GetExecutionMonikers()).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			DebugMode: ctx.Config.DebugMode,
			UseTUI:    ctx.Config.UseTUI,
		}).
		SetOutputDir(scanCfg.ScanOutDir)

	ctx.InitSummary = summary
}

// scannerImageConfigID maps a scanner type to its Docker image config ID.
// Multiple scanner types may share the same Docker image (e.g., all Trivy-based
// scanners use the "trivy-sbom" image entry in the security config).
func scannerImageConfigID(scannerType evidence.ScannerType) string {
	switch scannerType {
	case evidence.ScannerSBOM:
		return "trivy-sbom"
	case evidence.ScannerVuln:
		return "trivy-vuln"
	case evidence.ScannerSecrets, evidence.ScannerIaC, evidence.ScannerCompliance:
		return "trivy-sbom" // All Trivy-based scanners use same image
	case evidence.ScannerSAST:
		return "semgrep"
	case evidence.ScannerDAST:
		return "zap"
	default:
		return ""
	}
}

// getScannerImage returns the Docker image for a scanner ID from Security config.
// Returns empty string if scanner not found - caller should handle this as an error.
func getScannerImage(ctx *cmdframework.ExecutionContext, scannerID string) string {
	if ctx.EACConfig == nil || ctx.EACConfig.Security == nil {
		return ""
	}
	scanner, ok := ctx.EACConfig.Security.GetScanner(scannerID)
	if !ok {
		return ""
	}
	return scanner.FullImage()
}

// GetDockerImage returns the appropriate Docker image for a scanner type.
// Uses the eac-security contract to resolve scanner images.
func GetDockerImage(ctx *cmdframework.ExecutionContext, scannerType evidence.ScannerType) string {
	configID := scannerImageConfigID(scannerType)
	if configID == "" {
		return ""
	}
	return getScannerImage(ctx, configID)
}

// GetScannerImageFromConfig returns the Docker image for a scanner type using
// an EACConfig directly. This is useful for standalone commands (like zap)
// that do not use the full ExecutionContext.
func GetScannerImageFromConfig(cfg *config.EACConfig, scannerType evidence.ScannerType) string {
	configID := scannerImageConfigID(scannerType)
	if configID == "" || cfg == nil || cfg.Security == nil {
		return ""
	}
	scanner, ok := cfg.Security.GetScanner(configID)
	if !ok {
		return ""
	}
	return scanner.FullImage()
}

// CreateCommandConfig creates a standard command config for scan commands.
func CreateCommandConfig(scannerType evidence.ScannerType, scannerName string, monikers []string, debug, useTUI bool, tuiHeight int) *cmdframework.CommandConfig {
	return &cmdframework.CommandConfig{
		Type:        core.ActionScan,
		CommandPath: "scan",
		OutputDir:   paths.OutSecurityRelPath,
		Monikers:    monikers,
		UseTUI:      useTUI,
		TUIHeight:   tuiHeight,
		DebugMode:   debug,
	}
}

// MultiScanConfig holds configuration for running multiple scanners.
type MultiScanConfig struct {
	Scanners           []evidence.ScannerType // Scanners to run (empty = use defaults)
	SBOMFormat         string                 // SBOM format
	VulnSeverities     []evidence.Severity    // Vulnerability severity filter
	SemgrepConfig      string                 // SAST config
	ComplianceStandard string                 // Compliance standard
}

// RunMultiScan runs multiple scanners using the unified scan interface.
// If no scanners specified, uses default scanners from config based on module type.
func RunMultiScan(cmdCfg *cmdframework.CommandConfig, multiCfg *MultiScanConfig) int {
	// Store multi-scan config in typed fields
	cmdCfg.MultiScanConfig = multiCfg

	// Create scan context for incremental caching
	sctx := &scanContext{
		scanResults: make(map[string]bool),
	}
	cmdCfg.ScanCmdContext = sctx

	// Set up hooks for multi-scan
	hooks := &cmdframework.Hooks{
		AfterInit:    multiScanAfterInit,
		AfterResolve: multiScanAfterResolve,
		AfterExecute: scanAfterExecute,
	}

	// Register deps verifier for scanner tools
	cmdframework.SetDepsVerifier(scanDepsVerifier)

	return cmdframework.Run(cmdCfg, nil, hooks)
}

// multiScanAfterInit resolves scanners to run based on module types and sets up skip list.
func multiScanAfterInit(ctx *cmdframework.ExecutionContext) error {
	multiCfg, ok := ctx.Config.MultiScanConfig.(*MultiScanConfig)
	if !ok {
		return fmt.Errorf("MultiScanConfig not found or wrong type")
	}

	// Read skip_modules from eac-security contract and populate SkipMonikers.
	// The framework's applySkipFilter will use this list during module resolution.
	if ctx.EACConfig != nil && ctx.EACConfig.Security != nil {
		// Collect all modules that should be skipped
		var skipModules []string
		if ctx.ModuleRegistry != nil {
			for _, moniker := range ctx.ModuleRegistry.AllMonikers() {
				if ctx.EACConfig.Security.ShouldSkipModule(moniker) {
					skipModules = append(skipModules, moniker)
				}
			}
		}
		if len(skipModules) > 0 {
			ctx.Config.SkipMonikers = skipModules
		}
	}

	// If scanners were explicitly specified, use them for all modules
	if len(multiCfg.Scanners) > 0 {
		buildMultiScanInitSummary(ctx, multiCfg.Scanners)
		return nil
	}

	// Otherwise, we'll determine scanners per-module based on type
	// For init summary, show "default" as the scanner mode
	buildMultiScanInitSummary(ctx, nil)

	// Initialize UoW tracker for manifest generation
	sctx, _ := ctx.Config.ScanCmdContext.(*scanContext)
	if sctx != nil {
		sctx.tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, core.ActionScan)
	}

	return nil
}

// multiScanAfterResolve sets the component count after ModuleRegistry is available.
// It also handles incremental scan detection.
func multiScanAfterResolve(ctx *cmdframework.ExecutionContext) error {
	sctx, _ := ctx.Config.ScanCmdContext.(*scanContext)

	// Re-apply security skip filter now that ModuleRegistry is populated.
	// The AfterInit hook sets SkipMonikers but ModuleRegistry may have been nil
	// at that point, resulting in an empty skip list. Re-filter here to ensure
	// skipped modules are removed from ScopeMonikers.
	if ctx.EACConfig != nil && ctx.EACConfig.Security != nil && ctx.ModuleRegistry != nil {
		skipSet := make(map[string]bool)
		for _, moniker := range ctx.ModuleRegistry.AllMonikers() {
			if ctx.EACConfig.Security.ShouldSkipModule(moniker) {
				skipSet[moniker] = true
			}
		}
		if len(skipSet) > 0 {
			var filtered []string
			for _, m := range ctx.ScopeMonikers {
				if !skipSet[m] {
					filtered = append(filtered, m)
				}
			}
			ctx.ScopeMonikers = filtered
		}
	}

	// Now that ModuleRegistry is available, calculate and set UoW count
	if ctx.InitSummary != nil && ctx.ModuleRegistry != nil {
		uowCount := getScanUoWCount(ctx)
		ctx.InitSummary.SetUoWCount(uowCount)
	}

	// Clear scan state if --skip-cache
	// Note: This clears legacy state. UoW manifests are left intact but the
	// UoW-based detection respects the ForceRebuild flag in the worker.
	if ctx.Config.ForceRebuild {
		stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
		if err := stateMgr.ClearContext(core.ActionScan); err != nil {
			log.Warnf("Failed to clear scan state: %v", err)
		}
		return nil
	}

	// UoW-based incremental scan detection (devbox only, not CI)
	// Also run in dry-run mode to show which modules would be scanned/skipped
	if !environments.IsCI() && sctx != nil {
		detectUoWIncrementalScanChanges(ctx, sctx)

		// Pass cache times to orchestrator for TUI display
		if len(sctx.cacheTimes) > 0 && ctx.Orchestrator != nil {
			ctx.Orchestrator.SetCacheTimes(sctx.cacheTimes)
		}

		// Enable early cache detection for fast TUI feedback
		// Tabs will progressively "light up" blue as cache hits are detected
		if (len(sctx.cachedUoWs) > 0 || len(sctx.cachedModules) > 0) && ctx.Orchestrator != nil {
			verifier := &ScanCacheVerifier{
				cachedUoWs:    sctx.cachedUoWs,
				uowCacheTimes: sctx.uowCacheTimes,
				cachedModules: sctx.cachedModules,
			}
			ctx.Orchestrator.SetCacheDetection(verifier, sctx.cachedModules)
		}
	}

	// Pre-compute module input hashes if not already set by incremental detection.
	// In CI mode, incremental detection is skipped, so hashes are never computed.
	// Pre-computing ensures all workers for the same module get a consistent hash.
	if sctx != nil {
		preComputeScanModuleInputHashes(ctx, sctx)
	}

	return nil
}

// preComputeScanModuleInputHashes pre-computes input hashes for all execution modules.
// Skips modules that already have hashes from incremental detection.
func preComputeScanModuleInputHashes(ctx *cmdframework.ExecutionContext, sctx *scanContext) {
	if ctx.ModuleRegistry == nil {
		return
	}

	if sctx.moduleInputHashes == nil {
		sctx.moduleInputHashes = make(map[string]string)
	}

	for _, moniker := range ctx.GetExecutionMonikers() {
		// Skip if already pre-computed by incremental detection
		if _, ok := sctx.moduleInputHashes[moniker]; ok {
			continue
		}

		module, exists := ctx.ModuleRegistry.Get(moniker)
		if !exists {
			continue
		}

		h, err := computeScanInputHash(ctx, module)
		if err != nil {
			log.Debugf("Failed to compute input hash for %s: %v", moniker, err)
			continue
		}

		sctx.moduleInputHashes[moniker] = h
	}
}

// detectUoWIncrementalScanChanges performs UoW-level incremental scan detection.
// Instead of checking at module granularity, it checks each component:tool UoW.
// This enables partial caching - some components can be cached while others rescan.
func detectUoWIncrementalScanChanges(ctx *cmdframework.ExecutionContext, sctx *scanContext) {
	specs := ResolveScanUnitSpecs(ctx)
	result := caching.DetectIncrementalChanges(ctx, core.ActionScan, specs, "SCAN")
	if result == nil {
		return
	}

	// Always store pre-computed hashes for worker reuse
	sctx.moduleInputHashes = result.ModuleInputHashes

	if result.FreshRun {
		return
	}

	// Copy aggregated results to scan context
	sctx.cachedUoWs = result.CachedUoWs
	sctx.uowCacheTimes = result.UoWCacheTimes
	sctx.cachedModules = result.CachedModules
	sctx.cacheTimes = result.ModuleCacheTimes
}

// buildMultiScanInitSummary creates the init summary for multi-scan.
func buildMultiScanInitSummary(ctx *cmdframework.ExecutionContext, scanners []evidence.ScannerType) {
	// Format command name with scanner info
	command := "scan"
	if len(scanners) > 0 {
		var names []string
		for _, s := range scanners {
			names = append(names, string(s))
		}
		command = fmt.Sprintf("scan [%s]", names[0])
		if len(names) > 1 {
			command = fmt.Sprintf("scan [%d scanners]", len(names))
		}
	}

	summary := initsummary.New(command).
		SetRequest(ctx.Config.Monikers, ctx.GetExecutionMonikers()).
		SetExecutionContext(string(logging.GetExecutionContext())).
		SetFlags(initsummary.Flags{
			DebugMode: ctx.Config.DebugMode,
			UseTUI:    ctx.Config.UseTUI,
		}).
		SetOutputDir(ctx.EACConfig.Repository.Paths.Out.Scan)
	// Component count is set in multiScanAfterResolve after ModuleRegistry is available

	ctx.InitSummary = summary
}

// getGitCommit retrieves the current git commit SHA.
// In CI (GITHUB_SHA set), uses the environment variable.
// Locally, runs git rev-parse HEAD.
func getGitCommit(workspaceRoot string) string {
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		return sha
	}
	toolDef := tool.GlobalRegistry().GetOrAdhoc("git")
	execCtx := &tool.ExecutionContext{
		WorkspaceRoot: workspaceRoot,
		ModuleRoot:    workspaceRoot,
		ArgsOverrides: []string{"rev-parse", "HEAD"},
	}
	result, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err != nil || result.ExitCode != 0 {
		return ""
	}
	return strings.TrimSpace(string(result.Stdout))
}
