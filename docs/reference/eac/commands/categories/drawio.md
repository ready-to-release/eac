# DrawIO Commands

## Overview

The **drawio** category contains 6 commands for editing DrawIO diagram files embedded in PNG images.

## How DrawIO Editing Works

DrawIO diagrams are stored as `.drawio.png` files - PNG images with embedded XML metadata. This enables LLM-powered diagram editing through a decode-edit-embed workflow.

### Architecture

| Command  | Purpose                                    |
| -------- | ------------------------------------------ |
| `decode` | Extract and decode XML to human-readable   |
| `encode` | Encode human-readable XML to DrawIO format |
| `embed`  | Write encoded XML into PNG file            |
| `create` | Create new .drawio.png with blank content  |
| `info`   | Show diagram metadata                      |
| `render` | Render diagram XML to PNG image            |

### Editing Workflow

1. **Extract**: `drawio decode` extracts XML from PNG
2. **Edit**: Modify the human-readable XML (manually or via LLM)
3. **Encode**: `drawio encode` compresses XML to DrawIO format
4. **Embed**: `drawio embed` writes XML back into PNG

## Commands

<!-- book:category-commands drawio -->

## Common Use Cases

**Extract diagram for editing**:

```bash
r2r eac drawio decode --png diagram.drawio.png --output readable.xml
```

**Create a new diagram**:

```bash
r2r eac drawio create --output new-diagram.drawio.png
```

**Re-render after XML changes**:

```bash
r2r eac drawio render --xml edited.xml --output diagram.drawio.png
```

## Key Features

- Docker-based processing (drawio-cli container)
- Works with DinD (Docker-in-Docker) environments
- Supports stdin/stdout for Unix-style piping
- Creates PNG files if they don't exist

## See Also

- [create design](./create.md) - Generate Structurizr diagrams
- [serve design](./serve.md) - View Structurizr diagrams
