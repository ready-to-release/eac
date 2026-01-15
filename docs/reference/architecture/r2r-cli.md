# r2r-cli

The `r2r-cli` module is the Ready-to-Release command-line interface for executing containerized development workflows. It manages Docker-based extensions that provide isolated, reproducible development environments.

## System Context

Shows how r2r-cli interacts with developers, Docker, and GitHub Container Registry.

<!-- structurizr:r2r-cli:SystemContext -->

## Container Architecture

High-level view of r2r-cli's major subsystems.

<!-- structurizr:r2r-cli:Containers -->

## Subsystem Architecture

### Configuration System

Multi-layer configuration management (base, local, personal, dev configs).

<!-- structurizr:r2r-cli:ConfigurationSystem -->

### Parser System

EBNF-based command-line argument parsing.

<!-- structurizr:r2r-cli:ParserSystem -->

### Docker Orchestration

Container lifecycle management and execution.

<!-- structurizr:r2r-cli:DockerOrchestration -->

### Extension Management

Extension installation and lifecycle management.

<!-- structurizr:r2r-cli:ExtensionManagement -->

### Validation System

Configuration and command validation.

<!-- structurizr:r2r-cli:ValidationSystem -->

### Logging System

Structured logging with context management.

<!-- structurizr:r2r-cli:LoggingSystem -->

### Terminal System

Cross-platform terminal handling (Unix/Windows).

<!-- structurizr:r2r-cli:TerminalSystem -->

### GitHub Integration

GitHub Container Registry integration for image management.

<!-- structurizr:r2r-cli:GitHubIntegration -->

### Cache System

Registry response and metadata caching.

<!-- structurizr:r2r-cli:CacheSystem -->

### Session System

CLI session state management.

<!-- structurizr:r2r-cli:SessionSystem -->

### TUI System

Terminal UI components (spinners, progress bars).

<!-- structurizr:r2r-cli:TUISystem -->

## Workflow Diagrams

### Run Workflow

The flow when executing `r2r run <extension> <command>`.

<!-- structurizr:r2r-cli:RunWorkflow -->

### Install Workflow

The flow when installing extensions.

<!-- structurizr:r2r-cli:InstallWorkflow -->

## Design File

- **Location**: `specs/r2r-cli/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module r2r-cli`

## Key Features

| Feature                 | Description                                             |
| ----------------------- | ------------------------------------------------------- |
| Containerized Execution | Run commands in isolated Docker containers              |
| Multi-Config Support    | Layer configs from repository, local, personal, and dev |
| Extension Pinning       | SHA-based version pinning for reproducibility           |
| TTY Support             | Full terminal support including interactive sessions    |
| Cross-Platform          | Works on Linux, macOS, and Windows                      |
