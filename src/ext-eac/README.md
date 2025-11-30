# src/ext-eac

EAC Extension for R2R CLI - Repository management tooling wrapped as a Docker extension.

## Overview

The `eac` extension wraps the entire `src-commands` module suite (67 commands) into a containerized extension for R2R CLI. This enables portable, versioned access to repository management commands across any Docker-enabled environment.

## Features

- **67 Commands**: Full access to all src-commands functionality
- **Dynamic Discovery**: Extension metadata auto-generated from command registry
- **Multi-Platform**: Supports linux/amd64 and linux/arm64
- **Repository Context**: Operates on mounted repository via volume binding

## Command Categories

| Category    | Commands | Description                                                     |
| ----------- | -------- | --------------------------------------------------------------- |
| `build`     | 3        | Build modules and collect results                               |
| `design`    | 4        | Architecture diagrams using Structurizr DSL                     |
| `docs`      | 2        | MkDocs documentation server integration                         |
| `get`       | 9        | Structured data retrieval (modules, files, tests, dependencies) |
| `show`      | 10       | Human-readable data display                                     |
| `specs`     | 3        | Gherkin specification management with AI                        |
| `templates` | 7        | Template management for docs, specs, reports                    |
| `test`      | 6        | Module and suite testing                                        |
| `validate`  | 2        | Contract and dependency validation                              |
| `work`      | 7        | Git worktree management for parallel development                |
| Other       | 14       | Pipeline, release, completion, help, etc.                       |

## Usage

### Basic Command Execution

```bash
# List all modules
r2r run cmd show modules

# Test a specific module
r2r run cmd test module src-core

# Get dependency graph
r2r run cmd get dependencies

# Create specifications using AI
r2r run cmd specs create "User authentication flow"

# Build multiple modules
r2r run cmd build modules src-core src-commands
```

### Extension Metadata

```bash
# View extension metadata
r2r run cmd extension-meta

# Get help
r2r run cmd help

# Get help for specific command
r2r run cmd help show modules
```

### Volume Mounting Requirements

The extension requires repository context. Mount your repository root:

```bash
# Typical invocation (R2R CLI handles mounting automatically)
cd /path/to/repository
r2r run cmd show modules

# Manual Docker invocation (if needed)
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/ready-to-release/eac/extensions/eac show modules
```

### Docker-in-Docker Commands

Some commands require Docker access (e.g., `design serve`, `docs serve`). Mount the Docker socket:

```bash
# Commands that need Docker
r2r run cmd design serve
r2r run cmd docs serve

# Manual invocation with Docker socket
docker run --rm \
  -v $(pwd):/workspace \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -w /workspace \
  ghcr.io/ready-to-release/eac/extensions/eac design serve
```

## Extension Capabilities

- `repository-management` - Module and file discovery
- `module-discovery` - Find and inspect modules
- `dependency-analysis` - Analyze module dependencies
- `testing` - Run module and suite tests
- `documentation` - MkDocs integration
- `specifications` - AI-powered Gherkin specs
- `architecture-design` - Structurizr diagrams
- `git-worktree` - Parallel development workflows
- `build-automation` - Module build orchestration
- `pipeline-execution` - Multi-module pipelines

## Building the Extension

### From Repository Root

```bash
# Build multi-platform image
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f containers/ext-eac/Dockerfile \
  -t ghcr.io/ready-to-release/eac/extensions/eac:latest \
  .

# Build for local platform only
docker build \
  -f containers/ext-eac/Dockerfile \
  -t ext-eac:local \
  .
```

### Testing Locally

```bash
# Build local image
docker build -f containers/ext-eac/Dockerfile -t ext-eac:test .

# Test extension metadata
docker run --rm ext-eac:test extension-meta

# Test command execution
docker run --rm -v $(pwd):/workspace -w /workspace ext-eac:test show modules

# Test help
docker run --rm ext-eac:test help
```

## Technical Details

### Architecture

- **Entry Point**: `main.go` - Command router and metadata generator
- **Build**: Multi-stage Dockerfile (builder + runtime)
- **Base**: `golang:1.24-alpine` (build), `alpine:latest` (runtime)
- **Runtime Dependencies**: Git, Go, GitHub CLI (gh), Docker (socket mount), CA certificates, timezone data

### Command Dispatch

The extension uses direct pass-through to src-commands:

1. User invokes: `r2r run cmd show modules`
2. Extension receives: `["show", "modules"]`
3. Dispatches to: `src-commands` registry
4. Executes: Registered command function
5. Returns: Exit code to caller

### Metadata Generation

The `extension-meta` command dynamically generates YAML metadata by:

1. Querying `src-commands/registry`
2. Extracting command names, descriptions, and parameters
3. Marshaling to R2R CLI extension metadata format
4. Output: Complete, accurate command catalog

