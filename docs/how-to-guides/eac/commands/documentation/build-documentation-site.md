# Build Documentation Site

## What You'll Accomplish

Generate and preview documentation site using MkDocs with validation.

## Prerequisites

- MkDocs site module configured
- Documentation markdown files exist
- Books configuration (books.yml)

## Steps

### 1. Validate Documentation

```bash
r2r eac validate markdown
r2r eac validate books
```

**What happens**: Checks markdown syntax and books configuration

### 2. Build Documentation

```bash
r2r eac build docs-site
```

**What happens**: Generates static HTML site from markdown

### 3. Preview Documentation

```bash
r2r eac serve docs
```

**What happens**: Starts MkDocs development server with live reload

### 4. Open in Browser

Navigate to <http://localhost:8000>

**What happens**: View documentation site with navigation

## Books Configuration

Books.yml defines documentation structure:

```yaml
books:
  - name: "User Guide"
    nav: docs/user-guide/.nav.yml

  - name: "Reference"
    nav: docs/reference/.nav.yml
```

## Example Scenario

Building and previewing docs:

```bash
# Validate first
r2r eac validate markdown
# ✓ All markdown files valid

r2r eac validate books
# ✓ books.yml configuration valid

# Build site
r2r eac build docs-site
# Building documentation...
# ✓ Generated site/

# Preview locally
r2r eac serve docs
# Starting MkDocs server on http://localhost:8000
# INFO - Building documentation...
# INFO - Documentation built in 2.3s

# Open browser to http://localhost:8000
# Make changes to markdown, auto-reloads

# Stop server
# Ctrl+C

# Build for production
r2r eac build docs-site --clean
# ✓ Production site built to site/
```

## Validation Checks

```bash
# Check markdown syntax
r2r eac validate markdown

# Validate books config
r2r eac validate books

# Validate navigation structure
find docs -name ".nav.yml" -exec cat {} \;
```

## Common Issues

| Problem            | Solution                             |
| ------------------ | ------------------------------------ |
| "MkDocs not found" | Install MkDocs: `pip install mkdocs` |
| Broken links       | Check file paths in .nav.yml files   |
| Missing pages      | Ensure all referenced files exist    |

## Next Steps

- [Use Documentation Templates](./use-documentation-templates.md) → Consistent docs

## Related Commands

- [`serve docs`](../../../../reference/commands/serve/docs.md) - Start dev server
- [`validate markdown`](../../../../reference/commands/validate/markdown.md) - Check syntax
- [`validate books`](../../../../reference/commands/validate/books.md) - Validate config
