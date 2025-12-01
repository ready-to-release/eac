// Command: security compliance
// Short: Check compliance with security standards using Trivy
// Long: Verify compliance with security benchmarks and standards using Trivy.
// Long:
// Long: This command checks infrastructure and configurations against industry
// Long: standards like CIS Benchmarks, NIST, and PCI DSS. Results are saved as
// Long: timestamped evidence files with SHA256 integrity verification.
// Long:
// Long: Output: out/security/<module>/compliance/<timestamp>.json
// Long:
// Long: Example:
// Long:   security compliance --compliance k8s-cis              # Kubernetes CIS
// Long:   security compliance --compliance docker-cis           # Docker CIS
// Long:   security compliance --compliance k8s-nsa              # NSA/CISA K8s
// Long:   security compliance src-core --compliance k8s-cis     # Specific module
// Long:   security compliance src-core --debug                  # Debug logging
// Flag.compliance: type=string, default=k8s-cis, usage=Compliance standard (k8s-cis, docker-cis, k8s-nsa, etc.)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package compliance

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
	registry.Register(Compliance)
}

// Compliance command entry point
func Compliance() int {
	args := os.Args[3:] // Skip program name, "security", and "compliance"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printComplianceUsage()
		return 0
	}

	// Parse module monikers and flags
	var monikers []string
	complianceStandard := "k8s-cis" // Default to Kubernetes CIS
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--compliance":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --compliance requires a value\n")
				printComplianceUsage()
				return 1
			}
			i++
			complianceStandard = args[i]
		case "--debug", "-d":
			debug = true
		default:
			if strings.HasPrefix(arg, "--compliance=") {
				complianceStandard = strings.TrimPrefix(arg, "--compliance=")
			} else if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				printComplianceUsage()
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

	logger.Info("Starting compliance scanner",
		zap.String("standard", complianceStandard),
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
		fmt.Printf("✅ Scanning %s for %s compliance...\n", moniker, complianceStandard)

		// Run Trivy compliance scan
		findings, err := internal.RunTrivyCompliance(module.Files.Root, complianceStandard, logger)
		if err != nil {
			logger.Error("Compliance scan failed", zap.String("moniker", moniker), zap.Error(err))
			fmt.Fprintf(os.Stderr, "  ❌ Failed: %v\n", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerCompliance, err.Error())
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
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerCompliance, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			fmt.Fprintf(os.Stderr, "  ❌ Failed to write evidence: %v\n", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Compliance scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		fmt.Printf("  ✅ Success: %s\n", outputPath)
		successCount++
	}

	// Print summary
	fmt.Println()
	logger.Info("Compliance scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	fmt.Printf("Summary: %d succeeded, %d failed, %d total\n", successCount, failureCount, len(monikers))

	return exitCode
}

func printComplianceUsage() {
	fmt.Println("Check compliance with security standards using Trivy")
	fmt.Println()
	fmt.Println("Usage: security compliance [modules...] [flags]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  [modules...]          One or more module monikers to scan")
	fmt.Println("                        If no modules specified, scans all modules")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --compliance <std>    Compliance standard (default: k8s-cis)")
	fmt.Println("                        Options: k8s-cis, docker-cis, k8s-nsa, k8s-pss-baseline,")
	fmt.Println("                                 k8s-pss-restricted, awscli, pci-dss")
	fmt.Println("  --debug, -d           Enable debug logging")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  security compliance --compliance k8s-cis              # All modules, K8s CIS")
	fmt.Println("  security compliance --compliance docker-cis           # Docker CIS")
	fmt.Println("  security compliance --compliance k8s-nsa              # NSA/CISA K8s")
	fmt.Println("  security compliance src-core --compliance k8s-cis     # Specific module")
	fmt.Println("  security compliance src-core --debug                  # Debug logging")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  out/security/<module>/compliance/<timestamp>.json")
	fmt.Println()
	fmt.Println("External tool:")
	fmt.Println("  This command uses Trivy (Apache 2.0). See the NOTICE file in the")
	fmt.Println("  repository root for full attribution and licensing information.")
}
