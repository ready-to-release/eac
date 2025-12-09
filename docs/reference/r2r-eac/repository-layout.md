# Repository Layout and Module Structure

{{ page_breadcrumb() }}

## Overview

This repository is organized as a **modular monorepo** with clearly defined module boundaries using the EAC (Everything as Code) system.

**Quick Reference**: For complete documentation on the R2R CLI and EAC system, see:

- [R2R and EAC Overview](./overview.md) - System overview with repository structure
- [Modules](./modules.md) - Module system and dependency management
- [Contracts](./contracts.md) - Module contracts and configuration

All modules are defined in `.r2r/eac/modules.yml`, which serves as the central contract for module ownership, dependencies, and build configuration.

## Repository Structure

```text
eac/
├── .claude/                    # Claude Code CLI configuration
│   ├── agents/                 # Custom agent definitions
│   ├── commands/               # Custom command definitions
│   └── hooks/                  # Git hooks for Claude integration
│
├── .github/                    # GitHub Actions workflows and configuration
│   ├── actions/                # Reusable workflow actions
│   └── workflows/              # CI/CD pipeline definitions
│
├── .r2r/                       # R2R and EAC configuration (Everything as Code)
│   ├── cache/                  # Build cache
│   └── eac/                    # EAC configuration files
│       ├── modules.yml         # Module contracts (central registry)
│       ├── module-types.yml    # Module type definitions
│       ├── environments.yml    # Environment configurations
│       ├── test-suites.yml     # Test suite definitions
│       ├── books.yml           # PDF book generation config
│       ├── ai-config.yml       # AI provider configuration
│       └── system-dependencies.yml
│
├── .vscode/                    # VSCode workspace configuration
│   └── extensions/             # Custom VSCode extensions
│
├── containers/                 # Docker container definitions
│   ├── ext-eac/                # R2R CLI extension for EAC
│   ├── mkdocs-site/            # Documentation site builder
│   └── mkdocs-pdf/             # PDF documentation builder
│
├── contracts/                  # JSON schemas for configuration files
│   ├── eac-core/0.1.0/         # Core EAC schemas
│   │   ├── modules.schema.json
│   │   ├── module-types.schema.json
│   │   ├── environments.schema.json
│   │   ├── handlers.schema.json
│   │   └── ...
│   └── r2r-cli/0.1.0/          # CLI-specific schemas
│
├── docs/                       # MkDocs documentation site source
│   ├── assets/                 # Images, diagrams, etc.
│   ├── explanation/            # Conceptual documentation
│   ├── how-to-guides/          # Task-based guides
│   ├── reference/              # Technical reference (this file)
│   └── tutorials/              # Learning-oriented tutorials
│
├── go/                         # Go source code
│   ├── eac/                    # EAC implementation
│   │   ├── commands/           # CLI commands library (eac-commands module) with integrated AI providers
│   │   ├── core/               # Core domain libraries (eac-core module)
│   │   ├── mcp/                # MCP server implementations
│   │   └── specs/              # Shared BDD test infrastructure
│   └── r2r/                    # R2R implementation
│       └── cli/                # R2R CLI application (r2r-cli module)
│
├── release/                    # Release notes and changelogs
│   ├── books/                  # Books module releases
│   ├── docs/                   # Docs module releases
│   ├── ext-eac/                # Extension releases
│   └── r2r-cli/                # CLI releases
│
├── scripts/                    # Cross-platform scripts
│   ├── pwsh/                   # PowerShell scripts (Windows)
│   └── sh/                     # Shell scripts (Linux/macOS)
│
├── specs/                      # Gherkin BDD specifications
│   ├── eac-commands/           # Specs for commands module
│   ├── eac-core/               # Specs for core module
│   ├── r2r-cli/                # Specs for CLI module
│   └── repository/             # Repository-level validation specs
│
├── templates/                  # Project templates
│   ├── design/                 # Design document templates
│   ├── docs/                   # Documentation templates
│   └── specs/                  # Specification templates
│
├── typescript/                 # TypeScript source code
│   └── vscode-ext-commit/      # VSCode Git extension
│
├── go.work                     # Go workspace definition
├── mkdocs.yml                  # MkDocs site configuration
└── CHANGELOG.md                # Repository-level changelog
```

## Module Organization

Modules are organized into two categories:

**Deployable Modules** - Independently built, versioned, and deployed:

- **r2r-cli** - Go CLI application with cross-platform executables
- **ext-eac** - Docker extension for R2R CLI
- **docs-site** - MkDocs documentation site (GitHub Pages)

**Supporting Modules** - Shared code and infrastructure:

- **eac-core** - Core domain libraries
- **eac-commands** - Command implementations (110+ commands) with integrated AI providers (Anthropic, OpenAI)
- **eac-specs** - BDD test infrastructure
- **eac-mcp-commands** - MCP server for LLM tools

For detailed information on module types, capabilities, and configuration, see [Modules Documentation](./modules.md).

## Module Configuration

All modules are defined in `.r2r/eac/modules.yml`. Each module specifies:

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
      root: go/eac/commands
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
```

For complete module configuration reference, see:

- [Modules Contract](./contracts.md#modules-contract) - Full field reference and validation rules
- [Module Types Contract](./contracts.md#module-types-contract) - Type templates and capabilities
- [Modules Documentation](./modules.md) - Module system and lifecycle

---

## Complete Documentation

For comprehensive information about the R2R and EAC system:

### Core Documentation

- [R2R and EAC Overview](./overview.md) - System overview with repository structure
- [Architecture](./architecture.md) - System architecture, components, and execution models
- [Contracts](./contracts.md) - Contract system and YAML configuration
- [Modules](./modules.md) - Module system and dependency management

### Related Topics

- [Trunk-Based Development](../../explanation/continuous-delivery/workflow/trunk-based-development.md)
- [Command Reference](../commands/index.md) - All 110+ EAC commands

### Configuration Files

- `.r2r/eac/modules.yml` - Module registry and dependencies
- `.r2r/eac/module-types.yml` - Module type templates
- `contracts/eac-core/0.1.0/` - JSON schemas for validation

{{ diataxis_footer() }}
