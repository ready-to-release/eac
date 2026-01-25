# Generate Architecture Diagrams

## What You'll Accomplish

Create C4 model architecture diagrams using AI to analyze your code and generate Structurizr workspace files.

## Prerequisites

- Module to document exists
- AI provider configured
- Docker installed (for viewing diagrams)

## Steps

### 1. Generate Diagram for Module

```bash
r2r eac create design eac-core
```

**What happens**: AI analyzes module code and generates `workspace.dsl` file

### 2. Review Generated Diagram

```bash
cat specs/eac-core/.design/workspace.dsl
```

**What happens**: View Structurizr DSL definition

### 3. Visualize Diagram

```bash
r2r eac serve design eac-core
```

**What happens**: Starts Structurizr Lite server, opens in browser

### 4. Edit and Update

Edit `workspace.dsl` manually if needed, then update:

```bash
r2r eac update design eac-core
```

**What happens**: AI updates existing diagram with new information

## Diagram Types

C4 model levels generated:

- **System Context** - System and external dependencies
- **Container** - Major components/containers
- **Component** - Internal structure
- **Code** - Class/package details

## Example Scenario

Documenting core library module:

```bash
# Generate diagram
r2r eac create design eac-core

# Output:
# Analyzing eac-core module...
# Generating architecture diagram...
# ✓ Created specs/eac-core/.design/workspace.dsl

# View in browser
r2r eac serve design eac-core
# Starting Structurizr Lite on http://localhost:8080
# Open browser to view diagrams

# Make code changes, update diagram
r2r eac update design eac-core
# ✓ Updated specs/eac-core/.design/workspace.dsl with latest changes

# Validate syntax
r2r eac validate design eac-core
# ✓ workspace.dsl syntax valid
```

## Diagram Structure

Generated workspace.dsl includes:

```dsl
workspace "EAC Core Library" "Core domain libraries..." {
  model {
    # External systems
    filesystem = softwareSystem "File System" "Repository files..." "External"
    git_system = softwareSystem "Git" "Version control..." "External"

    # Your system
    eac_core = softwareSystem "EAC Core Library" "Foundational Go library..." {
      contracts = container "Contracts" "Module contract definitions..." "Go Package" {
        contract_types = component "Contract Types" "Module, Environment types..." "Go"
        contract_loader = component "Contract Loader" "Loads contracts from YAML..." "Go"
      }
      repository = container "Repository" "Repository operations..." "Go Package"
    }
  }

  views {
    systemContext eac_core
    container eac_core
    component contracts
  }
}
```

## Common Issues

| Problem              | Solution                              |
| -------------------- | ------------------------------------- |
| "AI analysis failed" | Ensure code is readable, has comments |
| Diagram incomplete   | Use `update design` to enhance        |
| Docker not found     | Install Docker for Structurizr Lite   |

## Next Steps

- [Build Documentation Site](./build-documentation-site.md) → Include diagrams

## Related Commands

- [`create design`](../../../../reference/eac/commands/create/design.md) - Generate diagram
- [`update design`](../../../../reference/eac/commands/update/design.md) - Update existing
- [`serve design`](../../../../reference/eac/commands/serve/design.md) - View in browser
- [`validate design`](../../../../reference/eac/commands/validate/design.md) - Check syntax
