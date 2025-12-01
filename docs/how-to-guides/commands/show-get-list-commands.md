# Show, Get, and List Commands

**Problem**: Understanding repository structure, modules, dependencies, and files requires manual exploration.

**Solution**: Use `show`, `get`, and `list` commands to query repository information in human-readable or machine-readable formats.

## Key Benefits

- Quick repository insights
- Human-readable tables (show commands)
- Machine-readable JSON (get commands)
- Module discovery and navigation
- Dependency visualization
- File ownership tracking

## Command Categories

| Category | Commands              | Format      | Use Case                           |
| -------- | --------------------- | ----------- | ---------------------------------- |
| **show** | Human-readable tables | Text/Tables | Interactive exploration, debugging |
| **get**  | Machine-readable data | JSON        | Scripting, automation, CI/CD       |
| **list** | File/item listings    | Text        | Quick lookups, piping              |

## Quick Reference

```bash
# Show commands (human-readable)
r2r eac show modules                    # Module table
r2r eac show dependencies               # Dependency graph
r2r eac show files                      # File ownership table
r2r eac show moduletypes                # Module type distribution
r2r eac show tests                      # Test suites table
r2r eac show environments               # Environment configs
r2r eac show config                     # EAC configuration

# Get commands (JSON output)
r2r eac get modules                     # Module contracts (JSON)
r2r eac get dependencies                # Dependency graph (JSON)
r2r eac get files                       # File mappings (JSON)
r2r eac get changed-modules             # Changed modules (JSON)
r2r eac get execution-order src-cli     # Build order (JSON)
r2r eac get config                      # EAC configuration (JSON)
r2r eac get tests                       # All tests (JSON)
r2r eac get environments                # Environment contracts (JSON)
r2r eac get changed-modules-ci          # Modules needing rebuild since last CI

# List commands
r2r eac list                            # List available extensions/commands
```

## Show Commands

### show modules

Display module table with details.

```bash
r2r eac show modules

# Output:
# ┌─────────────────┬──────────────┬────────────┬─────────────┐
# │ Module          │ Type         │ Files      │ Dependencies│
# ├─────────────────┼──────────────┼────────────┼─────────────┤
# │ src-commands    │ go-commands  │ 45         │ 2           │
# │ src-core        │ go-library   │ 32         │ 0           │
# │ src-auth        │ go-library   │ 28         │ 1           │
# │ docs            │ mkdocs-site  │ 156        │ 0           │
# └─────────────────┴──────────────┴────────────┴─────────────┘
```

### show dependencies

Display dependency graph.

```bash
r2r eac show dependencies

# Output:
# Module Dependencies:
#
# src-cli
#   └─→ src-commands
#       └─→ src-core
#
# src-auth
#   └─→ src-core
#
# src-api
#   ├─→ src-core
#   └─→ src-auth
```

### show files

Display file ownership by module.

```bash
r2r eac show files

# Output:
# ┌────────────────────────────────────┬─────────────────┐
# │ File                               │ Module          │
# ├────────────────────────────────────┼─────────────────┤
# │ src/commands/impl/work/work.go     │ src-commands    │
# │ src/core/repository/repo.go        │ src-core        │
# │ src/auth/jwt/token.go              │ src-auth        │
# │ docs/index.md                      │ docs            │
# └────────────────────────────────────┴─────────────────┘
#
# Total: 2,690 files across 15 modules
```

### show moduletypes

Display module type distribution.

```bash
r2r eac show moduletypes

# Output:
# Module Types:
#
# go-library       : 8 modules
# go-commands      : 3 modules
# go-cli           : 1 module
# mkdocs-site      : 1 module
# specifications   : 1 module
# contracts        : 1 module
#
# Total: 15 modules across 6 types
```

### show tests

Display test suites.

```bash
r2r eac show tests

# Output:
# ┌──────────────┬────────────┬────────────┬─────────────┐
# │ Suite        │ Module     │ Tests      │ Status      │
# ├──────────────┼────────────┼────────────┼─────────────┤
# │ integration  │ src-cli    │ 12         │ passing     │
# │ e2e          │ src-api    │ 8          │ passing     │
# │ smoke        │ src-auth   │ 3          │ passing     │
# └──────────────┴────────────┴────────────┴─────────────┘
```

