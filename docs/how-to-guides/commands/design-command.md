# Design Commands (Create, Update, Validate, Serve)

**Problem**: Creating and maintaining architecture diagrams is tedious and diagrams become outdated quickly.

**Solution**: Use the design commands (`create design`, `update design`, `validate design`, `serve design`) to generate and visualize C4 architecture diagrams as code using Structurizr DSL.

## Key Benefits

- AI-generated architecture diagrams from source code
- Diagrams as code (version controlled, reviewable)
- Interactive C4 model views (System Context, Container, Component)
- Auto-validation with Structurizr CLI
- Live browser preview with hot-reload

## Quick Start

```bash
# Generate architecture for a module
r2r eac create design eac-commands

# Validate workspace.dsl syntax
r2r eac validate design eac-commands

# View diagrams in browser
r2r eac serve design eac-commands
```

## Typical Workflow

### Creating Architecture Diagrams

```bash
# Generate from source code
r2r eac create design src-auth

# Output: specs/src-auth/.design/workspace.dsl created
# Includes: system context, containers, components

# View in browser
r2r eac serve design src-auth
# Opens http://localhost:9000 (or next available port in 9000-9999 range)
```

### Validation and Iteration

```bash
# Validate DSL syntax
r2r eac validate design src-api

# Fix any errors in workspace.dsl
nano specs/src-api/.design/workspace.dsl

# Revalidate
r2r eac validate design src-api

# Serve to view changes
r2r eac serve design src-api
```

### Debug AI Generation

```bash
# Generate with debug output
r2r eac create design eac-core --debug

# Inspect intermediate files
cat out/logs/design/debug-full-prompt.md
cat out/logs/design/debug-raw-ai-response.md
cat out/logs/design/debug-validation-result.json
```

## Command Reference

### create design

Generate Structurizr DSL workspace from source code.

```bash
r2r eac create design <module> [options]

# Options:
--output, -o <path>    # Custom output path
--force, -f            # Overwrite existing workspace.dsl
--skip-validation      # Skip automatic validation after generation
--debug, -d            # Save intermediate outputs to out/logs/design/

# Examples:
r2r eac create design src-auth
r2r eac create design src-api --output custom/path/workspace.dsl
r2r eac create design eac-core --force
r2r eac create design eac-commands --debug
r2r eac create design src-module --skip-validation
```

**What it does:**

1. Analyzes source code in `src/<module>/`
2. Uses AI to identify architecture elements
3. Generates Structurizr DSL with C4 model views
4. Validates DSL syntax with Docker/Structurizr CLI
5. Auto-retries if validation fails
6. Saves to `specs/<module>/.design/workspace.dsl`

**Requirements:**

- Docker must be installed and running
- Module must exist in `src/<module>/`
- AI provider configured (`r2r eac init --ai <provider>`)

### update design

Update existing Structurizr DSL workspace with changes from source code.

```bash
r2r eac update design <module> [options]

# Options:
--skip-validation      # Skip automatic validation after update
--debug, -d            # Save intermediate outputs to out/logs/design/

# Examples:
r2r eac update design src-auth
r2r eac update design src-api --skip-validation
r2r eac update design eac-core --debug
```

**What it does:**

1. Reads existing workspace.dsl from `specs/<module>/.design/`
2. Analyzes current source code in `src/<module>/`
3. Uses AI to update architecture based on code changes
4. Preserves manual customizations where possible
5. Validates updated DSL syntax with Docker/Structurizr CLI
6. Overwrites `specs/<module>/.design/workspace.dsl`

**When to use:**

- After significant code changes to a module
- When new components or relationships are added
- When architecture has evolved from initial design
- To refresh diagrams with latest code structure

**Best practices:**

- Review changes with `git diff` before committing
- Use `--debug` to inspect AI reasoning for updates
- Validate with `validate design` after manual adjustments
- Consider backing up workspace.dsl before major updates

**Requirements:**

- Existing workspace.dsl must be present
- Docker must be installed and running
- Module must exist in `src/<module>/`
- AI provider configured (`r2r eac init --ai <provider>`)

### validate design

Validate workspace.dsl syntax using Structurizr CLI.

```bash
r2r eac validate design <module> [options]

# Options:
--all                  # Validate all workspaces in specs/*/.design/

# Examples:
r2r eac validate design src-auth
r2r eac validate design src-api
r2r eac validate design --all
```

**Validation checks:**

- DSL syntax correctness
- Element relationships validity
- View definitions correctness
- Workspace renderability

