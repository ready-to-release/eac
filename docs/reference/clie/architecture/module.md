# clie-cli

The `clie-cli` module is the Ready-to-Release command-line interface for executing containerized development workflows.

It manages Docker-based extensions that provide isolated, reproducible development environments.

## System Context

Shows how clie-cli interacts with developers, Docker, and GitHub Container Registry.

<!-- structurizr:clie-cli:SystemContext -->

## Container Architecture

High-level view of clie-cli's major subsystems.

<!-- structurizr:clie-cli:Containers -->

## Subsystem Architecture

### Configuration System

Multi-layer configuration management (base, local, personal, dev configs).

<!-- structurizr:clie-cli:ConfigurationSystem -->

### Parser System

EBNF-based command-line argument parsing.

<!-- structurizr:clie-cli:ParserSystem -->

### Docker Orchestration

Container lifecycle management and execution.

<!-- structurizr:clie-cli:DockerOrchestration -->

### Extension Management

Extension installation and lifecycle management.

<!-- structurizr:clie-cli:ExtensionManagement -->

### Validation System

Configuration and command validation.

<!-- structurizr:clie-cli:ValidationSystem -->

### Logging System

Structured logging with context management.

<!-- structurizr:clie-cli:LoggingSystem -->

### Terminal System

Cross-platform terminal handling (Unix/Windows).

<!-- structurizr:clie-cli:TerminalSystem -->

### GitHub Integration

GitHub Container Registry integration for image management.

<!-- structurizr:clie-cli:GitHubIntegration -->

### Cache System

Registry response and metadata caching.

<!-- structurizr:clie-cli:CacheSystem -->

### Session System

CLI session state management.

<!-- structurizr:clie-cli:SessionSystem -->

### TUI System

Terminal UI components (spinners, progress bars).

<!-- structurizr:clie-cli:TUISystem -->

## Workflow Diagrams

### Run Workflow

The flow when executing `clie run <extension> <command>`.

<!-- structurizr:clie-cli:RunWorkflow -->

### Install Workflow

The flow when installing extensions.

<!-- structurizr:clie-cli:InstallWorkflow -->

## Design File

- **Location**: `specs/clie-cli/.design/workspace.dsl`
- **Interactive**: `clie eac serve-design --module clie-cli`

## Key Features

| Feature                 | Description                                             |
| ----------------------- | ------------------------------------------------------- |
| Containerized Execution | Run commands in isolated Docker containers              |
| Multi-Config Support    | Layer configs from repository, local, personal, and dev |
| Extension Pinning       | SHA-based version pinning for reproducibility           |
| TTY Support             | Full terminal support including interactive sessions    |
| Cross-Platform          | Works on Linux, macOS, and Windows                      |
