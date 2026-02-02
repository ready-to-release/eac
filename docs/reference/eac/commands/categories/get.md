# Get Commands

## Overview

Get commands provide structured YAML/JSON/TOML output designed for automation, CI/CD pipelines, build scripts, and programmatic processing.

All get commands output valid YAML/JSON/TOML that can be piped through `jq` or processed by other tools.

**Key Characteristics**:

- Valid YAML/JSON/TOML output
- Machine-readable structured data
- Designed for automation
- Deterministic and cacheable
- Non-zero exit codes on errors

**When to use**: In CI/CD pipelines, build scripts, or when you need programmatic access to repository data.

**For interactive use**: Use [show commands](./show.md) instead, which provide human-readable formatted output.

## All Get Commands

<!-- book:category-commands get -->

## Common Patterns

### JSON Output Structure

All get commands return valid JSON with consistent structure:

```json
{
  "data_field": [...],
  "metadata": {...},
  "total": N
}
```

### Processing with jq

Get commands are designed to be piped through `jq`:

```bash
# Extract specific fields
eac get modules | jq -r '.modules[].moniker'

# Filter results
eac get modules | jq '.modules[] | select(.type == "go-library")'

# Transform data
eac get dependencies | jq 'to_entries | map({module: .key, deps: .value})'

# Count results
eac get modules | jq '.modules | length'
```

### Caching Results

JSON output is deterministic and cacheable:

```bash
# Cache expensive queries
eac get files > files.json

# Query from cache
jq '.files[] | select(.module == "src-auth")' files.json
```

## get vs show Duality

Most get commands have corresponding `show` commands that provide the same information in human-readable format:

| get command          | show command          | Output Format    |
| -------------------- | --------------------- | ---------------- |
| `get modules`        | `show modules`        | JSON / Table     |
| `get dependencies`   | `show dependencies`   | JSON / Table     |
| `get files`          | `show files`          | JSON / Table     |
| `get config`         | `show config`         | JSON / Formatted |
| `get tests`          | `show tests`          | JSON / Table     |
| `get environments`   | `show environments`   | JSON / Table     |
| `get build-times`    | `show build-times`    | JSON / Table     |
| `get test-timings`   | `show test-timings`   | JSON / Table     |
| `get suite <name>`   | `show suite <name>`   | JSON / Formatted |
| `get artifacts <m>`  | `show artifacts <m>`  | JSON / Table     |
| `get valid-commands` | `show valid-commands` | JSON / Table     |

**Rule**: Use `get` for automation and scripts, `show` for interactive terminal use.

## Common Workflows

### CI/CD Integration

```bash
# Get changed modules
CHANGED=$(eac get changed-modules-ci | jq -r '.changed_modules[]')

# Build in dependency order
for module in $CHANGED; do
  ORDER=$(eac get execution order $module | jq -r '.execution_order[]')
  for dep in $ORDER; do
    eac build $dep
  done
done
```

### Module Analysis

```bash
# Get all module monikers
eac get modules | jq -r '.modules[].moniker'

# Find modules by type
eac get modules | jq '.modules[] | select(.type == "go-library")'

# Count modules by type
eac get modules | jq '.modules | group_by(.type) | map({type: .[0].type, count: length})'
```

### Dependency Analysis

```bash
# Get dependency graph
eac get dependencies | jq '.'

# Find dependencies of a module
eac get dependencies | jq '.dependencies["src-api"]'

# Find modules with no dependencies
eac get dependencies | jq 'to_entries[] | select(.value | length == 0) | .key'
```

### Build Optimization

```bash
# Get execution order for parallel builds
eac get execution order r2r-cli | jq -r '.execution_order[]'

# Get build dependencies
eac get build-deps src-api | jq '.dependencies[]'

# Analyze build performance
eac get build-times | jq '[.builds[]] | sort_by(.duration) | reverse'
```

## Performance Notes

### Fast Commands

Execute quickly (< 1s):

- `get modules`
- `get dependencies`
- `get config`
- `get environments`
- `get changed-modules`
- `get execution order`

### Moderate Commands

May take a few seconds:

- `get tests` (scans test files)
- `get build-times` (parses build logs)
- `get test-timings` (parses test logs)

### Expensive Commands

Avoid in tight loops:

