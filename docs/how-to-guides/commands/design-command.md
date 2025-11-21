# Design Command

**Problem**: Creating and maintaining architecture diagrams is tedious and diagrams become outdated quickly.

**Solution**: Use `design` to generate and visualize C4 architecture diagrams as code using Structurizr DSL.

## Key Benefits

- AI-generated architecture diagrams from source code
- Diagrams as code (version controlled, reviewable)
- Interactive C4 model views (System Context, Container, Component)
- Auto-validation with Structurizr CLI
- Live browser preview with hot-reload

## Quick Start

```bash
# Generate architecture for a module
r2r eac design create src-commands

# Validate workspace.dsl syntax
r2r eac design validate src-commands

# View diagrams in browser
r2r eac design serve src-commands
```

## Typical Workflow

### Creating Architecture Diagrams

```bash
# Generate from source code
r2r eac design create src-auth

# Output: specs/src-auth/design/workspace.dsl created
# Includes: system context, containers, components

# View in browser
r2r eac design serve src-auth
# Opens http://localhost:8080
```

### Validation and Iteration

```bash
# Validate DSL syntax
r2r eac design validate src-api

# Fix any errors in workspace.dsl
nano specs/src-api/design/workspace.dsl

# Revalidate
r2r eac design validate src-api

# Serve to view changes
r2r eac design serve src-api
```

### Debug AI Generation

```bash
# Generate with debug output
r2r eac design create src-core --debug

# Inspect intermediate files
cat out/debug-full-prompt.md
cat out/debug-raw-ai-response.md
cat out/debug-validation-result.json
```

## Command Reference

### design create

Generate Structurizr DSL workspace from source code.

```bash
r2r eac design create <module> [options]

# Options:
--output, -o <path>    # Custom output path
--force, -f            # Overwrite existing workspace.dsl
--debug, -d            # Save intermediate outputs to out/

# Examples:
r2r eac design create src-auth
r2r eac design create src-api --output custom/path/workspace.dsl
r2r eac design create src-core --force
r2r eac design create src-commands --debug
```

**What it does:**
1. Analyzes source code in `src/<module>/`
2. Uses AI to identify architecture elements
3. Generates Structurizr DSL with C4 model views
4. Validates DSL syntax with Docker/Structurizr CLI
5. Auto-retries if validation fails
6. Saves to `specs/<module>/design/workspace.dsl`

**Requirements:**
- Docker must be installed and running
- Module must exist in `src/<module>/`
- AI provider configured (`r2r eac init --ai <provider>`)

### design validate

Validate workspace.dsl syntax using Structurizr CLI.

```bash
r2r eac design validate <module> [options]

# Options:
--all                  # Validate all workspaces in specs/*/design/

# Examples:
r2r eac design validate src-auth
r2r eac design validate src-api
r2r eac design validate --all
```

**Validation checks:**
- DSL syntax correctness
- Element relationships validity
- View definitions correctness
- Workspace renderability

**Output:**
```
🔍 Validating module: src-auth
📄 Workspace: specs/src-auth/design/workspace.dsl
🐳 Using Docker: structurizr/cli:latest

✅ Workspace is valid

📊 Summary:
  Errors: 0
  Warnings: 0
  Execution time: 1.23s

📝 Results written to: out/design-validation-results.json
```

### design serve

View architecture diagrams in browser using Structurizr Lite.

```bash
r2r eac design serve <module> [options]

# Options:
--force, -f            # Stop existing container and start new one

# Examples:
r2r eac design serve src-auth
r2r eac design serve src-api --force
```

**What happens:**
1. Starts Structurizr Lite in Docker on port 8080
2. Mounts `specs/<module>/design/` directory
3. Opens browser to http://localhost:8080
4. Auto-reloads when workspace.dsl changes

**Viewer features:**
- Interactive C4 diagram navigation
- Zoom, pan, and explore elements
- Switch between views (System Context, Container, Component)
- Export diagrams as PNG/SVG

**Stopping the server:**
```bash
docker stop structurizr-lite-<module>
```

## Generated Workspace Structure

### Example workspace.dsl

```dsl
workspace "Authentication Module" "User authentication and authorization" {
    model {
        user = person "User" "Application user"
        admin = person "Admin" "System administrator"

        authSystem = softwareSystem "Authentication System" {

            api = container "API Layer" "REST API for authentication" "Go/Fiber" {
                loginHandler = component "Login Handler" "Processes login requests"
                tokenService = component "Token Service" "Manages JWT tokens"
                sessionStore = component "Session Store" "Stores user sessions"
            }

            database = container "Database" "User credentials and sessions" "PostgreSQL"
        }

        # Relationships
        user -> api "Authenticates via"
        api -> database "Reads/writes"
        loginHandler -> tokenService "Generates tokens"
        tokenService -> sessionStore "Stores sessions"
    }

    views {
        systemContext authSystem "SystemContext" {
            include *
            autoLayout
        }

        container authSystem "Containers" {
            include *
            autoLayout
        }

        component api "Components" {
            include *
            autoLayout
        }

        styles {
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85BBF0
                color #000000
            }
        }
    }
}
```

### File Organization

```
specs/
└── src-auth/
    └── design/
        ├── workspace.dsl           # Source DSL file
        ├── workspace.json          # Generated by Structurizr
        └── .structurizr/           # Viewer metadata
            └── workspace.dsl
```

## C4 Model Views

### System Context

