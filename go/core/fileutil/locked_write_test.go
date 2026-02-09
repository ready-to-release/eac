package fileutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAtomicWriteJSONWithLock_Success(t *testing.T) {
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

	err := AtomicWriteJSONWithLock(path, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWriteJSONWithLock failed: %v", err)
	}

	// Verify file exists and has correct content
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

func TestAtomicWriteJSONWithLock_NoLockFileLeftBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	data := map[string]string{"test": "data"}

	err := AtomicWriteJSONWithLock(path, data, 0644)
	if err != nil {
		t.Fatalf("AtomicWriteJSONWithLock failed: %v", err)
	}

	// Verify lock file was cleaned up
	lockPath := path + ".lock"
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Errorf("lock file was not cleaned up: %s", lockPath)
	}
}

func TestAtomicWriteJSONWithLock_Concurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	// Number of concurrent writers (reduced for Windows compatibility)
	numWriters := 5
	var wg sync.WaitGroup
	var failures sync.Map

	// Each writer writes their ID
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			data := map[string]int{"writer_id": id}
			err := AtomicWriteJSONWithLock(path, data, 0644)
			if err != nil {
				// Store failures instead of calling t.Errorf from goroutine
				failures.Store(id, err)
			}
		}(i)
	}

	wg.Wait()

	// Check for failures (some may timeout, which is acceptable)
	failureCount := 0
	failures.Range(func(key, value interface{}) bool {
		failureCount++
		return true
	})

	// Allow some failures due to lock contention, but not all
	if failureCount == numWriters {
		t.Fatal("all writers failed")
	}

	// Verify file exists and contains valid JSON from one of the writers
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var result map[string]int
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify we got a valid writer ID
	writerID, ok := result["writer_id"]
	if !ok {
		t.Fatal("writer_id not found in result")
	}

	if writerID < 0 || writerID >= numWriters {
		t.Errorf("invalid writer_id: %d", writerID)
	}

	// The important thing is that the file is valid JSON and not corrupted
	// The lock ensures atomic writes happened correctly
}

func TestAtomicWriteJSONWithLock_PreventsConcurrentCorruption(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	// Create a struct that will produce different size JSON outputs
	type testData struct {
		ID      int    `json:"id"`
		Message string `json:"message"`
	}

	numWriters := 5
	var wg sync.WaitGroup
	var failures sync.Map

	// Each writer writes their unique message
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Create messages of varying sizes
			message := string(make([]byte, id*100))
			for j := range message {
				message = message[:j] + "x"
			}

			data := testData{
				ID:      id,
				Message: message,
			}

			// Write with a small delay to increase chance of collision
			time.Sleep(time.Millisecond * time.Duration(id))
			err := AtomicWriteJSONWithLock(path, data, 0644)
			if err != nil {
				// Store failures instead of calling t.Errorf from goroutine;
				// some lock contention failures are expected on Windows.
				failures.Store(id, err)
			}
		}(i)
	}

	wg.Wait()

	// Allow some failures due to lock contention, but not all
	failureCount := 0
	failures.Range(func(key, value interface{}) bool {
		failureCount++
		return true
	})
	if failureCount == numWriters {
		t.Fatal("all writers failed")
	}

	// Verify file contains valid, uncorrupted JSON
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var result testData
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("file is corrupted (invalid JSON): %v", err)
	}

	// The lock should have prevented corruption
	// We should have a valid result from one of the writers
	if result.ID < 0 || result.ID >= numWriters {
		t.Errorf("invalid ID in result: %d", result.ID)
	}
}

func TestAtomicWriteJSONWithLock_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	// channels cannot be marshaled to JSON
	data := make(chan int)

	err := AtomicWriteJSONWithLock(path, data, 0644)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	// Verify no files were created
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not exist after error")
	}
}
