# Security Policy

## Reporting a Vulnerability

**For security vulnerabilities, please use GitHub's private vulnerability reporting:**

1. Go to the [Security tab](https://github.com/ready-to-release/eac/security)
2. Click "Report a vulnerability"
3. Provide detailed information about the issue

**Response times:**

- Initial acknowledgment: Within 48 hours
- Severity assessment: Within 7 days
- Fix timeline varies by severity

**Alternative:** For sensitive issues that cannot be reported via GitHub, contact: <security@ready-to-release.dev> (TODO: establish)

## Supported Versions

This is a **multi-module Go workspace**. All modules are currently in active development:

| Module           | Path                  | Status             | Go Version |
|------------------|-----------------------|--------------------|------------|
| EaC Core         | `go/core`             | Active Development | 1.24+      |
| EaC CLI          | `go/cli/eac`          | Active Development | 1.24+      |
| EaC MCP Commands | `go/cli/mcp`          | Active Development | 1.24+      |
| EaC Specs        | `go/specs`            | Active Development | 1.24+      |
| CLIE CLI          | `go/cli/clie`          | Active Development | 1.24+      |

Security patches are applied to the main branch and will be included in the next release.

## Security Practices

This repository implements comprehensive security measures:

- **Automated scanning:** CodeQL, Trivy, Semgrep, OWASP ZAP
- **Security commands:** `eac scan` with multiple scanner types
- **Continuous monitoring:** Security workflows run on every push and PR

For detailed information, see our comprehensive documentation:

- [Security Practices](docs/explanation/continuous-delivery/security/) - Shift-left security, SAST, DAST, supply chain security
- [Security Workflows](docs/reference/repository/continuous-delivery/workflows/security-workflows.md) - Automated security scanning
- [Scan Command](docs/how-to-guides/eac/commands/build-test-validate/scan-for-security-issues.md) - Running security scans locally

## License

- Code: [MIT License](LICENSE)
- Documentation: CC-BY-SA-4.0
