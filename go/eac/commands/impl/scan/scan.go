// Command: scan
// Short: Security scanning and evidence collection for audit compliance
// Long: Scan modules for security vulnerabilities using industry-standard tools.
// Long:
// Long: If no modules specified, scans all modules in the repository.
// Long: If no --scanner specified, uses default scanners for each module type
// Long: (configured in .r2r/eac/security-tools.yml).
// Long:
// Long: Supported scanners:
// Long:   - sbom: Software Bill of Materials (Trivy, CycloneDX format)
// Long:   - vuln: Vulnerability scanning (Trivy)
// Long:   - secrets: Secrets detection (Trivy)
// Long:   - iac: Infrastructure as Code scanning (Trivy)
// Long:   - compliance: CIS compliance checking (Trivy)
// Long:   - sast: Static Application Security Testing (Semgrep)
// Long:   - zap: Dynamic Application Security Testing (OWASP ZAP)
// Long:
// Long: Example:
// Long:   scan                                  # All modules, default scanners
// Long:   scan eac-core                         # Single module, default scanners
// Long:   scan --scanner sbom                   # All modules, SBOM only
// Long:   scan eac-core --scanner sbom,vuln     # Single module, specific scanners
// Long:
// Long: Evidence output: out/scan/<module>/<scanner>/
// Flag.scanner: type=string, default=, usage=Scanner types to run (comma-separated: sbom,vuln,secrets,iac,compliance,sast,zap)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging to out/logs/security
// Flag.tui: type=bool, default=auto, usage=Enable TUI console (default: auto-detect)
// Flag.no-tui: type=bool, default=false, usage=Disable TUI console
// Flag.tui-height: type=int, default=8, usage=Set TUI console height (3-20)
// Flag.sequential: type=bool, default=false, usage=Run scans sequentially instead of in parallel
// Flag.turbo: type=bool, default=false, usage=Enable turbo mode for faster scanning (increases parallelism)
// Flag.skip-cache: type=bool, default=false, usage=Skip incremental cache, force full scan
// Flag.skip-deps: type=bool, default=false, usage=Skip system dependency verification (trivy, semgrep, etc.)
// HasSideEffects: false
// Args: modules
package scan

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"

	// Import scanners package to trigger init() registration of native scanners
	_ "github.com/ready-to-release/eac/go/eac/commands/impl/scan/scanners"
)

var log = logging.C()

// ValidScannerTypes lists all valid scanner type strings.
var ValidScannerTypes = []string{"sbom", "vuln", "secrets", "iac", "compliance", "sast", "zap"}

func init() {
	registry.Register(Scan)
}

// Scan command entry point.
func Scan() int {
	args := os.Args[2:] // Skip program name and "scan"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printScanUsage()
		return 0
	}

	// Detect execution environment
	env := environment.Detect()

	// Parse scan-specific flags FIRST (before shared flags)
	// This ensures --scanner <value> is consumed before shared parsing
	// separates the value into positional args
	scanFlags, remainingAfterScan, err := ParseScanSpecificFlags(args)
	if err != nil {
		log.Errorf("Error: %v", err)
		printScanUsage()
		return 1
	}

	// Parse shared flags from remaining args
	shared, err := flags.ParseSharedFlagsWithEnv(flags.ScanConfig(), remainingAfterScan, env)
	if err != nil {
		log.Errorf("Error: %v", err)
		printScanUsage()
		return 1
	}

	// Check for unknown flags
	for _, arg := range shared.Remaining {
		if strings.HasPrefix(arg, "--") {
			log.Errorf("Error: unknown flag: %s", arg)
			return 1
		}
	}

	// Determine sequential mode: --roof 1 means sequential
	sequential := shared.MaxConcurrency == 1

	// Validate --turbo and sequential mutual exclusivity
	if shared.Turbo && sequential {
		log.Errorf("Error: --turbo and --roof 1 (sequential) cannot be used together")
		return 1
	}

	// Create command config
	cmdCfg := &cmdframework.CommandConfig{
		Type:           cmdframework.CommandTypeScan,
		ActionVerb:     "Scanning",
		OutputDir:      "out/scan",
		LogFileName:    "scan.log",
		Monikers:       shared.Monikers,
		Layered:        false, // Scan uses parallel execution
		Sequential:     sequential,
		Turbo:          shared.Turbo,
		MaxConcurrency: shared.MaxConcurrency,
		ForceRebuild:   shared.SkipCache,
		SkipDeps:       shared.SkipDeps,
		UseTUI:         shared.UseTUI,
		TUIHeight:      shared.TUIHeight,
		TUIASCIIMode:   shared.TUIASCIIMode,
		DebugMode:      shared.Debug,
		ShowTimings:    shared.ShowTimings,
		DryRun:         shared.DryRun,
	}

	// Create multi-scanner config
	multiCfg := &MultiScanConfig{
		Scanners:           scanFlags.Scanners,
		SBOMFormat:         scanFlags.SBOMFormat,
		VulnSeverities:     scanFlags.VulnSeverities,
		SemgrepConfig:      scanFlags.SemgrepConfig,
		ComplianceStandard: scanFlags.ComplianceStandard,
	}

	return RunMultiScan(cmdCfg, multiCfg)
}

