# Changelog

All notable changes to **eac** will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2026-02-20

### Added

- TUI with MVC architecture: view.go reduced from 3,719 to 235 lines; rendering split into render/ subpackage with 6 view files by concern
- Widget catalog: central registry of 11 singleton widgets + 1 tab template for consistent UI components
- Non-blocking boot: TUI renders instantly (<10ms); config loads in background (200-800ms)
- Progressive boot with BootState enum and unified render path
- Lock visualization in TUI status header showing resource contention and lock acquisition states
- BDD in-process dispatcher: CLI commands execute in-process via registry (~97% test time reduction)
- Test tag traceability threaded through full UoW pipeline for risk control traceability
- Manual test export/import with gherkin parsing and testing system integration
- Pluggable tool system with configurable execution modes (auto/container/system)
- Container credential forwarding: three-layer injection (global credentials, per-tool host-env, os.Getenv)
- Helper packages: gitexec, ghexec, goexec for tool-routed CLI execution
- Artifact matrix system with YAML-based definitions and container auto-discovery
- Granular test caching with test-set-aware invalidation rules
- Build dependency system for module-to-module build ordering
- Dependency failure propagation: CascadeFail removes transitive dependents from scheduler
- Unified work unit system with granular state management
- Resource capacity limiting with --roof flag
- Evidence artifact collection commands for CI dependencies
- DrawIO diagram editing support with container-based tooling
- Release management viewing commands (show/get subcommands)
- Books support commands for PDF documentation generation
- OSCAL catalog validation support
- Risk assessment and control creation commands with risk control catalog
- Design command with AI-powered specification generation
- Templates command system for docs/reports template installation
- Comprehensive security scanning suite with Windows path support

### Changed

- Pipeline latency fix: batched line delivery (50), channel buffers 500, cached layout metrics for ~10x throughput
- Cache staleness fix: pre-compute module input hashes once before parallel builds; changed staleness from ANY-mismatch to ALL-mismatch
- Test manifest removal: replaced legacy test.manifest.json with UoW-based testview aggregation
- Subprocess routing: all subprocesses routed through tool system with process group management
- Timeout configuration: centralized system with YAML defaults + Go fallback, domain-based categories
- Unified repository and module configuration into single .eac/repository.yml system
- Migrated configuration directory from .clie/eac to .eac
- Centralized contract system with contract-based validation and loading
- Multi-module restructure with extracted adapters and reorganized specs
- Reorganized modules by dependency layer (L0, L1, L2, L3)
- Platform-specific file locking for workunit (Windows + Unix)
- Restructured AI generation to JSON-first two-phase architecture
- Monorepo migration from src/ to go/cli/eac
- Unified release and CI orchestration system with CI dispatch layering
- Centralized logging with dual-output and debug support

[Unreleased]: https://github.com/ready-to-release/eac/compare/eac/0.0.1...HEAD
[0.0.1]: https://github.com/ready-to-release/eac/releases/tag/eac/0.0.1
