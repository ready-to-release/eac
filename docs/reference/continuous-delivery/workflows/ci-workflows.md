# CI Modules

{{ page_breadcrumb() }}

Reference for individual module CI workflows.

## Overview

Each module with build and test requirements has a dedicated CI workflow following a standard pattern. These workflows are invoked by the CI orchestrator (`trigger-ci.yaml`) and can also be triggered manually for testing individual modules.

**Location:** `.github/workflows/ci-{moniker}.yaml`

## Module CI Workflow Pattern

All module CI workflows follow a consistent structure:

### Standard Structure

```yaml
name: "ci-{moniker} (stage 1-7)"

on:
  workflow_call:
    inputs:
      ref:
        required: false
        type: string
        default: ''
      trigger_run_id:
        required: false
        type: string
        default: ''
  workflow_dispatch:
    inputs:
      ref:
        required: false
        type: string
      sha:
        required: false
        type: string
      trigger_run_id:
        required: false
        type: string

permissions:
  contents: read
  packages: read
  actions: read

jobs:
  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          ref: ⟪ inputs.ref || github.ref ⟫

      - name: Build Module
        uses: ./.github/actions/build-module
        with:
          module: {moniker}
          trigger-run-id: ⟪ inputs.trigger_run_id ⟫

  test:
    name: Test
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          ref: ⟪ inputs.ref || github.ref ⟫

      - name: Test Module
        uses: ./.github/actions/test-module
        with:
          module: {moniker}
          trigger-run-id: ⟪ inputs.trigger_run_id ⟫
          suites: ⟪ (inputs.ref != '' && startsWith(inputs.ref, 'refs/pull/')) && 'commit' || 'commit,acceptance' ⟫
```

### Workflow Naming Convention

Workflows are named with stage numbers indicating their position in the overall pipeline:

- `(stage 1-7)` - Standard module CI workflow (build and test)
- `(stage 8+)` - Release workflows (post-CI)

### Common Inputs

#### `ref` (string, optional)

Git reference to checkout. Used by the orchestrator to specify the correct branch/commit.

**Default:** `github.ref` (current reference)

#### `sha` (string, optional)

Commit SHA to test. Primarily used for display and tracking purposes.

**Default:** `github.sha` (current commit)

#### `trigger_run_id` (string, optional)

Run ID of the trigger workflow (`trigger-ci.yaml`). Used to download artifacts created by the trigger workflow.

**Default:** Empty string (builds artifacts locally)

**Usage:** When provided, the workflow downloads the commands binary and other artifacts from the trigger workflow instead of building them.

### Permissions

```yaml
permissions:
  contents: read    # Checkout code
  packages: read    # Download packages if needed
  actions: read     # Download artifacts from other workflows
```

### Job Structure

#### Build Job

**Purpose:** Build the module and produce artifacts.

**Steps:**

1. Checkout repository at specified ref
2. Build module using `.github/actions/build-module`
   - Downloads commands binary from trigger workflow if `trigger_run_id` is provided
   - Builds module using `commands build {moniker}`
   - Uploads build artifacts

**Artifacts Produced:**

- Module-specific build outputs (executables, libraries, containers, sites, etc.)
- Build logs and summaries

#### Test Job

**Purpose:** Run module test suites.

**Dependencies:** Requires `build` job to complete successfully

**Steps:**

1. Checkout repository at specified ref
2. Test module using `.github/actions/test-module`
   - Downloads commands binary and module artifacts from trigger workflow
   - Runs test suites based on context (PR vs trunk)
   - Generates test reports

**Test Suite Selection:**

**For Pull Requests:**

```yaml
suites: commit
```

Only runs commit test suite (fast feedback)

**For Main Branch:**

```yaml
suites: commit,acceptance
```

Runs both commit and acceptance test suites (full validation)

**Detection Logic:**

```yaml
⟪ (inputs.ref != '' && startsWith(inputs.ref, 'refs/pull/')) && 'commit' || 'commit,acceptance' ⟫
```

## Module CI Workflows Inventory

| Workflow                        | Module                | Type              | Stages      | Test Suites        |
| ------------------------------- | --------------------- | ----------------- | ----------- | ------------------ |
| `ci-eac-ai.yaml`                | eac-ai                | Go library        | Build, Test | commit, acceptance |
| `ci-eac-commands.yaml`          | eac-commands          | Go commands       | Build, Test | commit, acceptance |
| `ci-eac-core.yaml`              | eac-core              | Go library        | Build, Test | commit, acceptance |
| `ci-eac-mcp-commands.yaml`      | eac-mcp-commands      | Go MCP server     | Build, Test | commit, acceptance |
| `ci-ext-eac.yaml`               | ext-eac               | Docker extension  | Build, Test | commit, acceptance |
| `ci-r2r-cli.yaml`               | r2r-cli               | Go CLI            | Build, Test | commit, acceptance |
| `ci-vscode-ext-commit.yaml`     | vscode-ext-commit     | VSCode extension  | Build, Test | commit, acceptance |
| `ci-books.yaml`                 | books                 | PDF documentation | Build, Test | commit, acceptance |
| `ci-docs.yaml`                  | docs                  | MkDocs site       | Build, Test | commit, acceptance |
| `ci-scripts-cli-installer.yaml` | scripts-cli-installer | Shell scripts     | Build, Test | commit, acceptance |
| `ci-scripts-implicit-cli.yaml`  | scripts-implicit-cli  | Shell scripts     | Build, Test | commit, acceptance |

