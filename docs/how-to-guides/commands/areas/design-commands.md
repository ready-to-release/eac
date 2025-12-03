# Design Commands

Command reference for EAC's architecture design system.

## Quick Reference

| Command           | Description                                                  |
| ----------------- | ------------------------------------------------------------ |
| `create-design`   | Generate workspace.dsl for a module using AI                 |
| `update-design`   | Update existing workspace.dsl for a module using AI          |
| `validate-design` | Check workspace.dsl syntax using Structurizr CLI             |
| `serve-design`    | View architecture diagrams in browser using Structurizr Lite |

---

## create-design

Generate a Structurizr DSL workspace file for a module using AI analysis.

### Synopsis

```bash
r2r eac create-design <module> [options]
```

### Description

Analyzes module code structure, dependencies, and contracts to generate a C4 model architecture diagram in Structurizr DSL format. The AI examines:

- Package structure and interfaces
- Module dependencies from contracts
- External service calls
- Database and cache usage
- API endpoints

### Arguments

| Argument | Required | Description                           |
| -------- | -------- | ------------------------------------- |
| `module` | Yes      | Module moniker to generate design for |

### Flags

| Flag              | Short | Type   | Default                          | Description                                   |
| ----------------- | ----- | ------ | -------------------------------- | --------------------------------------------- |
| `--output`        | `-o`  | string | `<module>/.design/workspace.dsl` | Output file path                              |
| `--level`         | `-l`  | string | `container`                      | Diagram level (context, container, component) |
| `--include-tests` |       | bool   | `false`                          | Include test packages in diagram              |
| `--debug`         | `-d`  | bool   | `false`                          | Save AI prompts and responses                 |

### Examples

```bash
# Generate design for a module
r2r eac create-design eac-commands

# Generate component-level diagram
r2r eac create-design eac-core --level component

# Include test packages
r2r eac create-design eac-commands --include-tests

# Custom output path
r2r eac create-design eac-commands --output docs/architecture/commands.dsl

# Debug mode
r2r eac create-design eac-commands --debug
```

### Output

```text
Generating architecture design for eac-commands...

Analyzing module structure...
  ✓ 12 packages found
  ✓ 8 external dependencies
  ✓ 3 internal dependencies

Generating Structurizr DSL...
  ✓ Container diagram created
  ✓ 5 containers identified
  ✓ 12 relationships mapped

✓ Design saved: go/eac/commands/.design/workspace.dsl

To view: r2r eac serve-design eac-commands
```

### Exit Codes

| Code | Description                   |
| ---- | ----------------------------- |
| 0    | Design generated successfully |
| 1    | Error generating design       |
| 2    | Module not found              |

---

## update-design

Update an existing workspace.dsl file with changes detected in the codebase.

### Synopsis

```bash
r2r eac update-design <module> [options]
```

### Description

Compares current code structure with existing design and updates the workspace.dsl to reflect changes. Preserves manual customizations while adding new elements and removing obsolete ones.

Changes detected:

- New packages or components
- Removed packages
- New dependencies
- Changed relationships

### Arguments

| Argument | Required | Description              |
| -------- | -------- | ------------------------ |
| `module` | Yes      | Module moniker to update |

### Flags

| Flag        | Short | Type | Default | Description                                 |
| ----------- | ----- | ---- | ------- | ------------------------------------------- |
| `--dry-run` | `-n`  | bool | `false` | Show changes without applying               |
| `--force`   | `-f`  | bool | `false` | Overwrite without preserving customizations |
| `--debug`   | `-d`  | bool | `false` | Save AI prompts and responses               |

### Examples

```bash
# Update design for module
r2r eac update-design eac-commands

# Preview changes without applying
r2r eac update-design eac-commands --dry-run

# Force complete regeneration
r2r eac update-design eac-commands --force

# Debug mode
r2r eac update-design eac-commands --debug
```

### Output

```text
Updating architecture design for eac-commands...

Analyzing changes since last update...
  + New package: impl/security
  + New dependency: eac-ai
  - Removed package: impl/deprecated
  ~ Modified: impl/build (2 new components)

Updating workspace.dsl...
  ✓ Added container: Security
  ✓ Added relationship: Commands -> AI
  ✓ Removed container: Deprecated
  ✓ Updated container: Build

✓ Design updated: go/eac/commands/.design/workspace.dsl

Changes:
  Added: 2 elements, 1 relationship
  Removed: 1 element
  Modified: 1 element
```

### Dry Run Output

```text
Updating architecture design for eac-commands (dry-run)...

Changes that would be applied:
  + Add container "Security" with 3 components
  + Add relationship "Commands -> AI"
  - Remove container "Deprecated"
  ~ Update container "Build" (add 2 components)

No changes written (dry-run mode)
```

### Exit Codes

| Code | Description                                  |
| ---- | -------------------------------------------- |
| 0    | Design updated successfully                  |
| 1    | Error updating design                        |
| 2    | Module not found                             |
| 3    | No existing design found (use create-design) |

