// Package config provides book template expansion and source snippet support.
package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// BookTemplate defines a reusable book structure with placeholders.
type BookTemplate struct {
	Title       string   `yaml:"title,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Output      string   `yaml:"output,omitempty"`
	SiteURL     string   `yaml:"site_url,omitempty"`
	Sources     []Source `yaml:"sources,omitempty"`
}

// SourceSnippet is a named, reusable list of sources.
type SourceSnippet []Source

// BookGenerator creates multiple books from a template.
type BookGenerator struct {
	Template    string              `yaml:"template"`
	SiteURL     string              `yaml:"site_url,omitempty"`
	Categories  []GeneratorCategory `yaml:"categories,omitempty"`
	FromModules bool                `yaml:"from_modules,omitempty"`
}

// GeneratorCategory defines a category for diataxis generation.
type GeneratorCategory struct {
	Name        string `yaml:"name"`        // Category path (e.g., "reference/eac")
	Path        string `yaml:"path"`        // Output book name (defaults to name with / -> -)
	Title       string `yaml:"title"`       // Book title
	Description string `yaml:"description"` // Book description
}

// BooksConfigRaw is the raw YAML structure before expansion.
type BooksConfigRaw struct {
	Templates  map[string]BookTemplate  `yaml:"book-templates,omitempty"`
	Snippets   map[string]SourceSnippet `yaml:"source-snippets,omitempty"`
	Generators map[string]BookGenerator `yaml:"book-generators,omitempty"`
	Books      []Book                   `yaml:"books,omitempty"`
}

// ExpandBooks converts raw config with templates/generators into final books.
func (raw *BooksConfigRaw) ExpandBooks(modules []Module) *BooksConfig {
	if raw == nil {
		return &BooksConfig{}
	}

	expanded := &BooksConfig{
		Books: make([]Book, 0, len(raw.Books)),
	}

	// Track book names to prevent duplicates
	bookNames := make(map[string]bool)

	// 1. Copy explicit books (expanding snippet refs and templates)
	// Explicit books take precedence over generated books
	for _, book := range raw.Books {
		expandedBook := raw.expandBook(book)
		expanded.Books = append(expanded.Books, expandedBook)
		bookNames[expandedBook.Name] = true
	}

	// 2. Process generators (skip books that already exist as explicit)
	for _, gen := range raw.Generators {
		generatedBooks := raw.generateBooks(gen, modules)
		for _, book := range generatedBooks {
			if !bookNames[book.Name] {
				expanded.Books = append(expanded.Books, book)
				bookNames[book.Name] = true
			}
		}
	}

	return expanded
}

// expandBook expands a single book, resolving template and snippet references.
func (raw *BooksConfigRaw) expandBook(book Book) Book {
	// If book uses a template, start with template as base
	if book.Template != "" {
		tmpl, ok := raw.Templates[book.Template]
		if ok {
			book = mergeBookWithTemplate(book, tmpl)
		}
	}

	// Expand snippet references in sources
	book.Sources = raw.expandSources(book.Sources, book.Parameters)

	// Substitute parameters in all string fields
	book = substituteBookParams(book, book.Parameters)

	return book
}

// mergeBookWithTemplate merges book fields with template, book takes precedence.
func mergeBookWithTemplate(book Book, tmpl BookTemplate) Book {
	result := book

	// Template provides defaults, book overrides
	if result.Title == "" {
		result.Title = tmpl.Title
	}
	if result.Description == "" {
		result.Description = tmpl.Description
	}
	if result.Output == "" {
		result.Output = tmpl.Output
	}
	if result.SiteURL == "" {
		result.SiteURL = tmpl.SiteURL
	}

	// For sources: if book has no sources, use template sources
	// If book has sources, they completely override template
	if len(result.Sources) == 0 {
		result.Sources = make([]Source, len(tmpl.Sources))
		copy(result.Sources, tmpl.Sources)
	}

	return result
}

// expandSources expands snippet references into actual sources.
func (raw *BooksConfigRaw) expandSources(sources []Source, params map[string]string) []Source {
	var expanded []Source
	for _, src := range sources {
		if src.Type == "snippet" {
			// Find snippet and expand it
			snippetName := src.SnippetName
			snippet, ok := raw.Snippets[snippetName]
			if ok {
				// Merge parameters: source params override book params
				mergedParams := mergeMaps(params, src.Parameters)
				// Recursively expand (snippets can reference other snippets)
				snippetSources := raw.expandSources(snippet, mergedParams)
				// Substitute parameters in each source
				for _, ss := range snippetSources {
					expanded = append(expanded, substituteSourceParams(ss, mergedParams))
				}
			}
		} else {
			expanded = append(expanded, src)
		}
	}
	return expanded
}

// generateBooks creates books from a generator definition.
func (raw *BooksConfigRaw) generateBooks(gen BookGenerator, modules []Module) []Book {
	var books []Book
	tmpl, ok := raw.Templates[gen.Template]
	if !ok {
		return books
	}

	if gen.FromModules {
		// Generate evidence books from modules with evidence-book components
		for _, mod := range modules {
			for _, evidenceBookName := range mod.GetEvidenceBooks() {
				params := map[string]string{
					"moniker":  mod.Moniker,
					"site_url": gen.SiteURL,
				}
				book := createBookFromTemplate(evidenceBookName, tmpl, params)
				books = append(books, raw.expandBook(book))
			}
		}
	} else if len(gen.Categories) > 0 {
		// Generate diataxis PDFs from categories
		for _, cat := range gen.Categories {
			bookName := cat.Path
			if bookName == "" {
				bookName = strings.ReplaceAll(cat.Name, "/", "-")
			}
			params := map[string]string{
				"category":    cat.Name,
				"title":       cat.Title,
				"description": cat.Description,
				"site_url":    gen.SiteURL,
			}
			book := createBookFromTemplate(bookName, tmpl, params)
			book.Title = cat.Title
			book.Description = cat.Description
			books = append(books, raw.expandBook(book))
		}
	}

	return books
}

// createBookFromTemplate creates a new book from a template with parameters.
func createBookFromTemplate(name string, tmpl BookTemplate, params map[string]string) Book {
	return Book{
		Name:        name,
		Title:       tmpl.Title,
		Description: tmpl.Description,
		Output:      tmpl.Output,
		SiteURL:     tmpl.SiteURL,
		Sources:     cloneSources(tmpl.Sources),
		Parameters:  params,
	}
}

// cloneSources creates a deep copy of sources.
func cloneSources(sources []Source) []Source {
	if sources == nil {
		return nil
	}
	result := make([]Source, len(sources))
	for i, s := range sources {
		result[i] = s
		// Deep copy slices
		if s.Exclude != nil {
			result[i].Exclude = make([]string, len(s.Exclude))
			copy(result[i].Exclude, s.Exclude)
		}
		if s.Inserts != nil {
			result[i].Inserts = make([]InlineInsert, len(s.Inserts))
			copy(result[i].Inserts, s.Inserts)
		}
		if s.Frontmatter != nil {
			result[i].Frontmatter = make(map[string]any)
			for k, v := range s.Frontmatter {
				result[i].Frontmatter[k] = v
			}
		}
		if s.Parameters != nil {
			result[i].Parameters = make(map[string]string)
			for k, v := range s.Parameters {
				result[i].Parameters[k] = v
			}
		}
	}
	return result
}

// substituteBookParams substitutes {placeholder} patterns in book fields.
func substituteBookParams(book Book, params map[string]string) Book {
	if params == nil {
		return book
	}

	book.Title = substituteParams(book.Title, params)
	book.Description = substituteParams(book.Description, params)
	book.SiteURL = substituteParams(book.SiteURL, params)

	// Substitute in sources
	for i := range book.Sources {
		book.Sources[i] = substituteSourceParams(book.Sources[i], params)
	}

	return book
}

// substituteSourceParams substitutes {placeholder} patterns in source fields.
func substituteSourceParams(src Source, params map[string]string) Source {
	if params == nil {
		return src
	}

	src.From = substituteParams(src.From, params)
	src.To = substituteParams(src.To, params)
	src.Command = substituteParams(src.Command, params)
	src.Target = substituteParams(src.Target, params)

	// Substitute in exclude list
	for i := range src.Exclude {
		src.Exclude[i] = substituteParams(src.Exclude[i], params)
	}

	// Substitute in frontmatter string values
	if src.Frontmatter != nil {
		newFM := make(map[string]any)
		for k, v := range src.Frontmatter {
			if str, ok := v.(string); ok {
				newFM[k] = substituteParams(str, params)
			} else {
				newFM[k] = v
			}
		}
		src.Frontmatter = newFM
	}

	return src
}

// MergeBooksConfigs merges defaults with user config, expanding templates.
func MergeBooksConfigs(defaults, user *BooksConfigRaw, modules []Module) *BooksConfig {
	if defaults == nil && user == nil {
		return &BooksConfig{}
	}

	// Combine raw configs
	combined := &BooksConfigRaw{
		Templates:  make(map[string]BookTemplate),
		Snippets:   make(map[string]SourceSnippet),
		Generators: make(map[string]BookGenerator),
	}

	// Copy defaults first
	if defaults != nil {
		for k, v := range defaults.Templates {
			combined.Templates[k] = v
		}
		for k, v := range defaults.Snippets {
			combined.Snippets[k] = v
		}
		for k, v := range defaults.Generators {
			combined.Generators[k] = v
		}
		combined.Books = append(combined.Books, defaults.Books...)
	}

	// Override/extend with user config
	if user != nil {
		for k, v := range user.Templates {
			combined.Templates[k] = v
		}
		for k, v := range user.Snippets {
			combined.Snippets[k] = v
		}
		for k, v := range user.Generators {
			combined.Generators[k] = v
		}
		// User books override defaults by name
		combined.Books = mergeBookLists(combined.Books, user.Books)
	}

	// Expand everything
	return combined.ExpandBooks(modules)
}

// mergeBookLists merges two book lists, user books override defaults by name.
func mergeBookLists(defaults, user []Book) []Book {
	// Create map of default books by name
	byName := make(map[string]int)
	result := make([]Book, len(defaults))
	copy(result, defaults)
	for i, b := range result {
		byName[b.Name] = i
	}

	// Override or add user books
	for _, b := range user {
		if idx, exists := byName[b.Name]; exists {
			result[idx] = b
		} else {
			result = append(result, b)
			byName[b.Name] = len(result) - 1
		}
	}

	return result
}

// LoadBooksConfigRaw loads books config as raw structure (before expansion).
func LoadBooksConfigRaw(data []byte) (*BooksConfigRaw, error) {
	var raw BooksConfigRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing books config: %w", err)
	}
	return &raw, nil
}

// mergeMaps merges two string maps, with override values taking precedence.
func mergeMaps(base, override map[string]string) map[string]string {
	if base == nil && override == nil {
		return nil
	}
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}
