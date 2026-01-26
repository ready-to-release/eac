// Package config provides books configuration types and loading
package config

// BooksFileName is the name of the books configuration file.
const BooksFileName = "books.yml"

// BooksConfig holds the books configuration.
type BooksConfig struct {
	Books []Book `yaml:"books"`
}

// Book represents a single book configuration.
type Book struct {
	Name         string         `yaml:"name"`
	Title        string         `yaml:"title,omitempty"` // Book-specific title for cover page (optional, defaults to Name)
	Description  string         `yaml:"description"`
	Output       string         `yaml:"output,omitempty"`   // Default output mode: "site", "pdf-dark", "pdf-light", "pdf-all" (default: "pdf-dark" for book modules)
	SiteURL      string         `yaml:"site_url,omitempty"` // Base URL for GitHub Pages (e.g., https://ready-to-release.github.io/eac/)
	Sources      []Source       `yaml:"sources"`
	GeneratedNav []GeneratedNav `yaml:"generated_nav"`
}

// OutputMode constants for book default output.
const (
	OutputSite     = "site"
	OutputPDFDark  = "pdf-dark"
	OutputPDFLight = "pdf-light"
	OutputPDFAll   = "pdf-all"
)

// GetOutput returns the book's output mode, defaulting to "pdf-dark" if not specified.
func (b *Book) GetOutput() string {
	if b.Output == "" {
		return OutputPDFDark
	}
	return b.Output
}

// IsPDFOutput returns true if the book's output mode is a PDF format.
func (b *Book) IsPDFOutput() bool {
	output := b.GetOutput()
	return output == OutputPDFDark || output == OutputPDFLight || output == OutputPDFAll
}

// GetPDFTheme extracts the theme from a PDF output mode, or empty string for site mode.
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

// Source represents a content source (copy, command, or inline).
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

// InlineInsert maps a marker ID to an EAC command.
type InlineInsert struct {
	Marker  string `yaml:"marker"`
	Command string `yaml:"command"`
}

// GeneratedNav configures navigation for a generated section.
type GeneratedNav struct {
	Section    string `yaml:"section"`
	Title      string `yaml:"title"`
	InsertInto string `yaml:"insert_into"`
	Position   string `yaml:"position"`
}

// GetCopySources returns only copy-type sources.
func (b *Book) GetCopySources() []Source {
	var sources []Source
	for i := range b.Sources {
		if b.Sources[i].Type == "copy" {
			sources = append(sources, b.Sources[i])
		}
	}
	return sources
}

// GetCommandSources returns only command-type sources, sorted by order.
func (b *Book) GetCommandSources() []Source {
	var sources []Source
	for i := range b.Sources {
		if b.Sources[i].Type == "command" {
			sources = append(sources, b.Sources[i])
		}
	}
	return sources
}

// GetInlineSources returns only inline-type sources.
func (b *Book) GetInlineSources() []Source {
	var sources []Source
	for i := range b.Sources {
		if b.Sources[i].Type == "inline" {
			sources = append(sources, b.Sources[i])
		}
	}
	return sources
}

// GetBookByName finds a book by its name.
func (c *BooksConfig) GetBookByName(name string) *Book {
	for i := range c.Books {
		if c.Books[i].Name == name {
			return &c.Books[i]
		}
	}
	return nil
}

// GetBooksByNames returns books matching the given names.
// Used to resolve a module's books list to actual Book configs.
func (c *BooksConfig) GetBooksByNames(names []string) []*Book {
	var books []*Book
	for _, name := range names {
		if book := c.GetBookByName(name); book != nil {
			books = append(books, book)
		}
	}
	return books
}

// GetDefaultBooksByNames returns only the first book (the default) from the names list.
// The first book in a module's books list is the default; others require --all flag.
func (c *BooksConfig) GetDefaultBooksByNames(names []string) []*Book {
	if len(names) == 0 {
		return nil
	}
	// First book is the default
	if book := c.GetBookByName(names[0]); book != nil {
		return []*Book{book}
	}
	return nil
}

// SourceCount returns the total number of sources.
func (b *Book) SourceCount() int {
	return len(b.Sources)
}

// IsCopy returns true if this source is a copy type.
func (s *Source) IsCopy() bool {
	return s.Type == "copy"
}

// IsCommand returns true if this source is a command type.
func (s *Source) IsCommand() bool {
	return s.Type == "command"
}

// IsInline returns true if this source is an inline type.
func (s *Source) IsInline() bool {
	return s.Type == "inline"
}
