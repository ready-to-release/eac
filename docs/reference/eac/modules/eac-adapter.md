# eac-to-eac

The `eac-to-eac` module provides EAC command execution through both native and containerized adapters.

## System Context

Shows how the EAC adapter bridges CLI and container-based command execution.

<!-- structurizr:adapters:eac:SystemContext -->

## Container Architecture

High-level view of the EAC adapter packages.

<!-- structurizr:adapters:eac:Containers -->

## Design File

- **Location**: `specs/eac-to-eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module eac-to-eac`

## Key Responsibilities

| Package    | Purpose                                          |
| ---------- | ------------------------------------------------ |
| port       | EAC port interface (ExecConfig, Result)           |
| native     | Direct Go function call adapter                   |
| container  | Docker container-based command execution adapter  |
| factory    | Selects native or container adapter by environment |
| mock       | Test double for EAC command execution             |
