# evidence

Provides evidence collection for risk assessment, including discovery and loading of test results, security scans (SBOM, vulnerability, SAST, DAST, secrets, compliance, IaC), and evidence age validation.

## Key Types

- **`EvidenceType`** -- String enum for evidence types (unit-test, acceptance-test, vulnerability-scan, sbom, secrets-scan, sast-scan, compliance-scan, iac-scan, dast-scan)
- **`Evidence`** -- A piece of evidence with type, path, modification time, module, and associated control IDs
- **`TestResults`** -- Paths to test evidence files for a module (unit tests and acceptance tests)
- **`SecurityResults`** -- Paths to security evidence files for a module (SBOM, vuln, secrets, SAST, compliance, IaC, ZAP)
- **`SecurityEvidenceFile`** -- Standardized security evidence format with module, scanner, timestamp, SHA256, and findings
- **`EvidenceCollection`** -- All collected evidence for a module (tests, security, summaries, warnings)
- **`TestManifestData`** -- Simplified test manifest info for evidence collection (avoids import cycles)
- **`TestCase`** -- Single test case result with status, duration, and control IDs
- **`TestSummary`** -- Aggregate test result counts
- **`VulnerabilitySummary`** -- Aggregate vulnerability counts by severity
- **`SBOMSummary`** -- Aggregate SBOM component count
- **`ControlTestEvidence`** -- Evidence for a specific control from tests (pass/fail/skip counts)
- **`TestEvidence`** -- Single test covering a control with status and location
- **`EvidenceAgePolicy`** -- Policy for maximum evidence age
- **`ScannerType`** -- String enum for security scanner types (sbom, vuln, secrets, compliance, iac, sast, zap)
- **`TestEntryData`** -- Test entry data used by evidence extraction (avoids import cycles)
- **`SuiteResultData`** -- Per-suite test result information
- **`TestArtifactData`** -- Test artifact reference data

## Key Functions

- **`FindLatestSecurityScan()`** -- Find the security scan file for a module and scanner type
- **`LoadSecurityEvidence()`** -- Load and parse a standardized security evidence JSON file
- **`GetControlTestEvidenceFromManifest()`** -- Build comprehensive control evidence from test entries with suite filtering
- **`DefaultEvidenceAgePolicy()`** -- Return the default evidence age policy from configuration
- **`HasAnyEvidence()`** -- Check if an evidence collection has any evidence at all
- **`HasTestEvidence()`** -- Check if test evidence exists in a collection
- **`HasSecurityEvidence()`** -- Check if security evidence exists in a collection
- **`LatestModTime()`** -- Get the most recent modification time across all evidence

## Patterns

- Multi-scanner support: handles 7+ scanner types with per-scanner file discovery
- Config-driven paths: uses repository configuration for scan output directory resolution
- Suite filtering: supports composite suite filters like `"unit+integration"` for control evidence
- Control tag extraction: extracts `@control:` tags from test entries for control-evidence mapping
- Age policy: configurable maximum evidence age for staleness detection
- Import-cycle workaround: `TestManifestData`, `TestEntryData`, `SuiteResultData`, and `TestArtifactData` are local copies of test-internal types, avoiding a cycle between evidence and impl/test/internal (see `types_testing.go` doc comments)

## Internal Structure

| File | Responsibility |
| --- | --- |
| types_base.go | Core evidence types (`EvidenceType`, `Evidence`) and age policy |
| types_testing.go | Test-related types (`TestResults`, `TestCase`, `TestSummary`, `TestManifestData`, `ControlTestEvidence`, `TestEvidence`) |
| types_security.go | Security-related types (`SecurityResults`, `SecurityEvidenceFile`, `VulnerabilitySummary`, `SBOMSummary`) |
| types_collection.go | `EvidenceCollection` type and its convenience methods |
| security_loader.go | Security scan file discovery and parsing |
| test_loader.go | Test evidence loading with control tag extraction and suite filtering |

## Dependencies

- `core/config` -- evidence age configuration and scan output path resolution

## Role in System

The `evidence` package is the data collection layer of the risk assessment pipeline. It discovers and loads evidence files from test runs and security scans, organizing them into structured collections that feed into the scoring engine and OSCAL document generation.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- `security_loader.go` is 348 lines, making it the largest file in the package due to handling multiple scanner types.

### Optimization Opportunities

- None identified.
