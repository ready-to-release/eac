# Quick Start: Running Commands

This repository provides convenient wrapper scripts to run commands from `src/commands/`.

## One-Time Setup (PowerShell Only)

If you use PowerShell, run the setup script once to enable the `run` alias:

```powershell
.\setup-alias.ps1
```

Then reload your profile:
```powershell
. $PROFILE
```

After this, you can use the short `run` command instead of `.\run.ps1`.

## Usage

### PowerShell (Windows)

**After setup**:
```powershell
run example-show-modules
run example-show-files
```

**Without setup**:
```powershell
.\run.ps1 example-show-modules
.\run.ps1 example-show-files
```

### Bash (Unix/Linux/macOS)

```bash
./run.sh example-show-modules
./run.sh example-show-files
```

### List Available Commands

```bash
# PowerShell
run              # After setup
.\run.ps1        # Without setup

# Bash
./run.sh
```

## Available Commands

- `example-show-modules` - Display all module contracts
- `example-show-files` - Show file ownership and coverage

## How It Works

### PowerShell Setup

The `setup-alias.ps1` script adds this function to your PowerShell profile:

```powershell
function run {
    $runScript = Join-Path "C:\path\to\repo" "run.ps1"
    if (Test-Path $runScript) {
        & $runScript @args
    } else {
        Write-Error "run.ps1 not found"
    }
}
```

This allows you to use `run <command>` from anywhere in your terminal.

### Wrapper Scripts

Both `run.sh` and `run.ps1`:
1. Validate the command exists
2. Change to `src/commands/` directory
3. Execute `go run <command>.go` with arguments
4. Handle errors and show helpful messages

## Examples

### Show All Modules

```powershell
# PowerShell (after setup)
run example-show-modules

# Or
.\run.ps1 example-show-modules

# Bash
./run.sh example-show-modules
```

### Show Files with Coverage

```powershell
# PowerShell (after setup)
run example-show-files

# Or
.\run.ps1 example-show-files

# Bash
./run.sh example-show-files
```

### Pass Arguments

```powershell
# PowerShell (after setup)
run example-show-files --verbose

# Bash
./run.sh example-show-files --verbose
```

## Troubleshooting

### PowerShell: "run is not recognized"

Make sure you've:
1. Run `.\setup-alias.ps1`
2. Reloaded your profile: `. $PROFILE`
3. Or restarted PowerShell

### Bash: "Permission denied"

Make sure run.sh is executable:
```bash
chmod +x run.sh
./run.sh example-show-modules
```

### Command not found

Make sure the command exists in `src/commands/`:
```bash
ls src/commands/*.go
```

## Next Steps

- See `src/commands/README.md` for details on creating new commands
- See `out/IMPLEMENT-RUN-WRAPPERS.md` for implementation details
- Add new `.go` files to `src/commands/` and they're automatically available!
