# Serve Commands

Commands for starting development servers for documentation, architecture diagrams, and repository visualization.

**Key Features**:

- MkDocs documentation preview with live reload
- Structurizr Lite integration for architecture diagrams
- Docker-based diagram rendering
- Local development workflow support

## Commands in this Category

| Command                     | Purpose                               |
| --------------------------- | ------------------------------------- |
| [serve docs](./docs.md)     | Start or stop MkDocs server           |
| [serve design](./design.md) | View architecture diagrams in browser |

## Common Use Cases

### Documentation Server

```bash
# Start documentation server
eac serve docs

# Stop documentation server
eac serve docs --stop
```

### Architecture Diagrams

```bash
eac serve design src-auth
```

## See Also

- [create design](../create/design.md)
- [update design](../update/design.md)
- [validate design](../validate/design.md)
- [validate markdown](../validate/markdown.md)
