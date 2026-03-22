# deploy

<!-- book:cmd deploy -->

Deploys a module to a target environment using the configured deployer for each deployable component in the module.

## Usage

```bash
eac deploy <module> <environment> [flags]
```

Both arguments are required:

- `module` - Module moniker to deploy
- `environment` - Target environment moniker (must match an entry in `environments.yml`)

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--dry-run` | bool | Preview changes without applying (maps to `az what-if`) |
| `--component <name>` | string | Only deploy a specific component within the module |
| `--skip-deps` | bool | Skip system dependency verification |
| `--debug` | bool | Enable debug logs to console |
| `--tui` | bool | Enable TUI console |
| `--no-tui` | bool | Disable TUI console |

## Examples

```bash
# Deploy infra to development
eac deploy infra development

# Preview production changes without applying
eac deploy infra production --dry-run

# Deploy only a specific component
eac deploy infra development --component networking
```

## See Also

- [build](./build/build.md) - Build modules
