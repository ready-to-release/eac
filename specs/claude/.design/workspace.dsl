workspace "Claude Code Configuration" "Claude Code IDE integration and customization" {

    model {
        # External systems
        claude_code = softwareSystem "Claude Code" "AI coding assistant" "Consumer"
        github_api = softwareSystem "GitHub API" "GitHub integration" "External"
        filesystem = softwareSystem "File System" "Local repository files" "External"

        # Claude config system
        claude_config = softwareSystem "Claude Code Configuration" "Custom commands, hooks, and settings for Claude Code" {

            settings = container "Settings" "Claude Code behavior configuration" "JSON" {
                settings_json = component "settings.json" "Shared team settings" "JSON"
                settings_local = component "settings.local.json" "Personal local settings" "JSON"
            }

            commands = container "Custom Commands" "Slash command definitions" "Markdown" {
                boot_cmd = component "Boot Command" "Initializes Claude with project context" "Markdown"
                example_cmd = component "Example Command" "Template for custom commands" "Markdown"
            }

            agents = container "Custom Agents" "Specialized agent definitions" "Markdown" {
                boot_agent = component "Boot Agent" "Project initialization agent" "Markdown"
            }

            hooks = container "Hooks" "Event-driven automation scripts" "PowerShell" {
                session_start = component "Session Start Hook" "Runs on session initialization" "PowerShell"
                tool_use = component "Tool Use Hook" "Runs before tool execution" "PowerShell"
                file_write = component "File Write Hook" "Runs after file modifications" "PowerShell"
            }

            mcp_servers = container "MCP Servers" "Model Context Protocol servers" "Binary" {
                mcp_gh = component "GitHub MCP Server" "GitHub integration via MCP" "Go Binary"
            }

            logs = container "Session Logs" "Claude Code session logging" "JSONL" {
                session_logs = component "Session Logs" "JSONL session transcripts" "JSONL"
            }
        }

        # Relationships
        claude_code -> claude_config "Loads configuration from"
        hooks -> filesystem "Modifies files"
        hooks -> claude_code "Provides feedback to"
        mcp_servers -> github_api "Integrates with"
        commands -> claude_code "Extends capabilities of"
    }

    views {
        systemContext claude_config "SystemContext" {
            include *
            autoLayout lr
            title "Claude Code Configuration - System Context"
        }

        container claude_config "Containers" {
            include *
            autoLayout tb
            title "Claude Code Configuration - Structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
