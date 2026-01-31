// Command: get release-bundle
// Flag.with-versions: type=bool, default=false, usage=Resolve current versions from GitHub releases
// Flag.title-only: type=bool, default=false, usage=Output only the resolved release title
// Flag.format: type=string, usage=Output format: markdown, flat, shell, table
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--format markdown: Output as GitHub release notes markdown
//	--format flat: One module per line (moniker|version|tag|category)
//	--format shell: Shell variables for eval (RELEASE_MAP, ALL_RELEASED)
//	--format table: Markdown table for summaries
//
// Long:
// Long: Expected Output:
// Long: Structured release bundle configuration including:
// Long:   - title_format: Template for release title (with versions resolved if --with-versions)
// Long:   - headline: Map of label -> module info for title modules
// Long:   - categories: Grouped modules with their details
// Long: Use this in CI to create release notes without hardcoding module names.
// Long:
// Long: With --with-versions, each module includes:
// Long:   - version: Current released version (empty if not released)
// Long:   - tag: Full release tag (e.g., r2r-cli/1.0.0)
// Long:   - release_url: GitHub release URL
// Long:
// Long: With --format markdown, outputs release notes directly:
// Long:   ## Release Bundle
// Long:   ### Category Name
// Long:   | Module   | Version | Link                                                           |
// Long:   |----------|---------|----------------------------------------------------------------|
// Long:   | **Name** | `1.0.0` | [Release](https://github.com/org/repo/releases/tag/name/1.0.0) |
// Long:
// Long: With --format shell, outputs for eval:
// Long:   RELEASE_MAP='{"r2r-cli":{"tag":"r2r-cli/1.0.0","version":"1.0.0"},...}'
// Long:   ALL_RELEASED="true"
// Long:
// Long: Example:
// Long:   eval $(get release-bundle --format shell)
package get

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/domain/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetReleaseBundle)
}

// ReleaseBundleOutput is the structured output for get release-bundle.
type ReleaseBundleOutput struct {
	TitleFormat string                        `yaml:"title_format" json:"title_format"`
	Headline    map[string]HeadlineModule     `yaml:"headline" json:"headline"`
	Categories  []ReleaseBundleCategoryOutput `yaml:"categories" json:"categories"`
}

// HeadlineModule contains info about a headline module.
type HeadlineModule struct {
	Moniker    string `yaml:"moniker" json:"moniker"`
	Name       string `yaml:"name" json:"name"`
	Versioning string `yaml:"versioning" json:"versioning"`
	Version    string `yaml:"version,omitempty" json:"version,omitempty"`
	Tag        string `yaml:"tag,omitempty" json:"tag,omitempty"`
	ReleaseURL string `yaml:"release_url,omitempty" json:"release_url,omitempty"`
}

// ReleaseBundleCategoryOutput contains category info with enriched module details.
type ReleaseBundleCategoryOutput struct {
	Name        string                `yaml:"name" json:"name"`
	Description string                `yaml:"description,omitempty" json:"description,omitempty"`
	Modules     []ReleaseBundleModule `yaml:"modules" json:"modules"`
}

// ReleaseBundleModule contains module info for release bundle.
type ReleaseBundleModule struct {
	Moniker    string `yaml:"moniker" json:"moniker"`
	Name       string `yaml:"name" json:"name"`
	Versioning string `yaml:"versioning" json:"versioning"`
	Version    string `yaml:"version,omitempty" json:"version,omitempty"`
	Tag        string `yaml:"tag,omitempty" json:"tag,omitempty"`
	ReleaseURL string `yaml:"release_url,omitempty" json:"release_url,omitempty"`
}

// releaseBundleFlags defines valid flags for the get release-bundle command

