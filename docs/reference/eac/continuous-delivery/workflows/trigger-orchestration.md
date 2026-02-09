# Trigger Orchestration

Detailed reference for the trigger orchestration system (`change-trigger.yaml`).

## Overview

The `change-trigger.yaml` workflow implements incremental CI by detecting which modules have changed and orchestrating their build and test execution.

This workflow serves as the main entry point for continuous integration and handles all release triggering.

On push to main, after CI workflows complete, the workflow checks for pending releases from two sources:

- **Semver modules**: Changelog versions without corresponding git tags (clie, eac-ext)
- **Calver modules**: Modules that had CI dispatched and auto-release on every push (docs, books)

Releases are triggered in dependency order, with each layer completing before the next begins.

**Location:** `.github/workflows/change-trigger.yaml`

## Triggers

```yaml
on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
  workflow_dispatch:
    inputs:
      trigger-all:
        description: 'Trigger all workflows regardless of changes'
        required: false
        type: boolean
        default: false
```

The workflow runs automatically on pushes to main and pull requests targeting main. It can also be triggered manually with an option to bypass change detection.

## Permissions

```yaml
permissions:
  actions: write    # Required to dispatch child workflows
  contents: read    # Required to checkout code
```

## Pipeline Architecture

The workflow executes in multiple stages with two parallel flows:

```text
Stage 0: Build CI Tooling
         ↓
    ┌────┴────┐
    ↓         ↓
Flow A     Flow B
(Repo)     (Modules)
    ↓         ↓
    └────┬────┘
         ↓
    Summary
```

### Stage 0: Build CI Tooling

**Job:** `build-tooling`

Builds and caches the commands binary for use by all subsequent jobs.

**Steps:**

1. Checkout repository
2. Setup commands binary using `.github/actions/setup-commands`
3. Generate build summary

**Outputs:**

- Uploads `commands-binary` artifact for reuse

### Flow A: Repository Validation (Parallel)

Validates repository-level contracts and structure.

#### Job: `build-repository`

Builds the repository module.

**Dependencies:** `build-tooling`

**Steps:**

1. Checkout repository
2. Build repository module using `.github/actions/build-module`

#### Job: `validate-repository`

Runs repository commit test suite.

**Dependencies:** `build-repository`

**Steps:**

1. Checkout repository
2. Pre-compute git file list via GitHub API (optimization to avoid slow `git ls-files`)
3. Test repository module using `.github/actions/test-module` with `commit` suite

**Optimization:** Uses GitHub API to fetch file list (~2-5s) instead of `git ls-files` (~84s).

### Flow B: Module CI Orchestration (Parallel)

Detects changes, calculates execution order, and dispatches module CI workflows.

#### Job: `detect-changes`

Determines which modules need to be rebuilt based on file changes.

**Dependencies:** `build-tooling`

**Steps:**

1. Checkout repository with full history (`fetch-depth: 0`)
2. Download commands binary artifact
3. Detect changed modules using change detection logic (see below)
4. Filter to modules with CI workflows (`.github/workflows/ci-{moniker}.yaml`)
5. Generate step summary showing detected changes

**Outputs:**

- `changed-modules` - Space-separated list of module monikers
- `has-changes` - Boolean indicating if any modules changed
- `directly_changed` - Modules with direct file changes
- `invalidated` - Dependent modules triggered by dependency changes
- `base_sha` - Base commit SHA used for comparison
- `is_bootstrap` - Boolean indicating bootstrap mode (no previous CI)
- `changed_file_count` - Number of files changed
- `files_by_module` - JSON object mapping modules to changed files

**Change Detection Logic:**

**Manual Override:**

```bash
if [ "⟪ inputs.trigger-all ⟫" = "true" ]; then
  # Build all modules with CI workflows
fi
```

**Pull Request:**

```bash
# Compare against PR base SHA
RESULT=$(commands get changed-modules-ci --pr-base "⟪ github.event.pull_request.base.sha ⟫" --as-json)
```

**Push to Main:**

