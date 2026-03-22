# extension-meta

<!-- book:cmd extension-meta -->

Outputs extension metadata for [CLIE](../../clie/index.md) CLI integration.

> **Note:** This is a special-purpose command used by the [CLIE extension host](../../clie/index.md) to discover and configure EAC as a containerized extension. Most users do not need to call this directly.

## Usage

```bash
eac extension-meta
```

## Description

The `extension-meta` command outputs YAML-formatted metadata describing the EAC extension's capabilities, commands, requirements, and configuration. This metadata is consumed by the CLIE CLI during `clie install eac` to register available commands and validate system requirements.

For more on how CLIE discovers and manages extensions, see the [CLIE architecture documentation](../../clie/architecture/index.md).

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
