# Modules

## Overview

The EAC module system provides **independently buildable, testable units** with explicit contracts, dependency management, and file ownership.

**Key concepts**:

- **Modules** - Independently buildable units with explicit identity and dependencies
- **Components** - Typed components within modules (go, typescript, dockerfile, etc.)
- **Dependencies** - Explicit declaration enables topological build ordering
- **File Ownership** - Each file claimed by exactly one module via glob patterns

**Module definition** (`.eac/repository.yml`):

```yaml
modules:
  - moniker: core
    name: EAC Core Libraries
    depends_on: [logging]
    components:
      - type: go
        root: go/core
```

All modules are registered in `.eac/repository.yml` with dependencies, component types, and file ownership patterns.

---

## Module Organization

**28 modules** organized into focused groups:

### Core Framework

| Module        | Type       | Purpose                                  | Location     |
| ------------- | ---------- | ---------------------------------------- | ------------ |
| **core**      | go-library | Core utilities and libraries             | `go/core`    |
| **clibase**   | go-library | CLI framework (Cobra integration, flags) | `go/clibase` |
| **contracts** | yaml       | Contract schemas and definitions         | `contracts/` |

**See**: [core.md](core.md), [clibase.md](clibase.md)

### CLI Tools

| Module             | Type           | Purpose                            | Location             |
| ------------------ | -------------- | ---------------------------------- | -------------------- |
| **clie**           | go-cli         | CLIE CLI application               | `go/cli/clie`        |
| **eac**            | go-cli         | EAC CLI application                | `go/cli/eac`         |
| **eac-ext**        | clie-extension | EAC containerized extension        | `containers/eac-ext` |
| **eac-mcp-server** | go-mcp         | MCP server for AI tool integration | `go/mcp`             |

**See**: [eac.md](eac.md), [eac-ext.md](eac-ext.md), [eac-mcp-server.md](eac-mcp-server.md), [clie-eac-bundle.md](clie-eac-bundle.md)

### Commands

The **commands** module contains 7 command components:

- **base** - Command infrastructure and shared utilities
- **build** - Build automation commands
- **lint** - Code quality and linting
- **repository** - Repository management
- **scan** - Security scanning
- **test** - Test execution
- **update** - Update and maintenance

**Location**: `go/commands/`

**See**: [commands.md](commands.md)

### Adapters

The **adapters** module contains 17 adapter components:

**Test Frameworks**: gotest, godog, mocha, pytest, behave, reqnroll, cucumber

**Package Managers**: npm, pip, nuget, dotnet

**Infrastructure**: docker, gh, ai, eac

**Utilities**: tui

**Location**: `go/adapters/`

**See**: [adapters.md](adapters.md)

### Contracts

Contract schemas define interfaces between modules and external tools:

| Contract              | Purpose                            | Version |
| --------------------- | ---------------------------------- | ------- |
| **ai-provider**       | AI provider integration interface  | 0.1.0   |
| **clie**              | CLIE CLI framework configuration   | 0.1.0   |
| **container-runtime** | Container runtime interface        | 0.1.0   |
| **core**              | Core configuration and environment | 0.1.0   |
| **docs**              | Documentation generation contracts | 0.1.0   |
| **runner**            | Test runner interface              | 0.1.0   |
| **scanner**           | Security scanner interface         | 0.1.0   |
| **tui**               | Terminal UI component interface    | 0.1.0   |

**Location**: `contracts/`

**See**: [contracts.md](contracts.md)

### OCI Tools

Containerized development tools (12 container images):

