// Package scan provides the scan command implementation using cmdframework.
package scan

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/cli/eac/impl/scan/internal"
	"github.com/ready-to-release/eac/go/cli/eac/impl/scan/scanners"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/locking"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/hash"
	"github.com/ready-to-release/eac/go/core/logging"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

func init() {
	// Register scan component-level execution support
	cmdframework.RegisterUnitProvider(cmdframework.CommandTypeScan, ResolveScanUnitSpecs)
	cmdframework.RegisterUnitWorker(cmdframework.CommandTypeScan, scanUnitWorker)
}

// NOTE: Scanner-specific semaphores removed - parallelism is now controlled by
// the orchestrator's weighted semaphore system via component weights in getScanWeight()

// log is declared in scan.go

// ScanFrameworkConfig holds scan-specific configuration for the framework.
type ScanFrameworkConfig struct {
	// Scanner identity
	ScannerType  internal.ScannerType
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
	cachedModules map[string]bool      // Modules that are up-to-date (aggregated from UoWs for TUI)
	cacheTimes    map[string]time.Time // Module-level cache times for TUI
	scanResults   map[string]bool      // module -> passed (true = all scanners passed)
	mu            sync.Mutex           // Protects scanResults map for concurrent access

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
	// Store scan config in Extra for access in hooks/worker
	if cmdCfg.Extra == nil {
		cmdCfg.Extra = make(map[string]interface{})
	}
	cmdCfg.Extra["scanConfig"] = scanCfg
	cmdCfg.Extra["scanWorker"] = worker
	scanCfg.ScanResults = make(map[string]*ScanModuleResult)

	// Set up hooks
	hooks := &cmdframework.Hooks{
		AfterInit:    scanAfterInit,
		AfterExecute: scanAfterExecute,
	}

	return cmdframework.Run(cmdCfg, nil, hooks)
}

