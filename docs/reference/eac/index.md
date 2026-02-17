# EAC System Overview

## What is EAC?

> **EAC (Everything as Code)**

Is a comprehensive collection of commands designed to streamline **everything-as-code** workflows across your development lifecycle.

It helps teams codify and version-control all aspects of software development—configuration, contracts, architecture,
specifications, tests, security policies, and infrastructure.

**Key principle**: If it can be defined, it should be code. If it's code, it can be validated, versioned, and automated.

---

## Language Support

**Native Component Types:**

- ✅ **Go** - Component type with full build, test (gotest/godog), and cross-compilation support
- ✅ **TypeScript/JavaScript** - Component type with npm builds and test support (mocha, cucumber-js)
- ✅ **Docker** - Component type for multi-platform container builds via buildx
- ✅ **MkDocs** - Component type for documentation site and PDF generation
- ✅ **PowerShell** - Component type for script validation
- ✅ **Bash** - Component type for shell scripts

**Adapter Support:**

EAC includes adapters for package management and testing across multiple ecosystems:

- **Python** - pip adapter, pytest/behave test runners
- **.NET** - dotnet/nuget adapters, reqnroll test runner
- **Ruby** - cucumber test runner
- **Infrastructure** - GitHub CLI, AI providers, terminal UI

**Extensibility:**

Languages without native component types (Rust, Java, etc.) can use `dockerfile` components with custom build scripts.
The architecture supports adding new component types and adapters as needed.

See [Component Types Reference](./architecture/component-kinds.md) for complete component type documentation.

---

## EAC as an CLIE Extension

**EAC is delivered as an extension to CLIE**, an enterprise CLI framework for multi-platform containerized workflow execution.

**CLIE** provides:

- Cross-platform CLI (`clie`, `clie.exe`) for Windows, macOS, and Linux
- Docker-based extension execution with isolated, reproducible environments
- Git-aware working directory mounting
- Configurable extension management via `.clie/clie.yml`

**The EAC extension** (`eac-ext:latest`) provides:

- **Hundreds of commands** spanning build, test, validation, security, AI, documentation, and release workflows
- **Dual operation modes**: Run as a CLIE extension for containerized developer environments, or independently via the eac-mcp-server for AI tool integration using the Model Context Protocol (MCP)
- **Contract-driven architecture**: YAML contracts validated against JSON schemas to enforce consistency
- **Modular design**: Independent packages for core libraries, commands, AI integrations, and MCP servers

**Integration pattern**:

```text
Developer → CLIE CLI → Docker Container (eac-ext:latest) → EAC Commands → Repository
```

**Commands are organized by purpose**:

- Module management (`build`, `test`, `show-modules`)
- Validation (`validate-contracts`, `validate-dependencies`, `validate-specs`)
- Security scanning (`scan-vuln`, `scan-sast`, `scan-secrets`)
- AI workflows (`create-spec`, `create-design`, `create-commit-message`)
- Release management (`release-changelog`, `release-this`)
- Documentation (`serve-docs`, `serve-design`)

See the [Command Reference](./commands/index.md) for the complete list of available commands.

---

## Repository Structure

The repository is organized as a **modular monorepo** with 26+ independently buildable modules:

```text
eac/                            # Repository root
├── .eac/                       # Configuration directory
│   ├── repository.yml          # Module registry and dependencies
│   └── test-suites.yml         # Test suite definitions
│
├── contracts/                  # Versioned contract schemas (v0.1.0)
│   ├── ai-provider/            # AI provider contract
│   ├── clie/                   # CLIE CLI contract
│   ├── container-runtime/      # Container runtime contract
│   ├── core/                   # Core configuration contract
│   ├── docs/                   # Documentation contract
│   ├── runner/                 # Test runner contract
│   ├── scanner/                # Security scanner contract
│   └── tui/                    # Terminal UI contract
│
├── go/                         # Go source code
│   ├── cli/
│   │   ├── clie/               # CLIE CLI application
│   │   └── eac/                # EAC CLI application
│   ├── core/                   # Core utilities and libraries
│   ├── clibase/                # CLI framework
│   ├── mcp/                    # MCP server
│   ├── specs/                  # Specification infrastructure
│   │
│   ├── commands/               # Modular commands (7 modules)
│   │   ├── base/               # Command infrastructure
│   │   ├── build/              # Build commands
│   │   ├── lint/               # Lint commands
│   │   ├── repository/         # Repository commands
│   │   ├── scan/               # Scan commands
│   │   ├── test/               # Test commands
│   │   └── update/             # Update commands
│   │
│   └── adapters/               # Tool adapters (17 modules)
│       ├── ai/                 # AI provider adapter
│       ├── behave/             # Python BDD adapter
│       ├── cucumber/           # Ruby BDD adapter
│       ├── docker/             # Docker adapter
│       ├── dotnet/             # .NET adapter
│       ├── eac/                # EAC-to-EAC adapter
│       ├── gh/                 # GitHub adapter
│       ├── godog/              # Go BDD adapter
│       ├── gotest/             # Go test adapter
│       ├── mocha/              # JavaScript test adapter
│       ├── npm/                # npm adapter
│       ├── nuget/              # NuGet adapter
│       ├── pip/                # pip adapter
│       ├── pytest/             # pytest adapter
│       ├── reqnroll/           # .NET BDD adapter
│       └── tui/                # Terminal UI adapter
│
├── containers/                 # OCI tool images (12 containers)
│   ├── cgo-oci/                # C/Go cross-compilation
│   ├── dotnet-oci/             # .NET SDK
│   ├── drawio-oci/             # Diagram rendering
│   ├── eac-ext/                # EAC extension image
│   ├── git-oci/                # Git tools
│   ├── go-oci/                 # Go SDK
│   ├── gource-oci/             # Repository visualization
│   ├── mermaid-oci/            # Mermaid diagrams
│   ├── mkdocs-dev-oci/         # MkDocs dev server
│   ├── mkdocs-render-oci/      # MkDocs rendering
│   ├── nginx-oci/              # NGINX server
│   ├── pdf-cli-oci/            # PDF generation
│   └── pdf-oci/                # PDF utilities
│
├── specs/                      # BDD specifications
│   ├── {module}/.design/       # Structurizr C4 diagrams per module
│   └── {module}/features/      # Gherkin features per module
│
├── docs/                       # MkDocs documentation
│   ├── reference/              # Technical reference
│   ├── how-to-guides/          # Task-based guides
│   ├── explanation/            # Conceptual documentation
│   └── tutorials/              # Learning-oriented guides
│
├── out/                        # Build artifacts (gitignored)
│   ├── build/                  # Compiled binaries and images
│   ├── test/                   # Test results
│   └── evidence/               # Compliance evidence
│
├── .github/workflows/          # CI/CD pipelines
├── go.work                     # Go workspace definition
└── mkdocs.yml                  # Documentation configuration
```

