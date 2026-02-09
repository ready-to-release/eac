# Component Resolution

Component resolution is the process that turns module contracts (YAML
configuration) into executable Units of Work (UoWs). It determines what
gets built, in what order, and with which tool.

## Overview

```text
repository.yml          blueprints.yml         component-types.yml
      │                       │                        │
      └───────────┬───────────┘                        │
                  │                                    │
          Component Resolver                           │
                  │                                    │
                  ├── Merge blueprint defaults ◄────────┘
                  ├── Auto-discover components
                  ├── Resolve tool chains
                  ├── Apply build_after ordering
                  │
                  ▼
          List of UoW Specs (ready for orchestrator)
```

---

## Resolution Steps

### 1. Load Module Contract

The resolver reads the module's `components` block from `repository.yml`.
Each component entry has a **name** (the YAML key), an optional **type**
override, and optional **root**, **patterns**, and **build** configuration.

```yaml
modules:
  - moniker: eac
    template: go-exe
    components:
      go:
        root: go/cli/eac
        build:
          binary_name: eac
      godog:
        root: go/cli/eac/specs
```

### 2. Apply Blueprint Template

If the module declares a `template` (e.g., `go-exe`, `container-multiarch`),
the resolver loads the template from `blueprints.yml` and merges its
defaults with the module's explicit configuration.

Templates define:

- Default components (auto-added if not explicitly listed)
- Default patterns for each component type
- Build configuration (artifact matrix, docker_build settings)
- Component ordering (`build_after` relationships)

**Source**: `contracts/core/0.1.0/schemas/defaults/blueprints.yml`

**Example** (`go-exe` template):

```yaml
go-exe:
  components:
    go:
      files:
        source: ["**/*.go"]
        config: ["go.mod", "go.sum"]
    markdown: {}
    godog: {}
    gherkin: {}
    structurizr: {}
```

The module's explicit components override template defaults. Components
defined only in the template are auto-added.

### 3. Auto-Discover Components

Some components are **auto-discovered** based on filesystem conventions:

| Component    | Discovery Rule                              |
|-------------|---------------------------------------------|
| `gherkin`   | `specs/{moniker}/**/*.feature` exists        |
| `structurizr` | `specs/{moniker}/.design/workspace.dsl` exists |
| `markdown`  | `{root}/**/*.md` exists                      |

Auto-discovery runs after template merging, so explicitly defined
components always take priority.

**Source**: `go/core/config/component_discovery.go`

### 4. Resolve Tool Chains

Each component type maps to one or more **tools**. The tool chain
determines what builder/runner/linter handles the component:

| Component Type | Build Tool   | Test Tools          | Lint Tool        |
|---------------|--------------|---------------------|------------------|
| `go`          | `go`         | `gotest`, `godog`   | `golangci-lint`  |
| `dockerfile`  | `buildx`     | —                   | `hadolint`       |
| `typescript`  | `npm`        | `mocha`, `cucumber` | `eslint`         |
| `markdown`    | —            | —                   | `markdownlint`   |
| `gherkin`     | —            | —                   | `gherkin-lint`   |
| `structurizr` | `structurizr`| —                   | `structurizr`    |

For each component × tool combination, the resolver creates a UoW spec:

```text
Module: eac, Component: go, Action: build
  → UoW: build:eac:go:go

Module: eac, Component: go, Action: test
  → UoW: test:eac:go:gotest
  → UoW: test:eac:go:godog  (if godog component exists)
```

**Source**: `go/core/resolver/tool_chain.go`

---

## Build Ordering

### build_after

Component types can declare `build_after` relationships in
`blueprints.yml`. This creates intra-module dependency edges between UoWs.

```yaml
# In blueprints.yml component-type definitions:
dockerfile:
  build_after: [go]

docs-site:
  build_after: [docs-drawio, docs-mermaid]

docs-pdf:
  build_after: [docs-site]
```

**Meaning**: Within the same module, a `dockerfile` component's build UoW
depends on the `go` component's build UoW completing first.

**Resolution**: `ComponentResolver.ResolveForBuild()` creates a dependency
graph from `build_after` declarations and adds `DependsOn` edges to the
UoW specs.

**Source**: `go/core/resolver/component_resolver.go`

### Cross-Module Dependencies

Module-level `depends_on` creates edges between **all** UoWs across
modules:

```yaml
modules:
  - moniker: eac-ext
    depends_on: [eac]
```

This means every UoW in `eac-ext` depends on every UoW in `eac`.
Injected by `injectModuleDependencies()` in the command framework.

### Combined Ordering Example

```text
Module: eac
  build:eac:go:go              ← no dependencies (within module)
  build:eac:gherkin:gherkin    ← no dependencies

Module: eac-ext (depends_on: [eac])
  build:eac-ext:dockerfile:buildx  ← depends on ALL eac UoWs

Module: docs (depends_on: [repository])
  build:docs:drawio:drawio         ← phase 1 (build_after: nothing)
  build:docs:mermaid:mermaid       ← phase 1
  build:docs:site:mkdocs           ← phase 2 (build_after: [drawio, mermaid])
  build:docs:reference:pdf         ← phase 3 (build_after: [site])
```

---

## Component Weight and Scheduling

Each resolved UoW gets a **weight** for capacity scheduling:

```text
weight = tool.resources.cpus × component.amp.{action}
```

| Tool        | Default CPUs | Typical Amp | Effective Weight |
|-------------|-------------|-------------|------------------|
| `go`        | 2           | 1.0         | 2                |
| `buildx`    | 4           | 1.0         | 4                |
| `mkdocs`    | 4           | 1.25        | 5                |
| `gotest`    | 2           | 1.0         | 2                |
| `structurizr` | 2         | 1.0         | 2                |

Higher-weight UoWs consume more of the CPU capacity pool, so fewer run
in parallel. This prevents resource exhaustion during heavy builds.

---

## Display Order

Components within a module are displayed in a deterministic order for
consistent output across `show-modules`, `show-build-summary`, etc.

The display order follows these rules:

1. Buildable components first (go, typescript, dockerfile)
2. Test components (godog, gotest)
3. Non-buildable components (gherkin, structurizr, markdown, yaml)
4. Alphabetical within each category

**Source**: `go/core/config/display_order.go`

---

## Related Documentation

- [Build Execution System](./build-execution.md) — Orchestrator and UoW lifecycle
- [Cache System](./cache-system.md) — Incremental change detection
- [Component Types](./component-kinds.md) — Component type definitions
