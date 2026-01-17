// Package internal provides shared types and utilities for security scanners.
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/manifest"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"go.uber.org/zap"
)

// log is declared in docker.go

// ScanConfig holds the configuration for a scan command.
type ScanConfig struct {
	// ScannerType identifies the scanner (sbom, vuln, secrets, etc.)
	ScannerType ScannerType

	// ScannerName is the human-readable name for the scanner
	ScannerName string

	// ScannerEmoji is the emoji used for console output (e.g., "📦", "🔍", "🔐")
	ScannerEmoji string

	// Debug enables debug logging
	Debug bool

	// Monikers is the list of module monikers to scan
	Monikers []string

	// ArgOffset is the offset into os.Args where scanner-specific args start
	// Typically 3 for "scan <scanner> [args...]"
	ArgOffset int

	// CustomArgs holds scanner-specific arguments parsed by the caller
	CustomArgs interface{}
}

// ScanContext holds the runtime context for a scan operation.
type ScanContext struct {
	Config        *ScanConfig
	WorkspaceRoot string
	Logger        *logging.Logger
	EACConfig     *config.EACConfig
	ModuleReport  *reports.ModuleContractReport
	GitCommit     string
	ScanOutDir    string
}

// ModuleScanContext provides context for scanning a specific module.
type ModuleScanContext struct {
	*ScanContext
	Moniker       string
	Module        *modules.ModuleContract
	ModuleScanDir string
	ScanStart     time.Time
}

// ScanWorkerFunc is the function signature for module scan workers
// Returns: (findings interface{} or nil, error)
// If findings is nil and error is nil, the scan is skipped.
type ScanWorkerFunc func(ctx *ModuleScanContext) (interface{}, error)

// Run executes the scan framework with the given configuration and worker.
func Run(cfg *ScanConfig, worker ScanWorkerFunc) int {
	ctx, err := initialize(cfg)
	if err != nil {
		log.Errorf("Initialization failed: %v", err)
		return 1
	}
	defer func() { _ = ctx.Logger.Sync() }() //nolint:errcheck // best-effort sync

	// If no monikers provided, default to all modules
	if len(cfg.Monikers) == 0 {
		for _, module := range ctx.ModuleReport.Registry.All() {
			cfg.Monikers = append(cfg.Monikers, module.Moniker)
		}
	}

	// Display init summary
	displayInitSummary(ctx)

	// Scan each module
	exitCode := 0
	successCount := 0
	failureCount := 0

	for _, moniker := range cfg.Monikers {
		module, exists := ctx.ModuleReport.Registry.Get(moniker)
		if !exists {
			ctx.Logger.Error("Module not found", zap.String("moniker", moniker))
			log.Errorf("module not found: %s", moniker)
			failureCount++
			exitCode = 1
			continue
		}

		moduleCtx := &ModuleScanContext{
			ScanContext:   ctx,
			Moniker:       moniker,
			Module:        module,
			ModuleScanDir: filepath.Join(ctx.WorkspaceRoot, ctx.ScanOutDir, moniker),
			ScanStart:     time.Now(),
		}

		ctx.Logger.Info("Scanning module", zap.String("moniker", moniker), zap.String("root", module.Files.Root))
		log.Infof("%s Scanning %s...", cfg.ScannerEmoji, moniker)

		findings, err := worker(moduleCtx)
		if err != nil {
			handleScanFailure(moduleCtx, err)
			failureCount++
			exitCode = 1
			continue
		}

		if err := handleScanSuccess(moduleCtx, findings); err != nil {
			failureCount++
			exitCode = 1
			continue
		}

		successCount++
	}

	// Print summary
	log.Info("")
	ctx.Logger.Info(cfg.ScannerName+" scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(cfg.Monikers)))

	log.Infof("Summary: %d succeeded, %d failed, %d total", successCount, failureCount, len(cfg.Monikers))

	return exitCode
}

// initialize sets up the scan context with workspace, logger, config, and modules.
func initialize(cfg *ScanConfig) (*ScanContext, error) {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, err
	}

	var logger *logging.Logger
	if cfg.Debug {
		logger, err = logging.NewWithDebug("security", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("security", workspaceRoot)
	}
	if err != nil {
		return nil, err
	}

	eacCfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		logger.Error("Failed to load configuration", zap.Error(err))
		return nil, err
	}

	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		logger.Error("Failed to load module contracts", zap.Error(err))
		return nil, err
	}

	logger.Info("Starting "+cfg.ScannerName+" scanner",
		zap.Strings("modules", cfg.Monikers),
		zap.Bool("debug", cfg.Debug))

	return &ScanContext{
		Config:        cfg,
		WorkspaceRoot: workspaceRoot,
		Logger:        logger,
		EACConfig:     eacCfg,
		ModuleReport:  moduleReport,
		GitCommit:     GetGitCommit(workspaceRoot),
		ScanOutDir:    eacCfg.Repository.Paths.Out.Scan,
	}, nil
}

