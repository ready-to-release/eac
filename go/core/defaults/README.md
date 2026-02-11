# defaults

Default values and path derivation for module contracts, serving as the single
source of truth for fallback values when modules lack explicit configuration.

## Key Types

- **`TypeDefaults`** -- Default file patterns and repo paths for a module type
- **`FilesDefaults`** -- Default source, config, asset, and test patterns
- **`RepoDefaults`** -- Default specs, test-impl, and design paths
- **`ModuleDefaults`** -- Fully resolved defaults for a single module
- **`TypeDefaultsApplier`** -- Interface for retrieving type-based defaults

## Patterns

- Variable substitution: `SubstituteVariables` replaces `{moniker}`, `{root}`, `{type}` placeholders in patterns
- Priority resolution: Explicit module values override type defaults, which override generic fallbacks
- Forward-slash paths: All derived paths use `/` for cross-platform config file compatibility
- Nil vs empty distinction: Nil slices mean "not set" (apply defaults), empty slices mean "explicitly none"

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| defaults.go | Path derivation functions (`DesignPath`, `SpecsPath`, `WorkflowCIPath`, etc.) |
| type_defaults.go | `TypeDefaults`, `ModuleDefaults`, `ResolveDefaults`, variable substitution |
| cmd/serialize/ | CLI tool to serialize resolved module configs for comparison |
| cmd/thin/ | CLI tool to remove redundant fields that match type defaults |

## Dependencies

No internal dependencies in the library files. The `cmd/` sub-packages import
`core/domain/modules` and `core/paths` but are standalone tools, not library code.

## Role in System

`defaults` provides the fallback value layer in the `core` module's contract
loading pipeline. When the module loader processes `repository.yml`, it calls
`ResolveDefaults` to fill in any unset fields, ensuring every module contract
has complete path and file pattern information. The `cmd/` tools support
development workflows for comparing resolved configs and thinning out redundant
entries in the repository configuration.

## Code Health

### Tech Debt
- type_defaults.go: `ResolveDefaults` accepts 12 positional parameters; wrapping the current-value inputs into an options struct would improve call-site readability
- defaults.go: Path-building functions (`DesignPath`, `SpecsPath`, etc.) use inline string concatenation of `"specs/"` rather than referencing `paths.SpecsDir`, coupling the two packages implicitly

### Pain Points
- `TypeDefaults` in this package mirrors `config.TypeDefaults` to avoid import cycles; changes to one must be manually synchronized with the other

### Optimization Opportunities
- `SubstituteVariables` performs five sequential `strings.ReplaceAll` calls; a single-pass `strings.Replacer` would be marginally faster for patterns with many variables (trivial change, minimal benefit at current scale)
- The `cmd/serialize` and `cmd/thin` sub-packages are standalone CLI tools; moving them to a top-level `tools/` directory would clarify that they are not importable library code (organizational, no runtime impact)
