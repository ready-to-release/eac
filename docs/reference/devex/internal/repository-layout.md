# Repository Layout and Module Structure

## Overview

This repository is organized as a **modular monorepo** with clearly defined module boundaries using the EAC (Everything as Code) system.

**Quick Reference**: For complete documentation on the CLIE CLI and EAC system, see:

- [EAC Overview](../../eac/index.md) - System overview with repository structure
- [Modules](../../eac/modules/index.md) - Module system and dependency management
- [Contracts](../../eac/architecture/contracts.md) - Module contracts and configuration

All modules are defined in `.eac/repository.yml`, which serves as the central contract for module ownership,
dependencies, and build configuration.

## Repository Structure

```text
eac/
├── .claude/                          # Claude Code CLI configuration
│   ├── agents/                       # Custom agent definitions
│   ├── commands/                     # Custom command definitions
│   ├── hooks/                        # Git hooks for Claude integration
|   └── ...
│
├── .github/                          # GitHub Actions workflows and configuration
│   ├── actions/                      # Reusable workflow actions
│   ├── workflows/                    # CI/CD pipeline definitions
|   └── ...
│
├── .clie/                             # CLIE and extension configuration
│   ├── cache/                        # Build cache
│   └── eac/                          # User configuration overrides
│       ├── repository.yml            # Module contracts (central registry)
│       └── books.yml                 # PDF book generation config
│
├── .vscode/                          # VSCode workspace configuration
│   └── extensions/                   # Custom VSCode extensions
│
├── containers/                       # Docker container definitions
│   ├── eac-ext/                      # CLIE CLI extension for EAC
│   ├── mkdocs-site/                  # Documentation site builder
│   ├── mkdocs-pdf/                   # PDF documentation builder
│   └── ...
│
├── contracts/                        # JSON schemas and system defaults
│   ├── core/0.1.0/                   # Core EAC schemas and defaults
│   │   ├── schemas/
│   │   │   └── defaults/             # System default configurations
│   │   │       ├── blueprints.yml    # Component kind definitions
│   │   │       ├── tool-config.yml   # Tool definitions and resources
│   │   │       ├── environments.yml  # Environment configurations
│   │   │       ├── test-suites.yml   # Test suite definitions
│   │   │       └── ...               # Other system defaults
│   │   ├── repository.schema.json
│   │   └── ...
│   └── clie/0.1.0/                # CLI-specific schemas
│
├── docs/                             # MkDocs documentation site source
│   ├── assets/                       # Images, diagrams, etc.
│   ├── explanation/                  # Conceptual documentation
│   ├── how-to-guides/                # Task-based guides
│   ├── reference/                    # Technical reference (this file)
│   └── tutorials/                    # Learning-oriented tutorials
│
├── go/                               # Go source code
│   ├── adapters/                     # Adapters (AI, Docker, TUI)
│   ├── cli/                          # CLI implementations
│   │   └── eac/                      # EAC CLI commands library (eac-commands module) with integrated AI providers
│   ├── core/                         # Core domain libraries (eac-core module)
│   ├── godog/                        # Shared BDD test infrastructure
│   ├── specs/                        # Specs implementations
│   └── clie/                          # CLIE implementation
│       └── cli/                      # CLIE CLI application (clie module)
│
├── release/                          # Release notes and changelogs
│   ├── books/                        # Books module releases
│   ├── docs/                         # Docs module releases
│   ├── eac-ext/                      # Extension releases
│   └── clie/                      # CLI releases
│
├── scripts/                          # Cross-platform scripts
│   ├── pwsh/                         # PowerShell scripts (Windows)
│   └── sh/                           # Shell scripts (Linux/macOS)
│
├── specs/                            # Gherkin BDD specifications
│   ├── eac-commands/                 # Specs for commands module
│   ├── eac-core/                     # Specs for core module
│   ├── clie/                      # Specs for CLI module
│   └── repository/                   # Repository-level validation specs
│
├── templates/                        # Project templates
│   ├── design/                       # Design document templates
│   ├── docs/                         # Documentation templates
│   └── specs/                        # Specification templates
│
├── typescript/                       # TypeScript source code
│   └── vscode-commit/                # VSCode Git extension
│
├── go.work                           # Go workspace definition
├── mkdocs.yml                        # MkDocs site configuration
└── CHANGELOG.md                      # Repository-level changelog
```

## Module Organization

Modules are organized into two categories:

**Deployable Modules** - Independently built, versioned, and deployed:

- **clie** - Go CLI application with cross-platform executables
- **eac-ext** - Docker extension for CLIE CLI
- **docs-site** - MkDocs documentation site (GitHub Pages)

**Supporting Modules** - Shared code and infrastructure:

- **eac-core** - Core domain libraries
- **eac-commands** - Command implementations (hundreds of commands) with integrated AI providers (Anthropic, OpenAI)
- **eac-specs** - BDD test infrastructure
- **eac-mcp-commands** - MCP server for LLM tools

For detailed information on module types, capabilities, and configuration, see [Modules Documentation](../../eac/modules/index.md).

## Module Configuration

All modules are defined in `.eac/repository.yml`. Each module specifies:

- **Moniker** - Unique identifier (e.g., `eac-commands`)
- **Type** - Module type reference (e.g., `go-library`)
- **Dependencies** - Module dependencies (e.g., `depends_on: [eac-core]`)
- **File Ownership** - Glob patterns defining owned files

**Example**:

```yaml
modules:
  - moniker: eac-commands
    name: Go Commands Library
    type: go-commands
    depends_on: [eac-core]
    files:
      root: go/cli/eac
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
```

For complete module configuration reference, see:

- [Modules Contract](../../eac/architecture/contracts.md#modules-contract) - Full field reference and validation rules
- [Component Kinds Contract](../../eac/architecture/contracts.md#blueprints-contract) - Kind templates and capabilities
- [Modules Documentation](../../eac/modules/index.md) - Module system and lifecycle

---

## Complete Documentation

For comprehensive information about the EAC and CLIE system:

### Core Documentation

- [EAC Overview](../../eac/index.md) - System overview with repository structure
- [Architecture](../../eac/architecture/index.md) - System architecture, components, and execution models
- [Contracts](../../eac/architecture/contracts.md) - Contract system and YAML configuration
- [Modules](../../eac/modules/index.md) - Module system and dependency management

### Related Topics

- [Trunk-Based Development](../../../explanation/continuous-delivery/workflow/trunk-based-development.md)
- [Command Reference](../../eac/commands/index.md) - All the hundreds of EAC commands

### Configuration Files

- `.eac/repository.yml` - Module registry and dependencies (user config)
- `contracts/core/0.1.0/schemas/defaults/` - System default configurations
- `contracts/core/0.1.0/schemas/defaults/blueprints.yml` - Component kind definitions
- `contracts/core/0.1.0/` - JSON schemas for validation
