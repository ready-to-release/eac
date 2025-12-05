// Command: scan
// Short: Security scanning and evidence collection for audit compliance
// Long: The scan command provides audit-ready security evidence collection using industry-standard
// Long: tools. All scanners generate timestamped, SHA256-signed evidence files for regulatory compliance
// Long: and immutable audit trails.
// Long:
// Long: Supported scanners:
// Long:   - sbom: Software Bill of Materials (Trivy, CycloneDX format)
// Long:   - vuln: Vulnerability scanning (Trivy)
// Long:   - secrets: Secrets detection (Trivy)
// Long:   - compliance: CIS compliance checking (Trivy)
// Long:   - iac: Infrastructure as Code scanning (Trivy)
// Long:   - sast: Static Application Security Testing (Semgrep)
// Long:   - zap: Dynamic Application Security Testing (OWASP ZAP)
// Long:
// Long: All evidence files are written to out/security/<module>/<scanner>/<timestamp>.json
// Long:
// Long: External tools used (see NOTICE file for full attribution):
// Long:   - Trivy (Apache 2.0) - https://github.com/aquasecurity/trivy
// Long:   - Semgrep (LGPL 2.1) - https://github.com/semgrep/semgrep
// Long:   - OWASP ZAP (Apache 2.0) - https://github.com/zaproxy/zaproxy
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug logging to out/logs/security
// HasSideEffects: false
package scan

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Scan)
}

// Scan command entry point
func Scan() int {
	args := os.Args[2:] // Skip program name and "scan"

	if len(args) == 0 {
		printScanUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printScanUsage()
		return 0
	case "sbom", "vuln", "secrets", "compliance", "iac", "sast", "zap":
		// Handled by separate registrations in respective sub-command files
		return 0
	default:
		log.Errorf("unknown subcommand: %s", args[0])
		log.Info("")
		printScanUsage()
		return 1
	}
}

func printScanUsage() {
	log.Info("Security scanning and evidence collection for audit compliance")
	log.Info("")
	log.Info("Usage: scan <scanner> [modules...] [flags]")
	log.Info("")
	log.Info("Available scanners:")
	log.Info("  sbom                      Generate Software Bill of Materials (Trivy)")
	log.Info("  vuln                      Vulnerability scanning (Trivy)")
	log.Info("  secrets                   Secrets detection (Trivy)")
	log.Info("  compliance                CIS compliance checking (Trivy)")
	log.Info("  iac                       Infrastructure as Code scanning (Trivy)")
	log.Info("  sast                      Static Application Security Testing (Semgrep)")
	log.Info("  zap                       Dynamic Application Security Testing (OWASP ZAP)")
	log.Info("")
	log.Info("Module arguments:")
	log.Info("  [modules...]              One or more module monikers to scan")
	log.Info("                            If no modules specified, scans all modules")
	log.Info("")
	log.Info("Common flags:")
	log.Info("  --debug, -d               Enable debug logging")
	log.Info("")
	log.Info("Examples:")
	log.Info("  scan sbom eac-core                    # Single module")
	log.Info("  scan sbom eac-core r2r-cli            # Multiple modules")
	log.Info("  scan sbom                             # All modules")
	log.Info("  scan vuln eac-core --severity HIGH    # With scanner flags")
	log.Info("  scan sbom eac-core --debug            # With debug logging")
	log.Info("")
	log.Info("Evidence output:")
	log.Info("  out/security/<module>/<scanner>/<timestamp>.json")
	log.Info("")
	log.Info("Configuration:")
	log.Info("  Docker image versions are configured in .r2r/eac/security-tools.yml")
	log.Info("")
	log.Info("External tools:")
	log.Info("  This command uses third-party security tools. See the NOTICE file")
	log.Info("  in the repository root for full attribution and licensing information.")
	log.Info("")
	log.Info("Use 'scan <scanner> --help' for scanner-specific options.")
}
