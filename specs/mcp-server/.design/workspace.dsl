workspace "EAC MCP Commands Server" "Model Context Protocol server exposing EAC CLI commands" {

    model {
        # External actors and systems
        claude_desktop = person "Claude Desktop" "Claude Code desktop application consuming MCP server" "Client"
        filesystem = softwareSystem "File System" "Local file system for repository files" "External"

        # Dependencies (from module contracts)
        eac_core = softwareSystem "EAC Core" "Core domain libraries for repository operations" "Dependency"

        # MCP Server
        mcp_server = softwareSystem "EAC MCP Commands Server" "MCP server that dynamically exposes all EAC CLI commands as tools" {

            commands_mcp = container "Commands MCP Server" "Single-file MCP server that discovers and proxies EAC CLI commands via JSON-RPC over stdio." "Go (stdio MCP)" "MCP Server" {
                # JSON-RPC Protocol
                jsonrpc_handler = component "JSON-RPC Handler" "Handles MCP protocol: initialize, tools/list, tools/call" "Go (bufio/json)"
                request_router = component "Request Router" "Routes MCP methods to appropriate handlers" "Go"
                response_encoder = component "Response Encoder" "Encodes JSON-RPC responses to stdout" "Go (json)"

                # Command Discovery
                command_discovery = component "Command Discovery" "Discovers available commands via 'get commands'" "Go"
                tool_generator = component "Tool Generator" "Converts discovered commands to MCP tool definitions" "Go"

                # Command Execution
                cli_executor = component "CLI Executor" "Executes EAC CLI commands via 'go run'" "Go (os/exec)"
                repo_finder = component "Repository Finder" "Locates git repository root for command execution" "Go"
            }
        }

        # User interactions
        claude_desktop -> mcp_server "Invokes MCP tools" "JSON-RPC over stdio"

        # Dependency relationships (from module contracts)
        mcp_server -> eac_core "Uses repository utilities" "Go Import"

        # External system relationships
        commands_mcp -> filesystem "Finds repository root" "File I/O"

        # Component relationships - JSON-RPC flow
        jsonrpc_handler -> request_router "Parses and routes requests" "Function call"
        request_router -> command_discovery "Handles tools/list" "Function call"
        request_router -> cli_executor "Handles tools/call" "Function call"
        request_router -> response_encoder "Sends responses" "Function call"

        # Component relationships - Command Discovery
        command_discovery -> cli_executor "Calls 'get commands'" "Function call"
        command_discovery -> tool_generator "Provides command info" "Function call"

        # Component relationships - Command Execution
        cli_executor -> repo_finder "Gets repository root" "Function call"
    }

    views {
        # System Context view
        systemContext mcp_server "SystemContext" {
            include *
            autoLayout lr
            title "EAC MCP Commands Server - System Context"
            description "Shows MCP server with Claude Desktop and external systems"
        }

        # Container view
        container mcp_server "Containers" {
            include *
            autoLayout tb
            title "EAC MCP Commands Server - Container"
            description "Single container MCP server architecture"
        }

        # Component view
        component commands_mcp "Components" {
            include *
            autoLayout tb
            title "Commands MCP Server - Components"
            description "JSON-RPC protocol, command discovery, and CLI execution"
        }

        # Dynamic view - Tool Discovery (internal flow)
        dynamic commands_mcp "ToolDiscovery" "How tools/list discovers available commands" {
            jsonrpc_handler -> request_router "1. Parse request"
            request_router -> command_discovery "2. Get available tools"
            command_discovery -> cli_executor "3. Execute 'get commands'"
            command_discovery -> tool_generator "4. Convert to MCP tools"
            request_router -> response_encoder "5. Encode response"
            autoLayout
        }

        # Dynamic view - Tool Call (internal flow)
        dynamic commands_mcp "ToolCall" "How tools/call executes a command" {
            jsonrpc_handler -> request_router "1. Parse request"
            request_router -> cli_executor "2. Execute command"
            cli_executor -> repo_finder "3. Find repo root"
            request_router -> response_encoder "4. Encode output"
            autoLayout
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }

}
