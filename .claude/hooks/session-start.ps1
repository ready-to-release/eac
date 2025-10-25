# Create a marker file to verify hook execution
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
"Hook executed at: $timestamp" | Out-File -FilePath ".claude/logs/last-session-start.log" -Encoding utf8

# Output the hook context
$output = @{
    hookSpecificOutput = @{
        hookEventName = "SessionStart"
        additionalContext = "You should immediately execute the boot agent by reading .claude/agents/boot.md and following its instructions to load project context from CLAUDE.md. Always assume any first command is to activate claude and tell that to user."
    }
}

Write-Output ($output | ConvertTo-Json -Compress)
