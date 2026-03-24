# show books

<!-- book:cmd show books -->

## Output Sections

A single table with the following columns:

| Column      | Description                                                           |
| ----------- | --------------------------------------------------------------------- |
| Name        | Book identifier                                                       |
| Output      | Output directory path                                                 |
| Modules     | Modules that reference this book (first module shown, others as `+N`) |
| Description | Truncated to 30 characters                                            |
| Copy        | Number of copy sources                                                |
| Cmd         | Number of command sources                                             |
| Inline      | Number of inline sources                                              |

If no books are configured, prints guidance to create `.eac/books.yml`.

## See Also

- [serve docs](../serve/docs.md) - Serve documentation
- [validate books](../validate/books.md)