// scanAfterInit handles scan-specific initialization.
func scanAfterInit(ctx *cmdframework.ExecutionContext) error {
	scanCfg, ok := ctx.Config.Extra["scanConfig"].(*ScanFrameworkConfig)
	if !ok {
		return fmt.Errorf("scanConfig not found or wrong type")
	}

	// Set security output directory from config
	scanCfg.ScanOutDir = ctx.EACConfig.Repository.Paths.Out.Scan

	// Get git commit
	scanCfg.GitCommit = internal.GetGitCommit(ctx.WorkspaceRoot)

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

// recordScanResult records the result of a scan for state tracking.
func recordScanResult(sctx *scanContext, moniker string, passed bool) {
	if sctx == nil {
		return
	}
	sctx.mu.Lock()
	defer sctx.mu.Unlock()

	// Track module-level result (all scanners must pass for module to be cached)
	if existing, ok := sctx.scanResults[moniker]; ok {
		// If any scanner fails, mark module as failed
		if !passed {
			sctx.scanResults[moniker] = false
		} else if !existing {
			// Keep existing false value
		}
	} else {
		sctx.scanResults[moniker] = passed
	}
}

// scanUnitWorker runs scans for a single component using component-level execution.
// This is called by the UnitScheduler for parallel scan component execution.
// The component parameter is in "compName:scannerType" format (e.g., "go:trivy-vuln", "go:semgrep").
func scanUnitWorker(ctx *cmdframework.ExecutionContext, moniker, component string, logWriter io.Writer) int {
	// Get multi-scan config if available (contains scanner settings)
	multiCfg, _ := ctx.Config.Extra["multiScanConfig"].(*MultiScanConfig)
	if multiCfg == nil {
		multiCfg = &MultiScanConfig{
			SBOMFormat:     "cyclonedx-json",
			SemgrepConfig:  "auto",
			VulnSeverities: nil,
		}
	}

	// Get scan context for caching
	sctx, _ := ctx.Config.Extra["scanContext"].(*scanContext)

	// Parse component parameter: "compName:scannerType" (e.g., "go:trivy-vuln")
	parts := strings.SplitN(component, ":", 2)
	compName := parts[0]
	scannerTypeStr := ""
	if len(parts) == 2 {
		scannerTypeStr = parts[1]
	}

	// Build UnitID for UoW-level cache lookup
	unitID := workunit.UnitID{
		Context:   workunit.ContextScan,
		Module:    moniker,
		Component: compName,
		Tool:      scannerTypeStr,
	}

	// Check UoW-level cache first
	isCached := sctx != nil && sctx.cachedUoWs != nil && sctx.cachedUoWs[unitID.Longname()]
	log.Debugf("[SCAN-UOW-CACHE] Component worker for %s: unitID=%s, isCached=%v", component, unitID.Longname(), isCached)

	if isCached {
		if ctx.Config.DryRun {
			output.Writeln(logWriter, "⏭️  %s is up-to-date (would be skipped)", unitID.Longname())
		} else {
			output.Writeln(logWriter, "⏭️  Cached (unchanged)")
		}
		return -1 // -1 = skipped/cached = blue in TUI
	}

	// Get module contract
	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", moniker)
		return 1
	}

	// Validate component parameter parsing
	if compName == "" {
		output.Writeln(logWriter, "Error: invalid component format: %s", component)
		return 1
	}

	// Get component type for this component
	compTypeName := module.Components.GetComponentType(compName)
	compType := ctx.EACConfig.ComponentTypes.Get(compTypeName)

	// Skip non-scannable component types
	if compType == nil || !compType.IsScannable() {
		output.Writeln(logWriter, "⚠️  Component type %s is not scannable, skipping", compTypeName)
		return 0
	}

	// Use dash separator to match UnitID.DirName() for consistent manifest paths
	componentDir := compName
	if scannerTypeStr != "" {
		componentDir = compName + "-" + scannerTypeStr
	}

	// Acquire lock for this component+scanner with wait
	lockCfg := locking.UnitScanConfig(moniker, componentDir, paths.OutSecurityRelPath)
	lockFile, err := locking.AcquireWithWait(context.Background(), ctx.WorkspaceRoot, lockCfg,
		ctx.Orchestrator.GetRegistry(), locking.DefaultWaitConfig())
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer locking.ReleaseTracked(lockFile)

	// Get component root path
	componentRoot := module.Components.GetComponentRoot(compName)
	if componentRoot == "" {
		output.Writeln(logWriter, "Error: no root found for component %s", compName)
		return 1
	}

	// If scanner type specified (3-part key), run only that scanner
	if scannerTypeStr != "" {
		toolID := tool.ScannerToolIDForCategory(scannerTypeStr)
		if toolID == "" {
			output.Writeln(logWriter, "Error: invalid scanner type: %s", scannerTypeStr)
			return 1
		}
		return runUnitScanner(ctx, module, moniker, compName, componentRoot, internal.ScannerType(toolID), multiCfg, logWriter)
	}

	// 2-part key (no scanner specified): run all scanners from component type configuration
	var scanners []internal.ScannerType
	seenScanners := make(map[string]bool)
	for _, s := range compType.GetScanners() {
		if !seenScanners[s] {
			if toolID := tool.ScannerToolIDForCategory(s); toolID != "" {
				scanners = append(scanners, internal.ScannerType(toolID))
				seenScanners[s] = true
			}
		}
	}

	if len(scanners) == 0 {
		output.Writeln(logWriter, "⚠️  No scanners configured for component type: %s", compTypeName)
		return 0
	}

	// Run each scanner
	exitCode := 0
	for _, scannerType := range scanners {
		result := runUnitScanner(ctx, module, moniker, compName, componentRoot, scannerType, multiCfg, logWriter)
		if result != 0 {
			exitCode = 1 // Mark as failed but continue with other scanners
		}
	}

	return exitCode
}

