package tui3

import (
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/adapters/tui/tui3/cells"
)

func TestModel_UpdateMetrics(t *testing.T) {
	t.Run("populates cpuPercents after call", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// Call updateMetrics to collect system metrics
		m.updateMetrics()

		// cpuPercents should be populated (one entry per CPU core)
		if len(m.cpuPercents) == 0 {
			t.Error("cpuPercents should not be empty after updateMetrics()")
		}
	})

	t.Run("populates memPercent after call", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// Call updateMetrics to collect system metrics
		m.updateMetrics()

		// memPercent should be > 0 (some memory is always in use)
		if m.memPercent <= 0 {
			t.Errorf("memPercent should be > 0, got %v", m.memPercent)
		}
	})

	t.Run("rate limits updates", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// First call should update lastMetricsUpdate
		m.updateMetrics()
		firstUpdate := m.lastMetricsUpdate

		if firstUpdate.IsZero() {
			t.Fatal("lastMetricsUpdate should be set after first call")
		}

		// Second call immediately after should be rate-limited (not update timestamp)
		m.updateMetrics()
		secondUpdate := m.lastMetricsUpdate

		// The timestamps should be the same because the second call was rate-limited
		if !firstUpdate.Equal(secondUpdate) {
			t.Errorf("rate limiting failed: timestamps differ (first=%v, second=%v)", firstUpdate, secondUpdate)
		}
	})

	t.Run("updates after rate limit expires", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// First call
		m.updateMetrics()
		firstUpdate := m.lastMetricsUpdate

		// Simulate time passing by setting lastMetricsUpdate to the past
		m.lastMetricsUpdate = time.Now().Add(-2 * time.Second)
		expiredTimestamp := m.lastMetricsUpdate

		// Now updateMetrics should update again
		m.updateMetrics()

		// The timestamp should have been updated
		if m.lastMetricsUpdate.Equal(expiredTimestamp) {
			t.Error("updateMetrics should update after rate limit expires")
		}
		if m.lastMetricsUpdate.Before(firstUpdate) {
			t.Error("new timestamp should be after first update")
		}
	})
}

func TestModel_DeriveUoWCounters(t *testing.T) {
	tests := []struct {
		name        string
		units       []cells.SelectorUnit
		wantTotal   int
		wantRunning int
		wantDone    int
		wantCached  int
		wantFailed  int
	}{
		{
			name:        "empty units",
			units:       []cells.SelectorUnit{},
			wantTotal:   0,
			wantRunning: 0,
			wantDone:    0,
			wantCached:  0,
			wantFailed:  0,
		},
		{
			name: "all pending",
			units: []cells.SelectorUnit{
				{Moniker: "mod-a", Status: cells.UnitPending},
				{Moniker: "mod-b", Status: cells.UnitPending},
				{Moniker: "mod-c", Status: cells.UnitPending},
			},
			wantTotal:   3,
			wantRunning: 0,
			wantDone:    0,
			wantCached:  0,
			wantFailed:  0,
		},
		{
			name: "all running",
			units: []cells.SelectorUnit{
				{Moniker: "mod-a", Status: cells.UnitRunning},
				{Moniker: "mod-b", Status: cells.UnitRunning},
			},
			wantTotal:   2,
			wantRunning: 2,
			wantDone:    0,
			wantCached:  0,
			wantFailed:  0,
		},
		{
			name: "all complete",
			units: []cells.SelectorUnit{
				{Moniker: "mod-a", Status: cells.UnitComplete},
				{Moniker: "mod-b", Status: cells.UnitComplete},
			},
			wantTotal:   2,
			wantRunning: 0,
			wantDone:    2,
			wantCached:  0,
			wantFailed:  0,
		},
		{
			name: "all skipped counts as cached",
			units: []cells.SelectorUnit{
				{Moniker: "mod-a", Status: cells.UnitSkipped},
				{Moniker: "mod-b", Status: cells.UnitSkipped},
			},
			wantTotal:   2,
			wantRunning: 0,
			wantDone:    0,
			wantCached:  2,
			wantFailed:  0,
		},
		{
			name: "all failed",
			units: []cells.SelectorUnit{
				{Moniker: "mod-a", Status: cells.UnitFailed},
				{Moniker: "mod-b", Status: cells.UnitFailed},
			},
			wantTotal:   2,
			wantRunning: 0,
			wantDone:    0,
			wantCached:  0,
			wantFailed:  2,
		},
		{
			name: "mixed statuses",
			units: []cells.SelectorUnit{
				{Moniker: "mod-a", Status: cells.UnitPending},
				{Moniker: "mod-b", Status: cells.UnitQueued},
				{Moniker: "mod-c", Status: cells.UnitRunning},
				{Moniker: "mod-d", Status: cells.UnitRunning},
				{Moniker: "mod-e", Status: cells.UnitComplete},
				{Moniker: "mod-f", Status: cells.UnitComplete},
				{Moniker: "mod-g", Status: cells.UnitComplete},
				{Moniker: "mod-h", Status: cells.UnitSkipped},
				{Moniker: "mod-i", Status: cells.UnitFailed},
			},
			wantTotal:   9,
			wantRunning: 2,
			wantDone:    3,
			wantCached:  1,
			wantFailed:  1,
		},
		{
			name: "queued does not count as running",
			units: []cells.SelectorUnit{
				{Moniker: "mod-a", Status: cells.UnitQueued},
				{Moniker: "mod-b", Status: cells.UnitQueued},
				{Moniker: "mod-c", Status: cells.UnitRunning},
			},
			wantTotal:   3,
			wantRunning: 1,
			wantDone:    0,
			wantCached:  0,
			wantFailed:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(80, 24, "test", false)

			// Add all units
			for _, u := range tt.units {
				m.AddUnit(u.Moniker, u.DisplayName, u.Status, u.Weight)
			}

			// Call updateCells which should derive counters
			m.updateCells()

			// Verify counters
			if m.uowTotal != tt.wantTotal {
				t.Errorf("uowTotal = %d, want %d", m.uowTotal, tt.wantTotal)
			}
			if m.uowRunning != tt.wantRunning {
				t.Errorf("uowRunning = %d, want %d", m.uowRunning, tt.wantRunning)
			}
			if m.uowDone != tt.wantDone {
				t.Errorf("uowDone = %d, want %d", m.uowDone, tt.wantDone)
			}
			if m.uowCached != tt.wantCached {
				t.Errorf("uowCached = %d, want %d", m.uowCached, tt.wantCached)
			}
			if m.uowFailed != tt.wantFailed {
				t.Errorf("uowFailed = %d, want %d", m.uowFailed, tt.wantFailed)
			}
		})
	}
}

