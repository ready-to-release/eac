# serve Commands

## Overview

The **serve** category contains commands for starting development servers for documentation, architecture diagrams, and repository visualization.

## Commands

<!-- book:category-commands serve -->

## Common Use Cases

### Documentation Server

```bash
# Start documentation server
r2r eac serve docs

# Stop documentation server
r2r eac serve docs --stop
```

### Architecture Diagrams

```bash
r2r eac serve design src-auth
```

## Key Features

- MkDocs documentation preview with live reload
- Structurizr Lite integration for architecture diagrams
- Docker-based diagram rendering
- Local development workflow support

## See Also

- [create design](../create/design.md)
- [update design](../update/design.md)
- [validate design](../validate/design.md)
- [validate markdown](../validate/markdown.md)