- **`get files`** - Loads ~2,690 files (~19k tokens)
  - Use `show files-changed` or `show files-staged` instead when possible
  - Cache results if querying multiple times

## Best Practices

### Always Use jq

Validate and process JSON output:

```bash
# Good: Validate JSON and extract data
eac get modules | jq -r '.modules[].moniker'

# Avoid: Raw grep on JSON (fragile)
eac get modules | grep '"moniker"'
```

### Check Exit Codes

Commands return non-zero on errors:

```bash
if ! OUTPUT=$(eac get modules 2>&1); then
  echo "Error: Failed to get modules"
  exit 1
fi

echo "$OUTPUT" | jq '.modules[]'
```

### Cache Expensive Queries

```bash
# Cache files query
if [ ! -f files.json ]; then
  eac get files > files.json
fi

# Query from cache
jq '.files[] | select(.module == "src-auth")' files.json
```

### Use Raw Output (-r)

Extract values without quotes:

```bash
# With -r: produces raw strings
eac get modules | jq -r '.modules[].moniker'
# Output: eac-commands

# Without -r: produces quoted strings
eac get modules | jq '.modules[].moniker'
# Output: "eac-commands"
```

### Handle Empty Results

```bash
MODULES=$(eac get modules | jq -r '.modules[]')
if [ -z "$MODULES" ]; then
  echo "No modules found"
  exit 1
fi
```

## Integration Examples

### GitHub Actions

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
          MODULES=$(eac get changed-modules-ci | jq -r '.changed_modules | join(" ")')
          echo "modules=$MODULES" >> $GITHUB_OUTPUT

      - name: Build
        run: |
          for module in ⟪ steps.changed.outputs.modules ⟫; do
            eac build $module
          done
```

### Build Matrix

```yaml
jobs:
  generate-matrix:
    runs-on: ubuntu-latest
    outputs:
      matrix: ⟪ steps.set-matrix.outputs.matrix ⟫
    steps:
      - id: set-matrix
        run: |
          MATRIX=$(eac get modules | jq -c '{module: [.modules[].moniker]}')
          echo "matrix=$MATRIX" >> $GITHUB_OUTPUT

  build:
    needs: generate-matrix
    strategy:
      matrix: ⟪ fromJson(needs.generate-matrix.outputs.matrix) ⟫
    steps:
      - run: eac build ⟪ matrix.module ⟫
```

### Build Script

```bash
#!/bin/bash
set -e

# Get changed modules
CHANGED=$(eac get changed-modules | jq -r '.changed_modules[]')

if [ -z "$CHANGED" ]; then
  echo "No changes detected"
  exit 0
fi

# Build each in dependency order
for module in $CHANGED; do
  ORDER=$(eac get execution order $module | jq -r '.execution_order[]')

  for dep in $ORDER; do
    echo "Building $dep..."
    eac build $dep || exit 1
  done
done
```

## Common Issues

### Invalid JSON Output

**Problem**: `jq` reports invalid JSON

**Solution**: Check command exit code and stderr

```bash
if ! OUTPUT=$(eac get modules 2>&1); then
  echo "Command failed: $OUTPUT"
  exit 1
fi

if ! echo "$OUTPUT" | jq empty 2>/dev/null; then
  echo "Invalid JSON output"
  exit 1
fi
```

### Empty Output

**Problem**: Command returns `{}`

**Solution**: Verify repository state

```bash
# Check if modules exist
TOTAL=$(eac get modules | jq '.modules | length')
if [ "$TOTAL" -eq 0 ]; then
  echo "No modules found in repository"
fi
```

### Performance Issues

**Problem**: `get files` too slow

**Solution**: Use alternatives or cache

```bash
# Avoid: Calling get files repeatedly
for module in $(eac get modules | jq -r '.modules[].moniker'); do
  eac get files | jq ".files[] | select(.module == \"$module\")"
done

# Better: Cache once
eac get files > files.json
for module in $(eac get modules | jq -r '.modules[].moniker'); do
  jq ".files[] | select(.module == \"$module\")" files.json
done

# Best: Use get modules (includes file counts)
eac get modules | jq '.modules[] | "\(.moniker): \(.files) files"'
```

## See Also

- [Show Commands](./show.md) - Human-readable output
- [Output Formats](../overview/output-formats.md) - JSON vs formatted
- [Command Taxonomy](../overview/command-taxonomy.md) - Command organization
