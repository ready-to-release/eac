# Intelligent Module Detection

## How It Works

The `determineFileModule()` function now **intelligently extracts module names** from file paths using pattern matching, rather than hardcoding module mappings.

### Key Improvement

**Before:** Hardcoded mappings required code changes for each new module

```go
{"automation/", "infra"},  // All automation → generic "infra"
{"containers/", "infra"},  // All containers → generic "infra"
```

**After:** Dynamic extraction based on folder structure

```go
automation/<module-name>/... → module-name
containers/<module-name>/... → module-name
```

## Pattern Recognition

### 1. Automation Modules

**Pattern:** `automation/<module-name>/...`
**Module:** Extracted directly from folder name

| File Path | Detected Module | Type Prefix |
|-----------|----------------|-------------|
| `automation/sh/vscode/install.sh` | `sh-vscode` | `sh-` (shell) |
| `automation/pwsh-build/build.ps1` | `pwsh-build` | `pwsh-` (PowerShell) |
| `automation/sh-deploy/deploy.sh` | `sh-deploy` | `sh-` (shell) |

**Commit Example:**

```text
sh-vscode: feat: add automated installation

Implements automated VSCode extension
installation via shell script.
```

### 2. Container Modules

**Pattern:** `containers/<module-name>/...`
**Module:** Extracted from container folder name

| File Path | Detected Module |
|-----------|----------------|
| `containers/mkdocs/Dockerfile` | `mkdocs` |
| `containers/nginx-proxy/config.conf` | `nginx-proxy` |

**Commit Example:**

```text
mkdocs: chore: update base image

Updates mkdocs container to use Python
3.11 for improved build performance.
```

### 3. MCP Server Modules

**Pattern:** `src/mcp/<service>/...`
**Module:** `mcp-<service>`

| File Path | Detected Module |
|-----------|----------------|
| `src/mcp/pwsh/main.go` | `src-mcp-pwsh` |
| `src/mcp/vscode/main.go` | `src-mcp-vscode` |
| `src/mcp/docs/server.go` | `src-mcp-docs` |
| `src/mcp/github/api.go` | `src-mcp-github` |

### 4. VSCode Extension

**Pattern:** `.vscode/extensions/<name>/...`
**Module:** `vscode-extension` (or extracted name)

| File Path | Detected Module |
|-----------|----------------|
| `.vscode/extensions/claude-mcp-vscode/src/extension.ts` | `vscode-extension` |

### 5. Contract Modules

**Pattern:** `contracts/<name>/<version>/...`
**Module:** `contracts-<name>`

| File Path                                       | Detected Module              |
|-------------------------------------------------|------------------------------|
| `contracts/repository/0.1.0/definitions.yml`    | `contracts-repository`       |
| `contracts/deployable-units/0.1.0/src-mcp-pwsh.yml` | `contracts-deployable-units` |

### 6. Documentation

**Pattern:** `docs/...` or `*.md` (root)
**Module:** `docs`

| File Path                            | Detected Module |
|--------------------------------------|-----------------|
| `docs/reference/continuous-delivery/versioning.md` | `docs`          |
| `README.md`                          | `docs`          |
| `QUICKSTART.md`                      | `docs`          |

### 7. Configuration

**Patterns:**

- `.claude/...` → `claude-config`
- `.vscode/...` (non-extension) → `vscode-config`
- Root config files → `repo-config`

| File Path | Detected Module |
|-----------|----------------|
| `.claude/agents/vscode-extensionension-commit-button.md` | `claude-config` |
| `.vscode/settings.json` | `vscode-config` |
| `.gitignore` | `repo-config` |
| `mkdocs.yml` | `repo-config` |
| `package.json` | `repo-config` |

## Benefits

### 1. **Automatic Discovery**

New modules are automatically detected without code changes:

```bash
# Add new automation module
mkdir automation/py-test
# Automatically detected as: py-test
```

### 2. **Type Inference from Prefixes**

Module type is self-documenting through naming:

- `sh-*` → Shell scripts
- `pwsh-*` → PowerShell scripts
- `py-*` → Python scripts
- `mcp-*` → MCP servers

### 3. **Granular Commit Messages**

Instead of generic module names:

```text
❌ infra: add build script

✅ pwsh-build: add deployment automation
```

### 4. **Scalable Structure**

Repository grows naturally:

```text
automation/
  ├── sh-vscode/      → sh-vscode
  ├── pwsh-build/     → pwsh-build
  └── py-test/        → py-test (future)

containers/
  ├── mkdocs/         → mkdocs
  └── nginx-proxy/    → nginx-proxy
```

## Example Commit Message

With the improved detection, a multi-module change generates:

```text
# Revision 5b6ec4a81fb7d2d0a5d916967b4ea740815a29ae

This commit adds automated PowerShell build tooling and updates
the MkDocs container base image to Python 3.11 for improved
build performance and security. These infrastructure improvements
reduce build times by ~30% in CI/CD pipelines.

| Status   | File                                    | Module      |
| -------- | --------------------------------------- | ----------- |
| added    | automation/pwsh-build/build.ps1         | pwsh-build  |
| added    | automation/pwsh-build/README.md         | pwsh-build  |
| modified | containers/mkdocs/Dockerfile            | docs        |
| modified | containers/mkdocs/requirements.txt      | docs        |

---

pwsh-build: feat: add CI/CD build automation

Implements automated build pipeline with
artifact generation and test execution.

---

mkdocs: chore: update base image to Python 3.11

Updates base image for improved performance
and security patches.
```
