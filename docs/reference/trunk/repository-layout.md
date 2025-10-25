# Repository Layout and Module Structure

## Overview

This repository is organized as a monorepo with clearly defined module boundaries. Understanding module structure is essential for creating accurate semantic commit messages and version increments.

## Deployable Units vs Supporting Modules

The repository distinguishes between two categories:

### Deployable Units
Independently built, versioned, and deployed artifacts. Each has a detailed contract in `contracts/deployable-units/0.1.0/{moniker}.yml` defining:
- Build and deployment configuration
- Versioning strategy and current version
- Runtime dependencies and environment
- Integration points and APIs

**Current deployable units**:
- `src-mcp-pwsh`, `src-mcp-docs`, `src-mcp-github`, `src-mcp-vscode` (MCP servers)
- `vscode-ext-claude-commit` (VSCode extension)

### Supporting Modules
Non-deployable modules that support the system:
- Infrastructure (build/deploy tooling)
- Documentation (user guides)
- Configuration (IDE and tool settings)
- Contracts (definitions and schemas)

## Module Types

### 1. MCP Server Modules
Go-based Model Context Protocol servers providing integration capabilities.

**Characteristics**:
- Language: Go
- Versioning: Independent semantic versioning via git tags
- Tag format: `{module-name}/v{version}`
- Structure: `src/mcp/{name}/`

### 2. Extension Modules
IDE extensions and integrations.

**Characteristics**:
- Language: TypeScript/JavaScript
- Versioning: Independent semantic versioning via package.json
- Tag format: `{module-name}/v{version}`
- Structure: `.vscode/extensions/{name}/`

### 3. Infrastructure Modules
Build, deployment, and automation tooling.

**Characteristics**:
- Language: Shell/Docker/YAML
- Versioning: Not independently versioned
- Structure: `automation/`, `containers/`

### 4. Documentation Modules
User-facing and reference documentation.

**Characteristics**:
- Language: Markdown
- Versioning: Not independently versioned
- Structure: `docs/`, root `.md` files

### 5. Configuration Modules
IDE, tool, and repository configuration.

**Characteristics**:
- Language: JSON/YAML/Shell
- Versioning: Not independently versioned
- Structure: `.claude/`, `.vscode/`, root config files

### 6. Contract Modules
Versioned contract definitions and schemas.

**Characteristics**:
- Language: YAML/JSON
- Versioning: Semantic versioning in path structure
- Structure: `contracts/{name}/{version}/`

## Module Registry

### MCP Servers

| Module | Prefix | Location | Description |
|--------|--------|----------|-------------|
| `src-mcp-pwsh` | `src-mcp-pwsh:` | `src/mcp/pwsh/` | PowerShell command execution |
| `src-mcp-docs` | `src-mcp-docs:` | `src/mcp/docs/` | Documentation management |
| `src-mcp-github` | `src-mcp-github:` | `src/mcp/github/` | GitHub API integration |
| `src-mcp-vscode` | `src-mcp-vscode:` | `src/mcp/vscode/` | VSCode integration |

### Extensions

| Module | Prefix | Location | Description |
|--------|--------|----------|-------------|
| `vscode-ext-claude-commit` | `vscode-ext-claude-commit:` | `.vscode/extensions/claude-mcp-vscode/` | MCP VSCode extension UI |

### Infrastructure

| Module | Prefix | Location | Description |
|--------|--------|----------|-------------|
| `infrastructure` | `infra:` | `automation/`, `containers/` | Build and deployment automation |

### Documentation

| Module | Prefix | Location | Description |
|--------|--------|----------|-------------|
| `documentation` | `docs:` | `docs/`, `*.md` (root) | User guides and references |

### Configuration

| Module | Prefix | Location | Description |
|--------|--------|----------|-------------|
| `claude-config` | `config:` | `.claude/` | Claude Code configuration |
| `vscode-config` | `config:` | `.vscode/` (non-extension files) | VSCode workspace settings |
| `repository-config` | `config:` | Root config files | Repository-wide settings |

