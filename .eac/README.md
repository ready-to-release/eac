# EAC Configuration

All EAC configuration lives in this `.eac/` directory as YAML files. Each file has a JSON Schema for validation (referenced via the `yaml-language-server` directive at the top of each file).

## Configuration Files

| File | Schema | Purpose |
|------|--------|---------|
| `repository.yml` | `repository.schema.json` | **Primary config** -- repo settings, paths, conventions, modules |
| `environments.yml` | `environments.schema.json` | Test environment definitions (L0--L4) |
| `testing-tags.yml` | `testing-tags.schema.json` | Test tag taxonomy (`@L0`, `@Manual`, `@gxp`, etc.) |
| `test-suites.yml` | `test-suites.schema.json` | Test suite definitions (unit, integration, acceptance, manual) |
| `books.yml` | `books.schema.json` | PDF/book generation configuration |
| `commands.yml` | `commands.schema.json` | CLI command documentation mapping |
| `blueprints.yml` | `blueprints.schema.json` | Component kinds and artifact matrices |
| `lint-providers.yml` | `lint-providers.schema.json` | Linter tool definitions |
| `testing-mocks.yml` | `testing-mocks.schema.json` | BDD test mock configuration |
| `ai-provider.yml` | `ai-provider.schema.json` | AI provider configuration |
| `timeouts.yml` | `timeouts.schema.json` | Timeout values for all operations |

Schemas are at `contracts/core/0.1.0/schemas/`.

## How Defaults Work

EAC uses a **three-layer merge** with ascending priority:

```
1. Base defaults          (embedded at compile time)
2. Type-specific defaults (based on repository.type)
3. User config            (.eac/*.yml - version controlled)
4. Personal overrides     (.eac/*.personal.yml - gitignored)
```

### Merge Rules

- **Scalars**: non-zero wins (user value replaces default when non-empty)
- **Maps**: deep-merged recursively (user keys added to/override default keys)
- **Arrays**: replaced wholesale (user array fully replaces default array)
- **Modules**: fully replaced (user modules completely override default modules)

### Default Files

Default values are embedded from `contracts/core/0.1.0/schemas/defaults/`:

| Default File | Applied When |
|-------------|-------------|
| `repository.yml` | Always (base defaults) |
| `repository-mono.yml` | `repository.type = "mono"` |
| `repository-poly.yml` | `repository.type = "poly"` |
| `repository-adjacent.yml` | `repository.type = "adjacent"` |
| `environments.yml` | Always |
| `testing-tags.yml` | Always |
| `test-suites.yml` | Always |
| `timeouts.yml` | Always |
| `blueprints.yml` | Always |
| `books.yml` | Always |
| `lint-providers.yml` | Always |
| `logging.yml` | Always |
| `risk-scoring.yml` | Always |
| `tool-config.yml` | Always |
| `ai-config.yml` | Always |
| `ai-provider.yml` | Always |
| `commands.yml` | Always |
| `registries.yml` | Always |

## Repository Types

The `repository.type` field controls type-specific defaults:

| Setting | `mono` | `poly` | `adjacent` |
|---------|--------|--------|------------|
| `max_branch_age_days` | 7 | 2 | 3 |
| `pr.merge_strategy` | squash | squash | merge |
| `versioning.constraint` | patch-only | unrestricted | unrestricted |

## repository.yml Reference

This is the primary configuration file. Only the `repository` and `modules` sections are required; everything else has sensible defaults.

### Repository Settings

```yaml
repository:
  type: mono                    # mono | poly | adjacent
  trunk_branch: main            # Main branch name
  max_branch_age_days: 7        # Branch age warning threshold
  schemes: [SemVer, CalVer]     # Allowed versioning schemes
  pr:
    delete_branch_on_merge: true
    merge_strategy: squash      # squash | merge
  versioning:
    constraint: patch-only      # patch-only | unrestricted | calver-only
  parallelism:
    ci: 8                       # Max parallel workers in CI
    devbox: 0                   # 0 = auto-detect from CPU/RAM
  remote:
    type: ""                    # Defaults to github
    owner: your-org             # Required: GitHub org or username
    repo: your-repo             # Optional: auto-detected from git
  ghost-tracking:
    ghost-alias: ghost          # Prefix for dark launch code detection
```

### Paths

