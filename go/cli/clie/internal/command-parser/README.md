# command-parser

EBNF-aware command-line argument parser that identifies the boundary between CLI framework arguments and container pass-through arguments.

## Key Types

- **`Parser`** -- Stateless parser with hardcoded grammar elements from EBNF schema
- **`ParsedCommand`** -- Structured result containing binary name, global flags, subcommand, extension name, Viper args, container args, and argument boundary index

## Key Functions

- **`Parse`** -- Parses raw `os.Args` into a `ParsedCommand` structure
- **`SplitArguments`** -- Convenience wrapper returning Viper and container argument slices
- **`GetEmbeddedSchema`** -- Returns the EBNF grammar loaded from contracts at init

## Patterns

- Embedded contract: EBNF grammar loaded from `contracts/clie/0.1.0` at init via `clie.FS.ReadFile`
- Argument boundary detection: Once a non-CLIE-flag is seen after the extension name, all subsequent args become container args
- Hardcoded grammar: Valid binary names, global flags, and subcommands are defined as maps (EBNF schema loaded but not dynamically parsed)

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

- `parser.go:43` -- TODO: Grammar elements are hardcoded instead of parsed from the embedded EBNF schema dynamically.
- `parser.go:240` -- TODO: `IsValidExtensionName` should be implemented according to the Identifier production from EBNF.
- `parser.go:287-292` -- Five TODO items for future enhancements including dynamic EBNF parsing, subcommand-specific flag parsing, and grammar caching.

### Pain Points

- The EBNF schema file is loaded at init but only stored as a string; none of the parsing logic actually uses it. All grammar rules are duplicated as hardcoded maps in `NewParser()`.

### Optimization Opportunities

- Using `golang.org/x/exp/ebnf` to parse the schema dynamically would eliminate the hardcoded grammar duplication and ensure the parser stays in sync with the contract.
