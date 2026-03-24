# validate markdown

<!-- book:cmd validate markdown -->

## What It Checks

- Valid markdown syntax.
- Proper heading hierarchy (no skipped levels).
- Valid embedded code blocks (JSON, YAML syntax).

## Common Errors

- **Heading hierarchy skip** -- A heading jumps levels (e.g., `##` to `####`) without an intermediate heading.
- **Invalid code block** -- A fenced code block tagged as JSON or YAML contains syntax errors.

## See Also

- [validate](./validate.md)
- [validate Commands](../validate/index.md)
