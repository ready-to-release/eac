# Pipeline Configuration

This guide covers configuration options for EAC's pipeline orchestration system, including dependency settings, execution modes, and GitHub Actions integration.

## Configuration Files

| File                           | Purpose                  |
| ------------------------------ | ------------------------ |
| `.r2r/eac/pipeline/config.yml` | Pipeline settings        |
| `modules.yml`                  | Module dependencies      |
| `.github/workflows/`           | GitHub Actions workflows |

## Pipeline Settings

### Basic Configuration

`.r2r/eac/pipeline/config.yml`:

```yaml
# Execution settings
execution:
  # Maximum parallel jobs
  max_parallel: 4

  # Default timeout per module
  timeout: 30m

  # Continue on failure
  continue_on_error: false

  # Retry failed modules
  retry:
    enabled: true
    max_attempts: 2
    delay: 10s

# Dependency settings
dependencies:
  # Include transitive dependencies
  transitive: true

  # Fail if dependency fails
  fail_on_dependency_failure: true

  # Skip unchanged modules
  skip_unchanged: false

# Output settings
output:
  # Log directory
  log_dir: out/pipeline/

  # Log level: debug, info, warn, error
  log_level: info

  # Save execution report
  save_report: true
```

## Module Dependencies

### Defining Dependencies

In `modules.yml`:

```yaml
modules:
  - moniker: eac-core
    type: go-library
    # No dependencies - Layer 0

  - moniker: eac-ai
    type: go-library
    depends_on:
      - eac-core  # Layer 1

  - moniker: eac-commands
    type: go-commands
    depends_on:
      - eac-core
      - eac-ai    # Layer 2

  - moniker: r2r-cli
    type: go-cli
    depends_on:
      - eac-commands  # Layer 3
```

### Dependency Graph

```text
Layer 0: eac-core
           │
           ▼
Layer 1: eac-ai ──────┐
           │          │
           ▼          ▼
Layer 2: eac-commands
           │
           ▼
Layer 3: r2r-cli
```

### Execution Order

```bash
# View execution order
r2r eac get-execution-order r2r-cli

# Output:
# Layer 0: eac-core
# Layer 1: eac-ai
# Layer 2: eac-commands
# Layer 3: r2r-cli
```

## Execution Configuration

### Parallel Execution

```yaml
execution:
  # Maximum concurrent modules
  max_parallel: 4

  # Per-layer parallelism
  layer_parallel: true

  # Cross-layer parallelism (risky)
  cross_layer_parallel: false
```

### Timeout Settings

```yaml
execution:
  # Global timeout
  timeout: 30m

  # Per-module overrides
  timeouts:
    eac-core: 10m
    r2r-cli: 45m
    docs: 15m
```

### Retry Configuration

```yaml
execution:
  retry:
    enabled: true

    # Maximum retry attempts
    max_attempts: 2

    # Delay between retries
    delay: 10s

    # Exponential backoff
    backoff: true
    backoff_multiplier: 2

    # Only retry on specific errors
    retry_on:
      - timeout
      - network_error
```

### Error Handling

```yaml
execution:
  # Stop on first failure
  continue_on_error: false

  # Or continue and collect all failures
  # continue_on_error: true

  # Fail entire pipeline if any module fails
  fail_fast: true

  # Or allow partial success
  # fail_fast: false
```

## Pipeline Actions

### Build Pipeline

```yaml
actions:
  build:
    # Pre-build commands
    pre:
      - "r2r eac validate contracts"

    # Build command (per module type)
    commands:
      go-cli: "go build -o {output} ./..."
      go-library: "go build ./..."
      mkdocs-site: "mkdocs build"

    # Post-build commands
    post:
      - "r2r eac validate dependencies"
```

### Test Pipeline

```yaml
actions:
  test:
    # Test command per module type
    commands:
      go-cli: "go test -v ./..."
      go-library: "go test -v ./..."
      go-tests: "godog run"

    # Coverage settings
    coverage:
      enabled: true
      output: out/reports/coverage/
      threshold: 80

    # Test output format
    output_format: junit
```

### Full Pipeline

```yaml
actions:
  full:
    # Sequence of actions
    sequence:
      - validate
      - build
      - test
      - security

    # Stop on first action failure
    fail_fast: true
```

## GitHub Actions Integration

### Workflow Configuration

```yaml
# .github/workflows/ci.yml
name: CI Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  pipeline:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run Pipeline
        run: r2r eac run-pipeline

      - name: Upload Reports
        uses: actions/upload-artifact@v4
        with:
          name: pipeline-reports
          path: out/pipeline/
```

### Matrix Strategy

```yaml
jobs:
  prepare:
    runs-on: ubuntu-latest
    outputs:
      matrix: ${{ steps.set-matrix.outputs.matrix }}
    steps:
      - uses: actions/checkout@v4
      - id: set-matrix
        run: |
          LAYERS=$(r2r eac get-execution-order --json)
          echo "matrix=$LAYERS" >> $GITHUB_OUTPUT

  build:
    needs: prepare
    runs-on: ubuntu-latest
    strategy:
      matrix:
        layer: ${{ fromJson(needs.prepare.outputs.matrix) }}
    steps:
      - uses: actions/checkout@v4
      - name: Build Layer ${{ matrix.layer.index }}
        run: r2r eac build ${{ join(matrix.layer.modules, ' ') }}
```

