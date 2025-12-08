<!-- EDITOR
# Editor: how-to-guides/commands/areas/pipeline-overview.md

## Soul

Dependency-aware pipeline orchestration that executes modules in layers based on dependency graph, maximizing parallelism within layers while respecting build order.

## Sections

1. What is Pipeline Orchestration?
2. When to Use Pipeline Commands
3. Key Concepts
   - Module Dependencies
   - Execution Layers
   - Dependency Graph
   - Pipeline Execution
4. Workflow Overview
   - Local Pipeline Execution
   - CI/CD Integration
   - Waiting for Workflows
5. Execution Modes
   - Build Pipeline
   - Test Pipeline
   - Full Pipeline
6. Pipeline Status
   - Viewing Status
   - Status States
7. Integration Points
   - With GitHub Actions
   - With Module Contracts
   - With Build/Test Commands
   - With CI Summary
8. Advanced Usage
   - Custom Execution Order
   - Parallel Job Matrix
   - Changed Module Pipeline
9. Best Practices
10. Troubleshooting
11. CI Commands
12. Related Areas
-->

# Pipeline Orchestration

Pipeline orchestration in EAC provides dependency-aware execution of modules, enabling efficient CI/CD workflows that respect module relationships and maximize parallelism.

## What is Pipeline Orchestration?

EAC's pipeline system enables you to:

- **Execute modules in dependency order** - Build dependencies first
- **Maximize parallelism** - Run independent modules concurrently
- **Integrate with GitHub Actions** - Monitor and wait for workflows
- **Track pipeline status** - View CI state across modules

The system reads module contracts to determine execution order and groups modules into parallelizable layers.

## When to Use Pipeline Commands

Use pipeline commands when you need:

| Scenario                | Commands               |
| ----------------------- | ---------------------- |
| Build modules in order  | `pipeline run`         |
| Wait for CI to complete | `pipeline wait`        |
| Check CI status         | `pipeline status`      |

### Common Use Cases

- **Local development** - Build/test with correct dependency order
- **CI/CD workflows** - Orchestrate GitHub Actions jobs
- **Release preparation** - Ensure all modules build before release
- **Debugging** - Understand why a module failed

## Key Concepts

### Module Dependencies

Modules declare dependencies in contracts:

```yaml
# modules.yml
modules:
  - moniker: eac-commands
    depends_on:
      - eac-core
      - eac-ai

  - moniker: r2r-cli
    depends_on:
      - eac-commands
```

### Execution Layers

Dependencies determine execution layers:

```text
Layer 0: [eac-core, eac-ai]        # No dependencies
    │
    ▼
Layer 1: [eac-commands]            # Depends on Layer 0
    │
    ▼
Layer 2: [r2r-cli]                 # Depends on Layer 1
```

Modules in the same layer run in parallel.

### Dependency Graph

```text
                    ┌─────────────┐
                    │   r2r-cli   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │eac-commands │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌─────────┐  ┌─────────┐  ┌─────────┐
        │eac-core │  │ eac-ai  │  │eac-mcp  │
        └─────────┘  └─────────┘  └─────────┘
```

### Pipeline Execution

When running a pipeline:

1. **Analysis** - Build dependency graph from contracts
2. **Layering** - Group modules into parallel layers
3. **Execution** - Run each layer, wait for completion
4. **Reporting** - Aggregate results across modules

## Workflow Overview

### Local Pipeline Execution

```bash
# 1. View dependencies
r2r eac show dependencies

# 2. Get execution order for specific module
r2r eac get execution order r2r-cli

# 3. Run pipeline for module and dependencies
r2r eac pipeline run r2r-cli

# 4. Or run full pipeline
r2r eac pipeline run
```

### CI/CD Integration

```yaml
# GitHub Actions workflow
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run pipeline
        run: r2r eac pipeline run

      - name: Check status
        run: r2r eac pipeline status
```

### Waiting for Workflows

```bash
# Wait for current workflow runs to complete
r2r eac pipeline wait

# Wait with timeout
r2r eac pipeline wait --timeout 30m

# Wait for specific workflow
r2r eac pipeline wait --workflow ci
```

## Execution Modes

### Build Pipeline

```bash
# Build all modules in dependency order
r2r eac pipeline run --action build

# Build specific module and its dependencies
r2r eac pipeline run --action build r2r-cli
```

### Test Pipeline

```bash
# Test all modules
r2r eac pipeline run --action test

# Test with coverage
r2r eac pipeline run --action test --coverage
```

### Full Pipeline

