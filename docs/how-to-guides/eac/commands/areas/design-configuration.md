# Architecture Design Configuration

{{ page_breadcrumb() }}

This guide covers configuration options for EAC's architecture design system, including Structurizr DSL settings, themes, and AI generation options.

## Configuration Files

| File                             | Purpose                     |
| -------------------------------- | --------------------------- |
| `<module>/.design/workspace.dsl` | Module architecture diagram |
| `.r2r/eac/design/themes/`        | Custom diagram themes       |
| `.r2r/eac/ai/design/`            | AI prompt templates         |
| `.r2r/eac/design/config.yml`     | Global design settings      |

## Workspace DSL Structure

### Basic Workspace

```dsl
workspace "Module Name" "Description" {

    model {
        // People and systems
        user = person "User" "End user of the system"

        system = softwareSystem "My System" {
            description "Main system description"

            api = container "API" {
                description "REST API service"
                technology "Go"
            }

            db = container "Database" {
                description "Data storage"
                technology "PostgreSQL"
            }
        }

        // Relationships
        user -> api "Uses" "HTTPS"
        api -> db "Reads/Writes" "SQL"
    }

    views {
        container system "Containers" {
            include *
            autoLayout
        }

        theme default
    }

}
```

### Model Elements

#### People

```dsl
user = person "User" "Description" {
    tags "External"
}

admin = person "Administrator" {
    description "System administrator"
    tags "Internal"
}
```

#### Software Systems

```dsl
system = softwareSystem "Name" "Description" {
    tags "Tag1" "Tag2"
}

external = softwareSystem "External System" {
    description "Third-party service"
    tags "External"
}
```

#### Containers

```dsl
container "API" {
    description "REST API"
    technology "Go"
    tags "Service"
}

container "Web App" {
    description "Frontend application"
    technology "React"
    tags "Frontend"
}

container "Database" {
    description "Data storage"
    technology "PostgreSQL"
    tags "Database"
}
```

#### Components

```dsl
api = container "API" {
    component "AuthController" {
        description "Handles authentication"
        technology "Go"
    }

    component "UserService" {
        description "User management"
        technology "Go"
    }

    component "Repository" {
        description "Data access"
        technology "Go"
    }
}
```

### Relationships

```dsl
// Simple relationship
user -> api "Uses"

// With technology
api -> db "Reads/Writes" "SQL"

// With description and technology
user -> api "Authenticates" "HTTPS/JSON"

// Between components
authController -> userService "Validates users"
userService -> repository "Persists data"
```

## Views Configuration

### Container View

```dsl
views {
    container system "ContainerView" {
        include *
        exclude "element.tag==External"
        autoLayout lr
    }
}
```

### Component View

```dsl
views {
    component api "ComponentView" {
        include *
        autoLayout tb
    }
}
```

### Layout Options

| Option          | Description                |
| --------------- | -------------------------- |
| `autoLayout`    | Automatic layout (default) |
| `autoLayout lr` | Left to right              |
| `autoLayout rl` | Right to left              |
| `autoLayout tb` | Top to bottom              |
| `autoLayout bt` | Bottom to top              |

### Styling

```dsl
views {
    styles {
        element "Person" {
            shape Person
            background #08427B
            color #ffffff
        }

        element "Software System" {
            background #1168BD
            color #ffffff
        }

        element "Container" {
            background #438DD5
            color #ffffff
        }

        element "Database" {
            shape Cylinder
        }

        relationship "Relationship" {
            thickness 2
            color #707070
            style dashed
        }
    }
}
```

## Themes

### Built-in Themes

```dsl
views {
    theme default
    // or
    theme https://static.structurizr.com/themes/amazon-web-services-2020.04.30/theme.json
}
```

### Custom Theme

Create `.r2r/eac/design/themes/custom.json`:

```json
{
  "name": "Custom Theme",
  "elements": [
    {
      "tag": "Person",
      "shape": "Person",
      "background": "#08427B",
      "color": "#ffffff"
    },
    {
      "tag": "Software System",
      "shape": "RoundedBox",
      "background": "#1168BD",
      "color": "#ffffff"
    },
    {
      "tag": "Container",
      "shape": "RoundedBox",
      "background": "#438DD5",
      "color": "#ffffff"
    },
    {
      "tag": "Component",
      "shape": "Component",
      "background": "#85BBF0",
      "color": "#000000"
    },
    {
      "tag": "Database",
      "shape": "Cylinder",
      "background": "#438DD5",
      "color": "#ffffff"
    }
  ],
  "relationships": [
    {
      "tag": "Relationship",
      "thickness": 2,
      "color": "#707070",
      "dashed": false
    },
    {
      "tag": "Async",
      "thickness": 2,
      "color": "#707070",
      "dashed": true
    }
  ]
}
```

### Using Custom Theme

```dsl
views {
    theme .r2r/eac/design/themes/custom.json
}
```

## AI Configuration

### Prompt Templates

Location: `.r2r/eac/ai/design/`

```text
.r2r/eac/ai/design/
├── create-workspace.md     # Initial generation prompt
├── update-workspace.md     # Update prompt
└── analyze-module.md       # Code analysis prompt
```

### Create Workspace Prompt

