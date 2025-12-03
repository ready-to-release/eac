# Architecture Design

Architecture design in EAC uses the C4 model and Structurizr DSL to create living, diagram-as-code documentation that stays synchronized with your codebase.

## What is Architecture Design?

EAC's design system enables you to:

- **Generate architecture diagrams** using AI analysis of your code
- **Maintain diagrams as code** using Structurizr DSL
- **Visualize system structure** at multiple abstraction levels
- **Keep documentation current** through automated updates

The system uses AI to analyze module contracts, dependencies, and code structure to generate accurate C4 model diagrams.

## When to Use Design Commands

Use design commands when you need:

| Scenario                           | Commands          |
| ---------------------------------- | ----------------- |
| Creating initial architecture docs | `create-design`   |
| Updating after structural changes  | `update-design`   |
| Validating DSL syntax              | `validate-design` |
| Viewing diagrams in browser        | `serve-design`    |

### Common Use Cases

- **Onboarding** - Help new team members understand system structure
- **Architecture reviews** - Visualize proposed changes
- **Documentation** - Maintain up-to-date system documentation
- **Compliance** - Provide architecture artifacts for audits
- **Technical decisions** - Document architecture decision records

## Key Concepts

### C4 Model

The C4 model provides four levels of abstraction:

| Level         | Description                      | Audience               |
| ------------- | -------------------------------- | ---------------------- |
| **Context**   | System and external actors       | Business stakeholders  |
| **Container** | High-level technology choices    | Technical stakeholders |
| **Component** | Internal structure of containers | Developers             |
| **Code**      | Class/module level details       | Developers             |

EAC generates Container and Component level diagrams from module contracts.

### Structurizr DSL

Diagrams are defined using Structurizr DSL, a text-based format:

```dsl
workspace "System" {
    model {
        user = person "User"
        system = softwareSystem "My System" {
            api = container "API" {
                technology "Go"
            }
            db = container "Database" {
                technology "PostgreSQL"
            }
        }
        user -> api "Uses"
        api -> db "Reads/Writes"
    }
    views {
        container system {
            include *
            autoLayout
        }
    }
}
```

### Workspace Files

Each module can have a `workspace.dsl` file:

```text
go/eac/commands/.design/workspace.dsl
go/eac/core/.design/workspace.dsl
```

### AI Generation

When generating diagrams, AI analyzes:

1. **Module contracts** - Dependencies and relationships
2. **Code structure** - Packages and interfaces
3. **File organization** - Directory structure
4. **Existing documentation** - README files and comments

## Workflow Overview

### Initial Diagram Creation

```bash
# 1. Generate workspace.dsl for a module
r2r eac create-design eac-commands

# 2. Review generated DSL
cat go/eac/commands/.design/workspace.dsl

# 3. Validate syntax
r2r eac validate-design eac-commands

# 4. View in browser
r2r eac serve-design eac-commands
```

### Updating After Changes

```bash
# 1. After code changes, update the design
r2r eac update-design eac-commands

# 2. Review changes
git diff go/eac/commands/.design/workspace.dsl

# 3. Validate and view
r2r eac validate-design eac-commands
r2r eac serve-design eac-commands
```

### Team Collaboration

```bash
# 1. Create designs for all modules
r2r eac create-design --all

# 2. Commit to version control
git add **/.design/workspace.dsl
git commit -m "docs: add architecture diagrams"

# 3. CI validates on every PR
# In GitHub Actions:
r2r eac validate-design
```

## Diagram Types

### System Context

Shows how the system fits in its environment:

- External users and systems
- System boundaries
- High-level interactions

### Container Diagram

Shows the high-level technology architecture:

- Applications and services
- Data stores
- Communication protocols

### Component Diagram

Shows internal structure of a container:

- Modules and packages
- Internal dependencies
- Interfaces and APIs

## Integration Points

### With Module Contracts

Design generation reads from `modules.yml`:

- Module dependencies become relationships
- Module types inform container technology
- File patterns define component boundaries

### With Documentation

Generated diagrams integrate with MkDocs:

```markdown
## Architecture

![Container Diagram](./diagrams/container.png)

See [workspace.dsl](./.design/workspace.dsl) for source.
```

### With CI/CD

Validate designs in pipelines:

```yaml
- name: Validate architecture
  run: r2r eac validate-design

- name: Check for drift
  run: |
    r2r eac update-design --dry-run
    git diff --exit-code
```

## Visualization

### Structurizr Lite

`serve-design` launches Structurizr Lite in Docker:

```bash
r2r eac serve-design eac-commands
# Opens http://localhost:8080
```

Features:

- Interactive diagram exploration
- Multiple view types
- Export to PNG/SVG
- Documentation rendering

### Export Options

From Structurizr Lite, export:

- **PNG/SVG** - For documentation
- **PlantUML** - For integration with other tools
- **Mermaid** - For markdown rendering

## Next Steps

- [Design Configuration](design-configuration.md) - Configure DSL templates and themes
- [Design Commands](design-commands.md) - Full command reference

## Related Areas

- [Specifications](specifications-overview.md) - BDD specs that complement architecture docs
- [Templates](templates-overview.md) - Documentation templates including diagrams
- [Books](books-overview.md) - Aggregate designs into documentation books
