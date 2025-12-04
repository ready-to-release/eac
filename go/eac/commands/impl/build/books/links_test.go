package books

import (
	"testing"
)

func TestAddImageWidthConstraints(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		count    int
	}{
		{
			name:     "simple image without attributes",
			input:    `![CD Model](../assets/cd-model/diagram.png)`,
			expected: `![CD Model](../assets/cd-model/diagram.png){ width="100%" }`,
			count:    1,
		},
		{
			name:     "image with existing width - skip",
			input:    `![Image](path.png){ width="50%" }`,
			expected: `![Image](path.png){ width="50%" }`,
			count:    0,
		},
		{
			name:     "image with style attribute - skip",
			input:    `![Image](path.png){ style="max-width: 300px" }`,
			expected: `![Image](path.png){ style="max-width: 300px" }`,
			count:    0,
		},
		{
			name:     "icon image - skip",
			input:    `![Icon](../icons/check-icon.png)`,
			expected: `![Icon](../icons/check-icon.png)`,
			count:    0,
		},
		{
			name:     "favicon - skip",
			input:    `![](favicon.ico)`,
			expected: `![](favicon.ico)`,
			count:    0,
		},
		{
			name:     "badge image - skip",
			input:    `![Build Badge](badge.svg)`,
			expected: `![Build Badge](badge.svg)`,
			count:    0,
		},
		{
			name:     "multiple images",
			input:    "Some text\n![Img1](path1.png)\nMore text\n![Img2](path2.png)",
			expected: "Some text\n![Img1](path1.png){ width=\"100%\" }\nMore text\n![Img2](path2.png){ width=\"100%\" }",
			count:    2,
		},
		{
			name:     "drawio diagram",
			input:    `![Architecture](../assets/branching/overview.drawio.png)`,
			expected: `![Architecture](../assets/branching/overview.drawio.png){ width="100%" }`,
			count:    1,
		},
		{
			name:     "image with other attributes but no width",
			input:    `![Image](path.png){ loading="lazy" }`,
			expected: `![Image](path.png){ loading="lazy"  width="100%" }`,
			count:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := addImageWidthConstraints(tt.input)
			if result != tt.expected {
				t.Errorf("addImageWidthConstraints() result = %q, want %q", result, tt.expected)
			}
			if count != tt.count {
				t.Errorf("addImageWidthConstraints() count = %d, want %d", count, tt.count)
			}
		})
	}
}
