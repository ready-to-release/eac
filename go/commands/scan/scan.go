package scan

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/evidence"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/environment"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

var log = logging.C()

// ValidScannerTypes lists all valid scanner type strings.
var ValidScannerTypes = []string{"sbom", "vuln", "secrets", "iac", "compliance", "sast", "zap"}
type scanCommand struct{}

var _ core.SimpleCommandPort = (*scanCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&scanCommand{},
	}
}

func (c *scanCommand) Name() string { return "scan" }

func (c *scanCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "scan",
		Short:         "Security scanning and evidence collection for audit compliance",
		Long:          "Scan modules for security vulnerabilities using industry-standard tools.\n\nIf no modules specified, scans all modules in the repository.\nIf no --scanner specified, uses default scanners for each module type\n(configured in contracts/scanner/0.1.0/schemas/defaults/).\n\nSupported scanners:\n  - sbom: Software Bill of Materials (Trivy, CycloneDX format)\n  - vuln: Vulnerability scanning (Trivy)\n  - secrets: Secrets detection (Trivy)\n  - iac: Infrastructure as Code scanning (Trivy)\n  - compliance: CIS compliance checking (Trivy)\n  - sast: Static Application Security Testing (Semgrep)\n  - zap: Dynamic Application Security Testing (OWASP ZAP)\n\nExample:\n  scan                                  # All modules, default scanners\n  scan core                         # Single module, default scanners\n  scan --scanner sbom                   # All modules, SBOM only\n  scan core --scanner sbom,vuln     # Single module, specific scanners\n\nEvidence output: out/scan/<module>/<scanner>/",
		Args:          "modules",
		Flags: []core.FlagSpec{
			{Name: "scanner", Type: "string", Usage: "Scanner types to run (comma-separated: sbom,vuln,secrets,iac,compliance,sast,zap)"},
			{Name: "debug", Type: "bool", Shorthand: "d", DefaultValue: "false", Usage: "Enable debug logging to out/logs/security"},
			{Name: "tui", Type: "bool", DefaultValue: "auto", Usage: "Enable TUI console (default: auto-detect)"},
			{Name: "no-tui", Type: "bool", DefaultValue: "false", Usage: "Disable TUI console"},
			{Name: "tui-height", Type: "int", DefaultValue: "8", Usage: "Set TUI console height (3-20)"},
			{Name: "ascii", Type: "bool", DefaultValue: "false", Usage: "Use ASCII-only characters in TUI"},
			{Name: "skip-tui-delay", Type: "bool", DefaultValue: "false", Usage: "Skip TUI exit delay (exit immediately when done)"},
			{Name: "sequential", Type: "bool", DefaultValue: "false", Usage: "Run scans sequentially instead of in parallel"},
			{Name: "turbo", Type: "bool", DefaultValue: "false", Usage: "Enable turbo mode for faster scanning (increases parallelism)"},
			{Name: "skip-cache", Type: "bool", DefaultValue: "false", Usage: "Skip incremental cache, force full scan"},
			{Name: "skip-deps", Type: "bool", DefaultValue: "false", Usage: "Skip system dependency verification (trivy, semgrep, etc.)"},
		},
	}
}

func (c *scanCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Scan()
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
		Type:           core.ActionScan,
		CommandPath:    "scan",
		OutputDir:      paths.OutSecurityRelPath,
		Monikers:       shared.Monikers,
		Sequential:     sequential,
		Turbo:          shared.Turbo,
		MaxConcurrency: shared.MaxConcurrency,
		ForceRebuild:   shared.CacheConfig.ShouldSkipState(),
		SkipDeps:       shared.SkipDeps,
		UseTUI:         shared.UseTUI,
		TUIHeight:      shared.TUIHeight,
		TUIASCIIMode:   shared.TUIASCIIMode,
		TUI3Demo:       shared.TUI3Demo,
		SkipTUIDelay:   shared.SkipTUIDelay,
		DebugMode:      shared.Debug,
		ShowTimings:    shared.ShowTimings,
		DryRun:         shared.DryRun,
		CacheConfig:    shared.CacheConfig,
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
func parseScannerList(input string) ([]evidence.ScannerType, error) {
	var scanners []evidence.ScannerType
	for _, s := range strings.Split(input, ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			continue
		}
		toolID := tool.ScannerToolIDForCategory(s)
		if toolID == "" {
			return nil, fmt.Errorf("invalid scanner type: %s (valid: %s)", s, strings.Join(ValidScannerTypes, ", "))
		}
		scanners = append(scanners, evidence.ScannerType(toolID))
	}
	if len(scanners) == 0 {
		return nil, fmt.Errorf("no valid scanner types specified")
	}
	return scanners, nil
}

// parseSeverityList parses a comma-separated list of severity levels.
func parseSeverityList(input string) ([]evidence.Severity, error) {
	var severities []evidence.Severity
	for _, s := range strings.Split(input, ",") {
		s = strings.TrimSpace(strings.ToUpper(s))
		if s == "" {
			continue
		}
		severity, valid := evidence.ParseSeverity(s)
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
	log.Info("  --ascii                   Use ASCII-only characters in TUI")
	log.Info("  --skip-tui-delay          Skip TUI exit delay (exit immediately when done)")
	log.Info("")
	log.Info("Scanner-specific flags:")
	log.Info("  --format <fmt>            SBOM format (cyclonedx, spdx, spdx-json, github)")
	log.Info("  --severity <levels>       Vulnerability severity filter (LOW,MEDIUM,HIGH,CRITICAL)")
	log.Info("  --config <ruleset>        SAST config (auto, p/security-audit, p/owasp-top-ten)")
	log.Info("  --compliance <std>        Compliance standard (k8s-cis, docker-cis, k8s-nsa)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  scan                                  # All modules, default scanners")
	log.Info("  scan core                         # Single module, default scanners")
	log.Info("  scan core clie                 # Multiple modules")
	log.Info("  scan --scanner sbom                   # All modules, single scanner")
	log.Info("  scan core --scanner sbom,vuln     # Single module, multiple scanners")
	log.Info("  scan --severity HIGH                  # Vuln with severity filter")
	log.Info("  scan --scanner sast --config p/owasp-top-ten")
	log.Info("  scan core --debug                 # With debug logging")
	log.Info("  scan --sequential                     # Run all scans sequentially")
	log.Info("")
	log.Info("Evidence output:")
	log.Info("  out/scan/<module>/<scanner>/")
	log.Info("")
	log.Info("Configuration:")
	log.Info("  Default scanners per module type: contracts/security/.../policies.yml")
	log.Info("  Modules to skip (skip_modules):   contracts/security/.../policies.yml")
	log.Info("  Docker image versions:            contracts/security/.../scanners.yml")
	log.Info("")
	log.Info("External tools:")
	log.Info("  This command uses third-party security tools. See the NOTICE file")
	log.Info("  in the repository root for full attribution and licensing information.")
}
