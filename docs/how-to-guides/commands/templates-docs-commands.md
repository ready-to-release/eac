# Templates and Docs Commands

**Problem**: Setting up documentation and maintaining consistent project structure requires manual file creation and configuration.

**Solution**: Use `templates` to install project templates and `docs` to serve documentation sites.

## Key Benefits

- Quick project setup with templates
- Consistent documentation structure
- Live documentation preview
- MkDocs integration
- Customizable templates for specs and docs

## Quick Start

```bash
# Install documentation templates
r2r eac templates install-docs

# Install specification templates
r2r eac templates install-specs

# Serve documentation site
r2r eac docs serve

# Apply templates with variable replacement
r2r eac templates apply-docs --replace "project=MyProject"
```

## Templates Command

### templates install

Install templates to local directory.

```bash
r2r eac templates install [options]

# Subcommands:
r2r eac templates install-docs          # Install doc templates
r2r eac templates install-specs         # Install spec templates
r2r eac templates install-reports       # Install report templates
```

**What it installs:**

**Documentation templates** (`templates/docs/`):
- Module documentation structure
- How-to guides templates
- Reference documentation templates
- Tutorial templates

**Specification templates** (`templates/specs/`):
- Gherkin feature templates
- Scenario templates
- Step definition templates

**Report templates** (`templates/reports/`):
- Test report templates
- Build report templates
- Validation report templates

### templates list

List placeholder variables in templates.

```bash
r2r eac templates list

# Output:
# Available template variables:
# - {{PROJECT_NAME}}
# - {{MODULE_NAME}}
# - {{AUTHOR}}
# - {{VERSION}}
# - {{DESCRIPTION}}
```

### templates apply

Apply templates with variable replacement.

```bash
r2r eac templates apply-docs [options]

# Options:
--replace "key=value"      # Replace {{KEY}} with value

# Examples:
r2r eac templates apply-docs --replace "project=MyApp" --replace "author=John Doe"
```

## Docs Command

### docs serve

Start MkDocs documentation server.

```bash
r2r eac docs serve [options]

# Options:
start                  # Start documentation server (default)
stop                   # Stop documentation server

# Examples:
r2r eac docs serve start
r2r eac docs serve stop
r2r eac docs serve      # Defaults to start
```

**What it does:**
1. Starts MkDocs dev server
2. Watches for file changes
3. Auto-rebuilds on changes
4. Serves on http://localhost:8000
5. Provides live reload

## Typical Workflows

### Initial Project Setup

```bash
# 1. Install all templates
r2r eac templates install-docs
r2r eac templates install-specs
r2r eac templates install-reports

# 2. Apply with project variables
r2r eac templates apply-docs \
  --replace "project=MyProject" \
  --replace "author=TeamName" \
  --replace "version=1.0.0"

# 3. Serve documentation
r2r eac docs serve

# Opens http://localhost:8000
```

### Documentation Development

```bash
# Start docs server
r2r eac docs serve

# Edit docs in docs/
nano docs/how-to-guides/my-guide.md

# Browser auto-refreshes with changes

# When done
r2r eac docs serve stop
```

### Adding New Module Documentation

```bash
# Create module docs from template
cp templates/docs/module-template.md docs/modules/my-module.md

# Edit with module details
nano docs/modules/my-module.md

# Preview
r2r eac docs serve
```

### Specification Workflow

```bash
# Install spec templates
r2r eac templates install-specs

# Create spec from template
cp templates/specs/feature-template.feature specs/src-auth/login.feature

# Edit specification
nano specs/src-auth/login.feature

# Validate
r2r eac specs validate
```

## Documentation Structure

### MkDocs Site

```
docs/
├── index.md                          # Home page
├── getting-started/
│   ├── installation.md
│   └── quick-start.md
├── how-to-guides/
│   ├── commands/
│   │   ├── work-command.md
│   │   ├── specs-command.md
│   │   └── design-command.md
│   └── workflows/
│       └── tdd-workflow.md
├── reference/
│   ├── modules/
│   └── contracts/
└── tutorials/
    └── first-feature.md

mkdocs.yml                            # MkDocs configuration
```

### Template Variables

