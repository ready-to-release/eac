package diagrams

import (
	"strings"
	"testing"
)

func TestExtractStructurizrMarkers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
		module   string
		viewKey  string
	}{
		{
			name:     "single marker",
			content:  "# Title\n\n<!-- structurizr:eac-cli:SystemContext -->\n\nMore text.",
			expected: 1,
			module:   "eac-cli",
			viewKey:  "SystemContext",
		},
		{
			name:     "multiple markers",
			content:  "<!-- structurizr:eac-cli:SystemContext -->\n\n<!-- structurizr:clie:Containers -->",
			expected: 2,
		},
		{
			name:     "no markers",
			content:  "# Just a heading\n\nSome regular markdown.",
			expected: 0,
		},
		{
			name:     "marker with spaces",
			content:  "<!--  structurizr:eac-cli:SystemContext  -->",
			expected: 1,
			module:   "eac-cli",
			viewKey:  "SystemContext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markers := ExtractStructurizrMarkers(tt.content)

			if len(markers) != tt.expected {
				t.Fatalf("Expected %d markers, got %d", tt.expected, len(markers))
			}

			if tt.expected > 0 && tt.module != "" {
				if markers[0].Module != tt.module {
					t.Errorf("Expected module %q, got %q", tt.module, markers[0].Module)
				}
				if markers[0].ViewKey != tt.viewKey {
					t.Errorf("Expected viewKey %q, got %q", tt.viewKey, markers[0].ViewKey)
				}
			}
		})
	}
}

func TestExtractStructurizrMarkersPositions(t *testing.T) {
	content := "before\n<!-- structurizr:mod:View -->\nafter"
	markers := ExtractStructurizrMarkers(content)

	if len(markers) != 1 {
		t.Fatalf("Expected 1 marker, got %d", len(markers))
	}

	marker := markers[0]

	if marker.StartPos < 0 || marker.EndPos <= marker.StartPos {
		t.Errorf("Invalid positions: start=%d, end=%d", marker.StartPos, marker.EndPos)
	}

	extracted := content[marker.StartPos:marker.EndPos]
	if !strings.Contains(extracted, "structurizr:mod:View") {
		t.Errorf("FullMatch doesn't contain expected text, got: %s", extracted)
	}

	if marker.FullMatch != extracted {
		t.Errorf("FullMatch %q doesn't match extracted %q", marker.FullMatch, extracted)
	}
}

func TestExtractStructurizrMarkersMultipleModules(t *testing.T) {
	content := `# Architecture

<!-- structurizr:eac-cli:SystemContext -->

Some text.

<!-- structurizr:eac-cli:Containers -->

More text.

<!-- structurizr:clie:SystemContext -->
`

	markers := ExtractStructurizrMarkers(content)

	if len(markers) != 3 {
		t.Fatalf("Expected 3 markers, got %d", len(markers))
	}

	modules := make(map[string]int)
	for _, m := range markers {
		modules[m.Module]++
	}

	if modules["eac-cli"] != 2 {
		t.Errorf("Expected 2 eac-cli markers, got %d", modules["eac-cli"])
	}
	if modules["clie"] != 1 {
		t.Errorf("Expected 1 clie marker, got %d", modules["clie"])
	}
}
