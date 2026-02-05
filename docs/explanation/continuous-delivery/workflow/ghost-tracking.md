# Ghost Tracking

## What is a Ghost?

A **ghost** is any file, directory, or code construct that follows a naming convention (default: `ghost-` prefix) to mark it as **dark launch code** - code that exists in the codebase and can be deployed to production, but remains inactive or hidden.

Ghosts are the practical implementation of [feature hiding](feature-hiding.md) strategies.

```
ghost-checkout-v2.go          # Ghost file
ghost-monitoring/             # Ghost directory
ghost-payment-processor.yaml  # Ghost configuration
```

---

## Why Use Ghosts?

### The Dark Launch Practice

Dark launching means deploying code to production before it's "released":

| Traditional Flow | Dark Launch Flow |
|-----------------|------------------|
| Develop → Test → Release → Deploy | Develop → Deploy (hidden) → Activate → Release |
| Deployment = Risk | Deployment = Routine |
| Big bang releases | Gradual activation |

Ghosts make dark-launched code **discoverable** and **manageable**.

### Ghost Use Cases

| Use Case | Example | Benefit |
|----------|---------|---------|
| **Incomplete features** | `ghost-new-dashboard/` | Integrate to trunk early, activate when ready |
| **Monitoring probes** | `ghost-l4-probe.go` | Hidden observability, always in production |
| **Alternative implementations** | `ghost-payment-v2.go` | A/B testing, canary releases |
| **Experimental code** | `ghost-ml-recommendations/` | Safe experimentation in production environment |
| **Staged migrations** | `ghost-api-v3/` | Gradual backend transitions |

### The Ghost Lifecycle

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Create     │ ──▶ │   Deploy     │ ──▶ │   Activate   │ ──▶ │   Graduate   │
│   ghost-*    │     │   (hidden)   │     │   (release)  │     │   (rename)   │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                            │                    │
                            ▼                    ▼
                     ┌──────────────┐     ┌──────────────┐
                     │   Monitor    │     │   Rollback   │
                     │   & verify   │     │   (instant)  │
                     └──────────────┘     └──────────────┘
```

1. **Create**: Name new code with ghost prefix
2. **Deploy**: Code goes to production but remains inactive
3. **Activate**: Enable via configuration or feature flag
4. **Graduate**: Rename to remove ghost prefix when stable
5. **Or Rollback**: Deactivate instantly without redeployment

---

## Ghost Naming Convention

### File and Directory Patterns

The ghost alias (default: `ghost`) is used as a prefix:

```
ghost-<name>.<ext>      # Files: ghost-feature.go, ghost-config.yml
ghost-<name>/           # Directories: ghost-monitoring/
ghost.<ext>             # Single ghost file: ghost.go
```

### Code-Level Ghosts (Future)

In addition to files and directories, code constructs can be named as ghosts:

```go
// Variables
var ghostMetricsEnabled = false

// Functions
func ghostCollectTelemetry() { ... }

// Types
type ghostPaymentProcessor struct { ... }
```

This enables finer-grained ghost tracking at the symbol level.

---

## Configuration

### Repository Configuration

Configure ghost tracking in `.eac/repository.yml`:

```yaml
ghost-tracking:
  ghost-alias: ghost    # The prefix (default: "ghost")
```

Note: No exclusion configuration is needed - the scanner uses git-tracked files only, so `.gitignore` patterns are automatically respected.

### Custom Alias

Teams can use a different term if "ghost" doesn't fit:

```yaml
ghost-tracking:
  ghost-alias: hidden   # Use "hidden-" prefix instead
