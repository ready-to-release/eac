# CI/CD Workflow Reference

Comprehensive reference for the EAC repository CI/CD system. Covers architecture, workflow mapping, CLI commands, diagnostics, and debugging playbooks.

## 1. Architecture Overview

```text
Push/PR to main
  │
  ▼
change-trigger.yaml (CI Orchestrator)
  ├─ detect-changes ─── eac get changed-modules-ci
  │     Compares files against last successful CI
  │     Outputs: directly_changed + invalidated modules
  │
  ├─ dispatch ─── eac pipeline ci schedule
  │     Dispatches ci-{moniker}.yaml for each changed module
  │     Respects dependency order and concurrency limits
  │
  └─ dispatch-releases ─── gh workflow run release-trigger.yaml
        Only on push to main (not PRs)

ci-{moniker}.yaml (Per-Module CI)
  └─ calls _module-ci.yaml (Reusable Template)
       config → build → test → scan → evidence → summary

release-trigger.yaml (Release Orchestrator)
  ├─ Awaits all CI workflows
  ├─ Checks pending releases (semver from CHANGELOG, calver from dispatch)
  └─ Dispatches release-{moniker}.yaml for each pending release

release-{moniker}.yaml (Per-Module Release)
  └─ calls _module-release.yaml (Reusable Template)
       config → approve → release → evidence → cleanup → summary

cron-full-trigger.yaml (every 2 hours)
  └─ gh workflow run change-trigger.yaml --trigger-all=true
       Full rebuild to catch deviation between incremental and full builds
```

**Key design principles:**

- **Config-driven**: All behavior derived from `.eac/repository.yml` via `eac get ci-config` / `eac get release-config`
- **Module-per-workflow**: Each module moniker maps to its own `ci-{moniker}.yaml`
- **Smart change detection**: Only changed modules (and their dependents) get CI dispatched
- **Dependency-aware scheduling**: Modules dispatch in topological order; dependents wait for dependencies

## 2. Workflow File Map

### Orchestration Workflows

| File                     | Trigger                                             | Purpose                                                           |
| ------------------------ | --------------------------------------------------- | ----------------------------------------------------------------- |
| `change-trigger.yaml`    | push/PR to main, `workflow_dispatch`                | Detects changes, dispatches per-module CI, triggers releases      |
| `release-trigger.yaml`   | `workflow_dispatch` (from change-trigger or manual) | Awaits CI, detects pending releases, dispatches release workflows |
| `cron-full-trigger.yaml` | Schedule (every 2 hours)                            | Triggers full rebuild via change-trigger with `trigger-all=true`  |

### Reusable Workflow Templates

| File                   | Called By                      | Purpose                                                             |
| ---------------------- | ------------------------------ | ------------------------------------------------------------------- |
| `_module-ci.yaml`      | All `ci-*.yaml` workflows      | Config-driven CI: build, test (Linux/Windows/macOS), scan, evidence |
| `_module-release.yaml` | All `release-*.yaml` workflows | Config-driven release: approve, build/publish, evidence, cleanup    |

### Per-Module CI Workflows

| File                        | Module Moniker    |
| --------------------------- | ----------------- |
| `ci-core.yaml`              | core              |
| `ci-clie.yaml`              | clie              |
| `ci-eac.yaml`               | eac               |
| `ci-eac-ext.yaml`           | eac-ext           |
| `ci-docs.yaml`              | docs              |
| `ci-repository.yaml`        | repository        |
| `ci-mkdocs-render-oci.yaml` | mkdocs-render-oci |
| `ci-pdf-oci.yaml`           | pdf-oci           |
| `ci-implicit-cli.yaml`      | implicit-cli      |
| `ci-cli-installers.yaml`    | cli-installers    |
| `ci-eac-mcp-server.yaml`    | eac-mcp-server    |
| `ci-vscode-commit.yaml`     | vscode-commit     |

### Release Workflows

| File                           | Module          | Release Type        |
| ------------------------------ | --------------- | ------------------- |
| `release-clie.yaml`            | clie            | cli-binary (SemVer) |
| `release-eac-ext.yaml`         | eac-ext         | container (SemVer)  |
| `release-docs.yaml`            | docs            | docs-site (CalVer)  |
| `release-clie-eac-bundle.yaml` | clie-eac-bundle | bundle (SemVer)     |

### Support Workflows

