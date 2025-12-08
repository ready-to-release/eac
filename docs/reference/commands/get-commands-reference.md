# Get Commands Reference

**Purpose**: Get commands output structured JSON data for automation, scripting, and CI/CD integration.

**For interactive use**: See [Show Commands](../../how-to-guides/eac/commands/show-commands.md) for human-readable output.

## Overview

Get commands are designed for programmatic access to repository information:

- **Output format**: JSON (structured, machine-readable)
- **Use cases**: Automation, CI/CD pipelines, build scripts, analytics
- **Processing**: Pipe through `jq` for filtering and transformation
- **Exit codes**: Non-zero on errors for script error handling

## All Get Commands

### get modules

Get module contracts as JSON.

```bash
r2r eac get modules

# Output (JSON):
{
  "modules": [
    {
      "moniker": "eac-commands",
      "type": "go-commands",
      "path": "go/eac/commands",
      "dependencies": ["eac-core"],
      "files": 45
    }
  ]
}
```

**Use cases:**
- CI/CD scripts
- Build automation
- Module analysis tools

**jq examples:**

```bash
# Extract module names only
r2r eac get modules | jq -r '.modules[].moniker'

# Find modules of a specific type
r2r eac get modules | jq '.modules[] | select(.type == "go-library")'

# Count modules by type
r2r eac get modules | jq -r '.modules[].type' | sort | uniq -c

# Find modules with no dependencies
r2r eac get modules | jq '.modules[] | select(.dependencies | length == 0) | .moniker'
```

### get dependencies

Get dependency graph as JSON.

```bash
r2r eac get dependencies

# Output (JSON):
{
  "dependencies": {
    "r2r-cli": ["eac-commands"],
    "eac-commands": ["eac-core"],
    "src-auth": ["eac-core"],
    "src-api": ["eac-core", "src-auth"]
  }
}
```

**Diagram output formats:**

```bash
# PlantUML format
r2r eac get dependencies --as-plantuml

# Output:
@startuml
component "eac-core"
component "eac-commands"
component "r2r-cli"
component "src-auth"
component "src-api"

"r2r-cli" --> "eac-commands"
"eac-commands" --> "eac-core"
"src-auth" --> "eac-core"
"src-api" --> "eac-core"
"src-api" --> "src-auth"
@enduml

# Mermaid format
r2r eac get dependencies --as-mermaid

# Output:
graph TD
    r2r-cli --> eac-commands
    eac-commands --> eac-core
    src-auth --> eac-core
    src-api --> eac-core
    src-api --> src-auth
```

**jq examples:**

```bash
# Find all dependencies of a module
r2r eac get dependencies | jq '.dependencies["src-api"]'

# Find modules with no dependencies
r2r eac get dependencies | jq -r 'to_entries[] | select(.value | length == 0) | .key'

# Find most depended-on modules
r2r eac get dependencies | jq -r '
  [.dependencies | to_entries[] | .value[]] |
  group_by(.) |
  map({module: .[0], count: length}) |
  sort_by(.count) |
  reverse |
  .[]
'

# Count total dependencies
r2r eac get dependencies | jq '[.dependencies | to_entries[] | .value | length] | add'
```

### get files

Get file-to-module mappings.

```bash
r2r eac get files

# Output (JSON):
{
  "files": [
    {
      "path": "go/eac/commands/impl/work/work.go",
      "module": "eac-commands"
    }
  ],
  "total": 2690
}
```

**Performance Warning:**

> **This command is expensive!** It loads ~2,690 files (~19k tokens). Use sparingly and only when needed.

**Alternatives for specific queries:**
- `show files-changed` - Only changed files (human-readable)
- `show files-staged` - Only staged files (human-readable)
- `get changed-modules` - Affected modules only (JSON)

**jq examples:**

```bash
# All files in a specific module
r2r eac get files | jq '.files[] | select(.module == "src-auth")'

# Count files per module
r2r eac get files | jq -r '.files[].module' | sort | uniq -c

# Find files matching a pattern
r2r eac get files | jq '.files[] | select(.path | contains("auth/jwt"))'

# Extract just file paths
r2r eac get files | jq -r '.files[].path'
```

**Best practice**: Cache the output if you need to query it multiple times:

```bash
# Cache to file
r2r eac get files > files.json

# Query the cache
jq '.files[] | select(.module == "src-auth")' files.json
```

### get changed-modules

Get modules affected by changes.