### Module Organization

**28 modules** organized into focused groups. See [Modules Reference](./modules/index.md) for complete documentation.

**Core Framework** (3): core, clibase, contracts

**CLI Tools** (4): clie, eac, eac-ext, eac-mcp-server

**Framework Modules** (2):

- **commands** - Single module with 7 components: base, build, lint, repository, scan, test, update
- **adapters** - Single module with 17 components: ai, behave, cucumber, docker, dotnet, eac, gh, godog, gotest, mocha, npm, nuget, pip, pytest, reqnroll, tui, and more

**OCI Tools** (12): cgo-oci, dotnet-oci, drawio-oci, git-oci, go-oci, gource-oci, mermaid-oci, mkdocs-dev-oci, mkdocs-render-oci, nginx-oci, pdf-cli-oci, pdf-oci

**Supporting** (7): repository, docs, templates, clie-eac-bundle, cli-installers, vscode-commit, implicit-cli

All modules defined in `.eac/repository.yml` with explicit dependencies and file ownership.

---

## In This Section

| Document                                                   | Description                                           |
| ---------------------------------------------------------- | ----------------------------------------------------- |
| [Architecture](./architecture/index.md)                    | System architecture, components, and execution models |
| [Viewing Architecture](./architecture/viewing-diagrams.md) | How to view and work with C4 architecture diagrams    |
| [Modules](./architecture/modules.md)                       | Module system, dependencies, and build management     |
| [Component Types](./architecture/component-kinds.md)       | Language support and component type specifications    |
| [Contracts](./architecture/contracts.md)                   | YAML contracts, schemas, and validation system        |
| [Repository Layout](./architecture/repository-layout.md)   | Directory structure and file organization             |
| [Command Implementation](./architecture/command-implementation.md)   | Developer guide for implementing EAC commands   |

## Detailed Documentation

### System Architecture

**[Architecture](./architecture/index.md)**

Complete system architecture including component layers, execution models (Docker CLI, MCP Server, Fallback),
integration patterns, data flow, security architecture, and technology stack.

### Module System

**[Modules](./architecture/modules.md)**

Module system and dependency management including module registry,
dependency graph, file ownership patterns, module lifecycle (discovery, build, test, validation, release),
build system with caching, and Structurizr C4 architecture diagrams.

### Contract System

**[Contracts](./architecture/contracts.md)**

Contract-driven configuration including all 14 YAML contract files, modules contract (registry, dependencies, file ownership),
module types contract (templates, capabilities, artifacts), environments contract (L0-L4 testing pyramid),
validation system (schema, cross-reference), and configuration precedence.

### Repository Organization

**[Repository Layout](./architecture/repository-layout.md)**

Detailed repository structure with complete directory tree, module categories (deployable vs supporting),
module configuration examples, file ownership patterns, and navigation to related documentation.

### Extending EAC

**[Command Implementation](./architecture/command-implementation.md)**

Developer guide for implementing new commands in the EAC CLI including command structure, help system integration,
flag definitions, registry integration, and best practices.

## Related Documentation

- **[Command Reference](./commands/index.md)** - Complete reference for all EAC commands
- **[How-To Guides](../../how-to-guides/eac/index.md)** - Task-oriented guides for using EAC
- **[Decision Records](./decision-records/index.md)** - Architectural decisions and rationale