// runUnitScanner runs a single scanner for a component.
func runUnitScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, moniker, component, componentRoot string, scannerType internal.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) int {
	emoji := getScannerEmoji(scannerType)

	// Get scan context for result recording
	sctx, _ := ctx.Config.Extra["scanContext"].(*scanContext)

	// In dry-run mode, show what would happen without actually scanning
	if ctx.Config.DryRun {
		output.Writeln(logWriter, "🔍 %s/%s would be scanned (changed)", moniker, component)
		output.Writeln(logWriter, "   Scanner: %s", scannerType)
		return 0
	}

	// Parallelism controlled by orchestrator's weighted semaphore via component weights
	output.Writeln(logWriter, "%s Running %s scanner on %s/%s...", emoji, scannerType, moniker, component)
	scanStart := time.Now()

	// Compute input hash for UoW manifest (before scanning)
	inputHash, _ := computeScanInputHash(ctx, module)

	// Create scan config for this scanner
	scanCfg := &ScanFrameworkConfig{
		ScannerType:  scannerType,
		ScannerName:  string(scannerType),
		ScannerEmoji: emoji,
		ScanOutDir:   ctx.EACConfig.Repository.Paths.Out.Scan,
		GitCommit:    internal.GetGitCommit(ctx.WorkspaceRoot),
	}

	// Run the appropriate scanner on the component root
	findings, err := runScannerOnPath(ctx, componentRoot, scannerType, multiCfg, logWriter)
	if err != nil {
		handleScanFailure(ctx, scanCfg, module, moniker, component, scanStart, err, logWriter)
		// Record failed scan result
		recordScanResult(sctx, moniker, false)
		return 1
	}

	// Write evidence and update manifest
	_, writeErr := handleScanSuccess(ctx, scanCfg, module, moniker, component, scanStart, findings, logWriter)
	if writeErr != nil {
		// Record failed scan result
		recordScanResult(sctx, moniker, false)
		return 1
	}

	// Record successful scan result
	recordScanResult(sctx, moniker, true)

	// Write UoW manifest for incremental cache
	if sctx != nil {
		writeUoWScanManifest(ctx, sctx, moniker, component, string(scannerType), inputHash, scanStart)
	}

	return 0
}

// computeScanInputHash computes the input hash for a module's scan.
func computeScanInputHash(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract) (string, error) {
	patterns := module.GetGlobPatterns()
	files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
	if err != nil {
		return "", err
	}
	return hash.Files(ctx.WorkspaceRoot, files)
}

// writeUoWScanManifest writes a UoW manifest for a successful scan.
func writeUoWScanManifest(ctx *cmdframework.ExecutionContext, sctx *scanContext, moniker, component, tool, inputHash string, startTime time.Time) {
	// Initialize tracker if needed
	sctx.mu.Lock()
	if sctx.tracker == nil {
		sctx.tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, workunit.ContextScan)
	}
	tracker := sctx.tracker
	sctx.mu.Unlock()

	// Build UnitID for the tracker
	unitID := workunit.UnitID{
		Context:   workunit.ContextScan,
		Module:    moniker,
		Component: component,
		Tool:      tool,
	}

	// Create and record the manifest
	manifest := &coreoutput.UoWManifest{
		Context:    workunit.ContextScan,
		Module:     moniker,
		Component:  component,
		Tool:       tool,
		InputHash:  inputHash,
		ExecutedAt: startTime,
		ExitCode:   0, // Success
		Duration:   time.Since(startTime),
	}

	if err := tracker.RecordComplete(unitID, manifest); err != nil {
		log.Debugf("[SCAN-UOW-CACHE] Failed to write UoW manifest for %s/%s:%s: %v", moniker, component, tool, err)
	} else {
		log.Debugf("[SCAN-UOW-CACHE] Wrote UoW manifest for %s/%s:%s", moniker, component, tool)
	}
}

