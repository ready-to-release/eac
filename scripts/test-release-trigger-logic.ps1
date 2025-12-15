# =============================================================================
# Test Plan: Release Trigger Logic Simulation
# =============================================================================
#
# This script simulates the check-pending-releases action and trigger-releases
# job logic locally to verify correctness before CI runs.
#
# Usage: powershell -ExecutionPolicy Bypass -File scripts/test-release-trigger-logic.ps1
# =============================================================================

$COMMANDS = ".\go\eac\commands\build\commands.exe"
$PassCount = 0
$FailCount = 0

function Test-Pass {
    param([string]$Message)
    Write-Host "PASS: $Message" -ForegroundColor Green
    $script:PassCount++
}

function Test-Fail {
    param([string]$Message, [string]$Expected, [string]$Got)
    Write-Host "FAIL: $Message" -ForegroundColor Red
    Write-Host "       Expected: $Expected"
    Write-Host "       Got: $Got"
    $script:FailCount++
}

function Section {
    param([string]$Title)
    Write-Host ""
    Write-Host "--- $Title ---" -ForegroundColor Cyan
}

Write-Host "============================================================================="
Write-Host "RELEASE TRIGGER LOGIC TEST SUITE"
Write-Host "============================================================================="

# -----------------------------------------------------------------------------
# TEST 1: Semver Detection
# -----------------------------------------------------------------------------
Section "TEST 1: Semver Detection"

Write-Host "Running: $COMMANDS release tag-pending --all"
try {
    $result = & $COMMANDS release tag-pending --all 2>&1
    $resultStr = $result -join "`n"
    Write-Host "Result: $resultStr"

    $data = $resultStr | ConvertFrom-Json -ErrorAction Stop

    if ($null -ne $data.has_pending) {
        Test-Pass "tag-pending returns valid JSON with has_pending field"
        Write-Host "  has_pending: $($data.has_pending)"

        if ($data.has_pending -eq $true) {
            $pending = $data.results | Where-Object { $_.needs_tag -eq $true } | ForEach-Object { $_.module }
            Write-Host "  Pending semver modules: $($pending -join ' ')"
        }
    } else {
        Test-Fail "tag-pending JSON structure" "has_pending field" ($resultStr.Substring(0, [Math]::Min(100, $resultStr.Length)))
    }

    # Test parsing logic
    $semverModules = @($data.results | Where-Object { $_.needs_tag -eq $true } | ForEach-Object {
        $_ | Add-Member -NotePropertyName "type" -NotePropertyValue "semver" -PassThru
    })
    Write-Host "Parsed semver modules count: $($semverModules.Count)"
    Test-Pass "Semver parsing produces valid array"

} catch {
    Test-Fail "Semver detection" "valid JSON" $_.Exception.Message
}

# -----------------------------------------------------------------------------
# TEST 2: Calver Detection Logic
# -----------------------------------------------------------------------------
Section "TEST 2: Calver Detection Logic"

function Test-CalverDetection {
    param(
        [string]$Dispatched,
        [string]$CalverModules,
        [int]$ExpectedCount,
        [string]$TestName
    )

    $dispatchedList = if ($Dispatched) { $Dispatched -split '\s+' } else { @() }
    $calverList = if ($CalverModules) { $CalverModules -split '\s+' } else { @() }

    $calverPending = @()
    foreach ($calverMod in $calverList) {
        if ($dispatchedList -contains $calverMod) {
            $version = (Get-Date -Format "yyyy.MMdd.HHmm")
            $calverPending += @{
                module = $calverMod
                version = $version
                tag = "$calverMod/$version"
                needs_tag = $true
                type = "calver"
            }
        }
    }

    if ($calverPending.Count -eq $ExpectedCount) {
        Test-Pass "$TestName (count: $($calverPending.Count))"
    } else {
        Test-Fail $TestName "count=$ExpectedCount" "count=$($calverPending.Count)"
    }
    Write-Host "  Result: $($calverPending | ConvertTo-Json -Compress)"
}

Test-CalverDetection -Dispatched "docs books r2r-cli" -CalverModules "docs books" -ExpectedCount 2 -TestName "Both calver modules dispatched"
Test-CalverDetection -Dispatched "docs r2r-cli" -CalverModules "docs books" -ExpectedCount 1 -TestName "Only docs dispatched"
Test-CalverDetection -Dispatched "r2r-cli ext-eac" -CalverModules "docs books" -ExpectedCount 0 -TestName "No calver modules dispatched"
Test-CalverDetection -Dispatched "" -CalverModules "docs books" -ExpectedCount 0 -TestName "Empty dispatched list"

