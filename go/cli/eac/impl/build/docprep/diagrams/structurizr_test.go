package diagrams

import (
	"strings"
	"testing"
)

func TestExtractStructurizrMarkers(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		expected  int
		module    string
		component string
		viewKey   string
	}{
		{
			name:     "single 2-part marker",
			content:  "# Title\n\n<!-- structurizr:eac-cli:SystemContext -->\n\nMore text.",
			expected: 1,
			module:   "eac-cli",
			viewKey:  "SystemContext",
		},
		{
			name:      "single 3-part marker",
			content:   "# Title\n\n<!-- structurizr:adapters:tui-eac:SystemContext -->\n\nMore text.",
			expected:  1,
			module:    "adapters",
			component: "tui-eac",
			viewKey:   "SystemContext",
		},
		{
			name:     "multiple 2-part markers",
			content:  "<!-- structurizr:eac-cli:SystemContext -->\n\n<!-- structurizr:clie:Containers -->",
			expected: 2,
		},
		{
			name:     "no markers",
			content:  "# Just a heading\n\nSome regular markdown.",
			expected: 0,
		},
		{
			name:     "2-part marker with spaces",
			content:  "<!--  structurizr:eac-cli:SystemContext  -->",
			expected: 1,
			module:   "eac-cli",
			viewKey:  "SystemContext",
		},
		{
			name:      "3-part marker with spaces",
			content:   "<!--  structurizr:adapters:docker-eac:Containers  -->",
			expected:  1,
			module:    "adapters",
			component: "docker-eac",
			viewKey:   "Containers",
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
				if markers[0].Component != tt.component {
					t.Errorf("Expected component %q, got %q", tt.component, markers[0].Component)
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

	if marker.Component != "" {
		t.Errorf("Expected empty component for 2-part marker, got %q", marker.Component)
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

	// All should have empty Component (2-part markers)
	for _, m := range markers {
		if m.Component != "" {
			t.Errorf("Expected empty component for 2-part marker %s:%s, got %q",
				m.Module, m.ViewKey, m.Component)
		}
	}
}

func TestExtractStructurizrMarkersMixed2And3Part(t *testing.T) {
	content := `# Architecture

<!-- structurizr:eac:SystemContext -->

<!-- structurizr:adapters:tui-eac:SystemContext -->

<!-- structurizr:adapters:docker-eac:Containers -->

<!-- structurizr:eac:Containers -->
`

	markers := ExtractStructurizrMarkers(content)

	if len(markers) != 4 {
		t.Fatalf("Expected 4 markers, got %d", len(markers))
	}

	// First: 2-part eac:SystemContext
	if markers[0].Module != "eac" || markers[0].Component != "" || markers[0].ViewKey != "SystemContext" {
		t.Errorf("Marker 0: got module=%q component=%q view=%q",
			markers[0].Module, markers[0].Component, markers[0].ViewKey)
	}

	// Second: 3-part adapters:tui-eac:SystemContext
	if markers[1].Module != "adapters" || markers[1].Component != "tui-eac" || markers[1].ViewKey != "SystemContext" {
		t.Errorf("Marker 1: got module=%q component=%q view=%q",
			markers[1].Module, markers[1].Component, markers[1].ViewKey)
	}

	// Third: 3-part adapters:docker-eac:Containers
	if markers[2].Module != "adapters" || markers[2].Component != "docker-eac" || markers[2].ViewKey != "Containers" {
		t.Errorf("Marker 2: got module=%q component=%q view=%q",
			markers[2].Module, markers[2].Component, markers[2].ViewKey)
	}

	// Fourth: 2-part eac:Containers
	if markers[3].Module != "eac" || markers[3].Component != "" || markers[3].ViewKey != "Containers" {
		t.Errorf("Marker 3: got module=%q component=%q view=%q",
			markers[3].Module, markers[3].Component, markers[3].ViewKey)
	}
}

func TestExtractStructurizrMarkers3PartComponentPairs(t *testing.T) {
	content := `<!-- structurizr:adapters:docker-eac:SystemContext -->
<!-- structurizr:adapters:docker-eac:Containers -->
<!-- structurizr:adapters:docker-eac:ContainerAdapterComponents -->
<!-- structurizr:adapters:docker-eac:DockerClientComponents -->
<!-- structurizr:adapters:docker-eac:ServeComponents -->`

	markers := ExtractStructurizrMarkers(content)

	if len(markers) != 5 {
		t.Fatalf("Expected 5 markers, got %d", len(markers))
	}

	expectedViews := []string{
		"SystemContext", "Containers", "ContainerAdapterComponents",
		"DockerClientComponents", "ServeComponents",
	}
	for i, m := range markers {
		if m.Module != "adapters" {
			t.Errorf("Marker %d: expected module 'adapters', got %q", i, m.Module)
		}
		if m.Component != "docker-eac" {
			t.Errorf("Marker %d: expected component 'docker-eac', got %q", i, m.Component)
		}
		if m.ViewKey != expectedViews[i] {
			t.Errorf("Marker %d: expected view %q, got %q", i, expectedViews[i], m.ViewKey)
		}
	}
}
