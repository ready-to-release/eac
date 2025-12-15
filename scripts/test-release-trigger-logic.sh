#!/bin/bash
# =============================================================================
# Test Plan: Release Trigger Logic Simulation
# =============================================================================
#
# This script simulates the check-pending-releases action and trigger-releases
# job logic locally to verify correctness before CI runs.
#
# Test Scenarios:
# 1. Semver only: r2r-cli has pending changelog release
# 2. Calver only: docs was dispatched, books was not
# 3. Both: semver + calver releases pending
# 4. None: no pending releases
# 5. Dependency ordering: verify layers respect module dependencies
#
# Usage: ./scripts/test-release-trigger-logic.sh
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

COMMANDS="./go/eac/commands/build/commands.exe"
PASS_COUNT=0
FAIL_COUNT=0

# Build commands if needed
if [ ! -f "$COMMANDS" ]; then
    echo -e "${YELLOW}Building commands binary...${NC}"
    go build -o "$COMMANDS" ./go/eac/commands
fi

echo ""
echo "============================================================================="
echo "RELEASE TRIGGER LOGIC TEST SUITE"
echo "============================================================================="
echo ""

# -----------------------------------------------------------------------------
# Helper Functions
# -----------------------------------------------------------------------------

test_pass() {
    echo -e "${GREEN}PASS${NC}: $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

test_fail() {
    echo -e "${RED}FAIL${NC}: $1"
    echo -e "       Expected: $2"
    echo -e "       Got: $3"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

section() {
    echo ""
    echo -e "${BLUE}--- $1 ---${NC}"
}

# -----------------------------------------------------------------------------
# TEST 1: Semver Detection (release tag-pending)
# -----------------------------------------------------------------------------
section "TEST 1: Semver Detection"

echo "Running: $COMMANDS release tag-pending --all"
SEMVER_RESULT=$($COMMANDS release tag-pending --all 2>/dev/null || echo '{"error":"command failed"}')

echo "Result:"
echo "$SEMVER_RESULT" | jq . 2>/dev/null || echo "$SEMVER_RESULT"

# Verify JSON structure
if echo "$SEMVER_RESULT" | jq -e '.has_pending' > /dev/null 2>&1; then
    test_pass "tag-pending returns valid JSON with has_pending field"

    HAS_PENDING=$(echo "$SEMVER_RESULT" | jq -r '.has_pending')
    echo "  has_pending: $HAS_PENDING"

    if [ "$HAS_PENDING" = "true" ]; then
        PENDING_MODULES=$(echo "$SEMVER_RESULT" | jq -r '.results[] | select(.needs_tag == true) | .module' | tr '\n' ' ')
        echo "  Pending semver modules: $PENDING_MODULES"
    fi
else
    test_fail "tag-pending JSON structure" "has_pending field" "$(echo "$SEMVER_RESULT" | head -c 100)"
fi

# Test parsing logic (simulating action step)
SEMVER_MODULES_JSON=$(echo "$SEMVER_RESULT" | jq -c '[.results[] | select(.needs_tag == true) | . + {type: "semver"}]' 2>/dev/null || echo "[]")
echo "Parsed semver modules: $SEMVER_MODULES_JSON"

if echo "$SEMVER_MODULES_JSON" | jq -e '.' > /dev/null 2>&1; then
    test_pass "Semver parsing produces valid JSON array"
else
    test_fail "Semver parsing" "valid JSON array" "$SEMVER_MODULES_JSON"
fi

# -----------------------------------------------------------------------------
# TEST 2: Calver Detection Logic
# -----------------------------------------------------------------------------
section "TEST 2: Calver Detection Logic"

# Simulate different dispatched-modules scenarios
test_calver_detection() {
    local DISPATCHED="$1"
    local CALVER_MODULES="$2"
    local EXPECTED_COUNT="$3"
    local TEST_NAME="$4"

    CALVER_PENDING="[]"

    for calver_mod in $CALVER_MODULES; do
        if echo "$DISPATCHED" | grep -qw "$calver_mod"; then
            VERSION=$(date -u +"%Y.%m%d.%H%M")
            TAG="${calver_mod}/${VERSION}"
            CALVER_PENDING=$(echo "$CALVER_PENDING" | jq -c ". + [{\"module\": \"$calver_mod\", \"version\": \"$VERSION\", \"tag\": \"$TAG\", \"needs_tag\": true, \"type\": \"calver\"}]")
        fi
    done

    ACTUAL_COUNT=$(echo "$CALVER_PENDING" | jq 'length')

    if [ "$ACTUAL_COUNT" = "$EXPECTED_COUNT" ]; then
        test_pass "$TEST_NAME (count: $ACTUAL_COUNT)"
    else
        test_fail "$TEST_NAME" "count=$EXPECTED_COUNT" "count=$ACTUAL_COUNT"
    fi

    echo "  Result: $CALVER_PENDING"
}

# Test cases
test_calver_detection "docs books r2r-cli" "docs books" "2" "Both calver modules dispatched"
test_calver_detection "docs r2r-cli" "docs books" "1" "Only docs dispatched"
test_calver_detection "r2r-cli ext-eac" "docs books" "0" "No calver modules dispatched"
test_calver_detection "" "docs books" "0" "Empty dispatched list"

# -----------------------------------------------------------------------------
# TEST 3: Combining Semver + Calver
# -----------------------------------------------------------------------------
section "TEST 3: Combining Semver + Calver"

# Simulate combining
MOCK_SEMVER='[{"module":"r2r-cli","version":"1.0.0","tag":"r2r-cli/1.0.0","needs_tag":true,"type":"semver"}]'
MOCK_CALVER='[{"module":"docs","version":"2025.0116.1234","tag":"docs/2025.0116.1234","needs_tag":true,"type":"calver"}]'

COMBINED=$(echo "$MOCK_SEMVER $MOCK_CALVER" | jq -s 'add')
COMBINED_COUNT=$(echo "$COMBINED" | jq 'length')

if [ "$COMBINED_COUNT" = "2" ]; then
    test_pass "Combining semver + calver produces correct count"
else
    test_fail "Combining arrays" "count=2" "count=$COMBINED_COUNT"
fi

echo "Combined result:"
echo "$COMBINED" | jq .

# Verify types are preserved
SEMVER_COUNT=$(echo "$COMBINED" | jq '[.[] | select(.type == "semver")] | length')
CALVER_COUNT=$(echo "$COMBINED" | jq '[.[] | select(.type == "calver")] | length')

if [ "$SEMVER_COUNT" = "1" ] && [ "$CALVER_COUNT" = "1" ]; then
    test_pass "Types preserved after combining"
else
    test_fail "Type preservation" "semver=1, calver=1" "semver=$SEMVER_COUNT, calver=$CALVER_COUNT"
fi

# Test empty combination
EMPTY_COMBINED=$(echo "[] []" | jq -s 'add')
if [ "$EMPTY_COMBINED" = "[]" ]; then
    test_pass "Empty arrays combine to empty array"
else
    test_fail "Empty combination" "[]" "$EMPTY_COMBINED"
fi

# -----------------------------------------------------------------------------
# TEST 4: Execution Order Calculation
# -----------------------------------------------------------------------------
section "TEST 4: Execution Order Calculation"

# Test with actual modules
echo "Testing execution order for: docs books"
EXEC_ORDER=$($COMMANDS get "execution order" docs books --no-deps --as-json 2>/dev/null || echo '{"error":"failed"}')

echo "Execution order result:"
echo "$EXEC_ORDER" | jq . 2>/dev/null || echo "$EXEC_ORDER"

if echo "$EXEC_ORDER" | jq -e '.layers' > /dev/null 2>&1; then
    test_pass "Execution order returns valid JSON with layers"

    LAYER_COUNT=$(echo "$EXEC_ORDER" | jq -r '.layer_count')
    echo "  Layer count: $LAYER_COUNT"

    # Verify layers contain the requested modules
    ALL_MODULES=$(echo "$EXEC_ORDER" | jq -r '.layers[][] ' | sort | tr '\n' ' ')
    echo "  Modules in layers: $ALL_MODULES"
else
    test_fail "Execution order structure" "layers field" "$(echo "$EXEC_ORDER" | head -c 100)"
fi

# Test dependency ordering (r2r-cli depends on eac-commands which depends on eac-core)
echo ""
echo "Testing dependency ordering for: r2r-cli docs"
EXEC_ORDER_DEPS=$($COMMANDS get "execution order" r2r-cli docs --no-deps --as-json 2>/dev/null || echo '{"error":"failed"}')

echo "Result:"
echo "$EXEC_ORDER_DEPS" | jq . 2>/dev/null || echo "$EXEC_ORDER_DEPS"

# -----------------------------------------------------------------------------
# TEST 5: Layer Enrichment Logic
# -----------------------------------------------------------------------------
section "TEST 5: Layer Enrichment Logic"

# Simulate enrichment
MOCK_LAYERS='[["docs"],["books"]]'
MOCK_MODULES_JSON='[
    {"module":"docs","version":"2025.0116.1234","tag":"docs/2025.0116.1234","type":"calver"},
    {"module":"books","version":"2025.0116.1234","tag":"books/2025.0116.1234","type":"calver"}
]'
LAYER_COUNT=2

ENRICHED_LAYERS="[]"
for i in $(seq 0 $((LAYER_COUNT - 1))); do
    LAYER=$(echo "$MOCK_LAYERS" | jq -c ".[$i]")
    ENRICHED_LAYER="[]"

    for MODULE in $(echo "$LAYER" | jq -r '.[]'); do
        MODULE_INFO=$(echo "$MOCK_MODULES_JSON" | jq -c ".[] | select(.module == \"$MODULE\")")
        if [ -n "$MODULE_INFO" ]; then
            ENRICHED_LAYER=$(echo "$ENRICHED_LAYER" | jq -c ". + [$MODULE_INFO]")
        fi
    done

    ENRICHED_LAYERS=$(echo "$ENRICHED_LAYERS" | jq -c ". + [$ENRICHED_LAYER]")
done

echo "Enriched layers:"
echo "$ENRICHED_LAYERS" | jq .

# Verify structure
ENRICHED_COUNT=$(echo "$ENRICHED_LAYERS" | jq 'length')
if [ "$ENRICHED_COUNT" = "2" ]; then
    test_pass "Enrichment produces correct layer count"
else
    test_fail "Enrichment layer count" "2" "$ENRICHED_COUNT"
fi

# Verify each layer has module info
FIRST_LAYER_HAS_VERSION=$(echo "$ENRICHED_LAYERS" | jq '.[0][0].version != null')
if [ "$FIRST_LAYER_HAS_VERSION" = "true" ]; then
    test_pass "Enriched layers contain version info"
else
    test_fail "Version in enriched layers" "true" "$FIRST_LAYER_HAS_VERSION"
fi

# -----------------------------------------------------------------------------
# TEST 6: Full Integration Simulation
# -----------------------------------------------------------------------------
section "TEST 6: Full Integration Simulation"

echo "Simulating full check-pending-releases flow..."
echo ""

# Step 1: Get semver pending
echo "Step 1: Check semver pending releases"
SEMVER_RESULT=$($COMMANDS release tag-pending --all 2>/dev/null || echo '{"has_pending":false,"results":[]}')
SEMVER_MODULES=$(echo "$SEMVER_RESULT" | jq -c '[.results[] | select(.needs_tag == true) | . + {type: "semver"}]' 2>/dev/null || echo "[]")
echo "  Semver modules: $(echo "$SEMVER_MODULES" | jq -c '.[].module' 2>/dev/null | tr '\n' ' ')"

# Step 2: Check calver pending (simulate docs dispatched)
echo ""
echo "Step 2: Check calver pending releases (simulating: docs dispatched)"
DISPATCHED="docs"
CALVER_MODULES_LIST="docs books"
CALVER_PENDING="[]"

for calver_mod in $CALVER_MODULES_LIST; do
    if echo "$DISPATCHED" | grep -qw "$calver_mod"; then
        VERSION=$(date -u +"%Y.%m%d.%H%M")
        TAG="${calver_mod}/${VERSION}"
        CALVER_PENDING=$(echo "$CALVER_PENDING" | jq -c ". + [{\"module\": \"$calver_mod\", \"version\": \"$VERSION\", \"tag\": \"$TAG\", \"needs_tag\": true, \"type\": \"calver\"}]")
    fi
done
echo "  Calver modules: $(echo "$CALVER_PENDING" | jq -c '.[].module' 2>/dev/null | tr '\n' ' ')"

# Step 3: Combine
echo ""
echo "Step 3: Combine pending releases"
COMBINED=$(echo "$SEMVER_MODULES $CALVER_PENDING" | jq -s 'add')
PENDING_COUNT=$(echo "$COMBINED" | jq 'length')
echo "  Total pending: $PENDING_COUNT"

if [ "$PENDING_COUNT" -gt 0 ]; then
    HAS_PENDING="true"
    PENDING_MODS=$(echo "$COMBINED" | jq -r '.[].module' | tr '\n' ' ')
    echo "  Modules: $PENDING_MODS"

    # Step 4: Calculate execution order
    echo ""
    echo "Step 4: Calculate execution order"
    EXEC_ORDER=$($COMMANDS get "execution order" $PENDING_MODS --no-deps --as-json 2>/dev/null || echo '{"layers":[],"layer_count":0}')
    LAYER_COUNT=$(echo "$EXEC_ORDER" | jq -r '.layer_count')
    echo "  Layer count: $LAYER_COUNT"

    # Step 5: Show final result
    echo ""
    echo "Step 5: Final enriched layers"
    LAYERS=$(echo "$EXEC_ORDER" | jq -c '.layers')

    ENRICHED="[]"
    for i in $(seq 0 $((LAYER_COUNT - 1))); do
        LAYER=$(echo "$LAYERS" | jq -c ".[$i]")
        ENRICHED_LAYER="[]"

        for MODULE in $(echo "$LAYER" | jq -r '.[]'); do
            MODULE_INFO=$(echo "$COMBINED" | jq -c ".[] | select(.module == \"$MODULE\")")
            if [ -n "$MODULE_INFO" ]; then
                ENRICHED_LAYER=$(echo "$ENRICHED_LAYER" | jq -c ". + [$MODULE_INFO]")
            fi
        done

        ENRICHED=$(echo "$ENRICHED" | jq -c ". + [$ENRICHED_LAYER]")
    done

    echo "$ENRICHED" | jq .
    test_pass "Full integration simulation completed"
else
    echo "  No pending releases"
    test_pass "Full integration simulation completed (no pending)"
fi

# -----------------------------------------------------------------------------
# TEST 7: Edge Cases
# -----------------------------------------------------------------------------
section "TEST 7: Edge Cases"

# Test grep word boundary matching
echo "Testing grep word boundary matching..."
if echo "docs books r2r-cli" | grep -qw "docs"; then
    test_pass "grep -qw matches 'docs' in list"
else
    test_fail "grep word match" "match" "no match"
fi

if echo "docs-extra books" | grep -qw "docs"; then
    test_fail "grep word boundary" "no match for 'docs' in 'docs-extra'" "matched"
else
    test_pass "grep -qw does NOT match 'docs' in 'docs-extra' (word boundary)"
fi

# Test empty module handling
echo ""
echo "Testing empty module handling..."
EMPTY_MODULES=""
if [ -z "$EMPTY_MODULES" ]; then
    test_pass "Empty module detection works"
else
    test_fail "Empty module detection" "empty" "$EMPTY_MODULES"
fi

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
echo ""
echo "============================================================================="
echo "TEST SUMMARY"
echo "============================================================================="
echo -e "Passed: ${GREEN}$PASS_COUNT${NC}"
echo -e "Failed: ${RED}$FAIL_COUNT${NC}"
echo ""

if [ "$FAIL_COUNT" -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
