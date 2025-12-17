# Serve docs

<!-- book:cmd serve docs -->

## Workflow

### Making Documentation Changes

When working on documentation, use the `--reload` flag to quickly see your changes:

1. Start the documentation server:

   ```bash
   serve docs
   ```

2. Make changes to markdown files in `docs/`

3. Reload to see changes:

   ```bash
   serve docs --reload
   ```

4. Browser automatically refreshes to show updated content

### Port Management

The server automatically allocates ports in the 9000-9999 range. If you need a specific port:

```bash
serve docs --port 9725
```

If a server is already running on a different port, you'll need to stop it first or use `--reload` to restart on the same port.

## See Also

- [show books](../show/books.md)
- [validate books](../validate/books.md)
- [build](../build/build.md)
