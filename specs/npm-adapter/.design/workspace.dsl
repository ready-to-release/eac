workspace "NPM Adapter" "Isolated npm environments for parallel TypeScript test execution with incremental file syncing" {

    model {
        # Consumers
        cucumber_adapter = softwareSystem "Cucumber Adapter" "Uses npm isolation for parallel cucumber-js tests" "Dependent"
        mocha_adapter = softwareSystem "Mocha Adapter" "Uses npm isolation for parallel mocha tests" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Workspace paths" "Dependency"
        clibase = softwareSystem "Clibase" "RemoveAllWithRetry utility" "Dependency"

        # External systems
        npm = softwareSystem "npm" "Node package manager" "External"
        filesystem = softwareSystem "File System" "Source files, package.json, node_modules" "External"

        # NPM adapter system
        npm_adapter = softwareSystem "NPM Adapter" "Solves Windows EPERM errors and parallel test interference by providing isolated npm work directories with incremental file syncing" {

            # Isolation
            isolation = container "NPM Isolation" "Manages isolated work directories per module" "Go Package" {
                isolation_manager = component "NpmIsolation" "NewNpmIsolation, PrepareIsolatedEnv entry points" "Go"
                directory_sync = component "Directory Sync" "Incremental file sync (mtime/size comparison)" "Go"
                file_copy = component "File Copy" "copyFileIfChanged for individual files" "Go"
                package_detector = component "Package Change Detector" "Detects package.json changes for full reset" "Go"
                install_mutex = component "Install Mutex" "NpmInstallMu serializing npm install calls" "Go"
            }
        }

        # Consumer relationships
        cucumber_adapter -> npm_adapter "Isolates test environments" "Go Import"
        mocha_adapter -> npm_adapter "Isolates test environments" "Go Import"

        # Dependency relationships
        npm_adapter -> core "Uses workspace paths" "Go Import"
        npm_adapter -> clibase "Uses RemoveAllWithRetry" "Go Import"

        # External relationships
        isolation -> filesystem "Creates .cache/eac/npm/work/ directories, syncs files" "File I/O"
        isolation -> npm "npm ci runs in isolated directories" "CLI"

        # Component relationships
        isolation_manager -> directory_sync "Syncs src/, test/, features/, steps/"
        isolation_manager -> file_copy "Copies package.json, tsconfig.json, .mocharc.json"
        isolation_manager -> package_detector "Checks for full reset needed"
        isolation_manager -> install_mutex "Serializes npm install calls"
    }

    views {
        systemContext npm_adapter "SystemContext" {
            include *
            autoLayout lr
            title "NPM Adapter - System Context"
            description "Shows NPM adapter with test runner consumers and external systems"
        }

        container npm_adapter "Containers" {
            include *
            autoLayout tb
            title "NPM Adapter - Package Architecture"
            description "Shows NPM isolation internal structure"
        }

        component isolation "IsolationComponents" {
            include *
            autoLayout tb
            title "NPM Isolation - Components"
            description "Isolated environment preparation with incremental sync"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
