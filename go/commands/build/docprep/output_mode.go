package docprep

// OutputMode defines output-format-specific behavior.
// Implementations are stateless value objects.
type OutputMode interface {
	// Name returns the mode identifier: "site", "pdf-dark", "pdf-light".
	Name() string

	// IsPDF returns true for all PDF output modes.
	IsPDF() bool

	// Theme returns the theme name for PDF modes, empty for site.
	Theme() string

	// ShouldStripMacros returns true if Jinja2 macros should be removed.
	ShouldStripMacros() bool

	// ShouldInjectMacros returns true if navigation macros should be added.
	ShouldInjectMacros() bool

	// ShouldConvertDrawioToLinks returns true if .drawio refs become links.
	ShouldConvertDrawioToLinks() bool

	// ShouldStripExternalLinks returns true if links to files outside
	// staging should be stripped (text kept, href removed).
	ShouldStripExternalLinks() bool

	// ExtraPathPrefix returns a prefix to add to dynamically calculated
	// relative paths. Site mode returns "../" because MkDocs converts
	// file.md to file/index.html, adding a directory level. PDF returns "".
	ExtraPathPrefix() string

	// ShouldSkipDrawioFiles returns true if .drawio source files should
	// not be copied to staging.
	ShouldSkipDrawioFiles() bool
}

// SiteMode is the OutputMode for HTML site builds.
type SiteMode struct{}

func (SiteMode) Name() string                     { return "site" }
func (SiteMode) IsPDF() bool                      { return false }
func (SiteMode) Theme() string                    { return "" }
func (SiteMode) ShouldStripMacros() bool          { return false }
func (SiteMode) ShouldInjectMacros() bool         { return true }
func (SiteMode) ShouldConvertDrawioToLinks() bool { return false }
func (SiteMode) ShouldStripExternalLinks() bool   { return false }
func (SiteMode) ExtraPathPrefix() string          { return "../" }
func (SiteMode) ShouldSkipDrawioFiles() bool      { return false }

// PDFMode is the OutputMode for PDF builds.
type PDFMode struct {
	ThemeName string // "dark" or "light"
}

func (m PDFMode) Name() string                     { return "pdf-" + m.ThemeName }
func (PDFMode) IsPDF() bool                        { return true }
func (m PDFMode) Theme() string                    { return m.ThemeName }
func (PDFMode) ShouldStripMacros() bool            { return true }
func (PDFMode) ShouldInjectMacros() bool           { return false }
func (PDFMode) ShouldConvertDrawioToLinks() bool   { return true }
func (PDFMode) ShouldStripExternalLinks() bool     { return true }
func (PDFMode) ExtraPathPrefix() string            { return "" }
func (PDFMode) ShouldSkipDrawioFiles() bool        { return true }

// ModeFromString creates the appropriate OutputMode from a format string.
func ModeFromString(format string) OutputMode {
	switch format {
	case "site":
		return SiteMode{}
	case "pdf-dark":
		return PDFMode{ThemeName: "dark"}
	case "pdf-light":
		return PDFMode{ThemeName: "light"}
	case "pdf-all":
		return PDFMode{ThemeName: "dark"}
	default:
		return SiteMode{}
	}
}
