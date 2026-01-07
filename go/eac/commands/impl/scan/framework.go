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

// ScanFrameworkConfig holds scan-specific configuration for the framework
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

// ScanModuleResult holds scan results for a single module
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

// scanAfterInit handles scan-specific initialization
func scanAfterInit(ctx *cmdframework.ExecutionContext) error {
	scanCfg := ctx.Config.Extra["scanConfig"].(*ScanFrameworkConfig)

	// Set security output directory from config
	scanCfg.ScanOutDir = ctx.EACConfig.Repository.Paths.Out.Scan

	// Get git commit
	scanCfg.GitCommit = internal.GetGitCommit(ctx.WorkspaceRoot)

	// Build init summary
	buildScanInitSummary(ctx, scanCfg)

	return nil
}

// scanAfterExecute handles scan manifest updates
func scanAfterExecute(ctx *cmdframework.ExecutionContext) error {
	// Manifests are updated per-module in the worker
	// This hook is available for aggregate operations if needed
	return nil
}

// scanWorkerWrapper wraps the scanner-specific worker for cmdframework
func scanWorkerWrapper(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	scanCfg := ctx.Config.Extra["scanConfig"].(*ScanFrameworkConfig)
	worker := ctx.Config.Extra["scanWorker"].(ScanWorker)

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

// handleScanFailure handles a failed scan
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

// handleScanSuccess handles a successful scan
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

// updateScanManifest updates the scan manifest for a module
func updateScanManifest(ctx *cmdframework.ExecutionContext, scanCfg *ScanFrameworkConfig, module *modules.ModuleContract, moniker string, scanStart time.Time, status, evidencePath, errorMsg string) {
	duration := time.Since(scanStart)
	moduleScanDir := ctx.EACConfig.Repository.ScanModuleOutputPathAbs(ctx.WorkspaceRoot, moniker)

	mf, err := manifest.LoadOrCreateScanManifest(moduleScanDir, moniker, module.Type, scanCfg.GitCommit)
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

// buildScanInitSummary creates the init summary for scan commands
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

// GetDockerImage returns the appropriate Docker image for a scanner type
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

// CreateCommandConfig creates a standard command config for scan commands
func CreateCommandConfig(scannerType internal.ScannerType, scannerName string, monikers []string, debug bool, useTUI bool, tuiHeight int) *cmdframework.CommandConfig {
	return &cmdframework.CommandConfig{
		Type:       cmdframework.CommandTypeScan,
		ActionVerb: fmt.Sprintf("Scanning (%s)", scannerName),
		OutputDir:  "out/scan",
		LogFileName: fmt.Sprintf("%s.log", scannerType),
		Monikers:   monikers,
		Layered:    false, // Scan uses parallel execution
		UseTUI:     useTUI,
		TUIHeight:  tuiHeight,
		DebugMode:  debug,
	}
}

// MultiScanConfig holds configuration for running multiple scanners
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
		AfterInit: multiScanAfterInit,
	}

	return cmdframework.Run(cmdCfg, multiScanWorker, hooks)
}

// multiScanAfterInit resolves scanners to run based on module types and sets up skip list
func multiScanAfterInit(ctx *cmdframework.ExecutionContext) error {
	multiCfg := ctx.Config.Extra["multiScanConfig"].(*MultiScanConfig)

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

// multiScanWorker runs all configured scanners for a single module
func multiScanWorker(ctx *cmdframework.ExecutionContext, moniker string, logWriter io.Writer) int {
	multiCfg := ctx.Config.Extra["multiScanConfig"].(*MultiScanConfig)

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
		// Get default scanners for this module's type
		defaultScanners := ctx.EACConfig.SecurityTools.GetDefaultScanners(module.Type)
		for _, s := range defaultScanners {
			if scannerType, valid := internal.ParseScannerType(s); valid {
				scanners = append(scanners, scannerType)
			}
		}
	}

	if len(scanners) == 0 {
		output.Writeln(logWriter, "⚠️  No scanners configured for module type: %s", module.Type)
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

// runSingleScanner runs a single scanner for a module
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

// runScanner dispatches to the appropriate scanner implementation
func runScanner(ctx *cmdframework.ExecutionContext, module *modules.ModuleContract, scannerType internal.ScannerType, multiCfg *MultiScanConfig, logWriter io.Writer) (interface{}, error) {
	switch scannerType {
	case internal.ScannerSBOM:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		output.Writeln(logWriter, "  Format: %s", multiCfg.SBOMFormat)
		return internal.RunTrivySBOM(ctx.WorkspaceRoot, module.Files.Root, multiCfg.SBOMFormat, trivyImage)

	case internal.ScannerVuln:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		if len(multiCfg.VulnSeverities) > 0 {
			output.Writeln(logWriter, "  Severity filter: %v", multiCfg.VulnSeverities)
		}
		return internal.RunTrivyVuln(module.Files.Root, multiCfg.VulnSeverities, trivyImage)

	case internal.ScannerSecrets:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		return internal.RunTrivySecrets(module.Files.Root, trivyImage)

	case internal.ScannerIaC:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		return internal.RunTrivyIaC(module.Files.Root, trivyImage)

	case internal.ScannerCompliance:
		trivyImage := ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
		output.Writeln(logWriter, "  Using Trivy image: %s", trivyImage)
		output.Writeln(logWriter, "  Compliance standard: %s", multiCfg.ComplianceStandard)
		return internal.RunTrivyCompliance(module.Files.Root, multiCfg.ComplianceStandard, trivyImage)

	case internal.ScannerSAST:
		semgrepImage := ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage()
		output.Writeln(logWriter, "  Using Semgrep image: %s", semgrepImage)
		output.Writeln(logWriter, "  Config: %s", multiCfg.SemgrepConfig)
		return internal.RunSemgrepSAST(ctx.WorkspaceRoot, module.Files.Root, multiCfg.SemgrepConfig, semgrepImage)

	case internal.ScannerDAST:
		return nil, fmt.Errorf("ZAP scanner requires --target URL flag")

	default:
		return nil, fmt.Errorf("unknown scanner type: %s", scannerType)
	}
}

// getScannerEmoji returns the emoji for a scanner type
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

// buildMultiScanInitSummary creates the init summary for multi-scan
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

	ctx.InitSummary = summary
}
