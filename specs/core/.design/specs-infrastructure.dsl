workspace "EAC Specs Infrastructure" "Shared godog test runner, context, and common steps for BDD specifications" {

    model {
        # External systems
        godog = softwareSystem "Godog" "Cucumber/Gherkin test framework for Go" "External"
        eac_core = softwareSystem "EAC Core" "Core domain libraries" "External"
        filesystem = softwareSystem "File System" "Feature files and test assets" "External"

        # Specs infrastructure system
        eac_specs = softwareSystem "EAC Specs Infrastructure" "Shared BDD test infrastructure for all EAC modules" {

            # Internal shared infrastructure
            internal = container "Internal Package" "Shared test context, helpers, and step definitions" "Go Package" {
                test_context = component "Test Context" "Manages test state and isolated directories" "Go"
                test_runner = component "Test Runner" "Configures and runs godog test suites" "Go"
                common_steps = component "Common Steps" "Shared step definitions (Given/When/Then)" "Go"
                test_helpers = component "Test Helpers" "File creation, command execution utilities" "Go"
            }

            # Implementation modules (impl/)
            impl = container "Implementation Modules" "Per-module step definitions and feature tests" "Go Packages" {
                eac_commands_impl = component "EAC Commands Steps" "Steps for eac-commands features" "Go"
                eac_core_impl = component "EAC Core Steps" "Steps for eac-core features" "Go"
                clie_cli_impl = component "CLIE CLI Steps" "Steps for clie features" "Go"
                repository_impl = component "Repository Steps" "Steps for repository features" "Go"
                github_impl = component "GitHub Steps" "Steps for github workflow features" "Go"
            }
        }

        # External relationships
        eac_specs -> godog "Executes BDD tests" "Go Import"
        eac_specs -> eac_core "Uses test utilities" "Go Import"
        internal -> filesystem "Creates test files" "File I/O"

        # Internal relationships
        impl -> internal "Uses shared infrastructure"
        test_runner -> godog "Configures test execution"
        test_context -> test_helpers "Uses helper functions"
        common_steps -> test_context "Accesses test state"

        # Impl module relationships
        eac_commands_impl -> common_steps "Extends with command steps"
        eac_core_impl -> common_steps "Extends with core steps"
        clie_cli_impl -> common_steps "Extends with CLI steps"
    }

    views {
        systemContext eac_specs "SystemContext" {
            include *
            autoLayout lr
            title "EAC Specs Infrastructure - System Context"
        }

        container eac_specs "Containers" {
            include *
            autoLayout tb
            title "EAC Specs - Package Architecture"
        }

        component internal "InternalComponents" {
            include *
            autoLayout tb
            title "Internal Package - Components"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
