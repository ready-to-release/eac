workspace "Documentation Site" "MkDocs documentation for EAC methodology" {

    model {
        # External systems
        mkdocs = softwareSystem "MkDocs" "Static site generator" "External"
        github_pages = softwareSystem "GitHub Pages" "Documentation hosting" "External"
        readers = person "Readers" "Documentation consumers" "External"

        # Docs system
        docs = softwareSystem "Documentation Site" "MkDocs-based documentation explaining EAC methodology" {

            explanation = container "Explanation" "Conceptual documentation" "Markdown" {
                cd_docs = component "Continuous Delivery" "CD model, workflows, testing taxonomy" "Markdown"
                eac_docs = component "Everything as Code" "EAC principles and practices" "Markdown"
                specs_docs = component "Specifications" "BDD specification guidance" "Markdown"
                lifecycle_docs = component "Lifecycle" "Software lifecycle documentation" "Markdown"
                transformation = component "Transformation" "Organizational transformation guides" "Markdown"
            }

            reference = container "Reference" "Technical reference documentation" "Markdown" {
                decision_records = component "Decision Records" "Architecture decision records" "Markdown"
                specifications = component "Specifications Index" "Generated specification index" "Markdown"
            }

            tutorials = container "Tutorials" "Step-by-step guides" "Markdown" {
                getting_started = component "Getting Started" "Onboarding tutorials" "Markdown"
            }

            assets = container "Assets" "Static assets and diagrams" "Files" {
                diagrams = component "Diagrams" "DrawIO architecture diagrams" "PNG"
                references = component "Reference Images" "Book covers and citations" "PNG"
                pdfs = component "PDF References" "Research papers and guides" "PDF"
                logo = component "Logo" "EAC branding assets" "PNG"
            }

            config = container "Site Configuration" "MkDocs configuration" "YAML" {
                nav_files = component "Navigation Files" ".nav.yml navigation structure" "YAML"
            }
        }

        # Relationships
        readers -> docs "Reads documentation"
        mkdocs -> docs "Builds site from"
        docs -> github_pages "Deploys to"
        explanation -> assets "References diagrams from"
    }

    views {
        systemContext docs "SystemContext" {
            include *
            autoLayout lr
            title "Documentation Site - System Context"
        }

        container docs "Containers" {
            include *
            autoLayout tb
            title "Documentation Site - Content Structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
