# Dependency Management

## Introduction

Deployable modules depend on shared code. How these dependencies are managed fundamentally shapes your development workflow, testing strategy, and release process.

Two strategies exist:

| Strategy     | When             | Version               | Update Timing       |
| ------------ | ---------------- | --------------------- | ------------------- |
| **Implicit** | Same repository  | Always current (HEAD) | Automatic           |
| **Pinned**   | Cross-repository | Explicit version      | Consumer-controlled |

Understanding both is critical for choosing between monorepo and polyrepo architectures.

---

## Implicit Dependencies

**Definition:** Direct consumption of modules within the same repository, always using the current version.

This is **the core reason monorepos exist**.

### How It Works

In a monorepo, modules import each other directly without version numbers:

```go
// api-service/main.go
import (
    "github.com/company/monorepo/shared/models"
    "github.com/company/monorepo/shared/auth"
)
```

There is no version. The import always resolves to the current code in the repository.

### Why This Matters

**Atomic Changes:**

When you modify shared code, all consumers are updated in the same commit:

```text
Commit: "refactor(models): rename User.ID to User.UserID"

Files changed:
  shared/models/user.go        ← Change definition
  api-service/handlers/user.go ← Update usage
  worker-service/jobs/sync.go  ← Update usage
```

All changes are tested together. No version bump. No coordination across repos.

**No Diamond Dependency Problem:**

In polyrepo, this nightmare occurs:

```text
api-service depends on shared-models@v1.2.3
api-service depends on shared-auth@v2.0.0
shared-auth depends on shared-models@v1.1.0  ← CONFLICT!
```

In monorepo with implicit dependencies: **impossible**. There is only one version of shared-models.

**Immediate Feedback:**

Breaking changes are caught immediately:

```text
1. Developer changes shared/models/user.go
2. CI runs ALL tests for ALL modules
3. api-service tests fail immediately
4. Developer fixes api-service in same commit
5. Everything passes, single merge to trunk
```

No waiting for downstream repos to update. No "works in isolation, breaks in production."

### The Monorepo Contract

Implicit dependencies create a contract:

- **Shared code must not break consumers** (or must update them atomically)
- **All modules are tested together** in every pipeline run (optimized via incremental builds)
- **No versioning overhead** for internal modules
- **Single source of truth** for all code

This is why organizations adopt monorepos despite tooling complexity.

---

## Pinned Dependencies (Pin and Stitch)

**Definition:** Explicit version locking for dependencies outside the repository.

Required when:

- Consuming packages from external registries
- Depending on modules in other repositories (polyrepo)
- Publishing packages consumed by external systems

### How It Works

**Pin:** Lock dependencies to specific versions in your manifest:

```go
// go.mod
require (
    github.com/external/library v1.2.3
    github.com/other-team/shared-api v2.1.0
)
```

**Stitch:** At build time, resolve pinned versions and bundle into your artifact:

```text
1. Read go.mod
2. Download external/library@v1.2.3
3. Download other-team/shared-api@v2.1.0
4. Build api-service with exact versions
5. Produce immutable artifact
```

### Why Pin and Stitch

**Reproducible Builds:**

Same commit produces identical artifact regardless of when built. External packages don't change underneath you.

**Consumer Control:**

You decide when to update:

```text
external/library releases v1.3.0
├── api-service stays on v1.2.3 (not ready)
├── worker-service updates to v1.3.0 (needs feature)
└── batch-service stays on v1.2.3 (low priority)
```

No forced updates that destabilize multiple systems.

**Safe Rollbacks:**

Rolling back deploys the exact previous artifact with its bundled dependencies. No risk of dependency changes affecting rollback.

### Updating Pinned Dependencies

When updating:

```bash
# Create topic branch
git checkout -b chore/update-external-library

# Update the pin
go get github.com/external/library@v1.3.0

# Run full validation
go test ./...

# Merge via normal flow (Stages 2-4)
```

The update goes through standard validation, ensuring no regressions.

---

## Choosing Your Strategy

### Use Implicit Dependencies When

- Modules are in the same repository
- Teams can coordinate changes
- You want atomic cross-module changes
- You want immediate feedback on breaking changes

**This is the monorepo model.**

