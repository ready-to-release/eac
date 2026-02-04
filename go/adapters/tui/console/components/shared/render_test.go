package shared

import (
	"strings"
	"testing"
	"time"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{
			name:     "empty string",
			input:    "",
			maxWidth: 10,
			want:     "",
		},
		{
			name:     "string shorter than max",
			input:    "hello",
			maxWidth: 10,
			want:     "hello",
		},
		{
			name:     "string equal to max",
			input:    "hello",
			maxWidth: 5,
			want:     "hello",
		},
		{
			name:     "string longer than max with ellipsis",
			input:    "hello world",
			maxWidth: 8,
			want:     "hello...",
		},
		{
			name:     "very short max width",
			input:    "hello",
			maxWidth: 3,
			want:     "hel",
		},
		{
			name:     "max width 2",
			input:    "hello",
			maxWidth: 2,
			want:     "he",
		},
		{
			name:     "max width 1",
			input:    "hello",
			maxWidth: 1,
			want:     "h",
		},
		{
			name:     "zero max width",
			input:    "hello",
			maxWidth: 0,
			want:     "",
		},
		{
			name:     "negative max width",
			input:    "hello",
			maxWidth: -5,
			want:     "",
		},
		{
			name:     "long string truncation",
			input:    "this is a very long string that needs truncation",
			maxWidth: 20,
			want:     "this is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			width: 5,
			want:  "     ",
		},
		{
			name:  "string shorter than width",
			input: "hi",
			width: 5,
			want:  "hi   ",
		},
		{
			name:  "string equal to width",
			input: "hello",
			width: 5,
			want:  "hello",
		},
		{
			name:  "string longer than width",
			input: "hello world",
			width: 5,
			want:  "hello world",
		},
		{
			name:  "zero width",
			input: "hi",
			width: 0,
			want:  "hi",
		},
		{
			name:  "pad to width 10",
			input: "abc",
			width: 10,
			want:  "abc       ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PadRight(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("PadRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestPadToHeight(t *testing.T) {
	tests := []struct {
		name    string
		content string
		height  int
		want    string
	}{
		{
			name:    "empty content pad to height 3",
			content: "",
			height:  3,
			want:    "\n\n",
		},
		{
			name:    "single line pad to height 3",
			content: "line1",
			height:  3,
			want:    "line1\n\n",
		},
		{
			name:    "two lines pad to height 3",
			content: "line1\nline2",
			height:  3,
			want:    "line1\nline2\n",
		},
		{
			name:    "three lines at height 3",
			content: "line1\nline2\nline3",
			height:  3,
			want:    "line1\nline2\nline3",
		},
		{
			name:    "four lines truncate to height 3",
			content: "line1\nline2\nline3\nline4",
			height:  3,
			want:    "line1\nline2\nline3",
		},
		{
			name:    "height 1 keeps first line",
			content: "line1\nline2\nline3",
			height:  1,
			want:    "line1",
		},
		{
			name:    "height 0 returns empty",
			content: "line1\nline2",
			height:  0,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PadToHeight(tt.content, tt.height)
			if got != tt.want {
				t.Errorf("PadToHeight(%q, %d) = %q, want %q", tt.content, tt.height, got, tt.want)
			}
		})
	}
}

func TestHorizontalJoin(t *testing.T) {
	tests := []struct {
		name   string
		left   string
		right  string
		height int
	}{
		{
			name:   "simple join",
			left:   "L1\nL2",
			right:  "R1\nR2",
			height: 2,
		},
		{
			name:   "different heights",
			left:   "L1",
			right:  "R1\nR2\nR3",
			height: 3,
		},
		{
			name:   "empty left",
			left:   "",
			right:  "R1\nR2",
			height: 2,
		},
		{
			name:   "empty right",
			left:   "L1\nL2",
			right:  "",
			height: 2,
		},
		{
			name:   "both empty",
			left:   "",
			right:  "",
			height: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HorizontalJoin(tt.left, tt.right, tt.height)
			lines := strings.Split(got, "\n")

			// Verify correct number of lines
			if len(lines) != tt.height {
				t.Errorf("HorizontalJoin() produced %d lines, want %d", len(lines), tt.height)
			}

			// Verify each line has content from both sides
			for i, line := range lines {
				if !strings.Contains(line, " ") && tt.right != "" {
					t.Errorf("Line %d doesn't appear to join left and right: %q", i, line)
				}
			}
		})
	}
}

