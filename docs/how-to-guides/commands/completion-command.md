# Completion Command

**Problem**: Typing full command names and flags manually is slow and error-prone, especially with complex CLIs.

**Solution**: Use `completion` to generate shell-specific tab completion scripts that autocomplete commands, subcommands, and flags.

## Key Benefits

- Faster command entry with tab completion
- Discover available commands and flags interactively
- Reduce typos and syntax errors
- Support for bash, zsh, fish, and PowerShell

## Quick Reference

```bash
# Generate completion script for your shell
r2r eac completion bash    # Bash
r2r eac completion zsh     # Zsh
r2r eac completion fish    # Fish
r2r eac completion powershell  # PowerShell

# Quick install (Bash example)
r2r eac completion bash > ~/.bash_completion.d/r2r-eac
source ~/.bashrc
```

## Installation Instructions

### Bash

- **Option 1: User completion directory (Recommended)**

```bash
# Create completion directory if it doesn't exist
mkdir -p ~/.bash_completion.d

# Generate and save completion script
r2r eac completion bash > ~/.bash_completion.d/r2r-eac

# Add to ~/.bashrc (if not already present)
echo 'for f in ~/.bash_completion.d/*; do source "$f"; done' >> ~/.bashrc

# Reload shell configuration
source ~/.bashrc
```

- **Option 2: Direct sourcing**

```bash
# Generate completion script
r2r eac completion bash > ~/r2r-eac-completion.bash

# Add to ~/.bashrc
echo 'source ~/r2r-eac-completion.bash' >> ~/.bashrc

# Reload shell configuration
source ~/.bashrc
```

**System-wide installation (Linux):**

```bash
# Requires sudo
sudo r2r eac completion bash > /etc/bash_completion.d/r2r-eac
```

### Zsh

- **Option 1: User completion directory (Recommended)**

```bash
# Create completion directory if it doesn't exist
mkdir -p ~/.zsh/completion

# Generate and save completion script
r2r eac completion zsh > ~/.zsh/completion/_r2r-eac

# Add to ~/.zshrc (if not already present)
echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

# Reload shell configuration
source ~/.zshrc
```

- **Option 2: oh-my-zsh integration**

```bash
# Generate completion script
r2r eac completion zsh > ~/.oh-my-zsh/completions/_r2r-eac

# Reload shell configuration
source ~/.zshrc
```

**System-wide installation (macOS with Homebrew):**

```bash
# Find zsh site-functions directory
brew --prefix

# Generate completion (replace /usr/local with your brew prefix)
r2r eac completion zsh > /usr/local/share/zsh/site-functions/_r2r-eac

# Restart terminal
```

### Fish

- **User installation (Recommended)**

```bash
# Create completion directory if it doesn't exist
mkdir -p ~/.config/fish/completions

# Generate and save completion script
r2r eac completion fish > ~/.config/fish/completions/r2r-eac.fish

# Reload completions
fish_update_completions
```

**System-wide installation:**

```bash
# Requires sudo
sudo r2r eac completion fish > /usr/share/fish/vendor_completions.d/r2r-eac.fish
```

### PowerShell

- **Current user (Recommended)**

```powershell
# Create profile directory if it doesn't exist
New-Item -ItemType Directory -Force -Path (Split-Path $PROFILE)

# Generate completion script
r2r eac completion powershell | Out-File -Append $PROFILE -Encoding UTF8

# Reload profile
. $PROFILE
```

- **Alternative: Separate completion file**

```powershell
# Generate completion script
r2r eac completion powershell > "$HOME\Documents\PowerShell\r2r-eac-completion.ps1"

# Add to profile
Add-Content $PROFILE "`n. `"$HOME\Documents\PowerShell\r2r-eac-completion.ps1`""

# Reload profile
. $PROFILE
```

**All users (requires admin):**

```powershell
# Run PowerShell as Administrator
r2r eac completion powershell | Out-File -Append $PROFILE.AllUsersAllHosts -Encoding UTF8
```

## Examples

### Bash Usage

```bash
# After installation, tab completion works:
r2r eac <TAB>                  # Shows all commands
r2r eac build <TAB>            # Shows available modules
r2r eac test --<TAB>           # Shows available flags
r2r eac work c<TAB>            # Completes to "create"
r2r eac completion <TAB>       # Shows: bash fish powershell zsh
```

### Zsh Usage

```bash
# Tab completion with descriptions
r2r eac <TAB>
# build    -- Build one or more modules by moniker
# test     -- Test one or more modules by moniker
# validate -- Validate repository contracts and dependencies

# Flag completion
r2r eac test --<TAB>
# --as-cucumber -- Output in Cucumber JSON format
# --as-junit    -- Output in JUnit XML format
# --suite       -- Filter tests by suite
```

### Fish Usage

```bash
# Interactive completion with fuzzy matching
r2r eac <TAB>                  # Visual menu of commands
r2r eac specs c<TAB>           # Completes "create" with description
r2r eac test --s<TAB>          # Shows --suite flag
```

### PowerShell Usage

```powershell
# Tab completion cycles through options
r2r eac <TAB>                  # Cycles: build, test, validate...
r2r eac completion <TAB>       # Cycles: bash, fish, powershell, zsh
r2r eac test --<TAB>           # Cycles through flags
```

## Verification

Test that completion is working correctly:

```bash
# Type command and press TAB
r2r eac <TAB>

