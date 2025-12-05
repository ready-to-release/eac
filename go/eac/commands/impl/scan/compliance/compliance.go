// Command: scan compliance
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
// Long:   security compliance eac-core --compliance k8s-cis     # Specific module
// Long:   security compliance eac-core --debug                  # Debug logging
// Flag.compliance: type=string, default=k8s-cis, usage=Compliance standard (k8s-cis, docker-cis, k8s-nsa, etc.)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging
// HasSideEffects: false
// Args: modules
package compliance

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
				log.Error("--compliance requires a value")
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
				log.Errorf("unknown flag: %s", arg)
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
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	if debug {
		logger, err = logging.NewWithDebug("security", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("security", workspaceRoot)
	}
	if err != nil {
		log.Errorf("failed to initialize logger: %v", err)
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

	logger.Info("Starting compliance scanner",
		zap.String("standard", complianceStandard),
		zap.Strings("modules", monikers),
		zap.Bool("debug", debug))

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		logger.Error("Failed to load module contracts", zap.Error(err))
		log.Errorf("failed to load module contracts: %v", err)
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
			log.Errorf("module not found: %s", moniker)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Scanning module", zap.String("moniker", moniker), zap.String("root", module.Files.Root))
		log.Infof("✅ Scanning %s for %s compliance...", moniker, complianceStandard)

		// Run Trivy compliance scan
		findings, err := internal.RunTrivyCompliance(module.Files.Root, complianceStandard, trivyImage, logger)
		if err != nil {
			logger.Error("Compliance scan failed", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed: %v", err)

			// Write error evidence
			outputPath, writeErr := internal.WriteErrorEvidence(workspaceRoot, moniker, internal.ScannerCompliance, err.Error())
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
		outputPath, err := internal.WriteEvidence(workspaceRoot, moniker, internal.ScannerCompliance, findings)
		if err != nil {
			logger.Error("Failed to write evidence", zap.String("moniker", moniker), zap.Error(err))
			log.Errorf("  ❌ Failed to write evidence: %v", err)
			failureCount++
			exitCode = 1
			continue
		}

		logger.Info("Compliance scan completed", zap.String("moniker", moniker), zap.String("evidence", outputPath))
		log.Infof("  ✅ Success: %s", outputPath)
		successCount++
	}

	// Print summary
	log.Info("")
	logger.Info("Compliance scan summary",
		zap.Int("success", successCount),
		zap.Int("failed", failureCount),
		zap.Int("total", len(monikers)))

	log.Infof("Summary: %d succeeded, %d failed, %d total", successCount, failureCount, len(monikers))

	return exitCode
}

func printComplianceUsage() {
	log.Info("Check compliance with security standards using Trivy")
	log.Info("")
	log.Info("Usage: security compliance [modules...] [flags]")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  [modules...]          One or more module monikers to scan")
	log.Info("                        If no modules specified, scans all modules")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --compliance <std>    Compliance standard (default: k8s-cis)")
	log.Info("                        Options: k8s-cis, docker-cis, k8s-nsa, k8s-pss-baseline,")
	log.Info("                                 k8s-pss-restricted, awscli, pci-dss")
	log.Info("  --debug, -d           Enable debug logging")
	log.Info("")
	log.Info("Examples:")
	log.Info("  security compliance --compliance k8s-cis              # All modules, K8s CIS")
	log.Info("  security compliance --compliance docker-cis           # Docker CIS")
	log.Info("  security compliance --compliance k8s-nsa              # NSA/CISA K8s")
	log.Info("  security compliance eac-core --compliance k8s-cis     # Specific module")
	log.Info("  security compliance eac-core --debug                  # Debug logging")
	log.Info("")
	log.Info("Output:")
	log.Info("  out/security/<module>/compliance/<timestamp>.json")
	log.Info("")
	log.Info("External tool:")
	log.Info("  This command uses Trivy (Apache 2.0). See the NOTICE file in the")
	log.Info("  repository root for full attribution and licensing information.")
}