### show environments

Display environment configurations.

```bash
r2r eac show environments

# Output:
# ┌────────────┬─────────────────┬──────────────────────┐
# │ Environment│ Description     │ Variables            │
# ├────────────┼─────────────────┼──────────────────────┤
# │ dev        │ Development     │ DEBUG=true, PORT=3000│
# │ staging    │ Staging         │ DEBUG=false, PORT=80 │
# │ prod       │ Production      │ DEBUG=false, PORT=80 │
# └────────────┴─────────────────┴──────────────────────┘
```

### show config

Display all EAC configuration with defaults applied in human-readable format.

```bash
r2r eac show config

# Output:
# EAC Configuration:
#
# ┌─────────────────────┬──────────────────────────────────────┐
# │ Config              │ Value                                │
# ├─────────────────────┼──────────────────────────────────────┤
# │ repository.root     │ /home/user/projects/eac              │
# │ repository.name     │ eac                                  │
# │ ai.provider         │ anthropic                            │
# │ ai.model            │ claude-sonnet-4                      │
# │ build.parallel      │ true                                 │
# │ test.timeout        │ 30m                                  │
# └─────────────────────┴──────────────────────────────────────┘
#
# Total: 6 configuration settings
```

### show files-changed

Display changed files with module ownership.

```bash
r2r eac show-files-changed

# Output:
# ┌────────────────────────────────────┬─────────────────┐
# │ Changed File                       │ Module          │
# ├────────────────────────────────────┼─────────────────┤
# │ src/commands/impl/work/remove.go   │ src-commands    │
# │ src/core/repository/repo.go        │ src-core        │
# └────────────────────────────────────┴─────────────────┘
#
# Changed modules: src-commands, src-core
```

### show files-staged

Display staged files with module ownership.

```bash
r2r eac show-files-staged

# Output:
# ┌────────────────────────────────────┬─────────────────┐
# │ Staged File                        │ Module          │
# ├────────────────────────────────────┼─────────────────┤
# │ src/auth/jwt/token.go              │ src-auth        │
# │ src/auth/jwt/token_test.go         │ src-auth        │
# └────────────────────────────────────┴─────────────────┘
```

## Get Commands (JSON Output)

### get modules

Get module contracts as JSON.

```bash
r2r eac get modules

# Output (JSON):
{
  "modules": [
    {
      "moniker": "src-commands",
      "type": "go-commands",
      "path": "src/commands",
      "dependencies": ["src-core"],
      "files": 45
    }
  ]
}
```

**Use cases:**

- CI/CD scripts
- Build automation
- Module analysis tools

### get dependencies

Get dependency graph as JSON.

```bash
r2r eac get dependencies

# Output (JSON):
{
  "dependencies": {
    "src-cli": ["src-commands"],
    "src-commands": ["src-core"],
    "src-auth": ["src-core"],
    "src-api": ["src-core", "src-auth"]
  }
}
```

**Diagram output formats:**

```bash
# PlantUML format
r2r eac get dependencies --as-plantuml

# Output:
@startuml
component "src-core"
component "src-commands"
component "src-cli"
component "src-auth"
component "src-api"

"src-cli" --> "src-commands"
"src-commands" --> "src-core"
"src-auth" --> "src-core"
"src-api" --> "src-core"
"src-api" --> "src-auth"
@enduml

# Mermaid format
r2r eac get dependencies --as-mermaid

# Output:
graph TD
    src-cli --> src-commands
    src-commands --> src-core
    src-auth --> src-core
    src-api --> src-core
    src-api --> src-auth
```

### get files

Get file-to-module mappings.

```bash
r2r eac get files

# Output (JSON):
{
  "files": [
    {
      "path": "src/commands/impl/work/work.go",
      "module": "src-commands"
    }
  ],
  "total": 2690
}
```

**Warning:** This command loads ~2,690 files (~19k tokens). Use only when needed.

**Alternatives for specific queries:**

- `show-files-changed` - Only changed files
- `show-files-staged` - Only staged files

### get changed-modules

Get modules affected by changes.

```bash
r2r eac get changed-modules

# Output (JSON):
{
  "changed_modules": [
    "src-commands",
    "src-core"
  ]
}
```