### Contracts

| Module | Prefix | Location | Description |
|--------|--------|----------|-------------|
| `contracts-repository` | `contracts:` | `contracts/repository/{version}/` | Repository structure contracts |

## Module Identification Rules

### Path-Based Identification

1. **Exact match**: Check if path starts with known module location
2. **Parent directory**: Walk up to find nearest module root
3. **File type patterns**:
   - `src/mcp/*/**/*.go` → MCP server module
   - `.vscode/extensions/**/*.ts` → Extension module
   - `docs/**/*.md` → Documentation module
   - `.claude/**/*` → Claude config module
   - `contracts/{name}/{version}/**` → Contract module

### Commit Prefix Selection

**Single module change**: Use module prefix
```
src-mcp-pwsh: fix: handle empty command output
```

**Multiple modules (preferred)**: Separate commits per module
```
Commit 1: src-mcp-vscode: feat: add new tool
Commit 2: vscode-ext-claude-commit: feat: integrate new tool
```

**Multiple modules (alternative)**: Combined prefix
```
src-mcp-vscode,vscode-ext-claude-commit: feat: add complete feature
```

## Module Dependencies

```
vscode-ext-claude-commit
├── depends on: src-mcp-pwsh, src-mcp-docs, src-mcp-github, src-mcp-vscode
└── used by: end users

infrastructure
└── used by: all modules (build/deploy)

*-config
└── used by: all modules (configuration)

contracts-repository
└── defines: all module boundaries
```

## Version Management

**Independently Versioned Modules**:
- All MCP servers (git tags)
- VSCode extension (package.json + git tag)
- Contract modules (path-based versioning)

**Non-Versioned Modules**:
- Infrastructure (affects versioned modules)
- Documentation (informational only)
- Configuration (behavioral changes may trigger module versions)

## Quick Reference

### By Commit Prefix

- `src-mcp-pwsh:` → `src/mcp/pwsh/`
- `src-mcp-docs:` → `src/mcp/docs/`
- `src-mcp-github:` → `src/mcp/github/`
- `src-mcp-vscode:` → `src/mcp/vscode/`
- `vscode-ext-claude-commit:` → `.vscode/extensions/claude-mcp-vscode/`
- `infra:` → `automation/`, `containers/`
- `docs:` → `docs/`, root `*.md`
- `config:` → `.claude/`, `.vscode/`, root config
- `contracts:` → `contracts/{name}/{version}/`

### By Location

- `src/mcp/{name}/` → `mcp-{name}:`
- `.vscode/extensions/{name}/` → `vscode-ext-claude-commit:`
- `automation/`, `containers/` → `infra:`
- `docs/` → `docs:`
- `.claude/` → `config:`
- `contracts/{name}/{version}/` → `contracts:`

## Deployable Unit Contracts

Each deployable unit has a mandatory contract file at `contracts/deployable-units/0.1.0/{moniker}.yml`.

**Contract Schema** (all fields required for consistency):
- **Identity**: moniker, name, type, description
- **Versioning**: current_version, strategy, tag_format, changelog_path
- **Source**: root, includes, excludes
- **Build**: language, build_command, test_command, artifact_output
- **Deployment**: deployment_type, entry_point, runtime_dependencies
- **Commits**: prefix, version_affecting_types
- **Dependencies**: depends_on, used_by
- **Lifecycle**: start_command, stop_command, health_check
- **Integration**: protocol, transport, tools_provided/contributes
- **Metadata**: owner, documentation_path, api_documentation

**Example contracts**:
- [src-mcp-pwsh.yml](../../../contracts/deployable-units/0.1.0/src-mcp-pwsh.yml)
- [vscode-ext-claude-commit.yml](../../../contracts/deployable-units/0.1.0/vscode-ext-claude-commit.yml)

## References

- [Semantic Commits Guide](semantic-commits.md)
- [Versioning Documentation](versioning.md)
- [Deployable Units Contracts](../../../contracts/deployable-units/0.1.0/)