### Repository Context

Commands expect to run in repository root with access to:

- `contracts/` - Module contracts and boundaries
- `src/` - Source code modules
- `.git/` - Git repository for worktree commands
- `go.work` - Go workspace configuration

The extension sets `WORKDIR /workspace` and expects the repository to be mounted there.

## Dependencies

### Runtime Dependencies

The extension container includes:

- **Git**: Required for `work` commands, file tracking, and repository operations
- **Go**: Required for Go-based commands and tooling
- **GitHub CLI (gh)**: Required for GitHub operations (`work pr`, pipeline GitHub integration)
- **Docker**: Required for `design serve`, `docs serve` (via socket mount)
- **CA Certificates**: For HTTPS operations
- **Timezone Data**: For timestamp operations
- **curl**: For downloading additional tools during build

### External Integrations

- **AI Providers**: Some commands use AI (OpenAI, Anthropic) via API keys:

  - `specs create` - Generate Gherkin specifications
  - `design create` - Generate Structurizr diagrams
  - `commit` - Generate commit messages

  Pass API keys via environment variables:

  ```bash
  docker run --rm -v $(pwd):/workspace -w /workspace \
    -e OPENAI_API_KEY=$OPENAI_API_KEY \
    ext-eac specs create "Feature description"
  ```

## Module Contract

- **Moniker**: `ext-eac`
- **Type**: `go-r2r-extension`
- **Dependencies**: `src-commands`
- **Dockerfile**: `containers/ext-eac/Dockerfile`

## Labels

The container image includes comprehensive OCI and R2R CLI labels:

**OCI Labels**:

- `org.opencontainers.image.title`: "Command Extension"
- `org.opencontainers.image.description`: "Repository command tooling for R2R CLI"
- `org.opencontainers.image.version`: "1.0.0"
- `org.opencontainers.image.source`: Repository URL

**R2R CLI Labels**:

- `r2r-cli.extension.name`: "eac"
- `r2r-cli.extension.category`: "tools"
- `r2r-cli.capabilities`: Comma-separated capability list
- `r2r-cli.multi-arch`: "true"
- `r2r-cli.platforms`: "linux/amd64,linux/arm64"

## Examples

### Module Discovery

```bash
# List all modules
r2r run cmd show modules

# Get module details as JSON
r2r run cmd get modules --format json

# Show files owned by a module
r2r run cmd show files --module src-core
```

### Testing Workflows

```bash
# Test single module
r2r run cmd test module src-core

# Test multiple modules
r2r run cmd test modules src-core src-commands

# List test suites
r2r run cmd test list-suites

# Run specific test suite
r2r run cmd test suite integration
```

### Dependency Analysis

```bash
# Show dependency graph
r2r run cmd show dependencies

# Get execution order for modules
r2r run cmd get execution order src-commands

# Validate dependencies
r2r run cmd validate dependencies
```

### Git Worktree Management

```bash
# Create new workspace
r2r run cmd work create feature/new-feature

# List workspaces
r2r run cmd work list

# Commit with AI-generated message
r2r run cmd work commit

# Create PR with AI description
r2r run cmd work pr
```

### Specifications and Design

```bash
# Create Gherkin specification
r2r run cmd specs create "User can reset password via email"

# Validate specifications
r2r run cmd specs validate

# Generate architecture diagram
r2r run cmd design create src-core

# Serve diagrams in browser
r2r run cmd design serve
```

## Troubleshooting

### Command Not Found

If a command is not recognized:

1. Verify the command exists: `r2r run cmd help`
2. Check command syntax: `r2r run cmd help <command>`
3. Ensure repository is mounted correctly

### Permission Errors

If encountering permission errors:

1. Ensure volume mount is correct: `-v $(pwd):/workspace`
2. Check file permissions on host
3. Consider Docker user mapping if needed

### Docker-in-Docker Issues

If `design serve` or `docs serve` fail:

1. Verify Docker socket is mounted: `-v /var/run/docker.sock:/var/run/docker.sock`
2. Check Docker daemon is running on host
3. Ensure user has Docker socket permissions

### Repository Context Errors

If commands report missing files:

1. Verify you're in repository root: `pwd` should show repository root
2. Ensure working directory is set: `-w /workspace`
3. Check repository structure is intact

## Contributing

This extension is auto-generated from `src-commands`. To add new commands:

1. Add command implementation to `src/commands/impl/`
2. Follow src-commands conventions (see `src/commands/README.md`)
3. Rebuild extension image
4. Test with `docker run ext-eac:test extension-meta`

## License

MIT License - See repository LICENSE file

## Links

- **Repository**: https://github.com/ready-to-release/eac
- **src-commands**: `src/commands/` - Source command implementations
- **Dockerfile**: `containers/ext-eac/Dockerfile`
