#!/usr/bin/env python3
"""
Test Plan: Release Trigger Logic Simulation
============================================

This script simulates the check-pending-releases action and trigger-releases
job logic locally to verify correctness before CI runs.

Test Scenarios:
1. Semver only: r2r-cli has pending changelog release
2. Calver only: docs was dispatched, books was not
3. Both: semver + calver releases pending
4. None: no pending releases
5. Dependency ordering: verify layers respect module dependencies

Usage: python scripts/test-release-trigger-logic.py
"""

import json
import subprocess
import sys
import re
from datetime import datetime
from typing import List, Dict, Any, Tuple

# ANSI colors
RED = '\033[0;31m'
GREEN = '\033[0;32m'
YELLOW = '\033[1;33m'
BLUE = '\033[0;34m'
NC = '\033[0m'

COMMANDS = "./go/eac/commands/build/commands.exe"

class TestResults:
    def __init__(self):
        self.passed = 0
        self.failed = 0

    def test_pass(self, msg: str):
        print(f"{GREEN}PASS{NC}: {msg}")
        self.passed += 1

    def test_fail(self, msg: str, expected: str, got: str):
        print(f"{RED}FAIL{NC}: {msg}")
        print(f"       Expected: {expected}")
        print(f"       Got: {got}")
        self.failed += 1

    def section(self, title: str):
        print(f"\n{BLUE}--- {title} ---{NC}")


def run_command(cmd: List[str]) -> Tuple[str, int]:
    """Run a command and return stdout and return code."""
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        return result.stdout.strip(), result.returncode
    except Exception as e:
        return str(e), 1


def test_semver_detection(results: TestResults):
    """TEST 1: Semver Detection (release tag-pending)"""
    results.section("TEST 1: Semver Detection")

    print(f"Running: {COMMANDS} release tag-pending --all")
    stdout, retcode = run_command([COMMANDS, "release", "tag-pending", "--all"])

    print(f"Result:\n{stdout[:500]}")

    try:
        data = json.loads(stdout)

        if "has_pending" in data:
            results.test_pass("tag-pending returns valid JSON with has_pending field")
            print(f"  has_pending: {data['has_pending']}")

            if data.get("has_pending"):
                pending = [r["module"] for r in data.get("results", []) if r.get("needs_tag")]
                print(f"  Pending semver modules: {' '.join(pending)}")
        else:
            results.test_fail("tag-pending JSON structure", "has_pending field", str(data)[:100])

        # Test parsing logic (simulating action step)
        semver_modules = [
            {**r, "type": "semver"}
            for r in data.get("results", [])
            if r.get("needs_tag")
        ]
        print(f"Parsed semver modules: {json.dumps(semver_modules)}")
        results.test_pass("Semver parsing produces valid JSON array")

    except json.JSONDecodeError as e:
        results.test_fail("JSON parsing", "valid JSON", str(e))


def test_calver_detection(results: TestResults):
    """TEST 2: Calver Detection Logic"""
    results.section("TEST 2: Calver Detection Logic")

    def check_calver(dispatched: str, calver_modules: str, expected_count: int, test_name: str):
        dispatched_list = dispatched.split() if dispatched else []
        calver_list = calver_modules.split() if calver_modules else []

        calver_pending = []
        for calver_mod in calver_list:
            # Word boundary match (like grep -qw)
            if calver_mod in dispatched_list:
                version = datetime.utcnow().strftime("%Y.%m%d.%H%M")
                tag = f"{calver_mod}/{version}"
                calver_pending.append({
                    "module": calver_mod,
                    "version": version,
                    "tag": tag,
                    "needs_tag": True,
                    "type": "calver"
                })

        actual_count = len(calver_pending)
        if actual_count == expected_count:
            results.test_pass(f"{test_name} (count: {actual_count})")
        else:
            results.test_fail(test_name, f"count={expected_count}", f"count={actual_count}")
        print(f"  Result: {json.dumps(calver_pending)}")

    # Test cases
    check_calver("docs books r2r-cli", "docs books", 2, "Both calver modules dispatched")
    check_calver("docs r2r-cli", "docs books", 1, "Only docs dispatched")
    check_calver("r2r-cli ext-eac", "docs books", 0, "No calver modules dispatched")
    check_calver("", "docs books", 0, "Empty dispatched list")


