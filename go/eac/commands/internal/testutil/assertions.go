// Package testutil provides shared test utilities for command testing.
package testutil

import (
	"regexp"
	"strings"
	"testing"
)

// AssertContains fails if the string does not contain the substring.
func AssertContains(t *testing.T, s, substring string) {
	t.Helper()
	if !strings.Contains(s, substring) {
		t.Errorf("expected string to contain %q, got:\n%s", substring, s)
	}
}

// AssertNotContains fails if the string contains the substring.
func AssertNotContains(t *testing.T, s, substring string) {
	t.Helper()
	if strings.Contains(s, substring) {
		t.Errorf("expected string NOT to contain %q, got:\n%s", substring, s)
	}
}

// AssertContainsAll fails if the string does not contain all substrings.
func AssertContainsAll(t *testing.T, s string, substrings ...string) {
	t.Helper()
	for _, sub := range substrings {
		if !strings.Contains(s, sub) {
			t.Errorf("expected string to contain %q, got:\n%s", sub, s)
		}
	}
}

// AssertContainsAny fails if the string does not contain any of the substrings.
func AssertContainsAny(t *testing.T, s string, substrings ...string) {
	t.Helper()
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return
		}
	}
	t.Errorf("expected string to contain at least one of %v, got:\n%s", substrings, s)
}

// AssertMatches fails if the string does not match the regex pattern.
func AssertMatches(t *testing.T, s, pattern string) {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid regex pattern %q: %v", pattern, err)
	}
	if !re.MatchString(s) {
		t.Errorf("expected string to match pattern %q, got:\n%s", pattern, s)
	}
}

// AssertNotMatches fails if the string matches the regex pattern.
func AssertNotMatches(t *testing.T, s, pattern string) {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid regex pattern %q: %v", pattern, err)
	}
	if re.MatchString(s) {
		t.Errorf("expected string NOT to match pattern %q, got:\n%s", pattern, s)
	}
}

// AssertEqual fails if the values are not equal.
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

// AssertNotEqual fails if the values are equal.
func AssertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected == actual {
		t.Errorf("expected value to differ from %v", expected)
	}
}

// AssertNil fails if the value is not nil.
func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		t.Errorf("expected nil, got %v", value)
	}
}

// AssertNotNil fails if the value is nil.
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Error("expected non-nil value")
	}
}

// AssertTrue fails if the condition is false.
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("expected true: %s", msg)
	}
}

// AssertFalse fails if the condition is true.
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("expected false: %s", msg)
	}
}

// AssertEmpty fails if the string is not empty.
func AssertEmpty(t *testing.T, s string) {
	t.Helper()
	if s != "" {
		t.Errorf("expected empty string, got:\n%s", s)
	}
}

// AssertNotEmpty fails if the string is empty.
func AssertNotEmpty(t *testing.T, s string) {
	t.Helper()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

// AssertLineCount fails if the string does not have the expected number of lines.
func AssertLineCount(t *testing.T, s string, expected int) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	if s == "" {
		lines = []string{}
	}
	if len(lines) != expected {
		t.Errorf("expected %d lines, got %d:\n%s", expected, len(lines), s)
	}
}

// AssertMinLineCount fails if the string has fewer than the expected number of lines.
func AssertMinLineCount(t *testing.T, s string, minimum int) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	if s == "" {
		lines = []string{}
	}
	if len(lines) < minimum {
		t.Errorf("expected at least %d lines, got %d:\n%s", minimum, len(lines), s)
	}
}

// AssertHasPrefix fails if the string does not have the expected prefix.
func AssertHasPrefix(t *testing.T, s, prefix string) {
	t.Helper()
	if !strings.HasPrefix(s, prefix) {
		t.Errorf("expected string to have prefix %q, got:\n%s", prefix, s)
	}
}

// AssertHasSuffix fails if the string does not have the expected suffix.
func AssertHasSuffix(t *testing.T, s, suffix string) {
	t.Helper()
	if !strings.HasSuffix(s, suffix) {
		t.Errorf("expected string to have suffix %q, got:\n%s", suffix, s)
	}
}

// AssertSliceContains fails if the slice does not contain the element.
func AssertSliceContains[T comparable](t *testing.T, slice []T, element T) {
	t.Helper()
	for _, item := range slice {
		if item == element {
			return
		}
	}
	t.Errorf("expected slice to contain %v, got: %v", element, slice)
}

// AssertSliceNotContains fails if the slice contains the element.
func AssertSliceNotContains[T comparable](t *testing.T, slice []T, element T) {
	t.Helper()
	for _, item := range slice {
		if item == element {
			t.Errorf("expected slice NOT to contain %v, got: %v", element, slice)
			return
		}
	}
}

// AssertSliceLength fails if the slice does not have the expected length.
func AssertSliceLength[T any](t *testing.T, slice []T, expected int) {
	t.Helper()
	if len(slice) != expected {
		t.Errorf("expected slice length %d, got %d: %v", expected, len(slice), slice)
	}
}

// AssertMapHasKey fails if the map does not have the key.
func AssertMapHasKey[K comparable, V any](t *testing.T, m map[K]V, key K) {
	t.Helper()
	if _, ok := m[key]; !ok {
		t.Errorf("expected map to have key %v", key)
	}
}

// AssertMapNotHasKey fails if the map has the key.
func AssertMapNotHasKey[K comparable, V any](t *testing.T, m map[K]V, key K) {
	t.Helper()
	if _, ok := m[key]; ok {
		t.Errorf("expected map NOT to have key %v", key)
	}
}

// AssertError fails if err is nil.
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

// AssertNoError fails if err is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

// AssertErrorContains fails if err is nil or doesn't contain the substring.
func AssertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error containing %q, got nil", substring)
		return
	}
	if !strings.Contains(err.Error(), substring) {
		t.Errorf("expected error to contain %q, got: %v", substring, err)
	}
}