**Output:**

```text
🔍 Validating module: src-auth
📄 Workspace: specs/src-auth/.design/workspace.dsl
🐳 Using Docker: structurizr/cli:latest

✅ Workspace is valid

📊 Summary:
  Errors: 0
  Warnings: 0
  Execution time: 1.23s

📝 Results written to: out/logs/design/validation-results.json
```

### serve design

View architecture diagrams in browser using Structurizr Lite.

```bash
r2r eac serve design <module> [options]

# Options:
--force, -f            # Stop existing container and start new one

# Examples:
r2r eac serve design src-auth
r2r eac serve design src-api --force
```

**What happens:**

1. Finds available port in range 9000-9999 (dynamically allocated)
2. Starts Structurizr Lite in Docker on that port
3. Mounts `specs/<module>/.design/` directory
4. Opens browser to `http://localhost:<port>`
5. Auto-reloads when workspace.dsl changes

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

```text
specs/
└── src-auth/
    └── .design/
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
r2r eac create design src-module --debug
```

Creates debug files in `out/logs/design/`:

```text
out/logs/design/
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
nano .r2r/eac/ai/design/design.md

# Changes apply to all future generation
r2r eac create design src-module
```

### Manual Editing

Edit generated workspace.dsl:

```bash
# Generate initial structure
r2r eac create design src-module

# Edit manually
nano specs/src-module/.design/workspace.dsl

# Validate changes
r2r eac validate design src-module

# View updated diagrams
r2r eac serve design src-module
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
r2r eac create design src-module

# Validate
r2r eac validate design src-module

# Commit
git add specs/src-module/.design/
r2r eac work commit -m "docs: add architecture diagrams"

# View in MkDocs
r2r eac serve docs
```

### Review Workflow

```bash
# Create diagrams for new module
r2r eac create design src-new-feature

# Review in browser
r2r eac serve design src-new-feature

# Iterate based on feedback
nano specs/src-new-feature/.design/workspace.dsl
r2r eac validate design src-new-feature

# Commit for PR
r2r eac work commit --all
r2r eac work pr
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Validate architecture
  run: |
    r2r eac validate design --all
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
- **Validate often**: Run `validate design` before commits
- **Version control**: Commit workspace.dsl with code
- **Review in PRs**: Use diagrams for architecture discussions
- **Live documentation**: Use `serve design` during development

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Docker not found | Install Docker Desktop, ensure it's running |
| Module not found | Check module exists in `src/<module>/` |
| Validation fails | Use `--debug`, check DSL syntax manually |
| Port in use | Command auto-selects available port in 9000-9999 range |
| AI generates invalid DSL | Use `--debug`, check AI provider setup |
| Workspace exists | Use `--force` to overwrite |
| Browser doesn't open | Manually visit URL shown in command output |

## Advanced Usage

### Batch Generation

```bash
# Generate for multiple modules
for module in src-auth src-api eac-core; do
  r2r eac create design $module
done

# Validate all
r2r eac validate design --all
```

### Custom Output Paths

```bash
# Save to custom location
r2r eac create design src-module --output docs/architecture/system.dsl

# Validate custom path
docker run --rm -v "$(pwd):/workspace" structurizr/cli:latest \
  validate -w /workspace/docs/architecture/system.dsl
```

### Export Diagrams

```bash
# Start viewer
r2r eac serve design src-module

# In browser at the displayed URL (e.g., http://localhost:9000):
# - Click on a view
# - Click Export → PNG/SVG
# - Save to docs/images/
```

### Programmatic Access

Use Structurizr CLI directly:

```bash
# Export to JSON
docker run --rm -v "$(pwd):/workspace" structurizr/cli:latest \
  export -workspace /workspace/specs/src-module/.design/workspace.dsl \
  -format json -output /workspace/out/

# Generate PlantUML
docker run --rm -v "$(pwd):/workspace" structurizr/cli:latest \
  export -workspace /workspace/specs/src-module/.design/workspace.dsl \
  -format plantuml -output /workspace/out/
```

## Summary

1. **Create**: `r2r eac create design <module>`
2. **Update**: `r2r eac update design <module>` (refresh existing diagrams)
3. **Validate**: `r2r eac validate design <module>`
4. **View**: `r2r eac serve design <module>`
5. **Edit** (optional): Refine workspace.dsl manually
6. **Commit**: `git add specs/` and commit

Architecture diagrams as code provide living documentation that evolves with your codebase.
