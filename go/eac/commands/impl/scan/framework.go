// Package scan provides the scan command implementation using cmdframework.
package scan

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locking"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/manifest"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

func init() {
	// Register scan component-level execution support
	cmdframework.SetScanComponentWorkProvider(FlattenModulesToScanComponentWork)
	cmdframework.SetScanComponentWorker(scanComponentWorker)
}

// scannerSemaphores limits concurrent executions per scanner type.
// Each scanner type (trivy, semgrep, zap) gets its own semaphore with capacity 2.
// This allows some parallelism while preventing resource contention.
var (
	scannerSemaphores        = make(map[internal.ScannerType]chan struct{})
	semaphoreMu              sync.Mutex
	scannerSemaphoreCapacity = 3 // Max concurrent executions per scanner type
)

// getScannerSemaphore returns the semaphore for a scanner type, creating it if needed.
func getScannerSemaphore(scannerType internal.ScannerType) chan struct{} {
	semaphoreMu.Lock()
	defer semaphoreMu.Unlock()

	if sem, exists := scannerSemaphores[scannerType]; exists {
		return sem
	}

	// Create a semaphore with capacity 2 (two concurrent executions per scanner type)
	sem := make(chan struct{}, scannerSemaphoreCapacity)
	scannerSemaphores[scannerType] = sem
	return sem
}

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

// scanAfterExecute handles scan manifest updates.
func scanAfterExecute(ctx *cmdframework.ExecutionContext) error {
	// Manifests are updated per-module in the worker
	// This hook is available for aggregate operations if needed
	return nil
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

	// Acquire lock for this module
	lockCfg := locking.ScanConfig(moniker, paths.OutSecurityRelPath) // OutSecurityRelPath = "out/scan"
	lockFile, err := locking.Acquire(ctx.WorkspaceRoot, lockCfg)
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer locking.Release(lockFile)

	// Acquire scanner semaphore - only one of each scanner type runs at a time
	sem := getScannerSemaphore(scanCfg.ScannerType)
	output.Writeln(logWriter, "%s Waiting for %s scanner slot...", scanCfg.ScannerEmoji, scanCfg.ScannerType)
	sem <- struct{}{}
	output.Writeln(logWriter, "%s Acquired %s scanner slot", scanCfg.ScannerEmoji, scanCfg.ScannerType)
	defer func() {
		<-sem
		output.Writeln(logWriter, "%s Released %s scanner slot", scanCfg.ScannerEmoji, scanCfg.ScannerType)
	}()

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
		handleScanFailure(ctx, scanCfg, module, moniker, scanStart, err, logWriter)
		scanCfg.ScanResults[moniker] = result
		return 1
	}

	// Write evidence and update manifest
	evidencePath, writeErr := handleScanSuccess(ctx, scanCfg, module, moniker, scanStart, findings, logWriter)
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

// scanComponentWorker runs scans for a single component using component-level execution.
// This is called by the ComponentScheduler for parallel scan component execution.
func scanComponentWorker(ctx *cmdframework.ExecutionContext, moniker, component string, logWriter io.Writer) int {
	// Get multi-scan config if available (contains scanner settings)
	multiCfg, _ := ctx.Config.Extra["multiScanConfig"].(*MultiScanConfig)
	if multiCfg == nil {
		multiCfg = &MultiScanConfig{
			SBOMFormat:     "cyclonedx-json",
			SemgrepConfig:  "auto",
			VulnSeverities: nil,
		}
	}

	// Get module contract
	module, exists := ctx.ModuleRegistry.Get(moniker)
	if !exists {
		output.Writeln(logWriter, "Error: module not found: %s", moniker)
		return 1
	}

	// Get component type for this component
	compTypeName := module.Components.GetComponentType(component)

	// Acquire lock for this component
	lockCfg := locking.ComponentScanConfig(moniker, component, paths.OutSecurityRelPath)
	lockFile, err := locking.Acquire(ctx.WorkspaceRoot, lockCfg)
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer locking.Release(lockFile)

	// Get default scanners for this component type
	var scanners []internal.ScannerType
	seenScanners := make(map[string]bool)
	defaultScanners := ctx.EACConfig.SecurityTools.GetDefaultScanners(compTypeName)
	for _, s := range defaultScanners {
		if !seenScanners[s] {
			if scannerType, valid := internal.ParseScannerType(s); valid {
				scanners = append(scanners, scannerType)
				seenScanners[s] = true
			}
		}
	}

	if len(scanners) == 0 {
		output.Writeln(logWriter, "⚠️  No scanners configured for component type: %s", compTypeName)
		return 0
	}

	// Get component root path
	componentRoot := module.Components.GetComponentRoot(component)
	if componentRoot == "" {
		output.Writeln(logWriter, "Error: no root found for component %s", component)
		return 1
	}

	// Run each scanner
	exitCode := 0
	for _, scannerType := range scanners {
		result := runComponentScanner(ctx, module, moniker, component, componentRoot, scannerType, multiCfg, logWriter)
		if result != 0 {
			exitCode = 1 // Mark as failed but continue with other scanners
		}
	}

	return exitCode
}

