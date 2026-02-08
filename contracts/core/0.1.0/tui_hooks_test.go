package core

import (
	"context"
	"testing"
	"time"
)

func TestNullExitHoldController_WaitForRelease_ReturnsImmediately(t *testing.T) {
	ctrl := NullExitHoldController{}

	start := time.Now()
	result := ctrl.WaitForRelease(context.Background(), 5*time.Second)
	elapsed := time.Since(start)

	if !result {
		t.Error("expected WaitForRelease to return true")
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("expected immediate return, took %v", elapsed)
	}
}

func TestNullExitHoldController_HoldExit_ReturnsNoopRelease(t *testing.T) {
	ctrl := NullExitHoldController{}

	release := ctrl.HoldExit()
	if release == nil {
		t.Fatal("expected non-nil release function")
	}

	// Should not panic when called
	release()
	release() // Should be idempotent
}

func TestNullCommandSelectionHook_ReturnsOriginalCommand(t *testing.T) {
	hook := NullCommandSelectionHook{}

	req := CommandSelectionRequest{
		OriginalCommand: "test-command",
		Options: []CommandOption{
			{Name: "option1", Description: "First option"},
			{Name: "option2", Description: "Second option"},
		},
	}

	resp := hook.SelectCommand(context.Background(), req)

	if resp.SelectedCommand != "test-command" {
		t.Errorf("expected SelectedCommand=%q, got %q", "test-command", resp.SelectedCommand)
	}
	if resp.Cancelled {
		t.Error("expected Cancelled=false")
	}
}

func TestNullTUIHooks_AllMethodsWork(t *testing.T) {
	hooks := NullTUIHooks{}

	t.Run("SelectCommand", func(t *testing.T) {
		resp := hooks.SelectCommand(context.Background(), CommandSelectionRequest{
			OriginalCommand: "cmd",
		})
		if resp.SelectedCommand != "cmd" {
			t.Errorf("expected %q, got %q", "cmd", resp.SelectedCommand)
		}
	})

	t.Run("ReceiveUoWs", func(t *testing.T) {
		hooks.ReceiveUoWs(UoWData{})
	})

	t.Run("BeforeStart", func(t *testing.T) {
		if err := hooks.BeforeStart(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("HoldExit", func(t *testing.T) {
		release := hooks.HoldExit()
		release()
	})

	t.Run("WaitForRelease", func(t *testing.T) {
		if !hooks.WaitForRelease(context.Background(), time.Second) {
			t.Error("expected true")
		}
	})
}
