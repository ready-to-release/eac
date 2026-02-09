# Changelog

All notable changes to **eac** will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

<!-- Release Highlights (first release)

  TUI / User Interface:
  - MVC refactoring: view.go reduced from 3,719 to 235 lines; rendering split into render/ subpackage + 6 view files by concern
  - Widget catalog: central registry of 11 singleton widgets + 1 tab template for consistent UI components
  - Pipeline latency fix: batched line delivery (50), channel buffers 500, cached layout metrics for ~10x throughput
  - Non-blocking boot: TUI renders instantly (<10ms); config loads in background (200-800ms)
  - Progressive boot: BootState enum with unified render path; init and active phases share one layout
  - Lock visualization in TUI status header showing resource contention and lock acquisition states
  - ascii and skip-tui-delay flags for non-Unicode rendering and immediate exit
  - Tool lamp fix: moniker parsing uses index 3 for tool name instead of last segment

  Build System & Caching:
  - Cache staleness fix: pre-compute module input hashes once before parallel builds; changed staleness from ANY-mismatch to ALL-mismatch
  - Post-build copy: single source of truth for metadata file exclusion; fallback output dir detection
  - AfterExecute manifest assertion: shared helper validates manifests exist for all successful UoWs
  - Artifact matrix system with YAML-based definitions and container auto-discovery for -oci suffix
  - Granular test caching with test-set-aware invalidation rules
  - Build dependency system for module-to-module build ordering
  - Dry-run mode for build operations to preview changes without execution
  - Build agent tracking and diagnostics for better visibility
  - Cross-compiled binary path handling for multi-platform builds
  - Component weight calculation improvements for build scheduling
  - Structurizr diagram caching

  Test Framework:
  - Test manifest removal: replaced legacy test.manifest.json with UoW-based testview aggregation package
  - BDD in-process dispatcher: CLI commands execute in-process via registry (~97% test time reduction)
  - Test tag traceability: tags threaded through full UoW pipeline for risk control traceability
  - Gorunner context propagation: worker context propagated to go test/generate, fixing timeout failures
  - Legacy manifest cleanup: all legacy test/scan/lint.manifest.json references removed
  - Manual test export/import with gherkin parsing and testing system integration
  - Test results reporting commands for displaying and analyzing outcomes
  - Test suite configuration defaults system
  - Sequential test support alongside parallel test execution
  - Decentralized Gherkin step implementations into independent spec modules
  - Comprehensive testutil package for test infrastructure
  - Go generate support in test runners

  Tool System & Containers:
  - Subprocess routing: all subprocesses routed through tool system with process group management and context propagation
  - Executor mode: global override (auto/container/system) with per-tool bindings
  - Container credential forwarding: three-layer injection (global credentials, per-tool host-env, os.Getenv)
  - Helper packages: gitexec, ghexec, goexec for tool-routed CLI execution
  - Pluggable tool system with configurable execution modes
  - Neutral containerruntime provider with Docker-only discovery
  - Consolidated Docker utilities into shared framework
  - Architecture test prevents direct exec.Command regression

  Configuration & Templates:
  - Artifact matrix and container discovery: auto-discover -oci containers, artifact matrix expansion
  - Timeout configuration: centralized system with YAML defaults + Go fallback, domain-based categories (Docker, file lock, CI, TUI, capacity, evidence)
  - Book templates system with consolidated configuration defaults
  - Generated default config files during init with calculated platform-specific defaults
  - Separated linting and build configuration into provider-based architecture
  - Unified repository and module configuration into single .eac/repository.yml system
  - Migrated configuration directory from .clie/eac to .eac
  - Centralized contract system with contract-based validation and loading
  - Handler flags configuration system
  - Module file path specifications for explicit component roots
  - MkDocs config migrated to template-based architecture

  Scheduling & Execution:
  - Dependency failure propagation: CascadeFail removes transitive dependents immediately from scheduler
  - Unified work unit system with granular state management
  - Resource capacity limiting with --roof flag
  - Turbo mode support for PDF export concurrency control
  - Tracked locks across all frameworks for TUI visualization
  - Optimized parallelization and resource management

  New / Enhanced Commands:
  - eac update go-sums: run go mod tidy across all workspace modules
  - Enhanced show test-summary / show test-results with testview aggregation
  - Enhanced get test-results with UoW-based loading
  - Evidence artifact collection commands for CI dependencies
  - Repository scanner for resource scaling and capacity analysis
  - Git tag validation with diagnostic error messages
  - DrawIO diagram editing support with container-based tooling
  - CI run ID lookup command
  - Release management viewing commands (show/get subcommands)
  - Evidence workflow and release notes generation
  - Books support commands for PDF documentation generation
  - OSCAL catalog validation support
  - Consolidated scan command with skip filters for security scanning
  - Per-module CI change detection for targeted builds
  - Risk assessment and control creation commands with risk control catalog
  - Unused-steps command for specification coverage analysis
  - Comprehensive security scanning suite with Windows path support
  - Design command with AI-powered specification generation
  - Templates command system for docs/reports template installation
  - Auto-commit capability in commit message command

  Architecture:
  - Multi-module restructure with extracted adapters and reorganized specs
  - Container naming standardization and core resolver enhancement
  - Reorganized modules by dependency layer (L0, L1, L2, L3)
  - Platform-specific file locking for workunit (Windows + Unix)
  - Restructured AI generation to JSON-first two-phase architecture
  - Consolidated AI configuration framework into unified adapter
  - Migrated risk assessment reporting to template-based architecture
  - Separated books module from docs module
  - Migrated to go-git library for mockable repository operations
  - Monorepo migration from src/ to go/cli/eac
  - Centralized logging with dual-output and debug support

  CI/CD & Release:
  - Unified release and CI orchestration system
  - CI dispatch layering by artifact dependencies
  - Pipeline restructured to eliminate sequential blocking
  - CI inheritance with SHA passthrough
  - Self-healing releases with automatic retry logic
  - Concurrency controls for GitHub workflows
  - Release type system with workflow validation
  - Release bundle workflow and commands
  - macOS testing support in CI matrix
  - Module-owned CI workflows with centralized configuration
  - Release automation with changelog system and versioning

  Documentation:
  - Comprehensive security policy
  - Release workflow documentation with mermaid diagrams
  - Architectural decision records (ADRs)
  - Books infrastructure for PDF documentation generation
  - Documentation restructured following Diataxis framework (tutorials, how-to, explanation, reference)
  - Everything-as-code philosophy documentation
  - Testing taxonomy reference
  - EAC CLI installer guide
  - Manual testing guide and reference docs

-->

[Unreleased]: https://github.com/ready-to-release/eac/commits/main/go/cli/eac
