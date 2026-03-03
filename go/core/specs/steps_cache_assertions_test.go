// Package specs contains godog step implementations for core features.
//
// This file contains YAML parsing, YAML field assertions, and command output
// pattern-matching assertions for cache invalidation tests.
package specs

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"gopkg.in/yaml.v3"
)

func parseYAMLOutput(ctx *eacgodog.TestContext) {
	cacheCtx.parsedYAML = nil
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(ctx.CommandOutput), &parsed); err == nil {
		cacheCtx.parsedYAML = parsed
	}
}

func yamlFieldEquals(ctx *eacgodog.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	actual := fmt.Sprintf("%v", value)
	if actual != expected {
		return fmt.Errorf("expected %s to be %q, got %q", fieldPath, expected, actual)
	}
	return nil
}

func yamlFieldContains(ctx *eacgodog.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		if !strings.Contains(v, expected) {
			return fmt.Errorf("expected %s to contain %q, got %q\nFull output:\n%s", fieldPath, expected, v, ctx.CommandOutput)
		}
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == expected {
				return nil
			}
		}
		return fmt.Errorf("expected %s to contain %q, got %v\nFull output:\n%s", fieldPath, expected, v, ctx.CommandOutput)
	default:
		if !strings.Contains(fmt.Sprintf("%v", value), expected) {
			return fmt.Errorf("expected %s to contain %q, got %v\nFull output:\n%s", fieldPath, expected, value, ctx.CommandOutput)
		}
	}
	return nil
}

func yamlFieldNotContains(ctx *eacgodog.TestContext, fieldPath, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		if strings.Contains(v, expected) {
			return fmt.Errorf("expected %s NOT to contain %q, but it did", fieldPath, expected)
		}
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == expected {
				return fmt.Errorf("expected %s NOT to contain %q, but it did", fieldPath, expected)
			}
		}
	}
	return nil
}

func yamlFieldEmpty(ctx *eacgodog.TestContext, fieldPath string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return nil
	}

	switch v := value.(type) {
	case []interface{}:
		if len(v) != 0 {
			return fmt.Errorf("expected %s to be empty, got %v", fieldPath, v)
		}
	case string:
		if v != "" {
			return fmt.Errorf("expected %s to be empty, got %q", fieldPath, v)
		}
	case nil:
		// nil is empty
	default:
		return fmt.Errorf("expected %s to be empty, got %v", fieldPath, value)
	}
	return nil
}

func yamlFieldContainsExactly(ctx *eacgodog.TestContext, fieldPath string, count int, expected string) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	var occurrences int
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == expected {
				occurrences++
			}
		}
	case string:
		occurrences = strings.Count(v, expected)
	}

	if occurrences != count {
		return fmt.Errorf("expected %s to contain exactly %d occurrences of %q, got %d", fieldPath, count, expected, occurrences)
	}
	return nil
}

func yamlFieldUnderMs(ctx *eacgodog.TestContext, fieldPath string, maxMs int) error {
	value, err := extractYAMLField(ctx, fieldPath)
	if err != nil {
		return err
	}

	durationStr := fmt.Sprintf("%v", value)
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("failed to parse duration %q: %w", durationStr, err)
	}

	if duration.Milliseconds() > int64(maxMs) {
		return fmt.Errorf("expected %s under %dms, got %v", fieldPath, maxMs, duration)
	}
	return nil
}

func extractYAMLField(ctx *eacgodog.TestContext, fieldPath string) (interface{}, error) {
	if cacheCtx.parsedYAML == nil {
		output := ctx.CommandOutput
		yamlStart := findYAMLStart(output)
		if yamlStart > 0 {
			output = output[yamlStart:]
		}

		var data map[string]interface{}
		if err := yaml.Unmarshal([]byte(output), &data); err != nil {
			return nil, fmt.Errorf("failed to parse YAML output: %w\nOutput: %s", err, ctx.CommandOutput)
		}
		cacheCtx.parsedYAML = data
	}

	parts := strings.Split(fieldPath, ".")
	var current interface{} = cacheCtx.parsedYAML

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field %q not found at path %s", part, fieldPath)
			}
		case map[interface{}]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field %q not found at path %s", part, fieldPath)
			}
		default:
			return nil, fmt.Errorf("cannot navigate into %T at %s", current, part)
		}
	}

	return current, nil
}

func findYAMLStart(output string) int {
	patterns := []string{"modules:", "is_fresh_build:", "directly_changed:", "base_sha:", "module:", "head_sha:", "all_components:"}
	minIndex := len(output)

	for _, pattern := range patterns {
		idx := strings.Index(output, pattern)
		if idx >= 0 && idx < minIndex {
			lineStart := idx
			for lineStart > 0 && output[lineStart-1] != '\n' {
				lineStart--
			}
			if lineStart < minIndex {
				minIndex = lineStart
			}
		}
	}

	if minIndex < len(output) {
		return minIndex
	}
	return 0
}

func outputIndicatesWouldBuild(ctx *eacgodog.TestContext, module string) error {
	patterns := []string{
		fmt.Sprintf("%s.*would be built", module),
		fmt.Sprintf("%s.*will build", module),
		fmt.Sprintf("building.*%s", module),
		fmt.Sprintf("%s.*needs rebuild", module),
		fmt.Sprintf("%s.*changed", module),
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, ctx.CommandOutput)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("output does not indicate %s would be built. Output: %s", module, ctx.CommandOutput)
}

func outputIndicatesSkipped(ctx *eacgodog.TestContext, module string) error {
	patterns := []string{
		fmt.Sprintf("%s.*up-to-date", module),
		fmt.Sprintf("%s.*skipped", module),
		fmt.Sprintf("%s.*skip", module),
		fmt.Sprintf("skipping.*%s", module),
		fmt.Sprintf("%s.*cached", module),
		fmt.Sprintf("cached.*%s", module),
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, ctx.CommandOutput)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("output does not indicate %s is up-to-date or would be skipped. Output: %s", module, ctx.CommandOutput)
}

func outputIndicatesWouldLint(ctx *eacgodog.TestContext, module string) error {
	patterns := []string{
		fmt.Sprintf("%s.*would be linted", module),
		fmt.Sprintf("%s.*will lint", module),
		fmt.Sprintf("linting.*%s", module),
		fmt.Sprintf("%s.*needs linting", module),
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, ctx.CommandOutput)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("output does not indicate %s would be linted. Output: %s", module, ctx.CommandOutput)
}
