workspace "EAC AI Library" "AI/LLM provider integrations for code generation and analysis" {

    model {
        # External actors and systems
        eac_commands = softwareSystem "EAC Commands" "CLI commands using AI for generation" "Consumer"
        claude_api = softwareSystem "Claude API" "Anthropic Claude API" "External"
        openai_api = softwareSystem "OpenAI API" "OpenAI GPT API" "External"
        gemini_api = softwareSystem "Gemini API" "Google Gemini API" "External"
        claude_cli = softwareSystem "Claude CLI" "Claude Code CLI tool" "External"
        r2r_config = softwareSystem "R2R Config" "AI configuration from .r2r/eac/ai/" "External"

        # AI library system
        eac_ai = softwareSystem "EAC AI Library" "Go library providing unified AI provider integrations" {

            # Core AI execution
            executor = container "AI Executor" "Unified AI prompt execution with provider abstraction" "Go Package" {
                ai_executor = component "AI Executor" "Executes prompts via configured provider" "Go"
                executor_adapter = component "Executor Adapter" "Adapts executor for different use cases" "Go"
                response_handler = component "Response Handler" "Processes AI responses" "Go"
            }

            # Configuration
            config_pkg = container "Configuration" "AI provider configuration and loading" "Go Package" {
                config_types = component "Config Types" "AI configuration type definitions" "Go"
                config_loader = component "Config Loader" "Loads AI config from .r2r/eac/ai/" "Go"
                provider_selector = component "Provider Selector" "Selects provider based on config" "Go"
            }

            # Provider abstraction
            provider_pkg = container "Provider Interface" "Abstract provider interface and registry" "Go Package" {
                provider_interface = component "Provider Interface" "Abstract AI provider contract" "Go"
                provider_registry = component "Provider Registry" "Registry of available providers" "Go"
                mock_provider = component "Mock Provider" "Mock provider for testing" "Go"
            }

            # Provider implementations
            providers = container "Provider Implementations" "Concrete AI provider implementations" "Go Package" {
                claude_api_provider = component "Claude API Provider" "Anthropic Claude via REST API" "Go"
                claude_cli_provider = component "Claude CLI Provider" "Claude via claude CLI tool" "Go"
                openai_provider = component "OpenAI Provider" "OpenAI GPT via REST API" "Go"
                gemini_provider = component "Gemini Provider" "Google Gemini via REST API" "Go"
                test_provider = component "Test Provider" "Reads mock responses from files" "Go"
            }
        }

        # Consumer relationships
        eac_commands -> eac_ai "Uses for AI-powered generation"

        # External API relationships
        providers -> claude_api "Sends prompts" "HTTPS/REST"
        providers -> openai_api "Sends prompts" "HTTPS/REST"
        providers -> gemini_api "Sends prompts" "HTTPS/REST"
        providers -> claude_cli "Executes prompts" "CLI (os/exec)"
        config_pkg -> r2r_config "Loads configuration" "File I/O"

        # Internal container relationships
        executor -> config_pkg "Gets provider configuration"
        executor -> provider_pkg "Gets provider instance"
        provider_pkg -> providers "Creates provider instances"

        # Component relationships - Executor
        ai_executor -> executor_adapter "Wraps for specific use cases"
        ai_executor -> response_handler "Processes responses"

        # Component relationships - Config
        config_loader -> config_types "Creates config instances"
        config_loader -> provider_selector "Determines active provider"

        # Component relationships - Provider
        provider_registry -> provider_interface "Registers implementations"
        mock_provider -> provider_interface "Implements interface"

        # Component relationships - Implementations
        claude_api_provider -> provider_interface "Implements interface"
        claude_cli_provider -> provider_interface "Implements interface"
        openai_provider -> provider_interface "Implements interface"
        gemini_provider -> provider_interface "Implements interface"
        test_provider -> provider_interface "Implements interface"
    }

    views {
        systemContext eac_ai "SystemContext" {
            include *
            autoLayout lr
            title "EAC AI Library - System Context"
            description "Shows AI library with consumers and external APIs"
        }

        container eac_ai "Containers" {
            include *
            autoLayout tb
            title "EAC AI Library - Package Architecture"
            description "Shows executor, config, and provider packages"
        }

        component executor "ExecutorComponents" {
            include *
            autoLayout tb
            title "AI Executor - Components"
            description "Prompt execution and response handling"
        }

        component providers "ProvidersComponents" {
            include *
            autoLayout tb
            title "Provider Implementations - Components"
            description "Concrete AI provider implementations"
        }

        # Dynamic view - AI Execution Flow (internal containers only)
        dynamic eac_ai "ExecutionFlow" "How an AI prompt flows through internal containers" {
            executor -> config_pkg "1. Get provider config"
            executor -> provider_pkg "2. Get provider instance"
            provider_pkg -> providers "3. Create provider"
            autoLayout
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
