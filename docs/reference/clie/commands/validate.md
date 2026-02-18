# clie validate

Validate the syntax and structure of `.clie/clie.yml` configuration file.

## Syntax

```bash
clie validate [flags]
```

## Description

The `validate` command checks your CLIE configuration file for syntax errors and structural issues.

Use this to catch configuration errors before running commands.

**What it validates:**

- YAML syntax
- Required fields (extension name, image)
- Valid Docker image naming
- No duplicate extension names

## Examples

### Basic Validation

```bash
clie validate
```

**Output when valid:**

```text
✓ Configuration is valid
```

**Output when invalid:**

```text
✗ Configuration validation failed:
  - Line 3: Missing required field 'image' for extension 'eac'
  - Line 7: Invalid image name format 'invalid@image'
```

### Validate Before Commit

```bash
# Validate configuration
clie validate

# If valid, commit
if [ $? -eq 0 ]; then
    git add .clie/clie.yml
    git commit -m "Update CLIE configuration"
fi
```

## Validation Rules

### Extension Configuration

Each extension must have:

```yaml
extensions:
  - name: 'eac'                                        # Required: string
    image: 'ghcr.io/ready-to-release/eac-ext:latest' # Required: valid image reference
    description: 'EAC automation'                      # Optional: string
```

### Valid Image Reference Format

**Valid examples:**

- `ghcr.io/ready-to-release/eac-ext:latest`
- `ghcr.io/ready-to-release/eac-ext:v1.2.3`
- `eac-ext:dev` (local images)

**Invalid examples:**

- `invalid@image` (@ not allowed)
- `image name with spaces:latest`
- `:latest` (missing image name)

### Extension Names

**Valid names:**

- Alphanumeric characters
- Hyphens allowed
- No spaces
- Unique across configuration

**Examples:**

- ✅ `eac`
- ✅ `eac-dev`
- ❌ `my extension` (spaces not allowed)

## Common Validation Errors

### Error: "Missing required field"

```yaml
# Before (invalid)
extensions:
  - name: 'eac'

# After (valid)
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
```

### Error: "Invalid YAML syntax"

```yaml
# Before (invalid - bad indentation)
extensions:
- name: 'eac'
 image: 'ghcr.io/ready-to-release/eac-ext:latest'

# After (valid - consistent indentation)
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
```

### Error: "Duplicate extension"

```yaml
# Before (invalid)
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
  - name: 'eac'                                    # Duplicate!
    image: 'ghcr.io/ready-to-release/eac-ext:dev'

# After (valid)
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
  - name: 'eac-dev'                                # Renamed
    image: 'ghcr.io/ready-to-release/eac-ext:dev'
```

## See Also

- [CLIE CLI Overview](index.md) - Command overview
- [init command](init.md) - Initialize configuration
- [install command](install.md) - Install extensions
- [Configuration Reference](../configuration.md) - Detailed configuration guide
- [verify command](verify.md) - Verify system prerequisites