# Should show available commands like:
# build  completion  design  docs  get  help  init  pipeline  ...

# Type partial command and press TAB
r2r eac compl<TAB>

# Should complete to:
# r2r eac completion
```

If completion doesn't work, see Troubleshooting below.

## Typical Workflows

### First-time Setup

```bash
# 1. Identify your shell
echo $SHELL
# Output: /bin/bash (or /bin/zsh, /usr/bin/fish, etc.)

# 2. Generate and install completion
r2r eac completion bash > ~/.bash_completion.d/r2r-eac
echo 'for f in ~/.bash_completion.d/*; do source "$f"; done' >> ~/.bashrc

# 3. Reload configuration
source ~/.bashrc

# 4. Test completion
r2r eac <TAB>
```

### Updating Completion After CLI Changes

When the CLI adds new commands or flags:

```bash
# Bash
r2r eac completion bash > ~/.bash_completion.d/r2r-eac
source ~/.bashrc

# Zsh
r2r eac completion zsh > ~/.zsh/completion/_r2r-eac
rm -f ~/.zcompdump
source ~/.zshrc

# Fish
r2r eac completion fish > ~/.config/fish/completions/r2r-eac.fish
fish_update_completions

# PowerShell
r2r eac completion powershell | Out-File -Force "$HOME\Documents\PowerShell\r2r-eac-completion.ps1"
. $PROFILE
```

### Multi-shell Developers

If you use different shells on different systems:

```bash
# Generate all completion scripts to version control
mkdir -p completions
r2r eac completion bash > completions/r2r-eac.bash
r2r eac completion zsh > completions/_r2r-eac
r2r eac completion fish > completions/r2r-eac.fish
r2r eac completion powershell > completions/r2r-eac.ps1

# Install on each system as needed
# Bash: source completions/r2r-eac.bash
# Zsh: copy to fpath location
# Fish: copy to ~/.config/fish/completions/
# PowerShell: source in $PROFILE
```

## Troubleshooting

| Problem                              | Solution                                                                |
| ------------------------------------ | ----------------------------------------------------------------------- |
| Tab completion not working           | Verify shell with `echo $SHELL`, ensure you installed for correct shell |
| "command not found: compdef" (Zsh)   | Add `autoload -Uz compinit && compinit` to ~/.zshrc                     |
| Completion not loading after install | Restart terminal or run `source ~/.bashrc` (or ~/.zshrc)                |
| Old completions cached (Zsh)         | Delete cache: `rm -f ~/.zcompdump && source ~/.zshrc`                   |
| Permission denied                    | Use `sudo` for system-wide install or install to user directory         |
| PowerShell execution policy error    | Run `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`               |
| Completion shows wrong commands      | Regenerate: `r2r eac completion <shell> > <file>` and reload            |
| Fish completions not updating        | Run `fish_update_completions` or restart Fish                           |
| Bash completion directory missing    | Create it: `mkdir -p ~/.bash_completion.d`                              |

## Advanced Usage

### Programmatic Completion Generation

Generate completion for CI/CD or packaging:

```bash
# Generate all completions for distribution
for shell in bash zsh fish powershell; do
  r2r eac completion $shell > "dist/completions/r2r-eac.$shell"
done
```

### Custom Installation Locations

```bash
# Bash: Custom directory
r2r eac completion bash > /opt/completions/r2r-eac
echo 'source /opt/completions/r2r-eac' >> ~/.bashrc

# Zsh: Custom fpath
r2r eac completion zsh > /custom/path/_r2r-eac
echo 'fpath=(/custom/path $fpath)' >> ~/.zshrc
```

### Conditional Loading (Performance)

For large profiles, lazy-load completion:

```bash
# Add to ~/.bashrc (Bash)
if command -v r2r &> /dev/null; then
  source ~/.bash_completion.d/r2r-eac
fi

# Add to ~/.zshrc (Zsh)
if command -v r2r &> /dev/null; then
  fpath=(~/.zsh/completion $fpath)
  autoload -Uz compinit && compinit
fi
```

## Shell-Specific Features

### Bash

- Basic tab completion
- Command and flag completion
- No descriptions in completion menu

### Zsh

- Tab completion with descriptions
- Fuzzy matching support
- Completion menu with details
- Case-insensitive completion

### Fish

- Interactive visual menus
- Fuzzy search and filtering
- Rich descriptions for each option
- Automatic completion updates

### PowerShell

- Cycle-based completion (press TAB repeatedly)
- Parameter completion
- Supports both `-Flag` and `--flag` syntax
- Integration with PowerShell's completion system

## Summary

**Installation steps:**

1. Identify your shell: `echo $SHELL`
2. Generate completion: `r2r eac completion <shell>`
3. Install to appropriate location (see Installation Instructions)
4. Reload shell configuration
5. Test with `r2r eac <TAB>`

**Supported shells:**

- **Bash**: `~/.bash_completion.d/r2r-eac`
- **Zsh**: `~/.zsh/completion/_r2r-eac`
- **Fish**: `~/.config/fish/completions/r2r-eac.fish`
- **PowerShell**: Add to `$PROFILE`

Tab completion significantly improves CLI productivity by reducing typing, preventing errors, and making command discovery interactive.