// runComponentScanner runs a single scanner for a component.
func runComponentScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, moniker, component, componentRoot string, scannerType internal.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) int {
	emoji := getScannerEmoji(scannerType)

	// Acquire scanner semaphore
	sem := getScannerSemaphore(scannerType)
	output.Writeln(logWriter, "%s Waiting for %s scanner slot...", emoji, scannerType)
	sem <- struct{}{}
	output.Writeln(logWriter, "%s Acquired %s scanner slot", emoji, scannerType)
	defer func() {
		<-sem
		output.Writeln(logWriter, "%s Released %s scanner slot", emoji, scannerType)
	}()

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
		handleComponentScanFailure(ctx, scanCfg, module, moniker, component, scanStart, err, logWriter)
		return 1
	}

	// Write evidence and update manifest
	_, writeErr := handleComponentScanSuccess(ctx, scanCfg, module, moniker, component, scanStart, findings, logWriter)
	if writeErr != nil {
		return 1
	}

	return 0
}

// runScannerOnPath runs a scanner on a specific path.
func runScannerOnPath(ctx *cmdframework.ExecutionContext, targetPath string, scannerType internal.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) (interface{}, error) {
	switch scannerType {
	case internal.ScannerSBOM:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		output.Writeln(logWriter, "  Format: %s", multiCfg.SBOMFormat)
		return internal.RunTrivySBOM(ctx.WorkspaceRoot, targetPath, multiCfg.SBOMFormat, trivyImage)

	case internal.ScannerVuln:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		return internal.RunTrivyVuln(targetPath, multiCfg.VulnSeverities, trivyImage)

	case internal.ScannerSecrets:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		return internal.RunTrivySecrets(targetPath, trivyImage)

	case internal.ScannerIaC:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		return internal.RunTrivyIaC(targetPath, trivyImage)

	case internal.ScannerSAST:
		semgrepImage := ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage()
		output.Writeln(logWriter, "  Using Semgrep image: %s", semgrepImage)
		output.Writeln(logWriter, "  Config: %s", multiCfg.SemgrepConfig)
		return internal.RunSemgrepSAST(ctx.WorkspaceRoot, targetPath, multiCfg.SemgrepConfig, semgrepImage)

	case internal.ScannerDAST:
		return nil, fmt.Errorf("ZAP scanner requires --target URL flag")

	default:
		return nil, fmt.Errorf("unknown scanner type: %s", scannerType)
	}
}

// handleComponentScanFailure handles a failed component scan.
func handleComponentScanFailure(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig, module *modules.ModuleContract, moniker, component string, scanStart time.Time, scanErr error, logWriter io.Writer) {
	output.Writeln(logWriter, "  ❌ Failed: %v", scanErr)

	// Write error evidence to component directory
	outputPath, writeErr := internal.WriteComponentErrorEvidence(ctx.WorkspaceRoot, moniker, component, scanCfg.ScannerType, scanErr.Error())
	if writeErr != nil {
		output.Writeln(logWriter, "  Failed to write error evidence: %v", writeErr)
	} else {
		output.Writeln(logWriter, "  📄 Error evidence: %s", outputPath)
	}

	// Update scan manifest
	updateComponentScanManifest(ctx, scanCfg, module, moniker, component, scanStart, manifest.ScanStatusFailed, outputPath, scanErr.Error())
}

