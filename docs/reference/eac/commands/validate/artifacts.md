# Validate artifacts

<!-- book:cmd validate artifacts -->

## How It Works

Validates build artifacts exist for a module and its dependencies:

- **Artifact Location**: Checks `out/artifacts/<module>/` directory
- **Dependency Artifacts**: Verifies all transitive dependencies have artifacts
- **File Validation**: Ensures artifact files are complete and accessible
- **Checksum Verification**: Validates artifact integrity
- **Missing Detection**: Reports which artifacts are missing

Used before releases to ensure all required artifacts are available.

## See Also

- [validate](./validate.md)
- [get artifacts](../get/artifacts.md)
- [build](../build/build.md)
- [validate Commands](../categories/validate.md)
