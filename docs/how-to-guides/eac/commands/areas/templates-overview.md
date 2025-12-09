# Templates

{{ page_breadcrumb() }}

Templates in EAC provide a placeholder-based system for generating consistent documentation, reports, and project files with automatic value substitution.

## What are Templates?

EAC's template system enables you to:

- **Generate documentation** from templates with live data
- **Create standardized reports** with consistent formatting
- **Bootstrap project files** with proper structure
- **Maintain consistency** across documentation artifacts

The system uses placeholder syntax that gets replaced with values from contracts, configuration, and runtime data.

## When to Use Templates

Use template commands when you need:

| Scenario                               | Commands                                         |
| -------------------------------------- | ------------------------------------------------ |
| Apply templates with values            | `templates apply`, `templates apply docs`        |
| Install templates without substitution | `templates install`, `templates install reports` |
| List available placeholders            | `templates list`                                 |
| Extract tags from templates            | `templates tags`                                 |

### Common Use Cases

- **Documentation generation** - README files, guides, API docs
- **Report generation** - Compliance reports, status reports
- **Project bootstrapping** - Initial project structure
- **Standardization** - Consistent formats across modules

## Key Concepts

### Placeholder Syntax

Templates use placeholder syntax:

```markdown
# {{.Project.Name}}

Version: {{.Version}}
Last Updated: {{.Date}}

## Modules

{{range .Modules}}
- **{{.Moniker}}**: {{.Description}}
{{end}}
```

### Placeholder Categories

| Category        | Examples                                | Source                |
| --------------- | --------------------------------------- | --------------------- |
| **Project**     | `.Project.Name`, `.Project.Description` | Repository config     |
| **Module**      | `.Module.Moniker`, `.Module.Type`       | Module contracts      |
| **Version**     | `.Version`, `.Date`                     | Runtime/release info  |
| **Environment** | `.Env.Name`, `.Env.URL`                 | Environment contracts |
| **Custom**      | `.Custom.Key`                           | User-defined values   |

### Template Types

| Type              | Purpose            | Location                  |
| ----------------- | ------------------ | ------------------------- |
| **Documentation** | README, guides     | `.r2r/templates/docs/`    |
| **Reports**       | Compliance, status | `.r2r/templates/reports/` |
| **Project**       | Config files       | `.r2r/templates/project/` |

### Template Resolution

Templates are resolved from multiple locations:

1. Project templates (`.r2r/templates/`)
2. Default templates (built-in)
3. Custom template paths

## Workflow Overview

### Documentation Generation

```bash
# 1. List available templates
r2r eac templates list

# 2. Preview placeholder values
r2r eac templates list --show-values

# 3. Apply documentation templates
r2r eac templates apply docs

# 4. Review generated files
cat docs/README.md
```

### Report Generation

```bash
# 1. Install report templates (first time)
r2r eac templates install reports

# 2. Apply templates with current data
r2r eac templates apply --type reports

# 3. Review reports
cat out/reports/compliance-report.md
```

### Custom Templates

```bash
# 1. Create custom template
cat > .r2r/templates/docs/module-readme.md << 'EOF'
# {{.Module.Moniker}}

{{.Module.Description}}

## Installation

go get {{.Module.ImportPath}}

## Usage

See documentation at {{.DocsURL}}.
EOF

# 2. Apply to specific module
r2r eac templates apply --template module-readme.md --module eac-core
```

## Template Structure

### Documentation Templates

```

.r2r/templates/docs/
├── README.md.tmpl # Project README
├── module-readme.md.tmpl # Per-module README
├── contributing.md.tmpl # Contributing guide
└── architecture.md.tmpl # Architecture overview

```

### Report Templates

```text
.r2r/templates/reports/
├── compliance-report.md.tmpl # Compliance summary
├── security-report.md.tmpl # Security scan report
├── test-report.md.tmpl # Test results report
└── release-notes.md.tmpl # Release notes
```

### Example Template