| File          | Purpose                                 |
| ------------- | --------------------------------------- |
| `codeql.yaml` | CodeQL security scanning (Go + Actions) |
| `labeler.yml` | Auto-label PRs by affected modules      |

## 3. Module-to-Workflow Mapping

Modules are defined in `.eac/repository.yml`. Each module moniker maps to a CI workflow by convention:

```text
Module moniker: {moniker}
CI workflow:    .github/workflows/ci-{moniker}.yaml
Release workflow: .github/workflows/release-{moniker}.yaml
```

**Not all modules have CI workflows.** Only modules with independently testable components get their own CI. Modules like `contracts`, `clibase`, `adapters`, `commands`, and `templates` are tested as dependencies of other modules.

### Module Dependency Graph

```text
contracts (no CI)
  └─ core
       └─ clibase (no CI)
            └─ adapters (no CI)
                 └─ commands (no CI)
                      └─ eac ──────────────────┐
                           ├─ implicit-cli      │
                           ├─ vscode-commit     │
                           ├─ eac-mcp-server    │
                           └─ eac-ext ◄─────────┤ (depends_on_ci: clie)
                                                │
contracts ─── clie ────────────────────────────►┘
  │
  └─ (oci-tools group) ── mkdocs-render-oci, pdf-oci, drawio-oci, ...
       └─ repository
            └─ docs (depends_on: contracts, repository, oci-tools, vscode-commit, eac-mcp-server, adapters)

clie-eac-bundle (depends_on: eac, clie, eac-ext, docs, repository, oci-tools)
cli-installers (depends_on: eac, clie)
```

### Versioning Schemes

| Module          | Scheme   | Release Type | Changelog                              |
| --------------- | -------- | ------------ | -------------------------------------- |
| clie            | SemVer   | published    | `release/clie/CHANGELOG.md`            |
| eac             | SemVer   | published    | `release/eac/CHANGELOG.md`             |
| eac-ext         | SemVer   | published    | `release/eac-ext/CHANGELOG.md`         |
| docs            | CalVer   | published    | (auto-generated)                       |
| clie-eac-bundle | SemVer   | bundle       | `release/clie-eac-bundle/CHANGELOG.md` |
| vscode-commit   | Implicit | internal     | -                                      |
| eac-mcp-server  | SemVer   | internal     | `go/mcp/commands/CHANGELOG.md`         |

## 4. Reusable Workflow: `_module-ci.yaml`

### Inputs

| Input            | Required | Description                                                            |
| ---------------- | -------- | ---------------------------------------------------------------------- |
| `module`         | yes      | Module moniker to build and test                                       |
| `ref`            | no       | Git ref to checkout                                                    |
| `sha`            | no       | Commit SHA for artifact lookup                                         |
| `trigger_run_id` | no       | Run ID of trigger workflow (for downloading pre-built commands binary) |

### Job Structure

```text
config ─── Derives all CI settings from repository.yml
  │          eac get ci-config --module {module} --format github-output
  ▼
build ──── Builds binary (build-module action) or container (build-container action)
  │          Conditional: is-container determines which path
  ▼
┌─────────────────────────────────────────────────┐
│  test-linux ── Always if has-tests && !container │
│  test-windows ── If test-on-windows == true      │  (parallel)
│  test-macos ── If test-on-macos == true          │
│  scan ── If scans != ''                          │
│  container-test ── If container + test script    │
└─────────────────────────────────────────────────┘
  │
  ▼
evidence ── Builds compliance PDFs (if build-evidence == true && no failures)
  │
  ▼
summary ─── Generates CI summary (always runs)
              eac show ci-summary --build={result} --test-linux={result} ...
```

### Config-Driven Outputs

The `config` job runs `eac get ci-config` which returns:

| Output                  | Meaning                                         |
| ----------------------- | ----------------------------------------------- |
| `is-container`          | Module has a Dockerfile component               |
| `has-tests`             | Module has components with testers              |
| `test-on-windows`       | `artifact_matrix` includes windows              |
| `test-on-macos`         | `artifact_matrix` includes macos                |
| `scans`                 | Aggregated scanners (sbom, vuln, secrets, etc.) |
| `build-evidence`        | Has evidence-book components                    |
| `cross-compile-windows` | `artifact_matrix` is cross-platform             |
| `download-modules`      | CI dependencies (from `depends_on_ci`)          |
| `test-suites`           | PR test suites (typically: unit)                |
| `test-suites-full`      | Push-to-main test suites (unit,integration)     |

