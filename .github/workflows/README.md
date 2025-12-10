# GitHub Workflows Architecture

This repository implements a sophisticated CI/CD pipeline with:

- **Incremental CI** - Dependency-aware change detection for fast feedback
- **Parallel PR testing** - Fast commit suite for PRs
- **Sequential trunk builds** - Full acceptance testing on main branch
- **Automated releases** - Version-based deployment to multiple platforms
- **Security scanning** - CodeQL and supply chain attestations

---

## Table of Contents

1. [Workflow Overview](#1-workflow-overview)
2. [Architecture](#2-architecture)
3. [Standard Patterns](#3-standard-patterns)
4. [Special Cases](#4-special-cases)
5. [CI → Release Flow](#5-ci--release-flow)
6. [Key Concepts](#6-key-concepts)

---

## 1. Workflow Overview

### Workflow Categories

| Category              | Count | Purpose                                |
| --------------------- | ----- | -------------------------------------- |
| **CI Workflows**      | 11    | Build and test modules on PR/main      |
| **Release Workflows** | 4     | Publish artifacts to GitHub/GHCR/Pages |
| **Orchestration**     | 3     | Coordinate CI/release execution        |
| **Security**          | 1     | CodeQL scanning                        |

### Core Actions

| Action              | Purpose                                         |
| ------------------- | ----------------------------------------------- |
| `setup-commands`    | Provide commands binary to all workflows        |
| `build-module`      | Standardized build + artifact upload + summary  |
| `test-module`       | Standardized test + artifact download + summary |
| `setup-module-deps` | Setup system dependencies (Docker, Node, etc.)  |

---

## 2. Architecture

### Complete Pipeline

```text
┌─────────────────────────────────────────────────────────────┐
│                   GITHUB EVENTS (Triggers)                   │
└─────────────────────────────────────────────────────────────┘
         │                    │                  │
    Push/PR              CHANGELOG            Tag push
      main                change
         │                    │                  │
         ▼                    ▼                  ▼
┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐
│  trigger-ci     │  │ trigger-release │  │  Release     │
│  (orchestrate)  │  │ (detect tags)   │  │  Workflows   │
└─────────────────┘  └─────────────────┘  └──────────────┘
         │                    │
         ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│                    CI WORKFLOWS (11)                         │
│           Build → Test → Optionally trigger release          │
└─────────────────────────────────────────────────────────────┘

Standard Go Modules (8):
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ eac-core     │  │ eac-commands │  │ r2r-cli      │
│ Build → Test │  │ Build → Test │  │ Build → Test │
└──────────────┘  └──────────────┘  └──────────────┘

Special Modules (3):
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ ext-eac      │  │ docs         │  │ books        │
│ Docker build │  │ MkDocs site  │  │ PDF build    │
│ Push to GHCR │  │ → GH Pages   │  │ → Releases   │
└──────────────┘  └──────────────┘  └──────────────┘

┌─────────────────────────────────────────────────────────────┐
│                  RELEASE WORKFLOWS (4)                       │
└─────────────────────────────────────────────────────────────┘

┌──────────────────┐  ┌──────────────────┐
│ release-r2r-cli  │  │ release-ext-eac  │
│ Build binaries   │  │ Retag container  │
│ → GH Releases    │  │ → GHCR versioned │
└──────────────────┘  └──────────────────┘

┌──────────────────┐  ┌──────────────────┐
│ release-docs     │  │ release-books    │
│ Use CI artifacts │  │ Build PDFs       │
│ → GitHub Pages   │  │ → GH Releases    │
└──────────────────┘  └──────────────────┘
```

### Artifact Flow

```text
trigger-ci (build-tooling job)
  │
  ├─ Upload: commands-binary
  │
  ▼
ci-<module> (build job)
  │
  ├─ Download: commands-binary
  ├─ Build module
  ├─ Upload: build-artifacts-{module}
  │
  ▼
ci-<module> (test job)
  │
  ├─ Download: commands-binary
  ├─ Download: build-artifacts-{module}
  ├─ Run tests (commit for PRs, commit+acceptance for main)
  └─ Upload: test-results-{module}
```

---

## 3. Standard Patterns

### Adding a New CI Workflow

**Pattern for Go modules:**

```yaml
name: "ci-my-module (stage 1-6)"

on:
  workflow_call:
    inputs:
      ref: { type: string, default: '' }
      trigger_run_id: { type: string, default: '' }
  workflow_dispatch:
    inputs:
      ref: { type: string, default: '' }
      trigger_run_id: { type: string, default: '' }

permissions:
  contents: read
  actions: read

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ inputs.ref || github.ref }}

      - uses: ./.github/actions/build-module
        with:
          module: my-module
          trigger-run-id: ${{ inputs.trigger_run_id }}

  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ inputs.ref || github.ref }}

      - uses: ./.github/actions/test-module
        with:
          module: my-module
          trigger-run-id: ${{ inputs.trigger_run_id }}
          setup-deps: auto
          suites: ${{ startsWith(inputs.ref, 'refs/pull/') && 'commit' || 'commit,acceptance' }}
```

**Key points:**

- `trigger-run-id` enables artifact caching from orchestrator
- PRs run `commit` suite only (fast)
- Main branch runs `commit,acceptance` (full validation)
- `build-module` and `test-module` handle all artifact operations + summaries

---

## 4. Special Cases

### Container Build (ext-eac)

**Difference:** Builds Docker image, pushes to GHCR instead of uploading artifacts

```yaml
build:
  steps:
    - uses: ./.github/actions/build-module
      with:
        module: ext-eac
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}  # For GHCR push

test:
  steps:
    # Custom Docker-based testing (not test-module action)
    - name: Pull CI image
      run: docker pull ghcr.io/ready-to-release/ext-eac:sha-$SHORT_SHA

    - name: Test container
      run: ./r2r run eac show modules
```

**Why different:** Container testing requires Docker runtime, custom integration tests

---

### Documentation (ci-docs)

**Difference:** Triggers release workflow automatically on main branch

```yaml
build:
  - uses: ./.github/actions/build-module
    with:
      module: docs

test:
  - uses: ./.github/actions/test-module
    with:
      module: docs
      build-artifact-name: build-artifacts-docs

trigger-release:
  needs: test
  if: (github.ref == 'refs/heads/main') && !startsWith(github.ref, 'refs/pull/')
  steps:
    - run: |
        gh workflow run release-docs.yaml \
          --repo ${{ github.repository }} \
          -f ci_run_id=${{ github.run_id }}
```

**Artifact handoff:** `ci_run_id` parameter lets release workflow download CI artifacts

---

### Multi-Platform Testing (scripts-cli-installer)

**Difference:** Separate test jobs for Linux and Windows

```yaml
test-linux:
  runs-on: ubuntu-latest
  steps:
    - uses: ./.github/actions/test-module
      with:
        module: scripts-cli-installer
        setup-deps: auto

test-windows:
  runs-on: windows-latest
  steps:
    - uses: ./.github/actions/test-module
      with:
        module: scripts-cli-installer
        build-artifact-name: build-artifacts-scripts-cli-installer
```

---

## 5. CI → Release Flow

### Automated (Docs & Books)

CI workflows trigger releases automatically on main branch:

| CI Workflow | →   | Release Workflow | Trigger           | Parameters  |
| ----------- | --- | ---------------- | ----------------- | ----------- |
| ci-docs     | →   | release-docs     | `gh workflow run` | `ci_run_id` |
| ci-books    | →   | release-books    | `gh workflow run` | `ci_run_id` |

**Flow:**

1. PR merged to main
2. CI runs, builds artifacts
3. CI triggers release workflow with `ci_run_id`
4. Release downloads artifacts, deploys

---

### Manual (r2r-cli & ext-eac)

Developer-initiated release via CHANGELOG:

**Flow:**

1. Run `release this <module>` locally
2. CHANGELOG updated, PR created
3. PR reviewed and merged
4. `trigger-release` workflow detects CHANGELOG change
5. Creates git tag (e.g., `r2r-cli/0.1.0`)
6. Tag push triggers release workflow

**Workflows:**

- `release-r2r-cli`: Builds binaries from source, uploads to GitHub Releases
- `release-ext-eac`: Retags CI container image from `sha-{short}` to `{version}`

---

### Release Strategies

| Workflow        | Uses CI Artifacts | Strategy            | Reason                        |
| --------------- | ----------------- | ------------------- | ----------------------------- |
| release-r2r-cli | ❌                | Build from source   | Reproducible builds           |
| release-ext-eac | ✅                | Retag image         | Ensures tested image released |
| release-docs    | ✅                | Download artifacts  | No rebuild, CI-tested assets  |

---

## 6. Key Concepts

### Change Detection

**trigger-ci** workflow detects changes and only runs affected modules:

- Analyzes changed files
- Determines affected modules via dependency graph
- Dispatches only necessary CI workflows
- Passes `trigger-run-id` for artifact caching

**Override:** Use `trigger-all: true` to run all workflows

---

### Test Suites

**Two-tier strategy for fast PRs, thorough trunk:**

| Branch       | Suites              | Duration   | Purpose         |
| ------------ | ------------------- | ---------- | --------------- |
| Pull Request | `commit` only       | ~5-10 min  | Fast feedback   |
| Main (trunk) | `commit,acceptance` | ~15-30 min | Full validation |

**Implementation:**

```yaml
suites: ${{ startsWith(inputs.ref, 'refs/pull/') && 'commit' || 'commit,acceptance' }}
```

---

### Commands Binary Caching

All workflows need the `commands` binary. Three modes:

1. **CI orchestration:** `trigger-ci` builds once, uploads artifact
2. **CI modules:** Download from `trigger-run-id`, fall back to building
3. **Release/manual:** Build from source

**Result:** Most workflows reuse binary, faster execution

---

### Composite Action Design

**build-module** and **test-module** encapsulate:

- Commands setup
- Dependency installation
- Build/test execution
- Artifact upload/download
- Summary generation

**Why:** DRY principle, consistent patterns, single source of truth

---

### Security

**Supply Chain:**

- Build provenance attestations (Sigstore) for r2r-cli
- SBOM generation
- CI verification before release

**Code Scanning:**

- CodeQL on push/PR/schedule
- Go + GitHub Actions analysis

---

## How-To Guides

### Add a New Module to CI

1. Create module contract in `.r2r/eac/repository.yml`
2. Copy `ci-eac-core.yaml` → `ci-my-module.yaml`
3. Replace `eac-core` with `my-module` throughout
4. Update `trigger-ci.yaml` to include new workflow (if dependency-based orchestration needed)
5. Test with `workflow_dispatch`

### Trigger a Release

**For modules with automated releases (docs, books):**

- Merge to main → automatic

**For modules with manual releases (r2r-cli, ext-eac):**

```bash
# Update CHANGELOG
./commands release this my-module

# Create PR, get reviewed, merge

# trigger-release detects change and creates tag
```

### Debug a Failed Workflow

1. Check GitHub Actions summary for error
2. View workflow logs in `.github/workflows/`
3. Reproduce locally:

   ```bash
   # Build
   ./commands build my-module

   # Test
   ./commands test my-module --suite commit
   ```

### Update Commands Binary

Commands binary is rebuilt on every `trigger-ci` run. To force rebuild in a specific workflow:

```yaml
- uses: ./.github/actions/setup-commands
  with:
    trigger-run-id: ''  # Empty = always build from source
```

---

## Maintenance Notes

### Adding New Release Workflows

1. Create `release-my-module.yaml`
2. Update `trigger-release.yaml` module list (if using changelog-based release)
3. Document artifact strategy (rebuild vs reuse CI artifacts)

### Monitoring

**Scheduled full CI** runs every 2 hours (`cron-full-trigger`):

- Detects issues incremental CI might miss
- Validates change detection logic
- Monitor failure rate for insights

---

## Architecture Decisions

### Why Rebuild for Releases?

**r2r-cli:** Reproducible builds require fresh compilation
**books:** Release uses `--all` flag (different from CI subset)

### Why Retag for Containers?

**ext-eac:** Ensures released image is identical to tested CI image

### Why Auto-Trigger Some Releases?

**docs/books:** Deployment is low-risk, calver-tagged, auto-deploy on main merge is safe
