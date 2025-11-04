#!/bin/bash
#
# go.sh - Bash module for running Go commands from src/commands/
#
# Functions:
#   invoke_go_src_command <command-name> [args...]  - Run a Go command
#   get_go_src_commands                              - List available commands
#   new_run_alias                                    - Create 'run' alias

# Get repository root (three levels up from this script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

#
# invoke_go_src_command - Run a command from src/commands/
#
# Usage: invoke_go_src_command <command> [subcommand] [args...]
# Examples:
#   invoke_go_src_command show modules
#   invoke_go_src_command show files
#   invoke_go_src_command list-commands
#
invoke_go_src_command() {
    # Check if any args provided
    if [ $# -eq 0 ]; then
        echo "Usage: run <command> [subcommand] [args...]"
        echo ""
        echo "Available commands:"

        local commands=$(get_go_src_commands)
        if [ -n "$commands" ]; then
            echo "$commands" | sed 's|^|  |'
        else
            echo "  (no commands found)"
        fi
        return 1
    fi

    local commands_path="$REPO_ROOT/src/commands"

    # Run the command via dispatcher (let Go handle command validation)
    # Pass all arguments - Go will figure out what's command vs args
    (
        cd "$commands_path" || exit 1
        go run . "$@"
    )
}

#
# get_go_src_commands - Get available Go commands from src/commands/
#
# Usage: get_go_src_commands
#
# Returns: List of command names (one per line)
#
get_go_src_commands() {
    local commands_path="$REPO_ROOT/src/commands"

    if [ ! -d "$commands_path" ]; then
        return 0
    fi

    # Call the Go dispatcher to get the list of commands (single source of truth)
    (
        cd "$commands_path" || return 1
        go run . list-commands 2>/dev/null
    )
}

#
# new_run_alias - Create a 'run' alias for the current shell session
#
# Usage: new_run_alias
#
new_run_alias() {
    # Check if alias already exists
    if alias run >/dev/null 2>&1; then
        if [ -z "$RUN_WELCOME_EMITTED" ]; then
            echo "Alias 'run' already exists"
        fi
        return 0
    fi

    # Create the alias
    alias run='invoke_go_src_command'

    # Show welcome only once per session
    if [ -z "$RUN_WELCOME_EMITTED" ]; then
        echo "✅ Created alias 'run' for this session"
        echo ""
        echo "You can now use:"
        echo "  run <command-name> [args...]"
        echo ""
        echo "To make this permanent, add this to your ~/.bashrc or ~/.zshrc:"
        echo "  source \"$REPO_ROOT/scripts/sh/go-invoker/go.sh\""
        echo "  new_run_alias"
        echo ""
    fi
}

# Export functions so they can be used after sourcing
export -f invoke_go_src_command
export -f get_go_src_commands
export -f new_run_alias
