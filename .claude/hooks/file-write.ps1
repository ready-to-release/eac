# .claude/hooks/file-write.ps1
# Log file write operations

# Read JSON input from stdin
$input = [Console]::In.ReadToEnd()
$hookData = $input | ConvertFrom-Json -ErrorAction SilentlyContinue

$logDir = ".claude\logs"
$logFile = "$logDir\file-operations-$(Get-Date -Format 'yyyy-MM-dd').jsonl"

# Ensure log directory exists
New-Item -ItemType Directory -Force -Path $logDir | Out-Null

# Extract file path and tool name
$filePath = if ($hookData.tool_input.file_path) { $hookData.tool_input.file_path } else { "unknown" }
$toolName = if ($hookData.tool_name) { $hookData.tool_name } else { "unknown" }
$success = if ($hookData.tool_response.success -ne $null) { $hookData.tool_response.success } else { $true }

# Create file operation entry
$logEntry = @{
    timestamp = (Get-Date -Format "o")
    event = "file_operation"
    tool = $toolName
    filePath = $filePath
    success = $success
    sessionId = $hookData.session_id
} | ConvertTo-Json -Compress

# Append to log file
Add-Content -Path $logFile -Value $logEntry -Encoding UTF8
