workspace "Mocha Adapter" "TypeScript unit test execution via mocha with npm isolation and CTRF report generation" {

    model {
        # Consumers
        eac_cli = softwareSystem "EAC CLI" "Test commands dispatching to mocha runner" "Dependent"
        clibase = softwareSystem "Clibase" "Test runner registry and CTRF format" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Config, testing primitives" "Dependency"
        npm_adapter = softwareSystem "NPM Adapter" "NPM isolation for parallel test safety" "Dependency"

        # External systems
        mocha_framework = softwareSystem "Mocha" "TypeScript unit test framework" "External"
        node_runtime = softwareSystem "Node.js" "JavaScript runtime" "External"

        # Mocha adapter system
        mocha_adapter = softwareSystem "Mocha Adapter" "TypeScript mocha unit test execution with npm isolation for parallel safety and CTRF report generation" {

            # Runner
            runner = container "Mocha Runner" "TestTypeRunner implementation for mocha type" "Go Package" {
                runner_impl = component "MochaRunner" "Execute with npm isolation, install, mocha JSON output, CTRF conversion" "Go"
                test_info = component "Test Info" "GetTestInfo for TypeScript unit tests" "Go"
                type_registration = component "Type Registration" "Registers mocha type with unit test descriptor" "Go"
                module_resolver = component "Module Resolver" "findTsModuleForPath resolves TypeScript module" "Go"
            }

            # CTRF Converter
            ctrf_converter = container "CTRF Converter" "Transforms mocha JSON output to standardized CTRF format" "Go Package" {
                json_parser = component "JSON Parser" "Parses mocha --reporter json output" "Go"
                ctrf_mapper = component "CTRF Mapper" "Maps mocha results to CTRF Test entries" "Go"
            }
        }

        # Consumer relationships
        eac_cli -> mocha_adapter "Dispatches TypeScript unit tests" "Go Import"
        clibase -> mocha_adapter "Registered as test type runner" "Go Import"

        # Dependency relationships
        mocha_adapter -> core "Uses config, testing types" "Go Import"
        mocha_adapter -> npm_adapter "Uses NpmIsolation for parallel safety" "Go Import"

        # External relationships
        runner -> mocha_framework "Executes mocha --reporter json" "CLI"
        runner -> node_runtime "Runs via Node.js" "CLI"

        # Internal relationships
        runner -> ctrf_converter "Converts mocha JSON to CTRF"

        # Component relationships
        runner_impl -> test_info "Extracts test metadata"
        runner_impl -> type_registration "Registered as mocha type"
        runner_impl -> module_resolver "Resolves TypeScript module ownership"
        json_parser -> ctrf_mapper "Feeds parsed results for mapping"
    }

    views {
        systemContext mocha_adapter "SystemContext" {
            include *
            autoLayout lr
            title "Mocha Adapter - System Context"
            description "Shows mocha adapter with consumers and external systems"
        }

        container mocha_adapter "Containers" {
            include *
            autoLayout tb
            title "Mocha Adapter - Package Architecture"
            description "Shows mocha adapter internal structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
