# show build-summary

<!-- book:cmd show build-summary -->

## Output Sections

### On success

- **Header**: module name with build emoji
- **Status**: component types that were built
- **Build Output**: artifact verification table (artifact pattern, type, size/file count)
- **Artifacts**: artifact bundle name for download
- **Build Configuration** (collapsible): component types, container runtime, output directory

### On failure

- **Header**: module name with failure indicator
- **Status**: failure message
- **Diagnostics**: last 30 lines of each component build log
- **Timing**: build timing data if available
- **Build Configuration** (collapsible): same as success

## See Also

- [build](../build/build.md) - Build modules
- [show build-times](./build-times.md) - Performance analysis
- [get build-times](../get/build-times.md) - JSON output
