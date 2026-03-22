# Get book-description

<!-- book:cmd get book-description -->

Returns the description for a book PDF by matching its filename against the books configuration. Strips `-dark` and `-light` suffixes before matching, so `user-guide-dark.pdf` matches the `user-guide` book.

## Usage

```bash
eac get book-description <filename> [--default <value>]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `filename` | Yes | PDF filename to look up (e.g. `user-guide-dark.pdf`) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--default` | string | Default value to return if the book is not found |

## How It Works

1. Strips `.pdf`, `-dark`, and `-light` suffixes from the filename to derive the book name
2. Searches the books configuration for a matching book
3. Returns the book's description, title, or name (in that priority order)
4. If not found and `--default` is set, returns the default value
5. If not found and no default, exits with code 1

## Examples

```bash
# Get description for a book PDF
eac get book-description user-guide-dark.pdf

# Use a default if book not found
eac get book-description unknown.pdf --default "Documentation"
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Book found (or default used) |
| 1 | Book not found and no default provided |

## See Also

- [show books](../show/books.md)
- [validate books](../validate/books.md)
- [get Commands](../../categories/get.md)