### Test Suite Selection

- **Pull requests**: Run `test-suites` (fast, typically unit tests only)
- **Push to main**: Run `test-suites-full` (comprehensive, unit + integration)

## 5. Composite Actions

### Infrastructure Actions

| Action                    | Purpose                                                                          |
| ------------------------- | -------------------------------------------------------------------------------- |
| `setup-commands`          | Gets the `eac` commands binary (download from trigger or build from source)      |
| `setup-module-deps`       | Installs system deps (Go, Node, Docker Buildx, QEMU, UPX) based on module config |
| `download-artifact-retry` | Wraps `actions/download-artifact` with 3-attempt retry (10s, 30s backoff)        |
| `upload-artifact-retry`   | Wraps `actions/upload-artifact` with 3-attempt retry                             |

### Build/Test/Scan Actions

| Action                 | Purpose                                                                          |
| ---------------------- | -------------------------------------------------------------------------------- |
| `build-module`         | Builds a single module via `eac build --module {m} --skip-cache --skip-depm`     |
| `build-container`      | Builds container module, pushes to GHCR with `sha-{short}` tag                   |
| `test-module`          | Tests a module with configurable suites via `eac test --module {m} --suites {s}` |
| `scan-module`          | Runs security scans (sbom, vuln, secrets, compliance, iac, sast, zap)            |
| `build-evidence-books` | Downloads test/scan artifacts from module + deps, builds compliance PDFs         |

### Release Actions

| Action                       | Purpose                                                                      |
| ---------------------------- | ---------------------------------------------------------------------------- |
| `approve-release`            | Gates release: validates version, changelog, CI status                       |
| `extract-release-version`    | Extracts version from tag trigger or dispatch input                          |
| `attested-cli-build`         | Builds CLI binaries with supply-chain attestations, creates GitHub Release   |
| `attested-container-publish` | Retags CI container image with release tags, generates attestations          |
| `trigger-release`            | Dispatches a release workflow via `gh workflow run`                          |
| `check-pending-releases`     | Detects pending releases from CHANGELOG (semver), dispatch (calver), bundles |
| `await-dependency-ci`        | Waits for all transitive dependency modules to have passing CI               |
| `await-module-releases`      | Waits for in-flight module release workflows to complete                     |

### Cleanup Actions

| Action                   | Purpose                                                                |
| ------------------------ | ---------------------------------------------------------------------- |
| `cleanup-failed-release` | Removes orphaned tags and partial releases after failure               |
| `cleanup-pre-releases`   | Removes old CalVer pre-releases, keeping newest N                      |
| `update-evidence`        | Downloads evidence PDFs from CI and uploads to GitHub Release          |
| `download-ci-artifacts`  | Downloads build artifacts from a CI run, resolves run ID from SHA      |
| `download-module-deps`   | Downloads build artifacts for dependency modules with await + fallback |

## 6. EAC CLI Commands for CI

### Querying CI Status

```bash
# Get CI results for current HEAD (structured data)
eac get ci-results

# Get CI results for a specific commit
eac get ci-results abc1234

# Get CI results for specific modules
eac get ci-results abc1234 core clie

# Get CI results for a specific run ID
eac get ci-results 12345678

# Pretty-printed CI results
eac show ci-results
eac show ci-results abc1234
```

### Change Detection

```bash
# Detect which modules need CI (used by change-trigger)
eac get changed-modules-ci --format shell
# Outputs: CHANGED_MODULES, DIRECTLY_CHANGED, INVALIDATED, HAS_CHANGES, etc.

# With PR base SHA
eac get changed-modules-ci --pr-base abc1234 --format shell

# Local change detection
eac get changed-modules-local

# List all CI workflow modules
eac get ci-workflows
eac get ci-workflows --format json
```

### CI Configuration

```bash
# Derive CI config for a module (what _module-ci.yaml reads)
eac get ci-config --module clie --format github-output

# Derive release config for a module
eac get release-config --module clie --format github-output

# Get CI workflow file for a module
eac get module-ci-workflow --module clie
# Output: ci-clie.yaml
```

### Pipeline Orchestration