### Use Pinned Dependencies When

- Dependencies are external (npm, NuGet, PyPI)
- Modules are in separate repositories
- You publish packages for external consumption
- Teams need independent release cadences

**This is the polyrepo model (or external dependencies in any model).**

### Hybrid: Monorepo with External Dependencies

Most real systems use both:

```text
monorepo/
├── shared/models/       ← Implicit (same repo)
├── shared/auth/         ← Implicit (same repo)
├── api-service/
│   ├── go.mod           ← Pins external deps
│   └── main.go          ← Imports shared/ implicitly
└── worker-service/
    ├── go.mod           ← Pins external deps
    └── main.go          ← Imports shared/ implicitly
```

Internal modules: implicit. External packages: pinned.

---

## Anti-Patterns

### Implicit Dependencies Across Repository Boundaries

**This is the worst of both worlds.**

When teams split into multiple repositories but continue consuming each other without version pinning:

```go
// repo: api-service
// BAD: Importing from another repo without version pin
import "github.com/company/shared-models/user"  // Which version?!
```

**What goes wrong:**

| Problem                 | Consequence                                                                |
| ----------------------- | -------------------------------------------------------------------------- |
| Non-reproducible builds | Same commit produces different artifacts on different days                 |
| "It worked yesterday"   | Shared repo changed, broke your build without any change on your side      |
| Impossible rollbacks    | Rolling back your repo doesn't roll back the shared repo                   |
| Cascading breaks        | One breaking change ripples across all consuming repos unpredictably       |
| No testing guarantee    | Shared repo CI doesn't test your repo; your CI tests against moving target |

**The rule is absolute:**

- **Same repository → Implicit dependencies (always HEAD)**
- **Different repositories → Pinned dependencies (explicit versions)**

There is no middle ground. If you split repositories, you MUST pin versions.

If pinning feels like too much overhead, that's a signal you should be in a monorepo.

**The decision tree:**

```text
Do modules need to change atomically together?
├── Yes → Same repository (implicit)
└── No → Different repositories (pinned)

Are you pinning versions between repos?
├── Yes → Correct polyrepo usage
└── No → BROKEN. Fix immediately.
```

### Pinning Internal Monorepo Modules

```go
// BAD: Treating monorepo modules like external packages
require (
    github.com/company/monorepo/shared/models v1.2.3
)
```

This defeats the purpose of a monorepo. Use implicit imports.

### Version Ranges for External Dependencies

```yaml
# BAD: Non-deterministic
dependencies:
  external-lib: "^1.2.0"  # Could be 1.2.3 or 1.9.0
```

Pin to exact versions for reproducibility.

### Latest Tags

```yaml
# BAD: Changes without code change
dependencies:
  external-lib: latest
```

Builds become non-deterministic.

---

## Relationship to CD Model

**Stage 4 (Commit):**

- Implicit deps: All modules built and tested together
- Pinned deps: External versions locked in artifact

**Stage 5-7 (Testing):**

- Implicit deps: Integration tests run against current shared code
- Pinned deps: Tests validate specific external versions

**Rollback:**

- Implicit deps: Roll back entire monorepo state
- Pinned deps: Previous artifact contains bundled external deps

---

## Summary

| Aspect             | Implicit (Monorepo) | Pinned (Cross-repo)  |
| ------------------ | ------------------- | -------------------- |
| Version            | Always HEAD         | Explicit version     |
| Updates            | Automatic           | Consumer-controlled  |
| Breaking changes   | Caught immediately  | Caught when updating |
| Diamond dependency | Impossible          | Possible             |
| Coordination       | Required            | Independent          |
| Overhead           | Lower               | Higher               |

**Implicit dependencies are why monorepos exist.** They eliminate version management overhead and enable atomic cross-module changes.

**Pinned dependencies are required for external code.** They provide reproducibility and consumer control.

Most systems use both: implicit for internal modules, pinned for external packages.

---

## Next Steps

- [Trunk](../core-concepts/trunk.md) - Monorepo vs polyrepo patterns
- [Deployable Modules](../core-concepts/deployable-modules.md) - What gets built and versioned
- [Branching Strategies](branching-strategies.md) - How code flows through stages
