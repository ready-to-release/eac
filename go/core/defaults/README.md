# defaults

Default values and path derivation for module contracts, serving as the single
source of truth for fallback values when modules lack explicit configuration.

## Patterns

- Forward-slash paths: All derived paths use `/` for cross-platform config file compatibility

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| defaults.go | Path derivation functions (`DesignPath`, `SpecsPath`, `WorkflowCIPath`, etc.) |
| cmd/serialize/ | CLI tool to serialize resolved module configs for comparison |

## Dependencies

No internal dependencies in the library files. The `cmd/` sub-packages import
`core/domain/modules` and `core/paths` but are standalone tools, not library code.

## Code Health

### Tech Debt
- defaults.go: Path-building functions (`DesignPath`, `SpecsPath`, etc.) use inline string concatenation of `"specs/"` rather than referencing `paths.SpecsDir`, coupling the two packages implicitly
