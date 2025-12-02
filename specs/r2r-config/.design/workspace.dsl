workspace "R2R Configuration" "Repository-specific R2R CLI configuration and AI prompts" {

    model {
        # External systems
        r2r_cli = softwareSystem "R2R CLI" "Consumes configuration" "Consumer"
        eac_commands = softwareSystem "EAC Commands" "Uses AI prompts" "Consumer"
        ai_providers = softwareSystem "AI Providers" "Claude, OpenAI, Gemini" "External"

        # R2R config system
        r2r_config = softwareSystem "R2R Configuration" "Repository-specific configuration for R2R CLI and AI prompts" {

            cli_config = container "CLI Configuration" "R2R CLI settings" "YAML" {
                r2r_cli_yml = component "r2r-cli.yml" "CLI behavior and extension settings" "YAML"
                definitions_yml = component "definitions.yml" "Custom definitions and aliases" "YAML"
            }

            eac_config = container "EAC Configuration" "EAC-specific settings" "YAML" {
                repository_yml = component "repository.yml" "Repository metadata and settings" "YAML"
                handlers_yml = component "handlers.yml" "Event handlers configuration" "YAML"
                logging_yml = component "logging.yml" "Logging configuration" "YAML"
                agent_config_yml = component "agent-config.yml" "AI agent configuration" "YAML"
            }

            ai_prompts = container "AI Prompts" "Prompt templates for AI features" "Markdown" {
                commit_message = component "Commit Message Prompts" "Templates for generating commit messages" "Markdown"
                design_prompts = component "Design Prompts" "Templates for architecture generation" "Markdown"
                spec_prompts = component "Specification Prompts" "Templates for BDD spec generation" "Markdown"
                risk_prompts = component "Risk Prompts" "Templates for risk assessment" "Markdown"
            }

            cache = container "Cache" "Runtime cache for CLI operations" "JSON" {
                metadata_cache = component "Metadata Cache" "Cached repository metadata" "JSON"
                registry_cache = component "Registry Cache" "Extension registry cache" "JSON"
            }
        }

        # Relationships
        r2r_cli -> r2r_config "Loads configuration from"
        eac_commands -> ai_prompts "Uses prompts for AI generation"
        ai_prompts -> ai_providers "Sends prompts to"
        cli_config -> r2r_cli "Configures behavior"
    }

    views {
        systemContext r2r_config "SystemContext" {
            include *
            autoLayout lr
            title "R2R Configuration - System Context"
        }

        container r2r_config "Containers" {
            include *
            autoLayout tb
            title "R2R Configuration - Structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
