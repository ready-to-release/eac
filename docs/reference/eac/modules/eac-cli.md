# eac-cli

The `eac-cli` module provides all CLI command implementations for the EAC (Everything-as-Code) system.

It is the primary interface for developers interacting with the repository.

## System Context

Shows how eac-cli fits into the overall EAC ecosystem.

<!-- structurizr:eac-cli:SystemContext -->

## Container Architecture

High-level view of the major components within eac-cli.

<!-- structurizr:eac-cli:Containers -->

## Command Categories

### Development Commands

Commands for day-to-day development workflow.

<!-- structurizr:eac-cli:DevelopmentCommands -->

### Execution Commands

Commands for building, testing, and running modules.

<!-- structurizr:eac-cli:ExecutionCommands -->

### Query and Inspection

Commands for querying repository state and inspecting configurations.

<!-- structurizr:eac-cli:QueryAndInspection -->

### Release Commands

Commands for managing releases and changelogs.

<!-- structurizr:eac-cli:ReleaseCommands -->

### Security and Risk Commands

Commands for security scanning and risk assessment.

<!-- structurizr:eac-cli:SecurityAndRiskCommands -->

### Infrastructure Commands

Commands for CI/CD pipeline orchestration.

<!-- structurizr:eac-cli:InfrastructureCommands -->

## Component Architecture

### Registry Components

Command registration and discovery system.

<!-- structurizr:eac-cli:RegistryComponents -->

### Orchestrator Components

Build and test orchestration logic.

<!-- structurizr:eac-cli:OrchestratorComponents -->

### Query Components

Data retrieval and reporting components.

<!-- structurizr:eac-cli:QueryComponents -->

### Release Components

Release management and changelog generation.

<!-- structurizr:eac-cli:ReleaseComponents -->

### Commit Components

Git commit and workspace management.

<!-- structurizr:eac-cli:CommitComponents -->

### Design Components

Architecture diagram generation and validation.

<!-- structurizr:eac-cli:DesignComponents -->

### Render Components

Documentation rendering (books, reports).

<!-- structurizr:eac-cli:RenderComponents -->

### Test Components

Test execution and result processing.

<!-- structurizr:eac-cli:TestComponents -->

### Security Components

Security scanning integration.

<!-- structurizr:eac-cli:SecurityComponents -->

### Risk Components

OSCAL-based risk assessment.

<!-- structurizr:eac-cli:RiskComponents -->

## Design File

- **Location**: `specs/eac-cli/.design/workspace.dsl`
- **Interactive**: `r2r eac serve-design --module eac-cli`

## Key Dependencies

- `eac-core` - Core libraries and contracts
- Docker - For containerized operations
- External tools - PlantUML, Structurizr CLI, MkDocs
