package helputil

import (
	"bytes"
	"strings"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

func TestCategorizeFlags(t *testing.T) {
	t.Run("separates behavior flags from regular flags", func(t *testing.T) {
		flags := []core.FlagSpec{
			{Name: "format", Type: "string"},
			{Name: "with-cache", Type: "bool"},
			{Name: "no-cache", Type: "bool"},
			{Name: "debug", Type: "bool"},
		}

		regular, groups := CategorizeFlags(flags)

		if len(regular) != 2 {
			t.Errorf("Expected 2 regular flags, got %d", len(regular))
		}
		if len(groups) != 1 {
			t.Errorf("Expected 1 behavior group, got %d", len(groups))
		}

		// Check regular flags
		if regular[0].Name != "format" || regular[1].Name != "debug" {
			t.Error("Expected format and debug as regular flags")
		}

		// Check behavior group
		group := groups[0]
		if group.Behavior != "cache" {
			t.Errorf("Expected Behavior 'cache', got %q", group.Behavior)
		}
		if group.EnableFlag == nil || group.EnableFlag.Name != "with-cache" {
			t.Error("Expected EnableFlag to be 'with-cache'")
		}
		if group.DisableFlag == nil || group.DisableFlag.Name != "no-cache" {
			t.Error("Expected DisableFlag to be 'no-cache'")
		}
	})

	t.Run("handles multiple behavior groups", func(t *testing.T) {
		flags := []core.FlagSpec{
			{Name: "with-cache", Type: "bool"},
			{Name: "no-cache", Type: "bool"},
			{Name: "with-tui", Type: "bool"},
			{Name: "no-tui", Type: "bool"},
		}

		regular, groups := CategorizeFlags(flags)

		if len(regular) != 0 {
			t.Errorf("Expected 0 regular flags, got %d", len(regular))
		}
		if len(groups) != 2 {
			t.Errorf("Expected 2 behavior groups, got %d", len(groups))
		}

		// Verify both groups exist
		behaviors := make(map[string]bool)
		for _, g := range groups {
			behaviors[g.Behavior] = true
		}
		if !behaviors["cache"] || !behaviors["tui"] {
			t.Error("Expected both cache and tui behavior groups")
		}
	})

	t.Run("handles empty flags", func(t *testing.T) {
		regular, groups := CategorizeFlags(nil)

		if len(regular) != 0 {
			t.Errorf("Expected 0 regular flags, got %d", len(regular))
		}
		if len(groups) != 0 {
			t.Errorf("Expected 0 behavior groups, got %d", len(groups))
		}
	})

	t.Run("preserves order of behavior groups", func(t *testing.T) {
		flags := []core.FlagSpec{
			{Name: "with-cache", Type: "bool"},
			{Name: "no-cache", Type: "bool"},
			{Name: "with-tui", Type: "bool"},
			{Name: "no-tui", Type: "bool"},
			{Name: "with-parallel", Type: "bool"},
			{Name: "no-parallel", Type: "bool"},
		}

		_, groups := CategorizeFlags(flags)

		if len(groups) != 3 {
			t.Fatalf("Expected 3 groups, got %d", len(groups))
		}
		if groups[0].Behavior != "cache" {
			t.Errorf("Expected first group to be 'cache', got %q", groups[0].Behavior)
		}
		if groups[1].Behavior != "tui" {
			t.Errorf("Expected second group to be 'tui', got %q", groups[1].Behavior)
		}
		if groups[2].Behavior != "parallel" {
			t.Errorf("Expected third group to be 'parallel', got %q", groups[2].Behavior)
		}
	})
}

func TestFormatDefault(t *testing.T) {
	t.Run("returns ON for true default", func(t *testing.T) {
		flag := &core.FlagSpec{DefaultValue: "true"}
		result := FormatDefault(flag)
		if result != "ON" {
			t.Errorf("Expected 'ON', got %q", result)
		}
	})

	t.Run("returns OFF for false default", func(t *testing.T) {
		flag := &core.FlagSpec{DefaultValue: "false"}
		result := FormatDefault(flag)
		if result != "OFF" {
			t.Errorf("Expected 'OFF', got %q", result)
		}
	})
}

func TestPrintBehaviorGroup(t *testing.T) {
	t.Run("prints enable/disable pair", func(t *testing.T) {
		group := BehaviorGroup{
			Behavior: "cache",
			EnableFlag: &core.FlagSpec{
				Name:         "with-cache",
				Usage:        "Enable incremental caching",
				DefaultValue: "true",
			},
			DisableFlag: &core.FlagSpec{
				Name:  "no-cache",
				Usage: "Disable incremental caching",
			},
		}

		var buf bytes.Buffer
		PrintBehaviorGroup(&buf, group)

		output := buf.String()
		if !strings.Contains(output, "--with-cache / --no-cache") {
			t.Errorf("Expected flag pair in output, got %q", output)
		}
		if !strings.Contains(output, "Enable incremental caching") {
			t.Errorf("Expected usage in output, got %q", output)
		}
		if !strings.Contains(output, "(default: ON)") {
			t.Errorf("Expected default in output, got %q", output)
		}
	})

	t.Run("handles incomplete group gracefully", func(t *testing.T) {
		group := BehaviorGroup{
			Behavior:   "cache",
			EnableFlag: nil,
		}

		var buf bytes.Buffer
		PrintBehaviorGroup(&buf, group)

		output := buf.String()
		if output != "" {
			t.Errorf("Expected empty output for incomplete group, got %q", output)
		}
	})
}

func TestPrintBehaviorFlags(t *testing.T) {
	t.Run("prints section header and groups", func(t *testing.T) {
		groups := []BehaviorGroup{
			{
				Behavior: "cache",
				EnableFlag: &core.FlagSpec{
					Name:         "with-cache",
					Usage:        "Enable caching",
					DefaultValue: "true",
				},
				DisableFlag: &core.FlagSpec{Name: "no-cache"},
			},
		}

		var buf bytes.Buffer
		PrintBehaviorFlags(&buf, groups)

		output := buf.String()
		if !strings.Contains(output, "Behavior Flags:") {
			t.Errorf("Expected section header, got %q", output)
		}
	})

	t.Run("skips section for empty groups", func(t *testing.T) {
		var buf bytes.Buffer
		PrintBehaviorFlags(&buf, nil)

		output := buf.String()
		if output != "" {
			t.Errorf("Expected empty output for no groups, got %q", output)
		}
	})
}
