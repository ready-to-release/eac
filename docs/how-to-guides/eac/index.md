# EAC Guides

Guides for the Everything as Code (EAC) system.

Learn how to use CLI commands, write BDD specifications, configure integrations, and manage modules.

## In This Section

| Guide                   | Description                                               |
| ----------------------- | --------------------------------------------------------- |
| [Commands](./commands/index.md) | CLI commands for build, test, validate, release, and more |
| [Modules](./modules/index.md)   | Creating and configuring modules                          |

## Language Support

EAC currently provides native support for **Go** and **TypeScript** projects:

- **Go modules** - Full build, test, and cross-compilation support
- **TypeScript/npm** - Build and test integration
- **Docker containers** - Multi-platform builds for any language
- **Documentation** - MkDocs site generation

Other languages can use container-based or script-based builds.

See [R2R and EAC Reference](../../reference/eac/index.md#language-support) for details.

## Getting Started

New to EAC? Start here:

1. **[Commands](./commands/index.md)** - Learn the essential commands
2. **[Modules](./modules/index.md)** - Create your first module

## Common Tasks

### Development Workflow

- [Create Feature Workspace](./commands/development-workflow/create-feature-workspace.md)
- [Build and Test](./commands/build-test-validate/index.md)
- [Validate Before Commit](./commands/build-test-validate/validate-before-commit.md)
- [Make Commits with AI](./commands/development-workflow/make-commits-with-ai.md)

### Quality Assurance

- [Run Tests](./commands/build-test-validate/run-tests-for-module.md)
- [Scan for Security Issues](./commands/build-test-validate/scan-for-security-issues.md)
- [Validate Specifications](./commands/build-test-validate/validate-specifications.md)

### Release Management

- [Prepare Release](./commands/release-management/prepare-module-release.md)
- [Generate Changelog](./commands/release-management/generate-changelog.md)
- [Create Release Tag](./commands/release-management/create-release-tag.md)

## Need More Detail?

These how-to guides focus on accomplishing tasks. For comprehensive technical details:

- **[Command Reference](../../reference/eac/commands/index.md)** - Complete command syntax and options
- **[R2R and EAC Reference](../../reference/eac/index.md)** - Architecture, contracts, and extending EAC
