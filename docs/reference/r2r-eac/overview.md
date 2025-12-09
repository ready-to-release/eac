# R2R and EAC System Overview

{{ page_breadcrumb() }}

## What is EAC?

**EAC (Everything as Code)** is a comprehensive collection of commands designed to streamline **everything-as-code** workflows across your development lifecycle. It helps teams codify and version-control all aspects of software development—configuration, contracts, architecture, specifications, tests, security policies, and infrastructure.

**Key principle**: If it can be defined, it should be code. If it's code, it can be validated, versioned, and automated.

**Language support**: While the current implementation is built in Go and optimized for Go projects, the everything-as-code philosophy and command structure are designed to be language-agnostic. Future versions may support additional languages and ecosystems.

---

## EAC as an R2R Extension

**EAC is delivered as an extension to R2R** (Ready to Release), an enterprise CLI framework for containerized workflow execution.

**R2R** provides:

- Cross-platform CLI (`r2r`, `r2r.exe`) for Windows, macOS, and Linux
- Docker-based extension execution with isolated, reproducible environments
- Git-aware working directory mounting
- Configurable extension management via `.r2r/r2r-cli.yml`

**The EAC extension** (`ext-eac:latest`) provides:

- **105+ commands** spanning build, test, validation, security, AI, documentation, and release workflows
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

See the [Command Reference](../commands/index.md) for the complete list of available commands.

---

## Repository Structure

The repository is organized as a **modular monorepo** with clearly defined module boundaries:

```text
cli/
├── .r2r/                       # R2R and EAC configuration
│   ├── cache/                  # Build cache
│   └── eac/                    # EAC contracts (14 YAML files)
│       ├── modules.yml         # Module registry
│       ├── module-types.yml    # Module type definitions
│       ├── environments.yml    # Test environments (L0-L4)
│       ├── test-suites.yml     # Test suite definitions
│       └── ...
│
├── contracts/                  # JSON schemas for validation
│   └── eac-core/0.1.0/         # Versioned schemas
│       ├── modules.schema.json
│       ├── module-types.schema.json
│       └── ...
│
├── go/                         # Go source code
│   ├── eac/                    # EAC modules
│   │   ├── ai/                 # eac-ai (AI integrations)
│   │   ├── commands/           # eac-commands (105+ commands)
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
- **eac-ai** - AI provider integrations (Anthropic, OpenAI)
- **eac-commands** - Command implementations (105+ commands)
- **eac-specs** - BDD test infrastructure (Godog)
- **eac-mcp-commands** - MCP server for LLM tools

All modules defined in `.r2r/eac/modules.yml` with explicit dependencies and file ownership.

---

## Detailed Documentation

Explore these topics for comprehensive information about the R2R and EAC system:

### [Architecture](./architecture.md)

Complete system architecture including:

- Component layers (CLI, Extension, Core, MCP, Repository)
- Execution models (Docker CLI, MCP Server, Fallback CLI)
- Integration patterns (Docker orchestration, function calls, stdio/JSON-RPC, shared files)
- Data flow and security architecture
- Technology stack and performance considerations

### [Contracts](./contracts.md)

Contract system and YAML configuration including:

- All 14 contract files (modules, module types, environments, test suites, etc.)
- Modules contract (module registry, dependencies, file ownership)
- Module types contract (type templates, capabilities, build artifacts)
- Environments contract (testing pyramid L0-L4)
- Validation system (schema, cross-reference, file ownership, hierarchy)
- Configuration precedence and IDE integration

### [Modules](./modules.md)

Module system and dependency management including:

- Module registry and types
- Dependency graph and build order
- File ownership and glob patterns
- Module lifecycle (discovery, build, test, validation, release)
- Build system (artifacts, cache, incremental builds)
- Module designs (Structurizr C4 diagrams)
- Working with modules and best practices

### [Repository Layout](./repository-layout.md)

Detailed repository structure and organization including:

- Complete directory tree with descriptions
- Module categories (deployable vs supporting)
- Module configuration examples
- File ownership patterns
- Navigation to related documentation

### Command Reference

- [All Commands](../commands/index.md) - Complete command reference organized by category

{{ diataxis_footer() }}