| OCI Tool              | Purpose                  | Registry                                   |
| --------------------- | ------------------------ | ------------------------------------------ |
| **cgo-oci**           | C/Go cross-compilation   | ghcr.io/ready-to-release/cgo-oci           |
| **dotnet-oci**        | .NET SDK                 | ghcr.io/ready-to-release/dotnet-oci        |
| **drawio-oci**        | Diagram rendering        | ghcr.io/ready-to-release/drawio-oci        |
| **git-oci**           | Git tools                | ghcr.io/ready-to-release/git-oci           |
| **go-oci**            | Go SDK                   | ghcr.io/ready-to-release/go-oci            |
| **gource-oci**        | Repository visualization | ghcr.io/ready-to-release/gource-oci        |
| **mermaid-oci**       | Mermaid diagrams         | ghcr.io/ready-to-release/mermaid-oci       |
| **mkdocs-dev-oci**    | MkDocs dev server        | ghcr.io/ready-to-release/mkdocs-dev-oci    |
| **mkdocs-render-oci** | MkDocs rendering         | ghcr.io/ready-to-release/mkdocs-render-oci |
| **nginx-oci**         | NGINX server             | ghcr.io/ready-to-release/nginx-oci         |
| **pdf-cli-oci**       | PDF generation           | ghcr.io/ready-to-release/pdf-cli-oci       |
| **pdf-oci**           | PDF utilities            | ghcr.io/ready-to-release/pdf-oci           |

**Location**: `containers/`

**See**: [oci-tools.md](oci-tools.md)

### Supporting Modules

| Module              | Purpose                  | Location         |
| ------------------- | ------------------------ | ---------------- |
| **repository**      | Repository management    | `repository/`    |
| **docs**            | Documentation generation | `docs/`          |
| **templates**       | Project templates        | `templates/`     |
| **clie-eac-bundle** | Release bundle packaging | `bundle/`        |
| **cli-installers**  | CLIE installation        | `installers/`    |
| **vscode-commit**   | VS Code commit extension | `vscode-commit/` |
| **implicit-cli**    | Implicit CLI detection   | `implicit-cli/`  |

---

## Key Features

### Component Types

Modules contain typed components that determine build behavior:

- **go** - Go packages with build/test support
- **typescript** - TypeScript/JavaScript with npm builds
- **dockerfile** - Multi-platform container builds
- **book** - MkDocs documentation sites
- **pwsh** / **bash** - Script validation
- **gherkin** / **structurizr** - Specifications and architecture

**See**: [Component Types Reference](../architecture/component-kinds.md)

### Dependency Management

Modules declare dependencies via `depends_on` in `.eac/repository.yml`. Build system uses topological sort to determine build order and supports incremental builds based on change detection.

**Commands**:

```bash
eac show-dependencies          # View dependency graph
eac validate-dependencies      # Validate no cycles
eac get-changed-modules-ci     # Detect changed modules
```

**See**: [Architecture: Modules](../architecture/modules.md)

### File Ownership

Each file is claimed by exactly one module via glob patterns:

```yaml
components:
  - type: go
    root: go/core
    patterns:
      source: ["**/*.go"]
      tests: ["**/*_test.go"]
```

**Commands**:

```bash
eac show-files                 # Show file ownership
eac validate-module-files      # Validate no conflicts
```

---

## Module Lifecycle

1. **Discovery** - Load from `.eac/repository.yml`
2. **Build** - Topological build based on dependencies
3. **Test** - Execute test suites with tag filtering
4. **Validation** - Schema and dependency validation
5. **Release** - Version tagging and changelog generation

**See**: [Architecture: Modules](../architecture/modules.md) for detailed lifecycle documentation.

---

## Architecture Diagrams

Each module includes C4 architecture diagrams in Structurizr DSL format:

```bash
eac serve-design               # View all diagrams in browser
eac serve-design --module core # View specific module
```

**Design files**: `specs/{module}/.design/workspace.dsl`

**See**: [Viewing Architecture Diagrams](../architecture/viewing-diagrams.md)

---

## Related Documentation

- **[Architecture: Modules](../architecture/modules.md)** - Detailed module system architecture
- **[Architecture: Component Types](../architecture/component-kinds.md)** - Component type specifications
- **[Architecture: Contracts](../architecture/contracts.md)** - Contract system and schemas
- **[Commands Reference](../commands/index.md)** - CLI commands for working with modules
- **[Repository Layout](../architecture/repository-layout.md)** - Directory structure and organization