```bash
r2r eac get changed-modules

# Output (JSON):
{
  "changed_modules": [
    "eac-commands",
    "eac-core"
  ]
}
```

**Use cases:**
- Incremental builds
- Selective testing
- CI optimization

**jq examples:**

```bash
# Extract module names as space-separated list
r2r eac get changed-modules | jq -r '.changed_modules | join(" ")'

# Count changed modules
r2r eac get changed-modules | jq '.changed_modules | length'

# Check if specific module changed
r2r eac get changed-modules | jq '.changed_modules | contains(["src-auth"])'
```

### get execution order

Get build order for modules based on dependencies.

```bash
r2r eac get execution order r2r-cli

# Output (JSON):
{
  "execution_order": [
    "eac-core",
    "eac-commands",
    "r2r-cli"
  ]
}
```

**Use case**: Build modules in correct dependency order.

**jq examples:**

```bash
# Extract build order as array
r2r eac get execution order r2r-cli | jq -r '.execution_order[]'

# Get first module to build
r2r eac get execution order r2r-cli | jq -r '.execution_order[0]'

# Get last module (the target)
r2r eac get execution order r2r-cli | jq -r '.execution_order[-1]'
```

### get config

Get all EAC configuration with defaults applied in structured format.

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

**jq examples:**

```bash
# Get specific config value
r2r eac get config | jq -r '.ai.model'

# Check if parallel builds enabled
r2r eac get config | jq -r '.build.parallel'

# Get repository root
r2r eac get config | jq -r '.repository.root'

# List all config keys
r2r eac get config | jq -r 'paths(scalars) | join(".")'
```

### get tests

Get all tests in the repository in structured format.

```bash
r2r eac get tests

# Output (JSON):
{
  "tests": [
    {
      "suite": "integration",
      "module": "r2r-cli",
      "count": 12,
      "status": "passing",
      "path": "go/r2r/cli/tests/integration"
    },
    {
      "suite": "e2e",
      "module": "src-api",
      "count": 8,
      "status": "passing",
      "path": "go/eac/api/tests/e2e"
    }
  ],
  "total": 20
}
```

**Use cases:**
- Test coverage analysis
- CI/CD test orchestration
- Test suite management

**jq examples:**

```bash
# Tests for specific module
r2r eac get tests | jq '.tests[] | select(.module == "r2r-cli")'

# Total test count
r2r eac get tests | jq '.total'

# Count tests by status
r2r eac get tests | jq -r '.tests[].status' | sort | uniq -c

# Find failing tests
r2r eac get tests | jq '.tests[] | select(.status != "passing")'

# Sum all test counts
r2r eac get tests | jq '[.tests[].count] | add'
```

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

**jq examples:**

```bash
# Get environment names
r2r eac get environments | jq -r '.environments[].name'

# Get variables for specific environment
r2r eac get environments | jq '.environments[] | select(.name == "prod") | .variables'

# Export environment variables for dev
r2r eac get environments | jq -r '
  .environments[] |
  select(.name == "dev") |
  .variables |
  to_entries[] |
  "export \(.key)=\(.value)"
'

# Compare prod vs staging
r2r eac get environments | jq '[.environments[] | select(.name == "prod" or .name == "staging")]'
```

### get changed-modules-ci

Get modules requiring rebuild since last successful CI run.

```bash
r2r eac get changed-modules-ci

# Output (JSON):
{
  "changed_modules": [
    "eac-commands",
    "eac-core"
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

**jq examples:**

```bash
# Get changed modules as space-separated list
r2r eac get changed-modules-ci | jq -r '.changed_modules | join(" ")'

# Get commit range
r2r eac get changed-modules-ci | jq -r '"\(.base_commit)..\(.head_commit)"'

# Check if any modules changed
r2r eac get changed-modules-ci | jq '.changed_modules | length > 0'
```

### get suite

Get test suite information.

```bash
r2r eac get suite integration

# Output (JSON):
{
  "name": "integration",
  "module": "r2r-cli",
  "tests": 12,
  "status": "passing"
}
```

**jq examples:**

```bash
# Get test count
r2r eac get suite integration | jq '.tests'

# Check if passing
r2r eac get suite integration | jq -r '.status == "passing"'
```

### list

List available extensions/commands.

```bash
r2r eac list

