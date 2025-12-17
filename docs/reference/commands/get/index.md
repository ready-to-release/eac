# get Commands
Retrieve repository data in structured JSON format for automation and scripting.

## Commands in this Category

| Command | Purpose |
|---------|---------|
| [get](./get.md) | Base get command |
| [get approval-comments](./approval-comments.md) | Get PR approval comments in structured format |
| [get artifacts](./artifacts.md) | Get resolved artifacts for a module |
| [get build-deps](./build-deps.md) | Get build dependencies for a module |
| [get build-times](./build-times.md) | Get build timing information |
| [get changed-modules](./changed-modules.md) | Get modules affected by changed files |
| [get changed-modules-ci](./changed-modules-ci.md) | Get modules requiring rebuild since last CI |
| [get changelog](./changelog.md) | Get changelog data in structured format |
| [get commands](./commands.md) | Retrieve repository data in structured format |
| [get config](./config.md) | Get all EAC configuration |
| [get dependencies](./dependencies.md) | Get module dependency graph |
| [get environments](./environments.md) | Get all environment contracts |
| [get execution-order](./execution-order.md) | Get execution order for specific modules |
| [get files](./files.md) | Get repository files with module ownership |
| [get modules](./modules.md) | Get all module contracts |
| [get release-notes](./release-notes.md) | Get release notes data in structured format |
| [get specs](./specs.md) | Get specifications data in structured format |
| [get specs-unused-steps](./specs-unused-steps.md) | Detect unused godog step definitions |
| [get suite](./suite.md) | Get test suite information |
| [get test-timings](./test-timings.md) | Get test timing information |
| [get tests](./tests.md) | Get all tests in the repository |
| [get valid-commands](./valid-commands.md) | Get all valid commands |

## Quick Examples

```bash
# Get modules as JSON
r2r eac get modules | jq '.modules[].moniker'

# Get changed modules for CI
r2r eac get changed-modules-ci
```

## See Also

- [Category Overview](../categories/get.md)
- [show Commands](../show/index.md) - Human-readable output
