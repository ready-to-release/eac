package dependabot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DependabotPath is the path to dependabot.yml relative to the repo root.
const DependabotPath = ".github/dependabot.yml"

// Parse reads and parses the dependabot.yml file.
// If the file does not exist, returns an empty config with version 2.
func Parse(repoRoot string) (*DependabotConfig, error) {
	path := filepath.Join(repoRoot, DependabotPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &DependabotConfig{Version: 2}, nil
		}
		return nil, fmt.Errorf("reading dependabot.yml: %w", err)
	}

	return ParseBytes(data)
}

// parseContext tracks which YAML list field is currently being filled.
type parseContext int

const (
	ctxNone parseContext = iota
	ctxLabels
	ctxReviewers
	ctxAssignees
	ctxDirectories
	ctxGroupPatterns
	ctxGroupExcludePatterns
)

// ParseBytes parses dependabot.yml content from raw bytes.
func ParseBytes(data []byte) (*DependabotConfig, error) {
	config := &DependabotConfig{Version: 2}
	var current *UpdateEntry
	ctx := ctxNone
	currentGroupName := ""
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if trimmed == "version: 2" {
			continue
		}

		if trimmed == "updates:" {
			continue
		}

		// New entry starts with "- package-ecosystem:"
		if strings.HasPrefix(trimmed, "- package-ecosystem:") {
			if current != nil {
				config.Updates = append(config.Updates, *current)
			}
			current = &UpdateEntry{}
			current.PackageEcosystem = extractYAMLValue(trimmed[len("- package-ecosystem:"):])
			ctx = ctxNone
			currentGroupName = ""
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(trimmed, "directory:") && !strings.HasPrefix(trimmed, "directories:") {
			current.Directory = extractYAMLValue(trimmed[len("directory:"):])
			ctx = ctxNone
		} else if strings.HasPrefix(trimmed, "directories:") {
			ctx = ctxDirectories
		} else if strings.HasPrefix(trimmed, "interval:") {
			current.Schedule.Interval = extractYAMLValue(trimmed[len("interval:"):])
			ctx = ctxNone
		} else if strings.HasPrefix(trimmed, "day:") {
			current.Schedule.Day = extractYAMLValue(trimmed[len("day:"):])
			ctx = ctxNone
		} else if strings.HasPrefix(trimmed, "time:") {
			current.Schedule.Time = extractYAMLValue(trimmed[len("time:"):])
			ctx = ctxNone
		} else if strings.HasPrefix(trimmed, "timezone:") {
			current.Schedule.Timezone = extractYAMLValue(trimmed[len("timezone:"):])
			ctx = ctxNone
		} else if strings.HasPrefix(trimmed, "prefix:") {
			if current.CommitMessage == nil {
				current.CommitMessage = &CommitMessage{}
			}
			current.CommitMessage.Prefix = extractYAMLValue(trimmed[len("prefix:"):])
			ctx = ctxNone
		} else if trimmed == "labels:" {
			ctx = ctxLabels
		} else if trimmed == "reviewers:" {
			ctx = ctxReviewers
		} else if trimmed == "assignees:" {
			ctx = ctxAssignees
		} else if trimmed == "groups:" {
			if current.Groups == nil {
				current.Groups = make(map[string]Group)
			}
			ctx = ctxNone
		} else if trimmed == "patterns:" {
			ctx = ctxGroupPatterns
		} else if trimmed == "exclude-patterns:" {
			ctx = ctxGroupExcludePatterns
		} else if strings.HasPrefix(trimmed, "group-by:") {
			if currentGroupName != "" && current.Groups != nil {
				g := current.Groups[currentGroupName]
				g.GroupBy = extractYAMLValue(trimmed[len("group-by:"):])
				current.Groups[currentGroupName] = g
			}
		} else if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "- ") &&
			trimmed != "schedule:" && trimmed != "commit-message:" &&
			trimmed != "open-pull-requests-limit:" {
			// Could be a group name like "golang-x:" or "all-go-deps:"
			name := strings.TrimSuffix(trimmed, ":")
			if current.Groups != nil {
				currentGroupName = name
				if _, exists := current.Groups[name]; !exists {
					current.Groups[name] = Group{}
				}
			}
		} else if strings.HasPrefix(trimmed, "- ") {
			val := extractYAMLValue(trimmed[len("- "):])
			switch ctx {
			case ctxDirectories:
				current.Directories = append(current.Directories, val)
			case ctxLabels:
				current.Labels = append(current.Labels, val)
			case ctxReviewers:
				current.Reviewers = append(current.Reviewers, val)
			case ctxAssignees:
				current.Assignees = append(current.Assignees, val)
			case ctxGroupPatterns:
				if currentGroupName != "" && current.Groups != nil {
					g := current.Groups[currentGroupName]
					g.Patterns = append(g.Patterns, val)
					current.Groups[currentGroupName] = g
				}
			case ctxGroupExcludePatterns:
				if currentGroupName != "" && current.Groups != nil {
					g := current.Groups[currentGroupName]
					g.ExcludePatterns = append(g.ExcludePatterns, val)
					current.Groups[currentGroupName] = g
				}
			default:
				// Legacy fallback: assume labels
				current.Labels = append(current.Labels, val)
			}
		}
	}

	if current != nil {
		config.Updates = append(config.Updates, *current)
	}

	return config, nil
}

// extractYAMLValue extracts a value from a YAML line, removing quotes and whitespace.
func extractYAMLValue(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"'")
	return s
}

