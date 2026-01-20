package fileutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWrite_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	data := []byte("test data")

	err := AtomicWrite(path, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify file exists and has correct content
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("got %q, want %q", string(got), string(data))
	}
}

func TestAtomicWrite_NoTempFileLeftBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	data := []byte("test data")

	err := AtomicWrite(path, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	// Verify temp file was cleaned up
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file was not cleaned up: %s", tmpPath)
	}
}

func TestAtomicWrite_FailsGracefully(t *testing.T) {
	// Try to write to a path that can't exist
	path := filepath.Join("/nonexistent-directory-12345", "test.txt")
	data := []byte("test data")

	err := AtomicWrite(path, data, 0644)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAtomicWriteJSON_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := testStruct{
		Name:  "test",
		Value: 42,
	}

	err := AtomicWriteJSON(path, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWriteJSON failed: %v", err)
	}

	// Verify file exists and has correct JSON
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var got testStruct
	if err := json.Unmarshal(content, &got); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if got.Name != data.Name || got.Value != data.Value {
		t.Errorf("got %+v, want %+v", got, data)
	}
}

func TestAtomicWriteJSON_InvalidData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	// channels cannot be marshaled to JSON
	data := make(chan int)

	err := AtomicWriteJSON(path, data, 0644)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestAtomicWriteJSON_PrettyFormatted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	data := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	err := AtomicWriteJSON(path, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWriteJSON failed: %v", err)
	}

	// Verify file has pretty formatting (indentation)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Check for indentation (pretty print)
	expected := `{
  "name": "test",
  "value": 42
}`
	if string(content) != expected {
		t.Errorf("got:\n%s\nwant:\n%s", string(content), expected)
	}
}
