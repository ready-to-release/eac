# docs

Documentation contract schemas and embedded defaults for docs manifest
validation.

## Key Types

- **`FS`** -- Embedded filesystem containing the docs manifest JSON schema

## Patterns

- Embedded filesystem: `FS` bundles schemas via `//go:embed` for zero-file-IO access

## Internal Structure

| File / Sub-directory | Responsibility |
| --- | --- |
| embed.go | `FS` variable with `//go:embed` directive |
| schemas/manifest.schema.json | JSON Schema for docs manifest validation |

## Dependencies

None -- this is a leaf contract module with no internal dependencies.

## Role in System

The `docs` package (moniker: contracts-docs) provides the embedded JSON
schema used to validate documentation manifests. The docs command and its
update pipeline consume this schema to ensure manifest files conform to
the expected structure. Consuming modules read the schema from `FS` at
runtime without filesystem access.

## Code Health

### Tech Debt
- No test file; add an embed_test.go verifying `FS` can read `schemas/manifest.schema.json` to catch embed path drift
- No port interfaces defined -- this is purely an embed carrier; consider whether a `ManifestValidatorPort` belongs here or remains in the adapter

### Pain Points
- Package has no doc comments on `FS` explaining which schema version it bundles or how consumers should reference it
- If the schema evolves, there is no version constant to correlate schema content with package version

### Optimization Opportunities
- Add a one-line `SchemaVersion` constant alongside `FS` -- trivial effort, aids debugging when schemas change
- Add a 5-line embed_test.go confirming the file loads -- trivial effort, prevents silent embed breakage on path renames
