// Command: security
// Short: Security scanning and evidence collection for audit compliance
// Long: The security command provides audit-ready security evidence collection using industry-standard
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
package security

import (
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Security)
}

// Security command entry point
func Security() int {
	args := os.Args[2:] // Skip program name and "security"

	if len(args) == 0 {
		printSecurityUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printSecurityUsage()
		return 0
	case "sbom", "vuln", "secrets", "compliance", "iac", "sast", "zap":
		// Handled by separate registrations in respective sub-command files
		return 0
	default:
		log.Errorf("unknown subcommand: %s", args[0])
		log.Info("")
		printSecurityUsage()
		return 1
	}
}

func printSecurityUsage() {
	log.Info("Security scanning and evidence collection for audit compliance")
	log.Info("")
	log.Info("Usage: security <scanner> [modules...] [flags]")
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
	log.Info("  security sbom src-core                    # Single module")
	log.Info("  security sbom src-core src-cli            # Multiple modules")
	log.Info("  security sbom                             # All modules")
	log.Info("  security vuln src-core --severity HIGH    # With scanner flags")
	log.Info("  security sbom src-core --debug            # With debug logging")
	log.Info("")
	log.Info("Evidence output:")
	log.Info("  out/security/<module>/<scanner>/<timestamp>.json")
	log.Info("")
	log.Info("External tools:")
	log.Info("  This command uses third-party security tools. See the NOTICE file")
	log.Info("  in the repository root for full attribution and licensing information.")
	log.Info("")
	log.Info("Use 'security <scanner> --help' for scanner-specific options.")
}
