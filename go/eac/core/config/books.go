// Package config provides books configuration types and loading
package config

// BooksFileName is the name of the books configuration file
const BooksFileName = "books.yml"

// BooksConfig holds the books configuration
type BooksConfig struct {
	Books []Book `yaml:"books"`
}

// Book represents a single book configuration
type Book struct {
	Name         string         `yaml:"name"`
	Module       string         `yaml:"module"`
	Description  string         `yaml:"description"`
	Output       string         `yaml:"output,omitempty"`   // Default output mode: "site", "pdf-dark", "pdf-light", "pdf-all" (default: "pdf-dark" for book modules)
	SiteURL      string         `yaml:"site_url,omitempty"` // Base URL for GitHub Pages (e.g., https://ready-to-release.github.io/eac/)
	Default      *bool          `yaml:"default,omitempty"`  // Include in default builds (nil/true = yes, false = only with --all)
	Sources      []Source       `yaml:"sources"`
	GeneratedNav []GeneratedNav `yaml:"generated_nav"`
}

// IsDefault returns true if the book should be included in default builds.
// Books are included by default unless explicitly set to false.
func (b *Book) IsDefault() bool {
	if b.Default == nil {
		return true
	}
	return *b.Default
}

// OutputMode constants for book default output
const (
	OutputSite     = "site"
	OutputPDFDark  = "pdf-dark"
	OutputPDFLight = "pdf-light"
	OutputPDFAll   = "pdf-all"
)

// GetOutput returns the book's output mode, defaulting to "pdf-dark" if not specified
func (b *Book) GetOutput() string {
	if b.Output == "" {
		return OutputPDFDark
	}
	return b.Output
}

// IsPDFOutput returns true if the book's output mode is a PDF format
func (b *Book) IsPDFOutput() bool {
	output := b.GetOutput()
	return output == OutputPDFDark || output == OutputPDFLight || output == OutputPDFAll
}

// GetPDFTheme extracts the theme from a PDF output mode, or empty string for site mode
func (b *Book) GetPDFTheme() string {
	switch b.GetOutput() {
	case OutputPDFDark:
		return "dark"
	case OutputPDFLight:
		return "light"
	case OutputPDFAll:
		return "all"
	default:
		return ""
	}
}

// Source represents a content source (copy, command, or inline)
type Source struct {
	// Common field
	Type string `yaml:"type"` // "copy", "command", or "inline"

	// Copy source fields
	From    string   `yaml:"from,omitempty"`
	To      string   `yaml:"to,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`

	// Command source fields
	Command     string         `yaml:"command,omitempty"`
	Target      string         `yaml:"target,omitempty"`
	Frontmatter map[string]any `yaml:"frontmatter,omitempty"`
	Order       int            `yaml:"order,omitempty"`

	// Inline source fields
	MarkerPattern string         `yaml:"marker_pattern,omitempty"`
	Inserts       []InlineInsert `yaml:"inserts,omitempty"`
}

// InlineInsert maps a marker ID to an EAC command
type InlineInsert struct {
	Marker  string `yaml:"marker"`
	Command string `yaml:"command"`
}

// GeneratedNav configures navigation for a generated section
type GeneratedNav struct {
	Section    string `yaml:"section"`
	Title      string `yaml:"title"`
	InsertInto string `yaml:"insert_into"`
	Position   string `yaml:"position"`
}

// GetCopySources returns only copy-type sources
func (b *Book) GetCopySources() []Source {
	var sources []Source
	for _, s := range b.Sources {
		if s.Type == "copy" {
			sources = append(sources, s)
		}
	}
	return sources
}

// GetCommandSources returns only command-type sources, sorted by order
func (b *Book) GetCommandSources() []Source {
	var sources []Source
	for _, s := range b.Sources {
		if s.Type == "command" {
			sources = append(sources, s)
		}
	}
	return sources
}

// GetInlineSources returns only inline-type sources
func (b *Book) GetInlineSources() []Source {
	var sources []Source
	for _, s := range b.Sources {
		if s.Type == "inline" {
			sources = append(sources, s)
		}
	}
	return sources
}

// GetBookByName finds a book by its name
func (c *BooksConfig) GetBookByName(name string) *Book {
	for i := range c.Books {
		if c.Books[i].Name == name {
			return &c.Books[i]
		}
	}
	return nil
}

// GetBooksByModule returns all books that belong to a module
func (c *BooksConfig) GetBooksByModule(moniker string) []*Book {
	var books []*Book
	for i := range c.Books {
		if c.Books[i].Module == moniker {
			books = append(books, &c.Books[i])
		}
	}
	return books
}

// GetDefaultBooksByModule returns only default books for a module.
// Books with default: false are excluded unless --all flag is used.
func (c *BooksConfig) GetDefaultBooksByModule(moniker string) []*Book {
	var books []*Book
	for i := range c.Books {
		if c.Books[i].Module == moniker && c.Books[i].IsDefault() {
			books = append(books, &c.Books[i])
		}
	}
	return books
}

// SourceCount returns the total number of sources
func (b *Book) SourceCount() int {
	return len(b.Sources)
}

// IsCopy returns true if this source is a copy type
func (s *Source) IsCopy() bool {
	return s.Type == "copy"
}

// IsCommand returns true if this source is a command type
func (s *Source) IsCommand() bool {
	return s.Type == "command"
}

// IsInline returns true if this source is an inline type
func (s *Source) IsInline() bool {
	return s.Type == "inline"
}
