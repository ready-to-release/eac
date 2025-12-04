# Templates Commands

Command reference for EAC's template system.

## Quick Reference

| Command                     | Description                                  |
| --------------------------- | -------------------------------------------- |
| `templates`                 | Manage project templates                     |
| `templates apply`           | Apply templates with value replacements      |
| `templates apply docs`      | Apply documentation templates                |
| `templates install`         | Install templates without value replacements |
| `templates install reports` | Install report templates                     |
| `templates list`            | List template placeholder variables          |
| `templates tags`            | Extract template tags                        |

---

## templates

Manage project templates for documentation and specifications.

### Synopsis

```bash
r2r eac templates <subcommand> [options]
```

### Description

Parent command for template management. Use subcommands for specific operations.

### Subcommands

| Subcommand | Description                               |
| ---------- | ----------------------------------------- |
| `apply`    | Apply templates with value substitution   |
| `install`  | Install templates without substitution    |
| `list`     | List available templates and placeholders |
| `tags`     | Extract placeholder tags from templates   |

### Examples

```bash
# List available templates
r2r eac templates list

# Apply all templates
r2r eac templates apply

# Install templates
r2r eac templates install
```

---

## templates apply

Apply templates with value replacements.

### Synopsis

```bash
r2r eac templates apply [options]
```

### Description

Processes template files and replaces placeholders with values from contracts, configuration, and runtime data. Generates output files with substituted values.

### Flags

| Flag         | Short | Type   | Default | Description                            |
| ------------ | ----- | ------ | ------- | -------------------------------------- |
| `--type`     | `-t`  | string | `all`   | Template type (docs, reports, project) |
| `--template` |       | string | -       | Specific template file to apply        |
| `--module`   | `-m`  | string | -       | Apply to specific module               |
| `--output`   | `-o`  | string | -       | Custom output directory                |
| `--dry-run`  | `-n`  | bool   | `false` | Preview without writing                |

### Examples

```bash
# Apply all templates
r2r eac templates apply

# Apply documentation templates only
r2r eac templates apply --type docs

# Apply report templates
r2r eac templates apply --type reports

# Apply specific template
r2r eac templates apply --template module-readme.md.tmpl

# Apply to specific module
r2r eac templates apply --template module-readme.md.tmpl --module eac-commands

# Preview changes
r2r eac templates apply --dry-run

# Custom output directory
r2r eac templates apply --output out/docs/
```

### Output

```text
Applying templates...

Processing templates:
  ✓ README.md.tmpl → README.md
  ✓ module-readme.md.tmpl → go/eac/commands/README.md
  ✓ architecture.md.tmpl → docs/architecture.md
  ✓ contributing.md.tmpl → CONTRIBUTING.md

Placeholders replaced:
  {{.Project.Name}} → "eac"
  {{.Project.Description}} → "Everything as Code CLI toolkit"
  {{.Version}} → "2025.12.01"
  {{.Modules}} → 8 modules rendered

✓ 4 templates applied
```

### Exit Codes

| Code | Description                    |
| ---- | ------------------------------ |
| 0    | Templates applied successfully |
| 1    | Error applying templates       |
| 2    | Template not found             |
| 3    | Missing required values        |

---

## templates apply docs

Apply documentation templates with value replacements.

### Synopsis

```bash
r2r eac templates apply docs [options]
```

### Description

Specialized command for applying documentation templates. Equivalent to `templates apply --type docs` but with documentation-specific defaults and behavior.

### Flags

| Flag        | Short | Type   | Default | Description                       |
| ----------- | ----- | ------ | ------- | --------------------------------- |
| `--module`  | `-m`  | string | -       | Generate docs for specific module |
| `--output`  | `-o`  | string | `docs/` | Output directory                  |
| `--dry-run` | `-n`  | bool   | `false` | Preview without writing           |

### Examples

