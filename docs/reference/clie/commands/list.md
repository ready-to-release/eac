# clie list

List available extensions from the registry that can be installed.

## Syntax

```bash
clie list [flags]
```

## Description

The `list` command displays all extensions available in the CLIE registry.

Use this to discover extensions you can install.

## Examples

### List All Extensions

```bash
clie list
```

**Example output:**

```text
Available Extensions:

eac                Everything-as-Code automation framework
```

### Check Before Installing

```bash
# See what's available
clie list

# Install EAC
clie install eac
```

## EAC Extension

**Full name**: Everything-as-Code (EAC)

**Provides:**

- Build automation
- Test orchestration
- Code validation
- Security scanning
- Release management
- AI-powered workflows

**Install:**

```bash
clie install eac
```

**Documentation:**

- [EAC Commands Reference](../../eac/commands/index.md)
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md)

## Registry Information

Extensions are hosted in the GitHub Container Registry (GHCR):

- **Registry URL**: `ghcr.io/ready-to-release/`
- **Naming pattern**: `ext-<extension-name>`
- **Access**: Public read access (no authentication needed)

**Example:**

- `eac` → `ghcr.io/ready-to-release/eac-ext:latest`

## See Also

- [CLIE CLI Overview](index.md) - Command overview
- [install command](install.md) - Install extensions from the list
- [Configuration Reference](configuration.md) - Extension configuration
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Getting started guide
