# get changed-modules-local

{{ page_breadcrumb() }}

## Overview

**Command**: `r2r eac get changed-modules-local [module...]`
**Purpose**: Get modules requiring rebuild based on local build state
**Category**: [get](../categories/get.md)

## Description

Detects which modules need rebuilding by checking local build state cache. Unlike `get changed-modules` which checks git changes, this command uses build artifacts and timestamps to determine what needs rebuilding.

The command compares module source files against the last successful build state to identify:

- Modules that have changed and need rebuilding
- Modules that are up-to-date
- Specific reasons why each module needs rebuilding
- Whether this is a fresh build (no previous build state exists)

## Syntax

```bash
r2r eac get changed-modules-local [module...]
```

### Parameters

- `module...` - Optional module monikers to check. If not specified, all modules are checked.

### Output Format

JSON output includes:

```json
{
  "modules": ["module1", "module2"],
  "up_to_date": ["module3", "module4"],
  "change_reasons": {
    "module1": "source files modified",
    "module2": "dependency changed"
  },
  "is_fresh_build": false,
  "detection_time": "2.5ms"
}
```

## Examples

### Check All Modules

```bash
# Get all modules requiring rebuild
r2r eac get changed-modules-local | jq '.'
```

### Check Specific Modules

```bash
# Check only specific modules
r2r eac get changed-modules-local src-auth src-api
```

### Extract Module Lists

```bash
# Get list of modules needing rebuild
r2r eac get changed-modules-local | jq -r '.modules[]'

# Get list of up-to-date modules
r2r eac get changed-modules-local | jq -r '.up_to_date[]'

# Check change reasons
r2r eac get changed-modules-local | jq '.change_reasons'
```

### Build Only Changed Modules

```bash
# Build only modules that need it
CHANGED=$(r2r eac get changed-modules-local | jq -r '.modules[]')
if [ -n "$CHANGED" ]; then
  for module in $CHANGED; do
    r2r eac build $module
  done
fi
```

### Conditional Build Logic

```bash
# Check if any modules need building
RESULT=$(r2r eac get changed-modules-local)
IS_FRESH=$(echo "$RESULT" | jq -r '.is_fresh_build')
CHANGED_COUNT=$(echo "$RESULT" | jq -r '.modules | length')

if [ "$IS_FRESH" = "true" ]; then
  echo "Fresh build - building all modules"
  r2r eac build
elif [ "$CHANGED_COUNT" -gt 0 ]; then
  echo "Building $CHANGED_COUNT changed modules"
  echo "$RESULT" | jq -r '.modules[]' | xargs -L1 r2r eac build
else
  echo "All modules up-to-date"
fi
```

## Use Cases

### Local Development

Use this command during local development to avoid unnecessary rebuilds:

```bash
# Only rebuild what changed since last build
r2r eac get changed-modules-local | jq -r '.modules[]' | xargs -L1 r2r eac build
```

### Pre-Commit Hook

Integrate into pre-commit hooks to validate only affected modules:

```bash
# Validate only modules that changed locally
CHANGED=$(r2r eac get changed-modules-local | jq -r '.modules[]')
for module in $CHANGED; do
  r2r eac validate artifacts $module
done
```

### Build Status Reporting

Report on build freshness and required work:

```bash
# Generate build status report
RESULT=$(r2r eac get changed-modules-local)
echo "Modules needing rebuild: $(echo "$RESULT" | jq -r '.modules | length')"
echo "Up-to-date modules: $(echo "$RESULT" | jq -r '.up_to_date | length')"
echo "$RESULT" | jq -r '.change_reasons | to_entries[] | "\(.key): \(.value)"'
```

## Comparison with Related Commands

| Command | Purpose | Use When |
|---------|---------|----------|
| `get changed-modules-local` | Modules needing rebuild based on build state | Local development, incremental builds |
| `get changed-modules` | Modules affected by git changes | Working with uncommitted changes |
| `get changed-modules-ci` | Modules needing rebuild in CI | CI/CD pipelines, comparing against base commit |

## See Also

- [get changed-modules](./changed-modules.md) - Git-based change detection
- [get changed-modules-ci](./changed-modules-ci.md) - CI pipeline change detection
- [build](../other/build.md) - Build modules
- [validate artifacts](../validate/artifacts.md) - Validate build outputs

{{ diataxis_footer() }}
