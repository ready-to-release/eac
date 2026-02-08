package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertAttrListImages(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		adjustPaths bool
		wantResult  string
		wantCount   int
	}{
		{
			name:       "image with width attr",
			content:    `![Alt text](image.png){ width=50% }`,
			wantResult: `<img src="image.png" width="50%" alt="Alt text">`,
			wantCount:  1,
		},
		{
			name:       "image with multiple attrs",
			content:    `![Logo](logo.png){ width=200 height=100 }`,
			wantResult: `<img src="logo.png" width="200" height="100" alt="Logo">`,
			wantCount:  1,
		},
		{
			name:        "adjust paths prepends ../",
			content:     `![Alt](image.png){ width=50% }`,
			adjustPaths: true,
			wantResult:  `<img src="../image.png" width="50%" alt="Alt">`,
			wantCount:   1,
		},
		{
			name:        "absolute path not adjusted",
			content:     `![Alt](/assets/image.png){ width=50% }`,
			adjustPaths: true,
			wantResult:  `<img src="/assets/image.png" width="50%" alt="Alt">`,
			wantCount:   1,
		},
		{
			name:       "drawio files skipped",
			content:    `![Diagram](flow.drawio){ width=100% }`,
			wantResult: `![Diagram](flow.drawio){ width=100% }`,
			wantCount:  0,
		},
		{
			name:       "no recognized attrs returns unchanged",
			content:    `![Alt](image.png){ .some-class }`,
			wantResult: `![Alt](image.png){ .some-class }`,
			wantCount:  0,
		},
		{
			name:       "no attr_list images",
			content:    `![Alt](image.png)`,
			wantResult: `![Alt](image.png)`,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := ConvertAttrListImages(tt.content, tt.adjustPaths)
			assert.Equal(t, tt.wantResult, result)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestParseAttrListToHTML(t *testing.T) {
	tests := []struct {
		name  string
		attrs string
		want  string
	}{
		{
			name:  "width only",
			attrs: `width=50%`,
			want:  `width="50%"`,
		},
		{
			name:  "width and height",
			attrs: `width=200 height=100`,
			want:  `width="200" height="100"`,
		},
		{
			name:  "quoted values",
			attrs: `width="300px"`,
			want:  `width="300px"`,
		},
		{
			name:  "unknown attr filtered",
			attrs: `data-foo=bar`,
			want:  "",
		},
		{
			name:  "style attr",
			attrs: `style=border:1px`,
			want:  `style="border:1px"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAttrListToHTML(tt.attrs)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestAddImageWidthConstraints(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantResult string
		wantCount  int
	}{
		{
			name:       "adds width to plain image",
			content:    `![Photo](photo.png)`,
			wantResult: `![Photo](photo.png){ width="100%" }`,
			wantCount:  1,
		},
		{
			name:       "skips image with existing width",
			content:    `![Photo](photo.png){ width="50%" }`,
			wantResult: `![Photo](photo.png){ width="50%" }`,
			wantCount:  0,
		},
		{
			name:       "skips image with style",
			content:    `![Photo](photo.png){ style="border: 1px" }`,
			wantResult: `![Photo](photo.png){ style="border: 1px" }`,
			wantCount:  0,
		},
		{
			name:       "skips drawio",
			content:    `![Diagram](flow.drawio)`,
			wantResult: `![Diagram](flow.drawio)`,
			wantCount:  0,
		},
		{
			name:       "skips icon paths",
			content:    `![Icon](assets/icon-check.png)`,
			wantResult: `![Icon](assets/icon-check.png)`,
			wantCount:  0,
		},
		{
			name:       "skips badge paths",
			content:    `![Badge](assets/badge.png)`,
			wantResult: `![Badge](assets/badge.png)`,
			wantCount:  0,
		},
		{
			name:       "multiple images",
			content:    "![A](a.png)\n![B](b.png)",
			wantResult: "![A](a.png){ width=\"100%\" }\n![B](b.png){ width=\"100%\" }",
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := AddImageWidthConstraints(tt.content)
			assert.Equal(t, tt.wantResult, result)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestConvertDrawioImages(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		relFileDir string
		siteURL    string
		wantResult string
		wantCount  int
	}{
		{
			name:       "basic drawio conversion",
			content:    `![Diagram](flow.drawio)`,
			relFileDir: "docs/guide",
			siteURL:    "https://example.com/",
			wantResult: `[View interactive diagram](https://example.com/docs/guide/flow.drawio)`,
			wantCount:  1,
		},
		{
			name:       "absolute path drawio",
			content:    `![Diagram](/assets/flow.drawio)`,
			relFileDir: "docs/guide",
			siteURL:    "https://example.com/",
			wantResult: `[View interactive diagram](https://example.com/assets/flow.drawio)`,
			wantCount:  1,
		},
		{
			name:       "no drawio images",
			content:    `![Photo](photo.png)`,
			relFileDir: "docs",
			siteURL:    "https://example.com/",
			wantResult: `![Photo](photo.png)`,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := ConvertDrawioImages(tt.content, tt.relFileDir, tt.siteURL)
			assert.Equal(t, tt.wantResult, result)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}
