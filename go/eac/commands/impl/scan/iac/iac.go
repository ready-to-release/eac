// Command: scan iac
// Short: Scan Infrastructure as Code for misconfigurations using Trivy
// Long: Scan Infrastructure as Code (IaC) files for security misconfigurations using Trivy.
// Long:
// Long: This command detects security issues in IaC files including Terraform, CloudFormation,
// Long: Kubernetes manifests, Dockerfiles, and Helm charts. Results are saved as timestamped
// Long: evidence files with SHA256 integrity verification for audit compliance.
// Long:
// Long: Supported IaC types: Terraform, CloudFormation, Kubernetes, Docker, Helm
// Long:
// Long: Output: out/security/<module>/iac/<timestamp>.json
// Long:
// Long: Example:
// Long:   security iac                           # All modules
// Long:   security iac eac-core                  # Single module
// Long:   security iac eac-core r2r-cli          # Multiple modules
// Long:   security iac eac-core --debug          # Debug logging
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package iac

import (
	"github.com/ready-to-release/eac/go/eac/core/config"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"go.uber.org/zap"
)

var log = logging.C()

func init() {
	registry.Register(IaC)
}

// IaC command entry point
func IaC() int {
	args := os.Args[3:] // Skip program name, "security", and "iac"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printIaCUsage()
		return 0
	}

	// Parse module monikers and flags
	var monikers []string
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debug = true
		default:
			if strings.HasPrefix(arg, "--") {
				log.Errorf("unknown flag: %s", arg)
				printIaCUsage()
				return 1
			} else {
				monikers = append(monikers, arg)
			}
		}
	}

	// Initialize logger
	var logger *logging.Logger
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	if debug {
		logger, err = logging.NewWithDebug("security", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("security", workspaceRoot)
	}
	if err != nil {
		log.Errorf("failed to initialize logger: %v", err)
		return 1
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		logger.Error("Failed to load configuration", zap.Error(err))
		log.Errorf(" failed to load configuration: %v\n", err)
		return 1
	}

	// Get Docker image from config
	trivyImage := cfg.SecurityTools.DockerImages.Trivy.FullImage()
	logger.Debug("Using Trivy image", zap.String("image", trivyImage))

	logger.Info("Starting IaC scanner",
		zap.Strings("modules", monikers),
		zap.Bool("debug", debug))

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		logger.Error("Failed to load module contracts", zap.Error(err))
		log.Errorf("failed to load module contracts: %v", err)
		return 1
	}

	// If no monikers provided, default to all modules
	if len(monikers) == 0 {
		log.Info("ℹ️  No modules specified, scanning all modules...")
		logger.Info("No modules specified, using all modules")
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Scan each module
	exitCode := 0
	successCount := 0
	failureCount := 0

	for _, moniker := range monikers {
		// Get module from registry
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			logger.Error("Module not found", zap.String("moniker", moniker))
			log.Errorf("module not found: %s", moniker)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Scanning module", zap.String("moniker", moniker), zap.String("root", module.Files.Root))
		log.Infof("🏗️  Scanning %s...", moniker)

		// Run Trivy IaC scan
		findings, err := internal.RunTrivyIaC(module.Files.Root, trivyImage, logger)
		if err != nil {
			logger.Error("IaC scan failed", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed: %v", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerIaC, err.Error())
			if writeErr != nil {
				logger.Error("Failed to write error evidence", zap.Error(writeErr))
			} else {
				logger.Info("Error evidence written", zap.String("path", outputPath))
				log.Infof("  📄 Error evidence: %s", outputPath)
			}

			failureCount++
			exitCode = 1
			continue
		}

		// Write evidence file
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerIaC, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed to write evidence: %v", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("IaC scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		log.Infof("  ✅ Success: %s", outputPath)
		successCount++
	}

	// Print summary
	log.Info("")
	logger.Info("IaC scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	log.Infof("Summary: %d succeeded, %d failed, %d total", successCount, failureCount, len(monikers))

	return exitCode
}

func printIaCUsage() {
	log.Info("Scan Infrastructure as Code for misconfigurations using Trivy")
	log.Info("")
	log.Info("Usage: security iac [modules...] [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  [modules...]          One or more module monikers to scan")
	log.Info("                        If no modules specified, scans all modules")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --debug, -d           Enable debug logging")
	log.Info("")
	log.Info("Supported IaC types:")
	log.Info("  - Terraform (.tf)")
	log.Info("  - CloudFormation (.yaml, .yml, .json)")
	log.Info("  - Kubernetes manifests (.yaml, .yml)")
	log.Info("  - Dockerfiles")
	log.Info("  - Helm charts")
	log.Info("")
	log.Info("Examples:")
	log.Info("  security iac                           # All modules")
	log.Info("  security iac eac-core                  # Single module")
	log.Info("  security iac eac-core r2r-cli          # Multiple modules")
	log.Info("  security iac eac-core --debug          # Debug logging")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/security/<module>/iac/<timestamp>.json")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses Trivy (Apache 2.0). See the NOTICE file in the")
	log.Info("  repository root for full attribution and licensing information.")
}