```bash
# Compare against last successful CI run (uses CIRunQuerier via GitHub API)
RESULT=$(commands get changed-modules-ci --as-json)
```

**Bootstrap Mode:**

- Triggered when no previous successful CI run exists
- Builds all modules

**CI Cache Filtering:**

- After detecting changed modules, `get ci-dispatch` uses a `CICacheChecker` to
  skip modules that already have a successful CI build at the current HEAD SHA
- This avoids re-dispatching CI for modules whose source hasn't changed
- Bypass with `--skip-cache=remote:ci` to force all dispatches

**Dependency Invalidation:**

- When a module changes, all modules that depend on it are also marked for rebuild
- Ensures dependent modules are rebuilt when their dependencies change (cache invalidation)

#### Job: `calculate-order`

Computes dependency-based execution order for changed modules.

**Dependencies:** `detect-changes`

**Steps:**

1. Checkout repository
2. Download commands binary artifact
3. Calculate execution order using dependency graph
4. Generate execution plan (modules grouped into layers)
5. Generate step summary showing layer structure

**Command:**

```bash
PLAN=$(commands get execution-order $MODULES --skip-depm --as-json)
```

**Outputs:**

- `execution-plan` - JSON object with layered execution plan
- `layer-count` - Number of dependency layers

**Execution Plan Structure:**

```json
{
  "layer_count": 3,
  "layers": [
    ["eac-core", "eac-ai"],
    ["eac-commands", "eac-mcp-commands"],
    ["clie"]
  ]
}
```

**Layer explanation:**

- Layer 1: `eac-core`, `eac-ai` - No dependencies
- Layer 2: `eac-commands`, `eac-mcp-commands` - Depend on Layer 1
- Layer 3: `clie` - Depends on Layer 2

#### Job: `execute-layers`

Dispatches module CI workflows in dependency order.

**Dependencies:** `detect-changes`, `calculate-order`

**Steps:**

1. Checkout repository
2. Download commands binary artifact
3. Execute workflows based on execution mode (PR vs trunk)
4. Generate step summary

**Execution Modes:**

**Pull Request Mode:**

```bash
# Run all workflows in parallel for fast feedback
for module in $ALL_MODULES; do
  gh workflow run "ci-${module}.yaml" \
    --ref "$DISPATCH_REF" \
    -f ref=⟪ github.ref ⟫ \
    -f sha=⟪ github.sha ⟫ \
    -f trigger_run_id=⟪ github.run_id ⟫
done
# Child workflows report their own statuses
```

**Trunk Mode (Main Branch):**

```bash
# Run in dependency layers
for layer_idx in 0..$LAYER_COUNT; do
  # Dispatch all workflows in this layer
  for module in $LAYER_MODULES; do
    gh workflow run "ci-${module}.yaml" ...
  done

  # Wait for all workflows in layer to complete
  commands pipeline wait $LAYER_RUN_IDS --timeout 3600

  # Exit if layer fails
  if [ $? -ne 0 ]; then
    exit 1
  fi
done
```

**Key Differences:**

- **PR mode:** All modules run in parallel, no waiting, commit tests only
- **Trunk mode:** Sequential layers with waiting, full test suite (commit + acceptance)

### Summary Stage

**Job:** `summary`

Generates comprehensive summary of CI run.

**Dependencies:** All previous jobs (waits for both flows)

**Condition:** `always()` - Runs even if previous jobs fail

**Steps:**

1. Check build-tooling status
2. Check repository validation status
3. Check module dispatch status
4. Generate step summary with diagnostic information

**Summary Content:**

- Trigger information (event, branch, commit)
- Pipeline stage results (tooling, repository, modules)
- No-change explanation (if applicable)
- Validation passed confirmation
- Link to child workflow checks

## Change Detection Details

### For Pull Requests

Compares PR head against PR base to determine what the PR introduces:

```bash
commands get changed-modules-ci --pr-base "$PR_BASE_SHA" --as-json
```

### For Push to Main

Compares current commit against last successful CI run:

