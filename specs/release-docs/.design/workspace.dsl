workspace "Release Documentation" "Release notes and changelog management" {

    model {
        # External systems
        github_releases = softwareSystem "GitHub Releases" "Release publishing" "External"
        eac_commands = softwareSystem "EAC Commands" "Generates changelogs" "Consumer"
        developers = person "Developers" "Release managers" "External"

        # Release docs system
        release_docs = softwareSystem "Release Documentation" "Module changelogs and release notes" {

            changelogs = container "Changelogs" "Module-specific changelogs" "Markdown" {
                module_changelogs = component "Module Changelogs" "CHANGELOG.md per module" "Markdown"
            }

            release_process = container "Release Process" "Release workflow documentation" "Markdown" {
                release_checklist = component "Release Checklist" "Pre-release verification steps" "Markdown"
                versioning_guide = component "Versioning Guide" "CalVer/SemVer conventions" "Markdown"
            }
        }

        # Relationships
        developers -> release_docs "Manages releases via"
        eac_commands -> changelogs "Generates and updates"
        release_docs -> github_releases "Published to"
    }

    views {
        systemContext release_docs "SystemContext" {
            include *
            autoLayout lr
            title "Release Documentation - System Context"
        }

        container release_docs "Containers" {
            include *
            autoLayout tb
            title "Release Documentation - Structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
