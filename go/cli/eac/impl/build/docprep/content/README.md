# content

Content processing phases for the docprep pipeline, handling command execution, inline content insertion, image conversion, and command help marker expansion.

## Key Types

- **`CommandExecutor`** -- Interface for running EAC commands and capturing output
- **`ToolCommandExecutor`** -- Production `CommandExecutor` using the EAC adapter
- **`CommandHelp`** -- Parsed help output structure for a command
- **`CommandInfo`** -- Command metadata from `get valid-commands`
- **`FlagArg`** -- Flag or argument with name and description
- **`CategoryStats`** -- Category metadata for documentation grouping

## Patterns

- Command fragment caching: Pre-built markdown fragments from a prior build step are preferred over live command execution
- Marker replacement: `<!-- book:insert NAME -->` markers in markdown are replaced with command outputs at matching positions
- Attr-list to HTML: Markdown images with `{attrs}` syntax are converted to `<img>` tags for MkDocs compatibility
- Width constraints: Image dimensions are enforced via HTML tag conversion with max-width attributes
- Help marker expansion: `<!-- clie:command NAME -->` markers are expanded inline with formatted command help sections

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | `ExecuteCommands` runs book-defined commands and collects outputs, with fragment fallback |
| cmdhelp.go | `ProcessCommandMarkers` expands command help markers with formatted help output |
| inline.go | `InsertInlineContent` replaces `book:insert` markers in target files with command outputs |
| images.go | `ConvertAttrListImages` and `ProcessImageWidthConstraints` for image handling |

## Dependencies

- `adapters/eac` -- EAC binary adapter for command execution
- `core/config` -- book configuration for command and inline source definitions
- `core/paths` -- markdown command fragment output paths
- `docprep/staging` -- `FileIndex` for markdown file iteration

## Role in System

The content package provides four phases in the docprep preprocessing pipeline within `eac`. Phase 4 converts attr-list images to HTML. Phase 6 executes commands defined in `books.yml` and collects their output. Phase 8 inserts command output at marker positions in markdown files. Phase 9 expands command help markers into formatted documentation sections. These transformations prepare dynamic content before MkDocs rendering.

## Code Health

### Tech Debt

- None identified

### Pain Points

- No test coverage for `commands.go`, `images.go`

### Optimization Opportunities

- None identified
