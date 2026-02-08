package orchestrator

import (
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

func TestUpdateConfig_AppliesNonZeroFields(t *testing.T) {
	config := Config{
		ActionType: core.ActionBuild,
		TUI:        true,
	}
	orch := New(&config, nil)
	defer orch.Close()

	orch.UpdateConfig(ConfigUpdate{
		WorkspaceRoot:  "/workspace",
		OutputBaseDir:  "out/build",
		MaxConcurrency: 4,
		Turbo:          1.25,
		ComponentTypesDisplay: map[string]string{
			"mod:comp": "go",
		},
		ShowTimings: true,
		DryRun:      true,
	})

	if orch.config.WorkspaceRoot != "/workspace" {
		t.Errorf("WorkspaceRoot = %q, want %q", orch.config.WorkspaceRoot, "/workspace")
	}
	if orch.config.OutputBaseDir != "out/build" {
		t.Errorf("OutputBaseDir = %q, want %q", orch.config.OutputBaseDir, "out/build")
	}
	if orch.config.LogFileName() != "build.log" {
		t.Errorf("LogFileName() = %q, want %q", orch.config.LogFileName(), "build.log")
	}
	if orch.config.MaxConcurrency != 4 {
		t.Errorf("MaxConcurrency = %d, want 4", orch.config.MaxConcurrency)
	}
	if orch.config.Turbo != 1.25 {
		t.Errorf("Turbo = %v, want 1.25", orch.config.Turbo)
	}
	if orch.config.ComponentTypesDisplay["mod:comp"] != "go" {
		t.Errorf("ComponentTypesDisplay missing expected entry")
	}
	if !orch.config.ShowTimings {
		t.Error("ShowTimings should be true")
	}
	if !orch.config.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestUpdateConfig_PreservesExistingValues(t *testing.T) {
	config := Config{
		WorkspaceRoot:  "/original",
		OutputBaseDir:  "out/original",
		ActionType:     core.ActionTest,
		MaxConcurrency: 8,
		Turbo:          2.0,
		TUI:            true,
	}
	orch := New(&config, nil)
	defer orch.Close()

	// Update with only some fields set - zero/nil fields should not overwrite
	orch.UpdateConfig(ConfigUpdate{
		WorkspaceRoot: "/updated",
		// OutputBaseDir left empty - should preserve "out/original"
		// MaxConcurrency left 0 - should preserve 8
		// Turbo left 0 - should preserve 2.0
	})

	if orch.config.WorkspaceRoot != "/updated" {
		t.Errorf("WorkspaceRoot = %q, want %q", orch.config.WorkspaceRoot, "/updated")
	}
	if orch.config.OutputBaseDir != "out/original" {
		t.Errorf("OutputBaseDir = %q, want %q (should be preserved)", orch.config.OutputBaseDir, "out/original")
	}
	if orch.config.MaxConcurrency != 8 {
		t.Errorf("MaxConcurrency = %d, want 8 (should be preserved)", orch.config.MaxConcurrency)
	}
	if orch.config.Turbo != 2.0 {
		t.Errorf("Turbo = %v, want 2.0 (should be preserved)", orch.config.Turbo)
	}
	// ActionType is not in ConfigUpdate - should always be preserved
	if orch.config.ActionType != core.ActionTest {
		t.Errorf("ActionType = %q, want %q (should be preserved)", orch.config.ActionType, core.ActionTest)
	}
	if orch.config.ActionVerb() != "Testing" {
		t.Errorf("ActionVerb() = %q, want %q (derived from ActionType)", orch.config.ActionVerb(), "Testing")
	}
}

func TestUpdateConfig_EmptyUpdate(t *testing.T) {
	config := Config{
		WorkspaceRoot:  "/workspace",
		OutputBaseDir:  "out/build",
		ActionType:     core.ActionBuild,
		MaxConcurrency: 4,
		Turbo:          1.25,
	}
	orch := New(&config, nil)
	defer orch.Close()

	// Empty update should not change anything
	orch.UpdateConfig(ConfigUpdate{})

	if orch.config.WorkspaceRoot != "/workspace" {
		t.Errorf("WorkspaceRoot changed unexpectedly")
	}
	if orch.config.OutputBaseDir != "out/build" {
		t.Errorf("OutputBaseDir changed unexpectedly")
	}
	if orch.config.MaxConcurrency != 4 {
		t.Errorf("MaxConcurrency changed unexpectedly")
	}
	if orch.config.Turbo != 1.25 {
		t.Errorf("Turbo changed unexpectedly")
	}
}
