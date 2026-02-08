# System Context

High-level architecture showing how CLIE CLI interacts with external systems.

## System Context Diagram

Shows how CLIE CLI interacts with developers, CI/CD systems, Docker Engine, and GitHub Container Registry.

<!-- structurizr:clie-cli:SystemContext -->

**Key actors:**

- **Developer** - Uses CLIE CLI to execute containerized development workflows
- **CI/CD System** - Automated build and test pipelines

**External systems:**

- **Docker Engine** - Container runtime and orchestration
- **GitHub Container Registry** - Docker image registry for extensions
- **Git Repository** - Version control system

## Container Architecture

High-level view of CLIE CLI's major subsystems and how they interact.

<!-- structurizr:clie-cli:Containers -->

**Major subsystems:**

| Subsystem            | Purpose                                   |
| -------------------- | ----------------------------------------- |
| Command Layer        | Cobra-based command routing and execution |
| Configuration System | Multi-layer configuration management      |
| Command Parser       | EBNF-based argument parsing               |
| Docker Orchestration | Container lifecycle management            |
| Extension Management | Extension installation and lifecycle      |
| Validation System    | Configuration and command validation      |
| Logging System       | Structured logging with context           |
| Terminal System      | Cross-platform terminal handling          |
| GitHub Integration   | Container registry integration            |
| Cache System         | Registry response caching                 |
| Session System       | CLI session state management              |
| TUI System           | Terminal UI components                    |

## See Also

- [Subsystems](./subsystems.md) - Detailed component architecture for each subsystem
- [Workflows](./workflows.md) - Dynamic workflow diagrams
