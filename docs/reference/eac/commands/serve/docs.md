# serve docs

<!-- book:cmd serve docs -->

## Workflow

When working on documentation:

1. Start the server: `r2r eac serve docs`
2. Make changes to markdown files in `docs/`
3. Reload to see changes: `r2r eac serve docs --reload`

## Port Management

The server auto-allocates ports in the 9000-9999 range. For a specific port:

```bash
r2r eac serve docs --port 9725
```

## See Also

- [show books](../show/books.md) - List documentation books
- [validate books](../validate/books.md) - Validate book configuration
- [serve Commands](../categories/serve.md)
