# Build Execution System

The build execution system is the engine that turns module contracts into
executed build/test/lint/scan operations. It orchestrates units of work
across modules with parallel scheduling, capacity management, and phased
execution.

## Unit of Work (UoW)

Every operation in EAC is modeled as a **Unit of Work** (UoW). A UoW is
the atomic scheduling and caching unit.

### Identity

Each UoW is uniquely identified by four fields:

```text
action:module:component:tool
```

| Field       | Description                                 | Examples                          |
|-------------|---------------------------------------------|-----------------------------------|
| **Action**  | What operation to perform                   | `build`, `test`, `lint`, `scan`   |
| **Module**  | Which module owns this work                 | `eac`, `core`, `eac-ext`     |
| **Component** | Which component within the module          | `go`, `dockerfile`, `typescript`  |
| **Tool**    | Which tool executes the work                | `go`, `buildx`, `gotest`, `godog` |

**Examples**:

- `build:eac:go:go` — Compile the eac Go binary
- `build:eac-ext:dockerfile:buildx` — Build the eac-ext Docker image
- `test:core:go:gotest` — Run Go unit tests for the core module
- `lint:docs:markdown:markdownlint` — Lint documentation markdown

**Source**: `go/core/workunit/unit_id.go`

### Output Directory

Each UoW writes to a deterministic output path:

```text
out/{action}/{module}/{component}-{tool}/
```

Examples:

- `out/build/eac/go-go/` — contains cross-compiled binaries and `uow.manifest.json`
- `out/test/core/go-gotest/` — contains test results and `uow.manifest.json`

**Source**: `UnitID.OutDir()` in `go/core/workunit/unit_id.go`

### UoW Manifest

Every completed UoW produces a manifest at `{outdir}/uow.manifest.json`:

```json
{
  "id": "build:eac:go:go",
  "input_hash": "sha256:abc123...",
  "output_hash": "sha256:def456...",
  "executed_at": "2025-02-07T15:30:00Z",
  "duration_ms": 12345,
  "artifacts": [
    {
      "id": "eac-linux-amd64",
      "type": "executable",
      "path": "eac-linux-amd64",
      "sha256": "789abc..."
    }
  ]
}
```

**Source**: `go/core/output/manifest.go`, `go/core/output/types.go`

### UoW Specification

Before execution, each UoW is described by a `UnitSpec`:

| Field       | Purpose                                              |
|-------------|------------------------------------------------------|
| `ID`        | The `action:module:component:tool` identity          |
| `DependsOn` | List of UoW IDs that must complete first             |
| `Weight`    | Scheduling weight (CPU cost) for capacity management |
| `Tags`      | Metadata tags for filtering and display              |

**Source**: `go/core/workunit/unit_spec.go`

---

## Orchestrator

The orchestrator manages the execution of all UoWs for a command invocation.
It handles parallel scheduling, capacity limits, dependency ordering, and
result collection.

### Architecture

```text
Command Entry Point (build.go / test.go / lint.go)
    │
    ├── Framework: resolves modules, creates UoW specs
    │
    ├── Orchestrator: manages lifecycle
    │       │
    │       ├── Phase Runner: executes phases sequentially
    │       │       │
    │       │       └── Unit Scheduler: parallel UoW execution
    │       │               │
    │       │               ├── Capacity Manager: CPU/memory budgets
    │       │               └── Worker Pool: concurrent execution
    │       │
    │       └── Display: progress rendering (TUI or console)
    │
    └── Output: manifests, artifacts, summaries
```

**Source files**:

- `go/clibase/orchestrator/orchestrator_core.go` — lifecycle management
- `go/clibase/orchestrator/orchestrator_phases.go` — phase execution
- `go/clibase/orchestrator/unit_scheduler_core.go` — scheduling loop
- `go/clibase/orchestrator/unit_scheduler_execution.go` — UoW execution
- `go/clibase/orchestrator/unit_scheduler_capacity.go` — capacity management

### Execution Phases

The orchestrator runs UoWs in **phases**. Each phase completes before the
next begins. Within a phase, UoWs run in parallel up to capacity limits.

```text
Phase 1: Pre-build components (drawio diagrams, mermaid renders)
Phase 2: Main build (go, typescript, dockerfile, mkdocs)
Phase 3: Post-build (PDF generation, site assembly)
```

Phases are determined by the `build_after` relationships between component
types. See [Component Resolution](./component-resolution.md) for details.

### Capacity Management

The scheduler uses a **dual-pool capacity model** to prevent resource
exhaustion:

| Pool     | Purpose                           | Limit                        |
|----------|-----------------------------------|------------------------------|
| **CPU**  | Tracks total CPU weight in use    | System CPU count (default)   |
| **Slot** | Tracks concurrent UoW count       | Configurable max concurrency |

Each UoW has a **weight** derived from its tool's resource configuration
multiplied by the component's **amp** (amplifier) value:

```text
effective_weight = tool.resources.cpus * component.amp.{action}
```

A UoW is scheduled only when both pools have sufficient capacity. When a
UoW completes, its capacity is returned to both pools.

**Source**: `go/clibase/capacity/dual_pool.go`

### Dependency Ordering

UoWs declare dependencies via `DependsOn`. The scheduler only starts a UoW
when all its dependencies have completed successfully. Dependencies come
from two sources:

