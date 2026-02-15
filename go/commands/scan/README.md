# scan

Runs security scanners (SBOM, vulnerability, secrets, IaC, compliance, SAST, DAST) across modules in parallel using the command framework, with incremental UoW-level caching and evidence file output for audit compliance.

## Key Types

- **`MultiScanConfig`** -- Holds scanner selection, SBOM format, severity filter, SAST config, and compliance standard
- **`ScanFrameworkConfig`** -- Per-scanner identity (type, name, emoji), custom args, output directory, and results map
- **`ScanModuleResult`** -- Per-module scan outcome with evidence path, error message, and duration
- **`ScanSpecificFlags`** -- Parsed scan-only flags (scanner list, format, severity, config, compliance)
- **`scanContext`** -- Mutable state during execution: cached modules/UoWs, input hashes, result tracking, manifest tracker
- **`ScanWorker`** -- Function signature for scanner-specific work receiving module contract and scan config

## Patterns

- Multi-scanner dispatch: runs multiple scanner types per module, determined by component type configuration
- Framework delegation: delegates to `cmdframework.Run` with scan-registered unit provider and worker
- Hook-based lifecycle: `AfterInit`, `AfterResolve`, and `AfterExecute` hooks customize framework phases
- UoW-level caching: incremental detection at component:scanner granularity for partial cache hits
- Evidence-based output: writes timestamped evidence files with SHA256 integrity for audit compliance
- Docker-based scanners: resolves scanner images from eac-security contract configuration
- Skip-module policy: reads skip list from security contract to exclude modules from scanning

## Internal Structure

| File | Responsibility |
| --- | --- |
| scan.go | Command entry point, flag parsing, scanner list validation, usage display |
| framework.go | Framework integration: hooks, multi-scan orchestration, Docker image resolution |
| scanflags.go | Scan-specific flag parsing (--scanner, --format, --severity, --config, --compliance) |
| unit_work.go | Component resolution: maps modules to scannable `UnitSpec` work items |
| unit_worker.go | Unit worker: runs scanners per component, handles caching, evidence writing, manifests |
| testing.go | Test helpers: `Evidence` type alias, mock setup stubs (Docker-level mocking) |

## Dependencies

- `contracts/core` -- action type constants (`ActionScan`)
- `adapters/docker` -- Docker-based ZAP scan execution (zap sub-package)
- `clibase/cmdframework` -- parallel execution framework with TUI and hooks
- `clibase/caching` -- incremental change detection for UoW-level cache
- `clibase/environment` -- execution environment detection
- `clibase/flags` -- shared and registry-based flag parsing
- `clibase/initsummary` -- init summary reporting
- `clibase/locking` -- component-level file locks for concurrent scan safety
- `clibase/output` -- formatted log writing to worker streams
- `clibase/registry` -- command registration
- `core/config` -- global config and security scanner lookup
- `core/domain/modules` -- module contract for component roots and types
- `core/environments` -- CI detection for cache behavior
- `core/evidence` -- evidence file writing, reading, and scanner type constants
- `core/hash` -- input file hashing for cache keys
- `core/logging` -- structured logging
- `core/output` -- UoW manifest tracker for cache validation
- `core/paths` -- security output path constants
- `core/resolver` -- component-to-scanner resolution
- `core/tool` -- scanner bridge, tool registry, and platform filtering
- `core/workunit` -- `UnitID`, `UnitSpec`, and state management

## Role in System

The `scan` package provides the security scanning command for `eac`, orchestrating multiple industry-standard tools (Trivy, Semgrep, OWASP ZAP) across all modules with parallel execution and incremental caching. It produces timestamped evidence files for audit compliance, integrating with the `cmdframework` to share execution infrastructure with build, test, and lint commands while maintaining security-specific behavior like Docker image resolution and evidence integrity verification.

## Code Health

### Tech Debt
- No files over 300 lines; largest files are framework.go and unit_worker.go
- unit_worker_test.go provides unit tests for worker logic
- No unit tests for scan.go, scanflags.go, or unit_work.go

### Pain Points
- None identified
