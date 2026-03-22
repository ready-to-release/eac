# templates Commands

## Overview

The **templates** category contains commands for installing project templates for documentation,
AI prompts, reports, and specifications.

## Commands

<!-- book:category-commands templates -->

## Common Use Cases

### Install Documentation Templates

```bash
# Install to default location
eac templates install docs

# Install to custom location
eac templates install docs --destination ./custom-docs
```

### Install AI Prompt Templates

```bash
eac templates install ai
```

### Install Report Templates

```bash
eac templates install reports --debug
```

### Install Specification Templates

```bash
eac templates install specs
```

## Key Features

- Install documentation templates for consistent project structure
- Install AI prompt templates for code generation commands
- Install report templates for test and build summaries
- Install specification templates for compliance testing
- Debug logging support for troubleshooting

## See Also

- [create design](../commands/create/design.md) - AI architecture design
- [create spec](../commands/create/spec.md) - AI specification generation
- [get commit-message](../commands/get/commit-message.md) - AI commit messages
- [validate markdown](../commands/validate/markdown.md)
- [validate books](../commands/validate/books.md)
