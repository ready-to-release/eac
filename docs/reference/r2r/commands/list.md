# r2r list

List available extensions from the registry that can be installed.

## Syntax

```bash
r2r list [flags]
```

## Description

The `list` command displays all extensions available in the R2R registry. Use this to discover extensions you can install.

## Examples

### List All Extensions

```bash
r2r list
```

**Example output:**

```text
Available Extensions:

eac                Everything-as-Code automation framework
```

### Check Before Installing

```bash
# See what's available
r2r list

# Install EAC
r2r install eac
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
r2r install eac
```

**Documentation:**

- [EAC Commands Reference](../commands/index.md)
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md)

## Registry Information

Extensions are hosted in the GitHub Container Registry (GHCR):

- **Registry URL**: `ghcr.io/ready-to-release/`
- **Naming pattern**: `ext-<extension-name>`
- **Access**: Public read access (no authentication needed)

**Example:**

- `eac` → `ghcr.io/ready-to-release/ext-eac:latest`

## See Also

- [R2R CLI Overview](index.md) - Command overview
- [install command](install.md) - Install extensions from the list
- [Configuration Reference](../configuration/) - Extension configuration
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Getting started guide
