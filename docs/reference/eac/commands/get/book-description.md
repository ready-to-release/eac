# get book-description

<!-- book:cmd get book-description -->

## How It Works

1. Strips `.pdf`, `-dark`, and `-light` suffixes from the filename to derive the book name
2. Searches the books configuration for a matching book
3. Returns the book's description, title, or name (in that priority order)
4. If not found and `--default` is set, returns the default value
5. If not found and no default, exits with code 1

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Book found (or default used) |
| 1 | Book not found and no default provided |

## See Also

- [show books](../show/books.md)
- [validate books](../validate/books.md)
- [get Commands](../get/index.md)
