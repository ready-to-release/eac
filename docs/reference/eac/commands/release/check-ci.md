# release check-ci

<!-- book:cmd release check-ci -->

## How It Works

Verifies CI passed for a commit before allowing release:

- **Workflow Status**: Queries GitHub Actions API for workflow runs
- **Module Coverage**: Ensures all affected modules have passing CI
- **Artifact Validation**: Confirms build artifacts exist and are valid
- **Failure Prevention**: Blocks releases if CI is pending or failed
- **Timeout Handling**: Reports if CI is taking too long

Prevents releasing untested or broken code.

## See Also

- [release this](./this.md)
- [pipeline status](../pipeline/status.md)
- [release Commands](../release/index.md)
