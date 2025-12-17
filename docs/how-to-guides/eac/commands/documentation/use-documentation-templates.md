# Use Documentation Templates
## What You'll Accomplish

Install documentation templates to bootstrap consistent project documentation structure.

## Prerequisites

- Repository initialized with EAC
- Template files available in `templates/` directory

## Steps

### 1. Install Documentation Templates

```bash
r2r templates install docs
```

**What happens**: Installs documentation templates to `docs/reference/`

### 2. Install to Custom Location (Optional)

```bash
r2r templates install docs --destination ./custom-docs
```

**What happens**: Installs templates to your specified directory

### 3. Customize Templates

Edit the installed template files to replace placeholders with your specific values:

```bash
# Edit installed templates
code docs/reference/
```

**What happens**: You manually customize templates for your project

### 4. Verify Installed Templates

```bash
ls docs/reference/
```

**What happens**: See all installed documentation templates

## Template Types

Available templates for installation:

- **docs** - Documentation templates (README, guides, references)
- **ai** - AI prompt templates (for code generation commands)
- **reports** - Report templates (test reports, build summaries)
- **specs** - Specification templates (Gherkin scenarios for compliance)

## Example Scenario

Setting up documentation structure for a new project:

```bash
# Install documentation templates
r2r templates install docs
# ✓ Installed to docs/reference/
# ✓ Files: README.md, architecture.md, operations/...

# Install AI prompt templates
r2r templates install ai
# ✓ Installed to .r2r/eac/templates/ai/

# Install report templates
r2r templates install reports
# ✓ Installed to .r2r/templates/reports/

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
r2r templates install docs

# Install AI templates
r2r templates install ai

# Install report templates with debug logging
r2r templates install reports --debug

# Install specification templates
r2r templates install specs
```

## Template Destinations

| Template Type | Default Destination | Custom Path Support |
|---------------|---------------------|---------------------|
| docs | `docs/reference/` | Yes (`--destination`) |
| ai | `.r2r/eac/templates/ai/` | No |
| reports | `.r2r/templates/reports/` | No |
| specs | `specs/risk-controls/` | No |

## Common Issues

| Problem | Solution |
|---------|----------|
| Template directory not found | Ensure `templates/` directory exists in repository |
| Permission denied | Check write permissions for destination directory |
| Files already exist | Templates won't overwrite; delete or move existing files |

## Next Steps

- [Build Documentation Site](./build-documentation-site.md) → Generate static site from docs
- [create design](../../../../reference/commands/create/design.md) → Generate architecture diagrams

## Related Commands

- [`templates install`](../../../../reference/commands/templates/index.md) - Install templates overview
- [`templates install-docs`](../../../../reference/commands/templates/install-docs.md) - Documentation templates
- [`templates install-ai`](../../../../reference/commands/templates/install-ai.md) - AI prompt templates
- [`templates install-reports`](../../../../reference/commands/templates/install-reports.md) - Report templates