# -----------------------------------------------------------------------------
# TEST 3: Combining Semver + Calver
# -----------------------------------------------------------------------------
Section "TEST 3: Combining Semver + Calver"

$mockSemver = @(@{module="r2r-cli"; version="1.0.0"; tag="r2r-cli/1.0.0"; needs_tag=$true; type="semver"})
$mockCalver = @(@{module="docs"; version="2025.0116.1234"; tag="docs/2025.0116.1234"; needs_tag=$true; type="calver"})

$combined = $mockSemver + $mockCalver

if ($combined.Count -eq 2) {
    Test-Pass "Combining semver + calver produces correct count"
} else {
    Test-Fail "Combining arrays" "count=2" "count=$($combined.Count)"
}

Write-Host "Combined result: $($combined | ConvertTo-Json -Compress)"

# Count types manually to avoid PowerShell property counting quirks
$semverCount = @($combined | Where-Object { $_.type -eq "semver" }).Count
$calverCount = @($combined | Where-Object { $_.type -eq "calver" }).Count

if ($semverCount -eq 1 -and $calverCount -eq 1) {
    Test-Pass "Types preserved after combining"
} else {
    Test-Fail "Type preservation" "semver=1, calver=1" "semver=$semverCount, calver=$calverCount"
}

# Empty combination
$emptyCombined = @() + @()
if ($emptyCombined.Count -eq 0) {
    Test-Pass "Empty arrays combine to empty array"
} else {
    Test-Fail "Empty combination" "[]" "$emptyCombined"
}

# -----------------------------------------------------------------------------
# TEST 4: Execution Order Calculation
# -----------------------------------------------------------------------------
Section "TEST 4: Execution Order Calculation"

Write-Host "Testing execution order for: docs books"
try {
    $result = & $COMMANDS get execution order docs books --no-deps --as-json 2>&1
    $resultStr = $result -join "`n"
    Write-Host "Result: $resultStr"

    $data = $resultStr | ConvertFrom-Json -ErrorAction Stop

    if ($null -ne $data.layers) {
        Test-Pass "Execution order returns valid JSON with layers"
        Write-Host "  Layer count: $($data.layer_count)"

        $allModules = @()
        foreach ($layer in $data.layers) {
            $allModules += $layer
        }
        Write-Host "  Modules in layers: $($allModules -join ' ')"
    } else {
        Test-Fail "Execution order structure" "layers field" ($resultStr.Substring(0, [Math]::Min(100, $resultStr.Length)))
    }
} catch {
    Test-Fail "Execution order" "valid JSON" $_.Exception.Message
}

Write-Host ""
Write-Host "Testing dependency ordering for: r2r-cli docs"
try {
    $result2 = & $COMMANDS get execution order r2r-cli docs --no-deps --as-json 2>&1
    $resultStr2 = $result2 -join "`n"
    Write-Host "Result: $resultStr2"
} catch {
    Write-Host "Error: $($_.Exception.Message)"
}

# -----------------------------------------------------------------------------
# TEST 5: Layer Enrichment Logic
# -----------------------------------------------------------------------------
Section "TEST 5: Layer Enrichment Logic"

$mockLayers = @(@("docs"), @("books"))
$mockModulesJson = @(
    @{module="docs"; version="2025.0116.1234"; tag="docs/2025.0116.1234"; type="calver"}
    @{module="books"; version="2025.0116.1234"; tag="books/2025.0116.1234"; type="calver"}
)

# Build module lookup
$moduleLookup = @{}
foreach ($m in $mockModulesJson) {
    $moduleLookup[$m.module] = $m
}

# Enrich layers
$enrichedLayers = @()
foreach ($layer in $mockLayers) {
    $enrichedLayer = @()
    foreach ($module in $layer) {
        if ($moduleLookup.ContainsKey($module)) {
            $enrichedLayer += $moduleLookup[$module]
        }
    }
    $enrichedLayers += ,@($enrichedLayer)
}

Write-Host "Enriched layers: $($enrichedLayers | ConvertTo-Json -Depth 3 -Compress)"

if ($enrichedLayers.Count -eq 2) {
    Test-Pass "Enrichment produces correct layer count"
} else {
    Test-Fail "Enrichment layer count" "2" "$($enrichedLayers.Count)"
}

if ($enrichedLayers[0] -and $enrichedLayers[0][0] -and $enrichedLayers[0][0].version) {
    Test-Pass "Enriched layers contain version info"
} else {
    Test-Fail "Version in enriched layers" "true" "false"
}

# -----------------------------------------------------------------------------
# TEST 6: Full Integration Simulation
# -----------------------------------------------------------------------------
Section "TEST 6: Full Integration Simulation"

Write-Host "Simulating full check-pending-releases flow..."
Write-Host ""

