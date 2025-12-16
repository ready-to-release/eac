# templates Commands

{{ page_breadcrumb() }}

## Overview

The **templates** category contains 6 commands for installing project templates for documentation, AI prompts, reports, and specifications.

## Commands

| Command                                              | Purpose                                      |
| ---------------------------------------------------- | -------------------------------------------- |
| [templates install-docs](../templates/install-docs.md) | Install documentation templates            |
| [templates install-ai](../templates/install-ai.md)   | Install AI prompt templates                  |
| [templates install-reports](../templates/install-reports.md) | Install report templates          |
| [templates install-specs](../templates/install-specs.md) | Install specification templates        |

## Common Use Cases

### Install Documentation Templates

```bash
# Install to default location
r2r templates install docs

# Install to custom location
r2r templates install docs --destination ./custom-docs
```

### Install AI Prompt Templates

```bash
r2r templates install ai
```

### Install Report Templates

```bash
r2r templates install reports --debug
```

### Install Specification Templates

```bash
r2r templates install specs
```

## Key Features

- Install documentation templates for consistent project structure
- Install AI prompt templates for code generation commands
- Install report templates for test and build summaries
- Install specification templates for compliance testing
- Debug logging support for troubleshooting

## See Also

- [create design](../create/design.md) - AI architecture design
- [create spec](../create/spec.md) - AI specification generation
- [create commit-message](../create/commit-message.md) - AI commit messages
- [validate markdown](../validate/markdown.md)
- [validate books](../validate/books.md)

{{ diataxis_footer() }}
