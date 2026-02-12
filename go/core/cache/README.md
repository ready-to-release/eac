# cache

Two-dimensional taxonomy (Level x Type) for fine-grained cache control via the
`--skip-cache` CLI flag, plus CI cache checking to skip module builds when a
successful CI run already exists at HEAD.

## Key Types

| Type | Purpose |
|------|---------|
| `Config` | Parsed cache control configuration with skip specs |
| `Spec` | Cache specification as a level:type pair |
| `Level` | Where a cache lives (local, remote, all) |
| `Type` | What kind of cache (registry, state, asset, layer, work, all) |
| `DefaultSkipSpecs` | Pre-defined specs for bare `--skip-cache` flag |
| `CICacheChecker` | Checks whether a module's CI build can be skipped because a successful build exists at HEAD |
| `CICacheResult` | Result of a CI cache check with module, cached flag, reason, and last run SHA |
| `CIRunQuerier` | Port interface for querying CI workflow run status from an external system |
| `MockCIRunQuerier` | Test double for `CIRunQuerier` with configurable workflow-to-SHA mappings |

## Patterns

- 2D taxonomy: Combines `Level` (local/remote) with `Type` (state/asset/layer/etc.) for targeted invalidation
- Parse and match: `ParseSpec` parses user input; `Spec.Matches` checks against concrete level/type pairs
- Safe defaults: `DefaultSkipSpecs` skips state and work caches while preserving expensive assets
- Validation: `Spec.Validate` rejects invalid combinations like `remote:work`
- Port interface: `CIRunQuerier` decouples CI cache checking from GitHub API; callers inject implementations

## Internal Structure

| File | Purpose |
|------|---------|
| `cache.go` | `Level`, `Type`, `Spec` types with parsing and matching |
| `config.go` | `Config` with `ShouldSkip` and convenience methods |
| `defaults.go` | `DefaultSkipSpecs` for bare `--skip-cache` flag |
| `ci.go` | `CICacheChecker`, `CICacheResult`, `CIRunQuerier` interface, `MockCIRunQuerier` |

## Dependencies

No internal dependencies. This package depends only on the Go standard library.

## Role in System

`cache` provides the cache-control vocabulary for the `core` module. Commands
parse `--skip-cache=<spec>` into a `Config`, which build/test orchestrators
query to decide which caches to invalidate or bypass during execution. The
`Config` type exposes convenience methods like `ShouldSkipState`,
`ShouldForcePull`, and `ShouldForceNoCacheDocker` that translate the abstract
taxonomy into concrete build decisions. The `CICacheChecker` is used by CI
pipeline commands to skip module builds when the last successful CI run matches
the current HEAD SHA.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- config.go: `ShouldSkip` iterates the full `SkipSpecs` slice on every call; for current small spec counts this is fine, but if specs grow, a pre-computed `map[Level]map[Type]bool` at parse time would be O(1) (low priority, current usage is well under threshold)
- Good test coverage with cache_test.go, ci_test.go, config_test.go, and defaults_test.go
- All files are concise (all under 200 lines)