```

```
hidden-feature.go       # Detected as ghost
hidden-api-v2/          # Detected as ghost
ghost-something.go      # NOT detected (wrong prefix)
```

### Git-Based Discovery

Ghost tracking uses git-tracked files only - it does NOT walk the filesystem.

**What this means:**
- Only files in the git index are scanned
- `.gitignore` patterns are automatically respected
- `vendor/`, `node_modules/`, `out/` are excluded if gitignored
- Untracked files are not discovered as ghosts

**Benefits:**
- Fast and consistent across environments
- CI-optimized (uses GitHub Trees API)
- No need for custom exclusion configuration
- Matches what will actually be committed/deployed

---

## Platform Commands

### Discovering Ghosts

List all ghosts in machine-readable format:

```bash
eac get ghosts              # YAML output (default)
eac get ghosts --as-json    # JSON output
eac get ghosts --as-toml    # TOML output
```

Filter by type or ownership:

```bash
eac get ghosts --type file       # Only ghost files
eac get ghosts --type directory  # Only ghost directories
eac get ghosts --module core     # Ghosts in core module
eac get ghosts --unowned         # Ghosts not in any module
```

### Viewing Ghost Report

Human-readable report with statistics:

```bash
eac show ghosts
```

Example output:

```
# Ghost Tracking Report

## Summary

| Metric              | Value |
|---------------------|-------|
| Total Ghosts        | 5     |
| Files               | 3     |
| Directories         | 2     |
| Modules with Ghosts | 2     |
| Unowned             | 1     |

### Configuration
- Prefix: `ghost`
- Patterns: [ghost-*, ghost.*]

## Ghosts by Module

### core (3)

| Path                          | Type      | Ghost Name    |
|-------------------------------|-----------|---------------|
| `go/core/ghost-metrics.go`    | file      | metrics.go    |
| `go/core/ghost-telemetry/`    | directory | telemetry     |

## Unowned Ghosts

| Path                    | Type | Ghost Name |
|-------------------------|------|------------|
| `scripts/ghost-deploy.sh` | file | deploy.sh |
```

---

## Ghost Data Structure

Each ghost tracked by the platform includes:

| Field | Description |
|-------|-------------|
| `path` | Relative path from repository root |
| `name` | Base filename or directory name |
| `type` | `file` or `directory` |
| `ghost_name` | Name without the ghost prefix |
| `module` | Owning module (if inside a component) |
| `component` | Owning component name |
| `mod_time` | Last modification time |

Example YAML output:

```yaml
- path: go/core/ghost-metrics.go
  name: ghost-metrics.go
  type: file
  ghost_name: metrics.go
  module: core
  component: go
  mod_time: 2024-01-15T10:30:00Z
```

---

## Best Practices

### When to Ghost

**Do use ghosts for:**

- Features integrated to trunk before completion
- Production monitoring that should remain hidden
- Alternative implementations being tested
- Staged migrations and rollouts

**Don't use ghosts for:**

- Temporary debugging code (use branches or delete after)
- Dead code that will never activate (just delete it)
- Test fixtures (use `testdata/` which is excluded)

### Ghost Hygiene

1. **Regular reviews**: Run `eac show ghosts` periodically
2. **Set expiration**: Add comments with intended graduation date
3. **Track ownership**: Ensure ghosts belong to modules
4. **Graduate promptly**: Rename when feature is stable
5. **Clean up failures**: Remove ghosts that won't ship

### Naming Tips

Be descriptive in ghost names:

```
# Good - clear purpose
ghost-checkout-v2-redesign/
ghost-payment-stripe-migration.go
ghost-api-rate-limiting/

# Poor - unclear
ghost-new/
ghost-feature.go
ghost-test/
```

---

## Integration with Feature Hiding Strategies

Ghosts complement the [feature hiding strategies](feature-hiding.md):

| Strategy | Ghost Role |
|----------|------------|
| **Code-Level** | Ghost files contain inactive code |
| **Configuration** | Ghost config files control activation |
| **Feature Flags** | Ghosts hold flag-gated implementations |
| **Branch by Abstraction** | Ghost implementations behind interfaces |

Example: Branch by Abstraction with Ghosts

```go
// Active implementation
type PaymentProcessor struct { ... }

// Ghost implementation (exists, not injected)
type ghostPaymentProcessorV2 struct { ... }

// Factory uses config to select
func NewProcessor(useV2 bool) Processor {
    if useV2 {
        return &ghostPaymentProcessorV2{}
    }
    return &PaymentProcessor{}
}
```

---

## Next Steps

- [Feature Hiding](feature-hiding.md) - Hiding strategies overview
- [Release Toggling](../release-toggling/index.md) - Feature flags for activation
- [Trunk-Based Development](trunk-based-development.md) - Why small batches require hiding
