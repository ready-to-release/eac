# Build Changed Modules

## What You'll Accomplish

Build only the modules affected by your changes for efficient CI/CD pipelines.

## Prerequisites

### Required Setup

- Working in git repository with changes
- Module contracts defined

## Steps

### 1. Identify Changed Modules

```bash
r2r eac get changed-modules
```

**What happens**: Returns JSON list of modules affected by local changes

### 2. Get Build Order

```bash
r2r eac get execution-order $(r2r eac get changed-modules | jq -r '.changed_modules[]')
```

**What happens**: Determines correct build order respecting dependencies

### 3. Build Changed Modules

```bash
r2r eac get changed-modules | jq -r '.changed_modules[]' | xargs r2r eac build
```

**What happens**: Builds each changed module in dependency order

## CI Integration

For CI pipelines, use `get changed-modules-ci`:

```bash
# Get modules changed since last successful CI run
r2r eac get changed-modules-ci

# Build only what changed
MODULES=$(r2r eac get changed-modules-ci | jq -r '.changed_modules[]')
r2r eac build $MODULES
```

## Example Scenario

You changed auth module and want to build efficiently:

```bash
# Check what needs building
r2r eac get changed-modules
# {
#   "changed_modules": ["src-auth", "src-api"]
# }

# Get build order (src-auth must build before src-api)
r2r eac get execution-order src-auth src-api
# ["src-auth", "src-api"]

# Build in order
r2r eac build src-auth src-api
# Building src-auth... ✓
# Building src-api... ✓

# Or use CI detection
r2r eac get changed-modules-ci | jq -r '.changed_modules[]' | xargs r2r eac build
```

## GitHub Actions Example

```yaml
- name: Get Changed Modules
  id: changed
  run: |
    MODULES=$(r2r eac get changed-modules-ci | jq -r '.changed_modules | join(" ")')
    echo "modules=$MODULES" >> $GITHUB_OUTPUT

- name: Build Changed Modules
  run: r2r eac build ⟪ steps.changed.outputs.modules ⟫
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "No changes detected" | Ensure commits exist vs base branch |
| Build fails | Check dependencies are built first |

## Next Steps

- [Run Tests for Module](./run-tests-for-module.md) → Test changes

## Related Commands

- [`get changed-modules`](../../../../reference/commands/get/changed-modules.md) - Local changes
- [`get changed-modules-ci`](../../../../reference/commands/get/changed-modules-ci.md) - CI changes
- [`get execution-order`](../../../../reference/commands/get/execution-order.md) - Build order
