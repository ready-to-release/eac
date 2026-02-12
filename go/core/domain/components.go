package domain

// ComponentType defines component type metadata for domain validation.
// This is loaded from component-kinds in blueprints and used by the lint orchestrator.
// See also config.ComponentType in config/component_kinds.go which serves
// the config loading layer. These types are intentionally separate to avoid
// coupling between domain validation and config loading concerns.
type ComponentType struct {
	// Extensions are the file extensions belonging to this component type (e.g., [".go"], [".md", ".markdown"])
	Extensions []string `yaml:"extensions" json:"extensions"`

	// Linter is the linter tool to use (e.g., "go-lint", "markdown-lint")
	Linter string `yaml:"linter,omitempty" json:"linter,omitempty"`

	// LintInput specifies how files are passed to the linter:
	// - "packages": Lint by directory/package (e.g., Go modules use ./...)
	// - "files": Lint individual files (e.g., markdown files)
	LintInput string `yaml:"lint_input,omitempty" json:"lint_input,omitempty"`
}

// GetLintInput returns the lint input mode, defaulting to "files" if not specified.
func (c *ComponentType) GetLintInput() string {
	if c.LintInput == "" {
		return "files"
	}
	return c.LintInput
}

// HasLinter returns true if this component type has a linter configured.
func (c *ComponentType) HasLinter() bool {
	return c.Linter != ""
}

// ResolvedComponent is a component with its actual files computed at runtime.
type ResolvedComponent struct {
	Type  string   `json:"type"`
	Files []string `json:"files"`
	Count int      `json:"count"`
}
