# eac-core

The `eac-core` module provides foundational libraries and contracts used across the EAC ecosystem. It defines the contract system, configuration loading, and shared utilities.

## System Context

Shows how eac-core provides foundational services to other modules.

<!-- structurizr:eac-core:SystemContext -->

## Container Architecture

High-level view of the core library components.

<!-- structurizr:eac-core:Containers -->

## Component Architecture

### Contracts Components

Contract loading, parsing, and validation.

<!-- structurizr:eac-core:ContractsComponents -->

### Repository Components

Repository discovery and file operations.

<!-- structurizr:eac-core:RepositoryComponents -->

### Git Components

Git operations and repository state management.

<!-- structurizr:eac-core:GitComponents -->

### Logging Components

Structured logging with Zap.

<!-- structurizr:eac-core:LoggingComponents -->

### Testing Components

Test utilities and shared test infrastructure.

<!-- structurizr:eac-core:TestingComponents -->

## Design File

- **Location**: `specs/eac-core/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module eac-core`

## Key Responsibilities

| Component | Purpose |
|-----------|---------|
| Config | Multi-layer configuration loading (repository.yml, books.yml, etc.) |
| Contracts | Module contract definitions and validation |
| Paths | Standardized path calculations for repository layout |
| Repository | Git repository discovery and state |
| Logging | Consistent structured logging across all commands |