func TestModel_DeriveUoWCounters_AfterStatusUpdate(t *testing.T) {
	t.Run("counters update when unit status changes", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// Add units with initial statuses
		m.AddUnit("mod-a", "mod-a", cells.UnitPending, 1)
		m.AddUnit("mod-b", "mod-b", cells.UnitPending, 1)
		m.AddUnit("mod-c", "mod-c", cells.UnitPending, 1)

		m.updateCells()

		// Verify initial state
		if m.uowTotal != 3 {
			t.Errorf("initial uowTotal = %d, want 3", m.uowTotal)
		}
		if m.uowRunning != 0 {
			t.Errorf("initial uowRunning = %d, want 0", m.uowRunning)
		}

		// Update one unit to running
		m.UpdateUnit("mod-a", cells.UnitRunning)
		m.updateCells()

		if m.uowRunning != 1 {
			t.Errorf("after update uowRunning = %d, want 1", m.uowRunning)
		}

		// Complete the running unit
		m.UpdateUnit("mod-a", cells.UnitComplete)
		m.updateCells()

		if m.uowRunning != 0 {
			t.Errorf("after complete uowRunning = %d, want 0", m.uowRunning)
		}
		if m.uowDone != 1 {
			t.Errorf("after complete uowDone = %d, want 1", m.uowDone)
		}
	})
}

func TestModel_AddUnit_PreventsDuplicates(t *testing.T) {
	t.Run("duplicate registration only keeps one entry", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// Register unit twice (simulates SetInitSummary + StartUoW race)
		m.AddUnit("build:mod:go:go", "mod:go", cells.UnitPending, 1)
		m.AddUnit("build:mod:go:go", "mod:go", cells.UnitPending, 1)

		// Should only have one entry
		if len(m.units) != 1 {
			t.Errorf("expected 1 unit, got %d (duplicate registration)", len(m.units))
		}
	})

	t.Run("duplicate with different status updates existing", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// First registration as pending
		m.AddUnit("build:mod:go:go", "mod:go", cells.UnitPending, 1)
		// Second registration as running (should update, not add)
		m.AddUnit("build:mod:go:go", "mod:go", cells.UnitRunning, 1)

		if len(m.units) != 1 {
			t.Errorf("expected 1 unit, got %d", len(m.units))
		}
		if m.units[0].Status != cells.UnitRunning {
			t.Errorf("expected status Running, got %v", m.units[0].Status)
		}
	})

	t.Run("counters correct after duplicate registration and completion", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// Simulate the bug scenario: unit registered twice
		m.AddUnit("build:mod:go:go", "mod:go", cells.UnitPending, 1)
		m.AddUnit("build:mod:go:go", "mod:go", cells.UnitPending, 1) // duplicate

		// Update to running
		m.UpdateUnit("build:mod:go:go", cells.UnitRunning)
		m.updateCells()

		if m.uowRunning != 1 {
			t.Errorf("expected 1 running, got %d", m.uowRunning)
		}

		// Complete
		m.UpdateUnit("build:mod:go:go", cells.UnitComplete)
		m.updateCells()

		if m.uowRunning != 0 {
			t.Errorf("expected 0 running after complete, got %d", m.uowRunning)
		}
		if m.uowDone != 1 {
			t.Errorf("expected 1 done, got %d", m.uowDone)
		}
	})
}