// parseScannerList parses a comma-separated list of scanner types.
func parseScannerList(input string) ([]internal.ScannerType, error) {
	var scanners []internal.ScannerType
	for _, s := range strings.Split(input, ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			continue
		}
		scannerType, valid := internal.ParseScannerType(s)
		if !valid {
			return nil, fmt.Errorf("invalid scanner type: %s (valid: %s)", s, strings.Join(ValidScannerTypes, ", "))
		}
		scanners = append(scanners, scannerType)
	}
	if len(scanners) == 0 {
		return nil, fmt.Errorf("no valid scanner types specified")
	}
	return scanners, nil
}

// parseSeverityList parses a comma-separated list of severity levels.
func parseSeverityList(input string) ([]internal.Severity, error) {
	var severities []internal.Severity
	for _, s := range strings.Split(input, ",") {
		s = strings.TrimSpace(strings.ToUpper(s))
		if s == "" {
			continue
		}
		severity, valid := internal.ParseSeverity(s)
		if !valid {
			return nil, fmt.Errorf("invalid severity: %s (valid: LOW, MEDIUM, HIGH, CRITICAL)", s)
		}
		severities = append(severities, severity)
	}
	return severities, nil
}

func printScanUsage() {
	log.Info("Security scanning and evidence collection for audit compliance")
	log.Info("")
	log.Info("Usage: scan [modules...] [flags]")
	log.Info("")
	log.Info("If no modules specified, scans all modules in the repository.")
	log.Info("If no --scanner specified, uses default scanners for each module type.")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --scanner <types>         Scanner types to run (comma-separated)")
	log.Info("                            Options: sbom, vuln, secrets, iac, compliance, sast, zap")
	log.Info("  --turbo                   Enable turbo mode for faster scanning (increases parallelism)")
	log.Info("  --skip-cache              Skip incremental cache, force full scan")
	log.Info("  --skip-deps               Skip system dependency verification (trivy, semgrep)")
	log.Info("  --sequential              Run scans sequentially instead of in parallel")
	log.Info("  --debug, -d               Enable debug logging")
	log.Info("  --tui                     Enable TUI console (default for local)")
	log.Info("  --no-tui                  Disable TUI console")
	log.Info("  --tui-height N            Set TUI console height (3-20)")
	log.Info("")
	log.Info("Scanner-specific flags:")
	log.Info("  --format <fmt>            SBOM format (cyclonedx, spdx, spdx-json, github)")
	log.Info("  --severity <levels>       Vulnerability severity filter (LOW,MEDIUM,HIGH,CRITICAL)")
	log.Info("  --config <ruleset>        SAST config (auto, p/security-audit, p/owasp-top-ten)")
	log.Info("  --compliance <std>        Compliance standard (k8s-cis, docker-cis, k8s-nsa)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  scan                                  # All modules, default scanners")
	log.Info("  scan eac-core                         # Single module, default scanners")
	log.Info("  scan eac-core r2r-cli                 # Multiple modules")
	log.Info("  scan --scanner sbom                   # All modules, single scanner")
	log.Info("  scan eac-core --scanner sbom,vuln     # Single module, multiple scanners")
	log.Info("  scan --severity HIGH                  # Vuln with severity filter")
	log.Info("  scan --scanner sast --config p/owasp-top-ten")
	log.Info("  scan eac-core --debug                 # With debug logging")
	log.Info("  scan --sequential                     # Run all scans sequentially")
	log.Info("")
	log.Info("Evidence output:")
	log.Info("  out/scan/<module>/<scanner>/")
	log.Info("")
	log.Info("Configuration:")
	log.Info("  Default scanners per module type: .r2r/eac/security-tools.yml")
	log.Info("  Modules to skip (skip_modules):   .r2r/eac/security-tools.yml")
	log.Info("  Docker image versions:            .r2r/eac/security-tools.yml")
	log.Info("")
	log.Info("External tools:")
	log.Info("  This command uses third-party security tools. See the NOTICE file")
	log.Info("  in the repository root for full attribution and licensing information.")
}
