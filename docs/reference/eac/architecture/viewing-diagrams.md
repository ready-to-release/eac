# Viewing Architecture Diagrams

All modules include C4 architecture diagrams defined in Structurizr DSL format.

## Quick Start

```bash
# Start the Structurizr Lite server
eac serve-design

# Open browser to http://localhost:8080
```

The server starts a Docker container running Structurizr Lite, which renders all workspace.dsl files from the repository.

---

## Available Module Designs

Each module maintains its architecture diagrams in `specs/[module]/.design/workspace.dsl`:

| Module             | Design Location                                                                                                            | Description                                        |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------- |
| eac-cli            | [specs/eac-cli/.design/](https://github.com/ready-to-release/eac/tree/main/specs/eac-cli/.design/)                         | CLI command implementations                        |
| core               | [specs/core/.design/](https://github.com/ready-to-release/eac/tree/main/specs/core/.design/)                               | Core domain libraries (config, repository, git)    |
| clibase            | [specs/clibase/.design/](https://github.com/ready-to-release/eac/tree/main/specs/clibase/.design/)                         | CLI base framework (orchestrator, flags, render)   |
| eac-mcp-server     | [specs/eac-mcp-server/.design/](https://github.com/ready-to-release/eac/tree/main/specs/eac-mcp-server/.design/)           | MCP server for LLM tool integration                |
| clie-cli           | [specs/clie-cli/.design/](https://github.com/ready-to-release/eac/tree/main/specs/clie-cli/.design/)                       | Containerized workflow CLI                         |
| eac-ext            | [specs/eac-ext/.design/](https://github.com/ready-to-release/eac/tree/main/specs/eac-ext/.design/)                         | Docker extension container                         |
| ai-adapter         | [specs/ai-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/ai-adapter/.design/)                   | AI service integration adapter                     |
| docker-adapter     | [specs/docker-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/docker-adapter/.design/)           | Docker container runtime adapter                   |
| eac-adapter        | [specs/eac-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/eac-adapter/.design/)                 | EAC command execution adapter                      |
| tui-adapter        | [specs/tui-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/tui-adapter/.design/)                 | Terminal UI adapter (Bubbletea)                     |
| godog-adapter      | [specs/godog-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/godog-adapter/.design/)             | BDD test infrastructure (Godog)                    |
| cucumber-adapter   | [specs/cucumber-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/cucumber-adapter/.design/)       | Cucumber test runner adapter                       |
| gotest-adapter     | [specs/gotest-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/gotest-adapter/.design/)           | Go test runner adapter                             |
| mocha-adapter      | [specs/mocha-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/mocha-adapter/.design/)             | Mocha test runner adapter                          |
| npm-adapter        | [specs/npm-adapter/.design/](https://github.com/ready-to-release/eac/tree/main/specs/npm-adapter/.design/)                 | NPM dependency isolation adapter                   |
| docs               | [specs/docs/.design/](https://github.com/ready-to-release/eac/tree/main/specs/docs/.design/)                               | Documentation site architecture                    |
| contracts          | [specs/contracts/.design/](https://github.com/ready-to-release/eac/tree/main/specs/contracts/.design/)                     | Combined contract modules overview                 |

**Full list:** Check `specs/*/​.design/` directories in the repository.

---

## Diagram Levels (C4 Model)

Each module provides multiple diagram levels following the C4 model:

1. **System Context** - How the module fits in the overall system
2. **Container** - Components within the module
3. **Component** - Internal structure and relationships
4. **Code** - Implementation details (when applicable)

Not all modules include all levels - it depends on complexity.

---

## Working with Diagrams

### View All Modules

The `serve-design` command automatically loads all workspace.dsl files:

```bash
eac serve-design
```

Navigate between modules using the workspace dropdown in Structurizr Lite.

### View Specific Module

1. Run `eac serve-design`
2. Open `http://localhost:8080`
3. Select module from workspace dropdown (top-left)
4. Select diagram from the list

### Export Diagrams

In Structurizr Lite:

1. Click on a diagram
2. Use export controls to download as PNG or SVG
3. Diagrams maintain C4 model styling

---

## Structurizr DSL Files

All diagrams are defined in `workspace.dsl` files using Structurizr DSL.

**Format:** Plain text DSL (easy to version control)

**Example structure:**

```text
specs/
  eac-commands/
    .design/
      workspace.dsl          # Architecture definition
      .structurizr/          # Generated cache (not in git)
```

**Documentation:** [Structurizr DSL Language Reference](https://docs.structurizr.com/dsl/language)

---

## Updating Diagrams

### AI-Generated Updates

```bash
eac update-design <module-name>
```

Uses AI to analyze code and update the workspace.dsl file.

### Manual Editing

1. Edit `specs/<module>/.design/workspace.dsl`
2. Save changes
3. Refresh browser - Structurizr Lite auto-reloads

### Validate Syntax

```bash
eac validate-design
```

Checks all workspace.dsl files for syntax errors using Structurizr CLI.

---

## Creating New Diagrams

For new modules:

```bash
eac create-design <module-name>
```

Generates a workspace.dsl file based on code analysis.

---

## Tips

**Auto-refresh** - Structurizr Lite watches for file changes and reloads automatically

**Navigation** - Use the left sidebar to switch between workspaces (modules) and diagrams

**Zoom** - Use mouse wheel or diagram controls for zoom

**Themes** - Diagrams use consistent C4 model styling defined in workspace.dsl

**Docker Required** - The serve-design command requires Docker to run Structurizr Lite container

---

## Common Diagram Types

### eac-commands Module

- **System Context** - Commands in the CLIE/EAC ecosystem
- **Containers** - Command implementation structure
- **DevelopmentCommands** - AI-powered code generation commands
- **ExecutionCommands** - Build, test, and pipeline execution
- **InfrastructureCommands** - Security scanning and validation

### eac-core Module

- **System Context** - Core libraries role
- **Containers** - Package organization
- **Component** - Internal library structure

### clie-cli Module

- **System Context** - CLI in the broader ecosystem
- **Containers** - Application components
- **Component** - Command routing and execution

---

## Troubleshooting

**Server won't start:**

- Check Docker is running
- Check port 8080 is available
- Run `docker ps` to see if container is already running

**Diagrams not appearing:**

- Check workspace.dsl syntax with `eac validate-design`
- Look for error messages in browser console
- Verify .design folder contains workspace.dsl

**Changes not visible:**

- Hard refresh browser (Ctrl+F5 / Cmd+Shift+R)
- Restart serve-design command
- Check file was saved

---

## Related Commands

- [`serve-design`](../commands/serve/design.md) - Start Structurizr Lite server
- [`create-design`](../commands/create/design.md) - Generate architecture diagrams
- [`update-design`](../commands/update/design.md) - Update existing diagrams with AI
- [`validate-design`](../commands/validate/design.md) - Validate workspace.dsl syntax

---

## Related Documentation

- [Architecture Overview](./index.md) - System architecture overview
- [Modules](../modules/index.md) - Module system and organization
- [How-To: Generate Architecture Diagrams](../../../how-to-guides/eac/commands/documentation/generate-architecture-diagrams.md)
