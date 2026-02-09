# eac

The `eac` module provides all CLI command implementations for the EAC (Everything-as-Code) system.

It is the primary interface for developers interacting with the repository.

## System Context

Shows how eac fits into the overall EAC ecosystem.

<!-- structurizr:eac:SystemContext -->

## Container Architecture

High-level view of the major components within eac.

<!-- structurizr:eac:Containers -->

## Command Categories

### Development Commands

Commands for day-to-day development workflow.

<!-- structurizr:eac:DevelopmentCommands -->

### Execution Commands

Commands for building, testing, and running modules.

<!-- structurizr:eac:ExecutionCommands -->

### Query and Inspection

Commands for querying repository state and inspecting configurations.

<!-- structurizr:eac:QueryAndInspection -->

### Release Commands

Commands for managing releases and changelogs.

<!-- structurizr:eac:ReleaseCommands -->

### Security and Risk Commands

Commands for security scanning and risk assessment.

<!-- structurizr:eac:SecurityAndRiskCommands -->

### Infrastructure Commands

Commands for CI/CD pipeline orchestration.

<!-- structurizr:eac:InfrastructureCommands -->

## Component Architecture

### Registry Components

Command registration and discovery system.

<!-- structurizr:eac:RegistryComponents -->

### Orchestrator Components

Build and test orchestration logic.

<!-- structurizr:eac:OrchestratorComponents -->

### Query Components

Data retrieval and reporting components.

<!-- structurizr:eac:QueryComponents -->

### Release Components

Release management and changelog generation.

<!-- structurizr:eac:ReleaseComponents -->

### Commit Components

Git commit and workspace management.

<!-- structurizr:eac:CommitComponents -->

### Design Components

Architecture diagram generation and validation.

<!-- structurizr:eac:DesignComponents -->

### Render Components

Documentation rendering (books, reports).

<!-- structurizr:eac:RenderComponents -->

### Test Components

Test execution and result processing.

<!-- structurizr:eac:TestComponents -->

### Security Components

Security scanning integration.

<!-- structurizr:eac:SecurityComponents -->

### Risk Components

OSCAL-based risk assessment.

<!-- structurizr:eac:RiskComponents -->

## Design File

- **Location**: `specs/eac/.design/workspace.dsl`
- **Interactive**: `eac serve-design --module eac`

## Key Dependencies

- `eac-core` - Core libraries and contracts
- Docker - For containerized operations
- External tools - PlantUML, Structurizr CLI, MkDocs
