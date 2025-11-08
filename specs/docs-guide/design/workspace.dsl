workspace "Documentation Guide Architecture" "Diataxis-based documentation structure for the EAC framework" {

    model {
        # External actors
        developer = person "Developer" "Developer reading and contributing to documentation"
        contributor = person "Technical Writer" "Technical writer maintaining documentation"

        # Main documentation system
        docs_guide_system = softwareSystem "Documentation Guide" "Diataxis-structured documentation covering tutorials, how-to guides, reference, and explanation" {

            # Documentation containers (sections)
            tutorials = container "Tutorials" "Learning-oriented documentation guiding users through first steps and basic workflows." "Markdown" "Content" {
                gettingStarted = component "Getting Started" "Initial setup and first use tutorial" "Markdown"
                firstWorkflow = component "First Workflow" "Building first automation workflow" "Markdown"
            }

            howto_guides = container "How-To Guides" "Problem-oriented documentation providing step-by-step solutions to specific tasks." "Markdown" "Content" {
                installationGuides = component "Installation Guides" "Platform-specific installation instructions" "Markdown"
                configurationGuides = component "Configuration Guides" "Configuration and customization guides" "Markdown"
                integrationGuides = component "Integration Guides" "Third-party tool integration guides" "Markdown"
            }

            reference = container "Reference" "Information-oriented documentation covering CLI commands, APIs, configurations, and contracts." "Markdown" "Content" {
                commandReference = component "Command Reference" "Complete CLI command documentation" "Markdown"
                apiReference = component "API Reference" "Module API documentation" "Markdown"
                configReference = component "Configuration Reference" "Configuration schema and options" "Markdown"
                contractReference = component "Contract Reference" "Module contract specifications" "Markdown"
            }

            explanation = container "Explanation" "Understanding-oriented documentation covering architecture, concepts, and design decisions." "Markdown" "Content" {
                architecture = component "Architecture" "System architecture and design patterns" "Markdown"
                concepts = component "Concepts" "Core concepts and terminology" "Markdown"
                designDecisions = component "Design Decisions" "ADRs and architectural choices" "Markdown"
            }

            navigation = container "Navigation Structure" "MkDocs navigation configuration defining documentation hierarchy and organization." "YAML" "Infrastructure" {
                navConfig = component "Nav Config" ".nav.yml defining documentation structure" "YAML"
            }

            assets = container "Assets" "Images, diagrams, and static files supporting documentation content." "Static Files" "Infrastructure" {
                images = component "Images" "Screenshots and illustrative images" "PNG/SVG"
                diagrams = component "Diagrams" "Architecture and flow diagrams" "Mermaid/SVG"
            }
        }

        # User interactions
        developer -> docs_guide_system "Reads documentation" "Web browser"
        contributor -> docs_guide_system "Writes and maintains documentation" "Markdown editor"

        # Internal relationships

        # Navigation relationships
        navigation -> tutorials "Organizes tutorials section" "Nav structure"
        navigation -> howto_guides "Organizes how-to guides section" "Nav structure"
        navigation -> reference "Organizes reference section" "Nav structure"
        navigation -> explanation "Organizes explanation section" "Nav structure"

        # Asset relationships
        tutorials -> assets "Embeds images and diagrams" "Markdown links"
        howto_guides -> assets "Embeds images and diagrams" "Markdown links"
        reference -> assets "Embeds diagrams" "Markdown links"
        explanation -> assets "Embeds architecture diagrams" "Markdown links"

        # Cross-references between sections
        tutorials -> howto_guides "Links to detailed guides" "Markdown links"
        howto_guides -> reference "Links to command reference" "Markdown links"
        explanation -> reference "References contracts and APIs" "Markdown links"
    }

    views {
        # System Context view
        systemContext docs_guide_system "SystemContext" {
            include *
            autoLayout lr
            title "Documentation Guide - System Context"
            description "Shows the documentation system with users and contributors"
        }

        # Container view
        container docs_guide_system "Containers" {
            include *
            autoLayout tb
            title "Documentation Guide - Container Architecture"
            description "Shows the 4 Diataxis sections plus navigation and assets"
        }

        # Component views for documentation sections

        component tutorials "TutorialsComponents" {
            include *
            autoLayout tb
            title "Tutorials - Components"
            description "Learning-oriented tutorial content"
        }

        component howto_guides "HowToComponents" {
            include *
            autoLayout tb
            title "How-To Guides - Components"
            description "Problem-solving guide content"
        }

        component reference "ReferenceComponents" {
            include *
            autoLayout tb
            title "Reference - Components"
            description "Technical reference documentation"
        }

        component explanation "ExplanationComponents" {
            include *
            autoLayout tb
            title "Explanation - Components"
            description "Understanding-oriented conceptual content"
        }

        # Styles
        styles {
            element "Person" {
                shape person
                background #08427b
                color #ffffff
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
                shape roundedbox
            }
            element "Content" {
                background #2E7D32
                color #ffffff
            }
            element "Infrastructure" {
                background #F57C00
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
                shape component
            }
            relationship "Relationship" {
                thickness 2
                color #707070
                style solid
            }
            relationship "Markdown links" {
                thickness 2
                color #2E7D32
            }
            relationship "Nav structure" {
                thickness 2
                color #1976D2
                style dashed
            }
        }

        # Theme
        theme default
    }

    configuration {
        scope softwaresystem
    }

}
