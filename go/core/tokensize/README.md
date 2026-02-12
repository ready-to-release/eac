# tokensize

Estimates token counts for source files using a character-based heuristic
to identify files that may exceed AI context limits.

## Key Types

- **`Estimate`** -- Token estimation result with tokens, chars, bytes, lines
- **`EstimationMethod`** -- Algorithm identifier (currently `char/4`)

## Patterns

- Heuristic estimation: tokens approximated as character count divided by four
- Glob expansion: resolves doublestar patterns to absolute file paths

## Internal Structure

| File         | Responsibility                                          |
| ------------ | ------------------------------------------------------- |
| estimator.go | `EstimateFile`, `EstimateContent`, `ExpandGlobPatterns` |

## Dependencies

_No internal repository imports (leaf package)._

## Role in System

The `tokensize` package powers the `get token-size` command in `core`.
It provides a fast, dependency-free way to estimate whether source files
will fit within Claude's token window before submitting them for AI
generation.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- Consider adding a configurable multiplier per file extension (e.g., JSON is more token-dense than prose) for improved accuracy

### Assessed and Accepted
- `char/4` heuristic limitations are documented in `EstimateContent`'s doc comment
- Package is minimal (111 lines) and well-tested (191 lines)