func TestModel_DeriveUoWCounters_Weighted(t *testing.T) {
	t.Run("running counts weighted sum not unit count", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// Add units with different weights
		m.AddUnit("build:mod-a:go:go", "mod-a:go", cells.UnitRunning, 1)  // weight 1
		m.AddUnit("build:mod-b:pdf:pdf-tool", "mod-b:pdf", cells.UnitRunning, 4) // weight 4 (PDF)
		m.AddUnit("build:mod-c:site:mkdocs", "mod-c:site", cells.UnitPending, 1) // weight 1 (not running)

		m.updateCells()

		// Running should be 1 + 4 = 5 (weighted sum), not 2 (unit count)
		if m.uowRunning != 5 {
			t.Errorf("uowRunning = %d, want 5 (weighted sum: 1+4)", m.uowRunning)
		}

		// Total is still unit count
		if m.uowTotal != 3 {
			t.Errorf("uowTotal = %d, want 3", m.uowTotal)
		}
	})

	t.Run("weight defaults to 1 when not set", func(t *testing.T) {
		m := NewModel(80, 24, "test", false)

		// Add units without weight (Weight=0)
		m.AddUnit("build:mod-a:go:go", "mod-a:go", cells.UnitRunning, 0) // weight defaults to 1
		m.AddUnit("build:mod-b:go:go", "mod-b:go", cells.UnitRunning, 0) // weight defaults to 1

		m.updateCells()

		// Running should be 1 + 1 = 2 (each defaults to weight 1)
		if m.uowRunning != 2 {
			t.Errorf("uowRunning = %d, want 2 (default weight 1 each)", m.uowRunning)
		}
	})

	t.Run("weight preserved after Pending->Running transition via UpdateUnit", func(t *testing.T) {
		// This test mimics the exact runtime flow:
		// 1. UnitStartMsg adds unit with weight and Pending status
		// 2. UnitRunningMsg updates status to Running via UpdateUnit
		// 3. updateCells should use the preserved weight
		m := NewModel(80, 24, "test", false)

		// Add 3 docs units with weight 5 each (same as real docs-pdf)
		// These are added as Pending (mimics UnitStartMsg in update.go)
		m.AddUnit("build:books:howto:docs-pdf", "books:howto", cells.UnitPending, 5)
		m.AddUnit("build:books:developer:docs-pdf", "books:developer", cells.UnitPending, 5)
		m.AddUnit("build:books:operator:docs-pdf", "books:operator", cells.UnitPending, 5)

		m.updateCells()

		// All pending, running should be 0
		if m.uowRunning != 0 {
			t.Errorf("initial uowRunning = %d, want 0 (all pending)", m.uowRunning)
		}

		// Now update all 3 to Running (mimics UnitRunningMsg in update.go)
		m.UpdateUnit("build:books:howto:docs-pdf", cells.UnitRunning)
		m.UpdateUnit("build:books:developer:docs-pdf", cells.UnitRunning)
		m.UpdateUnit("build:books:operator:docs-pdf", cells.UnitRunning)

		m.updateCells()

		// Running should be 5 + 5 + 5 = 15 (weighted sum), NOT 3 (unit count)
		// This is the bug the user reported: "3x5 != 3"
		if m.uowRunning != 15 {
			t.Errorf("uowRunning = %d, want 15 (3 units × weight 5 each)", m.uowRunning)
		}

		// Verify individual unit weights are preserved
		for _, unit := range m.units {
			if unit.Weight != 5 {
				t.Errorf("unit %s weight = %d, want 5 (weight should be preserved)", unit.Moniker, unit.Weight)
			}
		}
	})
}