## Example: ci-eac-commands.yaml

Complete example of a module CI workflow for the `eac-commands` module:

### Workflow Definition

**File:** `.github/workflows/ci-eac-commands.yaml`

**Module:** eac-commands (Go commands library)

**Type:** go-commands

### Triggers

```yaml
on:
  workflow_call:
    inputs:
      ref:
        required: false
        type: string
        default: ''
      trigger_run_id:
        required: false
        type: string
        description: 'Run ID of trigger workflow (for artifact download)'
        default: ''
  workflow_dispatch:
    inputs:
      ref:
        required: false
        type: string
        description: 'Git ref to checkout'
        default: ''
      sha:
        required: false
        type: string
        description: 'Commit SHA to test'
        default: ''
      trigger_run_id:
        required: false
        type: string
        description: 'Run ID of trigger workflow (for artifact download)'
        default: ''
```

### Build Job

```yaml
build:
  name: Build
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v6
      with:
        ref: ⟪ inputs.ref || github.ref ⟫

    - name: Build Module
      uses: ./.github/actions/build-module
      with:
        module: eac-commands
        trigger-run-id: ⟪ inputs.trigger_run_id ⟫
```

**Build Action Behavior:**

1. Downloads commands binary from `trigger_run_id` if provided
2. Sets up Go environment (Go 1.21+)
3. Runs `commands build eac-commands`
4. Uploads build artifacts

### Test Job

```yaml
test:
  name: Test
  needs: build
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v6
      with:
        ref: ⟪ inputs.ref || github.ref ⟫

    - name: Test Module
      uses: ./.github/actions/test-module
      with:
        module: eac-commands
        trigger-run-id: ⟪ inputs.trigger_run_id ⟫
        # PRs run commit suite only; main/release branches run full suite
        suites: ⟪ (inputs.ref != '' && startsWith(inputs.ref, 'refs/pull/')) && 'commit' || 'commit,acceptance' ⟫
```

**Test Action Behavior:**

1. Downloads commands binary and build artifacts from `trigger_run_id`
2. Sets up test environment
3. Runs test suites:
   - **PR context:** `r2r eac test eac-commands --suite commit`
   - **Trunk context:** `r2r eac test eac-commands --suite commit,acceptance`
4. Uploads test results and reports

## Artifact Handling

### Artifact Download

When `trigger_run_id` is provided, workflows download artifacts from the trigger workflow:

```yaml
- name: Download commands binary
  uses: actions/download-artifact@v6
  with:
    name: commands-binary
    path: out/tools
    run-id: ⟪ inputs.trigger_run_id ⟫
```

### Artifact Upload

Build artifacts are uploaded for use by dependent workflows:

```yaml
- name: Upload build artifacts
  uses: actions/upload-artifact@v6
  with:
    name: build-{moniker}
    path: out/build/{moniker}/
```

## Status Checks

Each module CI workflow reports a status check to the commit:

**Check Name:** `CI: {module-name}`

**Examples:**

- `CI: eac-commands`
- `CI: r2r-cli`
- `CI: docs`

**Status Values:**

- **Success:** Build and all tests passed
- **Failure:** Build failed or tests failed
- **Cancelled:** Workflow was cancelled
- **Skipped:** Workflow was skipped (no changes detected)

## Manual Workflow Dispatch

Workflows can be triggered manually for individual module testing:

```bash
# Trigger CI for eac-commands module
gh workflow run ci-eac-commands.yaml

# Trigger with specific ref
gh workflow run ci-eac-commands.yaml \
  -f ref=refs/heads/feature-branch

# Trigger with commit SHA
gh workflow run ci-eac-commands.yaml \
  -f sha=abc123def456
```

## Debugging Module CI

### View Workflow Runs

```bash
# List recent runs for a module
gh run list --workflow ci-eac-commands.yaml --limit 10

# View specific run
gh run view <run-id>

# View failed step logs
gh run view <run-id> --log-failed
```

### Check Module Dependencies

```bash
# Get module dependencies
r2r eac get dependencies eac-commands

# Get modules that depend on this module
r2r eac get dependencies eac-commands --reverse
```

### Test Locally

```bash
# Build module locally
r2r eac build eac-commands

# Test module locally (commit suite)
r2r eac test eac-commands --suite commit

# Test module locally (all suites)
r2r eac test eac-commands --suite commit,acceptance
```

## References

- [Trigger Orchestration](./trigger-orchestration.md) - How modules are orchestrated
- [Overview](./overview.md) - Workflow architecture
- Reusable actions: `.github/actions/build-module`, `.github/actions/test-module`
- Module contracts: `.r2r/eac/repository.yml`

{{ diataxis_footer() }}
