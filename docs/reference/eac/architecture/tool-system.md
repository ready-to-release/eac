# Tool System

The tool system provides a unified, pluggable composition layer that assigns
any tool -- whether a host-installed binary or a Docker container -- to any
component type for any operation (build, lint, scan, test, serve).

Configuration lives in `tool-config.yml`. Defaults ship with the contracts
package; projects override them in `.eac/tool-config.yml`.

---

## Configuration File

The tool configuration is loaded in two layers:

1. **Defaults** from `contracts/core/0.1.0/schemas/defaults/tool-config.yml`
2. **Project overrides** from `.eac/tool-config.yml`

Project overrides merge on top of defaults. Tools and bindings in the override
replace defaults by key; credentials are merged additively.

### Top-Level Sections

```yaml
executor-mode: auto            # Global resolution strategy
credentials: {}                # Environment variables forwarded to containers
namespaces: {}                 # Lazy-loading groups
system-tools: {}               # Host-installed tool definitions
container-tools: {}            # Docker-based tool definitions
tool-bindings: {}              # Per-tool resolution overrides
component-tools: {}            # Component type → tool assignments
environments: {}               # Environment-specific overrides
caches: {}                     # Cache directory definitions
test-type-mapping: {}          # Test type → component type mapping
```

---

## Executor Mode

The `executor-mode` setting controls how tools are resolved when no explicit
per-tool binding exists.

| Mode | Behavior |
|---|---|
| `auto` (default) | Try the system variant first; if unavailable, fall back to the container variant |
| `system` | Always use system tools. Fails if the tool has no system variant |
| `container` | Always use container tools. Reproducible builds with no host dependencies |

```yaml
# .eac/tool-config.yml
executor-mode: container
```

Per-tool bindings (see below) always override the global executor mode.

### When to Use Each Mode

- **auto** -- good default for local development. Uses fast native tools when
  installed, containers otherwise.
- **system** -- CI environments where all tools are pre-installed on the runner.
- **container** -- reproducible builds across machines. No host tool installation
  required beyond Docker.

---

## Tool Bindings

Tool bindings override the executor mode for individual tools. This lets you
force a specific tool to always use the system or container variant regardless
of the global mode.

```yaml
tool-bindings:
  go: system          # Always use native Go, even in container mode
  golangci-lint: container   # Always lint in a container
  npm-build: auto     # Use the global executor-mode default
```

Valid values: `auto`, `system`, `container`.

---

## System Tools

System tools execute as host-installed binaries. Each entry defines the binary
name, optional arguments, verification command, and version constraint.

```yaml
system-tools:
  go:
    description: "Go compiler and toolchain"
    binary: go
    verify:
      command: "go version"
      pattern: "go version go(\\d+\\.\\d+)"
    version: ">=1.21"
    platforms: [linux, darwin, windows]
    resources:
      cpus: 4

  npm-build:
    description: "NPM build (runs 'npm run build')"
    binary: npm
    args: ["run", "build"]
    requirements: [npm, tsc]
```

### Fields

| Field | Description |
|---|---|
| `binary` | Executable name or path |
| `args` | Default arguments passed to the binary |
| `verify` | Verification command and version-extraction pattern |
| `version` | Required version constraint (e.g. `>=1.21`) |
| `platforms` | Limit to specific OS platforms (`linux`, `darwin`, `windows`). Empty means all. |
| `requirements` | Other tools that must be available (verified recursively) |
| `resources.cpus` | Scheduling weight (higher = more capacity reserved) |

---

## Container Tools

Container tools run inside Docker containers. They can reference a local
Dockerfile, an external image, or a legacy container definition.

```yaml
container-tools:
  golangci-lint:
    description: "Go linter aggregator"
    image: golangci/golangci-lint
    tag: "v2.1.6"
    command: ["golangci-lint", "run", "--output.json.path", "stdout"]
    mounts:
      - source: "{workspace}"
        target: /workspace
      - source: "{workspace}/.cache/golangci-lint"
        target: /root/.cache/golangci-lint
    workdir: /workspace/{module-rel}
    resources:
      cpus: 2

  go-build:
    description: "Go build with CGO and race detection support"
    localPath: containers/cgo-oci
    command: ["go", "build", "./..."]
    mounts:
      - source: "{module}"
        target: /app
    workdir: /app
    resources:
      cpus: 4
```