**Use cases:**

- Incremental builds
- Selective testing
- CI optimization

### get execution-order

Get build order for modules based on dependencies.

```bash
r2r eac get execution-order src-cli

# Output (JSON):
{
  "execution_order": [
    "src-core",
    "src-commands",
    "src-cli"
  ]
}
```

**Use case:** Build modules in correct dependency order.

### get config

Get all EAC configuration (6 configs with defaults) in structured format.

```bash
r2r eac get config

# Output (JSON):
{
  "repository": {
    "root": "/home/user/projects/eac",
    "name": "eac"
  },
  "ai": {
    "provider": "anthropic",
    "model": "claude-sonnet-4"
  },
  "build": {
    "parallel": true
  },
  "test": {
    "timeout": "30m"
  }
}
```

**Use cases:**

- CI/CD configuration validation
- Environment-specific settings
- Automation scripts

### get tests

Get all tests in the repository in structured format.

```bash
r2r eac get tests

# Output (JSON):
{
  "tests": [
    {
      "suite": "integration",
      "module": "src-cli",
      "count": 12,
      "status": "passing",
      "path": "src/cli/tests/integration"
    },
    {
      "suite": "e2e",
      "module": "src-api",
      "count": 8,
      "status": "passing",
      "path": "src/api/tests/e2e"
    }
  ],
  "total": 20
}
```

**Use cases:**

- Test coverage analysis
- CI/CD test orchestration
- Test suite management

### get environments

Get all environment contracts in structured format.

```bash
r2r eac get environments

# Output (JSON):
{
  "environments": [
    {
      "name": "dev",
      "description": "Development",
      "variables": {
        "DEBUG": "true",
        "PORT": "3000"
      }
    },
    {
      "name": "staging",
      "description": "Staging",
      "variables": {
        "DEBUG": "false",
        "PORT": "80"
      }
    },
    {
      "name": "prod",
      "description": "Production",
      "variables": {
        "DEBUG": "false",
        "PORT": "80"
      }
    }
  ]
}
```

**Use cases:**

- Environment deployment automation
- Configuration validation
- Infrastructure as code

### get changed-modules-ci

Get modules requiring rebuild since last successful CI run.

```bash
r2r eac get changed-modules-ci

# Output (JSON):
{
  "changed_modules": [
    "src-commands",
    "src-core"
  ],
  "base_commit": "abc123",
  "head_commit": "def456"
}
```

**Flags:**

- `--pr-base` - Base branch for PR comparison (default: main)
- `--workflow` - GitHub workflow name to check (default: CI)
- `--branch` - Branch to check CI status for (default: current branch)

**Examples:**

```bash
# Check changes for PR
r2r eac get changed-modules-ci --pr-base main

# Check specific workflow
r2r eac get changed-modules-ci --workflow "Build and Test"

# Check different branch
r2r eac get changed-modules-ci --branch develop
```

**Use cases:**

- CI optimization - only build/test changed modules
- PR validation
- Incremental deployments

### get suite

Get test suite information.

```bash
r2r eac get suite integration

# Output (JSON):
{
  "name": "integration",
  "module": "src-cli",
  "tests": 12,
  "status": "passing"
}
```

## Typical Workflows

### Module Discovery

```bash
# What modules exist?
r2r eac show modules

# What types are there?
r2r eac show moduletypes

# What depends on what?
r2r eac show dependencies
```

### Pre-commit Analysis

```bash
# What files changed?
r2r eac show-files-changed

# What modules are affected?
r2r eac get changed-modules

# Build only affected modules
MODULES=$(r2r eac get changed-modules | jq -r '.changed_modules[]')
for module in $MODULES; do
  r2r eac build $module
  r2r eac test $module
done
```

### Build Optimization

```bash
# Get execution order
r2r eac get execution-order src-cli | jq -r '.execution_order[]' | while read module; do
  echo "Building $module..."
  r2r eac build $module
done
```

### File Ownership

```bash
# Who owns this file?
r2r eac show files | grep "auth/jwt/token.go"

# All files in a module
r2r eac get files | jq '.files[] | select(.module == "src-auth")'
```

## Integration Patterns

### CI/CD Optimization

