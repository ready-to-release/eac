# Use Documentation Templates

## What You'll Accomplish

Install documentation templates to bootstrap consistent project documentation structure.

## Prerequisites

- Repository initialized with EAC
- Template files available in `templates/` directory

## How to Use

Install templates by running the appropriate command for your template type:

```bash
# Install to default location
templates install docs

# Install to custom location (docs type only)
templates install docs --destination ./custom-docs
```

Templates are copied as-is with `<< .Variable >>` placeholders preserved. After installation, edit the files to replace placeholders with your project-specific values.

## Template Types

Available templates for installation:

- **docs** - Documentation templates (README, guides, references)
- **ai** - AI prompt templates (for code generation commands)
- **reports** - Report templates (test reports, build summaries)
- **specs** - Specification templates (Gherkin scenarios for compliance)
- **claude** - Claude Code configuration templates (agents, commands, skills, MCP setup)

## Example Scenario

Setting up documentation structure for a new project:

```bash
# Install documentation templates
templates install docs
# Using templates from templates/docs
# Installing templates to docs/reference...
# ✓ Documentation templates installed successfully to docs/reference

# Install AI prompt templates
templates install ai
# Using templates from templates/ai
# Installing templates to .r2r/eac/templates/ai...
# ✓ AI prompt templates installed successfully to .r2r/eac/templates/ai

# Install report templates
templates install reports
# Using templates from templates/reports
# Installing templates to .r2r/templates/reports...
# ✓ Report templates installed successfully to .r2r/templates/reports

# Review installed docs
ls docs/reference/
# README.md
# architecture.md
# implementation-plan.md
# operations/
```

## Installing Multiple Template Types

```bash
# Install docs to default location
templates install docs

# Install AI templates
templates install ai

# Install report templates with debug logging
templates install reports --debug

# Install specification templates
templates install specs

# Install Claude Code templates
templates install claude
```

## Template Destinations

| Template Type   | Default Destination       | Custom Path Support   |
| --------------- | ------------------------- | --------------------- |
| docs            | `docs/reference/`         | Yes (`--destination`) |
| ai              | `.r2r/eac/templates/ai/`  | No                    |
| reports         | `.r2r/templates/reports/` | No                    |
| specs           | `specs/risk-controls/`    | No                    |
| claude          | `.claude/`                | No                    |

## Common Issues

> **WARNING**: Template installation will overwrite existing files at the destination without confirmation. Always backup important files before installing templates.

| Problem                      | Solution                                                 |
| ---------------------------- | -------------------------------------------------------- |
| Template directory not found | Ensure `templates/` directory exists in repository       |
| Permission denied            | Check write permissions for destination directory        |
| Files already exist          | Templates will overwrite existing files without warning; backup first |

## Next Steps

- [Build Documentation Site](./build-documentation-site.md) → Generate static site from docs
- [create design](../../../../reference/eac/commands/create/design.md) → Generate architecture diagrams

## Related Commands

- [`templates install`](../../../../reference/eac/commands/templates/index.md) - Install templates overview
- [`templates install-docs`](../../../../reference/eac/commands/templates/install-docs.md) - Documentation templates
- [`templates install-ai`](../../../../reference/eac/commands/templates/install-ai.md) - AI prompt templates
- [`templates install-reports`](../../../../reference/eac/commands/templates/install-reports.md) - Report templates
- [`templates install-specs`](../../../../reference/eac/commands/templates/install-specs.md) - Specification templates
- [`templates install-claude`](../../../../reference/eac/commands/templates/install-claude.md) - Claude Code templates