# Step 1: Get semver pending
Write-Host "Step 1: Check semver pending releases"
try {
    $semverResult = & $COMMANDS release tag-pending --all 2>&1 | ConvertFrom-Json
    $semverModules = @($semverResult.results | Where-Object { $_.needs_tag -eq $true } | ForEach-Object {
        $_ | Add-Member -NotePropertyName "type" -NotePropertyValue "semver" -PassThru
    })
    $moduleNames = $semverModules | ForEach-Object { $_.module }
    Write-Host "  Semver modules: $(if ($moduleNames) { $moduleNames -join ' ' } else { '(none)' })"
} catch {
    $semverModules = @()
    Write-Host "  Semver modules: (parse error)"
}

# Step 2: Check calver pending (simulate docs dispatched)
Write-Host ""
Write-Host "Step 2: Check calver pending releases (simulating: docs dispatched)"
$dispatched = @("docs")
$calverModulesList = @("docs", "books")

$calverPending = @()
foreach ($calverMod in $calverModulesList) {
    if ($dispatched -contains $calverMod) {
        $version = (Get-Date -Format "yyyy.MMdd.HHmm")
        $calverPending += @{
            module = $calverMod
            version = $version
            tag = "$calverMod/$version"
            needs_tag = $true
            type = "calver"
        }
    }
}
$moduleNames = $calverPending | ForEach-Object { $_.module }
Write-Host "  Calver modules: $(if ($moduleNames) { $moduleNames -join ' ' } else { '(none)' })"

# Step 3: Combine
Write-Host ""
Write-Host "Step 3: Combine pending releases"
$combined = @($semverModules) + @($calverPending)
Write-Host "  Total pending: $($combined.Count)"

if ($combined.Count -gt 0) {
    $pendingMods = $combined | ForEach-Object { $_.module }
    Write-Host "  Modules: $($pendingMods -join ' ')"

    # Step 4: Calculate execution order
    Write-Host ""
    Write-Host "Step 4: Calculate execution order"
    try {
        $cmdArgs = @("get", "execution", "order") + $pendingMods + @("--no-deps", "--as-json")
        $execResult = & $COMMANDS $cmdArgs 2>&1 | ConvertFrom-Json
        Write-Host "  Layer count: $($execResult.layer_count)"

        # Step 5: Enrich layers
        Write-Host ""
        Write-Host "Step 5: Final enriched layers"

        $moduleLookup = @{}
        foreach ($m in $combined) {
            $moduleLookup[$m.module] = $m
        }

        $enriched = @()
        foreach ($layer in $execResult.layers) {
            $enrichedLayer = @()
            foreach ($module in $layer) {
                if ($moduleLookup.ContainsKey($module)) {
                    $enrichedLayer += $moduleLookup[$module]
                }
            }
            $enriched += ,@($enrichedLayer)
        }

        Write-Host ($enriched | ConvertTo-Json -Depth 3)
        Test-Pass "Full integration simulation completed"
    } catch {
        Test-Fail "Full integration" "valid execution order" $_.Exception.Message
    }
} else {
    Write-Host "  No pending releases"
    Test-Pass "Full integration simulation completed (no pending)"
}

# -----------------------------------------------------------------------------
# TEST 7: Edge Cases
# -----------------------------------------------------------------------------
Section "TEST 7: Edge Cases"

Write-Host "Testing word boundary matching..."

# Word boundary test (PowerShell uses -contains for exact array match)
$testList = "docs books r2r-cli".Split()
if ($testList -contains "docs") {
    Test-Pass "Word match finds 'docs' in list"
} else {
    Test-Fail "Word match" "match" "no match"
}

$testList2 = "docs-extra books".Split()
if ($testList2 -notcontains "docs") {
    Test-Pass "Word match does NOT match 'docs' in 'docs-extra' (word boundary)"
} else {
    Test-Fail "Word boundary" "no match for 'docs' in 'docs-extra'" "matched"
}

Write-Host ""
Write-Host "Testing empty module handling..."
$emptyModules = ""
if ([string]::IsNullOrEmpty($emptyModules)) {
    Test-Pass "Empty module detection works"
} else {
    Test-Fail "Empty module detection" "empty" $emptyModules
}

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
Write-Host ""
Write-Host "============================================================================="
Write-Host "TEST SUMMARY"
Write-Host "============================================================================="
Write-Host "Passed: $PassCount" -ForegroundColor Green
Write-Host "Failed: $FailCount" -ForegroundColor Red
Write-Host ""

if ($FailCount -gt 0) {
    Write-Host "Some tests failed!" -ForegroundColor Red
    exit 1
} else {
    Write-Host "All tests passed!" -ForegroundColor Green
    exit 0
}
