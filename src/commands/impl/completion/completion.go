// Command: completion
// Short: Generate shell completion scripts
// Long: The completion command generates shell completion scripts for various shells.
// Long: Supported shells: bash, zsh, fish, powershell
// Long: These scripts enable tab-completion for command names, flags, and flag values.
// Long: To install completions, source the generated script in your shell configuration file.
// Flag.shell: type=string, required=true, completion=bash,zsh,fish,powershell, usage=Shell type for which to generate completions (bash, zsh, fish, or powershell)
// HasSideEffects: false
package completion

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Completion)
}

// Completion generates shell completion scripts
func Completion() int {
	args := os.Args[2:] // Skip program name and "completion"

	var shell string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--shell" {
			if i+1 < len(args) {
				shell = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "Error: --shell requires a value\n")
				return 1
			}
		} else if !strings.HasPrefix(arg, "-") {
			// Positional argument (shell name)
			if shell == "" {
				shell = arg
			}
		}
	}

	// Validate shell argument
	if shell == "" {
		fmt.Fprintf(os.Stderr, "Error: Shell type is required\n")
		fmt.Fprintf(os.Stderr, "\nUsage: completion <shell> or completion --shell <shell>\n")
		fmt.Fprintf(os.Stderr, "Supported shells: bash, zsh, fish, powershell\n")
		return 1
	}

	shell = strings.ToLower(shell)

	// Generate completion script based on shell type
	switch shell {
	case "bash":
		return generateBashCompletion()
	case "zsh":
		return generateZshCompletion()
	case "fish":
		return generateFishCompletion()
	case "powershell", "pwsh":
		return generatePowershellCompletion()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unsupported shell: %s\n", shell)
		fmt.Fprintf(os.Stderr, "Supported shells: bash, zsh, fish, powershell\n")
		return 1
	}
}

// generateBashCompletion generates Bash completion script
func generateBashCompletion() int {
	commands := registry.GetCommands()
	commandRegistry := registry.GetCommandRegistry()

	fmt.Println("# Bash completion for custom commands")
	fmt.Println("# Source this file in your ~/.bashrc or ~/.bash_profile")
	fmt.Println()
	fmt.Println("_custom_commands_completion() {")
	fmt.Println("    local cur prev words cword")
	fmt.Println("    _init_completion || return")
	fmt.Println()
	fmt.Println("    # Get all command names")
	fmt.Print("    local commands=\"")

	// Get sorted command names
	var cmdNames []string
	for cmdName := range commands {
		cmdNames = append(cmdNames, cmdName)
	}
	sort.Strings(cmdNames)
	fmt.Print(strings.Join(cmdNames, " "))
	fmt.Println("\"")
	fmt.Println()

	fmt.Println("    # If we're completing the first argument, suggest commands")
	fmt.Println("    if [ $cword -eq 1 ]; then")
	fmt.Println("        COMPREPLY=( $(compgen -W \"${commands}\" -- \"${cur}\") )")
	fmt.Println("        return")
	fmt.Println("    fi")
	fmt.Println()

	fmt.Println("    # Get the command being executed")
	fmt.Println("    local cmd=\"${words[1]}\"")
	fmt.Println()

	// Generate flag completions for each command
	fmt.Println("    # Command-specific flag completions")
	fmt.Println("    case \"$cmd\" in")

	for _, cmdName := range cmdNames {
		canonicalName := registry.GetCanonicalName(cmdName)
		reg := commandRegistry[canonicalName]
		if reg == nil || len(reg.Flags) == 0 {
			continue
		}

		fmt.Printf("        \"%s\")\n", cmdName)

		// Collect all flags
		var flags []string
		for _, flag := range reg.Flags {
			flags = append(flags, "--"+flag.Name)
			if flag.Shorthand != "" {
				flags = append(flags, "-"+flag.Shorthand)
			}
		}

		fmt.Printf("            local flags=\"%s\"\n", strings.Join(flags, " "))
		fmt.Println("            COMPREPLY=( $(compgen -W \"${flags}\" -- \"${cur}\") )")
		fmt.Println("            return")
		fmt.Println("            ;;")
	}

	fmt.Println("    esac")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("# Register completion function")
	fmt.Println("complete -F _custom_commands_completion go")
	fmt.Println()
	fmt.Println("# Installation instructions:")
	fmt.Println("# 1. Save this script to a file (e.g., ~/.bash_completion.d/commands)")
	fmt.Println("# 2. Source it in your ~/.bashrc: source ~/.bash_completion.d/commands")
	fmt.Println("# 3. Reload your shell: source ~/.bashrc")

	return 0
}