// Write writes the dependabot.yml file with proper formatting matching the existing style.
func Write(repoRoot string, config *DependabotConfig) error {
	path := filepath.Join(repoRoot, DependabotPath)

	content := FormatConfig(config)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating .github directory: %w", err)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

// FormatConfig produces a formatted dependabot.yml string matching the existing style.
func FormatConfig(config *DependabotConfig) string {
	var b strings.Builder
	b.WriteString("version: 2\n\nupdates:\n")

	prevEcosystem := ""
	for i, u := range config.Updates {
		ecosystem := u.PackageEcosystem

		// Add section separator comments between ecosystem groups
		if ecosystem != prevEcosystem && i > 0 {
			b.WriteString("\n")
		}

		// Add comment for the entry
		comment := entryComment(u)
		if comment != "" {
			fmt.Fprintf(&b, "  # %s\n", comment)
		}

		// Write the entry
		fmt.Fprintf(&b, "  - package-ecosystem: %q\n", u.PackageEcosystem)
		if len(u.Directories) > 0 {
			b.WriteString("    directories:\n")
			for _, d := range u.Directories {
				fmt.Fprintf(&b, "      - %q\n", d)
			}
		} else {
			fmt.Fprintf(&b, "    directory: %q\n", u.Directory)
		}
		b.WriteString("    schedule:\n")
		fmt.Fprintf(&b, "      interval: %q\n", u.Schedule.Interval)
		if u.Schedule.Day != "" {
			fmt.Fprintf(&b, "      day: %q\n", u.Schedule.Day)
		}
		if u.Schedule.Time != "" {
			fmt.Fprintf(&b, "      time: %q\n", u.Schedule.Time)
		}
		if u.Schedule.Timezone != "" {
			fmt.Fprintf(&b, "      timezone: %q\n", u.Schedule.Timezone)
		}
		if u.CommitMessage != nil && u.CommitMessage.Prefix != "" {
			b.WriteString("    commit-message:\n")
			fmt.Fprintf(&b, "      prefix: %q\n", u.CommitMessage.Prefix)
		}
		if len(u.Labels) > 0 {
			b.WriteString("    labels:\n")
			for _, label := range u.Labels {
				fmt.Fprintf(&b, "      - %q\n", label)
			}
		}
		if len(u.Reviewers) > 0 {
			b.WriteString("    reviewers:\n")
			for _, r := range u.Reviewers {
				fmt.Fprintf(&b, "      - %q\n", r)
			}
		}
		if len(u.Assignees) > 0 {
			b.WriteString("    assignees:\n")
			for _, a := range u.Assignees {
				fmt.Fprintf(&b, "      - %q\n", a)
			}
		}
		if u.OpenPullRequests != nil {
			fmt.Fprintf(&b, "    open-pull-requests-limit: %d\n", *u.OpenPullRequests)
		}
		if len(u.Groups) > 0 {
			b.WriteString("    groups:\n")
			names := make([]string, 0, len(u.Groups))
			for name := range u.Groups {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				g := u.Groups[name]
				fmt.Fprintf(&b, "      %s:\n", name)
				if len(g.Patterns) > 0 {
					b.WriteString("        patterns:\n")
					for _, p := range g.Patterns {
						fmt.Fprintf(&b, "          - %q\n", p)
					}
				}
				if g.GroupBy != "" {
					fmt.Fprintf(&b, "        group-by: %q\n", g.GroupBy)
				}
				if len(g.ExcludePatterns) > 0 {
					b.WriteString("        exclude-patterns:\n")
					for _, p := range g.ExcludePatterns {
						fmt.Fprintf(&b, "          - %q\n", p)
					}
				}
			}
		}

		prevEcosystem = ecosystem
	}

	return b.String()
}

// entryComment generates a descriptive comment for a dependabot entry.
func entryComment(u UpdateEntry) string {
	switch u.PackageEcosystem {
	case "github-actions":
		return "GitHub Actions"
	case "gomod":
		if u.IsConsolidated() {
			return "Go modules - all workspace modules (consolidated)"
		}
		dir := strings.TrimPrefix(u.Directory, "/")
		return fmt.Sprintf("Go modules - %s", dir)
	case "npm":
		dir := strings.TrimPrefix(u.Directory, "/")
		return fmt.Sprintf("npm - %s", dir)
	case "docker":
		dir := strings.TrimPrefix(u.Directory, "/")
		return fmt.Sprintf("Docker - %s", dir)
	case "pip":
		dir := strings.TrimPrefix(u.Directory, "/")
		return fmt.Sprintf("Python - %s", dir)
	default:
		return ""
	}
}

// DefaultUpdateEntry creates a new UpdateEntry with standard defaults for a discovered entry.
func DefaultUpdateEntry(entry EcosystemEntry) UpdateEntry {
	labels := []string{"dependencies"}

	switch entry.Ecosystem {
	case "gomod":
		labels = append(labels, "go")
	case "npm":
		labels = append(labels, "npm")
	case "pip":
		labels = append(labels, "python")
	case "docker":
		labels = append(labels, "docker")
	case "github-actions":
		labels = append(labels, "github-actions")
	}

	return UpdateEntry{
		PackageEcosystem: entry.Ecosystem,
		Directory:        entry.Directory,
		Schedule: Schedule{
			Interval: "weekly",
			Day:      "monday",
		},
		CommitMessage: &CommitMessage{
			Prefix: "chore(deps)",
		},
		Labels: labels,
	}
}
