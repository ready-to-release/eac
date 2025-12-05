// Command: scan vuln
// Short: Scan for vulnerabilities using Trivy
// Long: Scan for vulnerabilities in container images and filesystems using Trivy.
// Long:
// Long: This command scans for known CVEs and generates timestamped evidence files
// Long: with SHA256 integrity verification for audit compliance.
// Long:
// Long: Output: out/security/<module>/vuln/<timestamp>.json
// Long:
// Long: Example:
// Long:   security vuln                              # All modules
// Long:   security vuln eac-core                     # Single module
// Long:   security vuln eac-core r2r-cli             # Multiple modules
// Long:   security vuln eac-core --severity HIGH     # High severity only
// Long:   security vuln eac-core --severity CRITICAL,HIGH  # Multiple severities
// Long:   security vuln eac-core --debug             # Debug logging
// Flag.severity: type=string, default=, usage=Filter by severity (LOW,MEDIUM,HIGH,CRITICAL)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package vuln

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
	registry.Register(Vuln)
}

// Vuln command entry point
func Vuln() int {
	args := os.Args[3:] // Skip program name, "security", and "vuln"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printVulnUsage()
		return 0
	}

	// Parse module monikers and flags
	var monikers []string
	var severityFilter []internal.Severity
	debug := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--severity":
			if i+1 >= len(args) {
				log.Errorf( "Error: --severity requires a value\n")
				printVulnUsage()
				return 1
			}
			i++
			severityStr := args[i]
			// Parse comma-separated severity values
			for _, sev := range strings.Split(severityStr, ",") {
				sev = strings.TrimSpace(strings.ToUpper(sev))
				severity, valid := internal.ParseSeverity(sev)
				if !valid {
					log.Errorf( "Error: invalid severity: %s\n", sev)
					printVulnUsage()
					return 1
				}
				severityFilter = append(severityFilter, severity)
			}
		case "--debug", "-d":
			debug = true
		default:
			if strings.HasPrefix(arg, "--severity=") {
				severityStr := strings.TrimPrefix(arg, "--severity=")
				for _, sev := range strings.Split(severityStr, ",") {
					sev = strings.TrimSpace(strings.ToUpper(sev))
					severity, valid := internal.ParseSeverity(sev)
					if !valid {
						log.Errorf( "Error: invalid severity: %s\n", sev)
						printVulnUsage()
						return 1
					}
					severityFilter = append(severityFilter, severity)
				}
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf( "Error: unknown flag: %s\n", arg)
				printVulnUsage()
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

	logger.Info("Starting vulnerability scanner",
		zap.Any("severityFilter", severityFilter),
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
		log.Infof("🔍 Scanning %s...\n", moniker)

		// Run Trivy vulnerability scan
		findings, err := internal.RunTrivyVuln(module.Files.Root, severityFilter, trivyImage, logger)
		if err != nil {
			logger.Error("Vulnerability scan failed", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf( "  ❌ Failed: %v\n", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerVuln, err.Error())
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
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerVuln, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf( "  ❌ Failed to write evidence: %v\n", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Vulnerability scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		log.Infof("  ✅ Success: %s\n", outputPath)
		successCount++
	}

	// Print summary
	log.Info("")
	logger.Info("Vulnerability scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	log.Infof("Summary: %d succeeded, %d failed, %d total\n", successCount, failureCount, len(monikers))

	return exitCode
}

func printVulnUsage() {
	log.Info("Scan for vulnerabilities using Trivy")
	log.Info("")
	log.Info("Usage: security vuln [modules...] [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  [modules...]          One or more module monikers to scan")
	log.Info("                        If no modules specified, scans all modules")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --severity <levels>   Filter by severity (comma-separated)")
	log.Info("                        Options: LOW, MEDIUM, HIGH, CRITICAL")
	log.Info("                        Example: --severity HIGH,CRITICAL")
	log.Info("  --debug, -d           Enable debug logging")
	log.Info("")
	log.Info("Examples:")
	log.Info("  security vuln                              # All modules")
	log.Info("  security vuln eac-core                     # Single module")
	log.Info("  security vuln eac-core r2r-cli             # Multiple modules")
	log.Info("  security vuln eac-core --severity HIGH     # High severity only")
	log.Info("  security vuln eac-core --severity CRITICAL,HIGH  # Multiple severities")
	log.Info("  security vuln eac-core --debug             # Debug logging")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/security/<module>/vuln/<timestamp>.json")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses Trivy (Apache 2.0). See the NOTICE file in the")
	log.Info("  repository root for full attribution and licensing information.")
}
