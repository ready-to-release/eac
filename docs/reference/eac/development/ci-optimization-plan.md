# CI Workflow Optimization Plan

## Executive Summary

After aggressive analysis of all 11 CI workflows and 4 supporting actions, identified **10 optimization opportunities** ranging from critical to low priority. Estimated total savings: **30-50% reduction in CI time** and **significant reduction in log noise**.

---

## Critical Optimizations

### 1. Remove Verbose Diagnostics from Actions (Est. savings: 1-2 min/workflow)

**Problem**: `build-module` and `test-module` actions have extensive diagnostic steps that run on every invocation, adding significant log noise and time.

**Files affected**:

- `.github/actions/build-module/action.yaml` (lines 163-227)
- `.github/actions/test-module/action.yaml` (lines 116-210)

**Changes**:

```yaml
# Remove or make conditional (only on failure):
- Pre-Upload Diagnostics (build-module)
- Post-Upload Diagnostics (build-module)
- Pre-Download Diagnostics (test-module)
- Post-Download Diagnostics (test-module)
```

**Action**: Delete diagnostic steps or add `if: failure()` condition.

---

### 2. Merge Build+Test Jobs for Simple Modules (Est. savings: 30-60s/workflow)

**Problem**: Separate build and test jobs incur overhead:

- Job startup time (~15-30s each)
- Artifact upload from build job
- Artifact download to test job
- Redundant checkout in test job

**Workflows affected** (simple Go modules with no cross-platform tests):

- `ci-eac-core.yaml`
- `ci-eac-commands.yaml`
- `ci-eac-mcp-commands.yaml`
- `ci-vscode-ext-commit.yaml`
- `ci-docs.yaml`
- `ci-books.yaml`

**Proposed structure**:

```yaml
jobs:
  build-and-test:
    name: Build & Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - name: Build Module
        uses: ./.github/actions/build-module
        with:
          module: eac-commands
          upload-artifact: 'false'  # Don't upload, test runs locally
      - name: Test Module
        uses: ./.github/actions/test-module
        with:
          module: eac-commands
          download-artifact: 'false'  # Already have local artifacts
```

**Keep separate for**:

- `ci-r2r-cli.yaml` - needs cross-platform tests (Linux + Windows)
- `ci-ext-eac.yaml` - complex container workflow
- `ci-scripts-*` - need cross-platform tests

---

### 3. Eliminate Double Commands Binary Build in change-trigger (Est. savings: 30-60s)

**Problem**: `change-trigger.yaml` builds commands binary in `build-tooling`, then `detect-changes` downloads it again. Unnecessary artifact round-trip.

**Current flow**:

```text
build-tooling → uploads commands-binary artifact
detect-changes → downloads commands-binary artifact
```

**Proposed flow**:

```text
build-tooling → build + upload artifact
detect-changes → download artifact (only if not local)
```

**Or better**: Merge `build-tooling` into `detect-changes` since they run sequentially anyway.

---

## Medium Optimizations

### 4. Remove Legacy "(stage 1-7)" from Workflow Names

**Problem**: All CI workflow names include `(stage 1-7)` which is legacy from layered execution and no longer meaningful.

**Change**: Remove from all 11 CI workflows.

```yaml
# Before
name: "ci-eac-commands (stage 1-7)"

# After
name: "ci-eac-commands"
```

---

### 5. Remove Dead Code in ci-r2r-cli.yaml

**Problem**: Lines 79-84 "Restore Executable Permissions" runs after tests complete but serves no purpose.

```yaml
# DELETE THIS (dead code)
- name: Restore Executable Permissions
  run: |
    chmod +x out/build/r2r-cli/r2r-linux-amd64
    chmod +x out/build/r2r-cli/r2r-linux-arm64
    chmod +x out/build/r2r-cli/r2r-darwin-amd64
    chmod +x out/build/r2r-cli/r2r-darwin-arm64
```

---

### 6. Remove Auto-Release Trigger from docs/books CI

**Problem**: `ci-docs.yaml` and `ci-books.yaml` auto-trigger release workflows on main branch. This is aggressive - releases should be explicit.

**Files affected**:

- `ci-docs.yaml` (lines 80-101)
- `ci-books.yaml` (lines 80-100)

**Action**: Remove `trigger-release` job entirely. Releases should be manual or via changelog-based triggering.

---

### 7. Simplify test-module Artifact Handling

**Problem**: test-module always downloads build artifacts, even when running in same workflow where build just completed.

**Change**: Add `skip-download` mode for in-workflow testing.

```yaml
# test-module/action.yaml - add new mode
inputs:
  use-local-artifacts:
    description: 'Use local artifacts from build step (skip download)'
    default: 'false'
```

---

## Low Priority Optimizations

### 8. Consolidate Workflow Input Definitions

**Problem**: All 11 CI workflows have identical `on:` trigger definitions (40+ lines each). Total duplication: ~440 lines.

**Option A**: Accept duplication (YAML doesn't support includes)
**Option B**: Generate workflows from template

**Recommendation**: Accept for now, document pattern.

---

### 9. Reduce setup-commands Complexity

**Problem**: 258-line action with complex Windows cross-compilation logic.

**Observation**: Windows cross-compilation happens on every Linux build, even if no Windows tests needed.

**Potential fix**: Only cross-compile when workflow explicitly needs Windows.

---

### 10. Remove r2r-eac CI (Questionable Value)

**Problem**: `ci-r2r-eac.yaml` is a "static module" CI that only validates config and triggers release. It adds a CI workflow that doesn't actually test anything.

**Current behavior**:

```yaml
validate:  # Just checks module exists in config
trigger-release:  # Auto-triggers on main
```

**Consideration**: Is this necessary? The release bundle doesn't need CI validation.

---

## Implementation Priority

| Priority | Optimization                    | Est. Savings     | Complexity |
| -------- | ------------------------------- | ---------------- | ---------- |
| 1        | Remove verbose diagnostics      | 1-2 min/workflow | Low        |
| 2        | Merge build+test jobs           | 30-60s/workflow  | Medium     |
| 3        | Fix change-trigger double build | 30-60s           | Low        |
| 4        | Remove "(stage 1-7)" names      | Cosmetic         | Trivial    |
| 5        | Remove r2r-cli dead code        | Cosmetic         | Trivial    |
| 6        | Remove auto-release triggers    | Behavior change  | Low        |
| 7        | Simplify test-module downloads  | 10-20s/workflow  | Medium     |

---

## Estimated Total Impact

**Before optimization** (typical eac-commands CI):

- build-tooling: 45s
- build: 90s (including 30s diagnostics)
- test: 120s (including 40s artifact download + diagnostics)
- **Total: ~4.5 minutes**

**After optimization**:

- build-and-test (merged): 150s
- **Total: ~2.5 minutes**

> **Savings:** ~45% per simple module workflow
