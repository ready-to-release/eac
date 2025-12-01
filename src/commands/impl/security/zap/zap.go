// Command: security zap
// Short: Dynamic Application Security Testing using OWASP ZAP
// Long: Perform Dynamic Application Security Testing (DAST) using OWASP ZAP.
// Long:
// Long: This command performs black-box security testing of running web applications
// Long: using OWASP ZAP via Docker. It detects common vulnerabilities like XSS, SQL
// Long: injection, CSRF, and misconfigurations. Results are saved as timestamped
// Long: evidence files with SHA256 integrity verification for audit compliance.
// Long:
// Long: Note: Unlike other security commands, ZAP scans a running application URL
// Long: rather than module files. The module argument is used for evidence file
// Long: organization only.
// Long:
// Long: Output: out/security/<module>/zap/<timestamp>.json
// Long:
// Long: Example:
// Long:   security zap src-api --target http://localhost:8080              # Baseline scan
// Long:   security zap src-api --target http://localhost:8080 --scan-type full  # Full scan
// Long:   security zap src-api --target http://localhost:8080 --scan-type api   # API scan
// Long:   security zap src-api --target http://localhost:8080 --debug      # Debug logging
// Flag.target: type=string, default=, usage=Target URL to scan (required)
// Flag.scan-type: type=string, default=baseline, usage=Scan type (baseline, full, api)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: module
package zap

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
	registry.Register(ZAP)
}

// ZAP command entry point
func ZAP() int {
	args := os.Args[3:] // Skip program name, "security", and "zap"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printZAPUsage()
		return 0
	}

	// Parse module moniker and flags
	var moniker string
	targetURL := ""
	scanType := "baseline"
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--target":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --target requires a value\n")
				printZAPUsage()
				return 1
			}
			i++
			targetURL = args[i]
		case "--scan-type":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --scan-type requires a value\n")
				printZAPUsage()
				return 1
			}
			i++
			scanType = args[i]
			// Validate scan type
			if scanType != "baseline" && scanType != "full" && scanType != "api" {
				fmt.Fprintf(os.Stderr, "Error: invalid scan type: %s (must be baseline, full, or api)\n", scanType)
				printZAPUsage()
				return 1
			}
		case "--debug", "-d":
			debug = true
		default:
			if strings.HasPrefix(arg, "--target=") {
				targetURL = strings.TrimPrefix(arg, "--target=")
			} else if strings.HasPrefix(arg, "--scan-type=") {
				scanType = strings.TrimPrefix(arg, "--scan-type=")
				if scanType != "baseline" && scanType != "full" && scanType != "api" {
					fmt.Fprintf(os.Stderr, "Error: invalid scan type: %s (must be baseline, full, or api)\n", scanType)
					printZAPUsage()
					return 1
				}
			} else if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				printZAPUsage()
				return 1
			} else {
				if moniker == "" {
					moniker = arg
				} else {
					fmt.Fprintf(os.Stderr, "Error: only one module argument allowed for ZAP scans\n")
					printZAPUsage()
					return 1
				}
			}
		}
	}

	// Validate required arguments
	if moniker == "" {
		fmt.Fprintf(os.Stderr, "Error: module argument required\n")
		printZAPUsage()
		return 1
	}
	if targetURL == "" {
		fmt.Fprintf(os.Stderr, "Error: --target flag is required\n")
		printZAPUsage()
		return 1
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

	logger.Info("Starting ZAP DAST scanner",
		zap.String("module", moniker),
		zap.String("target", targetURL),
		zap.String("scanType", scanType),
		zap.Bool("debug", debug))

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		logger.Error("Failed to load module contracts", zap.Error(err))
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// Verify module exists
	_, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		logger.Error("Module not found", zap.String("moniker", moniker))
		fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
		return 1
	}

	logger.Info("Scanning target", zap.String("moniker", moniker), zap.String("target", targetURL))
	fmt.Printf("🕷️  Scanning %s at %s...\n", moniker, targetURL)

	// Run OWASP ZAP scan
	findings, err := internal.RunZAPScan(targetURL, scanType, workspaceRoot, logger)
	if err != nil {
		logger.Error("ZAP scan failed", zap.String("moniker", moniker), zap.Error(err))
		fmt.Fprintf(os.Stderr, "  ❌ Failed: %v\n", err)

		// Write error evidence
		outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerDAST, err.Error())
		if writeErr != nil {
			logger.Error("Failed to write error evidence", zap.Error(writeErr))
		} else {
			logger.Info("Error evidence written", zap.String("path", outputPath))
			fmt.Printf("  📄 Error evidence: %s\n", outputPath)
		}

		return 1
	}

	// Write evidence file
	outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerDAST, findings)
	if err != nil {
		logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
		fmt.Fprintf(os.Stderr, "  ❌ Failed to write evidence: %v\n", err)
		return 1
	}

	logger.Info("ZAP scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
	fmt.Printf("  ✅ Success: %s\n", outputPath)

	return 0
}

func printZAPUsage() {
	fmt.Println("Dynamic Application Security Testing using OWASP ZAP")
	fmt.Println()
	fmt.Println("Usage: security zap <module> --target <url> [flags]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <module>              Module moniker for evidence file organization")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  --target <url>        Target URL to scan (e.g., http://localhost:8080)")
	fmt.Println()
	fmt.Println("Optional Flags:")
	fmt.Println("  --scan-type <type>    Scan type (default: baseline)")
	fmt.Println("                        Options: baseline, full, api")
	fmt.Println("                        - baseline: Quick scan for common vulnerabilities")
	fmt.Println("                        - full:     Comprehensive scan (takes longer)")
	fmt.Println("                        - api:      API-specific scan")
	fmt.Println("  --debug, -d           Enable debug logging")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  security zap src-api --target http://localhost:8080              # Baseline scan")
	fmt.Println("  security zap src-api --target http://localhost:8080 --scan-type full  # Full scan")
	fmt.Println("  security zap src-api --target http://localhost:8080 --scan-type api   # API scan")
	fmt.Println("  security zap src-api --target http://localhost:8080 --debug      # Debug logging")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  out/security/<module>/zap/<timestamp>.json")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - Docker must be installed and running")
	fmt.Println("  - Target application must be accessible from Docker container")
	fmt.Println()
	fmt.Println("External tool:")
	fmt.Println("  This command uses OWASP ZAP (Apache 2.0) via Docker. See the NOTICE")
	fmt.Println("  file in the repository root for full attribution and licensing information.")
}
