# Show Artifacts

<!-- book:cmd show artifacts -->

Displays all build artifacts for a module in a formatted table, showing resolved names, existence status, and file paths. By default shows artifacts for the current platform.

## Usage

```bash
eac show artifacts <module> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `<module>` | Module name (required) |

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--all-platforms` | bool | Show artifacts for all platforms (linux, darwin, windows) |
| `--missing-only` | bool | Only show artifacts that do not exist on disk |
| `--os <os>` | string | Target OS override (default: current OS) |
| `--arch <arch>` | string | Target architecture override (default: current arch) |

## Output Sections

The report includes:

1. **Summary Header** - Module name, component types, build directory, platform, and metadata override count.
2. **Build Modes** - Breakdown of which artifacts are built in default mode (current platform) vs `--all` mode (all platforms).
3. **Artifact Table** - Each artifact with its resolved name, existence status (checkmark or X), and resolved file path. Includes total/exists/missing counts.
4. **Metadata Overrides** - If any artifact metadata is defined, shows the override mappings.

## Examples

```bash
# Show artifacts for a module on current platform
eac show artifacts eac

# Show artifacts across all platforms
eac show artifacts clie --all-platforms

# Show only missing artifacts
eac show artifacts eac --missing-only

# Show artifacts for a specific platform
eac show artifacts clie --os linux --arch arm64
```

## See Also

- [get artifacts](../get/artifacts.md) - JSON output
- [build](../build/build.md) - Build modules
