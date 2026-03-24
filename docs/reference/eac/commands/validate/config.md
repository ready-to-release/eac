# validate config

<!-- book:cmd validate config -->

## Validation Phases

1. **File Checks** -- Verifies contract defaults (`contracts/*/defaults/`), user config (`.eac/`), and personal overrides (`.eac/*.personal.yml`) are readable.
2. **Load Validation** -- Loads repository, environments, security, and testing configs with schema validation enabled.
3. **Cross-Reference Validation** -- Checks that component types reference valid entries in `component-types.yml`, module dependencies reference existing modules, and `component_deps` entries point to valid targets.

## Output Formats

- **text** -- Human-readable summary with file status, errors, and warnings.
- **json** -- Structured JSON with `valid`, `errors`, `warnings`, and `files_loaded` fields.
- **github** -- GitHub Actions annotation format (`::error file=...::message`).

## Common Errors

- **cannot read file** -- A config file exists but is not readable (permissions issue).
- **failed to load configuration** -- YAML syntax error or schema mismatch.
- **unknown type** -- A component references a type not defined in `component-types.yml`.
- **unknown dependency** -- A module's `depends_on` references a non-existent module.

## See Also

- [show config](../show/config.md)
- [validate](./validate.md)
