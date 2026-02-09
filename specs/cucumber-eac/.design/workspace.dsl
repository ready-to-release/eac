workspace "Cucumber Adapter" "TypeScript BDD test execution via cucumber-js with npm isolation and tag filter translation" {

    model {
        # Consumers
        eac_cli = softwareSystem "EAC CLI" "Test commands dispatching to cucumber runner" "Dependent"
        clibase = softwareSystem "Clibase" "Test runner registry" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Config, testing primitives" "Dependency"
        npm_adapter = softwareSystem "NPM Adapter" "NPM isolation for parallel test safety" "Dependency"

        # External systems
        cucumber_js = softwareSystem "cucumber-js" "TypeScript BDD test framework" "External"
        node_runtime = softwareSystem "Node.js" "JavaScript runtime" "External"

        # Cucumber adapter system
        cucumber_adapter = softwareSystem "Cucumber Adapter" "TypeScript BDD test execution via cucumber-js with npm isolation for parallel safety and tag filter translation" {

            # Runner
            runner = container "Cucumber Runner" "TestTypeRunner implementation for tscucumber type" "Go Package" {
                runner_impl = component "TsCucumberRunner" "Execute method with npm isolation, install, and cucumber-js execution" "Go"
                test_info = component "Test Info" "GetTestInfo and FindTestRoot for TypeScript features" "Go"
                type_registration = component "Type Registration" "Registers tscucumber type with BDD descriptor" "Go"
            }

            # Tag Translator
            tag_translator = container "Tag Translator" "Translates core.TagFilter to cucumber-js expression syntax" "Go Package" {
                translator_impl = component "CucumberTagTranslator" "AND combined, not @tag, (or) syntax translation" "Go"
            }
        }

        # Consumer relationships
        eac_cli -> cucumber_adapter "Dispatches TypeScript BDD tests" "Go Import"
        clibase -> cucumber_adapter "Registered as test type runner" "Go Import"

        # Dependency relationships
        cucumber_adapter -> core "Uses config, testing types" "Go Import"
        cucumber_adapter -> npm_adapter "Uses NpmIsolation for parallel safety" "Go Import"

        # External relationships
        runner -> cucumber_js "Executes cucumber-js with tag expressions" "CLI"
        runner -> node_runtime "Runs via Node.js" "CLI"

        # Internal relationships
        runner -> tag_translator "Translates tag filters to cucumber syntax"

        # Component relationships
        runner_impl -> test_info "Extracts TypeScript test metadata"
        runner_impl -> type_registration "Registered as tscucumber type"
    }

    views {
        systemContext cucumber_adapter "SystemContext" {
            include *
            autoLayout lr
            title "Cucumber Adapter - System Context"
            description "Shows cucumber adapter with consumers and external systems"
        }

        container cucumber_adapter "Containers" {
            include *
            autoLayout tb
            title "Cucumber Adapter - Package Architecture"
            description "Shows cucumber adapter internal structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