### Container Source Options

A container tool must specify exactly one source:

| Option | Description |
|---|---|
| `localPath` | Path to a directory containing a Dockerfile (built locally, tagged `{dirname}:local`) |
| `image` + `tag` | External image reference. Tag is required and must be a pinned version (mutable tags like `latest`, `dev`, `main` are rejected) |
| `container` | Legacy reference to a container defined in `repository.yml` |

### Fields

| Field | Description |
|---|---|
| `command` | Command to run inside the container |
| `entrypoint` | Override the container entrypoint |
| `mounts` | Volume mounts (see Mount Variables below) |
| `workdir` | Working directory inside the container |
| `env` | Static environment variables |
| `host-env` | Host environment variables to forward (see Credentials) |
| `network` | Docker network mode |
| `user` | Container user |
| `privileged` | Run in privileged mode |
| `resources` | CPU, memory, and shared memory limits |

All container tools automatically require `docker` as a dependency.

---

## Mount Variables

Mount sources support placeholder variables that are resolved at execution time.
These map host paths into the container.

| Variable | Resolves To |
|---|---|
| `{workspace}` | Repository root on the host |
| `{module}` | Absolute path to the module being processed |
| `{module-rel}` | Module path relative to the workspace root |
| `{output}` | Output directory for build artifacts |
| `{go_cache}` | Go module cache directory |
| `{npm_cache}` | NPM cache directory |

Mount example:

```yaml
mounts:
  - source: "{workspace}"
    target: /workspace
    readonly: true
  - source: "{module}"
    target: /app
  - source: "{output}"
    target: /output
```

Each mount has a `source` (host path or placeholder), a `target` (container
path), and optional `readonly` and `type` (`bind`, `volume`, or `tmpfs`) fields.

---

## Resource Constraints

Tools can declare resource requirements that affect both scheduling and
container provisioning.

```yaml
resources:
  cpus: 4          # Scheduling weight + Docker --cpus
  memory: "4g"     # Docker --memory
  shm_size: "256m" # Docker --shm-size
```

- **cpus** -- used as a scheduling weight by the capacity-aware scheduler.
  For container tools, also passed as `--cpus` to Docker. Defaults to 1.
- **memory** -- container memory limit (Docker `--memory`).
- **shm_size** -- shared memory size (Docker `--shm-size`). Useful for
  browser-based tools like headless Chrome.

---

## Tool Verification

System tools can declare how to verify their availability. Verification runs
at startup for bootstrap tools and on first use for others.

```yaml
verify:
  command: "go version"
  pattern: "go version go(\\d+\\.\\d+)"
version: ">=1.21"
```

Two verification methods:

1. **Command-based** -- runs a command and extracts a version using a regex
   pattern. The extracted version is compared against the `version` constraint.
2. **Environment-variable-based** -- checks that specific environment variables
   are set.

```yaml
verify:
  env_vars: [ANTHROPIC_API_KEY]
  require: "any"   # "any" or "all"
```

Verification results are cached for the lifetime of the process. Requirements
are checked recursively -- if tool A requires tool B, both must pass
verification.

---

## Credentials and Host Environment Forwarding

Container tools run in isolation and do not inherit host environment variables
by default. The credentials system provides an explicit allowlist of variables
to forward.

### Global Credentials

Defined at the top level, these variables are forwarded to **all** container
tools:

```yaml
credentials:
  host-env:
    - GITHUB_TOKEN
    - GOPRIVATE
    - NPM_TOKEN
    - SEMGREP_APP_TOKEN
  ci-env:
    - CI
    - GITHUB_ACTIONS
    - GITHUB_SHA
    - GITHUB_REF
    - GITHUB_REPOSITORY
    - GITHUB_RUN_ID
```

- **host-env** -- credentials and tokens needed by tools at runtime.
- **ci-env** -- CI system variables useful for reproducibility and reporting.

