# r2r install

Install the EAC extension by pulling its Docker image. When no extension name is provided, installs all configured extensions.

## Syntax

```bash
r2r install [extension-name] [flags]
```

## Description

The `install` command manages extension installation. It pulls Docker images from the container registry and registers them for use.

**Without extension name:**

1. Reads `.r2r/r2r-cli.yml` for configured extensions
2. Pulls Docker images for all extensions
3. Makes extension commands available

**With extension name:**

1. Adds the extension to `.r2r/r2r-cli.yml`
2. Pulls the extension's Docker image
3. Registers the extension for use via `r2r <extension>` command

## Flags

| Flag           | Description                                                   |
| -------------- | ------------------------------------------------------------- |
| `--load-local` | Use local development images instead of pulling from registry |
| `-h, --help`   | Display help information                                      |

## Examples

### Install EAC Extension

```bash
r2r install eac
```

**What happens:**

1. Adds eac to `.r2r/r2r-cli.yml`:

   ```yaml
   extensions:
     - name: 'eac'
       image: 'ghcr.io/ready-to-release/ext-eac:latest'
   ```

2. Pulls `ghcr.io/ready-to-release/ext-eac:latest` Docker image
3. Makes `r2r eac` commands available

### Install All Configured Extensions

After cloning a repository with existing R2R configuration:

```bash
r2r install
```

### Three-Step Setup

```bash
# 1. Initialize R2R configuration
r2r init

# 2. Install EAC extension
r2r install eac

# 3. Configure EAC
r2r eac init --ai-provider claude-api
```

### Local Development

Install using a local Docker image (for extension developers):

```bash
r2r install eac --load-local
```

## Configuration Changes

### Before Install

`.r2r/r2r-cli.yml`:

```yaml
extensions: []
```

### After `r2r install eac`

`.r2r/r2r-cli.yml`:

```yaml
extensions:
  - name: 'eac'
    image: 'ghcr.io/ready-to-release/ext-eac:latest'
    description: 'Everything-as-Code automation'
```

## Next Steps

After installing the extension:

1. **Configure the extension:**

   ```bash
   r2r eac init --ai-provider claude-api
   ```

2. **Verify installation:**

   ```bash
   r2r eac --help
   ```

3. **Start using commands:**

   ```bash
   r2r eac show modules
   r2r eac build
   r2r eac test
   ```

## See Also

- [R2R CLI Overview](index.md) - Command overview and architecture
- [init command](init.md) - Initialize configuration before install
- [list command](list.md) - Browse available extensions
- [cleanup command](cleanup.md) - Remove old images to free space
- [Configuration Reference](configuration.md) - Detailed configuration options
- [Quick Start Tutorial](../../tutorials/getting-started/quick-start.md) - Complete setup guide