```bash
# Apply all documentation templates
r2r eac templates apply docs

# Generate docs for specific module
r2r eac templates apply docs --module eac-commands

# Custom output directory
r2r eac templates apply docs --output generated/docs/

# Preview changes
r2r eac templates apply docs --dry-run
```

### Output

```text
Applying documentation templates...

Generating documentation:
  ✓ README.md - Project overview
  ✓ docs/getting-started.md - Quick start guide
  ✓ docs/api/index.md - API reference index

Module documentation:
  ✓ go/eac/commands/README.md
  ✓ go/eac/core/README.md
  ✓ scripts/r2r-cli/README.md

✓ Documentation generated: 6 files
```

### Exit Codes

| Code | Description                          |
| ---- | ------------------------------------ |
| 0    | Documentation generated successfully |
| 1    | Error generating documentation       |
| 2    | Template not found                   |

---

## templates install

Install templates without value replacements.

### Synopsis

```bash
r2r eac templates install [options]
```

### Description

Copies template files to the project without processing placeholders. Use this to set up templates that will be applied later or customized first.

### Flags

| Flag       | Short | Type   | Default           | Description                  |
| ---------- | ----- | ------ | ----------------- | ---------------------------- |
| `--type`   | `-t`  | string | `all`             | Template type to install     |
| `--output` | `-o`  | string | `.r2r/templates/` | Installation directory       |
| `--force`  | `-f`  | bool   | `false`           | Overwrite existing templates |

### Examples

```bash
# Install all default templates
r2r eac templates install

# Install documentation templates
r2r eac templates install --type docs

# Install to custom location
r2r eac templates install --output my-templates/

# Force overwrite existing
r2r eac templates install --force
```

### Output

```text
Installing templates...

Documentation templates:
  ✓ README.md.tmpl
  ✓ module-readme.md.tmpl
  ✓ contributing.md.tmpl
  ✓ architecture.md.tmpl

Project templates:
  ✓ .gitignore.tmpl
  ✓ Makefile.tmpl

✓ 6 templates installed to .r2r/templates/
```

### Exit Codes

| Code | Description                           |
| ---- | ------------------------------------- |
| 0    | Templates installed successfully      |
| 1    | Error installing templates            |
| 2    | Template already exists (use --force) |

---

## templates install reports

Install report templates without value replacements.

### Synopsis

```bash
r2r eac templates install reports [options]
```

### Description

Installs report template files for compliance, security, and status reporting. These can be customized before applying.

### Flags

| Flag       | Short | Type   | Default                   | Description                  |
| ---------- | ----- | ------ | ------------------------- | ---------------------------- |
| `--output` | `-o`  | string | `.r2r/templates/reports/` | Installation directory       |
| `--force`  | `-f`  | bool   | `false`                   | Overwrite existing templates |

### Examples

```bash
# Install report templates
r2r eac templates install reports

# Install to custom location
r2r eac templates install reports --output reports/templates/

# Force overwrite
r2r eac templates install reports --force
```

### Output

```text
Installing report templates...

Templates installed:
  ✓ compliance-report.md.tmpl - Compliance summary report
  ✓ security-report.md.tmpl - Security scan report
  ✓ test-report.md.tmpl - Test results report
  ✓ release-notes.md.tmpl - Release notes template
  ✓ status-report.md.tmpl - Project status report

✓ 5 report templates installed
```

### Exit Codes

| Code | Description                           |
| ---- | ------------------------------------- |
| 0    | Templates installed successfully      |
| 1    | Error installing templates            |
| 2    | Template already exists (use --force) |

---

## templates list

List template placeholder variables.

### Synopsis

```bash
r2r eac templates list [options]
```

### Description

Displays available templates and their placeholder variables. Optionally shows current values for placeholders.

### Flags

| Flag            | Short | Type   | Default | Description                     |
| --------------- | ----- | ------ | ------- | ------------------------------- |
| `--show-values` | `-v`  | bool   | `false` | Show current placeholder values |
| `--type`        | `-t`  | string | `all`   | Filter by template type         |
| `--json`        |       | bool   | `false` | Output as JSON                  |

