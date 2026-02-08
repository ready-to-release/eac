# markdown

Validation utilities for Markdown files, including structure checking, code block
syntax validation, and heading hierarchy enforcement.

## Key Types

- **`Validator`** -- Configurable markdown file and directory validator
- **`ValidatorOptions`** -- Controls which validations are applied
- **`ValidationResult`** -- Per-file validation outcome with errors and warnings
- **`ValidationError`** -- Structured error with line number and message
- **`ValidationWarning`** -- Structured warning with line number and message
- **`CodeBlock`** -- Extracted fenced code block with language and content
- **`Section`** -- Extracted heading section with level and content

## Patterns

- AST walking: Uses goldmark parser to walk the markdown AST for structural analysis
- Language-specific validation: JSON and YAML code blocks are validated for syntax correctness
- Options struct: `ValidatorOptions` configures behavior without modifying the validator
- Directory traversal: `ValidateDirectory` walks a tree, respecting exclude patterns

## Internal Structure

| File | Responsibility |
| --- | --- |
| validator.go | Validator implementation, AST extraction, result printing |

## Dependencies

No internal dependencies. Uses external `goldmark` for markdown parsing and
`gopkg.in/yaml.v3` for YAML code block validation.

## Role in System

`markdown` supports the `validate-markdown` command in the `core` module,
providing reusable validation logic that can be applied to any markdown file
or directory tree in the repository. It parses markdown into an AST, extracts
code blocks and sections, validates embedded JSON/YAML for syntax correctness,
and checks heading hierarchy for proper progression. The `PrintResults` method
produces human-readable validation summaries for CLI output.

## Code Health

### Tech Debt
- validator.go:306-307 -- commented-out `case "go"` block; decide whether to implement Go syntax validation or remove the placeholder
- `extractCodeBlocks` and `extractSections` share near-identical AST walk boilerplate; a generic walk-and-collect helper would reduce duplication

### Pain Points
- Only JSON and YAML code blocks are validated; other common languages (TOML, HCL) silently pass, which may surprise users expecting comprehensive checks
- `sanitizeMessage` collapses spaces with a loop (`for strings.Contains`); a single `strings.Join(strings.Fields(...))` call would be clearer

### Optimization Opportunities
- `ValidateDirectory` uses `filepath.Walk` (deprecated in favor of `filepath.WalkDir`); migrating avoids an extra `os.Stat` per entry (trivial, mechanical change)
- Code block extraction allocates intermediate `bytes.Buffer` per block; pre-sizing the buffer from `Lines().Len()` would reduce small allocations in large files (low priority)
