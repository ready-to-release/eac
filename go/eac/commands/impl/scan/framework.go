// Package scan provides the scan command implementation using cmdframework.
package scan

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/scanners"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locking"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/ready-to-release/eac/go/eac/core/environments"
	"github.com/ready-to-release/eac/go/eac/core/hash"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/manifest"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/tool"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

func init() {
	// Register scan component-level execution support
	cmdframework.RegisterUnitProvider(cmdframework.CommandTypeScan, FlattenModulesToScanUnits)
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
	cachedModules map[string]bool  // Modules that are up-to-date (cache hits)
	scanResults   map[string]bool  // module -> passed (true = all scanners passed)
	mu            sync.Mutex       // Protects scanResults map for concurrent access
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

	return cmdframework.Run(cmdCfg, scanWorkerWrapper, hooks)
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

// scanAfterExecute handles scan manifest aggregation after component-level execution.
// When scans run at component level, this aggregates component manifests into module-level manifests.
func scanAfterExecute(ctx *cmdframework.ExecutionContext) error {
	// Aggregate component manifests into module-level manifests
	// This is needed because component-level execution creates manifests at
	// out/scan/<module>/<component>/scan.manifest.json, but show scan-summary
	// expects them at out/scan/<module>/scan.manifest.json
	if err := aggregateComponentScanManifests(ctx); err != nil {
		log.Warnf("Failed to aggregate scan manifests: %v", err)
	}

	// Update incremental scan state (devbox only)
	if !environments.IsCI() && !ctx.Config.DryRun {
		sctx, _ := ctx.Config.Extra["scanContext"].(*scanContext)
		if sctx != nil {
			updateScanState(ctx, sctx)
		}
	}

	return nil
}

// updateScanState updates the scan state after execution.
// This follows the build pattern - computing moduleFiles fresh rather than storing in context.
func updateScanState(ctx *cmdframework.ExecutionContext, sctx *scanContext) {
	// Collect successfully scanned modules (exit code 0 or -1 for cached)
	scannedModules := make(map[string]bool)
	for _, result := range ctx.Results {
		if result.ExitCode == 0 || result.ExitCode == -1 {
			// Check if we have an explicit result recorded
			sctx.mu.Lock()
			if passed, ok := sctx.scanResults[result.Moniker]; ok {
				scannedModules[result.Moniker] = passed
			} else {
				// Default: exit code 0 means passed
				scannedModules[result.Moniker] = result.ExitCode == 0
			}
			sctx.mu.Unlock()
		}
	}

	// Also include component results (for component-level execution)
	for _, result := range ctx.UnitResults {
		if result.ExitCode == 0 || result.ExitCode == -1 {
			// Check if we have an explicit result recorded
			sctx.mu.Lock()
			if passed, ok := sctx.scanResults[result.Module]; ok {
				scannedModules[result.Module] = passed
			} else {
				// For component results, check if any failed
				if existing, exists := scannedModules[result.Module]; exists {
					// Keep false if already failed
					if !existing {
						scannedModules[result.Module] = false
					}
				} else {
					scannedModules[result.Module] = result.ExitCode == 0
				}
			}
			sctx.mu.Unlock()
		}
	}

	if len(scannedModules) == 0 {
		log.Debugf("[SCAN-CACHE] No scanned modules to update state for")
		return
	}

	log.Debugf("[SCAN-CACHE] Updating state for %d modules", len(scannedModules))
	for moniker, passed := range scannedModules {
		log.Debugf("[SCAN-CACHE] Module %s: passed=%v", moniker, passed)
	}

	// Update state using StateManager
	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
	for moniker, passed := range scannedModules {
		contract, ok := ctx.ModuleRegistry.Get(moniker)
		if !ok {
			continue
		}

		// Get source files and compute hash
		patterns := contract.GetGlobPatterns()
		files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
		if err != nil {
			log.Debugf("[SCAN-CACHE] Failed to expand patterns for %s: %v", moniker, err)
			continue
		}

		sourceHash, err := hash.Files(ctx.WorkspaceRoot, files)
		if err != nil {
			log.Debugf("[SCAN-CACHE] Failed to hash files for %s: %v", moniker, err)
			continue
		}

		if err := stateMgr.SaveModuleResult(workunit.ContextScan, moniker, passed, sourceHash); err != nil {
			log.Warnf("Failed to update scan state for %s: %v", moniker, err)
		}
	}
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

// scanWorkerWrapper wraps the scanner-specific worker for cmdframework.
func scanWorkerWrapper(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	scanCfg, ok := ctx.Config.Extra["scanConfig"].(*ScanFrameworkConfig)
	if !ok {
		output.Writeln(logWriter, "Error: scanConfig not found or wrong type")
		return 1
	}
	worker, ok := ctx.Config.Extra["scanWorker"].(ScanWorker)
	if !ok {
		output.Writeln(logWriter, "Error: scanWorker not found or wrong type")
		return 1
	}

	// Get module contract
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

	// Parallelism controlled by orchestrator's weighted semaphore via component weights
	output.Writeln(logWriter, "%s Scanning %s...", scanCfg.ScannerEmoji, moniker)
	scanStart := time.Now()

	// Run scanner-specific worker
	findings, err := worker(ctx, module, scanCfg, logWriter)
	duration := time.Since(scanStart)

	result := &ScanModuleResult{
		Moniker:  moniker,
		Duration: duration,
	}

	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		handleScanFailure(ctx, scanCfg, module, moniker, "", scanStart, err, logWriter)
		scanCfg.ScanResults[moniker] = result
		return 1
	}

	// Write evidence and update manifest
	evidencePath, writeErr := handleScanSuccess(ctx, scanCfg, module, moniker, "", scanStart, findings, logWriter)
	if writeErr != nil {
		result.Success = false
		result.ErrorMessage = writeErr.Error()
		scanCfg.ScanResults[moniker] = result
		return 1
	}

	result.Success = true
	result.EvidencePath = evidencePath
	scanCfg.ScanResults[moniker] = result

	return 0
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

	// Check incremental cache first - if module is cached, skip immediately (blue in TUI)
	if sctx != nil && sctx.cachedModules != nil && sctx.cachedModules[moniker] {
		if ctx.Config.DryRun {
			output.Writeln(logWriter, "⏭️  %s is up-to-date (would be skipped)", moniker)
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

	// Parse component parameter: "compName:scannerType" (e.g., "go:trivy-vuln")
	parts := strings.SplitN(component, ":", 2)
	compName := parts[0]
	scannerTypeStr := ""
	if len(parts) == 2 {
		scannerTypeStr = parts[1]
	}

	// Get component type for this component
	compTypeName := module.Components.GetComponentType(compName)
	compType := ctx.EACConfig.ComponentTypes.Get(compTypeName)

	// Skip non-scannable component types
	if compType == nil || !compType.IsScannable() {
		output.Writeln(logWriter, "⚠️  Component type %s is not scannable, skipping", compTypeName)
		return 0
	}

	// Use underscore separator for Windows compatibility in lock/output paths
	componentDir := compName
	if scannerTypeStr != "" {
		componentDir = compName + "_" + scannerTypeStr
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

	return 0
}

// logScannerConfig logs scanner-specific configuration for visibility.
func logScannerConfig(scannerType internal.ScannerType, logWriter io.Writer, multiCfg *MultiScanConfig, ctx *cmdframework.ExecutionContext) {
	switch scannerType {
	case internal.ScannerSBOM:
		output.Writeln(logWriter, "  Using Trivy image: %s", ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage())
		output.Writeln(logWriter, "  Format: %s", multiCfg.SBOMFormat)
	case internal.ScannerVuln:
		output.Writeln(logWriter, "  Using Trivy image: %s", ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage())
		if len(multiCfg.VulnSeverities) > 0 {
			output.Writeln(logWriter, "  Severity filter: %v", multiCfg.VulnSeverities)
		}
	case internal.ScannerSecrets, internal.ScannerIaC:
		output.Writeln(logWriter, "  Using Trivy image: %s", ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage())
	case internal.ScannerCompliance:
		output.Writeln(logWriter, "  Using Trivy image: %s", ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage())
		output.Writeln(logWriter, "  Compliance standard: %s", multiCfg.ComplianceStandard)
	case internal.ScannerSAST:
		output.Writeln(logWriter, "  Using Semgrep image: %s", ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage())
		output.Writeln(logWriter, "  Config: %s", multiCfg.SemgrepConfig)
	case internal.ScannerDAST:
		output.Writeln(logWriter, "  Using ZAP image: %s", ctx.EACConfig.SecurityTools.DockerImages.ZAP.FullImage())
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
		TrivyImage:         ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage(),
		SemgrepImage:       ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage(),
		ZAPImage:           ctx.EACConfig.SecurityTools.DockerImages.ZAP.FullImage(),
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

	updateScanManifest(ctx, scanCfg, module, moniker, component, scanStart,
		manifest.ScanStatusFailed, outputPath, scanErr.Error())
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
		updateScanManifest(ctx, scanCfg, module, moniker, component, scanStart,
			manifest.ScanStatusFailed, "", err.Error())
		return "", err
	}

	updateScanManifest(ctx, scanCfg, module, moniker, component, scanStart,
		manifest.ScanStatusPassed, outputPath, "")
	output.Writeln(logWriter, "  ✅ Success: %s", outputPath)
	return outputPath, nil
}

// updateScanManifest updates the scan manifest for a module or component.
func updateScanManifest(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig,
	module *modules.ModuleContract, moniker, component string,
	scanStart time.Time, status, evidencePath, errorMsg string) {

	duration := time.Since(scanStart)

	var scanDir, identifier, moduleType string
	if component != "" {
		scanDir = paths.ComponentScanOutputPath(ctx.WorkspaceRoot, moniker, component)
		identifier = moniker + "/" + component
		moduleType = module.Components.GetComponentType(component)
	} else {
		scanDir = ctx.EACConfig.Repository.ScanModuleOutputPathAbs(ctx.WorkspaceRoot, moniker)
		identifier = moniker
		moduleType = module.GetComponentTypesDisplay()
	}

	mf, err := manifest.LoadOrCreateScanManifest(scanDir, identifier, moduleType, scanCfg.GitCommit)
	if err != nil {
		log.Warnf("Failed to load/create scan manifest: %v", err)
		return
	}

	result := manifest.ScannerResult{
		Status:          status,
		RunTime:         time.Now(),
		DurationSeconds: duration.Seconds(),
		EvidencePath:    evidencePath,
		Error:           errorMsg,
	}
	mf.AddScannerResult(string(scanCfg.ScannerType), result)

	if err := mf.Save(scanDir); err != nil {
		log.Warnf("Failed to save scan manifest: %v", err)
	}
}

// aggregateComponentScanManifests aggregates component-level scan manifests into module-level manifests.
// This is called after component-level scan execution to create the module-level manifest that
// show scan-summary expects at out/scan/<module>/scan.manifest.json.
func aggregateComponentScanManifests(ctx *cmdframework.ExecutionContext) error {
	// Get unique modules from results
	moduleSet := make(map[string]bool)
	for _, result := range ctx.Results {
		// Component results have format "module/component", extract module
		moniker := result.Moniker
		if idx := strings.Index(moniker, "/"); idx > 0 {
			moniker = moniker[:idx]
		}
		moduleSet[moniker] = true
	}

	// For each module, aggregate component manifests
	for moniker := range moduleSet {
		if err := aggregateModuleScanManifest(ctx, moniker); err != nil {
			log.Warnf("Failed to aggregate scan manifest for %s: %v", moniker, err)
			// Continue with other modules
		}
	}

	return nil
}

// aggregateModuleScanManifest aggregates all component scan manifests for a module into a single module-level manifest.
func aggregateModuleScanManifest(ctx *cmdframework.ExecutionContext, moniker string) error {
	moduleScanDir := ctx.EACConfig.Repository.ScanModuleOutputPathAbs(ctx.WorkspaceRoot, moniker)

	// Read directory to find component subdirectories
	entries, err := os.ReadDir(moduleScanDir)
	if err != nil {
		// No scan directory yet - nothing to aggregate
		return nil
	}

	// Get module for type info
	module, exists := ctx.ModuleRegistry.Get(moniker)
	moduleType := ""
	if exists {
		moduleType = module.GetComponentTypesDisplay()
	}

	// Get git commit
	gitCommit := internal.GetGitCommit(ctx.WorkspaceRoot)

	// Create or load the module-level manifest
	moduleMf := manifest.NewScanManifest(moniker, moduleType, gitCommit)

	// Track total duration
	var totalDuration float64

	// Scan component subdirectories for manifests
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		componentDir := filepath.Join(moduleScanDir, entry.Name())
		componentMf, loadErr := manifest.LoadScanManifest(componentDir)
		if loadErr != nil {
			// No manifest in this directory - skip
			continue
		}

		// Merge scanner results from component manifest
		for scannerType, result := range componentMf.Scans {
			// Prefix scanner with component name for uniqueness
			key := fmt.Sprintf("%s/%s", entry.Name(), scannerType)
			moduleMf.AddScannerResult(key, result)
			totalDuration += result.DurationSeconds
		}

		// Merge artifacts with component prefix
		for _, artifact := range componentMf.Artifacts {
			artifact.Path = filepath.Join(entry.Name(), artifact.Path)
			artifact.ID = fmt.Sprintf("%s-%s", entry.Name(), artifact.ID)
			moduleMf.AddArtifact(artifact)
		}
	}

	// Only save if we found any scan results
	if len(moduleMf.Scans) == 0 {
		return nil
	}

	moduleMf.DurationSeconds = totalDuration

	// Save module-level manifest
	if err := moduleMf.Save(moduleScanDir); err != nil {
		return fmt.Errorf("failed to save aggregated manifest: %w", err)
	}

	log.Debugf("Aggregated scan manifest for %s: %d scanners, %d artifacts",
		moniker, len(moduleMf.Scans), len(moduleMf.Artifacts))

	return nil
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
		SetOutputDir(scanCfg.ScanOutDir).
		SetFlatExecution(true) // Scans run in parallel (no layered execution)

	ctx.InitSummary = summary
}

// GetDockerImage returns the appropriate Docker image for a scanner type.
func GetDockerImage(ctx *cmdframework.ExecutionContext, scannerType internal.ScannerType) string {
	switch scannerType {
	case internal.ScannerSBOM, internal.ScannerVuln, internal.ScannerSecrets, internal.ScannerIaC, internal.ScannerCompliance:
		return ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
	case internal.ScannerSAST:
		return ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage()
	case internal.ScannerDAST:
		return ctx.EACConfig.SecurityTools.DockerImages.ZAP.FullImage()
	default:
		return ""
	}
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
		Layered:     false, // Scan uses parallel execution
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

	// Read skip_modules from security-tools.yml and populate Extra["skipMonikers"]
	// The framework's applySkipFilter will use this list during module resolution
	if ctx.EACConfig != nil {
		skipModules := ctx.EACConfig.SecurityTools.GetSkipModules()
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
	if ctx.Config.ForceRebuild {
		stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
		if err := stateMgr.ClearContext(workunit.ContextScan); err != nil {
			log.Warnf("Failed to clear scan state: %v", err)
		}
		return nil
	}

	// Incremental scan detection (devbox only, not CI)
	// Also run in dry-run mode to show which modules would be scanned/skipped
	if !environments.IsCI() && sctx != nil {
		detectIncrementalScanChanges(ctx, sctx)
	}

	return nil
}

// detectIncrementalScanChanges detects which modules need rescanning.
// Instead of filtering modules from the execution plan, it stores which modules
// are cached so the component worker can skip them with -1 (blue in TUI).
// This follows the same pattern as build's detectIncrementalChanges.
func detectIncrementalScanChanges(ctx *cmdframework.ExecutionContext, sctx *scanContext) {
	startTime := time.Now()
	defer func() {
		ctx.SetChangeDetectionTiming(time.Since(startTime))
	}()

	// Collect modules for change detection
	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		return
	}

	// Build module files map for hash computation
	moduleFiles := make(map[string][]string)
	for _, moniker := range monikers {
		if contract, ok := ctx.ModuleRegistry.Get(moniker); ok {
			patterns := contract.GetGlobPatterns()
			files, err := hash.ExpandGlobPatterns(ctx.WorkspaceRoot, patterns)
			if err != nil {
				log.Debugf("Failed to expand patterns for %s: %v", moniker, err)
				continue
			}
			moduleFiles[moniker] = files
		}
	}

	// Use StateManager for change detection
	stateMgr := workunit.NewStateManager(ctx.WorkspaceRoot)
	rule := workunit.DefaultRules[workunit.ContextScan]

	// Create hash provider that computes hash from expanded files
	hashProvider := func(module string) (string, error) {
		files, ok := moduleFiles[module]
		if !ok {
			return "", fmt.Errorf("no files for module %s", module)
		}
		return hash.Files(ctx.WorkspaceRoot, files)
	}

	changeResult, err := stateMgr.DetectModuleChanges(workunit.ContextScan, monikers, rule, hashProvider)
	if err != nil {
		log.Debugf("Failed to detect scan changes: %v", err)
		return
	}

	log.Debugf("[SCAN-CACHE] DetectChanges result: FreshRun=%v Changed=%d UpToDate=%d",
		changeResult.FreshRun, len(changeResult.ChangedModules), len(changeResult.UpToDateModules))
	for moniker, reason := range changeResult.ChangeReasons {
		log.Debugf("[SCAN-CACHE] Changed: %s -> %s", moniker, reason)
	}

	detectionTime := time.Since(startTime)

	if changeResult.FreshRun {
		log.Debugf("Fresh scan detected, all modules will scan")
		if ctx.InitSummary != nil {
			ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
				Enabled:       true,
				DetectionTime: detectionTime,
				FreshBuild:    true,
			})
		}
		return
	}

	// Build set of changed modules
	changedSet := make(map[string]bool)
	for _, m := range changeResult.ChangedModules {
		changedSet[m] = true
	}

	// Build set of cached modules (modules that are up-to-date)
	// These will be skipped at the component worker level, not filtered from the plan.
	sctx.cachedModules = make(map[string]bool)
	var changedList []string
	var cachedList []string

	for _, moniker := range monikers {
		if changedSet[moniker] {
			changedList = append(changedList, moniker)
		} else {
			sctx.cachedModules[moniker] = true
			cachedList = append(cachedList, moniker)
		}
	}

	if ctx.InitSummary != nil {
		ctx.InitSummary.SetIncremental(&initsummary.IncrementalInfo{
			Enabled:       true,
			DetectionTime: detectionTime,
			Changed:       changedList,
			UpToDate:      cachedList,
			FreshBuild:    false,
		})
	}

	log.Debugf("Incremental scan: %d modules to scan, %d cached (will show blue in TUI)",
		len(changedList), len(cachedList))

	// Debug: Log the cached modules map keys
	if len(sctx.cachedModules) > 0 {
		var keys []string
		for k := range sctx.cachedModules {
			keys = append(keys, k)
		}
		log.Debugf("[SCAN-CACHE] cachedModules set with %d entries: %v", len(keys), keys)
	} else {
		log.Debugf("[SCAN-CACHE] cachedModules is empty or nil")
	}
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
		TrivyImage:         ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage(),
		SemgrepImage:       ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage(),
		ZAPImage:           ctx.EACConfig.SecurityTools.DockerImages.ZAP.FullImage(),
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
		SetOutputDir(ctx.EACConfig.Repository.Paths.Out.Scan).
		SetFlatExecution(true)
	// Component count is set in multiScanAfterResolve after ModuleRegistry is available

	ctx.InitSummary = summary
}
