# Templates and Serve Docs Commands

**Problem**: Setting up documentation and maintaining consistent project structure requires manual file creation and configuration.

**Solution**: Use `templates` to manage project templates with variable substitution and `serve docs` to serve documentation sites.

## Key Benefits

- Quick project setup with templates
- Consistent documentation structure
- Live documentation preview with MkDocs
- Variable substitution using JSON input files
- Automated report template installation
- Template variable discovery

## Quick Start

```bash
# Apply documentation templates with variables
r2r eac templates apply docs --input-json values.json

# Install report templates
r2r eac templates install reports

# List available template variables
r2r eac templates list --template path/to/template.md

# Serve documentation site
r2r eac serve docs

# Stop documentation server
r2r eac serve docs --stop
```

## Templates Command

The `templates` command provides template management with three main subcommands: `apply docs`, `install reports`, and `list`.

### templates apply docs

Apply documentation templates with variable substitution.

```bash
r2r eac templates apply docs [options]

# Options:
--source <path>              # Source template directory
--destination <path>         # Destination directory (default: .docs/reference)
--input-json <file>          # JSON file containing key-value pairs for substitution
--debug, -d                  # Enable debug output
```

**How it works:**

1. Reads templates from source directory
2. Loads variable values from JSON input file
3. Replaces `{{VARIABLE}}` placeholders with values from JSON
4. Writes processed files to destination directory

**JSON input file format:**

```json
{
  "PROJECT_NAME": "MyProject",
  "MODULE_NAME": "src-auth",
  "AUTHOR": "John Doe",
  "VERSION": "1.0.0",
  "DESCRIPTION": "Authentication module",
  "DATE": "2025-11-30"
}
```

**Examples:**

```bash
# Apply docs templates with values from JSON file
r2r eac templates apply docs --input-json project-values.json

# Specify custom source and destination
r2r eac templates apply docs \
  --source ./custom-templates \
  --destination ./docs/modules \
  --input-json values.json

# Enable debug output
r2r eac templates apply docs --input-json values.json --debug
```

### templates install reports

Install report templates without variable substitution.

```bash
r2r eac templates install reports [options]

# Options:
--source <path>              # Source template directory
--destination <path>         # Destination directory (default: .r2r/templates/reports)
--debug, -d                  # Enable debug output
```

**What it installs:**

Report templates are installed to `.r2r/templates/reports/` by default and include:

- Test report templates
- Build report templates
- Validation report templates
- CI/CD summary templates

**Examples:**

```bash
# Install report templates to default location
r2r eac templates install reports

# Install to custom destination
r2r eac templates install reports --destination ./templates/reports

# Install with debug output
r2r eac templates install reports --debug
```

### templates list

List all placeholder variables found in template files.

```bash
r2r eac templates list [options]

# Options:
--template <path>            # Path to template file or directory
--debug, -d                  # Enable debug output
```

**What it does:**

1. Scans template file(s) for `{{VARIABLE}}` patterns
2. Extracts unique variable names
3. Displays list of all placeholders found

**Output example:**

```
Available template variables:
- {{PROJECT_NAME}}
- {{MODULE_NAME}}
- {{AUTHOR}}
- {{VERSION}}
- {{DESCRIPTION}}
- {{DATE}}
```

**Examples:**

```bash
# List variables in a single template
r2r eac templates list --template ./templates/module-doc.md

# List variables in all templates in a directory
r2r eac templates list --template ./templates/

# List with debug output
r2r eac templates list --template ./templates/ --debug
```

## Serve Docs Command

The `serve docs` command manages the MkDocs documentation server.

### serve docs

Start or stop the MkDocs documentation server.

```bash
r2r eac serve docs [options]

# Options:
--no-browser                 # Don't automatically open browser
--port, -p <number>          # Port number (default: 8000)
--debug                      # Enable debug output
--stop                       # Stop the documentation server
```

**Default behavior (start server):**

1. Starts MkDocs development server
2. Watches for file changes
3. Auto-rebuilds on changes
4. Opens browser to http://localhost:8000
5. Provides live reload

**Examples:**

```bash
# Start server (opens browser automatically)
r2r eac serve docs

# Start without opening browser
r2r eac serve docs --no-browser

# Start on custom port
r2r eac serve docs --port 8080
r2r eac serve docs -p 8080

# Start with debug output
r2r eac serve docs --debug

# Stop the server
r2r eac serve docs --stop
```

## Typical Workflows

