# eac-commands

The `eac-commands` module provides all CLI command implementations for the EAC (Everything-as-Code) system.

It is the primary interface for developers interacting with the repository.

## System Context

Shows how eac-commands fits into the overall EAC ecosystem.

<!-- structurizr:eac-commands:SystemContext -->

## Container Architecture

High-level view of the major components within eac-commands.

<!-- structurizr:eac-commands:Containers -->

## Command Categories

### Development Commands

Commands for day-to-day development workflow.

<!-- structurizr:eac-commands:DevelopmentCommands -->

### Execution Commands

Commands for building, testing, and running modules.

<!-- structurizr:eac-commands:ExecutionCommands -->

### Query and Inspection

Commands for querying repository state and inspecting configurations.

<!-- structurizr:eac-commands:QueryAndInspection -->

### Release Commands

Commands for managing releases and changelogs.

<!-- structurizr:eac-commands:ReleaseCommands -->

### Security and Risk Commands

Commands for security scanning and risk assessment.

<!-- structurizr:eac-commands:SecurityAndRiskCommands -->

### Infrastructure Commands

Commands for CI/CD pipeline orchestration.

<!-- structurizr:eac-commands:InfrastructureCommands -->

## Component Architecture

### Registry Components

Command registration and discovery system.

<!-- structurizr:eac-commands:RegistryComponents -->

### Orchestrator Components

Build and test orchestration logic.

<!-- structurizr:eac-commands:OrchestratorComponents -->

### Query Components

Data retrieval and reporting components.

<!-- structurizr:eac-commands:QueryComponents -->

### Release Components

Release management and changelog generation.

<!-- structurizr:eac-commands:ReleaseComponents -->

### Commit Components

Git commit and workspace management.

<!-- structurizr:eac-commands:CommitComponents -->

### Design Components

Architecture diagram generation and validation.

<!-- structurizr:eac-commands:DesignComponents -->

### Render Components

Documentation rendering (books, reports).

<!-- structurizr:eac-commands:RenderComponents -->

### Test Components

Test execution and result processing.

<!-- structurizr:eac-commands:TestComponents -->

### Security Components

Security scanning integration.

<!-- structurizr:eac-commands:SecurityComponents -->

### Risk Components

OSCAL-based risk assessment.

<!-- structurizr:eac-commands:RiskComponents -->

## Design File

- **Location**: `specs/eac-commands/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module eac-commands`

## Key Dependencies

- `eac-core` - Core libraries and contracts
- Docker - For containerized operations
- External tools - PlantUML, Structurizr CLI, MkDocs
