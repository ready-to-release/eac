# UserPromptSubmit hook - Windows-compatible alternative to SessionStart
# This hook runs on the FIRST user prompt submission in a new session

# Check if we've already run the boot sequence in this session
$markerFile = ".claude/logs/session-boot-complete.log"

if (-not (Test-Path $markerFile)) {
    # Create marker to prevent re-running on every prompt
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    "Boot hook executed at: $timestamp" | Out-File -FilePath $markerFile -Encoding utf8

    # Output the hook context for Claude to receive
    $output = @{
        hookSpecificOutput = @{
            hookEventName = "UserPromptSubmit"
            additionalContext = "FIRST PROMPT DETECTED - Execute boot sequence: Read .claude/agents/boot.md and follow its instructions to load project context from CLAUDE.md."
        }
    }

    Write-Output ($output | ConvertTo-Json -Compress)
} else {
    # Already booted - output nothing (silent)
    Write-Output '{"hookSpecificOutput":{}}'
}
