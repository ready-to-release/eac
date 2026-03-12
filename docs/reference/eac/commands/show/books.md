# Show Books

<!-- book:cmd show books -->

Displays all books defined in `books.yml` in a formatted table, including each book's name, output path, associated modules, description, and source counts.

## Usage

```bash
eac show books [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--format <format>` | string | Output format (default: `table`) |

## Output Sections

A single table with the following columns:

| Column | Description |
|--------|-------------|
| Name | Book identifier |
| Output | Output directory path |
| Modules | Modules that reference this book (first module shown, others as `+N`) |
| Description | Truncated to 30 characters |
| Copy | Number of copy sources |
| Cmd | Number of command sources |
| Inline | Number of inline sources |

If no books are configured, prints guidance to create `.eac/books.yml`.

## Examples

```bash
# List all configured books
eac show books
```

## See Also

- [serve docs](../serve/docs.md) - Serve documentation
- [validate books](../validate/books.md)