Shows how the system fits in the world:
- External systems
- Users/actors
- High-level relationships

**Use for:** Executive overviews, system boundaries

### Container View

Shows high-level architecture:
- Applications/services
- Databases
- Major components
- Inter-container communication

**Use for:** Deployment planning, team organization

### Component View

Shows internal structure:
- Classes/packages
- Internal components
- Detailed relationships

**Use for:** Developer onboarding, code navigation

## Debug Mode

Use `--debug` to inspect AI generation:

```bash
r2r eac design create src-module --debug
```

Creates debug files in `out/`:

```
out/
├── debug-full-prompt.md           # AI prompt with source code
├── debug-raw-ai-response.dsl      # Unfiltered AI output
├── debug-cleaned-output.dsl       # After anti-corruption layer
└── debug-validation-result.json   # Structurizr validation
```

**When to use:**
- AI generates invalid DSL
- Missing architecture elements
- Incorrect relationships
- Customizing AI prompts

## Customization

### Custom Prompts

Override default AI behavior:

```bash
# Edit system prompt
nano .r2r/contracts/ai/design-create/system-prompt.md

# Changes apply to all future generation
r2r eac design create src-module
```

### Manual Editing

Edit generated workspace.dsl:

```bash
# Generate initial structure
r2r eac design create src-module

# Edit manually
nano specs/src-module/design/workspace.dsl

# Validate changes
r2r eac design validate src-module

# View updated diagrams
r2r eac design serve src-module
```

### Styling

Customize diagram appearance in workspace.dsl:

```dsl
views {
    styles {
        element "Software System" {
            background #1168bd
            color #ffffff
            fontSize 24
            shape RoundedBox
        }
        element "Person" {
            background #08427b
            color #ffffff
            shape Person
        }
        element "Database" {
            shape Cylinder
        }
    }
}
```

## Integration Patterns

### Documentation Workflow

```bash
# Generate architecture docs
r2r eac design create src-module

# Validate
r2r eac design validate src-module

# Commit
git add specs/src-module/design/
r2r eac work commit -m "docs: add architecture diagrams"

# View in MkDocs
r2r eac docs serve
```

### Review Workflow

```bash
# Create diagrams for new module
r2r eac design create src-new-feature

# Review in browser
r2r eac design serve src-new-feature

# Iterate based on feedback
nano specs/src-new-feature/design/workspace.dsl
r2r eac design validate src-new-feature

# Commit for PR
r2r eac work commit --all
r2r eac work pr
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Validate architecture
  run: |
    r2r eac design validate --all
    if [ $? -ne 0 ]; then
      echo "Architecture validation failed"
      exit 1
    fi
```

## Docker Requirements

### Check Docker

```bash
# Verify Docker is running
docker ps

# If not running:
# Windows: Start Docker Desktop
# Linux: sudo systemctl start docker
# macOS: Start Docker Desktop
```

### Structurizr Images

The command uses these Docker images:
- `structurizr/cli:latest` - For validation
- `structurizr/lite:latest` - For viewer

Images are pulled automatically on first use.

## Best Practices

- **Generate early**: Create diagrams when starting modules
- **Keep updated**: Regenerate when architecture changes significantly
- **Manual refinement**: AI generates structure, you add details
- **Validate often**: Run `design validate` before commits
- **Version control**: Commit workspace.dsl with code
- **Review in PRs**: Use diagrams for architecture discussions
- **Live documentation**: Use `design serve` during development

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Docker not found | Install Docker Desktop, ensure it's running |
| Module not found | Check module exists in `src/<module>/` |
| Validation fails | Use `--debug`, check DSL syntax manually |
| Port 8080 in use | Stop existing container: `docker stop structurizr-lite-<module>` |
| AI generates invalid DSL | Use `--debug`, check AI provider setup |
| Workspace exists | Use `--force` to overwrite |
| Browser doesn't open | Manually visit http://localhost:8080 |

## Advanced Usage

### Batch Generation

```bash
# Generate for multiple modules
for module in src-auth src-api src-core; do
  r2r eac design create $module
done

# Validate all
r2r eac design validate --all
```

### Custom Output Paths

```bash
# Save to custom location
r2r eac design create src-module --output docs/architecture/system.dsl

# Validate custom path
docker run --rm -v "$(pwd):/workspace" structurizr/cli:latest \
  validate -w /workspace/docs/architecture/system.dsl
```

### Export Diagrams

```bash
# Start viewer
r2r eac design serve src-module

# In browser at http://localhost:8080:
# - Click on a view
# - Click Export → PNG/SVG
# - Save to docs/images/
```

### Programmatic Access

Use Structurizr CLI directly:

```bash
# Export to JSON
docker run --rm -v "$(pwd):/workspace" structurizr/cli:latest \
  export -workspace /workspace/specs/src-module/design/workspace.dsl \
  -format json -output /workspace/out/

# Generate PlantUML
docker run --rm -v "$(pwd):/workspace" structurizr/cli:latest \
  export -workspace /workspace/specs/src-module/design/workspace.dsl \
  -format plantuml -output /workspace/out/
```

## Summary

1. **Create**: `r2r eac design create <module>`
2. **Validate**: `r2r eac design validate <module>`
3. **View**: `r2r eac design serve <module>`
4. **Edit** (optional): Refine workspace.dsl manually
5. **Commit**: `git add specs/` and commit

Architecture diagrams as code provide living documentation that evolves with your codebase.
