<!-- EDITOR
# Editor: reference/repository-layout.md

## Soul

Explains monorepo module boundaries, distinguishing deployable vs supporting modules with contract-based configuration.

## Sections

1. Overview
2. Repository Structure
3. Module Categories
4. Module Configuration
5. Module Types
6. References
-->

# Repository Layout and Module Structure

## Overview

This repository is organized as a monorepo with clearly defined module boundaries. Understanding module structure is essential for creating accurate semantic commit messages and version increments.

All modules are defined in `.r2r/eac/modules.yml`, which serves as the central contract for module ownership, dependencies, and build configuration.

## Repository Structure

```
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
│       ├── books.yml           # PDF book generation config
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
│   │   ├── ai/                 # AI provider integrations (eac-ai module)
│   │   ├── commands/           # CLI commands library (eac-commands module)
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

## Module Categories

The repository distinguishes between two categories of modules:

### Deployable Modules

Independently built, versioned, and deployed artifacts. Each module is versioned and can produce build artifacts (executables, containers, sites, etc.).

**Examples from this repository:**

- **r2r-cli** - Go CLI application with cross-platform executables
- **docs** - MkDocs documentation site (GitHub Pages)
- **books** - PDF documentation with generated content
- **ext-eac** - Docker extension for R2R CLI
- **vscode-ext-commit** - VSCode extension for Git operations

Each deployable module typically has:

- Versioning scheme (SemVer or CalVer)
- Build configuration and artifacts
- CI/CD workflows (`.github/workflows/ci-{moniker}.yaml`)
- Changelog (in `release/{moniker}/CHANGELOG.md` or module root)

### Supporting Modules

Non-deployable modules containing shared code, configuration, or infrastructure. These modules don't produce independent artifacts but support other modules.

**Examples from this repository:**

- **eac-core** - Core domain libraries shared by multiple modules
- **eac-ai** - AI provider integrations library
- **eac-specs** - Shared BDD test infrastructure
- **r2r-config** - Repository configuration files
- **github** - GitHub Actions workflows and configuration
- **vscode** - VSCode workspace configuration
- **templates** - Project templates

Supporting modules:

- Provide shared functionality or configuration
- Are not independently versioned or deployed
- Are dependencies for deployable modules
- Trigger builds of dependent modules when changed

## Module Configuration

All modules are defined in `.r2r/eac/modules.yml`. Each module contract specifies:

### Core Fields

```yaml
modules:
  - moniker: eac-commands              # Unique identifier (kebab-case)
    name: Go Commands Library          # Human-readable name
    type: go                           # Module type (see Module Types section)
    description: Go library containing CLI command implementations

    versioning:                        # Optional: for deployable modules
      scheme: SemVer                   # SemVer or CalVer

    depends_on:                        # Module dependencies
      - eac-ai
      - eac-core
```

### File Ownership

Modules declare ownership of files and directories:

```yaml
    files:
      root: go/eac/commands            # Module root directory

      source:                          # Source code patterns
        - "**/*.go"

      tests:                           # Test file patterns
        - "**/*_test.go"

      config:                          # Configuration files
        - go.mod
        - go.sum

      assets:                          # Documentation and assets
        - "**/README.md"
        - "**/assets/**"

      exclude:                         # Exclusion patterns
        - "ext/**"

      changelog: CHANGELOG.md          # Changelog file

      workflows:                       # GitHub Actions workflows
        ci: .github/workflows/ci-eac-commands.yaml
        release: ""                    # Empty if no release workflow

      repo:                            # Repository-level files
        specs:                         # BDD specifications
          - "specs/eac-commands/**"
        test_impl: "{test_impl_root}/eac-commands"  # Test implementation
        design: "{specs_root}/eac-commands/.design" # Design workspace
        other:                         # Other related files
          - .github/workflows/ci-eac-commands.yaml
```

### Metadata

Modules can include custom metadata for module-specific configuration. This is an optional escape hatch for special cases:

```yaml
    metadata:
      custom-property: "value"
```

Note: For artifact naming, prefer using literal patterns in `build.artifacts` (e.g., `pattern: "r2r-linux-amd64"`) rather than metadata overrides.

## Module Types

Module types are defined in `.r2r/eac/module-types.yml` and provide defaults and build configuration for modules of that type.

### Module Types

The unified module type system uses four base types. Behavior is driven by per-module artifact definitions in `modules.yml`:

- **go** - Go module (library, executable, or test)
  - Capabilities: `go_module`
  - Build artifacts driven by module config: none (library), executables (CLI), test results
  - Examples: `eac-core`, `eac-commands`, `r2r-cli`, `eac-specs`

- **container** - Docker container module
  - Capabilities: `docker_build`
  - Build artifacts: Container images
  - Examples: `docs`, `ext-eac`

- **typescript** - TypeScript/npm module
  - Capabilities: `npm_package`, `typescript`
  - Build artifacts: VSIX package, compiled JS
  - Example: `vscode-ext-commit`

- **static** - Static files module (no build step)
  - Capabilities: none
  - No build artifacts - file ownership only
  - Examples: `r2r-config`, `github`, `templates`

### Module Type Schema

Each module type defines:

```yaml
types:
  - name: go
    description: "Go module (library, executable, or test - driven by module artifacts)"

    capabilities:                      # Module capabilities
      - go_module

    defaults:                          # Default values for modules of this type
      files:
        source: ["**/*.go", "**/*.go.txt"]
        tests: ["**/*_test.go"]
        config: ["go.mod", "go.sum"]
        changelog: CHANGELOG.md
      repo:
        specs: ["{specs_root}/{moniker}/**"]
        test_impl: "{test_impl_root}/{moniker}"
        design: "{specs_root}/{moniker}/.design"
```

Pattern variables supported:

- `{moniker}` - Module moniker
- `{os}` - Target OS (linux, windows, darwin)
- `{arch}` - Target architecture (amd64, arm64)
- `{ext}` - Executable extension (.exe for Windows, empty otherwise)
- `{root}` - Module root directory
- `{specs_root}` - Specs root (from global defaults)
- `{test_impl_root}` - Test implementation root (from global defaults)

---

## References

- [Trunk-Based Development](../explanation/continuous-delivery/workflow/trunk-based-development.md)
- Module Configuration: `.r2r/eac/modules.yml`
- Module Types: `.r2r/eac/module-types.yml`
- JSON Schemas: `contracts/eac-core/0.1.0/`
