workspace "MCP PowerShell Server Architecture" "Model Context Protocol server for PowerShell command execution" {

    model {
        # External actors and systems
        claude_desktop = person "Claude Desktop" "Claude Code desktop application consuming MCP server" "External"
        powershell = softwareSystem "PowerShell" "Windows PowerShell Core (pwsh) for command execution" "External"
        powershell_modules = softwareSystem "PowerShell Modules" "Installed PowerShell modules and cmdlets" "External"

        # Main MCP PowerShell system
        mcp_pwsh_system = softwareSystem "MCP PowerShell Server" "Model Context Protocol server providing PowerShell command execution capabilities" {

            # Core containers
            mcp_server = container "MCP Server" "JSON-RPC server handling MCP protocol: initialize, tools/list, and tools/call over stdio transport." "Go" "Core" {
                requestHandler = component "Request Handler" "Handles MCP requests via JSON-RPC" "Go"
                jsonParser = component "JSON Parser" "Parses JSON-RPC messages from stdin" "Go (encoding/json)"
                responseEncoder = component "Response Encoder" "Encodes responses to JSON-RPC" "Go (encoding/json)"
                errorHandler = component "Error Handler" "Formats MCP error responses" "Go"
            }

            tool_registry = container "Tool Registry" "Registers and provides metadata for available MCP tools: execute-pwsh and get-pwsh-modules." "Go" "Core" {
                toolDefinitions = component "Tool Definitions" "Defines tool schemas and descriptions" "Go structs"
                schemaValidator = component "Schema Validator" "Validates tool arguments against schemas" "Go"
            }

            pwsh_executor = container "PowerShell Executor" "Executes PowerShell commands via pwsh subprocess, captures output/errors, and handles timeouts." "Go (os/exec)" "Execution" {
                commandExecutor = component "Command Executor" "Executes pwsh with -NoProfile -NonInteractive" "Go (os/exec)"
                outputCapture = component "Output Capture" "Captures combined stdout/stderr" "Go"
                contextManager = component "Context Manager" "Manages execution context and cancellation" "Go (context)"
            }

            module_inspector = container "Module Inspector" "Lists available PowerShell modules using Get-Module -ListAvailable and formats output as JSON." "Go" "Tool" {
                moduleQuery = component "Module Query" "Executes Get-Module cmdlet" "Go"
                jsonFormatter = component "JSON Formatter" "Formats module list as JSON" "Go"
            }

            stdio_transport = container "Stdio Transport" "Provides stdin/stdout transport for JSON-RPC communication using line-delimited JSON." "Go (bufio)" "Infrastructure" {
                stdinScanner = component "Stdin Scanner" "Scans stdin for JSON-RPC messages" "Go (bufio.Scanner)"
                stdoutWriter = component "Stdout Writer" "Writes JSON-RPC responses to stdout" "Go (json.Encoder)"
            }

        }

        # User interactions
        claude_desktop -> mcp_pwsh_system "Invokes PowerShell tools" "JSON-RPC over stdio"

        # External system relationships
        mcp_pwsh_system -> powershell "Executes PowerShell commands" "Process execution"
        pwsh_executor -> powershell_modules "Access installed modules" "PowerShell cmdlets"

        # Internal relationships

        # Stdio Transport relationships
        stdio_transport -> mcp_server "Delivers JSON-RPC messages" "Function calls"
        mcp_server -> stdio_transport "Sends JSON-RPC responses" "Function calls"

        # MCP Server relationships
        mcp_server -> tool_registry "Lists available tools" "Function calls"
        mcp_server -> pwsh_executor "Executes PowerShell commands" "Function calls"
        mcp_server -> module_inspector "Lists PowerShell modules" "Function calls"

        # Tool Registry relationships
        tool_registry -> pwsh_executor "Validates command arguments" "Function calls"

        # PowerShell Executor relationships
        pwsh_executor -> powershell "Spawns pwsh subprocess" "os/exec"

        # Module Inspector relationships
        module_inspector -> pwsh_executor "Executes Get-Module cmdlet" "Function calls"

        # Component relationships

        # MCP Server components
        jsonParser -> requestHandler "Parses incoming requests" "Go function calls"
        requestHandler -> responseEncoder "Formats responses" "Go function calls"
        requestHandler -> errorHandler "Handles errors" "Go function calls"

        # Tool Registry components
        toolDefinitions -> schemaValidator "Provides schemas for validation" "Go function calls"

        # PowerShell Executor components
        contextManager -> commandExecutor "Provides execution context" "Go function calls"
        commandExecutor -> outputCapture "Captures command output" "Go function calls"

        # Module Inspector components
        moduleQuery -> jsonFormatter "Formats module data" "Go function calls"

        # Stdio Transport components
        stdinScanner -> stdoutWriter "Bidirectional message flow" "Go function calls"
    }

    views {
        # System Context view
        systemContext mcp_pwsh_system "SystemContext" {
            include *
            autoLayout lr
            title "MCP PowerShell Server - System Context"
            description "Shows MCP server with Claude Desktop and PowerShell"
        }

        # Container view
        container mcp_pwsh_system "Containers" {
            include *
            autoLayout tb
            title "MCP PowerShell Server - Container Architecture"
            description "Shows all containers and their relationships"
        }

        # Component views for key containers

        component mcp_server "MCPServerComponents" {
            include *
            autoLayout tb
            title "MCP Server - Components"
            description "JSON-RPC request handling components"
        }

        component tool_registry "ToolRegistryComponents" {
            include *
            autoLayout tb
            title "Tool Registry - Components"
            description "Tool definition and validation components"
        }

        component pwsh_executor "PowerShellExecutorComponents" {
            include *
            autoLayout tb
            title "PowerShell Executor - Components"
            description "Command execution and output capture components"
        }

        component stdio_transport "StdioComponents" {
            include *
            autoLayout tb
            title "Stdio Transport - Components"
            description "Stdin/stdout message transport components"
        }

        # Filtered views

        container mcp_pwsh_system "CoreContainers" {
            include ->mcp_server->
            include ->tool_registry->
            include ->stdio_transport->
            autoLayout lr
            title "Core Containers"
            description "Core MCP protocol containers"
        }

        container mcp_pwsh_system "ExecutionContainers" {
            include ->pwsh_executor->
            include ->module_inspector->
            autoLayout lr
            title "Execution Containers"
            description "PowerShell execution containers"
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
            element "External" {
                background #999999
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
                shape roundedbox
            }
            element "Core" {
                background #2E7D32
                color #ffffff
            }
            element "Execution" {
                background #1976D2
                color #ffffff
            }
            element "Tool" {
                background #5E35B1
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
            relationship "Function calls" {
                thickness 2
                color #2E7D32
            }
            relationship "JSON-RPC over stdio" {
                thickness 3
                color #5E35B1
                style dashed
            }
            relationship "Process execution" {
                thickness 3
                color #1976D2
                style dashed
            }
            relationship "PowerShell cmdlets" {
                thickness 2
                color #D32F2F
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
