package scan

import (
	"context"
	"fmt"
	"io"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/evidence"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/locking"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

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
func scanUnitWorker(goCtx context.Context, ctx *cmdframework.ExecutionContext, spec core.UnitSpec, logWriter io.Writer) int {
	// Get multi-scan config if available (contains scanner settings)
	multiCfg, _ := ctx.Config.MultiScanConfig.(*MultiScanConfig)
	if multiCfg == nil {
		multiCfg = &MultiScanConfig{
			SBOMFormat:     "cyclonedx-json",
			SemgrepConfig:  "auto",
			VulnSeverities: nil,
		}
	}

	// Extract identity from spec - no more string parsing
	moniker := spec.ID.Module
	compName := spec.ID.ComponentName
	scannerTypeStr := spec.ID.Tool
	component := compName + ":" + scannerTypeStr // For display/logging

	// Get scan context for caching
	sctx, _ := ctx.Config.ScanCmdContext.(*scanContext)

	// Set up shared pipeline for cache + lock orchestration
	pipeline := &cmdframework.UnitPipeline{
		CachedUoWs: nil,
		LockStyle:  cmdframework.AlwaysLock,
		LockConfigFn: func(m, cd string) locking.Config {
			return locking.UnitScanConfig(m, cd, paths.OutSecurityRelPath)
		},
	}
	if sctx != nil {
		pipeline.CachedUoWs = sctx.cachedUoWs
	}

	// Use spec's UnitID directly for cache lookup
	unitID := spec.ID

	// Check UoW-level cache first
	log.Debugf("[SCAN-UOW-CACHE] Component worker for %s: unitID=%s", component, unitID.Longname())
	if cacheResult := pipeline.CheckCache(ctx, unitID, logWriter); cacheResult != 0 {
		return cacheResult
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
	compType := ctx.EACConfig.ComponentKinds.Get(compTypeName)

	// Skip non-scannable component types
	if compType == nil || !compType.IsScannable() {
		output.Writeln(logWriter, "⚠️  Component type %s is not scannable, skipping", compTypeName)
		return 0
	}

	// Acquire lock for this unit (component+scanner) with wait
	componentDir := cmdframework.UnitDir(compName, scannerTypeStr)
	release, err := pipeline.AcquireLock(ctx, moniker, componentDir)
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer release()

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
		return runUnitScanner(ctx, module, moniker, compName, componentRoot, evidence.ScannerType(toolID), multiCfg, logWriter)
	}

	// 2-part key (no scanner specified): run all scanners from component type configuration
	var scannersList []evidence.ScannerType
	seenScanners := make(map[string]bool)
	for _, s := range compType.GetScanners() {
		if !seenScanners[s] {
			if toolID := tool.ScannerToolIDForCategory(s); toolID != "" {
				scannersList = append(scannersList, evidence.ScannerType(toolID))
				seenScanners[s] = true
			}
		}
	}

	if len(scannersList) == 0 {
		output.Writeln(logWriter, "⚠️  No scanners configured for component type: %s", compTypeName)
		return 0
	}

	// Run each scanner
	exitCode := 0
	for _, scannerType := range scannersList {
		result := runUnitScanner(ctx, module, moniker, compName, componentRoot, scannerType, multiCfg, logWriter)
		if result != 0 {
			exitCode = 1 // Mark as failed but continue with other scanners
		}
	}

	return exitCode
}

// runUnitScanner runs a single scanner for a component.
func runUnitScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, moniker, component, componentRoot string, scannerType evidence.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) int {
	emoji := getScannerEmoji(scannerType)

	// Get scan context for result recording
	sctx, _ := ctx.Config.ScanCmdContext.(*scanContext)

	// In dry-run mode, show what would happen without actually scanning
	if ctx.Config.DryRun {
		output.Writeln(logWriter, "🔍 %s/%s would be scanned (changed)", moniker, component)
		output.Writeln(logWriter, "   Scanner: %s", scannerType)
		return 0
	}

	// Parallelism controlled by orchestrator's weighted semaphore via component weights
	output.Writeln(logWriter, "%s Running %s scanner on %s/%s...", emoji, scannerType, moniker, component)
	scanStart := time.Now()

	// Use pre-computed input hash for cache consistency.
	// This ensures the hash written to the manifest matches the hash used for detection.
	var inputHash string
	if sctx != nil && sctx.moduleInputHashes != nil {
		inputHash = sctx.moduleInputHashes[moniker]
	}
	if inputHash == "" {
		inputHash, _ = computeScanInputHash(ctx, module)
	}

	// Create scan config for this scanner
	scanCfg := &ScanFrameworkConfig{
		ScannerType:  scannerType,
		ScannerName:  string(scannerType),
		ScannerEmoji: emoji,
		ScanOutDir:   ctx.EACConfig.Repository.Paths.Out.Scan,
		GitCommit:    getGitCommit(ctx.WorkspaceRoot),
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
		compTypeName := module.Components.GetComponentType(component)
		writeUoWScanManifest(ctx, sctx, moniker, compTypeName, component, string(scannerType), inputHash, scanStart)
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
func writeUoWScanManifest(ctx *cmdframework.ExecutionContext, sctx *scanContext, moniker, compType, component, tool, inputHash string, startTime time.Time) {
	// Initialize tracker if needed
	sctx.mu.Lock()
	if sctx.tracker == nil {
		sctx.tracker = coreoutput.NewTracker(ctx.WorkspaceRoot, core.ActionScan)
	}
	tracker := sctx.tracker
	sctx.mu.Unlock()

	// Build UnitID for the tracker
	unitID := workunit.UnitID{
		Action:        core.ActionScan,
		Module:        moniker,
		ComponentType: compType,
		ComponentName: component,
		Tool:          tool,
	}

	// Create and record the manifest
	manifest := &coreoutput.UoWManifest{
		Action:     core.ActionScan,
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
func logScannerConfig(scannerType evidence.ScannerType, logWriter io.Writer, multiCfg *MultiScanConfig, ctx *cmdframework.ExecutionContext) {
	switch scannerType {
	case evidence.ScannerSBOM:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
		output.Writeln(logWriter, "  Format: %s", multiCfg.SBOMFormat)
	case evidence.ScannerVuln:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
		if len(multiCfg.VulnSeverities) > 0 {
			output.Writeln(logWriter, "  Severity filter: %v", multiCfg.VulnSeverities)
		}
	case evidence.ScannerSecrets, evidence.ScannerIaC:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
	case evidence.ScannerCompliance:
		output.Writeln(logWriter, "  Using Trivy image: %s", getTrivyImage(ctx))
		output.Writeln(logWriter, "  Compliance standard: %s", multiCfg.ComplianceStandard)
	case evidence.ScannerSAST:
		output.Writeln(logWriter, "  Using Semgrep image: %s", getSemgrepImage(ctx))
		output.Writeln(logWriter, "  Config: %s", multiCfg.SemgrepConfig)
	case evidence.ScannerDAST:
		output.Writeln(logWriter, "  Using ZAP image: %s", getZAPImage(ctx))
	}
}

// runScannerOnPath runs a scanner on a specific path using the scanner registry.
func runScannerOnPath(ctx *cmdframework.ExecutionContext, targetPath string, scannerType evidence.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) (interface{}, error) {
	// ZAP requires special handling (target URL)
	if scannerType == evidence.ScannerDAST {
		return nil, fmt.Errorf("ZAP scanner requires --target URL flag")
	}

	// Log scanner-specific configuration
	logScannerConfig(scannerType, logWriter, multiCfg, ctx)

	// Get scanner from registry (YAML-defined tools)
	scanFn := tool.GlobalScanBridge().GetScannerByToolID(string(scannerType))
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
		outputPath, writeErr = evidence.WriteComponentErrorEvidence(
			ctx.WorkspaceRoot, moniker, component, scanCfg.ScannerType, scanErr.Error())
	} else {
		outputPath, writeErr = evidence.WriteErrorEvidence(
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
		outputPath, err = evidence.WriteComponentEvidence(
			ctx.WorkspaceRoot, moniker, component, scanCfg.ScannerType, findings)
	} else {
		outputPath, err = evidence.WriteEvidence(
			ctx.WorkspaceRoot, moniker, scanCfg.ScannerType, findings)
	}

	if err != nil {
		output.Writeln(logWriter, "  ❌ Failed to write evidence: %v", err)
		return "", err
	}

	output.Writeln(logWriter, "  ✅ Success: %s", outputPath)
	return outputPath, nil
}

// multiScanWorker runs all configured scanners for a single module.
func multiScanWorker(_ context.Context, ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	multiCfg, ok := ctx.Config.MultiScanConfig.(*MultiScanConfig)
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
	var scannersList []evidence.ScannerType
	if len(multiCfg.Scanners) > 0 {
		scannersList = multiCfg.Scanners
	} else {
		// Get scanners from component types for each of the module's components
		seenScanners := make(map[string]bool)
		for _, componentName := range module.GetEnabledComponents() {
			compTypeName := module.Components.GetComponentType(componentName)
			compType := ctx.EACConfig.ComponentKinds.Get(compTypeName)
			if compType == nil || !compType.IsScannable() {
				continue // Skip non-scannable component types
			}
			for _, s := range compType.GetScanners() {
				if !seenScanners[s] {
					if toolID := tool.ScannerToolIDForCategory(s); toolID != "" {
						scannersList = append(scannersList, evidence.ScannerType(toolID))
						seenScanners[s] = true
					}
				}
			}
		}
	}

	if len(scannersList) == 0 {
		output.Writeln(logWriter, "⚠️  No scannable components in module: %s", module.GetComponentTypesDisplay())
		return 0
	}

	// Run each scanner sequentially
	exitCode := 0
	for _, scannerType := range scannersList {
		result := runSingleScanner(ctx, module, scannerType, multiCfg, logWriter)
		if result != 0 {
			exitCode = 1 // Mark as failed but continue with other scanners
		}
	}

	return exitCode
}

// runSingleScanner runs a single scanner for a module.
func runSingleScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, scannerType evidence.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) int {
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
		GitCommit:    getGitCommit(ctx.WorkspaceRoot),
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
func runScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, scannerType evidence.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) (interface{}, error) {
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
	if scannerType == evidence.ScannerDAST {
		return nil, fmt.Errorf("ZAP scanner requires --target URL flag")
	}

	// Log scanner-specific configuration
	logScannerConfig(scannerType, logWriter, multiCfg, ctx)

	// Get scanner from registry (YAML-defined tools)
	scanFn := tool.GlobalScanBridge().GetScannerByToolID(string(scannerType))
	if scanFn == nil {
		return nil, fmt.Errorf("no scanner found for type: %s", scannerType)
	}

	// Execute scanner
	return scanFn(ctx.WorkspaceRoot, moduleRoot, "", logWriter, tool.ScanOptions{
		ScanType: string(scannerType),
	})
}

// getScannerEmoji returns the emoji for a scanner type.
func getScannerEmoji(scannerType evidence.ScannerType) string {
	switch scannerType {
	case evidence.ScannerSBOM:
		return "📦"
	case evidence.ScannerVuln:
		return "🔍"
	case evidence.ScannerSecrets:
		return "🔐"
	case evidence.ScannerIaC:
		return "🏗️"
	case evidence.ScannerCompliance:
		return "✅"
	case evidence.ScannerSAST:
		return "🔬"
	case evidence.ScannerDAST:
		return "🌐"
	default:
		return "🔒"
	}
}