### Per-Tool Host Environment

Individual container tools can forward additional variables using `host-env`:

```yaml
container-tools:
  semgrep:
    host-env:
      - SEMGREP_APP_TOKEN
      - SEMGREP_RULES
```

Per-tool entries are additive to the global credentials list.

### Security Model

Only explicitly named variables are forwarded. There is no passthrough of the
full host environment. Variable names are logged at debug level; values are
never logged.

---

## Namespaces

Namespaces control lazy loading of tool definitions. Only the `bootstrap`
namespace is loaded at startup; other namespaces load on demand when a tool
from that namespace is first requested.

```yaml
namespaces:
  bootstrap: [go, docker, git]
  go: [golangci-lint, gcc, go-build, go-test-race]
  docs: [mkdocs-build, mkdocs-live, structurizr-cli, drawio, mermaid-render]
  node: [npm, tsc, npm-build, npm-test, eslint]
  security: [trivy-sbom, trivy-vuln, semgrep, zap]
  deploy: [az-bicep]
```

This keeps startup fast -- a repository with 50+ tools only verifies Docker,
Go, and Git on boot.

---

## Component Tool Assignments

The `component-tools` section maps component types to tools for each operation.
This is how the system knows which tool to use when building a `go` component
vs. a `typescript` component.

```yaml
component-tools:
  go:
    builder: go-build
    linter: golangci-lint
    tester: go-test-race
    scanners:
      - trivy-sbom
      - trivy-vuln

  typescript:
    builder: npm-build
    linter: eslint
    tester: npm-test
```

Operations that support multiple tools (`linters`, `scanners`, `servers`) accept
a list. Single-tool operations (`builder`, `tester`) accept a single tool name.

### Resolution Order

When resolving which tool to use for a component type and operation, the system
checks layers in this order (highest priority first):

1. **CLI overrides** -- flags passed on the command line
2. **Environment config** -- the active environment's overrides
3. **Project config** -- `.eac/tool-config.yml`
4. **Defaults** -- built-in contract defaults

---

## Environment Overrides

Different environments can override tool assignments. The environment is
auto-detected from CI variables or can be set explicitly.

```yaml
environments:
  ci:
    component-tools:
      go:
        builder: go-build
        linter: golangci-lint

  windows:
    component-tools:
      typescript:
        builder: npm-build
```

Auto-detected environments: `ci` (any CI system detected), `windows` (Windows
host), `local` (default fallback).

---

## Serve Configuration

Tools used for `serve` operations (local development servers) have additional
configuration:

```yaml
container-tools:
  mkdocs-live:
    localPath: containers/mkdocs-dev-oci
    serve:
      container_port: 8000
      host_port_range: "9000-9999"
      watch_enabled: true
      watch_paths: ["docs/", "mkdocs.yml"]
      healthcheck_url: /
      auto_open_browser: true
      restart_policy: unless-stopped
```

Defaults applied when `serve` is configured:

- `host_port_range`: `9000-9999`
- `restart_policy`: `unless-stopped`

---

## Caches

Cache directories can be declared for reuse across tool runs:

```yaml
caches:
  go-modules:
    path: ".cache/go"
    description: "Go module download cache"
  go-build:
    path: ".cache/go-build"
    description: "Go build cache"
```

These are referenced in mount configurations to persist caches between runs.

---

## Project Override Example

A minimal `.eac/tool-config.yml` that forces container mode and adds a
credential:

```yaml
executor-mode: container

credentials:
  host-env:
    - ARTIFACTORY_TOKEN

tool-bindings:
  go: system    # Keep Go native even in container mode

container-tools:
  custom-lint:
    image: my-registry.example.com/custom-lint
    tag: "2.1.0"
    command: ["lint", "--format=json"]
    mounts:
      - source: "{module}"
        target: /app
    workdir: /app
```

---

## Related Documentation

- [Build Execution System](./build-execution.md) -- how tools are orchestrated
- [Component Types](./component-kinds.md) -- what component types exist
- [Component Resolution](./component-resolution.md) -- how contracts become executable UoWs
