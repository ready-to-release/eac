# clie install

Install the EAC extension by pulling its Docker image.

When no extension name is provided, installs all configured extensions.

## Syntax

```bash
clie install [extension-name] [flags]
```

## Description

The `install` command manages extension installation.

It pulls Docker images from the container registry and registers them for use.

**Without extension name:**

1. Reads `.clie/clie.yml` for configured extensions
2. Pulls Docker images for all extensions
3. Makes extension commands available

**With extension name:**

1. Adds the extension to `.clie/clie.yml`
2. Pulls the extension's Docker image
3. Registers the extension for use via `clie <extension>` command

## Flags

| Flag           | Description                                                   |
| -------------- | ------------------------------------------------------------- |
| `--load-local` | Use local development images instead of pulling from registry |
| `-h, --help`   | Display help information                                      |

## Examples

### Install EAC Extension

```bash
clie install eac
```

**What happens:**

1. Adds eac to `.clie/clie.yml`:

   ```yaml
   extensions:
     - name: 'eac'
       image: 'ghcr.io/ready-to-release/eac-ext:latest'
   ```

2. Pulls `ghcr.io/ready-to-release/eac-ext:latest` Docker image
3. Makes `clie eac` commands available

### Install All Configured Extensions

After cloning a repository with existing CLIE configuration:

```bash
clie install
```

### Three-Step Setup

```bash
# 1. Initialize CLIE configuration
clie init

# 2. Install EAC extension
clie install eac

# 3. Configure EAC
clie eac init --ai-provider claude-api
```

### Local Development

Install using a local Docker image (for extension developers):

```bash
clie install eac --load-local
```

## Configuration Changes

### Before Install

`.clie/clie.yml`:

```yaml
extensions: []
```

### After `clie install eac`

`.clie/clie.yml`:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/eac-ext:latest'
    description: 'Everything-as-Code automation'
```

## Next Steps

After installing the extension:

1. **Configure the extension:**

   ```bash
   clie eac init --ai-provider claude-api
   ```

2. **Verify installation:**

   ```bash
   clie eac --help
   ```

3. **Start using commands:**

   ```bash
   clie eac show modules
   clie eac build
   clie eac test
   ```

## See Also

- [CLIE CLI Overview](index.md) - Command overview and architecture
- [init command](init.md) - Initialize configuration before install
- [list command](list.md) - Browse available extensions
- [cleanup command](cleanup.md) - Remove old images to free space
- [Configuration Reference](configuration.md) - Detailed configuration options
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Complete setup guide
