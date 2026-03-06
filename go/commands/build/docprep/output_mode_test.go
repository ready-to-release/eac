package docprep

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSiteMode(t *testing.T) {
	m := SiteMode{}
	assert.Equal(t, "site", m.Name())
	assert.False(t, m.IsPDF())
	assert.Equal(t, "", m.Theme())
	assert.False(t, m.ShouldStripMacros())
	assert.True(t, m.ShouldInjectMacros())
	assert.False(t, m.ShouldConvertDrawioToLinks())
	assert.False(t, m.ShouldStripExternalLinks())
	assert.Equal(t, "../", m.ExtraPathPrefix())
	assert.False(t, m.ShouldSkipDrawioFiles())
	assert.False(t, m.ShouldConvertAttrListImages())
}

func TestPDFModeDark(t *testing.T) {
	m := PDFMode{ThemeName: "dark"}
	assert.Equal(t, "pdf-dark", m.Name())
	assert.True(t, m.IsPDF())
	assert.Equal(t, "dark", m.Theme())
	assert.True(t, m.ShouldStripMacros())
	assert.False(t, m.ShouldInjectMacros())
	assert.True(t, m.ShouldConvertDrawioToLinks())
	assert.True(t, m.ShouldStripExternalLinks())
	assert.Equal(t, "", m.ExtraPathPrefix())
	assert.True(t, m.ShouldSkipDrawioFiles())
	assert.True(t, m.ShouldConvertAttrListImages())
}

func TestPDFModeLight(t *testing.T) {
	m := PDFMode{ThemeName: "light"}
	assert.Equal(t, "pdf-light", m.Name())
	assert.True(t, m.IsPDF())
	assert.Equal(t, "light", m.Theme())
}

func TestModeFromString(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantPDF  bool
	}{
		{"site", "site", false},
		{"pdf-dark", "pdf-dark", true},
		{"pdf-light", "pdf-light", true},
		{"pdf-all", "pdf-dark", true},
		{"unknown", "site", false},
		{"", "site", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m := ModeFromString(tt.input)
			assert.Equal(t, tt.wantName, m.Name())
			assert.Equal(t, tt.wantPDF, m.IsPDF())
		})
	}
}
