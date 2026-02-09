//go:build L0 && ov

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookTemplate_SnippetExpansion(t *testing.T) {
	raw := &BooksConfigRaw{
		Snippets: map[string]SourceSnippet{
			"common": {{Type: "copy", From: "docs/assets/**", To: "assets/"}},
		},
		Books: []Book{{
			Name:    "test",
			Sources: []Source{{Type: "snippet", SnippetName: "common"}},
		}},
	}

	expanded := raw.ExpandBooks(nil)

	require.Len(t, expanded.Books, 1)
	require.Len(t, expanded.Books[0].Sources, 1)
	assert.Equal(t, "copy", expanded.Books[0].Sources[0].Type)
	assert.Equal(t, "docs/assets/**", expanded.Books[0].Sources[0].From)
}

func TestBookTemplate_ParameterSubstitution(t *testing.T) {
	params := map[string]string{"moniker": "eac-ext"}
	src := Source{
		Type:    "command",
		Command: "show release-notes {moniker} latest",
	}

	result := substituteSourceParams(src, params)

	assert.Equal(t, "show release-notes eac-ext latest", result.Command)
}

func TestBookTemplate_TemplateExpansion(t *testing.T) {
	raw := &BooksConfigRaw{
		Templates: map[string]BookTemplate{
			"my-template": {
				Title:       "{title}",
				Description: "{description}",
				Output:      "pdf-dark",
				Sources: []Source{
					{Type: "copy", From: "docs/{category}/**", To: "./"},
				},
			},
		},
		Books: []Book{{
			Name:     "test-book",
			Template: "my-template",
			Parameters: map[string]string{
				"title":       "Test Title",
				"description": "Test Description",
				"category":    "tutorials",
			},
		}},
	}

	expanded := raw.ExpandBooks(nil)

	require.Len(t, expanded.Books, 1)
	book := expanded.Books[0]
	assert.Equal(t, "Test Title", book.Title)
	assert.Equal(t, "Test Description", book.Description)
	assert.Equal(t, "pdf-dark", book.Output)
	require.Len(t, book.Sources, 1)
	assert.Equal(t, "docs/tutorials/**", book.Sources[0].From)
}

func TestBookTemplate_NestedSnippets(t *testing.T) {
	raw := &BooksConfigRaw{
		Snippets: map[string]SourceSnippet{
			"inner": {{Type: "copy", From: "docs/assets/**", To: "assets/"}},
			"outer": {
				{Type: "copy", From: "docs/**/*.md", To: "./"},
				{Type: "snippet", SnippetName: "inner"},
			},
		},
		Books: []Book{{
			Name:    "test",
			Sources: []Source{{Type: "snippet", SnippetName: "outer"}},
		}},
	}

	expanded := raw.ExpandBooks(nil)

	require.Len(t, expanded.Books, 1)
	require.Len(t, expanded.Books[0].Sources, 2)
	assert.Equal(t, "docs/**/*.md", expanded.Books[0].Sources[0].From)
	assert.Equal(t, "docs/assets/**", expanded.Books[0].Sources[1].From)
}

func TestBookTemplate_GeneratorFromCategories(t *testing.T) {
	raw := &BooksConfigRaw{
		Templates: map[string]BookTemplate{
			"diataxis-pdf": {
				Output: "pdf-dark",
				Sources: []Source{
					{Type: "copy", From: "docs/{category}/**", To: "./"},
				},
			},
		},
		Generators: map[string]BookGenerator{
			"diataxis": {
				Template: "diataxis-pdf",
				SiteURL:  "https://example.com",
				Categories: []GeneratorCategory{
					{Name: "tutorials", Title: "Tutorials", Description: "Tutorial docs"},
					{Name: "reference/eac", Path: "reference-eac", Title: "EAC Ref", Description: "EAC reference"},
				},
			},
		},
	}

	expanded := raw.ExpandBooks(nil)

	require.Len(t, expanded.Books, 2)

	// First book: tutorials
	assert.Equal(t, "tutorials", expanded.Books[0].Name)
	assert.Equal(t, "Tutorials", expanded.Books[0].Title)
	require.Len(t, expanded.Books[0].Sources, 1)
	assert.Equal(t, "docs/tutorials/**", expanded.Books[0].Sources[0].From)

	// Second book: reference-eac (with custom path)
	assert.Equal(t, "reference-eac", expanded.Books[1].Name)
	assert.Equal(t, "EAC Ref", expanded.Books[1].Title)
	require.Len(t, expanded.Books[1].Sources, 1)
	assert.Equal(t, "docs/reference/eac/**", expanded.Books[1].Sources[0].From)
}

