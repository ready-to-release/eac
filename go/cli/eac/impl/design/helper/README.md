# design (helper)

Provides Structurizr workspace validation, export, serving, and utility functions for the design command family.

## Key Types

- **`StructurizrValidator`** -- Interface for validating workspaces using Structurizr CLI via Docker (ValidateModule, ValidateModuleFile, ValidateAll, IsDockerRunning)
- **`StructurizrValidatorImpl`** -- Concrete Docker-based implementation of StructurizrValidator
- **`MockStructurizrValidator`** -- Builder-pattern mock implementation for testing with configurable module/file results and Docker state
- **`ValidationResult`** -- Outcome of validating a module's workspace(s), with errors, warnings, file results, execution time, and raw CLI output
- **`FileValidationResult`** -- Outcome of validating a single DSL file within a module
- **`ValidationMessage`** -- A single error or warning with severity, message, and optional line/column
- **`ValidationSummary`** -- Aggregated results for multiple modules with pass/fail counts and totals
- **`StructurizrExporter`** -- Interface for exporting workspace views to SVG (ExportModule, ExportAll, IsDockerRunning)
- **`StructurizrExporterImpl`** -- Concrete implementation that exports via Structurizr CLI to PlantUML, then renders to SVG
- **`ExportResult`** -- Result of exporting a module's workspace including views, DSL hash, and execution time
- **`ExportedView`** -- A single exported view with view key, SVG path, module, and DSL hash
- **`ExportSummary`** -- Aggregated export results for multiple modules

## Key Functions

- `NewValidator` -- Creates a validator, returning a mock if `CLIE_MOCK_DOCKER` is set
- `NewExporter` / `NewExporterWithOutput` -- Creates a Structurizr exporter with optional custom output directory
- `StartStructurizrLite` -- Starts the Structurizr Lite Docker container and opens the browser
- `StopStructurizrLite` -- Stops a running Structurizr Lite container for a module
- `IsStructurizrLiteRunning` -- Checks if Structurizr Lite is currently running for a module
- `OpenBrowser` -- Opens a URL in the platform-default browser (Windows, macOS, Linux)
- `ValidateModuleName` -- Validates module names against path traversal, invalid characters, and Windows reserved names
- `CleanModuleName` -- Removes common prefixes/suffixes from module names (specs/, .design, src/)
- `FormatValidationResult` / `FormatValidationSummary` -- Formats validation results for console output
- `WriteValidationResultJSON` / `WriteValidationSummaryJSON` -- Writes validation results to JSON files
- `HashDSLContent` -- Returns first 8 characters of SHA256 hash of normalized DSL content for cache invalidation
- `ParseViewKeysFromDSL` -- Lightweight parser that extracts view keys from workspace.dsl without running Structurizr
- `GetStructurizrCLIImage` / `GetStructurizrLiteImage` -- Returns Docker images from tool-config.yml with fallback defaults

## Patterns

- **Interface with mock for testing**: `StructurizrValidator` interface has both a real Docker-based implementation and a builder-pattern mock (`MockStructurizrValidator`) with `WithValidResult`, `WithInvalidResult`, etc.
- **Environment-based mock activation**: `NewValidator` automatically returns a mock when `CLIE_MOCK_DOCKER` is set, enabling acceptance tests without Docker
- **Multi-file validation**: Modules can have multiple DSL files; validation iterates all non-fragment files (excluding `_*.dsl`)
- **Two-step Docker export**: Export uses Structurizr CLI to generate PlantUML, then PlantUML container to render SVG
- **Platform-aware operations**: Browser opening and Docker volume formatting handle Windows, macOS, and Linux differences
- **Limited buffer for Docker output**: `limitedBuffer` prevents memory exhaustion from runaway Docker container output

## Internal Structure

| File | Responsibility |
| --- | --- |
| validator.go | StructurizrValidator interface, StructurizrValidatorImpl, Docker-based validation execution, output parsing, module/file discovery, limitedBuffer, and formatDockerVolume |
| validator_mock.go | MockStructurizrValidator with builder methods for configuring test behavior and fixture loading |
| validation.go | Input validation utilities: ValidateModuleName, CleanModuleName, ValidateIdentifier |
| validation_formatter.go | Console formatting for ValidationResult and ValidationSummary, plus JSON file writing |
| export.go | StructurizrExporter interface and implementation: Docker-based PlantUML+SVG export pipeline, DSL hashing, view key parsing |
| serve.go | Structurizr Lite container lifecycle management: start, stop, and status checking |
| browser.go | Cross-platform browser opening (OpenBrowser, DetectBrowser) |
| constants.go | Package constants: file names, Docker configuration, timeouts, buffer limits, Docker image getters |

## Dependencies

- `go/adapters/docker` -- Docker serve helper for container lifecycle (StartServe, StopServe, IsServing, OpenBrowserWithFallback)
- `go/adapters/docker/util` -- Docker utility functions (IsDockerAvailable, FormatDockerVolume)
- `go/core/config` -- Loading repository configuration for specs directory path
- `go/core/environments` -- Environment variable names (EnvCLIEMockDocker)
- `go/core/logging` -- Component logger for info and error output
- `go/core/paths` -- Canonical path resolution for workspace DSL, design directories, and cache paths
- `go/core/repository` -- Repository root discovery
- `go/core/tool` -- Tool image resolution from tool-config.yml with defaults

## Role in System

This package provides the infrastructure for the `validate design`, `serve design`, and `update structurizr` commands. It handles the full lifecycle of Structurizr workspace operations -- from validating DSL syntax via Docker, to exporting views as SVGs, to serving Structurizr Lite for interactive editing. The mock validator enables acceptance testing of the design command family without requiring Docker.

## Code Health

### Tech Debt
- validator_mock.go:170 contains a fragile string slice check (`contentStr[0:7] == "invalid"`) that could panic on short content

### Pain Points
- None identified

### Optimization Opportunities
- None identified