// handleComponentScanSuccess handles a successful component scan.
func handleComponentScanSuccess(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig, module *modules.ModuleContract, moniker, component string, scanStart time.Time, findings interface{}, logWriter io.Writer) (string, error) {
	// Write evidence file to component directory
	outputPath, err := internal.WriteComponentEvidence(ctx.WorkspaceRoot, moniker, component, scanCfg.ScannerType, findings)
	if err != nil {
		output.Writeln(logWriter, "  ❌ Failed to write evidence: %v", err)
		updateComponentScanManifest(ctx, scanCfg, module, moniker, component, scanStart, manifest.ScanStatusFailed, "", err.Error())
		return "", err
	}

	// Update scan manifest
	updateComponentScanManifest(ctx, scanCfg, module, moniker, component, scanStart, manifest.ScanStatusPassed, outputPath, "")

	output.Writeln(logWriter, "  ✅ Success: %s", outputPath)
	return outputPath, nil
}

// updateComponentScanManifest updates the scan manifest for a component.
func updateComponentScanManifest(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig, module *modules.ModuleContract, moniker, component string, scanStart time.Time, status, evidencePath, errorMsg string) {
	duration := time.Since(scanStart)
	// Use component-level scan output directory
	componentScanDir := paths.ComponentScanOutputPath(ctx.WorkspaceRoot, moniker, component)

	mf, err := manifest.LoadOrCreateScanManifest(componentScanDir, moniker+"/"+component, module.Components.GetComponentType(component), scanCfg.GitCommit)
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

	if err := mf.Save(componentScanDir); err != nil {
		log.Warnf("Failed to save scan manifest: %v", err)
	}
}

// handleScanFailure handles a failed scan.
func handleScanFailure(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig, module *modules.ModuleContract, moniker string, scanStart time.Time, scanErr error, logWriter io.Writer) {
	output.Writeln(logWriter, "  ❌ Failed: %v", scanErr)

	// Write error evidence
	outputPath, writeErr := internal.WriteErrorEvidence(ctx.WorkspaceRoot, moniker, scanCfg.ScannerType, scanErr.Error())
	if writeErr != nil {
		output.Writeln(logWriter, "  Failed to write error evidence: %v", writeErr)
	} else {
		output.Writeln(logWriter, "  📄 Error evidence: %s", outputPath)
	}

	// Update scan manifest
	updateScanManifest(ctx, scanCfg, module, moniker, scanStart, manifest.ScanStatusFailed, outputPath, scanErr.Error())
}

// handleScanSuccess handles a successful scan.
func handleScanSuccess(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig, module *modules.ModuleContract, moniker string, scanStart time.Time, findings interface{}, logWriter io.Writer) (string, error) {
	// Write evidence file
	outputPath, err := internal.WriteEvidence(ctx.WorkspaceRoot, moniker, scanCfg.ScannerType, findings)
	if err != nil {
		output.Writeln(logWriter, "  ❌ Failed to write evidence: %v", err)
		updateScanManifest(ctx, scanCfg, module, moniker, scanStart, manifest.ScanStatusFailed, "", err.Error())
		return "", err
	}

	// Update scan manifest
	updateScanManifest(ctx, scanCfg, module, moniker, scanStart, manifest.ScanStatusPassed, outputPath, "")

	output.Writeln(logWriter, "  ✅ Success: %s", outputPath)
	return outputPath, nil
}

