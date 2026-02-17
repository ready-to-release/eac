# Scan

<!-- book:cmd scan -->

## Scanner Types

The scan command supports multiple scanner types via the `--scanner` flag:

| Type         | Description                          | Tool                             |
| ------------ | ------------------------------------ | -------------------------------- |
| `sbom`       | Software Bill of Materials           | Trivy                            |
| `vuln`       | Vulnerability scanning               | Trivy                            |
| `secrets`    | Secret detection                     | Trivy                            |
| `iac`        | Infrastructure as Code scanning      | Trivy                            |
| `compliance` | Compliance checking                  | Trivy                            |
| `sast`       | Static Application Security Testing  | Semgrep                          |
| `zap`        | Dynamic Application Security Testing | OWASP ZAP (see subcommand below) |

## Overview

The scan command runs security scanners with automatic detection and configuration.

## Usage Examples

### Run Default Scanners

```bash
# All modules with default scanners
eac scan

# Specific module with default scanners
eac scan eac-core
```

### Run Specific Scanner Types

```bash
# Single scanner type
eac scan --scanner vuln
eac scan --scanner sbom

# Multiple scanner types
eac scan --scanner vuln,secrets,sbom

# Specific module with specific scanners
eac scan eac-core --scanner vuln,secrets
```

### Dynamic Testing (ZAP)

ZAP is a special case requiring a URL target, so it has its own subcommand:

```bash
eac scan zap eac-api --target http://localhost:8080
```

## See Also

- [scan Commands Category](../categories/scan.md)
- [validate control-tags](../validate/control-tags.md) - Validate security control mappings
