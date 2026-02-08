workspace "TUI Adapter" "Terminal user interface implementation using bubbletea for interactive parallel task visualization" {

    model {
        # External actors
        developer = person "Developer" "Views build, test, scan progress in terminal"

        # Consumers
        eac_cli = softwareSystem "EAC CLI" "Commands rendering output through TUI" "Dependent"
        clibase = softwareSystem "Clibase" "Orchestrator sending execution events to TUI" "Dependent"

        # Dependencies
        core = softwareSystem "EAC Core" "Logging and domain types" "Dependency"
        tui_contract = softwareSystem "TUI Contract" "Console interface and display types" "Dependency"
        contracts_core = softwareSystem "Contracts Core" "ActionType and ExecutionObserver" "Dependency"

        # External systems
        terminal_sys = softwareSystem "Terminal" "stdout/stderr, TTY capabilities, terminal size" "External"

        # TUI adapter system
        tui_adapter = softwareSystem "TUI Adapter" "Bubbletea-based terminal UI for interactive parallel task visualization, phase management, and post-execution summaries" {

            # Parallel Console
            parallel_console = container "Parallel Console" "Rich multi-pane TUI for build/test output using bubbletea" "Go/Bubbletea" {
                console_model = component "Console Model" "Bubbletea Model with Init/Update/View cycle" "Go"
                message_pump = component "Message Pump" "Goroutine converting tea.Msg to Model updates" "Go"
                phase_tracker = component "Phase Tracker" "Tracks Init, Run, Summary phase transitions" "Go"
                uow_tracker = component "UoW Tracker" "Tracks unit-of-work start, running, complete states" "Go"
                tab_renderer = component "Tab Renderer" "Multi-tab pane rendering with scroll and selection" "Go"
                ring_buffer = component "Ring Buffer" "Fixed-size line buffer per output pane" "Go"
                exit_controller = component "Exit Controller" "Controlled exit timing with countdown" "Go"
            }

            # Console Render
            console_render = container "Console Render" "Rendering primitives for TUI widgets" "Go/Lipgloss" {
                icons = component "Icons" "Unicode and ASCII icon sets" "Go"
                lamps = component "Lamps" "Status indicator lamps (pass, fail, running)" "Go"
                styles = component "Styles" "Lipgloss style definitions for theming" "Go"
                text_utils = component "Text Utils" "Text truncation, wrapping, alignment" "Go"
            }

            # Observer
            observer = container "TUI Observer" "Translates ExecutionEvent to Console calls" "Go Package" {
                tui_observer = component "TUIObserver" "Maps OnUnitStart, OnUnitComplete to Console methods" "Go"
            }

            # Hooks
            hooks = container "TUI Hooks" "Bridges command framework lifecycle to TUI" "Go Package" {
                tui_hooks_impl = component "TUIHooksImpl" "OnPrepare, OnExecution, OnComplete lifecycle bridge" "Go"
            }

            # Registry
            tui_registry = container "TUI Registry" "Command-to-TUI binding registry" "Go Package" {
                binding_registry = component "Binding Registry" "Register, GetBinding, ResolveCommand with pattern matching" "Go"
                bootstrap = component "Bootstrap" "Factory registration for default TUI implementations" "Go"
            }

            # Selector
            selector = container "Selector Console" "Minimal TUI for interactive subcommand selection" "Go/Bubbletea" {
                selector_model = component "Selector Model" "Bubbletea model for list selection" "Go"
                option_types = component "Option Types" "CommandOption and selection result types" "Go"
            }

            # Stream
            stream = container "Stream" "Output stream utilities for TUI integration" "Go Package" {
                multi_writer = component "Multi Writer" "Attaches multiple output streams to Console" "Go"
                output_filter = component "Output Filter" "Filters and formats output for display" "Go"
            }

            # Environment
            tui_env = container "TUI Environment" "Environment detection for TUI capability" "Go Package" {
                tui_detector = component "TUI Detector" "IsInteractive(), ShouldUseTUI() environment probes" "Go"
            }
        }

        # Actor relationships
        developer -> tui_adapter "Views execution progress" "Terminal"

        # Consumer relationships
        eac_cli -> tui_adapter "Renders command output" "Go Import"
        clibase -> tui_adapter "Sends execution events" "Go Import"

        # Dependency relationships
        tui_adapter -> core "Uses logging and domain types" "Go Import"
        tui_adapter -> tui_contract "Implements Console interface" "Go Import"
        tui_adapter -> contracts_core "Uses ActionType and ExecutionObserver" "Go Import"

        # External relationships
        parallel_console -> terminal_sys "Renders to terminal via bubbletea" "stdout"
        selector -> terminal_sys "Renders selection list" "stdout"
        tui_env -> terminal_sys "Detects TTY and capabilities" "System calls"

        # Internal relationships
        tui_registry -> parallel_console "Creates ParallelConsole via factory"
        tui_registry -> selector "Creates SelectorConsole via factory"
        observer -> parallel_console "Translates events to Console calls"
        hooks -> parallel_console "Bridges lifecycle events"
        stream -> parallel_console "Attaches output writers"
        parallel_console -> console_render "Uses rendering primitives"

        # Component relationships - parallel console
        console_model -> message_pump "Receives messages from pump"
        console_model -> phase_tracker "Manages phase transitions"
        console_model -> uow_tracker "Tracks work unit states"
        console_model -> tab_renderer "Renders multi-tab output"
        tab_renderer -> ring_buffer "Reads buffered output lines"
        console_model -> exit_controller "Controls exit timing"

        # Component relationships - registry
        binding_registry -> bootstrap "Loads default TUI factories"
    }

    views {
        systemContext tui_adapter "SystemContext" {
            include *
            autoLayout lr
            title "TUI Adapter - System Context"
            description "Shows TUI adapter with consumers and external systems"
        }

        container tui_adapter "Containers" {
            include *
            autoLayout tb
            title "TUI Adapter - Package Architecture"
            description "Shows TUI adapter internal structure"
        }

        component parallel_console "ParallelConsoleComponents" {
            include *
            autoLayout tb
            title "Parallel Console - Components"
            description "Bubbletea-based multi-pane TUI"
        }

        theme default
    }

    configuration {
        scope softwaresystem
    }
}
