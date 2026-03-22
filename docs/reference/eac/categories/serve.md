# serve Commands

## Overview

The **serve** category contains commands for starting development servers for documentation, architecture diagrams, and repository visualization.

## Commands

<!-- book:category-commands serve -->

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

## Key Features

- MkDocs documentation preview with live reload
- Structurizr Lite integration for architecture diagrams
- Docker-based diagram rendering
- Local development workflow support

## See Also

- [create design](../commands/create/design.md)
- [update design](../commands/update/design.md)
- [validate design](../commands/validate/design.md)
- [validate markdown](../commands/validate/markdown.md)
