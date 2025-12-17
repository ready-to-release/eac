# Configuration and Setup
Learn how to configure EAC for your project and understand the contract system.

## Coming Soon

This section will provide comprehensive guides on:

### Getting Started with EAC

- **Initial Project Setup** - Setting up EAC in a new repository
- **Repository Structure** - Understanding the expected directory layout
- **First Module Creation** - Getting your first module running
- **Verifying Installation** - Ensuring everything is configured correctly

### Understanding Contracts

- **Module Contracts (.eac/repository.yml)** - Defining modules and their configurations
  - Module properties and metadata
  - Build pipeline configuration
  - Artifact definitions
  - File ownership patterns
  - **Language-Specific Module Types**:
    - `go` - Go modules with go.mod, cross-compilation, Go test support
    - `typescript` - TypeScript/npm modules with package.json, tsc builds
    - `container` - Docker builds via Dockerfile (any language)
    - `docs` - MkDocs documentation generation
    - `static` - File ownership only, no builds

- **Environment Contracts (.eac/environments.yml)** - Configuring deployment environments
  - Environment definitions
  - Environment-specific settings
  - Variable management

- **Books Configuration (books.yml)** - Organizing documentation
  - Book structure and navigation
  - Chapter organization
  - Content sources

- **Tag Contracts** - Valid tags for specifications and tests
  - Test tags
  - Control tags (compliance mapping)
  - Custom tag definitions

### MCP Server Setup

- **Configuring MCP Servers** - Setting up .mcp.json
  - Commands server configuration
  - GitHub server configuration
  - Authentication and tokens
  - Troubleshooting MCP connections

### AI Provider Configuration

- **Setting Up AI Integration** - Configuring AI providers for commands
  - Running `init` command
  - Supported AI providers (Anthropic, OpenAI)
  - API key management
  - Model selection

### Advanced Configuration

- **Custom Module Types** - Creating reusable module configurations
- **Pipeline Customization** - Tailoring build and test workflows
- **Validation Rules** - Customizing contract validation
- **Template Configuration** - Setting up project templates

## Related Guides

While this section is being developed, see:

- [Setup AI Provider](../commands/getting-started/setup-ai-provider.md) - Quick AI configuration
- [Creating Modules](../modules/creating-modules.md) - Module creation guide
- [Creating Module Types](../modules/creating-module-types.md) - Module type definitions

## Quick Start

For immediate setup needs:

```bash
# Initialize AI provider
go run ./go/eac/commands init

# View current configuration
go run ./go/eac/commands show config

# View all modules
go run ./go/eac/commands show modules
```