// logScannerConfig logs scanner-specific configuration for visibility.
func logScannerConfig(scannerType internal.ScannerType, logWriter io.Writer, multiCfg *MultiScanConfig, ctx *cmdframework.ExecutionContext) {
	switch scannerType {
	case internal.ScannerSBOM:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
		output.Writeln(logWriter, "  Format: %s", multiCfg.SBOMFormat)
	case internal.ScannerVuln:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
		if len(multiCfg.VulnSeverities) > 0 {
			output.Writeln(logWriter, "  Severity filter: %v", multiCfg.VulnSeverities)
		}
	case internal.ScannerSecrets, internal.ScannerIaC:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
	case internal.ScannerCompliance:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
		output.Writeln(logWriter, "  Compliance standard: %s", multiCfg.ComplianceStandard)
	case internal.ScannerSAST:
		output.Writeln(logWriter, "  Using Semgrep image: %s", getSemgrepImage(ctx))
		output.Writeln(logWriter, "  Config: %s", multiCfg.SemgrepConfig)
	case internal.ScannerDAST:
		output.Writeln(logWriter, "  Using ZAP image: %s", getZAPImage(ctx))
	}
}

// runScannerOnPath runs a scanner on a specific path using the scanner registry.
func runScannerOnPath(ctx *cmdframework.ExecutionContext, targetPath string, scannerType internal.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) (interface{}, error) {
	// ZAP requires special handling (target URL)
	if scannerType == internal.ScannerDAST {
		return nil, fmt.Errorf("ZAP scanner requires --target URL flag")
	}

	// Populate global scan context with execution-time config
	scanners.GlobalScanContext = &scanners.ScanContext{
		TrivyImage:         getTrivyImage(ctx),
		SemgrepImage:       getSemgrepImage(ctx),
		ZAPImage:           getZAPImage(ctx),
		SBOMFormat:         multiCfg.SBOMFormat,
		VulnSeverities:     multiCfg.VulnSeverities,
		SemgrepConfig:      multiCfg.SemgrepConfig,
		ComplianceStandard: multiCfg.ComplianceStandard,
		WorkspaceRoot:      ctx.WorkspaceRoot,
		GitCommit:          internal.GetGitCommit(ctx.WorkspaceRoot),
	}
	defer func() { scanners.GlobalScanContext = nil }()

	// Log scanner-specific configuration
	logScannerConfig(scannerType, logWriter, multiCfg, ctx)

	// Get scanner from registry (native or YAML)
	scanFn := scanners.GetScanner(string(scannerType))
	if scanFn == nil {
		return nil, fmt.Errorf("no scanner found for type: %s", scannerType)
	}

	// Execute scanner
	return scanFn(ctx.WorkspaceRoot, targetPath, "", logWriter, tool.ScanOptions{
		ScanType: string(scannerType),
	})
}

// handleScanFailure handles a failed scan (module or component level).
// If component is empty, treats as module-level scan.
func handleScanFailure(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig,
	module *modules.ModuleContract, moniker, component string,
	scanStart time.Time, scanErr error, logWriter io.Writer) {

	output.Writeln(logWriter, "  ❌ Failed: %v", scanErr)

	// Write error evidence
	var outputPath string
	var writeErr error
	if component != "" {
		outputPath, writeErr = internal.WriteComponentErrorEvidence(
			ctx.WorkspaceRoot, moniker, component, scanCfg.ScannerType, scanErr.Error())
	} else {
		outputPath, writeErr = internal.WriteErrorEvidence(
			ctx.WorkspaceRoot, moniker, scanCfg.ScannerType, scanErr.Error())
	}

	if writeErr != nil {
		output.Writeln(logWriter, "  Failed to write error evidence: %v", writeErr)
	} else {
		output.Writeln(logWriter, "  📄 Error evidence: %s", outputPath)
	}

}

