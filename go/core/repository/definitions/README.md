# repository/definitions

Enumerates, parses, and merges distributed `definitions.yml` files from across a repository into a single unified YAML structure. Each file's directory path determines its position in the merged output.

## Key Types

| Type | Purpose |
|------|---------|
| `DefinitionFile` | Represents a single `definitions.yml` with its filesystem path, parsed YAML content, and derived YAML path |

## Key Functions

| Function | Purpose |
|----------|---------|
| `EnumerateDefinitionFiles` | Walks a directory tree finding all `definitions.yml` files, parsing each into a `DefinitionFile` |
| `MergeDefinitions` | Merges multiple `DefinitionFile` entries into a single YAML document using directory-derived nesting |
| `ProcessDirectory` | Convenience function combining enumeration and merging in one call |

## Patterns

- **Directory-to-YAML-path mapping**: File at `foo/bar/definitions.yml` becomes nested under `foo.bar` in the merged output
- **Template skeleton exclusion**: Paths matching `*/templates/*/skeleton` are skipped to avoid parsing Handlebars syntax as YAML
- **Cross-platform path handling**: Normalizes backslashes to forward slashes for consistent behavior on Windows
- **Standard directory exclusions**: Skips `.git`, `node_modules`, `.vscode`, `.idea`, `out`

## Internal Structure

| File | Purpose |
|------|---------|
| `merger.go` | All types and functions: `DefinitionFile`, enumeration, merging, YAML path generation, nested value setting |

## Dependencies

None (uses only standard library and `gopkg.in/yaml.v3`).

## Role in System

Supports the definitions merge pipeline where teams maintain per-directory `definitions.yml` files that get combined into a single document. This enables distributed authoring of configuration data while producing a unified output for downstream consumption.

## Code Health

### Tech Debt

- None identified

### Pain Points

- Large test files: merger_edge_cases_test.go (333 lines), merger_unit_test.go (313 lines), merger_logic_test.go (299 lines), merger_test.go (297 lines)

### Optimization Opportunities

- None identified
