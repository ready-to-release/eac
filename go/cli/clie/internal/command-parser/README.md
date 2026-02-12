# command-parser

EBNF-aware command-line argument parser that identifies the boundary between CLI framework arguments and container pass-through arguments.

## Key Types

- **`Parser`** -- Stateless parser with grammar elements extracted from embedded EBNF schema
- **`ParsedCommand`** -- Structured result containing binary name, global flags, subcommand, extension name, Viper args, container args, and argument boundary index

## Key Functions

- **`Parse`** -- Parses raw `os.Args` into a `ParsedCommand` structure
- **`SplitArguments`** -- Convenience wrapper returning Viper and container argument slices
- **`GetEmbeddedSchema`** -- Returns the EBNF grammar loaded from contracts at init

## Patterns

- Embedded contract: EBNF grammar loaded from `contracts/clie/0.1.0` at init via `clie.FS.ReadFile`
- Schema-driven grammar: Global flags and subcommands are extracted from the EBNF schema at init using regex-based production rule parsing
- Argument boundary detection: Once a non-CLIE-flag is seen after the extension name, all subsequent args become container args

## Internal Structure

| File      | Responsibility                                                        |
| --------- | --------------------------------------------------------------------- |
| parser.go | Parser and ParsedCommand types, EBNF loading, argument boundary logic |

## Dependencies

- `contracts/clie/0.1.0` -- Embedded EBNF grammar schema

## Role in System

The command parser runs early in PersistentPreRunE to provide structured command information before Cobra processes arguments. Its primary purpose is detecting where CLI-framework arguments end and container pass-through arguments begin for the `run` command. This boundary is critical because the run command uses `DisableFlagParsing` to avoid Cobra intercepting flags meant for the container.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- `parser.go` is 333 lines, making it a moderately large file for argument parsing logic.

### Optimization Opportunities

- None identified.