// handleScanSuccess handles a successful scan (module or component level).
func handleScanSuccess(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig,
	module *modules.ModuleContract, moniker, component string,
	scanStart time.Time, findings interface{}, logWriter io.Writer) (string, error) {

	var outputPath string
	var err error
	if component != "" {
		outputPath, err = internal.WriteComponentEvidence(
			ctx.WorkspaceRoot, moniker, component, scanCfg.ScannerType, findings)
	} else {
		outputPath, err = internal.WriteEvidence(
			ctx.WorkspaceRoot, moniker, scanCfg.ScannerType, findings)
	}

	if err != nil {
		output.Writeln(logWriter, "  ❌ Failed to write evidence: %v", err)
		return "", err
	}

	output.Writeln(logWriter, "  ✅ Success: %s", outputPath)
	return outputPath, nil
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

// GetDockerImage returns the appropriate Docker image for a scanner type.
// Uses the eac-security contract to resolve scanner images.
func GetDockerImage(ctx *cmdframework.ExecutionContext, scannerType internal.ScannerType) string {
	if ctx.EACConfig == nil || ctx.EACConfig.Security == nil {
		return ""
	}

	// Map scanner type to scanner ID in eac-security contract
	var scannerID string
	switch scannerType {
	case internal.ScannerSBOM:
		scannerID = "trivy-sbom"
	case internal.ScannerVuln:
		scannerID = "trivy-vuln"
	case internal.ScannerSecrets, internal.ScannerIaC, internal.ScannerCompliance:
		scannerID = "trivy-sbom" // All Trivy-based scanners use same image
	case internal.ScannerSAST:
		scannerID = "semgrep"
	case internal.ScannerDAST:
		scannerID = "zap"
	default:
		return ""
	}

	scanner, ok := ctx.EACConfig.Security.GetScanner(scannerID)
	if !ok {
		return ""
	}
	return scanner.FullImage()
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

// getTrivyImage returns the Trivy Docker image from Security config.
func getTrivyImage(ctx *cmdframework.ExecutionContext) string {
	return getScannerImage(ctx, "trivy-sbom")
}

// getSemgrepImage returns the Semgrep Docker image from Security config.
func getSemgrepImage(ctx *cmdframework.ExecutionContext) string {
	return getScannerImage(ctx, "semgrep")
}

// getZAPImage returns the ZAP Docker image from Security config.
func getZAPImage(ctx *cmdframework.ExecutionContext) string {
	return getScannerImage(ctx, "zap")
}

// CreateCommandConfig creates a standard command config for scan commands.
func CreateCommandConfig(scannerType internal.ScannerType, scannerName string, monikers []string, debug, useTUI bool, tuiHeight int) *cmdframework.CommandConfig {
	return &cmdframework.CommandConfig{
		Type:        cmdframework.CommandTypeScan,
		CommandPath: "scan",
		ActionVerb:  fmt.Sprintf("Scanning (%s)", scannerName),
		OutputDir:   "out/scan",
		LogFileName: fmt.Sprintf("%s.log", scannerType),
		Monikers:    monikers,
		UseTUI:      useTUI,
		TUIHeight:   tuiHeight,
		DebugMode:   debug,
	}
}

// MultiScanConfig holds configuration for running multiple scanners.
type MultiScanConfig struct {
	Scanners           []internal.ScannerType // Scanners to run (empty = use defaults)
	SBOMFormat         string                 // SBOM format
	VulnSeverities     []internal.Severity    // Vulnerability severity filter
	SemgrepConfig      string                 // SAST config
	ComplianceStandard string                 // Compliance standard
}

// RunMultiScan runs multiple scanners using the unified scan interface.
// If no scanners specified, uses default scanners from config based on module type.
func RunMultiScan(cmdCfg *cmdframework.CommandConfig, multiCfg *MultiScanConfig) int {
	// Initialize cmdframework to get config and resolve modules
	if cmdCfg.Extra == nil {
		cmdCfg.Extra = make(map[string]interface{})
	}
	cmdCfg.Extra["multiScanConfig"] = multiCfg

	// Create scan context for incremental caching
	sctx := &scanContext{
		scanResults: make(map[string]bool),
	}
	cmdCfg.Extra["scanContext"] = sctx

	// Set up hooks for multi-scan
	hooks := &cmdframework.Hooks{
		AfterInit:    multiScanAfterInit,
		AfterResolve: multiScanAfterResolve,
		AfterExecute: scanAfterExecute,
	}

	return cmdframework.Run(cmdCfg, multiScanWorker, hooks)
}

// multiScanAfterInit resolves scanners to run based on module types and sets up skip list.
func multiScanAfterInit(ctx *cmdframework.ExecutionContext) error {
	multiCfg, ok := ctx.Config.Extra["multiScanConfig"].(*MultiScanConfig)
	if !ok {
		return fmt.Errorf("multiScanConfig not found or wrong type")
	}

	// Read skip_modules from eac-security contract and populate Extra["skipMonikers"]
	// The framework's applySkipFilter will use this list during module resolution
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
			ctx.Config.Extra["skipMonikers"] = skipModules
		}
	}

	// If scanners were explicitly specified, use them for all modules
	if len(multiCfg.Scanners) > 0 {
		ctx.Config.Extra["resolvedScanners"] = multiCfg.Scanners
		buildMultiScanInitSummary(ctx, multiCfg.Scanners)
		return nil
	}

	// Otherwise, we'll determine scanners per-module based on type
	// For init summary, show "default" as the scanner mode
	buildMultiScanInitSummary(ctx, nil)
	return nil
}

