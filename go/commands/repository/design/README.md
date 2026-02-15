# design

Provides architecture documentation operations using Structurizr DSL, including workspace validation via Docker, SVG diagram export through a PlantUML pipeline, and live-serving Structurizr Lite containers for interactive design editing.

## Key Types

- **`ValidationResult`** -- Per-module validation outcome with errors, warnings, file results, and timing
- **`FileValidationResult`** -- Per-file validation outcome within a multi-file module workspace
- **`ValidationMessage`** -- Single error or warning with severity, message, and optional line/column
- **`ValidationSummary`** -- Aggregated validation results across all modules
- **`StructurizrValidator`** -- Interface for DSL workspace validation (module, file, or all)
- **`ExportResult`** -- Per-module export outcome with views, DSL hash, and output directory
- **`ExportedView`** -- Single exported SVG view with view key, path, module, and DSL hash
- **`ExportSummary`** -- Aggregated export results across all modules
- **`StructurizrExporter`** -- Interface for exporting workspace views to SVG
- **`Output`** -- Unified output handler for console messages and structured logging

## Patterns

- Docker-based tooling: validation and export run Structurizr CLI/Lite via Docker containers
- Two-step export pipeline: Structurizr CLI exports to PlantUML, then PlantUML renders to SVG
- Content-addressed caching: DSL hash (SHA256 prefix) embedded in exported SVG filenames
- Multi-file validation: validates all DSL files in a module's design folder, skipping fragment files
- Mock injection via environment: `CLIE_MOCK_DOCKER` switches to mock validator for testing
- Volume mount path normalization: Windows paths converted to Docker-compatible format

## Internal Structure

| File | Responsibility |
| --- | --- |
| output.go | `Output` wrapper for dual console/log output with info, error, warn, debug, progress |
| helper/constants.go | Docker image defaults, timeouts, buffer limits, and tool image resolution |
| helper/serve.go | Start/stop Structurizr Lite Docker containers with dynamic port allocation |
| helper/validator.go | `StructurizrValidator` implementation: Docker-based DSL validation with output parsing |
| helper/export.go | `StructurizrExporter` implementation: PlantUML pipeline, SVG collection, DSL hashing |
| helper/validation.go | Input validation: module name safety, path traversal prevention, identifier checks |
| helper/validation_formatter.go | Console and JSON formatting for validation results and summaries |
| helper/browser.go | Cross-platform browser opening for serve URLs |
| helper/validator_mock.go | Mock validator for testing without Docker |

## Dependencies

- `adapters/docker` -- Docker container management for serve operations
- `adapters/docker/util` -- Docker availability check and volume path formatting
- `core/config` -- configuration loading for specs directory paths
- `core/environments` -- environment variable constants for mock control
- `core/logging` -- structured logging
- `core/paths` -- design directory, workspace DSL, and cache path resolution
- `core/repository` -- repository root discovery
- `core/tool` -- Docker image resolution from tool-config.yml

## Role in System

The `design` package provides architecture-as-code capabilities for `eac`, enabling teams to author Structurizr DSL workspaces, validate them in CI via Docker, export diagrams to SVG for documentation, and interactively edit designs with a live Structurizr Lite server. It is used by both the `validate design` and `serve design` commands.

## Code Health

### Tech Debt
- helper/export.go (477 lines) now has export_test.go (803 lines) covering all core functionality
- helper/validator_mock.go (310 lines) is large for a mock but serves multiple test scenarios

### Pain Points
- helper/validation_formatter.go (269 lines) lacks dedicated unit tests despite complex formatting logic

### Optimization Opportunities
- None identified
