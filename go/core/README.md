# core

Pure domain logic and shared infrastructure for the EAC workspace. This module contains all
business rules, domain types, and foundational utilities with zero CLI dependencies.

Module moniker: `core` | Dependencies: `contracts/core`

## Package Index

| Package | Purpose |
| --- | --- |
| [adapters](./adapters/) | Adapter layer wrapping concrete domain types to satisfy port interfaces |
| [ai](./ai/) | Consolidated access to AI configuration, prompt templating, and model interaction |
| [ai/config](./ai/config/) | AI provider configuration loading with three-tier fallback |
| [ai/generation](./ai/generation/) | Structured AI generation with retry strategies and format validation |
| [ai/mock](./ai/mock/) | File-based mock AI executor and validator for acceptance tests |
| [ai/templates](./ai/templates/) | Go template-based prompt rendering with data binding |
| [cache](./cache/) | Two-dimensional taxonomy (Level x Type) for fine-grained cache control |
| [changedetect](./changedetect/) | Change detection logic using file hashing and mtime comparison |
| [changelog](./changelog/) | Conventional commit parsing and changelog generation |
| [config](./config/) | Central configuration loading from YAML files across workspace layers |
| [ctrf](./ctrf/) | Common Test Report Format types and utilities |
| [defaults](./defaults/) | Default values and path derivation for module contracts |
| [defaults/cmd/serialize](./defaults/cmd/serialize/) | CLI tool to serialize resolved module configs for comparison |
| [defaults/cmd/thin](./defaults/cmd/thin/) | CLI tool to remove redundant fields that match type defaults |
| [docsync](./docsync/) | CLI command documentation scanning for missing or orphaned doc files |
| [domain](./domain/) | Core domain types shared across the module system |
| [domain/modules](./domain/modules/) | Module contract type system and registry with workspace context |
| [domain/reports](./domain/reports/) | Report generation for CLI display commands |
| [domain/schema](./domain/schema/) | JSON Schema validation for contract YAML files |
| [environments](./environments/) | Environment definitions including artifact modes and build contexts |
| [evidence](./evidence/) | Security scan evidence file writing and SHA256 integrity verification |
| [execution](./execution/) | Cache verification interface for dependency-based work unit scheduling |
| [fileutil](./fileutil/) | Atomic file writes, locked writes, and platform-aware cleanup |
| [ghost](./ghost/) | Ghost-prefixed file and directory discovery for dark launching |
| [git](./git/) | Pure Go git operations using go-git for repository interaction |
| [github](./github/) | Abstractions for GitHub API interactions (workflows, releases, PRs) |
| [hash](./hash/) | Deterministic file content hashing with mtime-based fast-path caching |
| [logging](./logging/) | Unified structured logging with dual-sink output (console + rolling file) |
| [markdown](./markdown/) | Markdown validation utilities including structure checking and code blocks |
| [module-deps](./module-deps/) | Internal module dependency verification via build constraint checking |
| [output](./output/) | Output aggregation types for command result collection |
| [ownership](./ownership/) | File ownership resolution to modules and components by directory root |
| [paths](./paths/) | Centralized path constants and builder functions for the repository |
| [platform](./platform/) | Platform-specific abstractions for command execution and console output |
| [releasenotes](./releasenotes/) | RELEASE-NOTES.md parsing, validation, and generation |
| [repository](./repository/) | Repository dependency graph and module resolution |
| [repository/definitions](./repository/definitions/) | Definition file enumeration and YAML merging across directory trees |
| [repository/gomod](./repository/gomod/) | Go module dependency graph building and validation |
| [repository/reports](./repository/reports/) | File-module ownership statistics reporting |
| [resolver](./resolver/) | Component resolver for module-to-component mapping |
| [resource](./resource/) | Resource capacity calculation with standard formulas |
| [scheduling](./scheduling/) | Dependency-ordered scheduling with priority heap |
| [specs](./specs/) | BDD specification parsing, scenario export, and Godog test integration |
| [specs/export/formats](./specs/export/formats/) | Pluggable export formatters (CSV, JSON, Markdown) for manual test scenarios |
| [specs/gherkin](./specs/gherkin/) | Gherkin feature file parsing and scenario utilities |
| [testing](./testing/) | Test discovery, tag inference, suite selection, and BDD test isolation |
| [testing/mocks](./testing/mocks/) | Mock implementations of core port interfaces for unit testing |
| [testutil](./testutil/) | Test fixtures and helpers for unit testing |
| [tokensize](./tokensize/) | Token count estimation using character-based heuristics |
| [tool](./tool/) | Build bridge integrating the tool system with executors |
| [validation](./validation/) | Structured validation types, error codes, and formatting utilities |
| [validation/formats/gherkin](./validation/formats/gherkin/) | Gherkin feature file structural and semantic validation |
| [validation/formats/json](./validation/formats/json/) | JSON schema validation for generated content |
| [validation/formats/oscal](./validation/formats/oscal/) | OSCAL catalog and profile validation |
| [validation/formats/structurizr](./validation/formats/structurizr/) | Structurizr DSL validation with quick, Docker, and composite modes |
| [workspace](./workspace/) | Workspace root detection using prioritized detection chain |
| [workunit](./workunit/) | Work unit action types and aggregation for build/test/scan operations |

## Architecture Notes

The core module is the foundation layer of the EAC system, deliberately free of CLI
or infrastructure concerns. Packages follow a layered structure: `domain` defines shared
types consumed by all other packages, `config` and `workspace` provide environment
bootstrapping, and higher-level packages like `specs`, `testing`, and `evidence` implement
business rules on top of these primitives. The `adapters` package bridges domain types
to port interfaces, enabling the module to be consumed by CLI and adapter layers above.

Key dependency flows within the module:

- **Foundation**: `domain`, `paths`, `platform`, `logging` have no internal dependencies
- **Configuration**: `config` reads YAML and depends on `domain` and `paths`
- **Business Logic**: `specs`, `testing`, `evidence`, `releasenotes` build on config and domain
- **Orchestration Support**: `scheduling`, `execution`, `workunit` provide primitives
  consumed by the clibase orchestrator
- **Adapters**: `adapters` depends on most domain packages to satisfy port interfaces
