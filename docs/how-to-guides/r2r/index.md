# R2R

Guides for working with the r2r (Ready to Release) command-line interface. Learn how to extend the CLI with custom commands packaged as Docker containers.

## In This Section

| Guide                                                              | Description                                                                                                           |
| ------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------- |
| [Creating Extensions](./creating-extensions.md)                    | Build production-ready extensions: project structure, metadata, testing, Dockerfile optimization, and EAC integration |
| [Testing in External Repositories](./testing-in-external-repos.md) | Test ext-eac and standalone extensions in external repositories using Docker                                          |

**See also**: [Local Development Workflows](../local-setup/local-dev-workflows.md)

Develop and test extensions locally (moved to Local Setup section)

### Quick Links

- **Creating your first extension?** Start with [Creating Extensions](./creating-extensions.md)
- **Developing ext-eac commands?** See [Local Development - EAC Repository](../local-setup/local-dev-workflows.md#eac-repository-development)
- **Building a standalone extension?** See [Local Development - External Repository Testing](../local-setup/local-dev-workflows.md#external-repository-testing)
- **Testing in another repo?** See [Testing in External Repositories](./testing-in-external-repos.md)

### Reference Implementation

For a complete, production-ready example, see [ext-env-check](https://github.com/ready-to-release/ext-env-check):

## See Also

- [R2R CLI Command Reference](../../reference/r2r/commands/index.md) - Complete command documentation
- [R2R CLI Configuration](../../reference/r2r/commands/configuration.md) - Configuration file reference
- [CLI vs Extensions Architecture](../../reference/eac/architecture/cli-integration.md) - Understanding the two-tier system

{{ diataxis_footer() }}