```bash
# Schedule and dispatch CI with concurrency limits (used by change-trigger)
eac pipeline ci schedule \
  --directly-changed "core clibase" \
  --invalidated "eac docs" \
  --head-sha abc1234 \
  --dispatch-ref main \
  --max-concurrent 6 \
  --timeout 3600 \
  --trigger-run-id 12345

# Dispatch a single workflow and wait for completion
eac pipeline ci dispatch-and-wait \
  --workflow ci-clie.yaml \
  --ref main \
  --timeout 300

# Get CI run ID for a workflow and commit
eac pipeline ci get-run-id --workflow ci-docs.yaml --sha abc123

# Generate diagnostic markdown for CI summaries
eac pipeline ci summary-link 12345678 --type test --artifact test-results-clie

# Check pipeline status for current HEAD
eac pipeline status
eac pipeline status --commit abc123

# Wait for all CI workflows to complete
eac pipeline await-ci --sha abc123 --timeout 1800
eac pipeline await-ci --pattern "ci-*.yaml" --exclude "ci-orchestrator"

# Filter modules for dispatch (check which already have valid CI)
eac get ci-dispatch \
  --directly-changed "core" \
  --invalidated "eac docs" \
  --format shell
# Outputs: DISPATCH, SKIPPED, DISPATCH_COUNT, SKIPPED_COUNT
```

### Show Commands (Human-Readable)

```bash
# Generate CI summary markdown (used in workflow summaries)
eac show ci-summary \
  --build=success \
  --test-linux=success \
  --test-on-windows --test-windows=failure \
  --scans-enabled --scan=success

# Show CI results (formatted table)
eac show ci-results abc1234

# Show dependency CI summary
eac show dependency-ci-summary --module clie --passed 5 --skipped 2 --status success
```

## 7. `gh` CLI Quick Reference

### Querying Workflow Runs

```bash
# List recent runs for a workflow
gh run list --workflow=ci-clie.yaml --limit 10

# List runs for a specific branch
gh run list --workflow=ci-clie.yaml --branch main

# List failed runs
gh run list --workflow=ci-clie.yaml --status failure

# View a specific run (summary)
gh run view 12345678

# View run with job details
gh run view 12345678 --json jobs --jq '.jobs[] | {name, status, conclusion}'

# View run logs (full)
gh run view 12345678 --log

# View logs for a specific failed job
gh run view 12345678 --log-failed

# Download run logs to file
gh run view 12345678 --log > run-12345678.log
```

### Querying Jobs and Steps

```bash
# Get all jobs for a run
gh api repos/{owner}/{repo}/actions/runs/12345678/jobs \
  --jq '.jobs[] | {name: .name, status: .status, conclusion: .conclusion}'

# Get failed steps for a run
gh api repos/{owner}/{repo}/actions/runs/12345678/jobs \
  --jq '.jobs[].steps[] | select(.conclusion == "failure") | {job: .name, step: .name}'
```

### Artifacts

```bash
# List artifacts for a run
gh run view 12345678 --json artifacts --jq '.artifacts[].name'

# Download specific artifact
gh run download 12345678 -n build-artifacts-clie

# Download all artifacts from a run
gh run download 12345678

# Download test results
gh run download 12345678 -n test-results-clie
```

### Dispatching Workflows

```bash
# Trigger CI for a specific module
gh workflow run ci-clie.yaml --ref main \
  -f ref=main \
  -f sha=$(git rev-parse HEAD) \
  -f trigger_run_id=""

# Trigger full rebuild
gh workflow run change-trigger.yaml --ref main -f trigger-all=true

# Trigger a release
gh workflow run release-clie.yaml --ref main \
  -f version=1.2.3 \
  -f skip-ci-check=false

# Re-run a failed workflow
gh run rerun 12345678

# Re-run only failed jobs
gh run rerun 12345678 --failed
```

### Checking Status

```bash
# Watch a run in real-time
gh run watch 12345678

# Get run status as JSON
gh run view 12345678 --json status,conclusion

# Check if CI passed for a commit
gh run list --commit abc1234 --json workflowName,conclusion \
  --jq '.[] | select(.workflowName | startswith("CI:")) | {workflowName, conclusion}'

# Find the orchestrator run for a commit
gh run list --workflow=change-trigger.yaml --commit abc1234 --limit 1
```

### Release Inspection

```bash
# List releases
gh release list --limit 10

# View a specific release
gh release view clie/1.2.3

# Download release assets
gh release download clie/1.2.3 -D ./release-assets/

# Check release tags
git tag -l "clie/*" | sort -V | tail -5
```

