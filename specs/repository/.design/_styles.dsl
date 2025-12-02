# Shared Structurizr DSL Styles (Archive)
#
# NOTE: Currently all workspace.dsl files use `theme default` only.
# This file is kept as a reference for custom styles if needed in the future.

styles {
    # === ELEMENT STYLES ===

    # People
    element "Person" {
        shape person
        background #08427b
        color #ffffff
    }
    element "Client" {
        shape person
        background #08427b
        color #ffffff
    }
    element "Automation" {
        background #2e7d32
        color #ffffff
        shape robot
    }

    # Software Systems
    element "Software System" {
        background #1168bd
        color #ffffff
    }
    element "External" {
        background #999999
        color #ffffff
    }
    element "Dependency" {
        background #1168bd
        color #ffffff
        shape roundedbox
    }
    element "Dependent" {
        background #1168bd
        color #ffffff
        shape roundedbox
    }
    element "Consumer" {
        background #1168bd
        color #ffffff
    }

    # Containers
    element "Container" {
        background #438dd5
        color #ffffff
        shape roundedbox
    }
    element "Go Package" {
        background #438dd5
        color #ffffff
    }
    element "Go Packages" {
        background #2E7D32
        color #ffffff
    }

    # Components
    element "Component" {
        background #85bbf0
        color #000000
        shape component
    }

    # Technology-specific containers
    element "Core" {
        background #2E7D32
        color #ffffff
    }
    element "Command" {
        background #1976D2
        color #ffffff
    }
    element "Infrastructure" {
        background #F57C00
        color #ffffff
    }
    element "MCP Server" {
        background #2E7D32
        color #ffffff
    }

    # Language/Platform specific
    element "Docker" {
        background #2496ED
        color #ffffff
    }
    element "Docker Stage" {
        background #85bbf0
        color #000000
    }
    element "PowerShell" {
        background #012456
        color #ffffff
    }
    element "Bash" {
        background #4EAA25
        color #ffffff
    }
    element "TypeScript" {
        background #3178C6
        color #ffffff
    }
    element "JSON" {
        background #F7DF1E
        color #000000
    }
    element "YAML" {
        background #CB171E
        color #ffffff
    }
    element "Markdown" {
        background #083fa1
        color #ffffff
    }
    element "Gherkin" {
        background #23D96C
        color #000000
    }
    element "CSS" {
        background #264de4
        color #ffffff
    }
    element "Files" {
        background #666666
        color #ffffff
    }

    # === RELATIONSHIP STYLES ===

    # Default
    relationship "Relationship" {
        thickness 2
        color #707070
        style solid
    }

    # Code dependencies
    relationship "Go Import" {
        thickness 2
        color #00ADD8
        style solid
    }
    relationship "Go package import" {
        thickness 2
        color #5E35B1
        style dashed
    }
    relationship "Function calls" {
        thickness 2
        color #2E7D32
    }
    relationship "Function call" {
        thickness 2
        color #2E7D32
    }
    relationship "Go function calls" {
        thickness 2
        color #2E7D32
    }

    # External integrations
    relationship "HTTPS/REST" {
        thickness 2
        color #D32F2F
        style dashed
    }
    relationship "HTTPS" {
        thickness 2
        color #D32F2F
        style dashed
    }
    relationship "File I/O" {
        thickness 2
        color #F57C00
        style dashed
    }
    relationship "CLI" {
        thickness 2
        color #7B1FA2
        style dashed
    }
    relationship "CLI (os/exec)" {
        thickness 2
        color #7B1FA2
        style dashed
    }
    relationship "Go (os/exec)" {
        thickness 2
        color #7B1FA2
    }
    relationship "CLI subprocess" {
        thickness 3
        color #1976D2
        style dashed
    }

    # Docker/Container
    relationship "Docker" {
        thickness 2
        color #2496ED
        style dashed
    }
    relationship "Docker API" {
        thickness 3
        color #1976D2
        style dashed
    }

    # Protocol specific
    relationship "JSON-RPC over stdio" {
        thickness 3
        color #5E35B1
        style dashed
    }
    relationship "JSON Schema" {
        thickness 2
        color #388E3C
        style dotted
    }
}