func TestBookTemplate_GeneratorFromModules(t *testing.T) {
	modules := []Module{
		{Moniker: "eac-ext", EvidenceBooks: []string{"release-evidence-eac-ext"}},
		{Moniker: "clie", EvidenceBooks: []string{"release-evidence-clie"}},
	}

	raw := &BooksConfigRaw{
		Templates: map[string]BookTemplate{
			"release-evidence": {
				Output: "pdf-dark",
				Sources: []Source{
					{Type: "command", Command: "show release-notes {moniker} latest", Target: "release-notes.md"},
				},
			},
		},
		Generators: map[string]BookGenerator{
			"evidence": {
				Template:    "release-evidence",
				FromModules: true,
				SiteURL:     "https://example.com",
			},
		},
	}

	expanded := raw.ExpandBooks(modules)

	require.Len(t, expanded.Books, 2)

	// First evidence book
	assert.Equal(t, "release-evidence-eac-ext", expanded.Books[0].Name)
	require.Len(t, expanded.Books[0].Sources, 1)
	assert.Equal(t, "show release-notes eac-ext latest", expanded.Books[0].Sources[0].Command)

	// Second evidence book
	assert.Equal(t, "release-evidence-clie", expanded.Books[1].Name)
	require.Len(t, expanded.Books[1].Sources, 1)
	assert.Equal(t, "show release-notes clie latest", expanded.Books[1].Sources[0].Command)
}

func TestBookTemplate_MergeBooksConfigs(t *testing.T) {
	defaults := &BooksConfigRaw{
		Templates: map[string]BookTemplate{
			"default-template": {Output: "pdf-dark"},
		},
		Snippets: map[string]SourceSnippet{
			"default-snippet": {{Type: "copy", From: "docs/**", To: "./"}},
		},
		Books: []Book{
			{Name: "default-book", Title: "Default Book"},
		},
	}

	user := &BooksConfigRaw{
		Templates: map[string]BookTemplate{
			"user-template": {Output: "site"},
		},
		Books: []Book{
			{Name: "default-book", Title: "Overridden Title"}, // Override default
			{Name: "user-book", Title: "User Book"},           // Add new
		},
	}

	merged := MergeBooksConfigs(defaults, user, nil)

	require.Len(t, merged.Books, 2)
	// Default book should be overridden
	assert.Equal(t, "Overridden Title", merged.Books[0].Title)
	// User book should be added
	assert.Equal(t, "User Book", merged.Books[1].Title)
}

func TestBookTemplate_FrontmatterSubstitution(t *testing.T) {
	params := map[string]string{"moniker": "eac-ext", "prefix_title": "EAC "}
	src := Source{
		Type:    "command",
		Command: "show release-notes {moniker} latest",
		Frontmatter: map[string]any{
			"title":       "{prefix_title}Release Notes",
			"description": "Release notes for {moniker}",
			"order":       1, // Non-string value should pass through
		},
	}

	result := substituteSourceParams(src, params)

	assert.Equal(t, "EAC Release Notes", result.Frontmatter["title"])
	assert.Equal(t, "Release notes for eac-ext", result.Frontmatter["description"])
	assert.Equal(t, 1, result.Frontmatter["order"])
}

func TestBookTemplate_SnippetWithParameters(t *testing.T) {
	raw := &BooksConfigRaw{
		Snippets: map[string]SourceSnippet{
			"evidence": {
				{Type: "command", Command: "show release-notes {moniker} latest", Target: "{prefix}release-notes.md"},
			},
		},
		Books: []Book{{
			Name: "test",
			Sources: []Source{{
				Type:        "snippet",
				SnippetName: "evidence",
				Parameters: map[string]string{
					"moniker": "eac-ext",
					"prefix":  "eac-",
				},
			}},
		}},
	}

	expanded := raw.ExpandBooks(nil)

	require.Len(t, expanded.Books, 1)
	require.Len(t, expanded.Books[0].Sources, 1)
	assert.Equal(t, "show release-notes eac-ext latest", expanded.Books[0].Sources[0].Command)
	assert.Equal(t, "eac-release-notes.md", expanded.Books[0].Sources[0].Target)
}

func TestLoadBooksConfigRaw(t *testing.T) {
	yaml := `
book-templates:
  my-template:
    output: pdf-dark
source-snippets:
  my-snippet:
    - type: copy
      from: "docs/**"
      to: "./"
books:
  - name: my-book
    title: My Book
`
	raw, err := LoadBooksConfigRaw([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, raw)

	assert.Contains(t, raw.Templates, "my-template")
	assert.Contains(t, raw.Snippets, "my-snippet")
	require.Len(t, raw.Books, 1)
	assert.Equal(t, "my-book", raw.Books[0].Name)
}
