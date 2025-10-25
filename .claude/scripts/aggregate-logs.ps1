# .claude/scripts/aggregate-logs.ps1
# Aggregate and analyze logs from .claude/logs

param(
    [string]$Date = (Get-Date -Format "yyyy-MM-dd"),
    [switch]$Summary
)

$logFiles = Get-ChildItem ".claude\logs\*-$Date.jsonl" -ErrorAction SilentlyContinue

if (-not $logFiles) {
    Write-Host "No logs found for date: $Date"
    exit
}

$allEntries = @()
foreach ($file in $logFiles) {
    $entries = Get-Content $file | ForEach-Object { $_ | ConvertFrom-Json }
    $allEntries += $entries
}

if ($Summary) {
    # Display summary
    $totalEvents = $allEntries.Count
    $eventTypes = $allEntries | Group-Object -Property event | Select-Object Name, Count

    Write-Host "`n=== Log Summary for $Date ==="
    Write-Host "Total Events: $totalEvents"
    Write-Host "`nEvent Breakdown:"
    $eventTypes | Format-Table -AutoSize
} else {
    # Display all entries
    $allEntries | Sort-Object timestamp | ConvertTo-Json -Depth 10
}
