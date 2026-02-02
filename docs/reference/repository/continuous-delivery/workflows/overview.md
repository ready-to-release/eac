# Workflow Architecture

Architectural overview of the GitHub Actions workflow system.

## Workflow Architecture

The workflow system is designed around three core architectural principles:

1. **Modular, dependency-aware design** - Each module has its own CI workflow, orchestrated based on dependency relationships
2. **Reusable actions pattern** - Common build and test logic is extracted into reusable composite actions
3. **Artifact propagation strategy** - Build artifacts are cached and reused across workflow runs and jobs

## Design Principles

### Incremental CI with Change Detection

The CI system performs incremental builds by detecting which modules have changed and only rebuilding those modules plus their dependents.

**For pull requests:**

- Compare against PR base SHA to determine what the PR introduces

**For push to main:**

- Compare against last successful CI run to build only what changed since last success

**Bootstrap mode:**

- When no previous successful CI run exists, build all modules

### Dependency-Based Execution Order

Modules are built in layers based on their dependency relationships:

- **Layer 1:** Modules with no dependencies
- **Layer 2:** Modules that depend only on Layer 1 modules
- **Layer N:** Modules that depend on modules in previous layers

Dependencies are defined in `.eac/repository.yml` under the `depends_on` field.

### Parallel Execution Within Layers

Modules within the same layer have no dependencies on each other and can be built in parallel.

**For pull requests:**

- All modules run in parallel for fast feedback
- Only commit test suite runs (quick validation)

**For main branch:**

- Modules run in dependency layers (sequential layers, parallel within each layer)
- Full test suite runs (commit + acceptance tests)

### Artifact Caching and Reuse

Build artifacts are uploaded and downloaded between workflow runs:

1. **Commands binary caching:** The `commands` binary is built once and cached as an artifact
2. **Module build artifacts:** Build outputs are uploaded and can be downloaded by dependent workflows
3. **Cross-job artifact sharing:** Artifacts from the trigger workflow are available to child workflows via `trigger_run_id` input

## Common Patterns

### `workflow_call` for Reusability

Module CI workflows use `workflow_call` trigger to be invoked by the orchestrator:

```yaml
on:
  workflow_call:
    inputs:
      ref:
        required: false
        type: string
      trigger_run_id:
        required: false
        type: string
```

### `workflow_dispatch` for Manual Triggers

Workflows support manual triggering for testing and debugging:

```yaml
on:
  workflow_dispatch:
    inputs:
      ref:
        required: false
        type: string
        description: 'Git ref to checkout'
```

### Artifact Download Conventions

Child workflows download artifacts from the trigger workflow using the `trigger_run_id` input:

```yaml
- name: Download commands binary
  uses: actions/download-artifact@v6
  with:
    name: commands-binary
    path: out/tools
    run-id: ⟪ inputs.trigger_run_id ⟫
```

### Status Check Naming Conventions

Workflow status checks follow a consistent naming pattern:

- **CI workflows:** `CI: {module-name}` (e.g., `CI: eac-commands`)
- **Release workflows:** `Release: {module-name}` (e.g., `Release: r2r-cli`)
- **Orchestration:** `CI Trigger`, `Release Trigger`
- **Security:** `CodeQL`

## Reusable Actions

The repository defines composite actions in `.github/actions/` for common operations:

### `.github/actions/setup-commands`

Sets up the commands binary for use in workflow steps.

**Outputs:**

- `commands-path` - Path to the commands executable
- `source` - Source of the commands binary (cached artifact or fresh build)

**Usage:**

```yaml
- name: Setup Commands Binary
  id: commands
  uses: ./.github/actions/setup-commands
```

### `.github/actions/build-module`

Builds a module using the commands binary.

**Inputs:**

- `module` (required) - Module moniker
- `trigger-run-id` (optional) - Run ID of trigger workflow for artifact download
- `setup-deps` (optional) - Dependency setup mode (auto, always, never)

**Usage:**

```yaml
- name: Build Module
  uses: ./.github/actions/build-module
  with:
    module: eac-commands
    trigger-run-id: ⟪ inputs.trigger_run_id ⟫
```

### `.github/actions/test-module`

Tests a module using the commands binary.

**Inputs:**

- `module` (required) - Module moniker
- `trigger-run-id` (optional) - Run ID of trigger workflow for artifact download
- `suites` (optional) - Comma-separated test suite list (default: commit)
- `setup-deps` (optional) - Dependency setup mode (auto, always, never)

**Usage:**

```yaml
- name: Test Module
  uses: ./.github/actions/test-module
  with:
    module: eac-commands
    trigger-run-id: ⟪ inputs.trigger_run_id ⟫
    suites: commit,acceptance
```

### `.github/actions/extract-release-version`

Extracts and validates version from git tag or workflow input.

**Inputs:**

- `module-prefix` (required) - Module tag prefix (e.g., r2r-cli)
- `legacy-prefixes` (optional) - Legacy tag prefixes for backwards compatibility
- `commands-path` (required) - Path to commands binary

**Outputs:**

- `version` - Extracted version (e.g., 1.0.0)
- `tag_name` - Full tag name (e.g., r2r-cli/1.0.0)
- `is_valid` - Whether version is valid semver (true/false)

**Usage:**

```yaml
- name: Extract and validate version
  id: extract_version
  uses: ./.github/actions/extract-release-version
  with:
    module-prefix: r2r-cli
    commands-path: ⟪ steps.commands.outputs.commands-path ⟫
```

## References

- [Trigger Orchestration](./trigger-orchestration.md) - Detailed change-trigger.yaml specification
- [CI Workflows](./ci-workflows.md) - Module CI workflow patterns
- [Repository Layout](../../../eac/architecture/repository-layout.md) - Module structure and dependencies
