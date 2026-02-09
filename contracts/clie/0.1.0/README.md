# clie

CLIE CLI contract schemas, EBNF grammar, and embedded defaults for CLI
configuration validation and command parsing.

## Key Types

- **`FS`** -- Embedded filesystem containing schema, grammar, and docs

## Patterns

- Embedded filesystem: `FS` bundles schema, EBNF, and markdown via `//go:embed`

## Internal Structure

| File / Sub-directory     | Responsibility                                |
| ------------------------ | --------------------------------------------- |
| embed.go                 | `FS` variable with `//go:embed` directives    |
| schemas/clie.schema.json | JSON Schema for clie.yml validation           |
| schemas/command.ebnf     | EBNF grammar for CLI command parsing          |
| schemas/clie.yml.md      | Reference documentation for CLI configuration |

## Dependencies

None -- this is a leaf contract module with no internal dependencies.

## Role in System

The `clie` package (moniker: contracts-clie) provides the embedded
schema and grammar that the CLI configuration loader uses for validation
and command parsing. All files are accessed through `FS` at runtime.

## Code Health

### Tech Debt

- No test file; add an embed_test.go verifying all three embedded files load from `FS` to catch path drift
- No port interfaces defined -- purely an embed carrier; if CLI config loading gains a port, it should live here

### Pain Points

- Three `//go:embed` directives in embed.go are fragile if schema files are renamed or moved; a single `schemas/*` glob would be simpler but less explicit
- No version constant correlating schema content to the 0.1.0 contract version

### Optimization Opportunities

- Add a 10-line embed_test.go that opens each embedded file -- trivial effort, catches silent breakage
- Consider a `SchemaVersion` constant to pair with the embedded JSON schema `$id` field -- trivial, aids diagnostics
