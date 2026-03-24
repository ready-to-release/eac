# scan

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

## See Also

- [validate control-tags](../validate/control-tags.md) - Validate security control mappings