Common template placeholders:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{PROJECT_NAME}}` | Project name | "MyApp" |
| `{{MODULE_NAME}}` | Module name | "src-auth" |
| `{{AUTHOR}}` | Author name | "John Doe" |
| `{{VERSION}}` | Version | "1.0.0" |
| `{{DESCRIPTION}}` | Description | "Authentication module" |
| `{{DATE}}` | Current date | "2025-01-21" |

## Integration Patterns

### Documentation in CI/CD

```yaml
# Build and deploy docs
- name: Build Documentation
  run: |
    pip install mkdocs mkdocs-material
    mkdocs build

- name: Deploy to GitHub Pages
  uses: peaceiris/actions-gh-pages@v3
  with:
    github_token: ${{ secrets.GITHUB_TOKEN }}
    publish_dir: ./site
```

### Pre-commit Documentation Check

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check if docs changed
if git diff --cached --name-only | grep -q '^docs/'; then
  echo "Building documentation..."
  mkdocs build --strict || exit 1
fi
```

### Template Automation

```bash
# Script to create new module with docs
create_module() {
  MODULE=$1

  # Create module structure
  mkdir -p src/$MODULE

  # Apply doc template
  r2r eac templates apply-docs \
    --replace "module=$MODULE" \
    --replace "date=$(date +%Y-%m-%d)"

  # Create spec from template
  cp templates/specs/module-spec.feature specs/$MODULE/
}

create_module "src-payments"
```

## MkDocs Configuration

### Basic mkdocs.yml

```yaml
site_name: My Project
site_description: Project documentation

theme:
  name: material
  palette:
    primary: indigo
    accent: indigo

nav:
  - Home: index.md
  - Getting Started:
      - Installation: getting-started/installation.md
      - Quick Start: getting-started/quick-start.md
  - How-To Guides:
      - Commands: how-to-guides/commands/
  - Reference:
      - Modules: reference/modules/
  - Tutorials:
      - First Feature: tutorials/first-feature.md

plugins:
  - search
  - awesome-pages

markdown_extensions:
  - pymdownx.highlight
  - pymdownx.superfences
  - admonition
  - toc:
      permalink: true
```

## Best Practices

- **Use templates**: Don't create docs from scratch
- **Consistent structure**: Follow established patterns
- **Live preview**: Use `docs serve` while writing
- **Version control**: Commit templates and generated docs
- **Update templates**: Keep templates current with best practices
- **Variable replacement**: Use templates for repeated structures
- **Link checking**: Validate internal links regularly

## Troubleshooting

| Problem | Solution |
|---------|----------|
| MkDocs not found | Install: `pip install mkdocs mkdocs-material` |
| Port 8000 in use | Stop existing server or use different port |
| Template not found | Run `templates install` first |
| Variables not replaced | Check `--replace` syntax: `key=value` |
| Docs not updating | Check file watch, restart server |
| Build fails | Check `mkdocs.yml` syntax |

## Advanced Usage

### Custom Templates

```bash
# Create custom template
mkdir -p templates/custom/
cat > templates/custom/my-template.md << 'EOF'
# {{TITLE}}

By {{AUTHOR}} - {{DATE}}

{{CONTENT}}
EOF

# Use custom template
cp templates/custom/my-template.md docs/custom-page.md
```

### Batch Template Application

```bash
# Apply to multiple files
for file in docs/**/*.md; do
  sed -i "s/{{PROJECT_NAME}}/MyProject/g" $file
  sed -i "s/{{VERSION}}/1.0.0/g" $file
done
```

### Documentation Validation

```bash
# Build with strict mode (fails on warnings)
mkdocs build --strict

# Check links
pip install linkchecker
mkdocs serve &
linkchecker http://localhost:8000
```

## Summary

**Templates:**
1. `r2r eac templates install-docs` - Install documentation templates
2. `r2r eac templates install-specs` - Install specification templates
3. `r2r eac templates apply-docs --replace "key=value"` - Apply with variables
4. `r2r eac templates list` - List available variables

**Documentation:**
1. `r2r eac docs serve` - Start documentation server
2. Edit files in `docs/`
3. Preview at http://localhost:8000
4. `r2r eac docs serve stop` - Stop server

Templates and docs commands streamline documentation creation and maintenance.