## 8. Dependency and Cascade Model

### Dependency Types

| Type            | Declared In      | Used For                                                                               |
| --------------- | ---------------- | -------------------------------------------------------------------------------------- |
| `depends_on`    | `repository.yml` | Build-time dependencies; change propagation                                            |
| `depends_on_ci` | `repository.yml` | CI artifact dependencies (merged into depends_on)                                      |
| `group`         | `repository.yml` | Collective reference (e.g., `depends_on: [oci-tools]` expands to all modules in group) |

### Change Detection Flow

1. **File ownership resolution**: Map changed files to modules via directory-root matching (most-specific root wins)
2. **Direct changes**: Modules with modified source files
3. **Dependency propagation**: If module A depends on module B, and B changed, A is marked "invalidated"
4. **CI cache check**: Invalidated modules with valid CI at HEAD SHA are skipped
5. **Dispatch**: Remaining modules dispatched in topological order

```bash
# The full flow in one command:
eac get changed-modules-ci --format shell
# Returns:
#   CHANGED_MODULES="core eac docs"
#   DIRECTLY_CHANGED="core"
#   INVALIDATED="eac docs"
#   HAS_CHANGES=true
```

### Dispatch Ordering

The scheduler uses **dependency-aware LPT (Longest Processing Time)** scheduling:

1. Modules with no unsatisfied dependencies are dispatched first
2. As modules complete, dependents become eligible
3. Concurrency limit: 6 for push to main, 20 for PRs
4. **Cascade failure**: If a dependency fails, all dependents are marked failed without running

### Release Dependency Layers

Releases dispatch in layers based on dependency order:

```text
Layer 1: clie, eac (no release dependencies on each other)
Layer 2: eac-ext (depends_on_ci: clie), docs
Layer 3: clie-eac-bundle (awaits all module releases)
```

The bundle release uses `await-module-releases` to wait for all in-flight releases before creating the aggregated release.

## 9. Common Failure Patterns

### Build Failures

**Symptom**: `build` job fails in `_module-ci.yaml`

| Cause                       | Diagnosis                                                                  |
| --------------------------- | -------------------------------------------------------------------------- |
| Go compilation error        | `gh run view {id} --log-failed`, look for `go build` output                |
| Missing dependency artifact | Check if dependency module CI completed; look for download-artifact errors |
| Container build fails       | Check Dockerfile context path, base image availability                     |
| Cross-compile failure       | Check UPX/CGo setup for target platform                                    |

```bash
# Quick diagnosis
gh run view {run-id} --log-failed
eac get ci-results {run-id}
```

### Test Failures

**Symptom**: `test-linux`, `test-windows`, or `test-macos` job fails

| Cause                       | Diagnosis                                                 |
| --------------------------- | --------------------------------------------------------- |
| Unit test assertion failure | Download test results artifact, check Go test output      |
| Integration test timeout    | Check if dependent services (Docker, GHCR) were available |
| Platform-specific failure   | Compare Linux vs Windows vs macOS logs                    |
| Flaky test                  | Check if re-run succeeds; look for race conditions        |

```bash
# Download and inspect test results
gh run download {run-id} -n test-results-{module}

# Check if test passes on retry
gh run rerun {run-id} --failed
```

### Dispatch/Orchestration Failures

**Symptom**: `change-trigger` or `release-trigger` fails

| Cause                    | Diagnosis                                                               |
| ------------------------ | ----------------------------------------------------------------------- |
| Change detection error   | Check `detect-changes` job logs for `eac get changed-modules-ci` output |
| Dispatch timeout         | Check `eac pipeline ci schedule` output; modules may be stuck           |
| Concurrency exhaustion   | Check for queued/in-progress runs consuming slots                       |
| Missing CI workflow file | Verify `ci-{moniker}.yaml` exists for all dispatched modules            |

```bash
# Check what the orchestrator dispatched
gh run view {orchestrator-run-id} --log

# Check status of dispatched module CIs
eac pipeline status --commit {sha}

# Find stuck or queued runs
gh run list --status in_progress --limit 20
gh run list --status queued --limit 20
```

### Release Failures

**Symptom**: Release workflow fails at approve, build, or publish stage

