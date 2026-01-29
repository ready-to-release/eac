# R2R and EAC System Overview

## What is EAC?

> **EAC (Everything as Code)**

Is a comprehensive collection of commands designed to streamline **everything-as-code** workflows across your development lifecycle.

It helps teams codify and version-control all aspects of software development—configuration, contracts, architecture,
specifications, tests, security policies, and infrastructure.

**Key principle**: If it can be defined, it should be code. If it's code, it can be validated, versioned, and automated.

---

## Language Support

**Currently Supported:**

- ✅ **Go** - Full build, test, cross-compilation, and validation support
- ✅ **TypeScript/npm** - Build via npm/tsc, testing via Mocha and Cucumber
- ✅ **Docker** - Multi-platform container builds via buildx
- ✅ **MkDocs** - Documentation site generation

**Extensible Architecture:**

EAC uses a capability-based dispatch system where commands delegate to language-specific handlers and runners.

While the current implementation focuses on Go and TypeScript projects,
the architecture is designed to support additional languages through custom handlers.

**Other Languages:**

Languages like Python, Rust, and Java can use the `container` or `static` component types with custom build scripts,
but lack native build/test handlers.

See [Component Types Reference](./architecture/component-types.md) for detailed language support per component type.

---

## EAC as an R2R Extension

**EAC is delivered as an extension to R2R** (Ready to Release), an enterprise CLI framework for containerized workflow execution.

**R2R** provides:

- Cross-platform CLI (`r2r`, `r2r.exe`) for Windows, macOS, and Linux
- Docker-based extension execution with isolated, reproducible environments
- Git-aware working directory mounting
- Configurable extension management via `.r2r/r2r-cli.yml`

**The EAC extension** (`ext-eac:latest`) provides:

- **Hundreds of commands** spanning build, test, validation, security, AI, documentation, and release workflows
- **Three execution models**: Docker CLI (`r2r eac <command>`), MCP server for LLM tools, and fallback mode for debugging
- **Contract-driven architecture**: YAML contracts validated against JSON schemas to enforce consistency
- **Modular design**: Independent packages for core libraries, commands, AI integrations, and MCP servers

**Integration pattern**:

```text
Developer → R2R CLI → Docker Container (ext-eac:latest) → EAC Commands → Repository
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

The repository is organized as a **modular monorepo** with clearly defined module boundaries:

```text
cli/
├── .r2r/                       # R2R and EAC configuration
│   ├── cache/                  # Build cache
│   └── eac/                    # User configuration overrides
│       ├── repository.yml      # Module registry
│       └── books.yml           # Book configuration
│
├── contracts/                  # JSON schemas and system defaults
│   └── eac-core/0.1.0/         # Versioned schemas and defaults
│       ├── defaults/           # System default configurations
│       │   ├── component-types.yml
│       │   ├── tool-config.yml
│       │   └── ...
│       ├── repository.schema.json
│       └── ...
│
├── go/                         # Go source code
│   ├── eac/                    # EAC modules
│   │   ├── commands/           # eac-commands (hundreds of commands with AI integrations)
│   │   ├── core/               # eac-core (domain libraries)
│   │   ├── mcp/                # eac-mcp-commands (MCP server)
│   │   └── specs/              # eac-specs (BDD infrastructure)
│   └── r2r/                    # R2R CLI
│       └── cli/                # r2r-cli (CLI application)
│
├── specs/                      # Gherkin BDD specifications
│   ├── eac-commands/           # Specs for commands module
│   │   └── .design/            # Structurizr architecture
│   ├── eac-core/               # Specs for core module
│   └── ...
│
├── docs/                       # MkDocs documentation site
│   ├── reference/              # Technical reference
│   ├── how-to-guides/          # Task-based guides
│   ├── explanation/            # Conceptual docs
│   └── tutorials/              # Learning-oriented
│
├── out/                        # Generated artifacts (not in Git)
│   ├── build/                  # Build artifacts
│   ├── test/                   # Test results
│   ├── logs/                   # Execution logs
│   └── evidence/               # Compliance evidence
│
├── containers/                 # Docker container definitions
│   └── ext-eac/                # EAC extension Dockerfile
│
├── .github/                    # GitHub Actions workflows
│   └── workflows/              # CI/CD pipelines
│
├── go.work                     # Go workspace definition
└── mkdocs.yml                  # Documentation site config
```

### Module Categories

**Deployable Modules** - Independently built, versioned, and deployed:

- **r2r-cli** - Go CLI application with cross-platform executables
- **ext-eac** - Docker extension for R2R CLI
- **docs-site** - MkDocs documentation site (GitHub Pages)

**Supporting Modules** - Shared code and infrastructure:

- **eac-core** - Core domain libraries (contracts, repository, git)
- **eac-commands** - Command implementations (hundreds of commands) with integrated AI providers (Anthropic, OpenAI)
- **eac-specs** - BDD test infrastructure (Godog)
- **eac-mcp-commands** - MCP server for LLM tools

All modules defined in `.r2r/eac/repository.yml` with explicit dependencies and file ownership.

---

## In This Section

| Document                                                 | Description                                           |
| -------------------------------------------------------- | ----------------------------------------------------- |
| [Architecture](./architecture/index.md)                  | System architecture, components, and execution models |
| [Viewing Architecture](./architecture/viewing-diagrams.md) | How to view and work with C4 architecture diagrams    |
| [Modules](./architecture/modules.md)                     | Module system, dependencies, and build management     |
| [Component Types](./architecture/component-types.md)     | Language support and component type specifications    |
| [Contracts](./architecture/contracts.md)                 | YAML contracts, schemas, and validation system        |
| [Repository Layout](./architecture/repository-layout.md) | Directory structure and file organization             |
| [Creating Commands](./architecture/creating-commands.md) | Developer guide for extending EAC with new commands   |

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

**[Creating Commands](./architecture/creating-commands.md)**

Developer guide for contributing new commands to the EAC CLI including command structure, help system integration,
flag definitions, registry integration, and best practices.

## Related Documentation

- **[Command Reference](./commands/index.md)** - Complete reference for all EAC commands
- **[How-To Guides](../../how-to-guides/eac/index.md)** - Task-oriented guides for using EAC
- **[Decision Records](../repository/decision-records/index.md)** - Architectural decisions and rationale
