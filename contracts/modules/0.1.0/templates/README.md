# Module Contract Templates

This directory contains template files for creating new module contracts.

## Using the Template

### Quick Start

1. **Copy the template:**
   ```bash
   cp contracts/modules/0.1.0/templates/default.yml contracts/modules/0.1.0/my-module.yml
   ```

2. **Edit required fields:**
   - `moniker`: Must match filename (without `.yml`)
   - `name`: Human-readable name for your module
   - `source.root`: Root directory path

3. **Customize optional fields** as needed

4. **Remove placeholders** like `(no-default)`

### Example: Creating "my-docs" Module

```bash
# 1. Copy template
cp contracts/modules/0.1.0/templates/default.yml contracts/modules/0.1.0/my-docs.yml

# 2. Edit the file
```

```yaml
# Version: 0.1.0

moniker: "my-docs"
name: "My Documentation"
type: "documentation"
description: "Documentation for my project"
parent: "."

versioning:
  version_scheme: "semver"

source:
  root: "docs/my-docs"
  includes:
    - "**/*.md"
  changelog_path: "docs/my-docs/CHANGELOG.md"
  exclude_children_owned_source: true

depends_on: []
used_by: []
```

## Field Reference

### Required Fields
- `moniker` - Unique ID (must match filename)
- `name` - Display name
- `source.root` - Root directory

### Common Customizations

**Module Type:**
```yaml
type: "documentation"  # or "source-code", "configuration", etc.
```

**Parent Module:**
```yaml
parent: "docs"  # Make this a child of another module
```

**Custom Includes:**
```yaml
source:
  includes:
    - "**/*.md"      # Only markdown files
    - "**/*.yaml"    # Only YAML files
```

**Shared Ownership with Children:**
```yaml
source:
  exclude_children_owned_source: false  # Allow overlap with children
```

## Regenerating Templates

If the schema changes, regenerate the template:

```bash
cd out
go run set-module-contract.go
```

## See Also

- [SET-MODULE-CONTRACT.md](../../../out/SET-MODULE-CONTRACT.md) - Detailed documentation
- [DEPENDS-ON-DEFAULT.md](../../../out/DEPENDS-ON-DEFAULT.md) - Default values
- [EXCLUDE-CHILDREN-OWNED-SOURCE.md](../../../out/EXCLUDE-CHILDREN-OWNED-SOURCE.md) - Ownership rules
