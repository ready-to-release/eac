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
r2r eac create design src-auth
```

**What happens**: AI analyzes module code and generates `workspace.dsl` file

### 2. Review Generated Diagram

```bash
cat src/auth/workspace.dsl
```

**What happens**: View Structurizr DSL definition

### 3. Visualize Diagram

```bash
r2r eac serve design src-auth
```

**What happens**: Starts Structurizr Lite server, opens in browser

### 4. Edit and Update

Edit `workspace.dsl` manually if needed, then update:

```bash
r2r eac update design src-auth
```

**What happens**: AI updates existing diagram with new information

## Diagram Types

C4 model levels generated:

- **System Context** - System and external dependencies
- **Container** - Major components/containers
- **Component** - Internal structure
- **Code** - Class/package details

## Example Scenario

Documenting authentication module:

```bash
# Generate diagram
r2r eac create design src-auth

# Output:
# Analyzing src-auth module...
# Generating architecture diagram...
# ✓ Created workspace.dsl

# View in browser
r2r eac serve design src-auth
# Starting Structurizr Lite on http://localhost:8080
# Open browser to view diagrams

# Make code changes, update diagram
r2r eac update design src-auth
# ✓ Updated workspace.dsl with latest changes

# Validate syntax
r2r eac validate design src-auth
# ✓ workspace.dsl syntax valid
```

## Diagram Structure

Generated workspace.dsl includes:

```dsl
workspace "src-auth" {
  model {
    user = person "User"
    authSystem = softwareSystem "Authentication" {
      api = container "Auth API" {
        loginHandler = component "Login Handler"
        tokenService = component "Token Service"
      }
      database = container "Database"
    }
  }

  views {
    systemContext authSystem
    container authSystem
    component api
  }
}
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "AI analysis failed" | Ensure code is readable, has comments |
| Diagram incomplete | Use `update design` to enhance |
| Docker not found | Install Docker for Structurizr Lite |

## Next Steps

- [Build Documentation Site](./build-documentation-site.md) → Include diagrams

## Related Commands

- [`create design`](../../../../reference/commands/create/design.md) - Generate diagram
- [`update design`](../../../../reference/commands/update/design.md) - Update existing
- [`serve design`](../../../../reference/commands/serve/design.md) - View in browser
- [`validate design`](../../../../reference/commands/validate/design.md) - Check syntax
