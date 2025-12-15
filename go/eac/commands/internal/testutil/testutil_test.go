// Package testutil provides shared test utilities for command testing.
package testutil

import (
	"fmt"
	"os"
	"testing"
)

func TestCapture(t *testing.T) {
	output := Capture(func() int {
		fmt.Println("stdout output")
		fmt.Fprintln(os.Stderr, "stderr output")
		return 0
	})

	output.AssertSuccess(t)

	if output.Stdout != "stdout output\n" {
		t.Errorf("expected stdout 'stdout output\\n', got %q", output.Stdout)
	}
	if output.Stderr != "stderr output\n" {
		t.Errorf("expected stderr 'stderr output\\n', got %q", output.Stderr)
	}
}

func TestCaptureFailure(t *testing.T) {
	output := Capture(func() int {
		fmt.Fprintln(os.Stderr, "error message")
		return 1
	})

	output.AssertFailure(t)
	output.AssertExitCode(t, 1)
}

func TestCaptureNoExit(t *testing.T) {
	output := CaptureNoExit(func() {
		fmt.Println("hello")
	})

	output.AssertSuccess(t)
	if output.Stdout != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output.Stdout)
	}
}

func TestAssertions(t *testing.T) {
	// Test string assertions
	AssertContains(t, "hello world", "world")
	AssertNotContains(t, "hello world", "foo")
	AssertContainsAll(t, "hello world", "hello", "world")
	AssertHasPrefix(t, "hello world", "hello")
	AssertHasSuffix(t, "hello world", "world")
	AssertNotEmpty(t, "hello")
	AssertMatches(t, "hello123", `hello\d+`)

	// Test slice assertions
	AssertSliceContains(t, []string{"a", "b", "c"}, "b")
	AssertSliceNotContains(t, []string{"a", "b", "c"}, "d")
	AssertSliceLength(t, []int{1, 2, 3}, 3)

	// Test map assertions
	m := map[string]int{"a": 1, "b": 2}
	AssertMapHasKey(t, m, "a")
	AssertMapNotHasKey(t, m, "c")

	// Test error assertions
	AssertNoError(t, nil)
	AssertError(t, fmt.Errorf("test error"))
	AssertErrorContains(t, fmt.Errorf("test error message"), "error")
}

func TestGitRepo(t *testing.T) {
	repo := NewGitRepo(t)

	// Verify repo was created
	if repo.Root() == "" {
		t.Fatal("expected non-empty repo root")
	}

	// Test file operations
	repo.WriteFile("test.txt", "hello")
	content := repo.ReadFile("test.txt")
	if content != "hello" {
		t.Errorf("expected 'hello', got %q", content)
	}

	// Test git operations
	repo.AddAll()
	repo.Commit("Initial commit")

	// Verify we're on a branch
	branch := repo.CurrentBranch()
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
}

func TestTempDir(t *testing.T) {
	dir := TempDir(t, "test-*")
	if dir == "" {
		t.Fatal("expected non-empty directory")
	}

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("temp dir should exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected directory")
	}
}

func TestWriteFixtureFile(t *testing.T) {
	dir := TempDir(t, "fixture-*")

	WriteFixtureFile(t, dir, "nested/path/file.txt", "content")

	content, err := os.ReadFile(dir + "/nested/path/file.txt")
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("expected 'content', got %q", string(content))
	}
}
