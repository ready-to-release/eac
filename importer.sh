#!/bin/bash
#
# importer.sh - Import the go-invoker module and optionally create 'run' alias
#
# Usage:
#   source ./importer.sh           # Import with alias
#   source ./importer.sh --no-alias # Import without alias
#
# Description:
#   This script sources the go.sh module and creates a session-wide 'run' alias.
#   The alias is temporary and only lasts for the current shell session.
#
#   Welcome messages are shown only once per session (tracked via $RUN_WELCOME_EMITTED).
#
# Note:
#   Must be sourced (source ./importer.sh) for the alias to work in your session.

set -e

# Parse arguments
NO_ALIAS=0
if [ "$1" = "--no-alias" ]; then
    NO_ALIAS=1
fi

# Check if welcome already shown this session
SHOW_WELCOME=1
if [ -n "$RUN_WELCOME_EMITTED" ]; then
    SHOW_WELCOME=0
fi

# Get script directory and module path
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_PATH="$SCRIPT_DIR/scripts/sh/go-invoker/go.sh"

# Check if module exists
if [ ! -f "$MODULE_PATH" ]; then
    echo "Error: Module not found at: $MODULE_PATH" >&2
    return 1 2>/dev/null || exit 1
fi

# Source the module
source "$MODULE_PATH"

# Show welcome only once per session
if [ "$SHOW_WELCOME" -eq 1 ]; then
    echo "✅ go-invoker module imported successfully!"
fi

# Create alias unless --no-alias specified
if [ "$NO_ALIAS" -eq 0 ]; then
    new_run_alias
else
    if [ "$SHOW_WELCOME" -eq 1 ]; then
        echo ""
        echo "You can now use:"
        echo "  invoke_go_src_command <command-name> [args...]"
        echo ""
        echo "To create the 'run' alias, call:"
        echo "  new_run_alias"
        echo ""
    fi
fi

# Mark welcome as shown for this session
export RUN_WELCOME_EMITTED=1
