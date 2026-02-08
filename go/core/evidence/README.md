# evidence

Writes and verifies security scan evidence files with SHA256 integrity hashes
for regulatory compliance and audit traceability.

## Key Types

- **`File`** -- Standardized evidence format with module, scanner, hash, findings
- **`ScannerType`** -- String alias for scanner tool IDs (trivy, semgrep, zap)
- **`ScanResult`** -- Outcome of a security scan (success, path, exit code)
- **`Severity`** -- Vulnerability severity level (LOW through CRITICAL)

## Patterns

- SHA256 integrity: findings are hashed and embedded in the evidence file
- Standardized schema: all scanners write the same `File` structure
- Error evidence: failed scans still produce evidence files for audit trail

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `File`, `ScannerType`, `ScanResult`, `Severity`, timestamps |
| evidence.go | `WriteEvidence`, `ReadEvidence`, `VerifyEvidence`, component variants |

## Dependencies

- `core/config` -- repository paths for output directories
- `core/domain` -- valid scanner category constants
- `core/tool` -- scanner tool ID constants

## Role in System

The `evidence` package provides the data layer for the `scan` command family
in `core`. Each scanner adapter writes evidence through this package, producing
JSON files under `out/scan/<module>/` that downstream compliance commands can
verify and aggregate.

## Code Health

### Tech Debt
- None identified

### Pain Points
- `evidence.go`: `WriteEvidence` and `WriteComponentEvidence` share ~80% of their logic; duplication could be reduced with a shared internal writer

### Optimization Opportunities
- Extract a common `writeFile(outputDir, module, scanner, findings)` helper to eliminate duplication between module-level and component-level writers (low effort, medium value)
- Package is compact (159 lines + 187 lines of tests) and well-covered; minimal issues
