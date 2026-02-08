# cache

Two-dimensional taxonomy (Level x Type) for fine-grained cache control via the
`--skip-cache` CLI flag.

## Key Types

- **`Config`** -- Parsed cache control configuration with skip specs
- **`Spec`** -- Cache specification as a level:type pair
- **`Level`** -- Where a cache lives (local, remote, all)
- **`Type`** -- What kind of cache (registry, state, asset, layer, work, all)
- **`DefaultSkipSpecs`** -- Pre-defined specs for bare `--skip-cache` flag

## Patterns

- 2D taxonomy: Combines `Level` (local/remote) with `Type` (state/asset/layer/etc.) for targeted invalidation
- Parse and match: `ParseSpec` parses user input; `Spec.Matches` checks against concrete level/type pairs
- Safe defaults: `DefaultSkipSpecs` skips state and work caches while preserving expensive assets
- Validation: `Spec.Validate` rejects invalid combinations like `remote:work`

## Internal Structure

| File | Responsibility |
| --- | --- |
| cache.go | `Level`, `Type`, `Spec` types with parsing and matching |
| config.go | `Config` with `ShouldSkip` and convenience methods |
| defaults.go | `DefaultSkipSpecs` for bare `--skip-cache` flag |

## Dependencies

No internal dependencies. This package depends only on the Go standard library.

## Role in System

`cache` provides the cache-control vocabulary for the `core` module. Commands
parse `--skip-cache=<spec>` into a `Config`, which build/test orchestrators
query to decide which caches to invalidate or bypass during execution. The
`Config` type exposes convenience methods like `ShouldSkipState`,
`ShouldForcePull`, and `ShouldForceNoCacheDocker` that translate the abstract
taxonomy into concrete build decisions.

## Code Health

### Tech Debt
- defaults.go:15 -- `DefaultSkipSpecs` is a package-level `var` (mutable slice); any caller could accidentally append to or mutate it; consider returning a copy from a function or using a frozen pattern
- config.go: `ShouldForcePull` and `ShouldForceNoCacheDocker` are thin wrappers that obscure the underlying `ShouldSkip` call; document or inline them to reduce indirection

### Pain Points
- None identified

### Optimization Opportunities
- `ShouldSkip` iterates the full `SkipSpecs` slice on every call; for the current small spec counts this is fine, but if specs grow, a pre-computed `map[Level]map[Type]bool` at parse time would be O(1) (low priority, current usage is well under threshold)
- None of the convenience methods on `Config` are generated; a code-generation approach could keep them in sync with new `Type` values automatically (deferred unless the taxonomy expands significantly)
