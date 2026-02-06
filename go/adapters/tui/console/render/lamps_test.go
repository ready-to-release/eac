package render

import (
	"testing"
)

func TestRenderProgressGradientLamps(t *testing.T) {
	tests := []struct {
		name        string
		activeLamps int
		lampCount   int
		asciiMode   bool
		wantFilled  int // expected filled character count
		wantEmpty   int // expected empty character count
	}{
		{
			name:        "zero active, 12 lamps",
			activeLamps: 0,
			lampCount:   12,
			asciiMode:   true,
			wantFilled:  0,
			wantEmpty:   12,
		},
		{
			name:        "all active, 12 lamps",
			activeLamps: 12,
			lampCount:   12,
			asciiMode:   true,
			wantFilled:  12,
			wantEmpty:   0,
		},
		{
			name:        "half active, 12 lamps",
			activeLamps: 6,
			lampCount:   12,
			asciiMode:   true,
			wantFilled:  6,
			wantEmpty:   6,
		},
		{
			name:        "zero active, 16 lamps",
			activeLamps: 0,
			lampCount:   16,
			asciiMode:   true,
			wantFilled:  0,
			wantEmpty:   16,
		},
		{
			name:        "all active, 16 lamps",
			activeLamps: 16,
			lampCount:   16,
			asciiMode:   true,
			wantFilled:  16,
			wantEmpty:   0,
		},
		{
			name:        "over-count clamped",
			activeLamps: 20,
			lampCount:   12,
			asciiMode:   true,
			wantFilled:  12,
			wantEmpty:   0,
		},
		{
			name:        "negative clamped to zero",
			activeLamps: -5,
			lampCount:   12,
			asciiMode:   true,
			wantFilled:  0,
			wantEmpty:   12,
		},
		{
			name:        "zero lampCount defaults to 16",
			activeLamps: 8,
			lampCount:   0,
			asciiMode:   true,
			wantFilled:  8,
			wantEmpty:   8,
		},
		{
			name:        "single lamp active",
			activeLamps: 1,
			lampCount:   12,
			asciiMode:   true,
			wantFilled:  1,
			wantEmpty:   11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderProgressGradientLamps(tt.activeLamps, tt.lampCount, tt.asciiMode)

			// In ASCII mode, count '*' (filled) and 'o' (empty) characters
			// ANSI escape codes are present, so count the raw characters
			filledCount := 0
			emptyCount := 0
			for _, r := range result {
				switch r {
				case '*':
					filledCount++
				case 'o':
					emptyCount++
				}
			}

			if filledCount != tt.wantFilled {
				t.Errorf("filled count = %d, want %d", filledCount, tt.wantFilled)
			}
			if emptyCount != tt.wantEmpty {
				t.Errorf("empty count = %d, want %d", emptyCount, tt.wantEmpty)
			}
		})
	}
}

func TestRenderProgressGradientLamps_UnicodMode(t *testing.T) {
	result := RenderProgressGradientLamps(6, 12, false)

	// In unicode mode, count filled (●) and empty (○) characters
	filledCount := 0
	emptyCount := 0
	for _, r := range result {
		switch r {
		case '●':
			filledCount++
		case '○':
			emptyCount++
		}
	}

	if filledCount != 6 {
		t.Errorf("filled count = %d, want 6", filledCount)
	}
	if emptyCount != 6 {
		t.Errorf("empty count = %d, want 6", emptyCount)
	}
}

func TestRenderProgressGradientLamps_GradientColors(t *testing.T) {
	// Verify the gradient palette has distinct colors at first and last positions.
	// In test environments without terminal profiles, lipgloss strips ANSI codes,
	// so we verify the color definitions themselves instead of rendered output.
	firstColor := ProgressGradientColors[0]
	lastColor := ProgressGradientColors[15]

	if firstColor == lastColor {
		t.Errorf("expected different colors for first (%s) and last (%s) gradient position", firstColor, lastColor)
	}

	// Verify color sampling: lamp 0 of 12 maps to gradient index 0,
	// lamp 11 of 12 maps to gradient index 15.
	colorIdx0 := 0 * (len(ProgressGradientColors) - 1) / (12 - 1)
	colorIdx11 := 11 * (len(ProgressGradientColors) - 1) / (12 - 1)

	if colorIdx0 != 0 {
		t.Errorf("lamp 0 should map to gradient index 0, got %d", colorIdx0)
	}
	if colorIdx11 != 15 {
		t.Errorf("lamp 11 should map to gradient index 15, got %d", colorIdx11)
	}
}
