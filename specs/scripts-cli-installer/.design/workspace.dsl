workspace "R2R CLI Installer Scripts" "Cross-platform installer scripts for R2R CLI" {

    model {
        # External systems
        user = person "Developer" "Developer installing R2R CLI"
        github_releases = softwareSystem "GitHub Releases" "R2R CLI release artifacts" "External"
        filesystem = softwareSystem "File System" "Local installation directory" "External"

        # Dependencies (from module contracts - used for testing)
        eac_core = softwareSystem "EAC Core" "Core domain libraries for testing scripts" "Dependency"

        # Installer system
        installer = softwareSystem "CLI Installer Scripts" "Cross-platform scripts for downloading and installing R2R CLI" {

            pwsh_installer = container "PowerShell Installer" "Windows/cross-platform PowerShell installer" "PowerShell" {
                download_handler = component "Download Handler" "Downloads release from GitHub" "PowerShell"
                platform_detector = component "Platform Detector" "Detects OS and architecture" "PowerShell"
                path_manager = component "Path Manager" "Adds to system PATH" "PowerShell"
            }

            bash_installer = container "Bash Installer" "Unix/Linux/macOS bash installer" "Bash" {
                download_handler_sh = component "Download Handler" "Downloads release using curl/wget" "Bash"
                platform_detector_sh = component "Platform Detector" "Detects OS and architecture" "Bash"
                path_manager_sh = component "Path Manager" "Updates shell profile PATH" "Bash"
            }
        }

        # Dependency relationships (from module contracts)
        installer -> eac_core "Tested using core test infrastructure" "Go Import"

        # User and external relationships
        user -> installer "Runs installer script"
        installer -> github_releases "Downloads CLI binary" "HTTPS"
        installer -> filesystem "Installs binary" "File I/O"

        download_handler -> github_releases "Fetches release"
        platform_detector -> download_handler "Provides platform info"
        path_manager -> filesystem "Updates PATH"

        download_handler_sh -> github_releases "Fetches release"
        platform_detector_sh -> download_handler_sh "Provides platform info"
        path_manager_sh -> filesystem "Updates shell profile"
    }

    views {
        systemContext installer "SystemContext" {
            include *
            autoLayout lr
            title "CLI Installer - System Context"
        }

        container installer "Containers" {
            include *
            autoLayout tb
            title "CLI Installer - Scripts"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