### Examples

```bash
# List all templates
r2r eac templates list

# Show placeholder values
r2r eac templates list --show-values

# Filter by type
r2r eac templates list --type docs

# JSON output
r2r eac templates list --json
```

### Output

```text
Available Templates
═══════════════════════════════════════════════════════

Documentation Templates (.r2r/templates/docs/):
─────────────────────────────────────────────────────
  README.md.tmpl
    Placeholders: {{.Project.Name}}, {{.Project.Description}}, {{.Version}}

  module-readme.md.tmpl
    Placeholders: {{.Module.Moniker}}, {{.Module.Description}}, {{.Module.ImportPath}}

Report Templates (.r2r/templates/reports/):
─────────────────────────────────────────────────────
  compliance-report.md.tmpl
    Placeholders: {{.Date}}, {{.ComplianceScore}}, {{.Findings}}

Available Placeholders:
─────────────────────────────────────────────────────
  Project:
    {{.Project.Name}}        - Project name
    {{.Project.Description}} - Project description
    {{.Project.Repo}}        - Repository name
    {{.Project.RepoURL}}     - Repository URL

  Module:
    {{.Module.Moniker}}      - Module identifier
    {{.Module.Type}}         - Module type
    {{.Module.Description}}  - Module description
    {{.Module.Path}}         - Module path

  Runtime:
    {{.Date}}                - Current date
    {{.DateTime}}            - Current date and time
    {{.Version}}             - Current version
    {{.GitCommit}}           - Git commit SHA
```

### Output with Values

```text
Available Placeholders (with current values):
─────────────────────────────────────────────────────
  {{.Project.Name}}        = "eac"
  {{.Project.Description}} = "Everything as Code CLI toolkit"
  {{.Project.Repo}}        = "ready-to-release/eac"
  {{.Version}}             = "2025.12.01"
  {{.Date}}                = "2025-12-01"
  {{.GitCommit}}           = "abc1234"
```

### Exit Codes

| Code | Description                 |
| ---- | --------------------------- |
| 0    | List generated successfully |
| 1    | Error listing templates     |

---

## templates tags

Extract template tags.

### Synopsis

```bash
r2r eac templates tags [template] [options]
```

### Description

Parses template files to extract all placeholder tags. Useful for validating templates and understanding data requirements.

### Arguments

| Argument   | Required | Description                                    |
| ---------- | -------- | ---------------------------------------------- |
| `template` | No       | Specific template to analyze (defaults to all) |

### Flags

| Flag       | Short | Type | Default | Description           |
| ---------- | ----- | ---- | ------- | --------------------- |
| `--unique` | `-u`  | bool | `false` | Show only unique tags |
| `--json`   |       | bool | `false` | Output as JSON        |

### Examples

```bash
# Extract tags from all templates
r2r eac templates tags

# Extract from specific template
r2r eac templates tags README.md.tmpl

# Unique tags only
r2r eac templates tags --unique

# JSON output
r2r eac templates tags --json
```

### Output

```text
Template Tags Analysis
═══════════════════════════════════════════════════════

README.md.tmpl:
  Line 1:  {{.Project.Name}}
  Line 3:  {{.Project.Description}}
  Line 10: {{.Version}}
  Line 15: {{range .Modules}}
  Line 16: {{.Moniker}}
  Line 16: {{.Description}}
  Line 17: {{end}}

module-readme.md.tmpl:
  Line 1:  {{.Module.Moniker}}
  Line 3:  {{.Module.Description}}
  Line 8:  {{.Module.ImportPath}}

Unique Tags:
─────────────────────────────────────────────────────
  {{.Module.Description}}
  {{.Module.ImportPath}}
  {{.Module.Moniker}}
  {{.Project.Description}}
  {{.Project.Name}}
  {{.Version}}

Control Structures:
  {{range .Modules}} ... {{end}}
```

