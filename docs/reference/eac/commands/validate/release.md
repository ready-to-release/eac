# validate release

<!-- book:cmd validate release -->

## What It Checks

- **File exists** -- Changelog exists at the path defined in the module contract.
- **Valid header** -- Title follows Keep a Changelog conventions.
- **Version format** -- Each entry uses valid semver (`x.y.z`) or calver (`YYYY.MM.DD`) format.
- **No duplicates** -- No two entries share the same version number.
- **Descending order** -- Versions are ordered newest-first (checked by date).
- **Non-empty versions** -- Each version has at least one entry (warning).

## Common Errors

- **changelog not found** -- The file does not exist at the contract-specified path.
- **invalid version format** -- A version does not match semver or calver patterns.
- **duplicate version** -- The same version number appears more than once.

## See Also

- [release changelog](../release/changelog.md)
- [release this](../release/this.md)
- [validate release-version](./release-version.md)
- [validate Commands](../validate/index.md)
