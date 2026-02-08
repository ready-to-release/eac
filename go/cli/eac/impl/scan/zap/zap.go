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
// Long:   Evidence files are written to out/scan/<module>/zap/
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
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/adapters/docker"
	"github.com/ready-to-release/eac/go/core/evidence"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(ZAP)
}

// ZAP command entry point
// Note: ZAP is special - it scans a running application URL, not module files,
// and only accepts a single module for evidence file organization.
func ZAP() int {
	args := os.Args[3:] // Skip program name, "security", and "zap"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printZAPUsage()
		return 0
	}

	// Define valid flags

	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
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
				log.Errorf("Error: --target requires a value\n")
				printZAPUsage()
				return 1
			}
			i++
			targetURL = args[i]
		case "--scan-type":
			if i+1 >= len(args) {
				log.Errorf("Error: --scan-type requires a value\n")
				printZAPUsage()
				return 1
			}
			i++
			scanType = args[i]
			// Validate scan type
			if scanType != "baseline" && scanType != "full" && scanType != "api" {
				log.Errorf("Error: invalid scan type: %s (must be baseline, full, or api)\n", scanType)
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
					log.Errorf("Error: invalid scan type: %s (must be baseline, full, or api)\n", scanType)
					printZAPUsage()
					return 1
				}
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf("Error: unknown flag: %s\n", arg)
				printZAPUsage()
				return 1
			} else {
				if moniker == "" {
					moniker = arg
				} else {
					log.Errorf("Error: only one module argument allowed for ZAP scans\n")
					printZAPUsage()
					return 1
				}
			}
		}
	}

	// Validate required arguments
	if moniker == "" {
		log.Errorf("Error: module argument required\n")
		printZAPUsage()
		return 1
	}
	if targetURL == "" {
		log.Errorf("Error: --target flag is required\n")
		printZAPUsage()
		return 1
	}

	// Initialize
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to find repository root: %v\n", err)
		return 1
	}

	if err := logging.ConfigureLoggingSimple(workspaceRoot, "commands", nil, debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	cfg, err := config.Load(config.DefaultLoadOptions())
	if err != nil {
		log.Debugf("Failed to load configuration: error=%v", err)
		log.Errorf(" failed to load configuration: %v\n", err)
		return 1
	}

	if cfg.Security == nil {
		log.Errorf("Error: security config not loaded\n")
		return 1
	}
	scanner, ok := cfg.Security.GetScanner("zap")
	if !ok {
		log.Errorf("Error: ZAP scanner not found in security config\n")
		return 1
	}
	zapImage := scanner.FullImage()
	log.Debugf("Using ZAP image: image=%s", zapImage)

	log.Debugf("Starting ZAP DAST scanner: module=%s, target=%s, scanType=%s, debug=%v", moniker, targetURL, scanType, debug)

	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Debugf("Failed to load module contracts: error=%v", err)
		log.Errorf("Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// Verify module exists
	_, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		log.Debugf("Module not found: moniker=%s", moniker)
		log.Errorf("Error: module not found: %s\n", moniker)
		return 1
	}

	log.Debugf("Scanning target: moniker=%s, target=%s", moniker, targetURL)
	log.Infof("🕷️  Scanning %s at %s...\n", moniker, targetURL)

	// Run OWASP ZAP scan
	findings, err := docker.RunZAPScan(targetURL, scanType, workspaceRoot, zapImage)
	if err != nil {
		log.Debugf("ZAP scan failed: moniker=%s, error=%v", moniker, err)
		log.Errorf("  ❌ Failed: %v\n", err)

		// Write error evidence
		outputPath, writeErr := evidence.WriteErrorEvidence(workspaceRoot, moniker, evidence.ScannerDAST, err.Error())
		if writeErr != nil {
			log.Debugf("Failed to write error evidence: error=%v", writeErr)
		} else {
			log.Debugf("Error evidence written: path=%s", outputPath)
			log.Infof("  📄 Error evidence: %s\n", outputPath)
		}

		return 1
	}

	// Write evidence file
	outputPath, err := evidence.WriteEvidence(workspaceRoot, moniker, evidence.ScannerDAST, findings)
	if err != nil {
		log.Debugf("Failed to write evidence: moniker=%s, error=%v", moniker, err)
		log.Errorf("  ❌ Failed to write evidence: %v\n", err)
		return 1
	}

	log.Debugf("ZAP scan completed: moniker=%s, evidence=%s", moniker, outputPath)
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
	log.Info("  out/scan/<module>/zap/")
	log.Info("")
	log.Info("Requirements:")
	log.Info("  - Docker must be installed and running")
	log.Info("  - Target application must be accessible from Docker container")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses OWASP ZAP (Apache 2.0) via Docker. See the NOTICE")
	log.Info("  file in the repository root for full attribution and licensing information.")
}
