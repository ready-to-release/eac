package tui

import (
	"testing"
	"time"
)

func TestDefaultTUIConfig(t *testing.T) {
	cfg := DefaultTUIConfig()

	// Verify timeouts match expected defaults
	t.Run("timeout defaults", func(t *testing.T) {
		if cfg.MetricsInterval != 500*time.Millisecond {
			t.Errorf("MetricsInterval: got %v, want 500ms", cfg.MetricsInterval)
		}
		if cfg.MinDisplayTime != 1500*time.Millisecond {
			t.Errorf("MinDisplayTime: got %v, want 1500ms", cfg.MinDisplayTime)
		}
		if cfg.ExitCountdown != 10*time.Second {
			t.Errorf("ExitCountdown: got %v, want 10s", cfg.ExitCountdown)
		}
		if cfg.FreezeCountdown != 120*time.Second {
			t.Errorf("FreezeCountdown: got %v, want 120s", cfg.FreezeCountdown)
		}
		if cfg.AutoScrollResume != 8*time.Second {
			t.Errorf("AutoScrollResume: got %v, want 8s", cfg.AutoScrollResume)
		}
	})

	// Verify layout defaults
	t.Run("layout defaults", func(t *testing.T) {
		if cfg.MaxTabs != 36 {
			t.Errorf("MaxTabs: got %d, want 36", cfg.MaxTabs)
		}
		if cfg.DefaultColumns != 4 {
			t.Errorf("DefaultColumns: got %d, want 4", cfg.DefaultColumns)
		}
		if cfg.MinColumns != 2 {
			t.Errorf("MinColumns: got %d, want 2", cfg.MinColumns)
		}
		if cfg.MaxColumns != 6 {
			t.Errorf("MaxColumns: got %d, want 6", cfg.MaxColumns)
		}
		if cfg.BufferSizePane != 500 {
			t.Errorf("BufferSizePane: got %d, want 500", cfg.BufferSizePane)
		}
		if cfg.BufferSizeResults != 100 {
			t.Errorf("BufferSizeResults: got %d, want 100", cfg.BufferSizeResults)
		}
		if cfg.BufferSizeUoW != 200 {
			t.Errorf("BufferSizeUoW: got %d, want 200", cfg.BufferSizeUoW)
		}
	})
}

func TestDefaultTUIConfig_NotNil(t *testing.T) {
	cfg := DefaultTUIConfig()
	if cfg == nil {
		t.Error("DefaultTUIConfig() returned nil")
	}
}
