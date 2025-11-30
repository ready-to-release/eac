workspace "MCP Servers Architecture" "Model Context Protocol servers for EAC commands and GitHub integration" {

    model {
        # External actors and systems
        claude_desktop = person "Claude Desktop" "Claude Code desktop application consuming MCP servers" "External"
        github_api = softwareSystem "GitHub API" "GitHub REST and GraphQL APIs for repository operations" "External"
        eac_cli = softwareSystem "EAC CLI" "Simply CLI commands binary for repository management" "External"
        filesystem = softwareSystem "File System" "Local file system for repository files and contracts" "External"

        # Main MCP servers system
        mcp_system = softwareSystem "MCP Servers" "Model Context Protocol servers providing GitHub and EAC commands integrations" {

            # GitHub MCP Server
            github_mcp = container "GitHub MCP Server" "Provides GitHub repository operations via MCP: view repos, create issues, list PRs, and manage workflow runs using gh CLI." "Go (stdio MCP)" "MCP Server" {
                ghRepoView = component "Repository Viewer" "Implements gh-repo-view tool" "Go MCP tool"
                ghIssueCreate = component "Issue Creator" "Implements gh-issue-create tool" "Go MCP tool"
                ghPrList = component "PR Lister" "Implements gh-pr-list tool" "Go MCP tool"
                ghRunList = component "Workflow Run Lister" "Implements gh-run-list tool" "Go MCP tool"
                ghCLI = component "GitHub CLI Wrapper" "Executes gh CLI commands" "Go (os/exec)"
            }

            # Commands MCP Server
            commands_mcp = container "Commands MCP Server" "Provides EAC repository management operations via MCP: modules, dependencies, specs, tests, build, design, docs, templates, and workspace management." "Go (stdio MCP)" "MCP Server" {
                # Module Discovery
                getModules = component "Module Discovery" "Implements get-modules, show-modules, show-moduletypes tools" "Go MCP tool"
                getFiles = component "File Mapper" "Implements get-files, show-files tools" "Go MCP tool"

                # Dependencies
                getDependencies = component "Dependency Analyzer" "Implements get-dependencies, show-dependencies, validate-dependencies tools" "Go MCP tool"
                getExecutionOrder = component "Execution Order" "Implements get-execution-order tool" "Go MCP tool"

                # Build & Test
                buildModule = component "Build Runner" "Implements build-module, build-modules tools" "Go MCP tool"
                testModule = component "Test Runner" "Implements test-module, test-modules, test-suite tools" "Go MCP tool"
                showTests = component "Test Lister" "Implements show-tests, get-tests, show-suite tools" "Go MCP tool"

                # Design & Architecture
                designCreate = component "Design Generator" "Implements create-design tool" "Go MCP tool"
                designValidate = component "Design Validator" "Implements validate-design tool" "Go MCP tool"
                designServe = component "Design Server" "Implements serve-design tool" "Go MCP tool"

                # Specifications
                specsCreate = component "Spec Generator" "Implements create-spec tool" "Go MCP tool"
                specsValidate = component "Spec Validator" "Implements validate-specs tool" "Go MCP tool"

                # Documentation
                docsServe = component "Docs Server" "Implements serve-docs tool" "Go MCP tool"

                # Templates
                templatesList = component "Template Lister" "Implements templates-list tool" "Go MCP tool"
                templatesInstall = component "Template Installer" "Implements templates-install tools" "Go MCP tool"
                templatesApply = component "Template Applicator" "Implements templates-apply tools" "Go MCP tool"

                # Git Operations
                commitAI = component "AI Commit Generator" "Implements commit tool" "Go MCP tool"
                showFilesChanged = component "Changed Files Viewer" "Implements show-files-changed, show-files-staged tools" "Go MCP tool"
                getChangedModules = component "Changed Modules Analyzer" "Implements get-changed-modules tool" "Go MCP tool"

                # Workspace Management
                workCreate = component "Workspace Creator" "Implements work-create tool" "Go MCP tool"
                workList = component "Workspace Lister" "Implements work-list tool" "Go MCP tool"
                workCommit = component "Workspace Committer" "Implements work-commit tool" "Go MCP tool"
                workOps = component "Workspace Operations" "Implements work-pr, work-merge, work-pull, work-remove tools" "Go MCP tool"

                # Core execution
                cliExecutor = component "CLI Executor" "Executes EAC CLI commands" "Go (os/exec)"
            }

            # Shared MCP infrastructure
            mcp_runtime = container "MCP Runtime" "Shared JSON-RPC over stdio transport for all MCP servers." "Go (stdio)" "Infrastructure" {
                stdinReader = component "Stdin Reader" "Reads JSON-RPC requests from stdin" "Go (bufio)"
                stdoutWriter = component "Stdout Writer" "Writes JSON-RPC responses to stdout" "Go (json)"
                requestRouter = component "Request Router" "Routes MCP requests to appropriate tools" "Go"
                errorHandler = component "Error Handler" "Formats MCP error responses" "Go"
            }
        }

        # User interactions
        claude_desktop -> mcp_system "Invokes MCP tools" "JSON-RPC over stdio"

        # External system relationships
        github_mcp -> github_api "Executes GitHub operations" "gh CLI / HTTPS"
        commands_mcp -> eac_cli "Executes repository commands" "CLI subprocess"
        commands_mcp -> filesystem "Reads contracts and files" "File I/O"

        # MCP Runtime relationships
        mcp_runtime -> github_mcp "Dispatches GitHub tool calls" "Function calls"
        mcp_runtime -> commands_mcp "Dispatches command tool calls" "Function calls"

        # GitHub MCP component relationships
        ghRepoView -> ghCLI "Executes gh repo view" "CLI call"
        ghIssueCreate -> ghCLI "Executes gh issue create" "CLI call"
        ghPrList -> ghCLI "Executes gh pr list" "CLI call"
        ghRunList -> ghCLI "Executes gh run list" "CLI call"

        # Commands MCP component relationships (all components use CLI executor)
        getModules -> cliExecutor "Executes: show modules, get modules" "CLI call"
        getFiles -> cliExecutor "Executes: show files, get files" "CLI call"
        getDependencies -> cliExecutor "Executes: show dependencies, validate dependencies" "CLI call"
        getExecutionOrder -> cliExecutor "Executes: get execution-order" "CLI call"
        buildModule -> cliExecutor "Executes: build module" "CLI call"
        testModule -> cliExecutor "Executes: test module, test suite" "CLI call"
        showTests -> cliExecutor "Executes: show tests, get tests" "CLI call"
        designCreate -> cliExecutor "Executes: create design" "CLI call"
        designValidate -> cliExecutor "Executes: validate design" "CLI call"
        designServe -> cliExecutor "Executes: serve design" "CLI call"
        specsCreate -> cliExecutor "Executes: create spec" "CLI call"
        specsValidate -> cliExecutor "Executes: validate specs" "CLI call"
        docsServe -> cliExecutor "Executes: serve docs" "CLI call"
        templatesList -> cliExecutor "Executes: templates list" "CLI call"
        templatesInstall -> cliExecutor "Executes: templates install" "CLI call"
        templatesApply -> cliExecutor "Executes: templates apply" "CLI call"
        commitAI -> cliExecutor "Executes: commit" "CLI call"
        showFilesChanged -> cliExecutor "Executes: show files-changed" "CLI call"
        getChangedModules -> cliExecutor "Executes: get changed-modules" "CLI call"
        workCreate -> cliExecutor "Executes: work create" "CLI call"
        workList -> cliExecutor "Executes: work list" "CLI call"
        workCommit -> cliExecutor "Executes: work commit" "CLI call"
        workOps -> cliExecutor "Executes: work pr/merge/pull/remove" "CLI call"

        # MCP Runtime component relationships
        stdinReader -> requestRouter "Delivers parsed requests" "Function call"
        requestRouter -> stdoutWriter "Sends responses" "Function call"
        requestRouter -> errorHandler "Handles errors" "Function call"
    }

    views {
        # System Context view
        systemContext mcp_system "SystemContext" {
            include *
            autoLayout lr
            title "MCP Servers - System Context"
            description "Shows MCP servers with Claude Desktop and external systems"
        }

        # Container view
        container mcp_system "Containers" {
            include *
            autoLayout tb
            title "MCP Servers - Container Architecture"
            description "Shows all MCP server containers and infrastructure"
        }

        # Component views for MCP servers

        component github_mcp "GitHubMCPComponents" {
            include *
            autoLayout tb
            title "GitHub MCP Server - Components"
            description "GitHub integration tools and gh CLI wrapper"
        }

        component commands_mcp "CommandsMCPComponents" {
            include *
            autoLayout tb
            title "Commands MCP Server - Components"
            description "EAC repository management tools and CLI executor"
        }

        component mcp_runtime "MCPRuntimeComponents" {
            include *
            autoLayout tb
            title "MCP Runtime - Components"
            description "JSON-RPC protocol implementation"
        }

        # Filtered views

        container mcp_system "MCPServers" {
            include ->github_mcp->
            include ->commands_mcp->
            autoLayout lr
            title "MCP Servers"
            description "All MCP server containers"
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
            relationship "Function call" {
                thickness 2
                color #2E7D32
            }
            relationship "JSON-RPC over stdio" {
                thickness 3
                color #5E35B1
                style dashed
            }
            relationship "CLI call" {
                thickness 2
                color #1976D2
                style dashed
            }
            relationship "CLI subprocess" {
                thickness 3
                color #1976D2
                style dashed
            }
            relationship "gh CLI / HTTPS" {
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
