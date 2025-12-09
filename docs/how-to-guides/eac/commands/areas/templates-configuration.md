# Templates Configuration

{{ page_breadcrumb() }}

This guide covers configuration options for EAC's template system, including placeholder definitions, template locations, and value sources.

## Configuration Files

| File                            | Purpose           |
| ------------------------------- | ----------------- |
| `.r2r/eac/templates/config.yml` | Template settings |
| `.r2r/templates/`               | Template files    |
| `.r2r/eac/templates/values.yml` | Custom values     |

## Template Settings

### Basic Configuration

`.r2r/eac/templates/config.yml`:

```yaml
# Template directories
paths:
  # Documentation templates
  docs: .r2r/templates/docs/

  # Report templates
  reports: .r2r/templates/reports/

  # Project templates
  project: .r2r/templates/project/

  # Output directories
  output:
    docs: docs/
    reports: out/reports/

# Template engine settings
engine:
  # Delimiter style
  left_delim: "{{"
  right_delim: "}}"

  # Missing key behavior: error, zero, invalid
  missing_key: error

  # HTML escaping
  html_escape: false

# File handling
files:
  # Template file extension
  extension: .tmpl

  # Preserve original extension
  preserve_extension: true

  # Overwrite existing files
  overwrite: true
```

## Template Locations

### Directory Structure

```text
.r2r/templates/
├── docs/
│   ├── README.md.tmpl
│   ├── CONTRIBUTING.md.tmpl
│   ├── module-readme.md.tmpl
│   └── api-reference.md.tmpl
├── reports/
│   ├── compliance-report.md.tmpl
│   ├── security-report.md.tmpl
│   ├── test-report.md.tmpl
│   └── release-notes.md.tmpl
└── project/
    ├── Makefile.tmpl
    ├── .gitignore.tmpl
    └── docker-compose.yml.tmpl
```

### Template Resolution Order

1. Project templates (`.r2r/templates/`)
2. Module-specific templates (`<module>/.templates/`)
3. Built-in templates (EAC defaults)

## Placeholder Configuration

### Built-in Placeholders

```yaml
# Automatically available placeholders
builtins:
  # Project information
  - .Project.Name
  - .Project.Description
  - .Project.Repo
  - .Project.RepoURL
  - .Project.ImportPath
  - .Project.License

  # Runtime information
  - .Date           # YYYY-MM-DD
  - .DateTime       # YYYY-MM-DD HH:MM:SS
  - .Timestamp      # Unix timestamp
  - .Year           # YYYY
  - .Version        # Current version
  - .GitCommit      # Short commit hash
  - .GitBranch      # Current branch
  - .GitTag         # Latest tag

  # Module information (when in module context)
  - .Module.Moniker
  - .Module.Type
  - .Module.Description
  - .Module.Path
  - .Module.ImportPath
```

### Custom Values

`.r2r/eac/templates/values.yml`:

```yaml
# Custom placeholder values
values:
  # Organization information
  Organization:
    Name: "My Company"
    URL: "https://example.com"
    Email: "contact@example.com"

  # Documentation settings
  Docs:
    URL: "https://docs.example.com"
    GettingStarted: "/getting-started"
    APIReference: "/api"

  # Support information
  Support:
    Email: "support@example.com"
    Slack: "#support"
    Issues: "https://github.com/org/repo/issues"

  # Custom fields
  Custom:
    Field1: "value1"
    Field2: "value2"
```

### Using Custom Values

```markdown
# {{.Project.Name}}

Maintained by {{.Values.Organization.Name}}

## Support

- Email: {{.Values.Support.Email}}
- Issues: {{.Values.Support.Issues}}
- Slack: {{.Values.Support.Slack}}

## Documentation

Visit [our docs]({{.Values.Docs.URL}}) for more information.
```

## Template Syntax

### Variables

```markdown
# Simple variable
{{.Project.Name}}

# Nested variable
{{.Module.Description}}

# Custom value
{{.Values.Organization.Name}}
```

### Conditionals

```markdown
{{if .Module.HasTests}}
## Testing

Run tests with:
\```bash
r2r eac test {{.Module.Moniker}}
\```

{{end}}

{{if not .Module.IsDeprecated}}

## Installation

\```bash
go get {{.Module.ImportPath}}
\```

{{end}}

### Loops

\```markdown
## Modules

{{range .Modules}}
### {{.Moniker}}

{{.Description}}

- Type: {{.Type}}
- Path: {{.Path}}
{{end}}
\```

### With Context

\```markdown
{{with .Module}}
# {{.Moniker}}

{{.Description}}

Type: {{.Type}}
{{end}}
\```
```

### Pipelines

```markdown
# Last updated: {{.Date | upper}}

# Module: {{.Module.Moniker | title}}

# Description: {{.Module.Description | default "No description"}}
```

### Built-in Functions

| Function  | Description     | Example                        |
| --------- | --------------- | ------------------------------ |
| `upper`   | Uppercase       | `{{.Name \| upper}}`           |
| `lower`   | Lowercase       | `{{.Name \| lower}}`           |
| `title`   | Title case      | `{{.Name \| title}}`           |
| `trim`    | Trim whitespace | `{{.Text \| trim}}`            |
| `default` | Default value   | `{{.Desc \| default "N/A"}}`   |
| `join`    | Join slice      | `{{.Tags \| join ", "}}`       |
| `split`   | Split string    | `{{.Path \| split "/"}}`       |
| `replace` | Replace string  | `{{.Name \| replace "-" "_"}}` |

## Output Configuration

### Documentation Output

