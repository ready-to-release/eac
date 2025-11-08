workspace "MCP Go Servers Architecture" "Model Context Protocol servers implemented in Go for GitHub, Docs, and Structurizr integrations" {

    model {
        # External actors and systems
        claude_desktop = person "Claude Desktop" "Claude Code desktop application consuming MCP servers" "External"
        github_api = softwareSystem "GitHub API" "GitHub REST API for repository operations" "External"
        filesystem = softwareSystem "File System" "Local file system for documentation and design files" "External"
        docker_daemon = softwareSystem "Docker Daemon" "Runs Structurizr and docs containers" "External"

        # Main MCP servers system
        mcp_go_system = softwareSystem "MCP Go Servers" "Model Context Protocol servers providing GitHub, docs, and design integrations" {

            # GitHub MCP Server
            github_mcp = container "GitHub MCP Server" "Provides GitHub repository operations via MCP: view repos, create issues, list PRs, and manage workflow runs." "Go (MCP SDK)" "MCP Server" {
                repoViewer = component "Repository Viewer" "Implements gh-repo-view tool" "Go (MCP tool)"
                issueCreator = component "Issue Creator" "Implements gh-issue-create tool" "Go (MCP tool)"
                prLister = component "PR Lister" "Implements gh-pr-list tool" "Go (MCP tool)"
                runLister = component "Run Lister" "Implements gh-run-list tool" "Go (MCP tool)"
                githubClient = component "GitHub Client" "GitHub API client wrapper" "Go (HTTP)"
            }

            docs_mcp = container "Docs MCP Server" "Provides documentation operations via MCP: search docs, get pages, list docs, build/serve site using Docker." "Go (MCP SDK)" "MCP Server" {
                docsSearcher = component "Docs Searcher" "Implements search-docs tool" "Go (MCP tool)"
                pageGetter = component "Page Getter" "Implements get-doc-page tool" "Go (MCP tool)"
                docsLister = component "Docs Lister" "Implements list-docs tool" "Go (MCP tool)"
                docsBuild = component "Docs Builder" "Implements build-docs tool (Docker)" "Go (MCP tool)"
                docsServer = component "Docs Server" "Implements serve-docs/stop-docs tools (Docker)" "Go (MCP tool)"
                fileReader = component "File Reader" "Reads markdown files from docs/" "Go (os)"
            }

            structurizr_mcp = container "Structurizr MCP Server" "Provides architecture design operations via MCP: serve Structurizr, manage design workspace, and render diagrams." "Go (MCP SDK)" "MCP Server" {
                designServer = component "Design Server" "Serves Structurizr in Docker container" "Go (MCP tool)"
                workspaceManager = component "Workspace Manager" "Manages DSL workspace files" "Go (MCP tool)"
                diagramRenderer = component "Diagram Renderer" "Renders architecture diagrams" "Go (MCP tool)"
            }

            mcp_runtime = container "MCP Runtime" "Go MCP SDK runtime providing stdio transport, tool registration, error handling, and JSON-RPC protocol implementation." "Go (MCP SDK)" "Infrastructure" {
                toolRegistry = component "Tool Registry" "Registers and dispatches MCP tools" "Go (MCP SDK)"
                stdioTransport = component "Stdio Transport" "JSON-RPC over stdio" "Go (MCP SDK)"
                errorHandler = component "Error Handler" "MCP error formatting and handling" "Go (MCP SDK)"
            }

            docker_client = container "Docker Client" "Shared Docker client for container operations in docs and design MCP servers." "Go (Docker SDK)" "Infrastructure" {
                containerManager = component "Container Manager" "Creates, starts, stops containers" "Go (Docker SDK)"
                imageManager = component "Image Manager" "Builds and manages Docker images" "Go (Docker SDK)"
            }
        }

        # User interactions
        claude_desktop -> mcp_go_system "Invokes MCP tools" "JSON-RPC over stdio"

        # External system relationships
        github_mcp -> github_api "Fetches repository data" "HTTPS/REST"
        docs_mcp -> filesystem "Reads documentation files" "File I/O"
        structurizr_mcp -> filesystem "Reads DSL workspace files" "File I/O"
        docs_mcp -> docker_daemon "Runs MkDocs containers" "Docker API"
        structurizr_mcp -> docker_daemon "Runs Structurizr containers" "Docker API"

        # Internal relationships

        # MCP Runtime relationships
        mcp_runtime -> github_mcp "Dispatches GitHub tool calls" "Function calls"
        mcp_runtime -> docs_mcp "Dispatches docs tool calls" "Function calls"
        mcp_runtime -> structurizr_mcp "Dispatches design tool calls" "Function calls"

        # Docker Client relationships
        docs_mcp -> docker_client "Manages docs containers" "Function calls"
        structurizr_mcp -> docker_client "Manages design containers" "Function calls"

        # Component relationships

        # GitHub MCP components
        repoViewer -> githubClient "Fetches repository details" "Go function calls"
        issueCreator -> githubClient "Creates GitHub issues" "Go function calls"
        prLister -> githubClient "Lists pull requests" "Go function calls"
        runLister -> githubClient "Lists workflow runs" "Go function calls"

        # Docs MCP components
        docsSearcher -> fileReader "Searches markdown content" "Go function calls"
        pageGetter -> fileReader "Reads markdown pages" "Go function calls"
        docsLister -> fileReader "Lists available docs" "Go function calls"
        docsBuild -> containerManager "Builds docs in container" "Go function calls"
        docsServer -> containerManager "Serves docs in container" "Go function calls"

        # Structurizr MCP components
        designServer -> containerManager "Runs Structurizr server" "Go function calls"
        workspaceManager -> fileReader "Reads DSL workspace files" "Go function calls"

        # MCP Runtime components
        stdioTransport -> toolRegistry "Receives tool invocations" "JSON-RPC"
        toolRegistry -> errorHandler "Handles tool errors" "Go function calls"
    }

    views {
        # System Context view
        systemContext mcp_go_system "SystemContext" {
            include *
            autoLayout lr
            title "MCP Go Servers - System Context"
            description "Shows MCP servers with Claude Desktop and external systems"
        }

        # Container view
        container mcp_go_system "Containers" {
            include *
            autoLayout tb
            title "MCP Go Servers - Container Architecture"
            description "Shows all Go-based MCP servers and infrastructure"
        }

        # Component views for MCP servers

        component github_mcp "GitHubMCPComponents" {
            include *
            autoLayout tb
            title "GitHub MCP Server - Components"
            description "GitHub integration tools and API client"
        }

        component docs_mcp "DocsMCPComponents" {
            include *
            autoLayout tb
            title "Docs MCP Server - Components"
            description "Documentation tools and file operations"
        }

        component structurizr_mcp "StructurizrMCPComponents" {
            include *
            autoLayout tb
            title "Structurizr MCP Server - Components"
            description "Architecture design tools and workspace management"
        }

        component mcp_runtime "MCPRuntimeComponents" {
            include *
            autoLayout tb
            title "MCP Runtime - Components"
            description "MCP SDK runtime and protocol implementation"
        }

        # Filtered views

        container mcp_go_system "MCPServers" {
            include ->github_mcp->
            include ->docs_mcp->
            include ->structurizr_mcp->
            autoLayout lr
            title "MCP Servers"
            description "All Go-based MCP server containers"
        }

        container mcp_go_system "Infrastructure" {
            include ->mcp_runtime->
            include ->docker_client->
            autoLayout lr
            title "Infrastructure"
            description "Shared infrastructure containers"
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
            element "MCP Server" {
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
            relationship "Function calls" {
                thickness 2
                color #2E7D32
            }
            relationship "JSON-RPC over stdio" {
                thickness 3
                color #5E35B1
                style dashed
            }
            relationship "Docker API" {
                thickness 3
                color #1976D2
                style dashed
            }
            relationship "HTTPS/REST" {
                thickness 2
                color #D32F2F
                style dashed
            }
            relationship "File I/O" {
                thickness 2
                color #F57C00
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