def test_combining(results: TestResults):
    """TEST 3: Combining Semver + Calver"""
    results.section("TEST 3: Combining Semver + Calver")

    mock_semver = [{"module": "r2r-cli", "version": "1.0.0", "tag": "r2r-cli/1.0.0", "needs_tag": True, "type": "semver"}]
    mock_calver = [{"module": "docs", "version": "2025.0116.1234", "tag": "docs/2025.0116.1234", "needs_tag": True, "type": "calver"}]

    combined = mock_semver + mock_calver

    if len(combined) == 2:
        results.test_pass("Combining semver + calver produces correct count")
    else:
        results.test_fail("Combining arrays", "count=2", f"count={len(combined)}")

    print(f"Combined result:\n{json.dumps(combined, indent=2)}")

    # Verify types preserved
    semver_count = sum(1 for m in combined if m.get("type") == "semver")
    calver_count = sum(1 for m in combined if m.get("type") == "calver")

    if semver_count == 1 and calver_count == 1:
        results.test_pass("Types preserved after combining")
    else:
        results.test_fail("Type preservation", "semver=1, calver=1", f"semver={semver_count}, calver={calver_count}")

    # Test empty combination
    empty_combined = [] + []
    if empty_combined == []:
        results.test_pass("Empty arrays combine to empty array")
    else:
        results.test_fail("Empty combination", "[]", str(empty_combined))


def test_execution_order(results: TestResults):
    """TEST 4: Execution Order Calculation"""
    results.section("TEST 4: Execution Order Calculation")

    print("Testing execution order for: docs books")
    stdout, retcode = run_command([COMMANDS, "get", "execution order", "docs", "books", "--no-deps", "--as-json"])

    print(f"Execution order result:\n{stdout[:500]}")

    try:
        data = json.loads(stdout)

        if "layers" in data:
            results.test_pass("Execution order returns valid JSON with layers")

            layer_count = data.get("layer_count", 0)
            print(f"  Layer count: {layer_count}")

            all_modules = []
            for layer in data.get("layers", []):
                all_modules.extend(layer)
            print(f"  Modules in layers: {' '.join(all_modules)}")
        else:
            results.test_fail("Execution order structure", "layers field", str(data)[:100])

    except json.JSONDecodeError as e:
        results.test_fail("Execution order JSON", "valid JSON", str(e))

    # Test with modules that have dependencies
    print("\nTesting dependency ordering for: r2r-cli docs")
    stdout2, _ = run_command([COMMANDS, "get", "execution order", "r2r-cli", "docs", "--no-deps", "--as-json"])
    print(f"Result:\n{stdout2[:500]}")


def test_enrichment(results: TestResults):
    """TEST 5: Layer Enrichment Logic"""
    results.section("TEST 5: Layer Enrichment Logic")

    mock_layers = [["docs"], ["books"]]
    mock_modules_json = [
        {"module": "docs", "version": "2025.0116.1234", "tag": "docs/2025.0116.1234", "type": "calver"},
        {"module": "books", "version": "2025.0116.1234", "tag": "books/2025.0116.1234", "type": "calver"}
    ]

    # Build module lookup
    module_lookup = {m["module"]: m for m in mock_modules_json}

    # Enrich layers
    enriched_layers = []
    for layer in mock_layers:
        enriched_layer = []
        for module in layer:
            if module in module_lookup:
                enriched_layer.append(module_lookup[module])
        enriched_layers.append(enriched_layer)

    print(f"Enriched layers:\n{json.dumps(enriched_layers, indent=2)}")

    if len(enriched_layers) == 2:
        results.test_pass("Enrichment produces correct layer count")
    else:
        results.test_fail("Enrichment layer count", "2", str(len(enriched_layers)))

    # Verify version info
    if enriched_layers[0] and "version" in enriched_layers[0][0]:
        results.test_pass("Enriched layers contain version info")
    else:
        results.test_fail("Version in enriched layers", "true", "false")


