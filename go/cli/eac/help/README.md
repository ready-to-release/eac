# help

Unified help printing for CLI commands, formatting usage, flags, and subcommand listings from registration metadata.

## Key Types

- **`BehaviorGroup`** -- Groups enable/disable flag pairs (e.g., `--with-tidy` / `--no-tidy`) for a single behavior
- **`PrintHelp`** -- Function rendering full help output from `CommandRegistration` metadata
- **`CategorizeFlags`** -- Separates behavior flags from regular flags for grouped display
- **`FormatDefault`** -- Formats default values with environment awareness (local vs CI contexts)
- **`PrintBehaviorFlags`** -- Renders behavior flag groups with enable/disable pairs and context-aware defaults

## Patterns

- Declarative flag categorization: Separates behavior flags (enable/disable pairs) from regular flags for grouped display
- Environment-aware defaults: Renders context-sensitive default values (e.g., "ON locally, OFF in CI") using `EnvDefaults` metadata
- Parent vs leaf rendering: Parent commands show grouped subcommands; leaf commands show flags and long description
- Shorthand formatting: Flags with shorthands render as `-s, --long-name` with optional default values
- Longest-first flag pair formatting: Behavior groups render as `--with-X / --no-X` with aligned descriptions
- Incomplete group filtering: `PrintBehaviorGroup` silently skips groups missing either the enable or disable flag

## Internal Structure

| File           | Responsibility                                                                                              |
| -------------- | ----------------------------------------------------------------------------------------------------------- |
| help.go        | `PrintHelp` renders full help output for parent and leaf commands, including examples and footer            |
| declarative.go | `CategorizeFlags`, `PrintBehaviorFlags`, `PrintBehaviorGroup`, and `FormatDefault` for behavior flag groups |

## Dependencies

- `clibase/registry` -- `CommandRegistration`, `FlagMetadata`, and subcommand group definitions

## Role in System

The help package provides the standard help rendering used by all CLI commands in `eac`. It reads `CommandRegistration` metadata from the registry to produce consistent, structured help output including grouped subcommands, behavior flag pairs, and usage examples.

For parent commands (e.g., `show`, `validate`), `PrintHelp` renders a grouped list of subcommands with their descriptions and a footer prompting for per-subcommand help. For leaf commands (e.g., `build`, `test`), it renders the long description, categorized flags (behavior flags first, then regular flags), and examples.

Commands delegate their `--help` output to `PrintHelp` rather than implementing custom help formatting. The `BehaviorGroup` type pairs enable/disable flags and presents them as a single line with context-aware default display, keeping help output concise while conveying environment-specific behavior differences.

Required flags are annotated with `[required]` in the output, and flags with non-empty defaults display them inline.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- None identified.

### Optimization Opportunities

- None identified.
