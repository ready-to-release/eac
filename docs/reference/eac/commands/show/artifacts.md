# show artifacts

<!-- book:cmd show artifacts -->

## Output Sections

The report includes:

1. **Summary Header** - Module name, component types, build directory, platform, and metadata override count.
2. **Build Modes** - Breakdown of which artifacts are built in default mode (current platform) vs `--all` mode (all platforms).
3. **Artifact Table** - Each artifact with its resolved name, existence status (checkmark or X), and resolved file path. Includes total/exists/missing counts.
4. **Metadata Overrides** - If any artifact metadata is defined, shows the override mappings.

## See Also

- [get artifacts](../get/artifacts.md) - JSON output
- [build](../build/build.md) - Build modules
