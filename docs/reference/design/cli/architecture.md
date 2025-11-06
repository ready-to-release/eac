# R2R CLI Architecture

**Version**: 1.0.0
**Date**: 2025-11-06
**Status**: Current

## Overview

The R2R (Ready to Release) CLI is an enterprise-grade automation framework that standardizes and containerizes development workflows. The architecture follows a modular design with clear separation of concerns between command handling, configuration management, Docker orchestration, and extension management.

## Architecture Model

This architecture is modeled using Structurizr DSL and can be visualized using [Structurizr Lite](https://github.com/structurizr/lite).

**DSL File**: [`workspace.dsl`](workspace.dsl)

### Viewing the Architecture

**Start Structurizr Lite:**

```bash
cd docs/reference/design/cli
docker run -d --name structurizr-cli -p 8081:8080 \
  -v "$(pwd):/usr/local/structurizr" \
  structurizr/lite
```

Then open http://localhost:8081 in your browser.

**Stop when done:**
```bash
docker stop structurizr-cli && docker rm structurizr-cli
```

**Reload after editing workspace.dsl:**
```bash
docker restart structurizr-cli
```

#### What You'll See

The interactive diagrams show:
- **System Context**: How the CLI fits into the broader system
- **Container Diagram**: The 9 main containers and their relationships

**Note**: On Windows with Git Bash, if the volume mount doesn't work, use `docker cp` instead:
```bash
docker run -d --name structurizr-cli -p 8081:8080 structurizr/lite
sleep 5
docker cp workspace.dsl structurizr-cli:/usr/local/structurizr/workspace.dsl
docker restart structurizr-cli
```

## Containers

### 1. CLI Application
**Technology**: Go (Cobra)
**Purpose**: Command-line interface with argument parsing and command routing

The main entry point that:
- Parses command-line arguments using the Command Parser
- Routes commands to appropriate handlers
- Coordinates operations across all containers
- Manages the overall execution flow

**Location**: `src/cli/main.go`, `src/cli/cmd/root.go`

### 2. Configuration Manager
**Technology**: Go (Viper)
**Purpose**: Loads, validates, and merges YAML configuration files with extension definitions

Responsible for:
- Loading base and override configuration files
- Validating configuration schema
- Checking for pinned extension versions (CI enforcement)
- Merging configuration layers
- Providing configuration to other containers

**Location**: `src/cli/internal/conf/`

### 3. Docker Orchestrator
**Technology**: Go (Docker SDK)
**Purpose**: Manages Docker container lifecycle, volumes, and networking

Handles:
- Container creation and execution
- Volume mounting (workspace, caches)
- Network configuration
- Container cleanup
- Resource limits (CPU, memory)
- Progress tracking and output streaming

**Location**: `src/cli/internal/docker/`

### 4. Extension Manager
**Technology**: Go
**Purpose**: Installs and manages CLI extensions from registries

Manages:
- Extension installation from GitHub Container Registry
- Extension metadata retrieval
- Local extension development (load_local mode)
- Extension versioning and updates

**Location**: `src/cli/internal/extensions/`

### 5. GitHub Registry Client
**Technology**: Go (HTTP)
**Purpose**: Fetches extension metadata and versions from GitHub Container Registry

Provides:
- GHCR API integration
- Tag listing and latest version detection
- Authentication for private registries
- Extension metadata fetching

**Location**: `src/cli/internal/github/`

### 6. Validation Engine
**Technology**: Go
**Purpose**: Validates configuration schema, extension metadata, and command syntax

Validates:
- YAML configuration structure
- Extension definitions (required fields, formats)
- Docker image references
- Environment variable names
- Resource limits
- Port and volume mappings

**Location**: `src/cli/internal/validator/`

### 7. Logger
**Technology**: Go (Zerolog)
**Purpose**: Structured logging with context and multiple output formats

Features:
- Structured JSON logging
- Log levels (debug, info, warn, error, fatal)
- Context propagation (command, operation ID)
- Console and file output
- Field-based metadata

**Location**: `src/cli/internal/logger/`

### 8. Registry Cache
**Technology**: Go (JSON file)
**Purpose**: Caches extension tags and metadata to reduce registry API calls

Caches:
- Extension version information
- Available tags
- Latest stable versions (sha- tags)
- Timestamp-based expiration (configurable TTL)

**Location**: `src/cli/internal/cache/`

### 9. Command Parser
**Technology**: Go
**Purpose**: Parses command-line arguments and detects argument boundaries for container execution

Handles:
- Subcommand detection
- Extension name extraction
- Viper flags vs container arguments separation
- Bash redirect pollution filtering
- Argument boundary detection

**Location**: `src/cli/internal/command-parser/`

## Key Relationships

### CLI Application Dependencies
- **Command Parser**: Parses command-line arguments before routing
- **Configuration Manager**: Loads configuration for command execution
- **Logger**: Logs all commands and errors
- **Docker Orchestrator**: Executes extensions in containers
- **Extension Manager**: Installs and manages extensions

### Configuration Manager Dependencies
- **Validation Engine**: Validates configuration schema and values
- **Registry Cache**: Checks for cached extension versions
- **GitHub Registry Client**: Fetches latest extension versions for pinning validation

### Extension Manager Dependencies
- **GitHub Registry Client**: Downloads extension container images
- **Registry Cache**: Updates cache with fetched metadata

### Supporting Container Dependencies
- **Docker Orchestrator → Logger**: Logs all container operations
- **GitHub Registry Client → Logger**: Logs all registry API calls

## Design Decisions

### 1. Container-Based Extension System
Extensions run in Docker containers to ensure:
- Isolation and security
- Consistent execution environment
- No local toolchain dependencies
- Portable across platforms

### 2. Layered Configuration
Configuration supports:
- Base configuration (`r2r.yaml`)
- Override files for environments
- Validation at load time
- Pinned versions in CI (enforced)

### 3. Registry Caching
Extension metadata is cached to:
- Reduce GHCR API calls
- Improve performance
- Enable offline operation (within TTL)
- Configurable cache duration (default: 5 minutes)

### 4. Structured Logging
All components use structured logging for:
- Debugging complex workflows
- Audit trails
- Integration with log aggregation systems
- Context propagation across operations

### 5. Command Parsing Strategy
Two-phase parsing approach:
- Custom parser for argument boundary detection
- Cobra for flag parsing and command routing
- Necessary for distinguishing CLI flags from container arguments

## Extension Points

### Adding New Commands
1. Create command file in `src/cli/cmd/`
2. Register in `cmd/root.go`
3. Use Configuration Manager for config access
4. Use Logger for structured logging

### Adding New Extension Sources
1. Implement registry client interface
2. Add to Extension Manager
3. Update configuration schema
4. Add validation rules

### Custom Validators
1. Add validation functions to Validation Engine
2. Register in configuration validation pipeline
3. Add tests in `validator/` package

## Testing Strategy

- **Unit Tests**: Test individual containers in isolation
- **Integration Tests**: Test container interactions
- **E2E Tests**: Test complete CLI workflows with Docker

**Test Locations**:
- `src/cli/internal/*/` - Unit tests co-located with code
- `src/cli/cmd/*_test.go` - Command handler tests

## Related Documentation

- [Structurizr Lite MCP Server Design](../structurizr-lite-mcp-design.md)
- [Configuration Schema](../../../explanation/cli/configuration.md) _(if exists)_
- [Extension Development Guide](../../../how-to-guides/cli/create-extension.md) _(if exists)_

## Maintenance

This architecture document should be updated when:
- New containers are added to the system
- Container responsibilities change significantly
- Major architectural decisions are made
- Key relationships between containers change

### Updating the Architecture

**Edit the DSL file directly:**
```bash
# Edit the architecture model
vi docs/reference/design/cli/workspace.dsl

# Reload in Structurizr Lite (if running)
docker restart structurizr-cli
```

**Or use the Structurizr Lite MCP Server for programmatic updates:**
```bash
# Start MCP server (in separate terminal)
cd src/mcp/structurizr-lite
make run

# Use JSON-RPC API to add/modify containers and relationships
# (See ../structurizr-lite-mcp-design.md for API documentation)
```

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-11-06 | Initial architecture documentation using Structurizr Lite MCP Server |
