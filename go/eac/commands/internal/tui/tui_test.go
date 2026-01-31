package tui

import (
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestSendAsync_NonBlocking verifies that sendAsync() returns immediately
// even when the message buffer is full or the consumer is slow.
func TestSendAsync_NonBlocking(t *testing.T) {
	c := New(Config{Height: 10})

	// Manually initialize the message channel (normally done in Start())
	c.msgChan = make(chan tea.Msg, 2) // Small buffer to easily fill

	// Fill the buffer
	c.msgChan <- tea.KeyMsg{}
	c.msgChan <- tea.KeyMsg{}

	// sendAsync should return immediately even with full buffer
	done := make(chan struct{})
	go func() {
		c.sendAsync(tea.KeyMsg{}) // This should not block
		close(done)
	}()

	select {
	case <-done:
		// Success - sendAsync returned without blocking
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sendAsync() blocked when buffer was full - should drop message and return immediately")
	}

	// Drain the channel
	close(c.msgChan)
	for range c.msgChan {
	}
}

// TestSendAsync_DropsWhenStopped verifies that sendAsync() safely handles
// calls after the console has been stopped.
func TestSendAsync_DropsWhenStopped(t *testing.T) {
	c := New(Config{Height: 10})
	c.msgChan = make(chan tea.Msg, 10)

	// Mark as stopped
	c.mu.Lock()
	c.stopped = true
	c.mu.Unlock()

	// Should not panic or block
	done := make(chan struct{})
	go func() {
		c.sendAsync(tea.KeyMsg{})
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sendAsync() blocked when console was stopped")
	}

	// Channel should be empty (message was dropped)
	select {
	case <-c.msgChan:
		t.Fatal("sendAsync() should not send when stopped")
	default:
		// Expected - no message sent
	}
}

// TestSendAsync_DropsWhenNilChannel verifies that sendAsync() safely handles
// calls when msgChan is nil.
func TestSendAsync_DropsWhenNilChannel(t *testing.T) {
	c := New(Config{Height: 10})
	// msgChan is nil by default

	// Should not panic
	done := make(chan struct{})
	go func() {
		c.sendAsync(tea.KeyMsg{})
		close(done)
	}()

	select {
	case <-done:
		// Success - no panic
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sendAsync() blocked with nil channel")
	}
}

// TestSendAsync_ConcurrentSafe verifies that sendAsync() can be called
// concurrently from multiple goroutines without races or panics.
func TestSendAsync_ConcurrentSafe(t *testing.T) {
	c := New(Config{Height: 10})
	c.msgChan = make(chan tea.Msg, 100)

	// Start a consumer
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		for range c.msgChan {
			// Consume messages
		}
	}()

	// Spawn many concurrent senders
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.sendAsync(tea.KeyMsg{})
			}
		}()
	}

	// All senders should complete quickly
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - all senders completed
	case <-time.After(5 * time.Second):
		t.Fatal("sendAsync() goroutines did not complete - possible deadlock")
	}

	// Clean up
	close(c.msgChan)
	<-consumerDone
}

// TestSendAsync_MessageDelivery verifies that messages sent via sendAsync()
// are actually delivered when the channel has capacity.
func TestSendAsync_MessageDelivery(t *testing.T) {
	c := New(Config{Height: 10})
	c.msgChan = make(chan tea.Msg, 10)

	// sendAsync checks c.program != nil, so we need a minimal program
	// We create a dummy program that we won't actually run
	c.program = tea.NewProgram(nil, tea.WithoutRenderer())

	// Send a message
	type testMsg struct{ value int }
	c.sendAsync(testMsg{value: 42})

	// Verify it was delivered
	select {
	case msg := <-c.msgChan:
		if tm, ok := msg.(testMsg); !ok || tm.value != 42 {
			t.Fatalf("received wrong message: %v", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("message was not delivered")
	}
}