```bash
# Build + Test + Validate
r2r eac pipeline run --action full

# Equivalent to:
r2r eac pipeline run --action build
r2r eac pipeline run --action test
r2r eac validate
```

## Pipeline Status

### Viewing Status

```bash
r2r eac pipeline status

# Output:
# Pipeline Status for main (commit abc123)
#
# | Module        | Build  | Test   | Overall |
# |---------------|--------|--------|---------|
# | eac-core      | ✅     | ✅     | ✅      |
# | eac-ai        | ✅     | ✅     | ✅      |
# | eac-commands  | ✅     | ⏳     | ⏳      |
# | r2r-cli       | ⏳     | -      | ⏳      |
#
# Legend: ✅ Passed  ❌ Failed  ⏳ Running  - Pending
```

### Status States

| State   | Symbol | Meaning                   |
| ------- | ------ | ------------------------- |
| Passed  | ✅     | All checks passed         |
| Failed  | ❌     | One or more checks failed |
| Running | ⏳     | Currently executing       |
| Pending | -      | Waiting for dependencies  |
| Skipped | ⊘      | Explicitly skipped        |

## Integration Points

### With GitHub Actions

Pipeline commands integrate with GitHub's API:

```bash
# Check workflow status
r2r eac pipeline status --repo owner/repo

# Wait for workflows
r2r eac pipeline wait --repo owner/repo --run-id 12345

# Trigger workflow
gh workflow run ci.yml
r2r eac pipeline wait
```

### With Module Contracts

Execution order derives from contracts:

```yaml
# modules.yml determines:
# - Which modules exist
# - Their dependencies
# - Build/test commands per type
```

### With Build/Test Commands

Pipeline executes standard commands:

```bash
# Pipeline internally runs:
r2r eac build <module>    # For build action
r2r eac test <module>     # For test action
```

### With CI Summary

Generate diagnostic on failure:

```bash
# On pipeline failure
r2r eac pipeline ci summary-link $RUN_ID --type build
```

## Advanced Usage

### Custom Execution Order

```bash
# Get execution order without running
r2r eac get execution order r2r-cli

# Output:
# Layer 0: eac-core, eac-ai
# Layer 1: eac-commands
# Layer 2: r2r-cli

# Run specific layers
r2r eac build eac-core eac-ai       # Layer 0
r2r eac build eac-commands          # Layer 1
r2r eac build r2r-cli               # Layer 2
```

### Parallel Job Matrix

```yaml
# Generate matrix from dependencies
jobs:
  prepare:
    outputs:
      matrix: ${{ steps.matrix.outputs.layers }}
    steps:
      - id: matrix
        run: |
          LAYERS=$(r2r eac get execution order --json)
          echo "layers=$LAYERS" >> $GITHUB_OUTPUT

  build:
    needs: prepare
    strategy:
      matrix: ${{ fromJson(needs.prepare.outputs.matrix) }}
```

### Changed Module Pipeline

```bash
# Only run pipeline for changed modules
CHANGED=$(r2r eac get changed-modules)
r2r eac pipeline run $CHANGED
```

## Best Practices

### Do's

- **Declare all dependencies** - Ensures correct execution order
- **Use layers for parallelism** - Faster CI/CD execution
- **Monitor pipeline status** - Catch failures early
- **Wait for completion** - Don't proceed without green CI

### Don'ts

- **Don't create cycles** - Circular dependencies break pipelines
- **Don't skip dependencies** - Build order matters
- **Don't ignore failures** - Fix before proceeding

## Troubleshooting

| Problem                    | Solution                        |
| -------------------------- | ------------------------------- |
| Module builds out of order | Check `depends_on` in contracts |
| Circular dependency error  | Review and break the cycle      |
| Pipeline hangs             | Check `pipeline wait` timeout   |
| Status not updating        | Verify GitHub API access        |

## Next Steps

- [Pipeline Configuration](pipeline-configuration.md) - Configure workflows and timeouts
- [Pipeline Commands](pipeline-commands.md) - Full command reference

## CI Commands

The pipeline CI commands provide workflow orchestration and diagnostics:

| Command | Description |
|---------|-------------|
| `pipeline ci dispatch-and-wait` | Dispatch workflow and wait for completion |
| `pipeline ci summary-link` | Generate CI diagnostic markdown |

See [Pipeline CI Commands](pipeline-ci-commands.md) for detailed documentation.

## Related Areas

- [Build Commands](build-overview.md) - Individual module build
- [Test Commands](test-overview.md) - Individual module test
- [Pipeline CI Commands](pipeline-ci-commands.md) - CI diagnostics and summaries
- [Validate Commands](validate-overview.md) - Dependency graph validation
