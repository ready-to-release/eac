# clie-cli

Command-line interface for running containerized development extensions. Parses CLI arguments, loads layered YAML configuration, manages Docker container lifecycles, and provides interactive TUI feedback.

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

- Layered config: Base `clie-cli.yml` merged with `.local.yml`, `.personal.yml`, `.dev.yml` overrides
- Lazy Docker connectivity: `ContainerHost` defers Ping until first use via `sync.Once`
- Argument boundary: Parser separates Viper-processed flags from container pass-through args
- Embedded contracts: EBNF grammar and JSON schema loaded from `contracts/clie-cli` at init
- Atomic file writes: Cache files written to `.tmp` then renamed for crash safety

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| main.go | Entry point; filters spurious shell redirect arguments |
| cmd/ | Cobra command tree (root, run, init, install, version, etc.) |
| internal/cache/ | Registry and metadata caching with TTL expiration |
| internal/command-parser/ | EBNF-based CLI argument parsing and boundary detection |
| internal/conf/ | YAML config loading, merging, validation, repo root discovery |
| internal/docker/ | Docker client abstraction, image pull, container lifecycle |
| internal/envconsts/ | Isolated environment variable constant definitions |
| internal/extensions/ | Extension image installation orchestration |
| internal/github/ | GitHub CLI auth and GHCR registry tag listing |
| internal/logging/ | Leveled logger with configurable sinks and formatters |
| internal/session/ | Cross-platform shell session identification |
| internal/terminal/ | Platform-specific terminal size detection |
| internal/tui/ | Bubble Tea spinner model for image pull progress |
| internal/validator/ | Command and config validation against schema and rules |
| internal/version/ | Semver parsing, comparison, and build-flag metadata |

## Dependencies

- `contracts/clie-cli/0.1.0` -- Embedded EBNF grammar and JSON schema

## Role in System

The clie CLI is the primary user-facing binary that developers invoke to run containerized extensions (e.g., `clie run eac`). It reads `.clie/clie-cli.yml` configuration, resolves extension images from GHCR, and delegates execution to Docker. The CLI is architecturally isolated from the core EAC modules to remain lightweight and independently distributable.
