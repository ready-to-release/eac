workspace "Go Command Invoker Scripts" "Cross-shell wrappers with tab completion for Go commands" {

    model {
        # External systems
        developer = person "Developer" "Developer running EAC commands"
        shell = softwareSystem "Shell" "PowerShell or Bash shell" "External"

        # Dependencies (from module contracts)
        eac_commands = softwareSystem "EAC Commands" "Go commands at go/cli/eac/" "Dependency"

        # Invoker system
        invoker = softwareSystem "Go Command Invoker" "Shell wrappers enabling direct command execution with tab completion" {

            pwsh_invoker = container "PowerShell Invoker" "PowerShell wrapper and completion" "PowerShell" {
                command_router = component "Command Router" "Routes commands to go run" "PowerShell"
                tab_completion = component "Tab Completion" "Provides command completion" "PowerShell"
                importer = component "Importer" "Imports functions into shell" "PowerShell"
            }

            bash_invoker = container "Bash Invoker" "Bash wrapper and completion" "Bash" {
                command_router_sh = component "Command Router" "Routes commands to go run" "Bash"
                tab_completion_sh = component "Tab Completion" "Provides command completion" "Bash"
                importer_sh = component "Importer" "Sources functions into shell" "Bash"
            }
        }

        # Dependency relationships (from module contracts)
        invoker -> eac_commands "Invokes commands" "CLI"

        # User and external relationships
        developer -> invoker "Types commands"
        invoker -> shell "Integrates with shell"

        command_router -> eac_commands "Invokes go run ./go/cli/eac"
        tab_completion -> eac_commands "Gets available commands"
        importer -> shell "Adds functions to session"

        command_router_sh -> eac_commands "Invokes go run"
        tab_completion_sh -> eac_commands "Gets available commands"
        importer_sh -> shell "Sources into profile"
    }

    views {
        systemContext invoker "SystemContext" {
            include *
            autoLayout lr
            title "Go Command Invoker - System Context"
        }

        container invoker "Containers" {
            include *
            autoLayout tb
            title "Go Command Invoker - Scripts"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
