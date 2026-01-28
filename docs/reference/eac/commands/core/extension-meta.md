# extension-meta

<!-- book:cmd extension-meta -->

Outputs extension metadata for r2r CLI integration.

## Usage

```bash
eac extension-meta
```

## Description

The `extension-meta` command outputs YAML-formatted metadata describing the EAC extension's capabilities, commands, requirements, and configuration. This metadata is used by the r2r CLI to discover and configure extensions.

## Output

The command outputs YAML-formatted metadata including:
- Extension name and version
- Available commands
- System requirements
- Configuration schema

## Examples

```bash
# Output extension metadata
eac extension-meta
```
