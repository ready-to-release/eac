# eac-adapter

The `eac-adapter` module provides EAC command execution through both native and containerized adapters.

## System Context

Shows how the EAC adapter bridges CLI and container-based command execution.

<!-- structurizr:eac-adapter:SystemContext -->

## Container Architecture

High-level view of the EAC adapter packages.

<!-- structurizr:eac-adapter:Containers -->

## Design File

- **Location**: `specs/eac-adapter/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module eac-adapter`

## Key Responsibilities

| Package    | Purpose                                          |
| ---------- | ------------------------------------------------ |
| port       | EAC port interface (ExecConfig, Result)           |
| native     | Direct Go function call adapter                   |
| container  | Docker container-based command execution adapter  |
| factory    | Selects native or container adapter by environment |
| mock       | Test double for EAC command execution             |
