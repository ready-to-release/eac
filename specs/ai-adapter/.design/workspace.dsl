workspace "AI Adapter" "Pluggable AI provider abstraction for LLM integration with configuration loading and schema validation" {

    model {
        # Consumers
        eac_cli = softwareSystem "EAC CLI" "Commit, spec generation, risk assessment, design commands" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Paths, environments, domain types" "Dependency"
        ai_provider_contract = softwareSystem "AI Provider Contract" "Provider, Executor, ConfigLoader interfaces" "Dependency"

        # External systems
        claude_api = softwareSystem "Claude API" "Anthropic Claude language model" "External"
        openai_api = softwareSystem "OpenAI API" "OpenAI GPT language models" "External"
        gemini_api = softwareSystem "Gemini API" "Google Gemini language models" "External"
        filesystem = softwareSystem "File System" "AI config YAML files (.eac/ai-provider.yml)" "External"

        # AI adapter system
        ai_adapter = softwareSystem "AI Adapter" "Pluggable AI provider abstraction with factory pattern, config loading, and schema validation" {

            # Executor
            executor = container "Executor" "Provider lifecycle management and execution orchestration" "Go Package" {
                executor_impl = component "Executor" "Main executor with LoadProvider, Execute, RegisterProvider" "Go"
                executor_adapter = component "ExecutorAdapter" "Bridges concrete Executor to domain AIExecutor interface" "Go"
            }

            # Config Loader
            config_loader = container "Config Loader" "Configuration loading with env var substitution and schema validation" "Go Package" {
                loader = component "Config Loader" "Loads .eac/ai-provider.yml with team/personal overrides" "Go"
                env_substitutor = component "Env Substitutor" "Substitutes ${VAR} in config values" "Go"
                schema_validator = component "Schema Validator" "Validates config against ai-provider schema" "Go"
            }

            # Provider Registry
            provider_registry = container "Provider Registry" "Factory-based provider registration and creation" "Go Package" {
                factory_registry = component "Factory Registry" "Maps provider names to ProviderFactory functions" "Go"
            }

            # Providers
            providers = container "Providers" "Concrete AI provider implementations" "Go Package" {
                anthropic_provider = component "Anthropic Provider" "Claude API integration via anthropic-sdk-go" "Go"
                openai_provider = component "OpenAI Provider" "OpenAI API integration via go-openai" "Go"
                gemini_provider = component "Gemini Provider" "Google Gemini via generative-ai-go" "Go"
                claude_cli_provider = component "Claude CLI Provider" "Claude via CLI subprocess" "Go"
                test_provider = component "Test Provider" "MockProvider returning configured responses" "Go"
            }

            # Tool Handler
            tool_handler = container "Tool Handler" "AI tool call handler integration" "Go Package" {
                handler = component "Tool Call Handler" "Handles AI tool use responses" "Go"
            }
        }

        # Consumer relationships
        eac_cli -> ai_adapter "Generates commit messages, specs, designs, risk assessments" "Go Import"

        # Dependency relationships
        ai_adapter -> core "Uses paths, environments" "Go Import"
        ai_adapter -> ai_provider_contract "Implements Provider and Executor interfaces" "Go Import"

        # External relationships
        anthropic_provider -> claude_api "Sends prompts, receives completions" "HTTPS"
        openai_provider -> openai_api "Sends prompts, receives completions" "HTTPS"
        gemini_provider -> gemini_api "Sends prompts, receives completions" "HTTPS"
        config_loader -> filesystem "Reads ai-provider.yml config files" "File I/O"

        # Internal relationships
        executor -> config_loader "Loads AI configuration"
        executor -> provider_registry "Creates providers via factories"
        executor -> providers "Executes prompts via selected provider"
        provider_registry -> providers "Registers provider factories"

        # Component relationships
        executor_impl -> executor_adapter "Adapted for domain interface"
        loader -> env_substitutor "Substitutes environment variables"
        loader -> schema_validator "Validates config against schema"
        factory_registry -> anthropic_provider "Creates Anthropic provider"
        factory_registry -> openai_provider "Creates OpenAI provider"
        factory_registry -> gemini_provider "Creates Gemini provider"
        factory_registry -> claude_cli_provider "Creates Claude CLI provider"
        factory_registry -> test_provider "Creates test mock provider"
    }

    views {
        systemContext ai_adapter "SystemContext" {
            include *
            autoLayout lr
            title "AI Adapter - System Context"
            description "Shows AI adapter with consumers, LLM services, and configuration"
        }

        container ai_adapter "Containers" {
            include *
            autoLayout tb
            title "AI Adapter - Package Architecture"
            description "Shows AI adapter internal structure"
        }

        component providers "ProvidersComponents" {
            include *
            autoLayout tb
            title "Providers - Components"
            description "Concrete AI provider implementations"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
