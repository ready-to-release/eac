workspace "Templates" "Reusable templates for specifications and documentation" {

    model {
        # External systems
        eac_commands = softwareSystem "EAC Commands" "Applies templates" "Consumer"
        r2r_cli = softwareSystem "R2R CLI" "Template installation" "Consumer"

        # Templates system
        templates = softwareSystem "Templates" "Reusable templates for project artifacts" {

            spec_templates = container "Specification Templates" "BDD specification templates" "Gherkin" {
                risk_controls = component "Risk Controls Catalog" "Security control specifications" "Feature Files"
            }

            doc_templates = container "Documentation Templates" "Documentation templates" "Markdown" {
                readme_templates = component "README Templates" "Project readme templates" "Markdown"
                changelog_templates = component "Changelog Templates" "Release changelog format" "Markdown"
            }

            report_templates = container "Report Templates" "Generated report templates" "Markdown" {
                security_reports = component "Security Reports" "Security scan report templates" "Markdown"
                compliance_reports = component "Compliance Reports" "Compliance assessment templates" "Markdown"
            }
        }

        # Relationships
        eac_commands -> templates "Applies templates from"
        r2r_cli -> templates "Installs templates from"
        spec_templates -> eac_commands "Provides risk control specs to"
    }

    views {
        systemContext templates "SystemContext" {
            include *
            autoLayout lr
            title "Templates - System Context"
        }

        container templates "Containers" {
            include *
            autoLayout tb
            title "Templates - Categories"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
