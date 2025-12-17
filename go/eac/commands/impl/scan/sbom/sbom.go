// Command: scan sbom
// Short: Generate Software Bill of Materials (SBOM)
// Long: Generate Software Bill of Materials (SBOM) using Trivy scanner.
// Long:
// Long: This command generates SBOM evidence files in CycloneDX format for audit compliance.
// Long: Evidence files are timestamped and SHA256-signed for integrity verification.
// Long:
// Long: Expected Output:
// Long:   Timestamped JSON evidence files are written to out/security/<module>/sbom/<timestamp>.json
// Long:   Files are generated in CycloneDX format by default (customizable via --format flag).
// Long:
// Long: Example:
// Long:   security sbom                              # All modules
// Long:   security sbom eac-core                     # Single module
// Long:   security sbom eac-core r2r-cli             # Multiple modules
// Long:   security sbom eac-core --format spdx-json  # SPDX format
// Long:   security sbom eac-core --debug             # Debug logging
// Flag.format: type=string, default=cyclonedx, usage=SBOM format (cyclonedx, spdx, spdx-json, github, etc.)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package sbom

import (
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"go.uber.org/zap"
)

var log = logging.C()

func init() {
	registry.Register(SBOM)
}

// SBOM command entry point
func SBOM() int {
	args := os.Args[3:] // Skip program name, "security", and "sbom"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printSBOMUsage()
		return 0
	}

	// Define valid flags

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		printSBOMUsage()
		return 1
	}

	// Parse module monikers and flags
	var monikers []string
	format := "cyclonedx" // Default to CycloneDX
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format":
			if i+1 >= len(args) {
				log.Errorf(" --format requires a value\n")
				printSBOMUsage()
				return 1
			}
			i++
			format = args[i]
		case "--debug", "-d":
			debug = true
		default:
			if strings.HasPrefix(arg, "--format=") {
				format = strings.TrimPrefix(arg, "--format=")
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf(" unknown flag: %s\n", arg)
				printSBOMUsage()
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

	logger.Info("Starting SBOM scanner",
		zap.String("format", format),
		zap.Strings("modules", monikers),
		zap.Bool("debug", debug))

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
		log.Infof("📦 Scanning %s...", moniker)

		// Run Trivy SBOM scan
		findings, err := internal.RunTrivySBOM(workspaceRoot, module.Files.Root, format, trivyImage, logger)
		if err != nil {
			logger.Error("SBOM scan failed", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed: %v", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerSBOM, err.Error())
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
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerSBOM, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed to write evidence: %v", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("SBOM scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		log.Infof("  ✅ Success: %s", outputPath)
		successCount++
	}

	// Print summary
	log.Info("")
	logger.Info("SBOM scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	log.Infof("Summary: %d succeeded, %d failed, %d total", successCount, failureCount, len(monikers))

	return exitCode
}

func printSBOMUsage() {
	log.Info("Generate Software Bill of Materials (SBOM)")
	log.Info("")
	log.Info("Usage: security sbom [modules...] [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  [modules...]          One or more module monikers to scan")
	log.Info("                        If no modules specified, scans all modules")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --format <format>     SBOM format (default: cyclonedx)")
	log.Info("                        Options: cyclonedx, spdx, spdx-json, github, json, sarif")
	log.Info("  --debug, -d           Enable debug logging")
	log.Info("")
	log.Info("Examples:")
	log.Info("  security sbom                              # All modules")
	log.Info("  security sbom eac-core                     # Single module")
	log.Info("  security sbom eac-core r2r-cli             # Multiple modules")
	log.Info("  security sbom eac-core --format spdx-json  # SPDX format")
	log.Info("  security sbom eac-core --debug             # Debug logging")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/security/<module>/sbom/<timestamp>.json")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses Trivy (Apache 2.0). See the NOTICE file in the")
	log.Info("  repository root for full attribution and licensing information.")
}