### Pipeline Status Check

```yaml
- name: Wait for Pipeline
  run: r2r eac pipeline-wait --timeout 30m

- name: Check Status
  run: r2r eac show-pipeline-status
```

## Pipeline Wait Configuration

```yaml
# Wait settings
wait:
  # Default timeout
  timeout: 30m

  # Poll interval
  interval: 30s

  # Workflows to wait for
  workflows:
    - ci
    - test

  # Check specific run ID
  run_id: null  # or specific ID

  # Fail on workflow failure
  fail_on_workflow_failure: true
```

### Wait Command Options

```bash
# Wait with default settings
r2r eac pipeline-wait

# Custom timeout
r2r eac pipeline-wait --timeout 1h

# Specific workflow
r2r eac pipeline-wait --workflow ci

# Specific run
r2r eac pipeline-wait --run-id 12345678
```

## Status Configuration

```yaml
# Status display settings
status:
  # Refresh interval for live display
  refresh: 5s

  # Show detailed module status
  detailed: true

  # Include timing information
  show_timing: true

  # Color output
  color: true
```

### Status Output

```bash
r2r eac show-pipeline-status

# Output:
# Pipeline Status for main (abc1234)
# ══════════════════════════════════
#
# │ Module        │ Build │ Test  │ Time   │ Status │
# ├───────────────┼───────┼───────┼────────┼────────┤
# │ eac-core      │ ✅    │ ✅    │ 45s    │ ✅     │
# │ eac-ai        │ ✅    │ ✅    │ 1m 12s │ ✅     │
# │ eac-commands  │ ✅    │ ⏳    │ 2m 30s │ ⏳     │
# │ r2r-cli       │ ⏳    │ -     │ -      │ ⏳     │
#
# Overall: Running (2/4 complete)
```

## Changed Modules Configuration

```yaml
# Changed module detection
changes:
  # Base reference for comparison
  base_ref: main

  # Include dependencies of changed modules
  include_dependents: true

  # File patterns to consider
  patterns:
    go-*:
      - "**/*.go"
      - "go.mod"
      - "go.sum"
    mkdocs-site:
      - "docs/**/*.md"
      - "mkdocs.yml"

  # Always include these modules
  always_include: []

  # Never include these modules
  exclude:
    - contracts
```

### Changed Module Commands

```bash
# Get changed modules
r2r eac get-changed-modules

# Get changed modules for CI
r2r eac get-changed-modules-ci

# Run pipeline for changed only
CHANGED=$(r2r eac get-changed-modules)
r2r eac run-pipeline $CHANGED
```

## Environment Variables

| Variable                | Description             | Default |
| ----------------------- | ----------------------- | ------- |
| `PIPELINE_MAX_PARALLEL` | Max parallel jobs       | `4`     |
| `PIPELINE_TIMEOUT`      | Global timeout          | `30m`   |
| `PIPELINE_LOG_LEVEL`    | Log verbosity           | `info`  |
| `GITHUB_TOKEN`          | GitHub API token        | -       |
| `GITHUB_REPOSITORY`     | Repository (owner/repo) | -       |

## Example Configurations

### Minimal Configuration

```yaml
execution:
  max_parallel: 4
  timeout: 30m

dependencies:
  transitive: true
```

### CI/CD Configuration

```yaml
execution:
  max_parallel: 8
  timeout: 45m
  continue_on_error: false
  retry:
    enabled: true
    max_attempts: 2

dependencies:
  transitive: true
  fail_on_dependency_failure: true

actions:
  full:
    sequence:
      - validate
      - build
      - test
      - security

wait:
  timeout: 60m
  workflows:
    - ci
    - security

output:
  log_dir: out/pipeline/
  save_report: true
```

### Development Configuration

```yaml
execution:
  max_parallel: 2
  timeout: 15m
  continue_on_error: true

dependencies:
  transitive: true
  skip_unchanged: true

changes:
  include_dependents: false

output:
  log_level: debug
```

## Troubleshooting

| Issue                    | Cause               | Solution               |
| ------------------------ | ------------------- | ---------------------- |
| Modules run out of order | Missing dependency  | Add to `depends_on`    |
| Pipeline hangs           | Timeout too short   | Increase timeout       |
| Circular dependency      | Invalid graph       | Check module contracts |
| Status not updating      | API rate limit      | Check GitHub token     |
| Parallel failures        | Resource contention | Reduce `max_parallel`  |

## Related Documentation

- [Pipeline Overview](pipeline-overview.md) - Concepts and workflows
- [Pipeline Commands](pipeline-commands.md) - Command reference
- [Build Commands](build-overview.md) - Build module operations
- [Test Commands](test-overview.md) - Test module operations