| Cause                   | Diagnosis                                                       |
| ----------------------- | --------------------------------------------------------------- |
| CI not passing          | Approve job fails waiting for CI; check `eac pipeline await-ci` |
| Version already exists  | Release/tag already created; check `gh release view {tag}`      |
| Missing changelog entry | SemVer modules require version in CHANGELOG                     |
| Container retag failure | Source CI image not found; check GHCR for `sha-{short}` tag     |
| Attestation failure     | Check OIDC token permissions                                    |

```bash
# Check release approval
gh run view {release-run-id} --log

# Check if tag/release already exists
gh release view {module}/{version} 2>&1

# Check container images
gh api user/packages/container/{module}/versions --jq '.[].metadata.container.tags'
```

### Artifact Download Failures

**Symptom**: `download-artifact-retry` exhausts all 3 attempts

| Cause                               | Diagnosis                                           |
| ----------------------------------- | --------------------------------------------------- |
| Artifact expired (90-day retention) | Check artifact creation date                        |
| Wrong run ID                        | Verify `trigger_run_id` passed through correctly    |
| Artifact name mismatch              | Check naming convention: `build-artifacts-{module}` |
| GitHub API transient error          | Usually resolves on re-run                          |

```bash
# Check artifacts exist for a run
gh run view {run-id} --json artifacts --jq '.artifacts[].name'

# Check if commands binary artifact exists
gh run view {trigger-run-id} --json artifacts \
  --jq '.artifacts[] | select(.name == "commands-binary") | {name, size_in_bytes}'
```

## 10. Debugging Playbook

### Step 1: Identify the Failure

```bash
# Find the failing run
gh run list --workflow=change-trigger.yaml --status failure --limit 5

# Or find CI for a specific commit
gh run list --commit $(git rev-parse HEAD) --limit 10

# Check overall pipeline status
eac pipeline status
```

### Step 2: Determine Failure Scope

```bash
# View the orchestrator run
gh run view {orchestrator-run-id}

# Get structured CI results for all modules
eac show ci-results {commit-sha}

# Check which modules were dispatched
gh run view {orchestrator-run-id} --log | grep -i "dispatch\|schedule"
```

### Step 3: Drill Into the Failing Module

```bash
# Find the module CI run
eac pipeline ci get-run-id --workflow ci-{module}.yaml --sha {commit-sha}

# View its logs
gh run view {module-run-id} --log-failed

# Download artifacts for local inspection
gh run download {module-run-id} -n test-results-{module}
gh run download {module-run-id} -n scan-results-{module}
```

### Step 4: Check Dependencies

```bash
# View module dependency graph
eac get dependencies --module {module}

# Check if dependency CIs passed
eac get ci-results {commit-sha} {dep-module-1} {dep-module-2}

# Verify dependency artifacts are available
gh run view {module-run-id} --log | grep -i "download.*artifact"
```

### Step 5: Local Reproduction

```bash
# Build the module locally
eac build --module {module}

# Run the same test suites CI uses
eac test --module {module} --suites unit
eac test --module {module} --suites unit,integration

# Run scans locally
eac scan --module {module}
```

### Step 6: Fix and Verify

```bash
# Push fix and watch CI
git push
gh run list --workflow=change-trigger.yaml --limit 1
gh run watch {new-run-id}

# Or manually trigger CI for just the affected module
gh workflow run ci-{module}.yaml --ref {branch} \
  -f ref={branch} \
  -f sha=$(git rev-parse HEAD)
```

### Quick Triage Cheatsheet

| Scenario                                | Command                                                              |
| --------------------------------------- | -------------------------------------------------------------------- |
| "What failed?"                          | `eac show ci-results`                                                |
| "Why was module X rebuilt?"             | Check orchestrator `detect-changes` job summary                      |
| "Is CI passing on main?"                | `eac pipeline status`                                                |
| "What's the run ID for module X?"       | `eac pipeline ci get-run-id --workflow ci-X.yaml --sha HEAD`         |
| "Re-run failed CI"                      | `gh run rerun {run-id} --failed`                                     |
| "Trigger full rebuild"                  | `gh workflow run change-trigger.yaml --ref main -f trigger-all=true` |
| "What modules have CI workflows?"       | `eac get ci-workflows`                                               |
| "What config does CI use for module X?" | `eac get ci-config --module X`                                       |
| "Download test results"                 | `gh run download {run-id} -n test-results-{module}`                  |
| "Watch a run in progress"               | `gh run watch {run-id}`                                              |
