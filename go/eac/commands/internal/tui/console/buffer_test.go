package console

import (
	"testing"
	"time"
)

func TestRingBuffer_Push(t *testing.T) {
	rb := NewRingBuffer(3)

	// Push less than capacity
	rb.Push(Line{Text: "line1", Level: LevelInfo})
	rb.Push(Line{Text: "line2", Level: LevelWarn})

	if rb.Count() != 2 {
		t.Errorf("expected count 2, got %d", rb.Count())
	}

	lines := rb.Last(2)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if lines[0].Text != "line1" || lines[1].Text != "line2" {
		t.Errorf("unexpected line order: %v", lines)
	}
}

func TestRingBuffer_Overflow(t *testing.T) {
	rb := NewRingBuffer(3)

	// Push more than capacity
	rb.Push(Line{Text: "line1"})
	rb.Push(Line{Text: "line2"})
	rb.Push(Line{Text: "line3"})
	rb.Push(Line{Text: "line4"}) // This should overwrite line1

	if rb.Count() != 3 {
		t.Errorf("expected count 3, got %d", rb.Count())
	}

	lines := rb.Last(3)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	// Oldest should be line2, newest should be line4
	if lines[0].Text != "line2" {
		t.Errorf("expected first line to be 'line2', got '%s'", lines[0].Text)
	}
	if lines[2].Text != "line4" {
		t.Errorf("expected last line to be 'line4', got '%s'", lines[2].Text)
	}
}

func TestRingBuffer_LastN(t *testing.T) {
	rb := NewRingBuffer(5)

	for i := 1; i <= 5; i++ {
		rb.Push(Line{Text: string(rune('0' + i))})
	}

	// Request fewer than available
	lines := rb.Last(2)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if lines[0].Text != "4" || lines[1].Text != "5" {
		t.Errorf("unexpected lines: %v", lines)
	}

	// Request more than available
	lines = rb.Last(10)
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
}

func TestRingBuffer_LastByLevel(t *testing.T) {
	rb := NewRingBuffer(10)

	rb.Push(Line{Text: "info1", Level: LevelInfo})
	rb.Push(Line{Text: "error1", Level: LevelError})
	rb.Push(Line{Text: "info2", Level: LevelInfo})
	rb.Push(Line{Text: "warn1", Level: LevelWarn})
	rb.Push(Line{Text: "error2", Level: LevelError})

	errors := rb.LastByLevel(10, LevelError)
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}
	if errors[0].Text != "error1" || errors[1].Text != "error2" {
		t.Errorf("unexpected error lines: %v", errors)
	}

	warnings := rb.LastByLevel(10, LevelWarn)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
}

func TestRingBuffer_Clear(t *testing.T) {
	rb := NewRingBuffer(5)

	rb.Push(Line{Text: "line1"})
	rb.Push(Line{Text: "line2"})

	rb.Clear()

	if rb.Count() != 0 {
		t.Errorf("expected count 0 after clear, got %d", rb.Count())
	}

	lines := rb.Last(5)
	if len(lines) != 0 {
		t.Errorf("expected no lines after clear, got %d", len(lines))
	}
}

func TestRingBuffer_Concurrency(t *testing.T) {
	rb := NewRingBuffer(100)

	// Start multiple goroutines pushing lines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				rb.Push(Line{
					Text:      "line",
					Timestamp: time.Now(),
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and count should be at capacity
	if rb.Count() != 100 {
		t.Errorf("expected count 100, got %d", rb.Count())
	}
}
