# Use Documentation Templates

{{ page_breadcrumb() }}

## What You'll Accomplish

Apply documentation templates for consistent project documentation with variable substitution.

## Prerequisites

- Templates available (or install them)
- Template variables defined

## Steps

### 1. Install Templates

```bash
r2r eac templates install
```

**What happens**: Installs templates to project without variable substitution

### 2. List Template Variables

```bash
r2r eac templates list
```

**What happens**: Shows available placeholder variables

### 3. Apply Templates with Values

```bash
r2r eac templates apply --module src-auth
```

**What happens**: Applies templates with variable substitution for src-auth module

### 4. Verify Generated Docs

```bash
ls docs/
```

**What happens**: Templates are applied with values filled in

## Template Types

Available templates:

- **Documentation templates** - README, guides, references
- **Report templates** - Test reports, build summaries
- **Specification templates** - Gherkin scenarios

## Example Scenario

Setting up documentation for new module:

```bash
# Install templates first
r2r eac templates install
# ✓ Installed documentation templates
# ✓ Installed report templates

# List available variables
r2r eac templates list
# Available placeholders:
# - {{MODULE_NAME}}
# - {{MODULE_TYPE}}
# - {{MODULE_PATH}}
# - {{MODULE_DESCRIPTION}}

# Apply templates for module
r2r eac templates apply --module src-auth

# Output:
# Applying templates for src-auth...
# ✓ Generated docs/README.md
# ✓ Generated docs/architecture.md
# ✓ Generated docs/api.md

# Review generated docs
cat docs/README.md
# # src-auth Module
#
# Authentication module for JWT token management...
```

## Custom Templates

```bash
# Install only docs templates
r2r eac templates apply-docs --module src-auth

# Install only report templates
r2r eac templates install-reports
```

## Template Variables

Common variables:

- `{{MODULE_NAME}}` - Module moniker
- `{{MODULE_TYPE}}` - Module type (go-library, etc.)
- `{{MODULE_PATH}}` - Path to module
- `{{DESCRIPTION}}` - Module description
- `{{DATE}}` - Current date

## Common Issues

| Problem | Solution |
|---------|----------|
| Template not found | Run `templates install` first |
| Variables not replaced | Check module exists in contracts |
| Wrong values | Update module.yml metadata |

## Next Steps

- [Build Documentation Site](./build-documentation-site.md) → Generate site

## Related Commands

- [`templates install`](../../../reference/commands/templates/install.md) - Install templates
- [`templates apply`](../../../reference/commands/templates/apply.md) - Apply with substitution
- [`templates list`](../../../reference/commands/templates/list.md) - List variables

{{ diataxis_footer() }}
