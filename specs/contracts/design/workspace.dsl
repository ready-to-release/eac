workspace "Contracts Module Architecture" "Contract definitions and validation for deployable units in the EAC framework" {

    model {
        # External actors and systems
        developer = person "Developer" "Developer defining module contracts and dependencies"
        cli_system = softwareSystem "CLI System" "Consumes contract definitions to manage modules" "External"

        # Main contracts system
        contracts_system = softwareSystem "Contracts Module" "Contract definitions and validation for deployable units" {

            # Core containers
            contract_loader = container "Contract Loader" "Loads and parses YAML contract files, validates schema, handles errors, and provides contract instances to other systems." "Go (YAML)" "Core" {
                yamlParser = component "YAML Parser" "Parses YAML contract files into Go structs" "Go (gopkg.in/yaml.v3)"
                schemaValidator = component "Schema Validator" "Validates contract structure against schema" "Go"
                fileLoader = component "File Loader" "Locates and reads contract files from filesystem" "Go"
            }

            contract_types = container "Contract Types" "Defines contract interfaces, base contract structure, versioning configuration, and source file patterns." "Go" "Core" {
                contractInterface = component "Contract Interface" "Generic contract interface with getters" "Go interface"
                baseContract = component "Base Contract" "Common contract fields (moniker, name, type)" "Go struct"
                versioningType = component "Versioning Type" "Version scheme configuration" "Go struct"
                sourceType = component "Source Type" "Source patterns and ownership rules" "Go struct"
            }

            error_handler = container "Error Handler" "Provides structured error types for contract loading failures, validation errors, and file system errors." "Go" "Core" {
                contractError = component "Contract Error" "Error type for contract-specific failures" "Go struct"
                validationError = component "Validation Error" "Error type for schema validation failures" "Go struct"
            }

            modules_contracts = container "Modules Contracts" "Stores module contract YAML definitions with moniker, name, type, description, dependencies, and source patterns." "YAML Files" "Data" {
                moduleYAML = component "Module YAML Files" "Individual contract files per module" "YAML"
            }

            reports_contracts = container "Reports Contracts" "Manages report generation contracts and validation for test results and summaries." "Go" "Integration"

        }

        # User interactions
        developer -> contracts_system "Defines module contracts" "YAML files"
        cli_system -> contracts_system "Loads contracts for module management" "Go package import"

        # Internal relationships

        # Contract Loader relationships
        contract_loader -> contract_types "Creates contract instances" "Function calls"
        contract_loader -> error_handler "Reports loading errors" "Function calls"
        contract_loader -> modules_contracts "Reads YAML contract files" "File I/O"

        # Component relationships within Contract Loader
        fileLoader -> yamlParser "Passes file content for parsing" "Go function calls"
        yamlParser -> schemaValidator "Validates parsed contracts" "Go function calls"

        # Contract Types internal relationships
        baseContract -> contractInterface "Implements contract interface" "Go interface implementation"
        baseContract -> versioningType "Contains versioning configuration" "Go struct composition"
        baseContract -> sourceType "Contains source configuration" "Go struct composition"

        # Error Handler relationships
        contract_loader -> contractError "Creates contract loading errors" "Go function calls"
        contract_loader -> validationError "Creates validation errors" "Go function calls"

        # Modules Contracts relationships
        modules_contracts -> contract_loader "Consumed by loader" "File reads"

        # Reports Contracts relationships
        reports_contracts -> contract_types "Uses base contract types" "Go package import"
    }

    views {
        # System Context view
        systemContext contracts_system "SystemContext" {
            include *
            autoLayout lr
            title "Contracts Module - System Context"
            description "Shows the Contracts module in context with external actors and systems"
        }

        # Container view
        container contracts_system "Containers" {
            include *
            autoLayout tb
            title "Contracts Module - Container Architecture"
            description "Shows the 5 core containers and their relationships"
        }

        # Component views for key containers

        component contract_loader "LoaderComponents" {
            include *
            autoLayout tb
            title "Contract Loader - Components"
            description "YAML parsing and validation components"
        }

        component contract_types "TypesComponents" {
            include *
            autoLayout tb
            title "Contract Types - Components"
            description "Contract structure and type definitions"
        }

        component error_handler "ErrorComponents" {
            include *
            autoLayout tb
            title "Error Handler - Components"
            description "Error types for contract operations"
        }

        # Filtered views

        container contracts_system "CoreContainers" {
            include ->contract_loader->
            include ->contract_types->
            include ->error_handler->
            autoLayout lr
            title "Core Containers"
            description "Core containers for contract loading and validation"
        }

        # Styles
        styles {
            element "Person" {
                shape person
                background #08427b
                color #ffffff
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "External" {
                background #999999
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
                shape roundedbox
            }
            element "Core" {
                background #2E7D32
                color #ffffff
            }
            element "Data" {
                background #F57C00
                color #ffffff
            }
            element "Integration" {
                background #5E35B1
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
                shape component
            }
            relationship "Relationship" {
                thickness 2
                color #707070
                style solid
            }
            relationship "Function calls" {
                thickness 2
                color #2E7D32
            }
            relationship "Go package import" {
                thickness 2
                color #1976D2
                style dashed
            }
            relationship "File I/O" {
                thickness 2
                color #D32F2F
                style dashed
            }
        }

        # Theme
        theme default
    }

    configuration {
        scope softwaresystem
    }

}
