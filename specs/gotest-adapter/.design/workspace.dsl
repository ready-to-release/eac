workspace "GoTest Adapter" "Executes Go unit tests and godog BDD tests with JSON event streaming and CTRF report generation" {

    model {
        # Consumers
        eac_cli = softwareSystem "EAC CLI" "Test commands dispatching to GoTest runner" "Dependent"
        clibase = softwareSystem "Clibase" "Test runner registry" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Config, testing primitives" "Dependency"
        godog_adapter = softwareSystem "Godog Adapter" "Tag translator for godog tests" "Dependency"

        # External systems
        go_test = softwareSystem "go test" "Go test binary with JSON output" "External"
        filesystem = softwareSystem "File System" "Test source files and CTRF reports" "External"

        # GoTest adapter system
        gotest_adapter = softwareSystem "GoTest Adapter" "Executes go test unit tests and godog BDD tests with JSON event streaming, parsing, and CTRF report generation" {

            # Runner
            runner = container "GoTest Runner" "TestTypeRunner implementation for gotest and godog types" "Go Package" {
                runner_impl = component "GoTestRunner" "Execute method orchestrating test execution" "Go"
                type_registration = component "Type Registration" "Registers gotest and godog types with fallback" "Go"
                test_info = component "Test Info" "GetTestInfo and FindTestRoot metadata extraction" "Go"
            }

            # Runner Helpers
            helpers = container "Runner Helpers" "Module lookup, feature extraction, go generate support" "Go Package" {
                module_lookup = component "Module Lookup" "Finds module for test package path" "Go"
                feature_extractor = component "Feature Extractor" "Extracts feature file paths for godog" "Go"
                go_generate = component "Go Generate" "Runs go generate before test execution" "Go"
            }

            # CTRF Converter
            ctrf_converter = container "CTRF Converter" "Transforms go test JSON events to standardized CTRF format" "Go Package" {
                event_parser = component "Event Parser" "Parses go test -json TestStart, TestPass, TestFail events" "Go"
                ctrf_mapper = component "CTRF Mapper" "Maps test events to CTRF Test entries" "Go"
            }
        }

        # Consumer relationships
        eac_cli -> gotest_adapter "Dispatches test execution" "Go Import"
        clibase -> gotest_adapter "Registered as test type runner" "Go Import"

        # Dependency relationships
        gotest_adapter -> core "Uses config, testing types" "Go Import"
        gotest_adapter -> godog_adapter "Uses GodogTagTranslator" "Go Import"

        # External relationships
        runner -> go_test "Executes go test -json" "CLI"
        ctrf_converter -> filesystem "Writes CTRF JSON reports" "File I/O"

        # Internal relationships
        runner -> helpers "Uses module lookup and feature extraction"
        runner -> ctrf_converter "Converts test output to CTRF"

        # Component relationships
        runner_impl -> type_registration "Registered as gotest and godog types"
        runner_impl -> test_info "Extracts test metadata"
        module_lookup -> feature_extractor "Finds features for modules"
        event_parser -> ctrf_mapper "Feeds parsed events for mapping"
    }

    views {
        systemContext gotest_adapter "SystemContext" {
            include *
            autoLayout lr
            title "GoTest Adapter - System Context"
            description "Shows GoTest adapter with consumers and external systems"
        }

        container gotest_adapter "Containers" {
            include *
            autoLayout tb
            title "GoTest Adapter - Package Architecture"
            description "Shows GoTest adapter internal structure"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