func GetReleaseBundle() int {
	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Parse flags
	withVersions := false
	titleOnly := false
	format := ""

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--with-versions":
			withVersions = true
		case arg == "--title-only":
			titleOnly = true
			withVersions = true // title needs versions resolved
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		}
	}

	// Handle title-only output (just the resolved title string)
	if titleOnly {
		data, err := getReleaseBundleData(workspaceRoot, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Println(data.TitleFormat)
		return 0
	}

	// Handle markdown format output
	if format == "markdown" {
		// Force with-versions for markdown output
		data, err := getReleaseBundleData(workspaceRoot, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Println(formatAsMarkdown(data))
		return 0
	}

	// Handle flat format output (one module per line for simple iteration)
	if format == "flat" {
		// Force with-versions for flat output
		data, err := getReleaseBundleData(workspaceRoot, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Println(formatAsFlat(data))
		return 0
	}

	// Handle shell format output (for eval in bash)
	if format == "shell" {
		// Force with-versions for shell output
		data, err := getReleaseBundleData(workspaceRoot, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(formatAsShell(data))
		return 0
	}

	// Handle table format output (markdown table for summaries)
	if format == "table" {
		// Force with-versions for table output
		data, err := getReleaseBundleData(workspaceRoot, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(formatAsTable(data))
		return 0
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return getReleaseBundleData(workspaceRoot, withVersions)
	})
}

// formatAsFlat generates flat output for simple bash iteration.
// Format: moniker|version|tag|category (one per line).
func formatAsFlat(data *ReleaseBundleOutput) string {
	var sb strings.Builder

	for _, cat := range data.Categories {
		for _, mod := range cat.Modules {
			version := mod.Version
			if version == "" {
				version = "-"
			}
			tag := mod.Tag
			if tag == "" {
				tag = "-"
			}
			fmt.Fprintf(&sb, "%s|%s|%s|%s\n", mod.Moniker, version, tag, cat.Name)
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// formatAsShell generates shell variable output for eval.
// Output: RELEASE_MAP='{"mod":{"tag":"...","version":"..."},...}'
//
//	ALL_RELEASED="true|false"
func formatAsShell(data *ReleaseBundleOutput) string {
	var sb strings.Builder

	// Build release map JSON
	allReleased := true
	var mapParts []string
	for _, cat := range data.Categories {
		for _, mod := range cat.Modules {
			if mod.Version != "" && mod.Tag != "" {
				// Escape quotes in values
				tag := strings.ReplaceAll(mod.Tag, `"`, `\"`)
				version := strings.ReplaceAll(mod.Version, `"`, `\"`)
				mapParts = append(mapParts, fmt.Sprintf(`"%s":{"tag":"%s","version":"%s"}`,
					mod.Moniker, tag, version))
			} else {
				allReleased = false
			}
		}
	}

	releaseMap := "{" + strings.Join(mapParts, ",") + "}"
	fmt.Fprintf(&sb, "RELEASE_MAP='%s'\n", releaseMap)
	fmt.Fprintf(&sb, "ALL_RELEASED=\"%t\"\n", allReleased)

	return sb.String()
}

// formatAsTable generates markdown table output for summaries.
func formatAsTable(data *ReleaseBundleOutput) string {
	tb := render.NewTableBuilder().
		WithHeaders("Module", "Version", "Status")

	for _, cat := range data.Categories {
		for _, mod := range cat.Modules {
			if mod.Version != "" {
				tb.AddRow(mod.Moniker, fmt.Sprintf("`%s`", mod.Version), ":white_check_mark:")
			} else {
				tb.AddRow(mod.Moniker, "—", ":grey_question:")
			}
		}
	}

	return tb.Build()
}

// formatAsMarkdown generates release notes markdown from bundle data.
func formatAsMarkdown(data *ReleaseBundleOutput) string {
	var sb strings.Builder

	sb.WriteString("## Release Bundle\n\n")
	sb.WriteString("This release provides links to the latest versions of all modules in the EAC ecosystem.\n\n")

	for _, cat := range data.Categories {
		sb.WriteString("### " + cat.Name + "\n\n")
		if cat.Description != "" {
			sb.WriteString(cat.Description + "\n\n")
		}

		tb := render.NewTableBuilder().
			WithHeaders("Module", "Version", "Link")

		for _, mod := range cat.Modules {
			if mod.Version != "" {
				tb.AddRow(
					fmt.Sprintf("**%s**", mod.Name),
					fmt.Sprintf("`%s`", mod.Version),
					fmt.Sprintf("[GitHub Release](%s)", mod.ReleaseURL),
				)
			} else {
				tb.AddRow(fmt.Sprintf("**%s**", mod.Name), "*not released*", "-")
			}
		}
		sb.WriteString(tb.Build())
		sb.WriteString("\n")
	}

	return sb.String()
}

// getGitHubReleases fetches release tags from GitHub using gh CLI.
func getGitHubReleases() (map[string]string, error) {
	// Run: gh release list --limit 100 --json tagName -q '.[].tagName'
	cmd := exec.Command("gh", "release", "list", "--limit", "100", "--json", "tagName", "-q", ".[].tagName")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub releases: %w", err)
	}

	// Parse output - one tag per line
	releases := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		tag := strings.TrimSpace(line)
		if tag == "" {
			continue
		}
		// Parse tag format: <moniker>/<version>
		parts := strings.SplitN(tag, "/", 2)
		if len(parts) == 2 {
			moniker := parts[0]
			// Only keep the latest version for each module (first seen = latest)
			if _, exists := releases[moniker]; !exists {
				releases[moniker] = tag
			}
		}
	}

	return releases, nil
}

// getGitHubRepoURL returns the GitHub repository URL.
func getGitHubRepoURL() string {
	// Try to get from GITHUB_REPOSITORY env var
	if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
		return "https://github.com/" + repo
	}
	// Fallback: use gh CLI to get repo info
	cmd := exec.Command("gh", "repo", "view", "--json", "url", "-q", ".url")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getReleaseBundleData(workspaceRoot string, withVersions bool) (*ReleaseBundleOutput, error) {
	// Load module contracts via the reports API (same as get modules)
	report, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load modules: %w", err)
	}

	// Find the release module with release_bundle config
	var releaseBundle *struct {
		TitleFormat string
		Headline    map[string]string
		Categories  []struct {
			Name        string
			Description string
			Modules     []string
		}
	}

	for _, mod := range report.Modules {
		if mod.ReleaseBundle != nil {
			releaseBundle = &struct {
				TitleFormat string
				Headline    map[string]string
				Categories  []struct {
					Name        string
					Description string
					Modules     []string
				}
			}{
				TitleFormat: mod.ReleaseBundle.TitleFormat,
				Headline:    mod.ReleaseBundle.Headline,
			}
			for _, cat := range mod.ReleaseBundle.Categories {
				releaseBundle.Categories = append(releaseBundle.Categories, struct {
					Name        string
					Description string
					Modules     []string
				}{
					Name:        cat.Name,
					Description: cat.Description,
					Modules:     cat.Modules,
				})
			}
			break
		}
	}

	if releaseBundle == nil {
		return nil, fmt.Errorf("no module with release_bundle configuration found")
	}

	// Fetch GitHub releases if --with-versions
	var releases map[string]string
	var repoURL string
	if withVersions {
		releases, err = getGitHubReleases()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch GitHub releases: %w", err)
		}
		repoURL = getGitHubRepoURL()
	}

	// Build output
	output := &ReleaseBundleOutput{
		TitleFormat: releaseBundle.TitleFormat,
		Headline:    make(map[string]HeadlineModule),
		Categories:  []ReleaseBundleCategoryOutput{},
	}

	// Populate headline modules
	for label, moniker := range releaseBundle.Headline {
		mod, ok := report.GetModuleByMoniker(moniker)
		if !ok {
			return nil, fmt.Errorf("headline module %q not found", moniker)
		}
		versioning := ""
		if mod.Versioning != nil {
			versioning = mod.Versioning.Scheme
		}

		headline := HeadlineModule{
			Moniker:    moniker,
			Name:       mod.Name,
			Versioning: versioning,
		}

		// Add version info if requested
		if withVersions {
			if tag, ok := releases[moniker]; ok {
				parts := strings.SplitN(tag, "/", 2)
				if len(parts) == 2 {
					headline.Version = parts[1]
					headline.Tag = tag
					if repoURL != "" {
						headline.ReleaseURL = repoURL + "/releases/tag/" + tag
					}
				}
			}
		}

		output.Headline[label] = headline
	}

	// If --with-versions, resolve title format with actual versions
	if withVersions {
		resolvedTitle := releaseBundle.TitleFormat
		for label, headline := range output.Headline {
			placeholder := "{" + label + "_version}"
			version := headline.Version
			if version == "" {
				version = "unreleased"
			}
			resolvedTitle = strings.ReplaceAll(resolvedTitle, placeholder, version)
		}
		output.TitleFormat = resolvedTitle
	}

	// Populate categories with enriched module info
	for _, cat := range releaseBundle.Categories {
		catOutput := ReleaseBundleCategoryOutput{
			Name:        cat.Name,
			Description: cat.Description,
			Modules:     []ReleaseBundleModule{},
		}

		for _, moniker := range cat.Modules {
			mod, ok := report.GetModuleByMoniker(moniker)
			if !ok {
				return nil, fmt.Errorf("module %q in category %q not found", moniker, cat.Name)
			}
			versioning := ""
			if mod.Versioning != nil {
				versioning = mod.Versioning.Scheme
			}

			bundleMod := ReleaseBundleModule{
				Moniker:    moniker,
				Name:       mod.Name,
				Versioning: versioning,
			}

			// Add version info if requested
			if withVersions {
				if tag, ok := releases[moniker]; ok {
					parts := strings.SplitN(tag, "/", 2)
					if len(parts) == 2 {
						bundleMod.Version = parts[1]
						bundleMod.Tag = tag
						if repoURL != "" {
							bundleMod.ReleaseURL = repoURL + "/releases/tag/" + tag
						}
					}
				}
			}

			catOutput.Modules = append(catOutput.Modules, bundleMod)
		}

		output.Categories = append(output.Categories, catOutput)
	}

	return output, nil
}
