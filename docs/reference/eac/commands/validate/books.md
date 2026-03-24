# validate books

<!-- book:cmd validate books -->

## What It Checks

- **Schema validation** -- `books.yml` parses correctly against its JSON schema.
- **Duplicate book names** -- No two books share the same name.
- **Module references** -- Reports orphaned books not referenced by any module.
- **Command sources** -- All commands in `command` and `inline` sources reference valid EAC `show` commands.
- **Generated navigation** -- `section` and `insert_into` fields are present; `position` values are valid (`first`, `last`, or `after:<item>`).

## Common Errors

- **Duplicate book name** -- Two books share the same `name`. Rename one.
- **Unknown command** -- A command source references an invalid command.
- **Generated nav missing 'section'** -- A `generated_nav` entry lacks the required `section` field.
- **No module references this book** -- The book is orphaned; no module declares it.

## See Also

- [validate](./validate.md)
- [show books](../show/books.md)
- [serve docs](../serve/docs.md)
- [validate Commands](../validate/index.md)
