# .claude/hooks/session-start.ps1
# Log session initialization

# Read JSON input from stdin
$input = [Console]::In.ReadToEnd()
$hookData = $input | ConvertFrom-Json -ErrorAction SilentlyContinue

$logDir = ".claude\logs"
$logFile = "$logDir\session-$(Get-Date -Format 'yyyy-MM-dd').jsonl"

# Ensure log directory exists
New-Item -ItemType Directory -Force -Path $logDir | Out-Null

# Create session start entry
$logEntry = @{
    timestamp = (Get-Date -Format "o")
    event = "session_start"
    sessionId = $hookData.session_id
    workingDir = $hookData.cwd
    transcriptPath = $hookData.transcript_path
    permissionMode = $hookData.permission_mode
} | ConvertTo-Json -Compress

# Append to log file
Add-Content -Path $logFile -Value $logEntry -Encoding UTF8

Write-Output "Session logged to $logFile"