// generateZshCompletion generates Zsh completion script
func generateZshCompletion() int {
	commands := registry.GetCommands()
	commandRegistry := registry.GetCommandRegistry()

	fmt.Println("#compdef _custom_commands go")
	fmt.Println("# Zsh completion for custom commands")
	fmt.Println("# Place this file in your $fpath (e.g., ~/.zsh/completions/_commands)")
	fmt.Println()
	fmt.Println("function _custom_commands {")
	fmt.Println("    local line state")
	fmt.Println()

	// Get sorted command names
	var cmdNames []string
	for cmdName := range commands {
		cmdNames = append(cmdNames, cmdName)
	}
	sort.Strings(cmdNames)

	fmt.Println("    _arguments -C \\")
	fmt.Println("        '1: :->commands' \\")
	fmt.Println("        '*::arg:->args'")
	fmt.Println()
	fmt.Println("    case $state in")
	fmt.Println("        commands)")
	fmt.Println("            local -a commands")
	fmt.Println("            commands=(")

	for _, cmdName := range cmdNames {
		canonicalName := registry.GetCanonicalName(cmdName)
		reg := commandRegistry[canonicalName]
		desc := ""
		if reg != nil {
			desc = reg.Short
		}
		fmt.Printf("                '%s:%s'\n", cmdName, desc)
	}

	fmt.Println("            )")
	fmt.Println("            _describe 'command' commands")
	fmt.Println("            ;;")
	fmt.Println("        args)")
	fmt.Println("            case $line[1] in")

	// Generate flag completions for each command
	for _, cmdName := range cmdNames {
		canonicalName := registry.GetCanonicalName(cmdName)
		reg := commandRegistry[canonicalName]
		if reg == nil || len(reg.Flags) == 0 {
			continue
		}

		fmt.Printf("                %s)\n", cmdName)
		fmt.Println("                    _arguments \\")

		for _, flag := range reg.Flags {
			flagSpec := fmt.Sprintf("                        '")
			if flag.Shorthand != "" {
				flagSpec += fmt.Sprintf("(-%s --%s)", flag.Shorthand, flag.Name)
			}
			flagSpec += fmt.Sprintf("{-%s,--%s}", flag.Shorthand, flag.Name)
			flagSpec += fmt.Sprintf("[%s]'", flag.Usage)
			fmt.Println(flagSpec + " \\")
		}

		fmt.Println("                    ;;")
	}

	fmt.Println("            esac")
	fmt.Println("            ;;")
	fmt.Println("    esac")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("# Installation instructions:")
	fmt.Println("# 1. Save this file to ~/.zsh/completions/_commands (create directory if needed)")
	fmt.Println("# 2. Add to ~/.zshrc: fpath=(~/.zsh/completions $fpath)")
	fmt.Println("# 3. Add to ~/.zshrc: autoload -Uz compinit && compinit")
	fmt.Println("# 4. Reload your shell: source ~/.zshrc")

	return 0
}

// generateFishCompletion generates Fish completion script
func generateFishCompletion() int {
	commands := registry.GetCommands()
	commandRegistry := registry.GetCommandRegistry()

	fmt.Println("# Fish completion for custom commands")
	fmt.Println("# Place this file in ~/.config/fish/completions/commands.fish")
	fmt.Println()

	// Get sorted command names
	var cmdNames []string
	for cmdName := range commands {
		cmdNames = append(cmdNames, cmdName)
	}
	sort.Strings(cmdNames)

	// Generate completions for each command
	for _, cmdName := range cmdNames {
		canonicalName := registry.GetCanonicalName(cmdName)
		reg := commandRegistry[canonicalName]

		// Command completion
		desc := ""
		if reg != nil {
			desc = reg.Short
		}
		fmt.Printf("complete -c go -n '__fish_use_subcommand' -a '%s' -d '%s'\n", cmdName, desc)

		// Flag completions
		if reg != nil && len(reg.Flags) > 0 {
			for _, flag := range reg.Flags {
				shortFlag := ""
				if flag.Shorthand != "" {
					shortFlag = fmt.Sprintf(" -s %s", flag.Shorthand)
				}
				fmt.Printf("complete -c go -n '__fish_seen_subcommand_from %s'%s -l %s -d '%s'\n",
					cmdName, shortFlag, flag.Name, flag.Usage)
			}
		}
	}

	fmt.Println()
	fmt.Println("# Installation instructions:")
	fmt.Println("# 1. Save this file to ~/.config/fish/completions/commands.fish")
	fmt.Println("# 2. Restart your Fish shell or run: source ~/.config/fish/completions/commands.fish")

	return 0
}

// generatePowershellCompletion generates PowerShell completion script
func generatePowershellCompletion() int {
	commands := registry.GetCommands()
	commandRegistry := registry.GetCommandRegistry()

	fmt.Println("# PowerShell completion for custom commands")
	fmt.Println("# Add this to your PowerShell profile ($PROFILE)")
	fmt.Println()

	fmt.Println("$commandsScriptBlock = {")
	fmt.Println("    param($commandName, $parameterName, $wordToComplete, $commandAst, $fakeBoundParameters)")
	fmt.Println()

	// Get sorted command names
	var cmdNames []string
	for cmdName := range commands {
		cmdNames = append(cmdNames, cmdName)
	}
	sort.Strings(cmdNames)

	fmt.Println("    $commands = @(")
	for _, cmdName := range cmdNames {
		canonicalName := registry.GetCanonicalName(cmdName)
		reg := commandRegistry[canonicalName]
		desc := ""
		if reg != nil {
			desc = reg.Short
		}
		fmt.Printf("        @{ Name = '%s'; Description = '%s' }\n", cmdName, desc)
	}
	fmt.Println("    )")
	fmt.Println()

	fmt.Println("    $commands | Where-Object { $_.Name -like \"$wordToComplete*\" } | ForEach-Object {")
	fmt.Println("        [System.Management.Automation.CompletionResult]::new($_.Name, $_.Name, 'ParameterValue', $_.Description)")
	fmt.Println("    }")
	fmt.Println("}")
	fmt.Println()

	fmt.Println("Register-ArgumentCompleter -Native -CommandName go -ScriptBlock $commandsScriptBlock")
	fmt.Println()
	fmt.Println("# Installation instructions:")
	fmt.Println("# 1. Open your PowerShell profile: notepad $PROFILE")
	fmt.Println("# 2. Paste this script into your profile")
	fmt.Println("# 3. Restart PowerShell or run: . $PROFILE")

	return 0
}