```text
# {{.Project.Name}}

> {{.Project.Description}}

## Quick Start

  # Install
  go install {{.Project.ImportPath}}@latest

  # Verify
  {{.Project.CLI}} --version

## Modules

| Module | Type | Description |
| ------ | ---- | ----------- |
{{range .Modules -}}
| {{.Moniker}} | {{.Type}} | {{.Description}} |
{{end}}

## Documentation

- User Guide: {{.DocsURL}}/guide
- API Reference: {{.DocsURL}}/api
- Contributing: {{.RepoURL}}/CONTRIBUTING.md

---

Generated on {{.Date}} by {{.Project.Name}}
```

## Integration Points

### With Module Contracts

Templates access contract data:

```yaml
# modules.yml
modules:
  - moniker: eac-commands
    description: CLI command implementations
    # Accessible as {{.Module.Description}}
```

### With MkDocs

Generate pages for documentation site:

```bash
# Generate docs
r2r eac templates apply docs

# Build site
r2r eac build docs

# Serve locally
r2r eac serve docs
```

### With CI/CD

Automate documentation updates:

```yaml
- name: Update documentation
  run: |
    r2r eac templates apply docs
    git add docs/
    git diff --cached --quiet || git commit -m "docs: update from templates"
```

### With Books

Templates feed into book generation:

```bash
# Apply templates first
r2r eac templates apply docs

# Then build book
r2r eac build docs-book
```

## Placeholder Reference

### Project Placeholders

| Placeholder                | Description         |
| -------------------------- | ------------------- |
| `{{.Project.Name}}`        | Project name        |
| `{{.Project.Description}}` | Project description |
| `{{.Project.Repo}}`        | Repository name     |
| `{{.Project.RepoURL}}`     | Full repository URL |
| `{{.Project.ImportPath}}`  | Go import path      |

### Module Placeholders

| Placeholder               | Description                |
| ------------------------- | -------------------------- |
| `{{.Module.Moniker}}`     | Module identifier          |
| `{{.Module.Type}}`        | Module type (go-cli, etc.) |
| `{{.Module.Description}}` | Module description         |
| `{{.Module.Path}}`        | Relative path              |
| `{{.Module.ImportPath}}`  | Go import path             |

### Runtime Placeholders

| Placeholder      | Description           |
| ---------------- | --------------------- |
| `{{.Date}}`      | Current date          |
| `{{.DateTime}}`  | Current date and time |
| `{{.Version}}`   | Current version       |
| `{{.GitCommit}}` | Current git commit    |
| `{{.GitBranch}}` | Current git branch    |

### Control Structures

```markdown
{{/* Conditionals */}}
{{if .Module.HasTests}}
## Testing
Run tests with: `r2r eac test {{.Module.Moniker}}`
{{end}}

{{/* Loops */}}
{{range .Modules}}
- {{.Moniker}}
{{end}}

{{/* With context */}}
{{with .Module}}
Module: {{.Moniker}} ({{.Type}})
{{end}}
```

## Best Practices

### Template Design

- **Keep it simple** - Complex logic belongs in code, not templates
- **Use consistent naming** - Placeholder names should be intuitive
- **Document placeholders** - Help users understand available values
- **Test templates** - Verify output with real data

### Do's

```markdown
{{/* Good: Clear, simple placeholders */}}
# {{.Project.Name}}
Version: {{.Version}}
```

### Don'ts

```markdown
{{/* Bad: Complex logic in templates */}}
{{if and (eq .Module.Type "go-cli") (gt (len .Module.Files) 10)}}
...complex nested logic...
{{end}}
```

## Troubleshooting

| Problem                  | Solution                                      |
| ------------------------ | --------------------------------------------- |
| Placeholder not replaced | Check spelling, use `templates list`          |
| Missing value            | Ensure data source exists (contracts, config) |
| Template syntax error    | Validate Go template syntax                   |
| Empty output             | Check conditionals, verify data exists        |

## Next Steps

- [Templates Configuration](templates-configuration.md) - Configure template paths and custom values
- [Templates Commands](templates-commands.md) - Full command reference

## Related Areas

- [Books](books-overview.md) - Aggregate templated docs into books
- [Design](design-overview.md) - Architecture diagrams in documentation
- [Release](release-overview.md) - Release notes generation

{{ diataxis_footer() }}
