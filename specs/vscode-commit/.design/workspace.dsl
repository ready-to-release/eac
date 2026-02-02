workspace "VSCode Git Commit Extension" "VSCode extension providing Git toolbar and commit UI" {

    model {
        # External systems
        vscode = softwareSystem "VSCode" "Visual Studio Code editor" "External"
        git = softwareSystem "Git" "Version control system" "External"
        eac_commands = softwareSystem "EAC Commands" "CLI for commit message generation" "External"

        # Extension system
        vscode_ext = softwareSystem "VSCode Git Commit Extension" "Extension providing Git toolbar integration and commit UI" {

            extension = container "Extension" "TypeScript VSCode extension" "TypeScript" {
                extension_main = component "Extension Main" "Extension activation and command registration" "TypeScript"
                status_bar = component "Stable Status Bar" "Non-flickering status bar management" "TypeScript"
                progress_buffer = component "Progress Frame Buffer" "Smooth progress indicator updates" "TypeScript"
            }
        }

        # Relationships
        vscode -> vscode_ext "Loads extension"
        vscode_ext -> git "Executes git commands" "CLI"
        vscode_ext -> eac_commands "Generates commit messages" "CLI"

        extension_main -> status_bar "Updates status"
        extension_main -> progress_buffer "Shows progress"
        status_bar -> vscode "Renders in status bar"
    }

    views {
        systemContext vscode_ext "SystemContext" {
            include *
            autoLayout lr
            title "VSCode Git Commit Extension - System Context"
        }

        container vscode_ext "Containers" {
            include *
            autoLayout tb
            title "VSCode Extension - Container"
        }

        component extension "ExtensionComponents" {
            include *
            autoLayout tb
            title "Extension - Components"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
