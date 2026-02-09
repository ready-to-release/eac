# clie

The `clie` module is the Ready-to-Release command-line interface for executing containerized development workflows.

It manages Docker-based extensions that provide isolated, reproducible development environments.

## System Context

Shows how clie interacts with developers, Docker, and GitHub Container Registry.

<!-- structurizr:clie:SystemContext -->

## Container Architecture

High-level view of clie's major subsystems.

<!-- structurizr:clie:Containers -->

## Subsystem Architecture

### Configuration System

Multi-layer configuration management (base, local, personal, dev configs).

<!-- structurizr:clie:ConfigurationSystem -->

### Parser System

EBNF-based command-line argument parsing.

<!-- structurizr:clie:ParserSystem -->

### Docker Orchestration

Container lifecycle management and execution.

<!-- structurizr:clie:DockerOrchestration -->

### Extension Management

Extension installation and lifecycle management.

<!-- structurizr:clie:ExtensionManagement -->

### Validation System

Configuration and command validation.

<!-- structurizr:clie:ValidationSystem -->

### Logging System

Structured logging with context management.

<!-- structurizr:clie:LoggingSystem -->

### Terminal System

Cross-platform terminal handling (Unix/Windows).

<!-- structurizr:clie:TerminalSystem -->

### GitHub Integration

GitHub Container Registry integration for image management.

<!-- structurizr:clie:GitHubIntegration -->

### Cache System

Registry response and metadata caching.

<!-- structurizr:clie:CacheSystem -->

### Session System

CLI session state management.

<!-- structurizr:clie:SessionSystem -->

### TUI System

Terminal UI components (spinners, progress bars).

<!-- structurizr:clie:TUISystem -->

## Workflow Diagrams

### Run Workflow

The flow when executing `clie run <extension> <command>`.

<!-- structurizr:clie:RunWorkflow -->

### Install Workflow

The flow when installing extensions.

<!-- structurizr:clie:InstallWorkflow -->

## Design File

- **Location**: `specs/clie/.design/workspace.dsl`
- **Interactive**: `clie eac serve-design --module clie`

## Key Features

| Feature                 | Description                                             |
| ----------------------- | ------------------------------------------------------- |
| Containerized Execution | Run commands in isolated Docker containers              |
| Multi-Config Support    | Layer configs from repository, local, personal, and dev |
| Extension Pinning       | SHA-based version pinning for reproducibility           |
| TTY Support             | Full terminal support including interactive sessions    |
| Cross-Platform          | Works on Linux, macOS, and Windows                      |
