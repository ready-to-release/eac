workspace "VSCode Configuration" "VSCode workspace settings and extensions" {

    model {
        # External systems
        vscode = softwareSystem "VSCode" "Visual Studio Code editor" "Consumer"
        developers = person "Developers" "Team members using VSCode" "External"

        # VSCode config system
        vscode_config = softwareSystem "VSCode Configuration" "Shared workspace settings and recommended extensions" {

            settings = container "Workspace Settings" "Shared editor configuration" "JSON" {
                settings_json = component "settings.json" "Shared workspace settings" "JSON"
                settings_local = component "settings.local.json.template" "Template for personal settings" "JSON"
            }

            keybindings = container "Keybindings" "Custom keyboard shortcuts" "JSON" {
                keybindings_json = component "keybindings.json" "Custom key mappings" "JSON"
            }

            tasks = container "Tasks" "Build and run tasks" "JSON" {
                tasks_json = component "tasks.json" "Workspace task definitions" "JSON"
            }

            extensions = container "Extensions" "Recommended and bundled extensions" "JSON" {
                extensions_json = component "extensions.json" "Recommended extensions list" "JSON"
                vscode_ext_commit = component "Bundled Commit Extension" "Pre-built commit extension" "TypeScript"
            }

            styles = container "Editor Styles" "Custom styling" "CSS" {
                markdown_preview = component "Markdown Preview" "Dark theme for markdown preview" "CSS"
            }
        }

        # Relationships
        developers -> vscode_config "Uses workspace settings"
        vscode -> vscode_config "Loads configuration from"
        extensions -> vscode "Installs into"
    }

    views {
        systemContext vscode_config "SystemContext" {
            include *
            autoLayout lr
            title "VSCode Configuration - System Context"
        }

        container vscode_config "Containers" {
            include *
            autoLayout tb
            title "VSCode Configuration - Structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