1. **Intra-module** (`build_after`): Component A builds after component B
   within the same module. Example: `dockerfile` builds after `go`.
2. **Cross-module** (`depends_on`): All UoWs in module A depend on all UoWs
   in module B. Example: `eac-ext` depends on `eac`.

Cross-module dependencies are injected by `injectModuleDependencies()` in
`go/clibase/cmdframework/execute.go`. This function reads each module's
`depends_on` list and creates edges from every UoW in the dependent module
to every UoW in the dependency module.

### State Machine

Each UoW progresses through states:

```text
Pending → Running → Completed
                  → Failed
                  → Cached (skipped)
```

The state manager tracks transitions and computes aggregate status for
display and summary reporting.

**Source**: `go/core/workunit/unit_state.go`, `go/core/workunit/state_manager.go`

---

## Build Framework

The build command framework connects the command entry point to the
orchestrator. It handles module resolution, scope filtering, input hashing,
and incremental change detection.

### Flow

```text
1. Parse flags (module scope, force, artifacts mode)
2. Load module registry from repository.yml
3. Resolve scope (which modules to build)
4. Resolve components per module (via ComponentResolver)
5. Create UoW specs with dependencies
6. Compute input hashes per module
7. Detect incremental changes (skip unchanged UoWs)
8. Submit to orchestrator for parallel execution
9. Collect results, write summaries
```

**Source**: `go/cli/eac/impl/build/framework.go`, `go/cli/eac/impl/build/build.go`

### Module Scope Resolution

The build command accepts explicit module names or builds all modules:

```bash
eac build                    # All modules
eac build eac eac-ext    # Specific modules
eac build --module eac   # Single module
```

When specific modules are requested, the framework resolves their
`depends_on` graph and includes all transitive dependencies in the build
scope (unless `--skip-depm` is used).

### Unit Workers

Each builder (Go, buildx, mkdocs, etc.) implements the build logic for its
component type. The unit worker dispatches to the appropriate builder based
on the UoW's tool field.

| Tool       | Builder                    | Builds                          |
|------------|----------------------------|---------------------------------|
| `go`       | `GoHandler`                | Go binaries and libraries       |
| `buildx`   | `BuildxHandler`            | Docker images via buildx        |
| `mkdocs`   | `MkDocsHandler`            | Documentation sites             |
| `scripts`  | `ScriptsHandler`           | Bash/PowerShell scripts         |
| `npm`      | `NpmHandler`               | TypeScript/JavaScript packages  |
| `structurizr` | `StructurizrHandler`    | C4 architecture diagrams        |

**Source**: `go/cli/eac/impl/build/unit_worker.go`,
`go/cli/eac/impl/build/builders/`

---

## Builders

### Go Builder

Compiles Go binaries with cross-compilation support.

**Key features**:

- Cross-platform builds via `GOOS`/`GOARCH` matrix
- UPX compression for reduced binary size
- Version injection via `-ldflags`
- `GOWORK=off` for isolated module compilation

**Output**: `out/build/{module}/go-go/{binary}-{os}-{arch}`

**Source**: `go/cli/eac/impl/build/builders/go.go`

### Buildx Builder

Builds Docker images using `docker buildx build`.

**Key features**:

- Multi-platform builds (`linux/amd64`, `linux/arm64`)
- BuildKit cache support (GHA cache, registry cache)
- Local builds with `--load`, CI builds with `--push`
- Tag management with template variables

**Command construction**:

```text
docker buildx build
  --builder {builder}
  --platform {platforms}
  --tag {tags}
  --file {dockerfile}
  [--load | --push]
  [--cache-from ... --cache-to ...]
  {context}
```

The build context, Dockerfile path, platforms, tags, and cache settings all
come from the module's `docker_build` configuration, with template defaults
provided by the module's blueprint template (e.g., `container-multiarch`).

**Source**: `go/cli/eac/impl/build/builders/buildx.go`

### Other Builders

- **MkDocs** (`mkdocs.go`): Runs mkdocs-render-oci container for HTML/PDF
- **Structurizr** (`structurizr.go`): Generates C4 diagram PNGs
- **DrawIO** (`drawio.go`): Renders `.drawio` files to PNG
- **Scripts** (`scripts.go`): Executes bash/pwsh build scripts
- **PDF** (`pdf.go`): Generates PDF documents from markdown

---

## Output Tracking

### Output Tracker

During build execution, the `OutputTracker` records artifacts as they are
produced. Each artifact gets a SHA256 hash, type, and path recorded in the
UoW manifest.

**Source**: `go/core/output/tracker.go`

### Output Reader

After builds, the `OutputReader` loads manifests from disk to answer
queries like:

- "What was the input hash for this UoW?"
- "When was this UoW last executed?"
- "What artifacts did this UoW produce?"

This is used by the cache system to determine whether a UoW needs to
re-execute.

**Source**: `go/core/output/reader.go`

### Output Aggregation

Build summaries aggregate results across all UoWs:

- Total/passed/failed/cached counts
- Duration per UoW
- Artifact inventory

**Source**: `go/core/output/reader.go`, `go/core/output/types.go`

---

## Related Documentation

- [Cache System](./cache-system.md) — Input hashing and incremental builds
- [Component Resolution](./component-resolution.md) — How components become UoWs
- [Component Types](./component-kinds.md) — Component type definitions
- [Dependencies](./dependencies.md) — Module dependency graph