```bash
commands get changed-modules-ci --as-json
```

The command uses a `CIRunQuerier` (backed by the GitHub Actions API) to find the
last successful run of `change-trigger.yaml` on the main branch.

### CI Cache Integration

Module CI dispatch uses the core cache system's **CI cache** (`remote:ci`) to
determine which modules already have a valid CI build at the current HEAD SHA.

The `get ci-dispatch` command creates a `CICacheChecker` that:

1. Queries GitHub Actions for each module's last successful `ci-{module}.yaml` run
2. Compares the run's HEAD SHA against the current commit SHA
3. Skips dispatch for modules where the SHAs match (CI-cached)

This is a first-class cache type in the 2D taxonomy. It can be bypassed with:

```bash
commands get ci-dispatch --skip-cache=remote:ci  # Force dispatch all modules
```

See [Cache System](../../architecture/cache-system.md#ci-cache-architecture) for
the full CI cache architecture including the port/adapter pattern.

### Bootstrap Mode

When no previous successful CI run is found:

- `is_bootstrap` is set to `true`
- All modules are marked for rebuild
- Establishes baseline for future incremental builds

### Dependency Invalidation

The change detection includes transitive invalidation:

1. **Directly changed modules:** Modules with modified files
2. **Invalidated modules:** Modules that depend on directly changed modules

**Example:**

```text
eac-core (changed) → eac-commands (invalidated) → clie (invalidated)
```

If `eac-core` changes, both `eac-commands` and `clie` are rebuilt.

## Workflow Dispatch Parameters

### `trigger-all` (boolean)

Bypasses change detection and triggers all module CI workflows.

**Use cases:**

- Testing full build after infrastructure changes
- Periodic full validation
- Debugging change detection logic

**Behavior:**

- Discovers all `ci-*.yaml` workflows dynamically
- Dispatches all discovered workflows
- Skips change detection and dependency analysis

## Environment Variables

- `GH_TOKEN` - GitHub token for workflow dispatch and API calls (automatically provided)
- `GITHUB_TOKEN` - GitHub token for checkout and artifact operations (automatically provided)

## Artifacts

### Produced

- `commands-binary` - Cached commands executable for reuse

### Consumed

- None (creates initial artifacts)

## Step Summaries

The workflow generates rich GitHub step summaries:

### Change Detection Summary

Shows:

- Base SHA or bootstrap mode
- Files changed count
- Directly changed modules
- Invalidated (dependent) modules
- CI workflows to run
- Per-module reasoning (why each module was triggered)

### Execution Order Summary

Shows:

- Execution mode (PR parallel vs trunk layered)
- Layer structure with module groupings
- Explanation of execution strategy

### Final Summary

Shows:

- Pipeline stage results (pass/fail)
- Total workflows dispatched
- Link to child workflow status checks
- Troubleshooting information (if no changes)

## Performance Optimizations

1. **Artifact caching:** Commands binary built once and reused
2. **GitHub API file list:** Uses fast API instead of slow `git ls-files`
3. **Incremental builds:** Only builds changed modules and their dependents
4. **Parallel execution:** Modules in same layer run concurrently

## Failure Modes

### Build Tooling Failure

If `build-tooling` fails:

- All downstream jobs are skipped
- Summary explains tooling build failure

### Repository Validation Failure

If `build-repository` or `validate-repository` fails:

- Module workflows are still dispatched (validation is independent)
- Summary shows repository validation failure

### Module Dispatch Failure

If `execute-layers` fails to dispatch workflows:

- Summary shows dispatch failure
- Indicates orchestration problem (not module failure)

### No Changes Detected

If no modules need rebuilding:

- Summary explains why (base SHA, file count, affected modules)
- Workflow succeeds (no work to do)

## References

- Workflow file: `.github/workflows/change-trigger.yaml`
- [Overview](./overview.md) - Workflow architecture
- [CI Workflows](./ci-workflows.md) - Module CI workflow patterns
- [Repository Layout](../../architecture/repository-layout.md) - Module structure