```markdown
# Architecture Diagram Generation

## Context
Generate a Structurizr DSL workspace for a software module.

## Module Information
- Moniker: {{.Module.Moniker}}
- Type: {{.Module.Type}}
- Description: {{.Module.Description}}
- Dependencies: {{.Module.Dependencies}}

## Code Analysis
- Packages: {{.Analysis.Packages}}
- Interfaces: {{.Analysis.Interfaces}}
- External Calls: {{.Analysis.ExternalCalls}}

## Guidelines
1. Use C4 model conventions
2. Include all significant containers
3. Show key relationships
4. Use appropriate technology labels
5. Apply consistent styling

## Output
Generate complete workspace.dsl content.
```

### Update Workspace Prompt

```markdown
# Architecture Diagram Update

## Existing Workspace
{{.ExistingDSL}}

## Changes Detected
- New packages: {{.Changes.NewPackages}}
- Removed packages: {{.Changes.RemovedPackages}}
- New dependencies: {{.Changes.NewDeps}}

## Guidelines
1. Preserve existing structure where valid
2. Add new elements for new code
3. Remove obsolete elements
4. Update relationships

## Output
Generate updated workspace.dsl content.
```

## Global Settings

### Design Config

`.r2r/eac/design/config.yml`:

```yaml
# Default settings for all diagrams
defaults:
  autoLayout: tb
  theme: default

# Module-specific overrides
modules:
  eac-commands:
    autoLayout: lr
    theme: custom

  eac-core:
    includeTests: false

# Generation settings
generation:
  includeInternal: true
  minComponents: 3
  maxDepth: 2

# Validation settings
validation:
  requireDescription: true
  requireTechnology: true
  checkRelationships: true
```

### Per-Module Settings

Override in module contract:

```yaml
# modules.yml
modules:
  - moniker: eac-commands
    type: go-commands
    design:
      enabled: true
      autoLayout: lr
      includeTests: false
```

## Docker Configuration

### Structurizr Lite

`serve design` uses Docker:

```yaml
# Default container settings
structurizr:
  image: structurizr/lite:latest
  port: 8080
  volume: .design:/usr/local/structurizr
```

### Custom Port

```bash
# Use custom port
r2r eac serve design eac-commands --port 9090
```

### Persistent Container

```bash
# Keep container running
r2r eac serve design eac-commands --detach

# Stop later
docker stop structurizr-eac-commands
```

## Integration Settings

### MkDocs Integration

Export diagrams for documentation:

```yaml
# mkdocs.yml
plugins:
  - structurizr:
      workspace_dir: .design
      output_dir: docs/diagrams
      format: svg
```

### CI/CD Integration

```yaml
# .github/workflows/docs.yml
- name: Validate designs
  run: r2r eac validate design

- name: Export diagrams
  run: |
    r2r eac serve design --export svg
    cp .design/*.svg docs/diagrams/
```

## Environment Variables

| Variable           | Description             | Default           |
| ------------------ | ----------------------- | ----------------- |
| `STRUCTURIZR_PORT` | Structurizr Lite port   | `8080`            |
| `DESIGN_THEME`     | Default theme           | `default`         |
| `DESIGN_OUTPUT`    | Export output directory | `.design/export/` |

## Validation Rules

### DSL Validation

```bash
r2r eac validate design eac-commands
```

Validates:

- DSL syntax correctness
- Element references exist
- Relationship targets valid
- View includes valid elements

### Common Errors

| Error                 | Cause                         | Fix                          |
| --------------------- | ----------------------------- | ---------------------------- |
| `Unknown identifier`  | Element not defined           | Add element definition       |
| `Circular reference`  | Self-referencing relationship | Remove or fix relationship   |
| `Missing description` | Required field empty          | Add description              |
| `Invalid technology`  | Unsupported value             | Use standard technology name |

## Example Configurations

### Microservice Module

```dsl
workspace "User Service" "User management microservice" {

    model {
        user = person "API Consumer"

        userService = softwareSystem "User Service" {
            api = container "REST API" {
                technology "Go"
                description "HTTP endpoints"
            }

            grpc = container "gRPC Server" {
                technology "Go"
                description "Internal communication"
            }

            db = container "User Database" {
                technology "PostgreSQL"
                tags "Database"
            }

            cache = container "Cache" {
                technology "Redis"
                tags "Cache"
            }
        }

        user -> api "HTTP/JSON"
        api -> db "SQL"
        api -> cache "GET/SET"
        grpc -> db "SQL"
    }

    views {
        container userService "Containers" {
            include *
            autoLayout lr
        }

        styles {
            element "Database" {
                shape Cylinder
            }
            element "Cache" {
                shape Cylinder
                background #DC382D
            }
        }
    }
}
```

### CLI Module

```dsl
workspace "CLI Tool" "Command-line interface" {

    model {
        user = person "Developer"

        cli = softwareSystem "CLI" {
            commands = container "Commands" {
                technology "Go"
                description "Command implementations"
            }

            core = container "Core Library" {
                technology "Go"
                description "Shared functionality"
            }

            config = container "Configuration" {
                technology "YAML"
                description "Config files"
            }
        }

        user -> commands "Executes"
        commands -> core "Uses"
        commands -> config "Reads"
    }

    views {
        container cli {
            include *
            autoLayout tb
        }
    }
}
```

## Related Documentation

- [Design Overview](design-overview.md) - Concepts and workflows
- [Design Commands](design-commands.md) - Command reference
- [Templates Configuration](templates-configuration.md) - Documentation templates

{{ diataxis_footer() }}
