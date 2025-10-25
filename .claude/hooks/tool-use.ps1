# .claude/hooks/tool-use.ps1
# Log all tool invocations

param($tool, $parameters, $result)

$logDir = ".claude\logs"
$logFile = "$logDir\session-$(Get-Date -Format 'yyyy-MM-dd').jsonl"

# Ensure log directory exists
New-Item -ItemType Directory -Force -Path $logDir | Out-Null

# Create tool use entry
$logEntry = @{
    timestamp = (Get-Date -Format "o")
    event = "tool_use"
    tool = $tool
    parameters = $parameters
    result = if ($result.Length -gt 1000) { $result.Substring(0, 1000) + "..." } else { $result }
} | ConvertTo-Json -Compress -Depth 10

# Append to log file
Add-Content -Path $logFile -Value $logEntry -Encoding UTF8