### Exit Codes

| Code | Description                 |
| ---- | --------------------------- |
| 0    | Tags extracted successfully |
| 1    | Error parsing templates     |
| 2    | Template not found          |

---

## Common Workflows

### Documentation Generation

```bash
# 1. List available templates
r2r eac templates list

# 2. Preview placeholder values
r2r eac templates list --show-values

# 3. Apply documentation templates
r2r eac templates apply docs

# 4. Build documentation site
r2r eac build docs

# 5. Serve locally
r2r eac serve docs
```

### Report Generation

```bash
# 1. Install report templates (first time)
r2r eac templates install reports

# 2. Customize templates if needed
# Edit .r2r/templates/reports/*.tmpl

# 3. Apply templates with current data
r2r eac templates apply --type reports

# 4. Review generated reports
cat out/reports/compliance-report.md
```

### Custom Template Setup

```bash
# 1. Install default templates as base
r2r eac templates install

# 2. Customize templates
vim .r2r/templates/docs/README.md.tmpl

# 3. Verify placeholders
r2r eac templates tags README.md.tmpl

# 4. Preview output
r2r eac templates apply --template README.md.tmpl --dry-run

# 5. Apply template
r2r eac templates apply --template README.md.tmpl
```

### Module Documentation

```bash
# Generate README for each module
for module in eac-commands eac-core r2r-cli; do
  r2r eac templates apply docs --module $module
done

# Or apply to all modules at once
r2r eac templates apply --template module-readme.md.tmpl
```

---

## Template Syntax

### Placeholders

```markdown
# {{.Project.Name}}

{{.Project.Description}}

Version: {{.Version}}
```

### Conditionals

```markdown
{{if .Module.HasTests}}
## Testing

Run tests with: `r2r eac test {{.Module.Moniker}}`
{{end}}
```

### Loops

```markdown
## Modules

{{range .Modules}}
- **{{.Moniker}}**: {{.Description}}
{{end}}
```

### Context Blocks

```markdown
{{with .Module}}
Module: {{.Moniker}} ({{.Type}})
Path: {{.Path}}
{{end}}
```

### Comments

```markdown
{{/* This is a comment and won't appear in output */}}
```

---

## Integration Patterns

### CI/CD Documentation Updates

```yaml
name: Update Docs

on:
  push:
    branches: [main]

jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Generate documentation
        run: r2r eac templates apply docs

      - name: Commit changes
        run: |
          git add docs/
          git diff --cached --quiet || \
            git commit -m "docs: update from templates"
          git push
```

### With MkDocs

```bash
# Generate templated docs
r2r eac templates apply docs

# Build MkDocs site
r2r eac build docs

# Serve locally
r2r eac serve docs
```

### With Books

```bash
# Apply templates first
r2r eac templates apply docs

# Then build book
r2r eac build docs-book
```

---

## Best Practices

### Template Design

- **Keep it simple** - Complex logic belongs in code
- **Use consistent naming** - Intuitive placeholder names
- **Document placeholders** - Help users understand values
- **Test templates** - Verify output with real data

### Do's

- Use clear, simple placeholders
- Provide default values where possible
- Keep templates focused on one purpose
- Version control templates

### Don'ts

- Avoid complex nested logic in templates
- Don't hardcode values that should be placeholders
- Don't create templates for one-time use

---

## Troubleshooting

| Problem                  | Solution                               |
| ------------------------ | -------------------------------------- |
| Placeholder not replaced | Check spelling with `templates list`   |
| Missing value            | Ensure data source exists              |
| Template syntax error    | Validate Go template syntax            |
| Empty output             | Check conditionals, verify data exists |
| Template not found       | Check path and file extension          |

---

## Related Documentation

- [Templates Overview](templates-overview.md) - Concepts and placeholder reference
- [Templates Configuration](templates-configuration.md) - Configuration options
- [Books Commands](books-commands.md) - Documentation aggregation
