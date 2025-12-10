# Architecture Documentation

{{ page_breadcrumb() }}

**Status:** Placeholder - Content coming soon

**Estimated time:** 60 minutes
**Prerequisites:** [Your First Module](../getting-started/first-module.md), understanding of C4 model

## Planned Content

This tutorial teaches you how to create and maintain architecture documentation using the C4 model and Structurizr, with AI-assisted diagram generation.

### What You'll Learn

- Understand the C4 model (Context, Container, Component, Code)
- Create `workspace.dsl` files for modules
- Generate architecture diagrams with AI: `r2r create design`
- Validate DSL syntax: `r2r validate design`
- View diagrams in browser: `r2r serve design`
- Update diagrams as architecture evolves: `r2r update design`
- Keep documentation synchronized with code

### Tutorial Structure

1. **C4 model fundamentals**
   - Context: System in environment
   - Container: Deployable units
   - Component: Groupings of code
   - Code: Class diagrams (optional)
   - Choosing the right level

2. **Structurizr DSL basics**
   - Workspace structure
   - Define systems, containers, components
   - Relationships and dependencies
   - Styling and layout

3. **Creating architecture documentation**
   - Generate with AI: `r2r create design <module>`
   - AI analyzes code and generates `workspace.dsl`
   - Review and refine generated documentation
   - Understand generated structure

4. **Validating architecture**
   - Check syntax: `r2r validate design`
   - Uses Structurizr CLI (Docker)
   - Fix validation errors
   - Ensure diagrams render

5. **Viewing diagrams**
   - Start server: `r2r serve design`
   - Open in browser (Structurizr Lite)
   - Navigate between views
   - Export diagrams (PNG, SVG)

6. **Updating documentation**
   - Code changes → Architecture changes
   - Update with AI: `r2r update design <module>`
   - Manual edits for precision
   - Version control diagrams

7. **Best practices**
   - Keep diagrams simple and focused
   - Update when architecture changes
   - Use consistent naming
   - Document key decisions
   - Link to decision records

### Example: Module Architecture

The tutorial documents the architecture of a module:

**Generated `workspace.dsl`:**
```dsl
workspace "EAC Commands" {

    model {
        user = person "Developer" "Uses r2r CLI"

        system = softwareSystem "EAC Commands" {
            cli = container "CLI" "Command-line interface" "Go" {
                cmd = component "Commands" "Cobra command handlers"
                contracts = component "Contracts" "Repository contracts"
                pipeline = component "Pipeline" "CI/CD orchestration"
            }
        }

        user -> cli "Executes commands"
        cmd -> contracts "Reads"
        cmd -> pipeline "Orchestrates"
    }

    views {
        systemContext system {
            include *
            autolayout lr
        }

        container system {
            include *
            autolayout lr
        }

        component cli {
            include *
            autolayout lr
        }
    }
}
```

### Key Concepts Covered

- C4 model levels
- Structurizr DSL syntax
- AI-assisted documentation generation
- Architecture validation
- Living documentation practices

### Creating Architecture Documentation

```bash
# Generate architecture from code
r2r create design eac-commands

# Review generated workspace.dsl
cat go/eac/commands/design/workspace.dsl

# Validate DSL syntax
r2r validate design

# View in browser
r2r serve design
# Opens http://localhost:8080

# Update after code changes
r2r update design eac-commands
```

### C4 Model Levels

**Level 1 - Context:**
- System and its users
- External dependencies
- High-level overview

**Level 2 - Container:**
- Deployable units (apps, services, databases)
- Communication protocols
- Technology choices

**Level 3 - Component:**
- Major code structures
- Responsibilities
- Interactions

**Level 4 - Code:**
- Class diagrams (rarely used)
- Generated from code

### Structurizr Lite

The tutorial uses Structurizr Lite (Docker) to view diagrams:

- Lightweight diagram viewer
- No account required
- Runs locally
- Export diagrams

### Documentation Workflow

1. **Initial creation**
   - Generate with AI: `r2r create design`
   - Review and refine
   - Commit to version control

2. **Ongoing maintenance**
   - Code changes → Update documentation
   - Use AI or manual edits
   - Validate before committing
   - Review in pull requests

3. **Consumption**
   - View in browser during development
   - Export for presentations
   - Link from README
   - Include in onboarding

### Best Practices

- Generate diagrams from code when possible
- Keep DSL files in module directories
- Version control all architecture documentation
- Update diagrams when architecture changes
- Use AI to bootstrap, refine manually
- Document key architectural decisions
- Link to decision records
- Keep diagrams simple and focused

### Integration with Development

- Architecture reviews in PRs
- Link diagrams in documentation
- Use in onboarding new team members
- Reference in technical discussions
- Export for presentations

### Tools Used

- **Structurizr CLI**: Validate DSL syntax (Docker)
- **Structurizr Lite**: View diagrams (Docker)
- **AI (Claude)**: Generate and update DSL
- **Git**: Version control diagrams

### Next Steps

After completing this tutorial, you'll maintain living architecture documentation. Explore other specialized topics: [Effective BDD Scenarios](./effective-bdd-scenarios.md), [Security Scanning](./security-scanning-workflow.md), or [TypeScript Setup](./typescript-module-setup.md).

{{ diataxis_footer() }}
