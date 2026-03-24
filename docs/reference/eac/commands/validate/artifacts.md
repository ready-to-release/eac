# validate artifacts

<!-- book:cmd validate artifacts -->

## What It Checks

- Target module artifacts in `out/build/` (executables, files, directories).
- All transitive dependency artifacts (recursive, unless `--skip-depm`).
- Platform-specific artifacts for the target OS/architecture.
- Marker files for modules with no traditional build outputs.

## Common Errors

- **module not found** -- The moniker does not exist. Check `eac show modules`.
- **Missing artifacts detected** -- Required build outputs are missing. Run `eac build <module>`.

## See Also

- [show artifacts](../show/artifacts.md)
- [validate](./validate.md)