### Initial Project Setup

```bash
# 1. Create JSON file with project variables
cat > project-values.json << 'EOF'
{
  "PROJECT_NAME": "MyProject",
  "AUTHOR": "TeamName",
  "VERSION": "1.0.0",
  "DESCRIPTION": "My awesome project",
  "DATE": "2025-11-30"
}
EOF

# 2. Apply documentation templates
r2r eac templates apply docs --input-json project-values.json

# 3. Install report templates
r2r eac templates install reports

# 4. Serve documentation
r2r eac serve docs
```

### Documentation Development

```bash
# Start docs server
r2r eac serve docs

# Edit docs in docs/
nano docs/how-to-guides/my-guide.md

# Browser auto-refreshes with changes

# When done, stop server
r2r eac serve docs --stop
```

### Creating Module Documentation from Template

```bash
# 1. Create module-specific values file
cat > module-values.json << 'EOF'
{
  "MODULE_NAME": "src-payments",
  "DESCRIPTION": "Payment processing module",
  "AUTHOR": "Payments Team",
  "DATE": "2025-11-30"
}
EOF

# 2. Apply template with module values
r2r eac templates apply docs \
  --source ./templates/module-template \
  --destination ./docs/modules \
  --input-json module-values.json

# 3. Preview in documentation
r2r eac serve docs
```

### Discovering Template Variables

```bash
# Check what variables a template needs
r2r eac templates list --template ./templates/new-template.md

# Output shows required variables:
# - {{PROJECT_NAME}}
# - {{MODULE_NAME}}
# - {{FEATURE_NAME}}

# Create JSON file with those values
cat > values.json << 'EOF'
{
  "PROJECT_NAME": "MyApp",
  "MODULE_NAME": "eac-core",
  "FEATURE_NAME": "authentication"
}
EOF

# Apply template
r2r eac templates apply docs --input-json values.json
```

## Documentation Structure

### MkDocs Site

```text
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

| Variable           | Description  | Example                 |
| ------------------ | ------------ | ----------------------- |
| `{{PROJECT_NAME}}` | Project name | "MyApp"                 |
| `{{MODULE_NAME}}`  | Module name  | "src-auth"              |
| `{{AUTHOR}}`       | Author name  | "John Doe"              |
| `{{VERSION}}`      | Version      | "1.0.0"                 |
| `{{DESCRIPTION}}`  | Description  | "Authentication module" |
| `{{DATE}}`         | Current date | "2025-11-30"            |
| `{{FEATURE_NAME}}` | Feature name | "login"                 |

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

### Template Automation Script

```bash
#!/bin/bash
# Script to create new module with docs

create_module() {
  MODULE=$1
  DESCRIPTION=$2

  # Create module structure
  mkdir -p src/$MODULE

  # Create values file
  cat > /tmp/${MODULE}-values.json << EOF
{
  "MODULE_NAME": "$MODULE",
  "DESCRIPTION": "$DESCRIPTION",
  "AUTHOR": "$(git config user.name)",
  "DATE": "$(date +%Y-%m-%d)",
  "VERSION": "0.1.0"
}
EOF

  # Apply doc template
  r2r eac templates apply docs \
    --source ./templates/module \
    --destination ./docs/modules \
    --input-json /tmp/${MODULE}-values.json

  echo "Module $MODULE documentation created"
}

create_module "src-payments" "Payment processing module"
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

- **Use JSON files**: Store template values in version-controlled JSON files
- **Discover variables first**: Use `templates list` to find required variables before applying
- **Consistent structure**: Follow established documentation patterns
- **Live preview**: Use `serve docs` while writing documentation
- **Version control templates**: Keep templates and JSON value files in git
- **Port management**: Use `--port` flag if default port 8000 is occupied
- **Debug mode**: Enable `--debug` flag when troubleshooting template issues
- **Custom destinations**: Use `--destination` to organize generated docs

## Troubleshooting

| Problem                | Solution                                                  |
| ---------------------- | --------------------------------------------------------- |
| MkDocs not found       | Install: `pip install mkdocs mkdocs-material`             |
| Port 8000 in use       | Use `--port` flag: `serve docs --port 8080`               |
| Template not found     | Check `--source` path is correct                          |
| Variables not replaced | Verify JSON file format and use `--input-json` flag       |
| Missing variables      | Use `templates list` to discover required variables       |
| Docs not updating      | Restart server with `serve docs --stop` then `serve docs` |
| Build fails            | Check `mkdocs.yml` syntax with `mkdocs build --strict`    |
| Server won't stop      | Use `--stop` flag: `serve docs --stop`                    |

