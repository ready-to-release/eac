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
// HasSideEffects: false
// Args: modules
package scan

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/scan/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/cmdframework"
	"github.com/ready-to-release/eac/go/eac/commands/internal/environment"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

// ValidScannerTypes lists all valid scanner type strings
var ValidScannerTypes = []string{"sbom", "vuln", "secrets", "iac", "compliance", "sast", "zap"}

func init() {
	registry.Register(Scan)
}

// scanArgs holds parsed arguments for the scan command
type scanArgs struct {
	Monikers         []string
	Scanners         []internal.ScannerType
	Debug            bool
	Sequential       bool
	UseTUI           bool
	TUIHeight        int
	TUIExplicitlySet bool
	// Scanner-specific options
	SBOMFormat         string               // SBOM format (cyclonedx, spdx, etc.)
	VulnSeverities     []internal.Severity  // Vuln severity filter
	SemgrepConfig      string               // SAST config
	ComplianceStandard string               // Compliance standard
}

// Scan command entry point
func Scan() int {
	args := os.Args[2:] // Skip program name and "scan"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printScanUsage()
		return 0
	}

	// Parse arguments (empty args = scan all modules with defaults)
	parsed, err := parseArgs(args)
	if err != nil {
		log.Errorf("Error: %v", err)
		printScanUsage()
		return 1
	}

	// Validate TUI usage
	env := environment.Detect()
	if err := env.ValidateTUI(parsed.TUIExplicitlySet, parsed.UseTUI); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Create command config
	cmdCfg := &cmdframework.CommandConfig{
		Type:        cmdframework.CommandTypeScan,
		ActionVerb:  "Scanning",
		OutputDir:   "out/scan",
		LogFileName: "scan.log",
		Monikers:    parsed.Monikers,
		Layered:     false, // Scan uses parallel execution
		Sequential:  parsed.Sequential,
		UseTUI:      parsed.UseTUI,
		TUIHeight:   parsed.TUIHeight,
		DebugMode:   parsed.Debug,
	}

	// Create multi-scanner config
	multiCfg := &MultiScanConfig{
		Scanners:           parsed.Scanners,
		SBOMFormat:         parsed.SBOMFormat,
		VulnSeverities:     parsed.VulnSeverities,
		SemgrepConfig:      parsed.SemgrepConfig,
		ComplianceStandard: parsed.ComplianceStandard,
	}

	return RunMultiScan(cmdCfg, multiCfg)
}

// parseArgs parses command line arguments
func parseArgs(args []string) (*scanArgs, error) {
	env := environment.Detect()
	parsed := &scanArgs{
		UseTUI:             env.ShouldUseTUI(),
		TUIHeight:          tui.DefaultHeight,
		SBOMFormat:         "cyclonedx",
		SemgrepConfig:      "auto",
		ComplianceStandard: "k8s-cis",
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--scanner":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--scanner requires a value")
			}
			i++
			scanners, err := parseScannerList(args[i])
			if err != nil {
				return nil, err
			}
			parsed.Scanners = scanners
		case "--format":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--format requires a value")
			}
			i++
			parsed.SBOMFormat = args[i]
		case "--severity":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--severity requires a value")
			}
			i++
			severities, err := parseSeverityList(args[i])
			if err != nil {
				return nil, err
			}
			parsed.VulnSeverities = severities
		case "--config":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--config requires a value")
			}
			i++
			parsed.SemgrepConfig = args[i]
		case "--compliance":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--compliance requires a value")
			}
			i++
			parsed.ComplianceStandard = args[i]
		case "--debug", "-d":
			parsed.Debug = true
		case "--sequential":
			parsed.Sequential = true
		case "--tui":
			parsed.UseTUI = true
			parsed.TUIExplicitlySet = true
		case "--no-tui":
			parsed.UseTUI = false
			parsed.TUIExplicitlySet = true
		case "--tui-height":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--tui-height requires a value")
			}
			i++
			height, err := strconv.Atoi(args[i])
			if err != nil || height < 3 || height > 20 {
				return nil, fmt.Errorf("--tui-height must be a number between 3 and 20")
			}
			parsed.TUIHeight = height
		default:
			// Handle --flag=value syntax
			if strings.HasPrefix(arg, "--scanner=") {
				scanners, err := parseScannerList(strings.TrimPrefix(arg, "--scanner="))
				if err != nil {
					return nil, err
				}
				parsed.Scanners = scanners
			} else if strings.HasPrefix(arg, "--format=") {
				parsed.SBOMFormat = strings.TrimPrefix(arg, "--format=")
			} else if strings.HasPrefix(arg, "--severity=") {
				severities, err := parseSeverityList(strings.TrimPrefix(arg, "--severity="))
				if err != nil {
					return nil, err
				}
				parsed.VulnSeverities = severities
			} else if strings.HasPrefix(arg, "--config=") {
				parsed.SemgrepConfig = strings.TrimPrefix(arg, "--config=")
			} else if strings.HasPrefix(arg, "--compliance=") {
				parsed.ComplianceStandard = strings.TrimPrefix(arg, "--compliance=")
			} else if strings.HasPrefix(arg, "--tui-height=") {
				heightStr := strings.TrimPrefix(arg, "--tui-height=")
				height, err := strconv.Atoi(heightStr)
				if err != nil || height < 3 || height > 20 {
					return nil, fmt.Errorf("--tui-height must be a number between 3 and 20")
				}
				parsed.TUIHeight = height
			} else if strings.HasPrefix(arg, "--") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			} else {
				// Assume it's a module moniker
				parsed.Monikers = append(parsed.Monikers, arg)
			}
		}
	}

	return parsed, nil
}

// parseScannerList parses a comma-separated list of scanner types
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

// parseSeverityList parses a comma-separated list of severity levels
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
	log.Info("  Docker image versions:            .r2r/eac/security-tools.yml")
	log.Info("")
	log.Info("External tools:")
	log.Info("  This command uses third-party security tools. See the NOTICE file")
	log.Info("  in the repository root for full attribution and licensing information.")
}
