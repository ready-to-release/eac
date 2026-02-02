package tui

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestExitHoldController_NoHolds_ReturnsImmediately(t *testing.T) {
	ctrl := NewExitHoldController()

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

func TestExitHoldController_SingleHold_BlocksUntilRelease(t *testing.T) {
	ctrl := NewExitHoldController()

	release := ctrl.HoldExit()

	// Start goroutine to release after delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		release()
	}()

	start := time.Now()
	result := ctrl.WaitForRelease(context.Background(), 5*time.Second)
	elapsed := time.Since(start)

	if !result {
		t.Error("expected WaitForRelease to return true after release")
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("expected to block for ~50ms, only took %v", elapsed)
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("expected to unblock quickly after release, took %v", elapsed)
	}
}

func TestExitHoldController_MultipleHolds_RequiresAllReleased(t *testing.T) {
	ctrl := NewExitHoldController()

	release1 := ctrl.HoldExit()
	release2 := ctrl.HoldExit()
	release3 := ctrl.HoldExit()

	var wg sync.WaitGroup
	wg.Add(1)

	// Start goroutine to wait for release
	var result bool
	var elapsed time.Duration
	go func() {
		defer wg.Done()
		start := time.Now()
		result = ctrl.WaitForRelease(context.Background(), 5*time.Second)
		elapsed = time.Since(start)
	}()

	// Release holds one by one with delays
	time.Sleep(20 * time.Millisecond)
	release1()
	time.Sleep(20 * time.Millisecond)
	release2()
	time.Sleep(20 * time.Millisecond)
	release3()

	wg.Wait()

	if !result {
		t.Error("expected WaitForRelease to return true")
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("expected to block until all released (~60ms), only took %v", elapsed)
	}
}

func TestExitHoldController_Timeout_ReturnsFalse(t *testing.T) {
	ctrl := NewExitHoldController()

	_ = ctrl.HoldExit() // Hold without releasing

	start := time.Now()
	result := ctrl.WaitForRelease(context.Background(), 50*time.Millisecond)
	elapsed := time.Since(start)

	if result {
		t.Error("expected WaitForRelease to return false on timeout")
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("expected to wait for timeout (~50ms), only waited %v", elapsed)
	}
	if elapsed > 150*time.Millisecond {
		t.Errorf("expected to return after timeout, took %v", elapsed)
	}
}

func TestExitHoldController_ContextCancelled_ReturnsFalse(t *testing.T) {
	ctrl := NewExitHoldController()

	_ = ctrl.HoldExit() // Hold without releasing

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result := ctrl.WaitForRelease(ctx, 5*time.Second)
	elapsed := time.Since(start)

	if result {
		t.Error("expected WaitForRelease to return false on context cancellation")
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("expected to wait for cancellation (~50ms), only waited %v", elapsed)
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("expected to return after cancellation, took %v", elapsed)
	}
}

func TestExitHoldController_ReleaseIdempotent(t *testing.T) {
	ctrl := NewExitHoldController()

	release := ctrl.HoldExit()

	// Call release multiple times - should not panic and not decrement below 0
	release()
	release()
	release()

	// Should still return immediately (no holds)
	result := ctrl.WaitForRelease(context.Background(), 100*time.Millisecond)
	if !result {
		t.Error("expected WaitForRelease to return true")
	}
}

func TestExitHoldController_ConcurrentAccess(t *testing.T) {
	ctrl := NewExitHoldController()

	// Test concurrent holds and releases
	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release := ctrl.HoldExit()
			time.Sleep(time.Duration(10+i*5) * time.Millisecond)
			release()
		}()
	}

	// Wait for all to start
	time.Sleep(5 * time.Millisecond)

	// Wait for release should eventually succeed
	result := ctrl.WaitForRelease(context.Background(), 2*time.Second)
	if !result {
		t.Error("expected WaitForRelease to return true after all released")
	}

	wg.Wait()
}

func TestExitHoldController_HoldAfterRelease(t *testing.T) {
	ctrl := NewExitHoldController()

	// First hold-release cycle
	release1 := ctrl.HoldExit()
	release1()

	// Verify no holds
	if !ctrl.WaitForRelease(context.Background(), 100*time.Millisecond) {
		t.Error("expected no holds after release")
	}

	// Second hold-release cycle
	release2 := ctrl.HoldExit()

	// Start waiter
	done := make(chan bool)
	go func() {
		done <- ctrl.WaitForRelease(context.Background(), 5*time.Second)
	}()

	// Release after delay
	time.Sleep(20 * time.Millisecond)
	release2()

	result := <-done
	if !result {
		t.Error("expected WaitForRelease to return true")
	}
}