// handleScanFailure handles a failed scan by writing error evidence and updating manifest.
func handleScanFailure(ctx *ModuleScanContext, scanErr error) {
	ctx.Logger.Error(ctx.Config.ScannerName+" scan failed",
		zap.String("moniker", ctx.Moniker),
		zap.Error(scanErr))
	log.Errorf("  ❌ Failed: %v", scanErr)

	// Write error evidence
	outputPath, writeErr := WriteErrorEvidence(ctx.WorkspaceRoot, ctx.Moniker, ctx.Config.ScannerType, scanErr.Error())
	if writeErr != nil {
		ctx.Logger.Error("Failed to write error evidence", zap.Error(writeErr))
	} else {
		ctx.Logger.Info("Error evidence written", zap.String("path", outputPath))
		log.Infof("  📄 Error evidence: %s", outputPath)
	}

	// Update scan manifest with failure
	UpdateScanManifest(ctx, manifest.ScanStatusFailed, outputPath, scanErr.Error())
}

// handleScanSuccess handles a successful scan by writing evidence and updating manifest.
func handleScanSuccess(ctx *ModuleScanContext, findings interface{}) error {
	// Write evidence file
	outputPath, err := WriteEvidence(ctx.WorkspaceRoot, ctx.Moniker, ctx.Config.ScannerType, findings)
	if err != nil {
		ctx.Logger.Error("Failed to write evidence",
			zap.String("moniker", ctx.Moniker),
			zap.Error(err))
		log.Errorf("  ❌ Failed to write evidence: %v", err)

		// Update scan manifest with failure
		UpdateScanManifest(ctx, manifest.ScanStatusFailed, "", err.Error())
		return err
	}

	// Update scan manifest with success
	UpdateScanManifest(ctx, manifest.ScanStatusPassed, outputPath, "")

	ctx.Logger.Info(ctx.Config.ScannerName+" scan completed",
		zap.String("moniker", ctx.Moniker),
		zap.String("evidence", outputPath))
	log.Infof("  ✅ Success: %s", outputPath)
	return nil
}

// GetGitCommit retrieves the current git commit SHA.
// In CI (GITHUB_SHA set), uses the environment variable.
// Locally, runs git rev-parse HEAD.
func GetGitCommit(workspaceRoot string) string {
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		return sha
	}
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workspaceRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// UpdateScanManifest loads or creates the scan manifest, adds the scanner result, and saves it.
func UpdateScanManifest(ctx *ModuleScanContext, status, evidencePath, errorMsg string) {
	duration := time.Since(ctx.ScanStart)

	mf, err := manifest.LoadOrCreateScanManifest(ctx.ModuleScanDir, ctx.Moniker, ctx.Module.Type, ctx.GitCommit)
	if err != nil {
		ctx.Logger.Warn("Failed to load/create scan manifest", zap.Error(err))
		return
	}

	result := manifest.ScannerResult{
		Status:          status,
		RunTime:         time.Now(),
		DurationSeconds: duration.Seconds(),
		EvidencePath:    evidencePath,
		Error:           errorMsg,
	}
	mf.AddScannerResult(string(ctx.Config.ScannerType), result)

	if err := mf.Save(ctx.ModuleScanDir); err != nil {
		ctx.Logger.Warn("Failed to save scan manifest", zap.Error(err))
		return
	}

	ctx.Logger.Debug("Scan manifest updated", zap.String("path", manifest.GetScanManifestPath(ctx.ModuleScanDir)))
}

// TrivyImage returns the configured Trivy Docker image.
func (ctx *ScanContext) TrivyImage() string {
	return ctx.EACConfig.SecurityTools.DockerImages.Trivy.FullImage()
}

// SemgrepImage returns the configured Semgrep Docker image.
func (ctx *ScanContext) SemgrepImage() string {
	return ctx.EACConfig.SecurityTools.DockerImages.Semgrep.FullImage()
}

// ZapImage returns the configured ZAP Docker image.
func (ctx *ScanContext) ZapImage() string {
	return ctx.EACConfig.SecurityTools.DockerImages.ZAP.FullImage()
}

// displayInitSummary outputs the initialization summary for scan commands.
func displayInitSummary(ctx *ScanContext) {
	cfg := ctx.Config

	log.Info("═══ Scan Initialization ═══")
	log.Infof("Scanner: %s", cfg.ScannerName)
	log.Infof("Execution Context: %s", detectExecutionContext())

	// Modules
	log.Info("── Modules ──")
	if len(cfg.Monikers) <= 5 {
		log.Infof("  Target: %d (%s)", len(cfg.Monikers), strings.Join(cfg.Monikers, ", "))
	} else {
		first5 := strings.Join(cfg.Monikers[:5], ", ")
		log.Infof("  Target: %d (%s, ...)", len(cfg.Monikers), first5)
	}

	// Execution plan
	log.Info("── Execution Plan ──")
	log.Infof("  Mode: Sequential (all modules)")

	// Flags
	log.Info("── Flags ──")
	if cfg.Debug {
		log.Info("  🐛 debug: enabled")
	} else {
		log.Info("  (defaults)")
	}

	log.Info("═══════════════════════════════")
	log.Info("")
}

// detectExecutionContext returns a human-readable execution context string.
func detectExecutionContext() string {
	env := "devbox"
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		env = "CI"
	}
	return env
}