```yaml
paths:
  specs_root: specs
  containers_root: containers
  templates: templates
  ai_prompts:
    system: templates/ai        # Shipped with repo
    team: .eac/templates/ai     # Team customizations
  out:
    root: out
    build: out/build
    test: out/test
    logs: out/logs
    scan: out/scan
    tools: out/tools
```

### Conventions

```yaml
conventions:
  godog_test: godog_test.go
  package_json: package.json
  changelog: CHANGELOG.md
  build_log: build.log
  build_timing: build-timing.txt
  test_timing: test-timing.txt
  specification: specification.feature
  risk_catalog: controls.catalog.json
  risk_controls_dir: .risk-controls
  design_dir: .design
  workspace_dsl: workspace.dsl
```

### Modules

Each module defines a build/test/release unit:

```yaml
modules:
  - moniker: my-module          # Unique identifier
    description: What it does
    group: my-group             # Optional: group multiple modules
    depends_on: [other-module]  # Build/release dependencies
    depends_on_ci: [dep]        # CI-only dependencies
    versioning:
      scheme: SemVer            # SemVer | CalVer | Implicit
      changelog: CHANGELOG.md   # Path to changelog (SemVer/CalVer only)
      release_type: published   # published | internal | bundle
    components:
      - type: go                # Component kind (from blueprints)
        root: path/to/code      # Source root directory
        name: optional-name     # Override auto-generated name
        build:
          binary_name: mybin
          artifact_matrix: cross-platform-upx
        patterns:
          source: ["**/*.go"]   # Override default source patterns
          config: ["**/*.yml"]
          exclude: ["vendor/**"]
          data: ["testdata/**"]
        design:
          root: specs/my-module/.design
          patterns: ["workspace.dsl"]
```

### Container Registries

```yaml
registries:
  ghcr.io:
    org: my-org                 # Required: your GitHub org
    cleanup:                    # Optional: prune policy
      enabled: true
      keep: 10
      min_age_days: 7
      image_tags:
        preserve: ["v*", "latest", "[0-9]*.[0-9]*.[0-9]*"]
        prune: ["sha-*", "dev-*", "pr-*", "ci"]
```

## Personal Overrides

Create `*.personal.yml` files to override any config locally without affecting the team:

```
.eac/repository.personal.yml
.eac/timeouts.personal.yml
.eac/ai-provider.personal.yml
.eac/testing-mocks.personal.yml
```

These files are gitignored and deep-merged on top of the team config. Example:

```yaml
# .eac/timeouts.personal.yml
# Longer timeouts for slow network
docker:
  image_pull: 10m
http:
  download: 10m
```

## Timeouts

All timeouts use Go duration syntax (`500ms`, `5s`, `2m`, `1h`). Override in `.eac/timeouts.yml` or `.eac/timeouts.personal.yml`:

| Category | Key | Default | Description |
|----------|-----|---------|-------------|
| docker | query | 5s | Daemon queries |
| docker | image_pull | 5m | Pull from registry |
| docker | image_build | 30m | Dockerfile build |
| docker | container_start | 60s | Container startup |
| docker | container_exec | 30m | Container execution |
| http | request | 30s | Standard HTTP requests |
| http | download | 5m | File downloads |
| ci | poll_interval | 10s | Status polling interval |
| ci | workflow_completion | 5m | Workflow completion wait |
| long_operations | security_scan | 30m | Security scan |
| long_operations | build | 30m | Build operation |
| long_operations | test | 30m | Test execution |
| long_operations | worker_timeout | 5m | Kill hung workers |
| long_operations | evidence_max_age | 24h | Evidence validity |

## CLI Commands

Inspect the loaded configuration:

```bash
# Structured YAML output of all config
clie get config

# Human-readable summary tables
clie show config

# Show config source files and layers (verbose)
clie show config --verbose

# Validate all contracts and schemas
clie validate config
```

## Validation

All config files are validated against JSON Schema (Draft 2020-12) at load time. Schemas are in `contracts/core/0.1.0/schemas/`. The `yaml-language-server` directive at the top of each YAML file enables IDE validation.

To enable schema support in your editor, ensure it supports the `yaml-language-server` comment format:

```yaml
# yaml-language-server: $schema=../contracts/core/0.1.0/schemas/repository.schema.json
```
