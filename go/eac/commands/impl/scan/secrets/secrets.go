// Command: scan secrets
// Short: Detect secrets and credentials using Trivy
// Long: Scan for hardcoded secrets and credentials in source code using Trivy.
// Long:
// Long: This command detects API keys, passwords, tokens, and other sensitive data
// Long: that may have been accidentally committed to the repository. Results are
// Long: saved as timestamped evidence files with SHA256 integrity verification.
// Long:
// Long: Expected Output:
// Long:   Timestamped JSON evidence files are written to out/security/<module>/secrets/<timestamp>.json
// Long:
// Long: Example:
// Long:   security secrets                       # All modules
// Long:   security secrets eac-core              # Single module
// Long:   security secrets eac-core r2r-cli      # Multiple modules
// Long:   security secrets eac-core --debug      # Debug logging
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package secrets

import (
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
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
	registry.Register(Secrets)
}

// Secrets command entry point
func Secrets() int {
	args := os.Args[3:] // Skip program name, "security", and "secrets"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printSecretsUsage()
		return 0
	}

	// Define valid flags

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		printSecretsUsage()
		return 1
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
				log.Errorf(" unknown flag: %s\n", arg)
				printSecretsUsage()
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
		log.Errorf(" failed to find repository root: %v\n", err)
		return 1
	}

	if debug {
		logger, err = logging.NewWithDebug("security", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("security", workspaceRoot)
	}
	if err != nil {
		log.Errorf(" failed to initialize logger: %v\n", err)
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

	logger.Info("Starting secrets scanner",
		zap.Strings("modules", monikers),
		zap.Bool("debug", debug))

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		logger.Error("Failed to load module contracts", zap.Error(err))
		log.Errorf(" failed to load module contracts: %v\n", err)
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
			log.Errorf(" module not found: %s\n", moniker)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Scanning module", zap.String("moniker", moniker), zap.String("root", module.Files.Root))
		log.Infof("🔐 Scanning %s...", moniker)

		// Run Trivy secrets scan
		findings, err := internal.RunTrivySecrets(module.Files.Root, trivyImage, logger)
		if err != nil {
			logger.Error("Secrets scan failed", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed: %v", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerSecrets, err.Error())
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
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerSecrets, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed to write evidence: %v", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Secrets scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		log.Infof("  ✅ Success: %s", outputPath)
		successCount++
	}

	// Print summary
	log.Info("")
	logger.Info("Secrets scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	log.Infof("Summary: %d succeeded, %d failed, %d total", successCount, failureCount, len(monikers))

	return exitCode
}

func printSecretsUsage() {
	log.Info("Detect secrets and credentials using Trivy")
	log.Info("")
	log.Info("Usage: security secrets [modules...] [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  [modules...]          One or more module monikers to scan")
	log.Info("                        If no modules specified, scans all modules")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --debug, -d           Enable debug logging")
	log.Info("")
	log.Info("Examples:")
	log.Info("  security secrets                       # All modules")
	log.Info("  security secrets eac-core              # Single module")
	log.Info("  security secrets eac-core r2r-cli      # Multiple modules")
	log.Info("  security secrets eac-core --debug      # Debug logging")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/security/<module>/secrets/<timestamp>.json")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses Trivy (Apache 2.0). See the NOTICE file in the")
	log.Info("  repository root for full attribution and licensing information.")
}
