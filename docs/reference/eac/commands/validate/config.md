# Validate config

<!-- book:cmd validate config -->

Validates the complete configuration stack from all sources: contract defaults, user overrides, and personal overrides. Runs three validation phases: file checks, schema validation, and cross-reference validation.

## Usage

```bash
eac validate config [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--strict` | bool | Treat warnings as errors |
| `--format` | string | Output format: `text` (default), `json`, `github` |

## Validation Phases

1. **File Checks** -- Verifies contract defaults (`contracts/*/defaults/`), user config (`.eac/`), and personal overrides (`.eac/*.personal.yml`) are readable.
2. **Load Validation** -- Loads repository, environments, security, and testing configs with schema validation enabled.
3. **Cross-Reference Validation** -- Checks that component types reference valid entries in `component-types.yml`, module dependencies reference existing modules, and `component_deps` entries point to valid targets.

## Output Formats

- **text** -- Human-readable summary with file status, errors, and warnings.
- **json** -- Structured JSON with `valid`, `errors`, `warnings`, and `files_loaded` fields.
- **github** -- GitHub Actions annotation format (`::error file=...::message`).

## Examples

```bash
# Basic validation
eac validate config

# Strict mode (warnings become errors)
eac validate config --strict

# Machine-readable output
eac validate config --format json
```

## Common Errors

- **cannot read file** -- A config file exists but is not readable (permissions issue).
- **failed to load configuration** -- YAML syntax error or schema mismatch.
- **unknown type** -- A component references a type not defined in `component-types.yml`.
- **unknown dependency** -- A module's `depends_on` references a non-existent module.

## See Also

- [show config](../show/config.md)
- [validate](./validate.md)
