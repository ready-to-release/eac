workspace "Repository Module" "Root configuration module defining contracts and validation rules" {

    model {
        # External systems
        go_toolchain = softwareSystem "Go Toolchain" "Compiles Go modules" "External"

        # Dependencies (from module contracts)
        eac_core = softwareSystem "EAC Core" "Core domain libraries for validation" "Dependency"
        clie_config = softwareSystem "CLIE Config" "Repository configuration definitions" "Dependency"

        # Repository module system
        repository = softwareSystem "Repository Module" "Defines repository contracts, validation rules, and module ownership" {

            contracts = container "Contracts" "YAML contract definitions" "YAML" {
                modules_yml = component "repository.yml" "Module definitions and ownership" "YAML"
                environments_yml = component "environments.yml" "Environment configurations" "YAML"
                module_types_yml = component "module-types.yml" "Module type definitions" "YAML"
                test_suites_yml = component "test-suites.yml" "Test suite definitions" "YAML"
                testing_tags_yml = component "testing-tags.yml" "Test tag contracts" "YAML"
                system_deps_yml = component "system-dependencies.yml" "External system dependencies" "YAML"
            }

            validation_rules = container "Validation Rules" "BDD specifications for repository integrity" "Gherkin" {
                module_hierarchy = component "Module Hierarchy Valid" "Validates dependency graph structure" "Feature"
                module_isolation = component "Module Isolation" "Ensures modules own their files" "Feature"
                validate_deps = component "Validate Dependencies" "Checks go.mod against contracts" "Feature"
                go_modules_tidy = component "Go Modules Tidy" "Ensures go.mod files are tidy" "Feature"
                test_tags = component "Test Tags Contracted" "Validates test tags are defined" "Feature"
                no_unused_steps = component "No Unused Steps" "Detects orphaned step definitions" "Feature"
                markdown_syntax = component "Markdown Syntax" "Validates markdown formatting" "Feature"
                feature_tags = component "Feature Level Tags" "Enforces tag placement rules" "Feature"
            }
        }

        # Dependency relationships (from module contracts)
        repository -> eac_core "Uses validation libraries" "Go Import"
        repository -> clie_config "Reads configuration from" "File I/O"

        # External system relationships
        validation_rules -> go_toolchain "Validates Go modules"
    }

    views {
        systemContext repository "SystemContext" {
            include *
            autoLayout lr
            title "Repository Module - System Context"
        }

        container repository "Containers" {
            include *
            autoLayout tb
            title "Repository Module - Containers"
        }

        component contracts "ContractsComponents" {
            include *
            autoLayout tb
            title "Contracts - YAML Definitions"
        }

        component validation_rules "ValidationComponents" {
            include *
            autoLayout tb
            title "Validation Rules - BDD Specifications"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
