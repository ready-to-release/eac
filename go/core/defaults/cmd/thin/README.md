# defaults/cmd/thin

Standalone CLI tool that reads `repository.yml` and `module-types.yml`, then removes fields from module definitions that are redundant because they match their type defaults. Specifically targets empty `specs: []` entries that duplicate the type default.

## Key Functions

| Function | Purpose |
|----------|---------|
| `main` | Entry point: loads types and modules, removes redundant fields, writes updated `repository.yml` |
| `findRepoRoot` | Walks up directory tree looking for `.git` to locate repository root |

## Patterns

- **Schema thinning**: Removes explicitly-stated values that match implicit defaults, reducing config noise
- **Raw YAML manipulation**: Uses `map[string]interface{}` to preserve YAML structure while removing fields
- **Safe writes**: Adds header comment and writes back to same file

## Internal Structure

| File | Purpose |
|------|---------|
| `main.go` | CLI entry point, type defaults loading, redundancy detection, YAML write-back |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/paths` | `EACConfigPath` for locating `.eac/` configuration directory |

## Role in System

One-time migration utility used when introducing or changing module type defaults. After updating `module-types.yml` with new defaults, run this tool to clean up `repository.yml` by removing fields that now match the type default, keeping the config minimal.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: The tool only handles `specs: []` removal (lines 103-109 of `main.go`). Other redundant fields (source patterns, config patterns) are detected but not yet removed.
- **Optimization Opportunities**: Could be generalized to handle all redundant field types, not just specs.
