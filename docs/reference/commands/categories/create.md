# create Commands

{{ page_breadcrumb() }}

## Overview

The **create** category contains 7 commands for AI-powered content generation and documentation creation.

## Commands

| Command                                              | Purpose                                               |
| ---------------------------------------------------- | ----------------------------------------------------- |
| [create pr](../create/pr.md)                         | Generate pull request with AI-generated description   |
| [create spec](../create/spec.md)                     | Generate Gherkin specifications from natural language |
| [create squash-message](../create/squash-message.md) | Generate squash commit message from branch commits    |
| [create commit-message](../create/commit-message.md) | Generate AI-powered commit messages                   |
| [create design](../create/design.md)                 | Generate workspace.dsl for a module using AI          |
| [create risk-profile](../create/risk-profile.md)     | Create OSCAL profile from risk assessment             |
| [create risk-assess](../create/risk-assess.md)       | Update OSCAL assessment-results with evidence         |

## Common Use Cases

**Pull Request Creation**:

```bash
r2r eac create pr
```

**Specification Generation**:

```bash
r2r eac create spec "User can login with email and password"
```

**Architecture Documentation**:

```bash
r2r eac create design src-auth
```

## Key Features

- AI-powered content generation using configured providers
- Integration with Git workflows
- BDD specification support
- OSCAL compliance documentation
- Architecture diagram generation

## See Also

- [AI Configuration](../other/init.md)
- [update Commands](./update.md)
- [templates Commands](./templates.md)

{{ diataxis_footer() }}
