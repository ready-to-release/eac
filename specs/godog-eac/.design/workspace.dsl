workspace "Godog Adapter" "Shared BDD test infrastructure for godog with test context, fixture management, and cached repository data" {

    model {
        # Consumers
        eac_specs = softwareSystem "EAC Module Specs" "BDD specifications across all modules" "Dependent"
        gotest_adapter = softwareSystem "GoTest Adapter" "Go test runner using godog tag translator" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Config, testing primitives, domain types" "Dependency"
        clibase = softwareSystem "Clibase" "Test runner interfaces" "Dependency"
        contracts_core = softwareSystem "Contracts Core" "Port interfaces" "Dependency"
        godog_framework = softwareSystem "Godog Framework" "cucumber/godog BDD test framework" "External"

        # External systems
        git_system = softwareSystem "Git" "Tracked file listing for test cache" "External"
        filesystem = softwareSystem "File System" "Feature files, fixtures, test workspaces" "External"

        # Godog adapter system
        godog_adapter = softwareSystem "Godog Adapter" "Shared BDD test infrastructure providing test context, fixture pooling, cached repository data, and in-process command dispatch" {

            # Test Context
            test_context = container "Test Context" "Godog-specific test context with isolation and mock config" "Go Package" {
                context_impl = component "TestContext" "Wraps SharedTestContext with godog-specific state" "Go"
                scenario_init = component "Scenario Initializer" "Creates per-scenario test context" "Go"
                mock_config = component "Mock Config" "Mocking environment builder for isolated tests" "Go"
            }

            # Test Cache
            test_cache = container "Test Cache" "Process-level repository cache with thread-safe RWMutex" "Go Package" {
                cache_impl = component "TestCache" "Thread-safe cache with double-checked locking" "Go"
                file_queries = component "File Queries" "FilesByExtension, FilesBySuffix, FilesInDir" "Go"
                ci_optimization = component "CI Optimization" "Pre-computed .git/cached-files.txt lookup" "Go"
            }

            # Step Definitions
            steps = container "Shared Step Definitions" "Reusable step definitions for BDD specs" "Go Package" {
                common_steps = component "Common Steps" "RegisterCommonSteps for config, module, file assertions" "Go"
            }

            # Fixtures
            fixtures = container "Fixtures" "Test environment setup with template-based fast-copy" "Go Package" {
                fixture_pool = component "Fixture Pool" "Template-based fixtures for fast scenario setup" "Go"
                module_setup = component "Module Setup" "CreateGoModule, SetupEACConfig, ApplyTemplate" "Go"
                git_fixtures = component "Git Fixtures" "Git state management for test scenarios" "Go"
            }

            # Templates
            templates = container "Templates" "Named EAC configuration templates with parameter substitution" "Go Package" {
                template_registry = component "Template Registry" "Named templates with {{PLACEHOLDER}} substitution" "Go"
                template_params = component "Template Params" "Parameter map for template rendering" "Go"
            }

            # Dispatcher
            dispatcher = container "In-Process Dispatcher" "Direct command dispatch avoiding subprocess overhead" "Go Package" {
                dispatcher_impl = component "Dispatcher" "MakeInProcessDispatcher for fast BDD execution" "Go"
            }

            # Tag Translator
            tag_translator = container "Tag Translator" "Translates core.TagFilter to godog tag expressions" "Go Package" {
                translator_impl = component "GodogTagTranslator" "AND/OR/NOT tag expression syntax" "Go"
            }

            # Runner Config
            runner_config = container "Runner Config" "Godog runner configuration and option building" "Go Package" {
                config_builder = component "RunnerConfig" "BuildOptions for godog test execution" "Go"
                tag_filter_builder = component "Tag Filter Builder" "Constructs tag filters from env and config" "Go"
            }
        }

        # Consumer relationships
        eac_specs -> godog_adapter "Uses shared test infrastructure" "Go Import"
        gotest_adapter -> godog_adapter "Uses tag translator" "Go Import"

        # Dependency relationships
        godog_adapter -> core "Uses config, testing primitives" "Go Import"
        godog_adapter -> clibase "Uses test runner interfaces" "Go Import"
        godog_adapter -> contracts_core "Uses port interfaces" "Go Import"

        # External relationships
        godog_adapter -> godog_framework "Integrates with godog scenario API" "Go Import"
        test_cache -> git_system "Lists tracked files via git ls-files" "CLI"
        test_cache -> filesystem "Reads .git/cached-files.txt (CI)" "File I/O"
        fixtures -> filesystem "Creates isolated test workspaces" "File I/O"

        # Internal relationships
        test_context -> test_cache "Queries cached repository data"
        test_context -> fixtures "Sets up test environment"
        test_context -> dispatcher "Dispatches commands in-process"
        test_context -> steps "Registers common step definitions"
        fixtures -> templates "Applies named configuration templates"
        runner_config -> tag_translator "Translates tag filters"

        # Component relationships
        context_impl -> scenario_init "Creates per-scenario context"
        context_impl -> mock_config "Builds mocking environment"
        cache_impl -> file_queries "Provides file query methods"
        cache_impl -> ci_optimization "Checks for pre-computed cache"
        fixture_pool -> module_setup "Creates module test fixtures"
        fixture_pool -> git_fixtures "Manages git state"
        template_registry -> template_params "Substitutes template parameters"
        config_builder -> tag_filter_builder "Builds tag filter options"
    }

    views {
        systemContext godog_adapter "SystemContext" {
            include *
            autoLayout lr
            title "Godog Adapter - System Context"
            description "Shows godog adapter with BDD specs, dependencies, and external systems"
        }

        container godog_adapter "Containers" {
            include *
            autoLayout tb
            title "Godog Adapter - Package Architecture"
            description "Shows godog adapter internal structure"
        }

        component test_context "TestContextComponents" {
            include *
            autoLayout tb
            title "Test Context - Components"
            description "Per-scenario test context with isolation"
        }

        component test_cache "TestCacheComponents" {
            include *
            autoLayout tb
            title "Test Cache - Components"
            description "Thread-safe repository data cache"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