---

## validate-design

Validate workspace.dsl syntax using Structurizr CLI.

### Synopsis

```bash
r2r eac validate-design [module] [options]
```

### Description

Checks Structurizr DSL files for syntax errors and structural issues:

- DSL syntax validation
- Element reference verification
- Relationship target validation
- View configuration checks

Requires Docker for Structurizr CLI execution.

### Arguments

| Argument | Required | Description                                   |
| -------- | -------- | --------------------------------------------- |
| `module` | No       | Module to validate (validates all if omitted) |

### Flags

| Flag       | Short | Type | Default | Description                       |
| ---------- | ----- | ---- | ------- | --------------------------------- |
| `--all`    | `-a`  | bool | `false` | Validate all modules with designs |
| `--strict` |       | bool | `false` | Fail on warnings                  |

### Examples

```bash
# Validate specific module
r2r eac validate-design eac-commands

# Validate all modules
r2r eac validate-design --all

# Strict validation
r2r eac validate-design eac-commands --strict
```

### Output (Success)

```text
Validating design: eac-commands

  ✓ DSL syntax valid
  ✓ All element references valid
  ✓ All relationship targets exist
  ✓ View configuration valid

✓ Design validation passed
```

### Output (Failure)

```text
Validating design: eac-commands

  ✗ DSL syntax errors:
    - Line 15: Unknown identifier 'unknownSystem'
    - Line 42: Missing closing brace

  ⚠ Warnings:
    - Line 30: Container 'api' has no relationships

✗ Design validation failed (2 errors, 1 warning)
```

### Exit Codes

| Code | Description           |
| ---- | --------------------- |
| 0    | Validation passed     |
| 1    | Validation failed     |
| 2    | Design file not found |
| 3    | Docker not available  |

---

## serve-design

View architecture diagrams in browser using Structurizr Lite.

### Synopsis

```bash
r2r eac serve-design <module> [options]
```

### Description

Launches Structurizr Lite in Docker to provide an interactive viewer for architecture diagrams. Features:

- Multiple diagram views
- Interactive navigation
- Export to PNG/SVG
- Documentation rendering

### Arguments

| Argument | Required | Description    |
| -------- | -------- | -------------- |
| `module` | Yes      | Module to view |

### Flags

| Flag       | Short | Type   | Default | Description                                 |
| ---------- | ----- | ------ | ------- | ------------------------------------------- |
| `--port`   | `-p`  | int    | `8080`  | Port to serve on                            |
| `--detach` | `-d`  | bool   | `false` | Run in background                           |
| `--export` | `-e`  | string | -       | Export format (png, svg) instead of serving |

### Examples

```bash
# View design in browser
r2r eac serve-design eac-commands
# Opens http://localhost:8080

# Use custom port
r2r eac serve-design eac-commands --port 9090

# Run in background
r2r eac serve-design eac-commands --detach

# Export diagrams instead of serving
r2r eac serve-design eac-commands --export png
```

### Output

```text
Starting Structurizr Lite...

  Module: eac-commands
  Workspace: go/eac/commands/.design/workspace.dsl
  URL: http://localhost:8080

  Press Ctrl+C to stop

Structurizr Lite is running...
```

### Export Output

```text
Exporting diagrams for eac-commands...

  ✓ Container diagram: .design/export/container.png
  ✓ Component diagram: .design/export/component.png

Exported 2 diagrams to .design/export/
```

### Exit Codes

| Code | Description             |
| ---- | ----------------------- |
| 0    | Server stopped normally |
| 1    | Error starting server   |
| 2    | Design file not found   |
| 3    | Docker not available    |
| 4    | Port already in use     |

---

## Common Workflows

### Initial Design Creation

```bash
# 1. Generate design
r2r eac create-design eac-commands

# 2. Validate syntax
r2r eac validate-design eac-commands

# 3. View in browser
r2r eac serve-design eac-commands

# 4. Customize as needed
# Edit go/eac/commands/.design/workspace.dsl

# 5. Validate again
r2r eac validate-design eac-commands
```

### Keeping Designs Current

```bash
# After code changes
r2r eac update-design eac-commands --dry-run

# Review changes, then apply
r2r eac update-design eac-commands

# Validate
r2r eac validate-design eac-commands
```

### CI/CD Integration

```bash
# Validate all designs in CI
r2r eac validate-design --all

# Check for drift
r2r eac update-design eac-commands --dry-run
if [ $? -ne 0 ]; then
  echo "Design needs update"
  exit 1
fi
```

### Documentation Export

```bash
# Export all diagrams
for module in eac-commands eac-core r2r-cli; do
  r2r eac serve-design $module --export svg
  cp $module/.design/export/*.svg docs/diagrams/
done
```

---

## Related Documentation

- [Design Overview](design-overview.md) - Concepts and workflows
- [Design Configuration](design-configuration.md) - Configuration reference
- [Templates Commands](templates-commands.md) - Documentation templates
