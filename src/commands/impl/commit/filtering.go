package commit

import (
	"strings"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/core/contracts"
)

// stripAgentNoise removes common initialization/greeting patterns from agent output
// using the contract-based anti-corruption framework.
//
// This function uses the generalized contracts.ApplyWithFallback which loads
// anti-corruption rules from contracts/commit-message/0.1.0/anti-corruption.yml
// and applies them to filter out conversational wrappers, initialization messages, and other noise.
//
// The agentType parameter determines the content start marker:
//   - "top-level": looks for conventional commit format (<type>(<scope>): <summary>)
//   - "module": looks for plain module names
func stripAgentNoise(output string, agentType string, workspaceRoot string) string {
	// Load anti-corruption rules using contract loader
	loader := contracts.NewSpecContractLoader(workspaceRoot, "ai/commit-message", "0.1.0")
	rules, err := loader.LoadAntiCorruptionRules()

	// Determine content marker based on agent type
	var contentMarker string
	if agentType == "top-level" {
		// For top-level, look for conventional commit patterns
		// We'll use pattern matching as a pre-filter, then apply contract rules
		lines := strings.Split(output, "\n")
		matcher := &commitPatternMatcher{agentType: agentType}
		firstValidLineIdx := matcher.findFirstValidLine(lines)
		if firstValidLineIdx != -1 {
			// Found valid pattern - start from there
			output = strings.Join(lines[firstValidLineIdx:], "\n")
		}
		contentMarker = "" // Let anti-corruption rules handle the rest
	} else {
		// For module sections, content starts with module name (lowercase alphanumeric)
		contentMarker = "" // Anti-corruption rules will handle it
	}

	// Apply anti-corruption with fallback to hardcoded patterns
	if err != nil {
		// Fallback: use hardcoded rules if contract not available
		return applyHardcodedFiltering(output, contentMarker)
	}

	return contracts.ApplyWithFallback(output, rules, contentMarker)
}

// commitPatternMatcher handles pattern matching for different agent types
type commitPatternMatcher struct {
	agentType string
}

func (m *commitPatternMatcher) findFirstValidLine(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m.agentType == "top-level" {
			if m.isValidTopLevelLine(trimmed) {
				return i
			}
		} else if m.agentType == "module" {
			if m.isValidModuleLine(trimmed) {
				return i
			}
		} else {
			// Unknown agent type - use generic heuristic
			if m.isValidTopLevelLine(trimmed) {
				return i
			}
		}
	}
	return -1
}

func (m *commitPatternMatcher) isValidTopLevelLine(trimmed string) bool {
	// Conventional commit with scope
	// Format: <type>(<scope>): <summary>
	// Examples: feat(multi-module): add feature, fix(cli): resolve bug
	for _, commitType := range commitmessage.StandardCommitTypes {
		// Look for <type>(<scope>): pattern
		pattern := commitType + "("
		if strings.HasPrefix(trimmed, pattern) {
			// Verify it has closing paren and colon
			if strings.Contains(trimmed, "):") {
				return true
			}
		}
	}

	return false
}

func (m *commitPatternMatcher) isValidModuleLine(trimmed string) bool {
	// Plain module name (followed by dashes on next line)
	// Must be:
	// - Not empty
	// - Not markdown headers (no #)
	// - Not dashes
	// - Not conventional commit format (no : in first part)
	// - Lowercase alphanumeric with dashes/underscores only
	if trimmed != "" &&
		!strings.HasPrefix(trimmed, "#") &&
		trimmed != "---" &&
		!strings.Contains(trimmed, ":") {
		// Check if it looks like a module name (lowercase alphanumeric with dashes)
		isModuleName := true
		for _, ch := range trimmed {
			if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
				isModuleName = false
				break
			}
		}
		if isModuleName {
			return true
		}
	}

	return false
}

// applyHardcodedFiltering applies basic hardcoded filtering when contract not available
func applyHardcodedFiltering(output string, contentStartMarker string) string {
	output = strings.TrimSpace(output)

	// Remove code fences
	if strings.HasPrefix(output, "```") {
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		output = strings.Join(lines, "\n")
	}

	lines := strings.Split(output, "\n")
	return filterWithHardcodedRules(lines)
}

// filterWithHardcodedRules applies hardcoded noise patterns when YAML rules unavailable
func filterWithHardcodedRules(lines []string) string {
	var cleaned []string
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip common noise patterns (hardcoded)
		if strings.Contains(trimmed, "**Initialized and ready") ||
			strings.Contains(trimmed, "**INITIALIZED") ||
			strings.Contains(trimmed, "Loading project context") ||
			strings.Contains(trimmed, "ready to assist") ||
			(len(trimmed) > 0 && trimmed[0] > 127 && strings.Contains(trimmed, "**")) {
			continue
		}

		// Skip horizontal rules before content starts
		if !foundContent && (trimmed == "---" || trimmed == "___" || trimmed == "***") {
			continue
		}

		// Skip empty lines before content starts
		if !foundContent && trimmed == "" {
			continue
		}

		// Content has started
		if trimmed != "" {
			foundContent = true
		}

		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}
