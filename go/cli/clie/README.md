# clie

Command-line interface for running containerized development extensions.

Parses CLI arguments, loads layered YAML configuration, manages Docker container lifecycles, and provides interactive TUI feedback.

## Key Types

- **`Config`** -- Top-level configuration with registry, defaults, extensions
- **`Extension`** -- Single extension definition (image, env, volumes)
- **`ContainerHost`** -- Manages Docker client and container operations
- **`Parser`** -- EBNF-aware command-line argument parser
- **`CommandValidator`** -- Validates parsed commands against business rules
- **`EmbeddedValidator`** -- JSON Schema validator using embedded contract
- **`RegistryCache`** -- TTL-based cache for GHCR tag/digest data
- **`Logger`** -- Leveled logger with raw, timestamped, JSON formatters
- **`Info`** -- Semver version metadata set via build flags
- **`Installer`** -- Orchestrates extension image pull operations

## Patterns

- Layered config: Base `clie.yml` merged with `.local.yml`, `.personal.yml`, `.dev.yml` overrides
- Lazy Docker connectivity: `ContainerHost` defers Ping until first use via `sync.Once`
- Argument boundary: Parser separates Viper-processed flags from container pass-through args
- Embedded contracts: EBNF grammar and JSON schema loaded from `contracts/clie` at init
- Atomic file writes: Cache files written to `.tmp` then renamed for crash safety

## Internal Structure

| File/Sub-package         | Responsibility                                                |
| ------------------------ | ------------------------------------------------------------- |
| main.go                  | Entry point; filters spurious shell redirect arguments        |
| doc.go                   | Package documentation comment                                 |
| cmd/                     | Cobra command tree (root, run, init, install, version, etc.)  |
| internal/cache/          | Registry and metadata caching with TTL expiration             |
| internal/command-parser/ | EBNF-based CLI argument parsing and boundary detection        |
| internal/conf/           | YAML config loading, merging, validation, repo root discovery |
| internal/docker/         | Docker client abstraction, image pull, container lifecycle    |
| internal/envconsts/      | Isolated environment variable constant definitions            |
| internal/extensions/     | Extension image installation orchestration                    |
| internal/github/         | GitHub CLI auth and GHCR registry tag listing                 |
| internal/logging/        | Leveled logger with configurable sinks and formatters         |
| internal/session/        | Cross-platform shell session identification                   |
| internal/terminal/       | Platform-specific terminal size detection                     |
| internal/tui/            | Bubble Tea spinner model for image pull progress              |
| internal/validator/      | Command and config validation against schema and rules        |
| internal/version/        | Semver parsing, comparison, and build-flag metadata           |

## Dependencies

- `contracts/clie/0.1.0` -- Embedded EBNF grammar and JSON schema

## Role in System

The clie CLI is the primary user-facing binary that developers invoke to run containerized extensions (e.g., `clie run eac`). It reads `.clie/clie.yml` configuration, resolves extension images from GHCR, and delegates execution to Docker. The CLI is architecturally isolated from the core EAC modules to remain lightweight and independently distributable.

## Code Health

### Tech Debt

- `internal/docker/hosting_test.go` contains TODO comments at lines 64 and 72 indicating incomplete test coverage for extension validation scenarios with proper mocks.

### Pain Points

- Multiple files exceed 300 lines: `cmd/run.go` (535 lines), `internal/conf/config-extensions.go` (430 lines), `internal/github/registry.go` (433 lines), `internal/conf/config.go` (390 lines), `cmd/update.go` (379 lines), `internal/command-parser/parser.go` (333 lines), `internal/conf/config-validation.go` (324 lines), `cmd/cleanup.go` (302 lines).
- No test files exist for `internal/cache` package (0 test files).
- No test files exist for `internal/extensions` package (0 test files).
- No test files exist for `internal/session` package (0 test files).
- No test files exist for `internal/tui` package (0 test files).

### Optimization Opportunities

- None identified.
