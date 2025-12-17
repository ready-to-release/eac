// Command: scan zap
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
// Long: Expected Output:
// Long:   Timestamped JSON evidence files are written to out/security/<module>/zap/<timestamp>.json
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

	// Define valid flags
	commandFlags := []flags.FlagDefinition{
		{Name: "--target", HasValue: true, ValueType: "string"},
		{Name: "--scan-type", HasValue: true, ValueType: "string"},
		{Name: "--debug", Shorthand: "-d", HasValue: false},
	}

	// Validate flags
	if err := flags.ValidateFlags(args, commandFlags); err != nil {
		log.Errorf("%v", err)
		printZAPUsage()
		return 1
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
				log.Errorf( "Error: --target requires a value\n")
				printZAPUsage()
				return 1
			}
			i++
			targetURL = args[i]
		case "--scan-type":
			if i+1 >= len(args) {
				log.Errorf( "Error: --scan-type requires a value\n")
				printZAPUsage()
				return 1
			}
			i++
			scanType = args[i]
			// Validate scan type
			if scanType != "baseline" && scanType != "full" && scanType != "api" {
				log.Errorf( "Error: invalid scan type: %s (must be baseline, full, or api)\n", scanType)
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
					log.Errorf( "Error: invalid scan type: %s (must be baseline, full, or api)\n", scanType)
					printZAPUsage()
					return 1
				}
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf( "Error: unknown flag: %s\n", arg)
				printZAPUsage()
				return 1
			} else {
				if moniker == "" {
					moniker = arg
				} else {
					log.Errorf( "Error: only one module argument allowed for ZAP scans\n")
					printZAPUsage()
					return 1
				}
			}
		}
	}

	// Validate required arguments
	if moniker == "" {
		log.Errorf( "Error: module argument required\n")
		printZAPUsage()
		return 1
	}
	if targetURL == "" {
		log.Errorf( "Error: --target flag is required\n")
		printZAPUsage()
		return 1
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

	// Load configuration
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		logger.Error("Failed to load configuration", zap.Error(err))
		log.Errorf(" failed to load configuration: %v\n", err)
		return 1
	}

	// Get Docker image from config
	zapImage := cfg.SecurityTools.DockerImages.ZAP.FullImage()
	logger.Debug("Using ZAP image", zap.String("image", zapImage))

	logger.Info("Starting ZAP DAST scanner",
		zap.String("module", moniker),
		zap.String("target", targetURL),
		zap.String("scanType", scanType),
		zap.Bool("debug", debug))

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		logger.Error("Failed to load module contracts", zap.Error(err))
		log.Errorf( "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// Verify module exists
	_, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		logger.Error("Module not found", zap.String("moniker", moniker))
		log.Errorf( "Error: module not found: %s\n", moniker)
		return 1
	}

	logger.Info("Scanning target", zap.String("moniker", moniker), zap.String("target", targetURL))
	log.Infof("🕷️  Scanning %s at %s...\n", moniker, targetURL)

	// Run OWASP ZAP scan
	findings, err := internal.RunZAPScan(targetURL, scanType, workspaceRoot, zapImage, logger)
	if err != nil {
		logger.Error("ZAP scan failed", zap.String("moniker", moniker), zap.Error(err))
		log.Errorf( "  ❌ Failed: %v\n", err)

		// Write error evidence
		outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerDAST, err.Error())
		if writeErr != nil {
			logger.Error("Failed to write error evidence", zap.Error(writeErr))
		} else {
			logger.Info("Error evidence written", zap.String("path", outputPath))
			log.Infof("  📄 Error evidence: %s\n", outputPath)
		}

		return 1
	}

	// Write evidence file
	outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerDAST, findings)
	if err != nil {
		logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
		log.Errorf( "  ❌ Failed to write evidence: %v\n", err)
		return 1
	}

	logger.Info("ZAP scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
	log.Infof("  ✅ Success: %s\n", outputPath)

	return 0
}

func printZAPUsage() {
	log.Info("Dynamic Application Security Testing using OWASP ZAP")
	log.Info("")
	log.Info("Usage: security zap <module> --target <url> [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  <module>              Module moniker for evidence file organization")
	log.Info("")
	log.Info("Required Flags:")
	log.Info("  --target <url>        Target URL to scan (e.g., http://localhost:8080)")
	log.Info("")
	log.Info("Optional Flags:")
	log.Info("  --scan-type <type>    Scan type (default: baseline)")
	log.Info("                        Options: baseline, full, api")
	log.Info("                        - baseline: Quick scan for common vulnerabilities")
	log.Info("                        - full:     Comprehensive scan (takes longer)")
	log.Info("                        - api:      API-specific scan")
	log.Info("  --debug, -d           Enable debug logging")
	log.Info("")
	log.Info("Examples:")
	log.Info("  security zap src-api --target http://localhost:8080              # Baseline scan")
	log.Info("  security zap src-api --target http://localhost:8080 --scan-type full  # Full scan")
	log.Info("  security zap src-api --target http://localhost:8080 --scan-type api   # API scan")
	log.Info("  security zap src-api --target http://localhost:8080 --debug      # Debug logging")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/security/<module>/zap/<timestamp>.json")
	log.Info("")
	log.Info("Requirements:")
	log.Info("  - Docker must be installed and running")
	log.Info("  - Target application must be accessible from Docker container")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses OWASP ZAP (Apache 2.0) via Docker. See the NOTICE")
	log.Info("  file in the repository root for full attribution and licensing information.")
}