# Output varies based on available extensions
```

**Use case**: Discovering available commands and extensions.

## Integration Patterns

### CI/CD Optimization

**GitHub Actions - Build only changed modules:**

```yaml
name: Incremental Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Get Changed Modules
        id: changed
        run: |
          MODULES=$(r2r eac get changed-modules | jq -r '.changed_modules | join(" ")')
          echo "modules=$MODULES" >> $GITHUB_OUTPUT

      - name: Build Changed Modules
        run: |
          for module in ${{ steps.changed.outputs.modules }}; do
            echo "Building $module..."
            r2r eac build $module
            r2r eac test $module
          done
```

**GitHub Actions - Build matrix:**

```yaml
name: Build Matrix

on: [push]

jobs:
  generate-matrix:
    runs-on: ubuntu-latest
    outputs:
      matrix: ${{ steps.set-matrix.outputs.matrix }}
    steps:
      - uses: actions/checkout@v2
      - id: set-matrix
        run: |
          MATRIX=$(r2r eac get modules | jq -c '{module: [.modules[].moniker]}')
          echo "matrix=$MATRIX" >> $GITHUB_OUTPUT

  build:
    needs: generate-matrix
    runs-on: ubuntu-latest
    strategy:
      matrix: ${{ fromJson(needs.generate-matrix.outputs.matrix) }}
    steps:
      - uses: actions/checkout@v2
      - name: Build ${{ matrix.module }}
        run: r2r eac build ${{ matrix.module }}
```

### Incremental Builds

**Build only changed modules in dependency order:**

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
  ORDER=$(r2r eac get execution order $module | jq -r '.execution_order[]')
  for dep in $ORDER; do
    echo "Building $dep..."
    r2r eac build module $dep || exit 1
  done
done
```

### Module Analytics

**Analyze module metrics:**

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

**Generate module report:**

```bash
#!/bin/bash
# module-report.sh

echo "Module Report"
echo "============="
echo ""

# Module count by type
echo "Modules by Type:"
r2r eac get modules | jq -r '.modules[].type' | sort | uniq -c

echo ""
echo "Largest Modules:"
r2r eac get modules | jq -r '.modules | sort_by(.files) | reverse | .[0:5] | .[] | "\(.moniker): \(.files) files"'

echo ""
echo "Modules with Most Dependencies:"
r2r eac get modules | jq -r '.modules | sort_by(.dependencies | length) | reverse | .[0:5] | .[] | "\(.moniker): \(.dependencies | length) deps"'
```

## Advanced Usage with jq

### Dependency Analysis

**Find modules with no dependencies:**

```bash
r2r eac get modules | jq '.modules[] | select(.dependencies | length == 0) | .moniker'
```

**Find most depended-on modules:**

```bash
r2r eac get dependencies | jq -r '
  [.dependencies | to_entries[] | .value[]] |
  group_by(.) |
  map({module: .[0], count: length}) |
  sort_by(.count) |
  reverse |
  .[]
'
```

**Detect circular dependencies (basic check):**

```bash
# This checks if a module depends on itself through any path
# More sophisticated circular dependency detection requires graph traversal
r2r eac get dependencies | jq 'to_entries[] | select(.value | contains([.key]))'
```

### File Statistics

**Files per module:**

```bash
r2r eac get modules | jq '.modules[] | "\(.moniker): \(.files) files"'
```

**Largest modules:**

```bash
r2r eac get modules | jq '.modules | sort_by(.files) | reverse | .[0:5]'
```

**Total file count:**

```bash
r2r eac get modules | jq '[.modules[].files] | add'
```

### Build Matrix Generation

**Generate build matrix for GitHub Actions:**

```bash
r2r eac get modules | jq '{module: [.modules[].moniker]}'

# Output:
# {
#   "module": ["r2r-cli", "eac-commands", "eac-core", ...]
# }
```

**Generate build matrix for changed modules only:**

```bash
r2r eac get changed-modules | jq '{module: .changed_modules}'
```

### Test Analysis

**Find modules without tests:**

```bash
# Get all modules
ALL_MODULES=$(r2r eac get modules | jq -r '.modules[].moniker')

# Get modules with tests
TESTED_MODULES=$(r2r eac get tests | jq -r '.tests[].module' | sort -u)

# Find difference
comm -23 <(echo "$ALL_MODULES" | sort) <(echo "$TESTED_MODULES")
```

**Calculate test coverage percentage:**