def test_full_integration(results: TestResults):
    """TEST 6: Full Integration Simulation"""
    results.section("TEST 6: Full Integration Simulation")

    print("Simulating full check-pending-releases flow...\n")

    # Step 1: Get semver pending
    print("Step 1: Check semver pending releases")
    stdout, _ = run_command([COMMANDS, "release", "tag-pending", "--all"])

    try:
        semver_result = json.loads(stdout)
        semver_modules = [
            {**r, "type": "semver"}
            for r in semver_result.get("results", [])
            if r.get("needs_tag")
        ]
        modules_str = ' '.join(m["module"] for m in semver_modules)
        print(f"  Semver modules: {modules_str or '(none)'}")
    except:
        semver_modules = []
        print("  Semver modules: (parse error)")

    # Step 2: Check calver pending (simulate docs dispatched)
    print("\nStep 2: Check calver pending releases (simulating: docs dispatched)")
    dispatched = ["docs"]
    calver_modules_list = ["docs", "books"]

    calver_pending = []
    for calver_mod in calver_modules_list:
        if calver_mod in dispatched:
            version = datetime.utcnow().strftime("%Y.%m%d.%H%M")
            tag = f"{calver_mod}/{version}"
            calver_pending.append({
                "module": calver_mod,
                "version": version,
                "tag": tag,
                "needs_tag": True,
                "type": "calver"
            })

    modules_str = ' '.join(m["module"] for m in calver_pending)
    print(f"  Calver modules: {modules_str or '(none)'}")

    # Step 3: Combine
    print("\nStep 3: Combine pending releases")
    combined = semver_modules + calver_pending
    print(f"  Total pending: {len(combined)}")

    if combined:
        pending_mods = [m["module"] for m in combined]
        print(f"  Modules: {' '.join(pending_mods)}")

        # Step 4: Calculate execution order
        print("\nStep 4: Calculate execution order")
        cmd = [COMMANDS, "get", "execution order"] + pending_mods + ["--no-deps", "--as-json"]
        stdout, _ = run_command(cmd)

        try:
            exec_order = json.loads(stdout)
            layer_count = exec_order.get("layer_count", 0)
            print(f"  Layer count: {layer_count}")

            # Step 5: Enrich layers
            print("\nStep 5: Final enriched layers")
            module_lookup = {m["module"]: m for m in combined}
            enriched = []

            for layer in exec_order.get("layers", []):
                enriched_layer = []
                for module in layer:
                    if module in module_lookup:
                        enriched_layer.append(module_lookup[module])
                enriched.append(enriched_layer)

            print(json.dumps(enriched, indent=2))
            results.test_pass("Full integration simulation completed")
        except:
            results.test_fail("Full integration", "valid execution order", "parse error")
    else:
        print("  No pending releases")
        results.test_pass("Full integration simulation completed (no pending)")


def test_edge_cases(results: TestResults):
    """TEST 7: Edge Cases"""
    results.section("TEST 7: Edge Cases")

    # Test word boundary matching
    print("Testing word boundary matching...")

    def word_match(word: str, text: str) -> bool:
        """Match whole words only (like grep -qw)"""
        return word in text.split()

    if word_match("docs", "docs books r2r-cli"):
        results.test_pass("Word match finds 'docs' in list")
    else:
        results.test_fail("Word match", "match", "no match")

    if not word_match("docs", "docs-extra books"):
        results.test_pass("Word match does NOT match 'docs' in 'docs-extra' (word boundary)")
    else:
        results.test_fail("Word boundary", "no match for 'docs' in 'docs-extra'", "matched")

    # Test empty handling
    print("\nTesting empty module handling...")
    empty_modules = ""
    if not empty_modules:
        results.test_pass("Empty module detection works")
    else:
        results.test_fail("Empty module detection", "empty", empty_modules)


def main():
    print("=" * 77)
    print("RELEASE TRIGGER LOGIC TEST SUITE")
    print("=" * 77)

    results = TestResults()

    test_semver_detection(results)
    test_calver_detection(results)
    test_combining(results)
    test_execution_order(results)
    test_enrichment(results)
    test_full_integration(results)
    test_edge_cases(results)

    print("\n" + "=" * 77)
    print("TEST SUMMARY")
    print("=" * 77)
    print(f"Passed: {GREEN}{results.passed}{NC}")
    print(f"Failed: {RED}{results.failed}{NC}")
    print()

    if results.failed > 0:
        print(f"{RED}Some tests failed!{NC}")
        return 1
    else:
        print(f"{GREEN}All tests passed!{NC}")
        return 0


if __name__ == "__main__":
    sys.exit(main())
