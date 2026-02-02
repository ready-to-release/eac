# Create Commands

## Overview

The **create** category contains commands for AI-powered content generation and documentation creation.

## How AI Generation Works

All create commands use structured AI generation with format-specific validation:

### Supported Output Formats

Each command generates content in a specific structured format:

| Command          | Output Format                    | Validation                                      |
| ---------------- | -------------------------------- | ----------------------------------------------- |
| `commit-message` | Plaintext (conventional commits) | Format validation and retry                     |
| `spec`           | Gherkin (.feature files)         | Syntax, step definitions, and quality standards |
| `design`         | Structurizr DSL (workspace.dsl)  | DSL syntax validation via Structurizr CLI       |
| `pr`             | Markdown (GitHub PR format)      | Format validation and retry                     |
| `risk-profile`   | OSCAL Profile JSON               | JSON schema validation                          |
| `risk-assess`    | OSCAL Assessment Results JSON    | JSON schema validation                          |
| `squash-message` | Plaintext (commit messages)      | Format validation and retry                     |

### Generation Process

1. **AI Generation**: AI generates content following the specified output format
2. **Validation**: Output is validated against format-specific rules (syntax, schema, standards)
3. **Automatic Retry**: If validation fails, AI receives error feedback and regenerates improved output
4. **Quality Assurance**: Validated output is ready to use without manual syntax corrections

This architecture ensures all generated content is syntactically correct and follows best practices.

## Commands

<!-- book:category-commands create -->

## Common Use Cases

**Pull Request Creation**:

```bash
eac create pr
```

**Specification Generation**:

```bash
eac create spec "User can login with email and password"
```

**Architecture Documentation**:

```bash
eac create design src-auth
```

## Key Features

- AI-powered content generation using configured providers
- Integration with Git workflows
- BDD specification support
- OSCAL compliance documentation
- Architecture diagram generation

## See Also

- [AI Configuration](../init/init.md)
- [update Commands](./update.md)
- [templates Commands](./templates.md)
