# Show deps-setup-summary

<!-- book:cmd show deps-setup-summary -->

Generate a dependencies setup summary, formatted as Markdown suitable for `$GITHUB_STEP_SUMMARY`.

Shows which build dependencies were installed versus already available on the CI runner. Supports Go, Node.js, Docker Buildx, QEMU, and UPX.

## Usage

```
eac show deps-setup-summary --module=<name> --deps=<list> [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--module` | string | Module name (required) |
| `--deps` | string | Comma-separated list of dependencies (`go`, `node`, `docker`, `buildx`, `upx`) |
| `--go-available` | bool | Whether Go was already available |
| `--node-available` | bool | Whether Node.js was already available |
| `--buildx-available` | bool | Whether Docker Buildx was already available |
| `--qemu-available` | bool | Whether QEMU was already available |
| `--upx-available` | bool | Whether UPX was already available |

## Output Sections

- **Header**: "Dependencies Setup for `<module>`"
- **Dependencies table**: one row per dependency with status ("Already available" or "Installed"/"Configured")

The `--deps` flag controls which rows appear. The `--*-available` flags determine whether each dependency is shown as pre-existing or newly installed.

## Examples

```bash
# Go module with Go pre-installed
eac show deps-setup-summary --module=core --deps=go --go-available

# Container module needing buildx and QEMU
eac show deps-setup-summary --module=oci-tools --deps=docker,buildx

# Full setup with Node.js
eac show deps-setup-summary --module=docs --deps=go,node --go-available --node-available

# Redirect to GitHub Actions step summary
eac show deps-setup-summary --module=core --deps=go --go-available >> "$GITHUB_STEP_SUMMARY"
```

## See Also

- [show](show.md)
- [dependencies](dependencies.md)