## Advanced Usage

### Custom Template with Variable Discovery

```bash
# 1. Create custom template
mkdir -p templates/custom/
cat > templates/custom/api-doc.md << 'EOF'
# {{API_NAME}} API Documentation

**Version**: {{VERSION}}
**Author**: {{AUTHOR}}
**Last Updated**: {{DATE}}

## Overview

{{DESCRIPTION}}

## Endpoints

{{ENDPOINTS}}

## Authentication

{{AUTH_METHOD}}
EOF

# 2. Discover what variables are needed
r2r eac templates list --template templates/custom/api-doc.md

# Output:
# - {{API_NAME}}
# - {{VERSION}}
# - {{AUTHOR}}
# - {{DATE}}
# - {{DESCRIPTION}}
# - {{ENDPOINTS}}
# - {{AUTH_METHOD}}

# 3. Create corresponding JSON file
cat > api-values.json << 'EOF'
{
  "API_NAME": "User Service",
  "VERSION": "2.0.0",
  "AUTHOR": "API Team",
  "DATE": "2025-11-30",
  "DESCRIPTION": "User management API",
  "ENDPOINTS": "See endpoints section below",
  "AUTH_METHOD": "Bearer token (JWT)"
}
EOF

# 4. Apply template
r2r eac templates apply docs \
  --source templates/custom \
  --destination docs/api \
  --input-json api-values.json
```

### Multiple Environment Configurations

```bash
# Create environment-specific values
cat > values-dev.json << 'EOF'
{
  "ENVIRONMENT": "Development",
  "API_URL": "https://dev.example.com",
  "DEBUG_MODE": "enabled"
}
EOF

cat > values-prod.json << 'EOF'
{
  "ENVIRONMENT": "Production",
  "API_URL": "https://api.example.com",
  "DEBUG_MODE": "disabled"
}
EOF

# Apply for different environments
r2r eac templates apply docs \
  --source templates/config \
  --destination docs/environments/dev \
  --input-json values-dev.json

r2r eac templates apply docs \
  --source templates/config \
  --destination docs/environments/prod \
  --input-json values-prod.json
```

### Documentation Validation

```bash
# Build with strict mode (fails on warnings)
mkdocs build --strict

# Check links
pip install linkchecker
r2r eac serve docs --no-browser &
sleep 5
linkchecker http://localhost:8000
r2r eac serve docs --stop
```

### Serve on Custom Port for Parallel Development

```bash
# Terminal 1: Main docs on default port
r2r eac serve docs

# Terminal 2: Feature branch docs on different port
cd /path/to/feature-branch
r2r eac serve docs --port 8001
```

## Command Reference

### Templates Commands

| Command                     | Description                                              | Key Flags                                              |
| --------------------------- | -------------------------------------------------------- | ------------------------------------------------------ |
| `templates apply docs`      | Apply documentation templates with variable substitution | `--source`, `--destination`, `--input-json`, `--debug` |
| `templates install reports` | Install report templates without substitution            | `--source`, `--destination`, `--debug`                 |
| `templates list`            | List placeholder variables in templates                  | `--template`, `--debug`                                |

### Serve Commands

| Command      | Description         | Key Flags                                     |
| ------------ | ------------------- | --------------------------------------------- |
| `serve docs` | Start MkDocs server | `--no-browser`, `--port`, `--debug`, `--stop` |

## Summary

**Templates:**

1. `r2r eac templates apply docs --input-json values.json` - Apply templates with JSON values
2. `r2r eac templates install reports` - Install report templates
3. `r2r eac templates list --template <path>` - Discover template variables

**Documentation:**

1. `r2r eac serve docs` - Start documentation server (opens browser)
2. Edit files in `docs/`
3. Preview at http://localhost:8000 (auto-reloads)
4. `r2r eac serve docs --stop` - Stop server

**Key Differences from Legacy Approach:**

- Use `--input-json <file>` instead of multiple `--replace "key=value"` flags
- Use `serve docs --stop` instead of `docs serve stop` subcommand
- Use `templates list` to discover required variables before applying
- Default destinations: `.docs/reference` for docs, `.r2r/templates/reports` for reports

The templates and serve docs commands streamline documentation creation with variable substitution from JSON files and provide integrated MkDocs server management.
