workspace "CLIE Configuration" "Repository-specific CLIE CLI configuration and AI prompts" {

    model {
        # External systems
        clie_cli = softwareSystem "CLIE CLI" "Consumes configuration" "Consumer"
        eac_commands = softwareSystem "EAC Commands" "Uses AI prompts" "Consumer"
        ai_providers = softwareSystem "AI Providers" "Claude, OpenAI, Gemini" "External"

        # CLIE config system
        clie_config = softwareSystem "CLIE Configuration" "Repository-specific configuration for CLIE CLI and AI prompts" {

            cli_config = container "CLI Configuration" "CLIE CLI settings" "YAML" {
                clie_cli_yml = component "clie-cli.yml" "CLI behavior and extension settings" "YAML"
                definitions_yml = component "definitions.yml" "Custom definitions and aliases" "YAML"
            }

            eac_config = container "EAC Configuration" "EAC-specific settings" "YAML" {
                repository_yml = component "repository.yml" "Repository metadata and settings" "YAML"
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
        clie_cli -> clie_config "Loads configuration from"
        eac_commands -> ai_prompts "Uses prompts for AI generation"
        ai_prompts -> ai_providers "Sends prompts to"
        cli_config -> clie_cli "Configures behavior"
    }

    views {
        systemContext clie_config "SystemContext" {
            include *
            autoLayout lr
            title "CLIE Configuration - System Context"
        }

        container clie_config "Containers" {
            include *
            autoLayout tb
            title "CLIE Configuration - Structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