// multiScanAfterResolve sets the component count after ModuleRegistry is available.
// It also handles incremental scan detection.
func multiScanAfterResolve(ctx *cmdframework.ExecutionContext) error {
	sctx, _ := ctx.Config.Extra["scanContext"].(*scanContext)

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
		if err := stateMgr.ClearContext(workunit.ContextScan); err != nil {
			log.Warnf("Failed to clear scan state: %v", err)
		}
		return nil
	}

	// UoW-based incremental scan detection (devbox only, not CI)
	// Also run in dry-run mode to show which modules would be scanned/skipped
	if !environments.IsCI() && sctx != nil {
		detectUoWIncrementalScanChanges(ctx, sctx)
	}

	return nil
}

// detectUoWIncrementalScanChanges performs UoW-level incremental scan detection.
// Instead of checking at module granularity, it checks each component:tool UoW.
// This enables partial caching - some components can be cached while others rescan.
func detectUoWIncrementalScanChanges(ctx *cmdframework.ExecutionContext, sctx *scanContext) {
	startTime := time.Now()
	defer func() {
		ctx.SetChangeDetectionTiming(time.Since(startTime))
	}()

	// Get all expected UoWs from resolved unit specs
	units := ResolveScanUnitSpecs(ctx)
	if len(units) == 0 {
		return
	}

	// Build list of expected UoWs
	expectedUoWs := make([]workunit.UnitID, len(units))
	for i, spec := range units {
		expectedUoWs[i] = spec.ID
	}

	// Collect module files for hash computation
	moduleFiles := make(map[string][]string)
	for _, id := range expectedUoWs {
		if _, ok := moduleFiles[id.Module]; ok {
			continue // Already collected
		}
		if contract, ok := ctx.ModuleRegistry.Get(id.Module); ok {
			patterns := contract.GetGlobPatterns()
			files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
			if err != nil {
				log.Debugf("Failed to expand patterns for %s: %v", id.Module, err)
				continue
			}
			moduleFiles[id.Module] = files
		}
	}

	// Use shared helpers for change detection and aggregation
	reader := coreoutput.NewReader(ctx.WorkspaceRoot)
	getInputHash := coreoutput.InputHashProvider(hash.NewModuleInputHashProvider(ctx.WorkspaceRoot, moduleFiles))

	aggResult, err := coreoutput.AggregateUoWChanges(reader, workunit.ContextScan, expectedUoWs, getInputHash)
	if err != nil {
		log.Debugf("Failed to detect UoW changes: %v", err)
		return
	}

	// Log change detection results
	log.Debugf("[SCAN-UOW-CACHE] DetectUoWChanges result: FreshRun=%v Changed=%d UpToDate=%d",
		aggResult.UoWResult.FreshRun, len(aggResult.UoWResult.Changed), len(aggResult.UoWResult.UpToDate))
	for longname, reason := range aggResult.UoWResult.ChangeReasons {
		log.Debugf("[SCAN-UOW-CACHE] Changed: %s -> %s", longname, reason)
	}

	detectionTime := time.Since(startTime)

	if aggResult.UoWResult.FreshRun {
		log.Debugf("Fresh scan detected (UoW mode), all components will scan")
		if ctx.InitSummary != nil {
			ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: detectionTime,
				FreshBuild:    true,
			})
		}
		return
	}

	// Copy aggregated results to context
	sctx.cachedUoWs = aggResult.CachedUoWs
	sctx.uowCacheTimes = aggResult.UoWCacheTimes
	sctx.cachedModules = aggResult.CachedModules
	sctx.cacheTimes = aggResult.ModuleCacheTimes

	// Log module-level aggregation
	agg := workunit.NewUoWAggregator(expectedUoWs)
	for _, id := range aggResult.UoWResult.UpToDate {
		agg.MarkCached(id)
	}
	for module := range aggResult.CachedModules {
		total, cached := agg.Stats(module)
		log.Debugf("[SCAN-UOW-CACHE] Module %s: %d/%d UoWs cached -> module cached=%v",
			module, cached, total, true)
	}
	for _, module := range aggResult.ChangedModules {
		total, cached := agg.Stats(module)
		log.Debugf("[SCAN-UOW-CACHE] Module %s: %d/%d UoWs cached -> module cached=%v",
			module, cached, total, false)
	}

	// Report incremental detection in init summary
	if ctx.InitSummary != nil {
		ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
			Enabled:       true,
			DetectionTime: detectionTime,
			Changed:       aggResult.ChangedModules,
			UpToDate:      aggResult.UpToDateModules,
			FreshBuild:    false,
		})
	}

	log.Debugf("Incremental (UoW mode): %d modules to scan, %d cached, %d UoWs cached",
		len(aggResult.ChangedModules), len(aggResult.UpToDateModules), len(sctx.cachedUoWs))
}

