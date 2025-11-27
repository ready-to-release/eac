#!/bin/bash
# Test script to verify each repository validation test catches errors

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

echo "=== Repository Validation Test Suite ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

passed=0
failed=0

run_test() {
    local test_name="$1"
    local break_cmd="$2"
    local restore_cmd="$3"
    local test_cmd="$4"

    echo "Testing: $test_name"

    # Break it
    eval "$break_cmd" 2>&1 > /dev/null

    # Run test - should fail
    if eval "$test_cmd" 2>&1 | grep -q "FAIL\\|error\\|❌"; then
        echo -e "${GREEN}  ✓ Test correctly detected the error${NC}"
        ((passed++))
    else
        echo -e "${RED}  ✗ Test FAILED to detect the error${NC}"
        ((failed++))
    fi

    # Restore
    eval "$restore_cmd" 2>&1 > /dev/null

    echo ""
}

# Test 1: markdown-syntax
run_test "markdown-syntax validation" \
    "echo 'Invalid JSON: { broken' >> README.md" \
    "git checkout README.md" \
    "go run ./src/commands validate markdown"

# Test 2: go-modules-tidy
run_test "go-modules-tidy validation" \
    "echo '// temp' >> src/cli/main.go" \
    "git checkout src/cli/main.go" \
    "go run ./src/commands validate go-tidy"

# Test 3: module-hierarchy
run_test "module-hierarchy validation" \
    "sed -i.bak 's/depends_on: \\[\\]/depends_on: [\"nonexistent-module\"]/' contracts/modules/0.1.0/src-cli.yml" \
    "mv contracts/modules/0.1.0/src-cli.yml.bak contracts/modules/0.1.0/src-cli.yml" \
    "go run ./src/commands validate module-hierarchy"

# Test 4: module-files
run_test "module-files validation" \
    "echo 'test' > untracked-file.txt" \
    "rm -f untracked-file.txt" \
    "go run ./src/commands validate module-files"

# Test 5: validate-dependencies
run_test "validate-dependencies validation" \
    "echo 'require github.com/invalid/module v0.0.0' >> src/cli/go.mod" \
    "git checkout src/cli/go.mod" \
    "go run ./src/commands validate dependencies"

# Test 6: test-tags-contracted
run_test "test-tags-contracted validation" \
    "sed -i.bak '1s/^/@invalid-tag\\n/' specs/repository/markdown-syntax/specification.feature" \
    "mv specs/repository/markdown-syntax/specification.feature.bak specs/repository/markdown-syntax/specification.feature" \
    "go run ./src/commands validate test-tags"

echo "==================================="
echo -e "${GREEN}Passed: $passed${NC}"
if [ $failed -gt 0 ]; then
    echo -e "${RED}Failed: $failed${NC}"
    exit 1
else
    echo -e "${GREEN}All validation tests working correctly!${NC}"
fi