func TestHorizontalJoin_PreservesWidth(t *testing.T) {
	left := "short\nlonger line"
	right := "R1\nR2"
	height := 2

	got := HorizontalJoin(left, right, height)
	lines := strings.Split(got, "\n")

	// All lines should have the same left padding width
	if len(lines) >= 2 {
		// The "short" line should be padded to match "longer line"
		// Find where the right content starts
		idx1 := strings.LastIndex(lines[0], "R")
		idx2 := strings.LastIndex(lines[1], "R")
		if idx1 != idx2 {
			t.Errorf("Left side not uniformly padded: R at %d and %d", idx1, idx2)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "zero duration",
			duration: 0,
			want:     "0ms",
		},
		{
			name:     "1 millisecond",
			duration: time.Millisecond,
			want:     "1ms",
		},
		{
			name:     "100 milliseconds",
			duration: 100 * time.Millisecond,
			want:     "100ms",
		},
		{
			name:     "500 milliseconds",
			duration: 500 * time.Millisecond,
			want:     "500ms",
		},
		{
			name:     "999 milliseconds",
			duration: 999 * time.Millisecond,
			want:     "999ms",
		},
		{
			name:     "exactly 1 second",
			duration: time.Second,
			want:     "1.0s",
		},
		{
			name:     "1.5 seconds",
			duration: 1500 * time.Millisecond,
			want:     "1.5s",
		},
		{
			name:     "30 seconds",
			duration: 30 * time.Second,
			want:     "30.0s",
		},
		{
			name:     "59.9 seconds",
			duration: 59900 * time.Millisecond,
			want:     "59.9s",
		},
		{
			name:     "exactly 1 minute",
			duration: time.Minute,
			want:     "1m00s",
		},
		{
			name:     "1 minute 30 seconds",
			duration: time.Minute + 30*time.Second,
			want:     "1m30s",
		},
		{
			name:     "5 minutes 5 seconds",
			duration: 5*time.Minute + 5*time.Second,
			want:     "5m05s",
		},
		{
			name:     "10 minutes",
			duration: 10 * time.Minute,
			want:     "10m00s",
		},
		{
			name:     "1 hour (60 minutes)",
			duration: time.Hour,
			want:     "60m00s",
		},
		{
			name:     "1 hour 30 minutes",
			duration: time.Hour + 30*time.Minute + 45*time.Second,
			want:     "90m45s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestTreeBranch(t *testing.T) {
	tests := []struct {
		name   string
		isLast bool
		want   string
	}{
		{
			name:   "not last item",
			isLast: false,
			want:   "|-",
		},
		{
			name:   "last item",
			isLast: true,
			want:   "`-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TreeBranch(tt.isLast)
			// The actual characters are Unicode box drawing characters
			// Check that the results are different
			if (tt.isLast && got == TreeBranch(false)) || (!tt.isLast && got == TreeBranch(true)) {
				t.Errorf("TreeBranch(%v) should differ from TreeBranch(%v)", tt.isLast, !tt.isLast)
			}
		})
	}
}

func TestRepeatChar(t *testing.T) {
	tests := []struct {
		name string
		char string
		n    int
		want string
	}{
		{
			name: "repeat dash 5 times",
			char: "-",
			n:    5,
			want: "-----",
		},
		{
			name: "repeat star 3 times",
			char: "*",
			n:    3,
			want: "***",
		},
		{
			name: "repeat zero times",
			char: "x",
			n:    0,
			want: "",
		},
		{
			name: "repeat negative times",
			char: "x",
			n:    -5,
			want: "",
		},
		{
			name: "repeat empty char",
			char: "",
			n:    5,
			want: "",
		},
		{
			name: "repeat multi-char string",
			char: "ab",
			n:    3,
			want: "ababab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RepeatChar(tt.char, tt.n)
			if got != tt.want {
				t.Errorf("RepeatChar(%q, %d) = %q, want %q", tt.char, tt.n, got, tt.want)
			}
		})
	}
}

func TestLabelValue(t *testing.T) {
	tests := []struct {
		name  string
		label string
		value string
	}{
		{
			name:  "simple label value",
			label: "Status",
			value: "Running",
		},
		{
			name:  "empty value",
			label: "Name",
			value: "",
		},
		{
			name:  "empty label",
			label: "",
			value: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LabelValue(tt.label, tt.value)
			// Verify the output contains both the label and value
			if !strings.Contains(got, tt.value) && tt.value != "" {
				t.Errorf("LabelValue() doesn't contain value %q", tt.value)
			}
			// Note: The label may be styled, so we just check that the function returns something
			if got == "" && (tt.label != "" || tt.value != "") {
				t.Errorf("LabelValue() returned empty string for non-empty inputs")
			}
		})
	}
}

