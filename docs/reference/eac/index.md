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

## Running EAC

EAC can be run in two modes:

### Standalone CLI

Install the `eac` binary directly. No Docker required.

```text
Developer → EAC CLI → EAC Commands → Repository
```

### Via CLIE Extension Host (Optional)

EAC is also available as a CLIE extension (`eac-ext:latest`) for containerized execution. CLIE provides Docker-based isolation, reproducible environments, and git-aware volume mounting.

```text
Developer → CLIE CLI → Docker Container (eac-ext:latest) → EAC Commands → Repository
```

See [CLIE CLI Reference](../clie/index.md) for details on the extension host.

### Via MCP Server

LLM tools connect via the Model Context Protocol for AI-assisted workflows.

```text
LLM Tool → MCP Protocol → eac-mcp-commands → Repository
```

### Capabilities

- **Hundreds of commands** spanning build, test, validation, security, AI, documentation, and release workflows
- **Contract-driven architecture**: YAML contracts validated against JSON schemas to enforce consistency
- **Modular design**: Independent packages for core libraries, commands, AI integrations, and MCP servers

**Commands are organized by purpose**:

- Module management (`build`, `test`, `show-modules`)
- Validation (`validate-contracts`, `validate-dependencies`, `validate-specs`)
- Security scanning (`scan-vuln`, `scan-sast`, `scan-secrets`)
- AI workflows (`create-spec`, `create-design`, `get-commit-message`)
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

## Related Documentation

- **[Command Reference](./commands/index.md)** - Complete reference for all EAC commands
- **[How-To Guides](../../how-to-guides/eac/index.md)** - Task-oriented guides for using EAC
- **[Decision Records](./architecture/decisions/index.md)** - Architectural decisions and rationale
