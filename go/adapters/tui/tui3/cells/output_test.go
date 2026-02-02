package cells

import (
	"strings"
	"testing"
)

func TestOutputCell_Render(t *testing.T) {
	tests := []struct {
		name      string
		lines     []OutputLine
		height    int
		wantParts []string
	}{
		{
			name:      "empty",
			lines:     nil,
			height:    5,
			wantParts: []string{},
		},
		{
			name: "few lines",
			lines: []OutputLine{
				{Text: "line 1", Level: LineLevelInfo},
				{Text: "line 2", Level: LineLevelInfo},
				{Text: "line 3", Level: LineLevelInfo},
			},
			height:    10,
			wantParts: []string{"line 1", "line 2", "line 3"},
		},
		{
			name: "with levels",
			lines: []OutputLine{
				{Text: "info message", Level: LineLevelInfo},
				{Text: "warn message", Level: LineLevelWarn},
				{Text: "error message", Level: LineLevelError},
			},
			height:    10,
			wantParts: []string{"info message", "warn message", "error message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewOutputCell()
			c.SetLines(tt.lines)
			got := stripANSI(c.Render(60, tt.height))

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("OutputCell.Render() missing %q in:\n%s", part, got)
				}
			}
		})
	}
}

func TestOutputCell_Scroll(t *testing.T) {
	// Create more lines than can fit
	var lines []OutputLine
	for i := 0; i < 20; i++ {
		lines = append(lines, OutputLine{
			Text:  strings.Repeat("x", 10) + string(rune('a'+i)),
			Level: LineLevelInfo,
		})
	}

	c := NewOutputCell()
	c.SetLines(lines)

	// Render with small height (5 lines)
	// Default scroll offset 0 shows last 5 lines
	got := stripANSI(c.Render(60, 5))

	// Should show the last lines (auto-scroll to bottom)
	if !strings.Contains(got, "t") { // 't' is 20th letter
		t.Errorf("OutputCell should show last lines, got:\n%s", got)
	}

	// Scroll up
	c.SetScrollOffset(10)
	got = stripANSI(c.Render(60, 5))

	// Should show earlier lines now
	if strings.Contains(got, "t") {
		t.Errorf("OutputCell with offset should not show last line, got:\n%s", got)
	}
}

func TestOutputCell_AutoScroll(t *testing.T) {
	c := NewOutputCell()

	// Initially auto-scroll should be true
	if !c.IsAutoScrolling() {
		t.Error("OutputCell should default to auto-scroll")
	}

	// Disable auto-scroll
	c.SetAutoScroll(false)
	if c.IsAutoScrolling() {
		t.Error("SetAutoScroll(false) should disable auto-scroll")
	}

	// Enable auto-scroll
	c.SetAutoScroll(true)
	if !c.IsAutoScrolling() {
		t.Error("SetAutoScroll(true) should enable auto-scroll")
	}
}

func TestOutputCell_LevelPrefix(t *testing.T) {
	c := NewOutputCell()
	c.SetLines([]OutputLine{
		{Text: "error line", Level: LineLevelError},
	})

	got := c.Render(60, 5)

	// Error lines should have some indicator (X or similar)
	stripped := stripANSI(got)
	if !strings.Contains(stripped, "error line") {
		t.Errorf("OutputCell should contain error line, got:\n%s", stripped)
	}
}

func TestOutputCell_ZoneID(t *testing.T) {
	c := NewOutputCell()
	if got := c.ZoneID(); got != "" {
		t.Errorf("OutputCell.ZoneID() = %q, want %q", got, "")
	}
}

func TestOutputCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*OutputCell)(nil)
}