// updateScanManifest updates the scan manifest for a module.
func updateScanManifest(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig, module *modules.ModuleContract, moniker string, scanStart time.Time, status, evidencePath, errorMsg string) {
	duration := time.Since(scanStart)
	moduleScanDir := ctx.EACConfig.Repository.ScanModuleOutputPathAbs(ctx.WorkspaceRoot, moniker)

	mf, err := manifest.LoadOrCreateScanManifest(moduleScanDir, moniker, module.GetComponentTypesDisplay(), scanCfg.GitCommit)
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

	if err := mf.Save(moduleScanDir); err != nil {
		log.Warnf("Failed to save scan manifest: %v", err)
	}
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

	// Set up hooks for multi-scan
	hooks := &cmdframework.Hooks{
		AfterInit:    multiScanAfterInit,
		AfterResolve: multiScanAfterResolve,
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
func multiScanAfterResolve(ctx *cmdframework.ExecutionContext) error {
	// Now that ModuleRegistry is available, calculate and set component count
	if ctx.InitSummary != nil && ctx.ModuleRegistry != nil {
		componentCount := getScanComponentCount(ctx)
		ctx.InitSummary.SetComponentCount(componentCount)
	}
	return nil
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

	// Acquire lock for this module
	lockCfg := locking.ScanConfig(moniker, paths.OutSecurityRelPath) // OutSecurityRelPath = "out/scan"
	lockFile, err := locking.Acquire(ctx.WorkspaceRoot, lockCfg)
	if err != nil {
		output.Writeln(logWriter, "Error: %v", err)
		return 1
	}
	defer locking.Release(lockFile)

	// Determine which scanners to run
	var scanners []internal.ScannerType
	if len(multiCfg.Scanners) > 0 {
		scanners = multiCfg.Scanners
	} else {
		// Get default scanners for each of the module's package types
		seenScanners := make(map[string]bool)
		for _, pkgType := range module.GetEnabledComponents() {
			defaultScanners := ctx.EACConfig.SecurityTools.GetDefaultScanners(pkgType)
			for _, s := range defaultScanners {
				if !seenScanners[s] {
					if scannerType, valid := internal.ParseScannerType(s); valid {
						scanners = append(scanners, scannerType)
						seenScanners[s] = true
					}
				}
			}
		}
	}

	if len(scanners) == 0 {
		output.Writeln(logWriter, "⚠️  No scanners configured for module packages: %s", module.GetComponentTypesDisplay())
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

	// Acquire scanner semaphore - only one of each scanner type runs at a time
	sem := getScannerSemaphore(scannerType)
	output.Writeln(logWriter, "%s Waiting for %s scanner slot...", emoji, scannerType)
	sem <- struct{}{}
	output.Writeln(logWriter, "%s Acquired %s scanner slot", emoji, scannerType)
	defer func() {
		<-sem
		output.Writeln(logWriter, "%s Released %s scanner slot", emoji, scannerType)
	}()

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
		handleScanFailure(ctx, scanCfg, module, module.Moniker, scanStart, err, logWriter)
		return 1
	}

	// Write evidence and update manifest
	_, writeErr := handleScanSuccess(ctx, scanCfg, module, module.Moniker, scanStart, findings, logWriter)
	if writeErr != nil {
		return 1
	}

	return 0
}

// runScanner dispatches to the appropriate scanner implementation.
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

	switch scannerType {
	case internal.ScannerSBOM:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		output.Writeln(logWriter, "  Format: %s", multiCfg.SBOMFormat)
		return internal.RunTrivySBOM(ctx.WorkspaceRoot, moduleRoot, multiCfg.SBOMFormat, trivyImage)

	case internal.ScannerVuln:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		if len(multiCfg.VulnSeverities) > 0 {
			output.Writeln(logWriter, "  Severity filter: %v", multiCfg.VulnSeverities)
		}
		return internal.RunTrivyVuln(moduleRoot, multiCfg.VulnSeverities, trivyImage)

	case internal.ScannerSecrets:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		return internal.RunTrivySecrets(moduleRoot, trivyImage)

	case internal.ScannerIaC:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		return internal.RunTrivyIaC(moduleRoot, trivyImage)

	case internal.ScannerCompliance:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		output.Writeln(logWriter, "  Compliance standard: %s", multiCfg.ComplianceStandard)
		return internal.RunTrivyCompliance(moduleRoot, multiCfg.ComplianceStandard, trivyImage)

	case internal.ScannerSAST:
		semgrepImage := ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage()
		output.Writeln(logWriter, "  Using Semgrep image: %s", semgrepImage)
		output.Writeln(logWriter, "  Config: %s", multiCfg.SemgrepConfig)
		return internal.RunSemgrepSAST(ctx.WorkspaceRoot, moduleRoot, multiCfg.SemgrepConfig, semgrepImage)

	case internal.ScannerDAST:
		return nil, fmt.Errorf("ZAP scanner requires --target URL flag")

	default:
		return nil, fmt.Errorf("unknown scanner type: %s", scannerType)
	}
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
