// Command: security sbom
// Short: Generate Software Bill of Materials (SBOM)
// Long: Generate Software Bill of Materials (SBOM) using Trivy scanner.
// Long:
// Long: This command generates SBOM evidence files in CycloneDX format for audit compliance.
// Long: Evidence files are timestamped and SHA256-signed for integrity verification.
// Long:
// Long: Output: out/security/<module>/sbom/<timestamp>.json
// Long:
// Long: Example:
// Long:   security sbom                              # All modules
// Long:   security sbom src-core                     # Single module
// Long:   security sbom src-core src-cli             # Multiple modules
// Long:   security sbom src-core --format spdx-json  # SPDX format
// Long:   security sbom src-core --debug             # Debug logging
// Flag.format: type=string, default=cyclonedx, usage=SBOM format (cyclonedx, spdx, spdx-json, github, etc.)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package sbom

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/security/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
	"go.uber.org/zap"
)

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

	// Parse module monikers and flags
	var monikers []string
	format := "cyclonedx" // Default to CycloneDX
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --format requires a value\n")
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
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
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

	logger.Info("Starting SBOM scanner",
		zap.String("format", format),
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
		fmt.Printf("📦 Scanning %s...\n", moniker)

		// Run Trivy SBOM scan
		findings, err := internal.RunTrivySBOM(module.Files.Root, format, logger)
		if err != nil {
			logger.Error("SBOM scan failed", zap.String("moniker", moniker), zap.Error(err))
			fmt.Fprintf(os.Stderr, "  ❌ Failed: %v\n", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerSBOM, err.Error())
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
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerSBOM, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			fmt.Fprintf(os.Stderr, "  ❌ Failed to write evidence: %v\n", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("SBOM scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		fmt.Printf("  ✅ Success: %s\n", outputPath)
		successCount++
	}

	// Print summary
	fmt.Println()
	logger.Info("SBOM scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	fmt.Printf("Summary: %d succeeded, %d failed, %d total\n", successCount, failureCount, len(monikers))

	return exitCode
}

func printSBOMUsage() {
	fmt.Println("Generate Software Bill of Materials (SBOM)")
	fmt.Println()
	fmt.Println("Usage: security sbom [modules...] [flags]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  [modules...]          One or more module monikers to scan")
	fmt.Println("                        If no modules specified, scans all modules")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --format <format>     SBOM format (default: cyclonedx)")
	fmt.Println("                        Options: cyclonedx, spdx, spdx-json, github, json, sarif")
	fmt.Println("  --debug, -d           Enable debug logging")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  security sbom                              # All modules")
	fmt.Println("  security sbom src-core                     # Single module")
	fmt.Println("  security sbom src-core src-cli             # Multiple modules")
	fmt.Println("  security sbom src-core --format spdx-json  # SPDX format")
	fmt.Println("  security sbom src-core --debug             # Debug logging")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  out/security/<module>/sbom/<timestamp>.json")
	fmt.Println()
	fmt.Println("External tool:")
	fmt.Println("  This command uses Trivy (Apache 2.0). See the NOTICE file in the")
	fmt.Println("  repository root for full attribution and licensing information.")
}
