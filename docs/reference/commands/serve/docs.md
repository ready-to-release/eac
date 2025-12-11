# serve docs

{{ page_breadcrumb() }}

## Overview

**Command**: `serve docs [flags]`
**Purpose**: Start, stop, or reload MkDocs documentation server
**Category**: [serve](../categories/serve.md)

The serve docs command manages the MkDocs documentation server using Docker. It can start a server on a specified port, stop a running server, reload to pick up changes, and open the documentation in your browser.

## Syntax

```bash
serve docs [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--no-browser` | Don't open browser after starting server |
| `-p, --port` | Port number for MkDocs server (default: auto-allocated 9000-9999) |
| `--stop` | Stop the running MkDocs server |
| `--reload` | Reload documentation by restarting the running server |
| `--debug` | Enable debug mode with log streaming |
| `-h, --help` | Show help message |

## Examples

```bash
# Start server with auto-allocated port
serve docs

# Start server on specific port
serve docs --port 9001

# Start without opening browser
serve docs --no-browser

# Reload to pick up documentation changes
serve docs --reload

# Stop the running server
serve docs --stop

# Enable debug mode with log streaming
serve docs --debug
```

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
- [build books](../../other/build.md)

{{ diataxis_footer() }}
