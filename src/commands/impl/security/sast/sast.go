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
// Long:   security sast src-core                 # Single module
// Long:   security sast src-core src-cli         # Multiple modules
// Long:   security sast --config p/security-audit  # Specific ruleset
// Long:   security sast --config p/owasp-top-ten   # OWASP Top 10
// Long:   security sast src-core --debug         # Debug logging
// Flag.config: type=string, default=auto, usage=Semgrep config (auto, p/security-audit, p/owasp-top-ten, etc.)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package sast

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/security/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
	"go.uber.org/zap"
)

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
				fmt.Fprintf(os.Stderr, "Error: --config requires a value\n")
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
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
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
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	if debug {
		logger, err = logging.NewWithDebug("security", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("security", workspaceRoot)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize logger: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// If no monikers provided, default to all modules
	if len(monikers) == 0 {
		fmt.Println("ℹ️  No modules specified, scanning all modules...")
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
			fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Scanning module", zap.String("moniker", moniker), zap.String("root", module.Files.Root))
		fmt.Printf("🔬 Scanning %s...\n", moniker)

		// Run Semgrep SAST scan
		findings, err := internal.RunSemgrepSAST(module.Files.Root, config, logger)
		if err != nil {
			logger.Error("SAST scan failed", zap.String("moniker", moniker), zap.Error(err))
			fmt.Fprintf(os.Stderr, "  ❌ Failed: %v\n", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerSAST, err.Error())
			if writeErr != nil {
				logger.Error("Failed to write error evidence", zap.Error(writeErr))
			} else {
				logger.Info("Error evidence written", zap.String("path", outputPath))
				fmt.Printf("  📄 Error evidence: %s\n", outputPath)
			}

			failureCount++
			exitCode = 1
			continue
		}

		// Write evidence file
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerSAST, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			fmt.Fprintf(os.Stderr, "  ❌ Failed to write evidence: %v\n", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("SAST scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		fmt.Printf("  ✅ Success: %s\n", outputPath)
		successCount++
	}

	// Print summary
	fmt.Println()
	logger.Info("SAST scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	fmt.Printf("Summary: %d succeeded, %d failed, %d total\n", successCount, failureCount, len(monikers))

	return exitCode
}

func printSASTUsage() {
	fmt.Println("Static Application Security Testing using Semgrep")
	fmt.Println()
	fmt.Println("Usage: security sast [modules...] [flags]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  [modules...]          One or more module monikers to scan")
	fmt.Println("                        If no modules specified, scans all modules")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --config <ruleset>    Semgrep ruleset (default: auto)")
	fmt.Println("                        Options: auto, p/security-audit, p/owasp-top-ten,")
	fmt.Println("                                 p/cwe-top-25, p/ci, p/r2c-security-audit")
	fmt.Println("  --debug, -d           Enable debug logging")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  security sast                          # All modules, auto-detect")
	fmt.Println("  security sast src-core                 # Single module")
	fmt.Println("  security sast src-core src-cli         # Multiple modules")
	fmt.Println("  security sast --config p/security-audit  # Security audit ruleset")
	fmt.Println("  security sast --config p/owasp-top-ten   # OWASP Top 10")
	fmt.Println("  security sast src-core --debug         # Debug logging")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  out/security/<module>/sast/<timestamp>.json")
	fmt.Println()
	fmt.Println("External tool:")
	fmt.Println("  This command uses Semgrep (LGPL 2.1). See the NOTICE file in the")
	fmt.Println("  repository root for full attribution and licensing information.")
}