```yaml
output:
  docs:
    # Output directory
    path: docs/

    # File naming
    naming:
      # Use template name
      preserve_name: true
      # Or use module name
      # pattern: "{module}-{template}"

    # Post-processing
    post_process:
      # Format markdown
      format_markdown: true
      # Add frontmatter
      add_frontmatter: true
```

### Report Output

```yaml
output:
  reports:
    path: out/reports/

    # Date-based subdirectories
    date_subdirs: true
    # Creates: out/reports/2024-12-01/

    # Archive old reports
    archive:
      enabled: true
      keep_days: 30
```

## Apply Modes

### Apply Documentation Templates

```yaml
apply:
  docs:
    # Templates to apply
    templates:
      - README.md.tmpl
      - CONTRIBUTING.md.tmpl

    # Apply to each module
    per_module:
      enabled: true
      templates:
        - module-readme.md.tmpl
      output_pattern: "{module}/README.md"
```

### Apply Report Templates

```yaml
apply:
  reports:
    templates:
      - compliance-report.md.tmpl
      - security-report.md.tmpl

    # Include runtime data
    include_data:
      - security_scan_results
      - test_results
      - compliance_status
```

### Install Without Values

```yaml
install:
  # Copy templates without processing
  docs:
    source: .r2r/templates/docs/
    destination: docs/templates/

  reports:
    source: .r2r/templates/reports/
    destination: out/templates/
```

## Tag Extraction

### Tag Configuration

```yaml
tags:
  # Tag pattern
  pattern: "{{/* @tag: (.*) */}}"

  # Extract tags from templates
  extract:
    enabled: true
    output: out/template-tags.yml
```

### Using Tags

```markdown
{{/* @tag: project-info */}}
# {{.Project.Name}}

{{/* @tag: module-list */}}
## Modules
{{range .Modules}}
- {{.Moniker}}
{{end}}

{{/* @tag: support-info */}}
## Support
Contact: {{.Values.Support.Email}}
```

### Extracted Tags

```yaml
# out/template-tags.yml
tags:
  - name: project-info
    file: README.md.tmpl
    line: 1

  - name: module-list
    file: README.md.tmpl
    line: 4

  - name: support-info
    file: README.md.tmpl
    line: 10
```

## MkDocs Integration

### MkDocs Templates

```yaml
mkdocs:
  # Generate navigation from templates
  generate_nav: true

  # Template for nav structure
  nav_template: .r2r/templates/mkdocs/nav.yml.tmpl

  # Auto-include generated docs
  auto_include: true
```

### Navigation Template

```yaml
# nav.yml.tmpl
nav:
  - Home: index.md
  - Getting Started: getting-started.md
  - Modules:
    {{range .Modules}}
    - {{.Moniker}}: modules/{{.Moniker}}.md
    {{end}}
  - API Reference: api/index.md
```

## Environment Variables

| Variable          | Description        | Default                         |
| ----------------- | ------------------ | ------------------------------- |
| `TEMPLATE_PATH`   | Template directory | `.r2r/templates/`               |
| `TEMPLATE_OUTPUT` | Output directory   | `docs/`                         |
| `TEMPLATE_VALUES` | Custom values file | `.r2r/eac/templates/values.yml` |

## Example Configurations

### Minimal Configuration

```yaml
paths:
  docs: .r2r/templates/docs/
  output:
    docs: docs/

engine:
  missing_key: error
```

### Documentation Project

```yaml
paths:
  docs: .r2r/templates/docs/
  reports: .r2r/templates/reports/
  output:
    docs: docs/
    reports: out/reports/

engine:
  missing_key: error
  html_escape: false

apply:
  docs:
    templates:
      - README.md.tmpl
      - CONTRIBUTING.md.tmpl
      - CODE_OF_CONDUCT.md.tmpl
    per_module:
      enabled: true
      templates:
        - module-readme.md.tmpl
      output_pattern: "modules/{module}/README.md"

mkdocs:
  generate_nav: true
  auto_include: true
```

### Enterprise Configuration

```yaml
paths:
  docs: .r2r/templates/docs/
  reports: .r2r/templates/reports/
  project: .r2r/templates/project/
  output:
    docs: docs/
    reports: out/reports/

engine:
  missing_key: error
  html_escape: false

files:
  extension: .tmpl
  overwrite: false  # Don't overwrite manual changes

apply:
  docs:
    templates:
      - README.md.tmpl
      - CONTRIBUTING.md.tmpl
      - SECURITY.md.tmpl
      - CHANGELOG.md.tmpl

  reports:
    templates:
      - compliance-report.md.tmpl
      - security-report.md.tmpl
      - audit-report.md.tmpl
    include_data:
      - security_scan_results
      - compliance_status
      - audit_evidence

output:
  reports:
    date_subdirs: true
    archive:
      enabled: true
      keep_days: 90

tags:
  extract:
    enabled: true
    output: out/template-tags.yml
```

## Troubleshooting

| Issue                    | Cause                       | Solution             |
| ------------------------ | --------------------------- | -------------------- |
| Placeholder not replaced | Wrong path or missing value | Check `.Values` path |
| Template syntax error    | Invalid Go template         | Validate syntax      |
| File not generated       | Output path wrong           | Check output config  |
| Missing data             | Data source not loaded      | Verify data sources  |
| Encoding issues          | Non-UTF8 content            | Use UTF-8 encoding   |

## Related Documentation

- [Templates Overview](templates-overview.md) - Concepts and workflows
- [Templates Commands](templates-commands.md) - Command reference
- [Books Configuration](books-configuration.md) - Book generation

{{ diataxis_footer() }}
