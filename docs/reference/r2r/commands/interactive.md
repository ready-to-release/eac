# r2r interactive

Start an extension container in interactive mode for debugging and exploration.

## Syntax

```bash
r2r interactive <extension-name> [flags]
```

## Description

The `interactive` command starts an extension's Docker container with an interactive shell. Useful for debugging, exploring the container environment, or running commands manually inside the extension.

**Use cases:**

- Debugging extension behavior
- Exploring container filesystem and tools
- Testing commands inside the container
- Troubleshooting extension issues

## Examples

### Basic Interactive Session

```bash
r2r interactive eac
```

Opens a shell inside the EAC extension container:

```text
root@container:/workspace#
```

### Run Commands in Container

```bash
r2r interactive eac

# Inside container:
root@container:/workspace# ls
root@container:/workspace# which eac
root@container:/workspace# eac --help
root@container:/workspace# exit
```

### Debug Extension Issue

```bash
# Start interactive session
r2r interactive eac

# Inside container, run the failing command
root@container:/workspace# eac build --debug

# Examine logs, filesystem, environment
root@container:/workspace# env | grep EAC
root@container:/workspace# ls -la .r2r/eac/
```

## Container Environment

When you enter interactive mode, you have access to:

- **Extension binaries** - The extension's command-line tools
- **Workspace mount** - Your project files at `/workspace`
- **Configuration** - Extension config from `.r2r/<extension>/`
- **Shell environment** - Bash shell
- **Debugging tools** - Tools included in the container image

### Workspace Mounting

Your current directory is mounted at `/workspace` in the container:

```bash
# Host filesystem
/home/user/my-project/
├── .r2r/
├── src/
└── README.md

# Inside container
/workspace/
├── .r2r/
├── src/
└── README.md
```

**Note:** Changes made inside `/workspace` affect your host files.

## See Also

- [R2R CLI Overview](index.md) - Command overview
- [metadata command](metadata.md) - Get extension metadata
- [install command](install.md) - Install extensions
- [Configuration Reference](configuration.md) - Extension configuration
