package cells

import "testing"

func TestCommandCell_Render(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		modules  []string
		width    int
		expected string
	}{
		{
			name:     "short command fits",
			command:  "build",
			modules:  []string{"a", "b", "c"},
			width:    50,
			expected: "build a b c",
		},
		{
			name:     "empty modules",
			command:  "build",
			modules:  []string{},
			width:    50,
			expected: "build",
		},
		{
			name:     "single module",
			command:  "test",
			modules:  []string{"core"},
			width:    30,
			expected: "test core",
		},
		{
			name:     "truncate long list",
			command:  "build",
			modules:  []string{"contracts", "core", "eac-cli", "docs", "ai-adapter", "tui-adapter"},
			width:    40,
			// 40 - 4 = 36 max before truncation
			// "build contracts core" = 24 <= 36
			// "build contracts core eac-cli" = 38 > 36, truncate
			expected: "build contracts core ...",
		},
		{
			name:     "truncate very narrow",
			command:  "build",
			modules:  []string{"contracts", "core"},
			width:    20,
			// 20 - 4 = 16 max before truncation
			// "build contracts" = 15 <= 16
			// "build contracts core" = 24 > 16, truncate
			expected: "build contracts ...",
		},
		{
			name:     "exact fit no truncation",
			command:  "build",
			modules:  []string{"mod1", "mod2"},
			width:    20, // "build mod1 mod2" = 15 chars, 20-4=16, 15 <= 16
			expected: "build mod1 mod2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCommandCell()
			c.SetCommand(tt.command)
			c.SetModules(tt.modules)
			got := stripANSI(c.Render(tt.width, 1))
			if got != tt.expected {
				t.Errorf("CommandCell.Render() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCommandCell_TruncationBoundary(t *testing.T) {
	// Test 1: All modules fit
	c := NewCommandCell()
	c.SetCommand("build")
	c.SetModules([]string{"mod1", "mod2"}) // Only 2 modules

	// "build mod1 mod2" = 15 chars
	// With width 20: 20-4=16 max, 15 <= 16, all modules fit
	got := stripANSI(c.Render(20, 1))
	if got != "build mod1 mod2" {
		t.Errorf("all fit: got %q, want %q", got, "build mod1 mod2")
	}

	// Test 2: Truncation occurs
	c2 := NewCommandCell()
	c2.SetCommand("build")
	c2.SetModules([]string{"mod1", "mod2", "mod3"}) // 3 modules

	// With width 20: 20-4=16 max
	// "build mod1 mod2" = 15 <= 16, fits
	// "build mod1 mod2 mod3" = 21 > 16, truncate
	got = stripANSI(c2.Render(20, 1))
	if got != "build mod1 mod2 ..." {
		t.Errorf("truncate: got %q, want %q", got, "build mod1 mod2 ...")
	}

	// Test 3: Narrow width forces early truncation
	// With width 15: 15-4=11 max
	// "build mod1" = 10 <= 11, ok
	// "build mod1 mod2" = 15 > 11, truncate
	got = stripANSI(c2.Render(15, 1))
	if got != "build mod1 ..." {
		t.Errorf("narrow truncate: got %q, want %q", got, "build mod1 ...")
	}
}

func TestCommandCell_NeverExceedsWidth(t *testing.T) {
	c := NewCommandCell()
	c.SetCommand("build")
	c.SetModules([]string{"verylongmodulename1", "verylongmodulename2", "verylongmodulename3"})

	widths := []int{10, 20, 30, 40, 50, 60, 80, 100}
	for _, w := range widths {
		got := stripANSI(c.Render(w, 1))
		if len(got) > w {
			t.Errorf("width %d: output %q exceeds width (len=%d)", w, got, len(got))
		}
	}
}

func TestCommandCell_ZoneID(t *testing.T) {
	c := NewCommandCell()
	if got := c.ZoneID(); got != "res-command" {
		t.Errorf("CommandCell.ZoneID() = %q, want %q", got, "res-command")
	}
}

func TestCommandCell_ImplementsCell(t *testing.T) {
	var _ Cell = (*CommandCell)(nil)
}
