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
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
)

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
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printSecurityUsage()
		return 1
	}
}

func printSecurityUsage() {
	fmt.Println("Security scanning and evidence collection for audit compliance")
	fmt.Println()
	fmt.Println("Usage: security <scanner> [modules...] [flags]")
	fmt.Println()
	fmt.Println("Available scanners:")
	fmt.Println("  sbom                      Generate Software Bill of Materials (Trivy)")
	fmt.Println("  vuln                      Vulnerability scanning (Trivy)")
	fmt.Println("  secrets                   Secrets detection (Trivy)")
	fmt.Println("  compliance                CIS compliance checking (Trivy)")
	fmt.Println("  iac                       Infrastructure as Code scanning (Trivy)")
	fmt.Println("  sast                      Static Application Security Testing (Semgrep)")
	fmt.Println("  zap                       Dynamic Application Security Testing (OWASP ZAP)")
	fmt.Println()
	fmt.Println("Module arguments:")
	fmt.Println("  [modules...]              One or more module monikers to scan")
	fmt.Println("                            If no modules specified, scans all modules")
	fmt.Println()
	fmt.Println("Common flags:")
	fmt.Println("  --debug, -d               Enable debug logging")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  security sbom src-core                    # Single module")
	fmt.Println("  security sbom src-core src-cli            # Multiple modules")
	fmt.Println("  security sbom                             # All modules")
	fmt.Println("  security vuln src-core --severity HIGH    # With scanner flags")
	fmt.Println("  security sbom src-core --debug            # With debug logging")
	fmt.Println()
	fmt.Println("Evidence output:")
	fmt.Println("  out/security/<module>/<scanner>/<timestamp>.json")
	fmt.Println()
	fmt.Println("External tools:")
	fmt.Println("  This command uses third-party security tools. See the NOTICE file")
	fmt.Println("  in the repository root for full attribution and licensing information.")
	fmt.Println()
	fmt.Println("Use 'security <scanner> --help' for scanner-specific options.")
}