// multiScanWorker runs all configured scanners for a single module.
func multiScanWorker(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	multiCfg, ok := ctx.Config.Extra["multiScanConfig"].(*MultiScanConfig)
	if !ok {
		output.Writeln(logWriter, "Error: multiScanConfig not found or wrong type")
		return 1
	}

	// Get module to determine its type
	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", moniker)
		return 1
	}

	// Acquire lock for this module with wait
	lockCfg := locking.ScanConfig(moniker, paths.OutSecurityRelPath) // OutSecurityRelPath = "out/scan"
	lockFile, err := locking.AcquireWithWait(context.Background(), ctx.WorkspaceRoot, lockCfg,
		ctx.Orchestrator.GetRegistry(), locking.DefaultWaitConfig())
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer locking.ReleaseTracked(lockFile)

	// Determine which scanners to run
	var scanners []internal.ScannerType
	if len(multiCfg.Scanners) > 0 {
		scanners = multiCfg.Scanners
	} else {
		// Get scanners from component types for each of the module's components
		seenScanners := make(map[string]bool)
		for _, componentName := range module.GetEnabledComponents() {
			compTypeName := module.Components.GetComponentType(componentName)
			compType := ctx.EACConfig.ComponentTypes.Get(compTypeName)
			if compType == nil || !compType.IsScannable() {
				continue // Skip non-scannable component types
			}
			for _, s := range compType.GetScanners() {
				if !seenScanners[s] {
					if toolID := tool.ScannerToolIDForCategory(s); toolID != "" {
						scanners = append(scanners, internal.ScannerType(toolID))
						seenScanners[s] = true
					}
				}
			}
		}
	}

	if len(scanners) == 0 {
		output.Writeln(logWriter, "⚠️  No scannable components in module: %s", module.GetComponentTypesDisplay())
		return 0
	}

	// Run each scanner sequentially
	exitCode := 0
	for _, scannerType := range scanners {
		result := runSingleScanner(ctx, module, scannerType, multiCfg, logWriter)
		if result != 0 {
			exitCode = 1 // Mark as failed but continue with other scanners
		}
	}

	return exitCode
}

