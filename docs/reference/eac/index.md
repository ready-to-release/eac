# EAC System Overview

## What is EAC?

> **EAC (Everything as Code)**

Is a comprehensive collection of commands designed to streamline **everything-as-code** workflows across your development lifecycle.

It helps teams codify and version-control all aspects of software development - configuration, contracts, architecture,
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

The repository is organized as a **modular monorepo** with 28 independently buildable modules across six groups (Core Framework, CLI Tools, Commands, Adapters, OCI Tools, Supporting).

See [Repository Layout](./architecture/repository-layout.md) for the full directory structure and [Modules Reference](./modules/index.md) for the complete module catalog.

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

## Related Documentation

- **[Command Reference](./commands/index.md)** - Complete reference for all EAC commands
- **[How-To Guides](../../how-to-guides/eac/index.md)** - Task-oriented guides for using EAC
- **[Decision Records](./decision-records/index.md)** - Architectural decisions and rationale
