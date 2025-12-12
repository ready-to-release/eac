# serve Commands

{{ page_breadcrumb() }}

## Overview

The **serve** category contains 2 commands for starting development servers for documentation and architecture diagrams.

## Commands

| Command                            | Purpose                                                      |
| ---------------------------------- | ------------------------------------------------------------ |
| [serve docs](../serve/docs.md)     | Start or stop MkDocs server                                  |
| [serve design](../serve/design.md) | View architecture diagrams in browser using Structurizr Lite |

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

{{ diataxis_footer() }}