// runSingleScanner runs a single scanner for a module.
func runSingleScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, scannerType internal.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) int {
	emoji := getScannerEmoji(scannerType)

	// Parallelism controlled by orchestrator's weighted semaphore via component weights
	output.Writeln(logWriter, "%s Running %s scanner...", emoji, scannerType)
	scanStart := time.Now()

	// Create scan config for this scanner
	scanCfg := &ScanFrameworkConfig{
		ScannerType:  scannerType,
		ScannerName:  string(scannerType),
		ScannerEmoji: emoji,
		ScanOutDir:   ctx.EACConfig.Repository.Paths.Out.Scan,
		GitCommit:    internal.GetGitCommit(ctx.WorkspaceRoot),
	}

	// Run the appropriate scanner
	findings, err := runScanner(ctx, module, scannerType, multiCfg, logWriter)
	if err != nil {
		handleScanFailure(ctx, scanCfg, module, module.Moniker, "", scanStart, err, logWriter)
		return 1
	}

	// Write evidence and update manifest
	_, writeErr := handleScanSuccess(ctx, scanCfg, module, module.Moniker, "", scanStart, findings, logWriter)
	if writeErr != nil {
		return 1
	}

	return 0
}

// runScanner dispatches to the appropriate scanner implementation using the scanner registry.
func runScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, scannerType internal.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) (interface{}, error) {
	// Get the module's scannable root (buildable package root or first available)
	moduleRoot := module.Components.GetBuildableRoot()
	if moduleRoot == "" {
		for _, root := range module.GetComponentRoots() {
			moduleRoot = root
			break
		}
	}
	if moduleRoot == "" {
		return nil, fmt.Errorf("no package root found for module %s", module.Moniker)
	}

	// ZAP requires special handling (target URL)
	if scannerType == internal.ScannerDAST {
		return nil, fmt.Errorf("ZAP scanner requires --target URL flag")
	}

	// Populate global scan context with execution-time config
	scanners.GlobalScanContext = &scanners.ScanContext{
		TrivyImage:         getTrivyImage(ctx),
		SemgrepImage:       getSemgrepImage(ctx),
		ZAPImage:           getZAPImage(ctx),
		SBOMFormat:         multiCfg.SBOMFormat,
		VulnSeverities:     multiCfg.VulnSeverities,
		SemgrepConfig:      multiCfg.SemgrepConfig,
		ComplianceStandard: multiCfg.ComplianceStandard,
		WorkspaceRoot:      ctx.WorkspaceRoot,
		GitCommit:          internal.GetGitCommit(ctx.WorkspaceRoot),
	}
	defer func() { scanners.GlobalScanContext = nil }()

	// Log scanner-specific configuration
	logScannerConfig(scannerType, logWriter, multiCfg, ctx)

	// Get scanner from registry (native or YAML)
	scanFn := scanners.GetScanner(string(scannerType))
	if scanFn == nil {
		return nil, fmt.Errorf("no scanner found for type: %s", scannerType)
	}

	// Execute scanner
	return scanFn(ctx.WorkspaceRoot, moduleRoot, "", logWriter, tool.ScanOptions{
		ScanType: string(scannerType),
	})
}

// getScannerEmoji returns the emoji for a scanner type.
func getScannerEmoji(scannerType internal.ScannerType) string {
	switch scannerType {
	case internal.ScannerSBOM:
		return "📦"
	case internal.ScannerVuln:
		return "🔍"
	case internal.ScannerSecrets:
		return "🔐"
	case internal.ScannerIaC:
		return "🏗️"
	case internal.ScannerCompliance:
		return "✅"
	case internal.ScannerSAST:
		return "🔬"
	case internal.ScannerDAST:
		return "🌐"
	default:
		return "🔒"
	}
}

// buildMultiScanInitSummary creates the init summary for multi-scan.
func buildMultiScanInitSummary(ctx *cmdframework.ExecutionContext, scanners []internal.ScannerType) {
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
