// Command: security sast
// Short: Static Application Security Testing using Semgrep
// Long: Perform Static Application Security Testing (SAST) using Semgrep.
// Long:
// Long: This command analyzes source code for security vulnerabilities, bugs, and code
// Long: quality issues using Semgrep's extensive rule library. Results are saved as
// Long: timestamped evidence files with SHA256 integrity verification for audit compliance.
// Long:
// Long: Output: out/security/<module>/sast/<timestamp>.json
// Long:
// Long: Example:
// Long:   security sast                          # All modules, auto-detect rules
// Long:   security sast eac-core                 # Single module
// Long:   security sast eac-core r2r-cli         # Multiple modules
// Long:   security sast --config p/security-audit  # Specific ruleset
// Long:   security sast --config p/owasp-top-ten   # OWASP Top 10
// Long:   security sast eac-core --debug         # Debug logging
// Flag.config: type=string, default=auto, usage=Semgrep config (auto, p/security-audit, p/owasp-top-ten, etc.)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package sast

import (
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/security/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"go.uber.org/zap"
)

var log = logging.C()

func init() {
	registry.Register(SAST)
}

// SAST command entry point
func SAST() int {
	args := os.Args[3:] // Skip program name, "security", and "sast"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printSASTUsage()
		return 0
	}

	// Parse module monikers and flags
	var monikers []string
	config := "auto" // Default to auto-detect
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--config":
			if i+1 >= len(args) {
				log.Errorf( "Error: --config requires a value\n")
				printSASTUsage()
				return 1
			}
			i++
			config = args[i]
		case "--debug", "-d":
			debug = true
		default:
			if strings.HasPrefix(arg, "--config=") {
				config = strings.TrimPrefix(arg, "--config=")
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf( "Error: unknown flag: %s\n", arg)
				printSASTUsage()
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
		log.Errorf( "Error: failed to find repository root: %v\n", err)
		return 1
	}

	if debug {
		logger, err = logging.NewWithDebug("security", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("security", workspaceRoot)
	}
	if err != nil {
		log.Errorf( "Error: failed to initialize logger: %v\n", err)
		return 1
	}
	defer logger.Sync()

	logger.Info("Starting SAST scanner",
		zap.String("config", config),
		zap.Strings("modules", monikers),
		zap.Bool("debug", debug))

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		logger.Error("Failed to load module contracts", zap.Error(err))
		log.Errorf( "Error: failed to load module contracts: %v\n", err)
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
			log.Errorf( "Error: module not found: %s\n", moniker)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Scanning module", zap.String("moniker", moniker), zap.String("root", module.Files.Root))
		log.Infof("🔬 Scanning %s...\n", moniker)

		// Run Semgrep SAST scan
		findings, err := internal.RunSemgrepSAST(workspaceRoot, module.Files.Root, config, logger)
		if err != nil {
			logger.Error("SAST scan failed", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf( "  ❌ Failed: %v\n", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerSAST, err.Error())
			if writeErr != nil {
				logger.Error("Failed to write error evidence", zap.Error(writeErr))
			} else {
				logger.Info("Error evidence written", zap.String("path", outputPath))
				log.Infof("  📄 Error evidence: %s\n", outputPath)
			}

			failureCount++
			exitCode = 1
			continue
		}

		// Write evidence file
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerSAST, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf( "  ❌ Failed to write evidence: %v\n", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("SAST scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		log.Infof("  ✅ Success: %s\n", outputPath)
		successCount++
	}

	// Print summary
	log.Info("")
	logger.Info("SAST scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	log.Infof("Summary: %d succeeded, %d failed, %d total\n", successCount, failureCount, len(monikers))

	return exitCode
}

func printSASTUsage() {
	log.Info("Static Application Security Testing using Semgrep")
	log.Info("")
	log.Info("Usage: security sast [modules...] [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  [modules...]          One or more module monikers to scan")
	log.Info("                        If no modules specified, scans all modules")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --config <ruleset>    Semgrep ruleset (default: auto)")
	log.Info("                        Options: auto, p/security-audit, p/owasp-top-ten,")
	log.Info("                                 p/cwe-top-25, p/ci, p/r2c-security-audit")
	log.Info("  --debug, -d           Enable debug logging")
	log.Info("")
	log.Info("Examples:")
	log.Info("  security sast                          # All modules, auto-detect")
	log.Info("  security sast eac-core                 # Single module")
	log.Info("  security sast eac-core r2r-cli         # Multiple modules")
	log.Info("  security sast --config p/security-audit  # Security audit ruleset")
	log.Info("  security sast --config p/owasp-top-ten   # OWASP Top 10")
	log.Info("  security sast eac-core --debug         # Debug logging")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/security/<module>/sast/<timestamp>.json")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses Semgrep (LGPL 2.1). See the NOTICE file in the")
	log.Info("  repository root for full attribution and licensing information.")
}