```yaml
# GitHub Actions - Build only changed modules
- name: Get Changed Modules
  id: changed
  run: |
    MODULES=$(r2r eac get changed-modules | jq -r '.changed_modules | join(" ")')
    echo "modules=$MODULES" >> $GITHUB_OUTPUT

- name: Build Changed Modules
  run: |
    for module in ${{ steps.changed.outputs.modules }}; do
      r2r eac build $module
      r2r eac test $module
    done
```

### Incremental Builds

```bash
#!/bin/bash
# build-changed.sh

# Get changed modules
CHANGED=$(r2r eac get changed-modules | jq -r '.changed_modules[]')

if [ -z "$CHANGED" ]; then
  echo "No changes detected"
  exit 0
fi

# Build in dependency order
for module in $CHANGED; do
  ORDER=$(r2r eac get execution-order $module | jq -r '.execution_order[]')
  for dep in $ORDER; do
    echo "Building $dep..."
    r2r eac build module $dep || exit 1
  done
done
```

### Module Analytics

```bash
#!/bin/bash
# analyze-modules.sh

# Get all modules
MODULES=$(r2r eac get modules | jq -r '.modules[].moniker')

# Analyze each
for module in $MODULES; do
  TYPE=$(r2r eac get modules | jq -r ".modules[] | select(.moniker == \"$module\") | .type")
  FILES=$(r2r eac get modules | jq -r ".modules[] | select(.moniker == \"$module\") | .files")
  DEPS=$(r2r eac get dependencies | jq -r ".dependencies[\"$module\"] | length // 0")

  echo "$module: type=$TYPE, files=$FILES, deps=$DEPS"
done
```

## Best Practices

- **Use show for humans**: Interactive exploration and debugging
- **Use get for scripts**: Automation and CI/CD pipelines
- **Avoid get files in loops**: It's expensive; use targeted queries
- **Cache results**: Store JSON output for repeated queries
- **Parse with jq**: Use jq for JSON processing in scripts
- **Check exit codes**: Commands return non-zero on errors

## Troubleshooting

| Problem                     | Solution                                                |
| --------------------------- | ------------------------------------------------------- |
| Empty output                | No modules/files found, check repository structure      |
| JSON parse error            | Pipe through `jq` for validation                        |
| Slow `get files`            | Use `show-files-changed` or `show-files-staged` instead |
| Module not found            | Check module contract exists                            |
| Circular dependency warning | Review `show dependencies`, fix architecture            |

## Advanced Usage

### Dependency Analysis

```bash
# Find modules with no dependencies
r2r eac get modules | jq '.modules[] | select(.dependencies | length == 0) | .moniker'

# Find most depended-on modules
r2r eac get dependencies | jq -r '
  [.dependencies | to_entries[] | .value[]] |
  group_by(.) |
  map({module: .[0], count: length}) |
  sort_by(.count) |
  reverse |
  .[]
'
```

### File Statistics

```bash
# Files per module
r2r eac get modules | jq '.modules[] | "\(.moniker): \(.files) files"'

# Largest modules
r2r eac get modules | jq '.modules | sort_by(.files) | reverse | .[0:5]'
```

### Build Matrix

```bash
# Generate build matrix for GitHub Actions
r2r eac get modules | jq '{module: [.modules[].moniker]}'

# Output:
# {
#   "module": ["src-cli", "src-commands", "src-core", ...]
# }
```

## Summary

**Show commands** (human-readable):

- `show modules` - Module table
- `show dependencies` - Dependency graph
- `show files` - File ownership
- `show moduletypes` - Module type distribution
- `show tests` - Test suites table
- `show environments` - Environment configs
- `show config` - EAC configuration
- `show-files-changed` - Changed files only
- `show-files-staged` - Staged files only

**Get commands** (JSON):

- `get modules` - Module data
- `get dependencies` - Dependency data (supports `--as-plantuml`, `--as-mermaid`)
- `get files` - File mappings
- `get changed-modules` - Affected modules
- `get changed-modules-ci` - Modules needing rebuild since last CI
- `get execution-order` - Build order
- `get config` - EAC configuration
- `get tests` - All tests
- `get environments` - Environment contracts
- `get suite` - Test suite information

Use show commands interactively, get commands in automation.
