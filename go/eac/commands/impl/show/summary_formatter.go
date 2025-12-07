// Package show provides commands for displaying repository information
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SummaryFormatter provides utilities for pretty-printing summaries
type SummaryFormatter struct {
	module    string
	status    string
	startTime time.Time
}

// NewSummaryFormatter creates a new formatter
func NewSummaryFormatter(module, status string) *SummaryFormatter {
	return &SummaryFormatter{
		module:    module,
		status:    status,
		startTime: time.Now(),
	}
}

// Header generates a summary header
func (f *SummaryFormatter) Header(summaryType string) string {
	var emoji string
	switch summaryType {
	case "build":
		emoji = "📦"
	case "test":
		emoji = "🧪"
	default:
		emoji = "📋"
	}
	return fmt.Sprintf("## %s %s: %s\n\n", emoji, strings.Title(summaryType), f.module)
}

// StatusSection generates a status section
func (f *SummaryFormatter) StatusSection(message string) string {
	emoji := "✅"
	if f.status != "success" {
		emoji = "❌"
	}
	return fmt.Sprintf("### %s Status\n%s\n\n", emoji, message)
}

// Table generates a markdown table
func (f *SummaryFormatter) Table(headers []string, rows [][]string) string {
	var sb strings.Builder

	// Headers
	sb.WriteString("|")
	for _, h := range headers {
		sb.WriteString(fmt.Sprintf(" %s |", h))
	}
	sb.WriteString("\n|")

	// Separator
	for range headers {
		sb.WriteString("--------|")
	}
	sb.WriteString("\n")

	// Rows
	for _, row := range rows {
		sb.WriteString("|")
		for _, cell := range row {
			sb.WriteString(fmt.Sprintf(" %s |", cell))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

// Section generates a section with title
func (f *SummaryFormatter) Section(title, content string) string {
	return fmt.Sprintf("### %s\n%s\n\n", title, content)
}

// CollapsibleSection generates a collapsible details section
func (f *SummaryFormatter) CollapsibleSection(summary, content string) string {
	return fmt.Sprintf("<details>\n<summary>%s</summary>\n\n%s\n</details>\n\n", summary, content)
}

// CodeBlock generates a code block
func (f *SummaryFormatter) CodeBlock(language, code string) string {
	return fmt.Sprintf("```%s\n%s\n```\n\n", language, code)
}

// Divider generates a horizontal rule
func (f *SummaryFormatter) Divider() string {
	return "---\n\n"
}

// Footer generates a summary footer
func (f *SummaryFormatter) Footer(duration time.Duration) string {
	var icon string
	if f.status == "success" {
		icon = "✅"
	} else {
		icon = "❌"
	}
	return fmt.Sprintf("---\n*%s Completed in %.1fs*\n", icon, duration.Seconds())
}

// GetFileCount returns the count of files in a directory matching a pattern
func GetFileCount(dir, pattern string) (int, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, nil
	}

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return 0, err
	}
	return len(matches), nil
}

// GetDirectorySize returns the size of a directory in human-readable format
func GetDirectorySize(dir string) (string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "0 B", nil
	}

	var size int64
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return formatBytes(size), nil
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// BulletList generates a bullet list
func (f *SummaryFormatter) BulletList(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
	sb.WriteString("\n")
	return sb.String()
}

// Link generates a markdown link
func Link(text, url string) string {
	return fmt.Sprintf("[%s](%s)", text, url)
}

// Bold generates bold text
func Bold(text string) string {
	return fmt.Sprintf("**%s**", text)
}

// Code generates inline code
func Code(text string) string {
	return fmt.Sprintf("`%s`", text)
}

// Emoji returns an emoji for a given status or type
func Emoji(name string) string {
	emojis := map[string]string{
		"success":     "✅",
		"failure":     "❌",
		"warning":     "⚠️",
		"info":        "ℹ️",
		"build":       "📦",
		"test":        "🧪",
		"metrics":     "📊",
		"diagnostics": "🔍",
		"tips":        "💡",
		"config":      "⚙️",
		"artifact":    "📁",
		"time":        "⏱️",
		"chart":       "📈",
	}
	if emoji, ok := emojis[name]; ok {
		return emoji
	}
	return ""
}

// readLogTail reads the last N lines from a log file
func readLogTail(path string, lines int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// Split into lines
	allLines := strings.Split(string(data), "\n")

	// Get last N lines
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}

	tailLines := allLines[start:]
	return strings.Join(tailLines, "\n")
}