```bash
TOTAL_MODULES=$(r2r eac get modules | jq '.modules | length')
TESTED_MODULES=$(r2r eac get tests | jq '[.tests[].module] | unique | length')

echo "scale=2; $TESTED_MODULES * 100 / $TOTAL_MODULES" | bc
```

## Performance Notes

### get files is Expensive

The `get files` command loads ~2,690 files (~19k tokens) and should be used sparingly:

**Avoid:**
```bash
# DON'T do this in a loop
for module in $(r2r eac get modules | jq -r '.modules[].moniker'); do
  r2r eac get files | jq ".files[] | select(.module == \"$module\")"
done
```

**Better:**
```bash
# Cache the result once
r2r eac get files > files.json

# Query the cache
for module in $(r2r eac get modules | jq -r '.modules[].moniker'); do
  jq ".files[] | select(.module == \"$module\")" files.json
done
```

**Best:**
```bash
# Use get modules which includes file counts
r2r eac get modules | jq '.modules[] | "\(.moniker): \(.files) files"'

# Or use show files-changed for changed files only
r2r eac show files-changed
```

### Caching Strategies

**Cache expensive queries:**

```bash
# Cache files
r2r eac get files > .cache/files.json

# Cache modules
r2r eac get modules > .cache/modules.json

# Cache dependencies
r2r eac get dependencies > .cache/dependencies.json

# Query from cache
jq '.modules[] | select(.type == "go-library")' .cache/modules.json
```

**Invalidate cache on changes:**

```bash
# Check if cache is stale
if [ .cache/files.json -ot .r2r/eac/modules.yml ]; then
  echo "Cache is stale, regenerating..."
  r2r eac get files > .cache/files.json
fi
```

## Best Practices

1. **Always pipe through jq**: Validate and process JSON output
2. **Cache expensive queries**: Store results for repeated queries
3. **Check exit codes**: Commands return non-zero on errors
4. **Use specific queries**: Avoid `get files` when alternatives exist
5. **Combine commands**: Use multiple get commands for complex queries
6. **Handle empty results**: Check for empty arrays/objects before processing
7. **Use raw output**: Add `-r` flag to jq for script-friendly output

**Example with error handling:**

```bash
#!/bin/bash
set -e  # Exit on error

# Get changed modules with error handling
if ! CHANGED=$(r2r eac get changed-modules); then
  echo "Error: Failed to get changed modules"
  exit 1
fi

# Check if any modules changed
if [ $(echo "$CHANGED" | jq '.changed_modules | length') -eq 0 ]; then
  echo "No modules changed"
  exit 0
fi

# Process changed modules
echo "$CHANGED" | jq -r '.changed_modules[]' | while read module; do
  echo "Building $module..."
  r2r eac build $module
done
```

## Troubleshooting

| Problem                | Solution                                              |
| ---------------------- | ----------------------------------------------------- |
| JSON parse error       | Pipe through `jq` for validation and detailed errors  |
| Empty output           | Check if modules/files exist in repository            |
| Slow `get files`       | Use alternatives or cache the result                  |
| Module not found       | Verify module contract exists in `.r2r/eac/modules.yml` |
| Invalid jq syntax      | Test jq filter separately: `echo '{"test":1}' \| jq '.test'` |
| Exit code issues       | Check `$?` after command and handle non-zero          |
| Memory issues          | Avoid loading all files repeatedly, use caching       |

## Command Reference Table

| Command                  | Output             | Performance | Use Case                          |
| ------------------------ | ------------------ | ----------- | --------------------------------- |
| `get modules`            | Module contracts   | Fast        | Module metadata, CI/CD            |
| `get dependencies`       | Dependency graph   | Fast        | Build order, architecture         |
| `get files`              | File mappings      | **SLOW**    | File analysis (cache results!)    |
| `get changed-modules`    | Changed modules    | Fast        | Incremental builds                |
| `get changed-modules-ci` | CI changes         | Medium      | CI optimization                   |
| `get execution order`    | Build order        | Fast        | Sequential builds                 |
| `get config`             | EAC configuration  | Fast        | Config validation                 |
| `get tests`              | Test suites        | Fast        | Test orchestration                |
| `get environments`       | Environments       | Fast        | Deployment automation             |
| `get suite`              | Suite info         | Fast        | Specific test suite details       |
| `list`                   | Extensions         | Fast        | Command discovery                 |

**For human-readable output**: Use [show commands](../../how-to-guides/eac/commands/show-commands.md) instead.
