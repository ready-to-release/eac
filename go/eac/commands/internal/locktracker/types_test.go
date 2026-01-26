//go:build L1
// +build L1

package locktracker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// LockType Tests
// =============================================================================

// TestLockType_Constants verifies that all LockType constants are defined and distinct.
func TestLockType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		lockType LockType
		wantStr  string
	}{
		{
			name:     "mutex type",
			lockType: LockTypeMutex,
			wantStr:  "mutex",
		},
		{
			name:     "rwmutex type",
			lockType: LockTypeRWMutex,
			wantStr:  "rwmutex",
		},
		{
			name:     "semaphore type",
			lockType: LockTypeSemaphore,
			wantStr:  "semaphore",
		},
		{
			name:     "filelock type",
			lockType: LockTypeFileLock,
			wantStr:  "filelock",
		},
		{
			name:     "weighted type",
			lockType: LockTypeWeighted,
			wantStr:  "weighted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStr, string(tt.lockType), "LockType constant should have expected string value")
		})
	}
}

// TestLockType_Distinctness verifies all LockType constants have unique values.
func TestLockType_Distinctness(t *testing.T) {
	lockTypes := []LockType{
		LockTypeMutex,
		LockTypeRWMutex,
		LockTypeSemaphore,
		LockTypeFileLock,
		LockTypeWeighted,
	}

	seen := make(map[LockType]bool)
	for _, lt := range lockTypes {
		if seen[lt] {
			t.Errorf("LockType %q is not unique", lt)
		}
		seen[lt] = true
	}

	assert.Equal(t, 5, len(seen), "Expected 5 distinct LockType values")
}

// =============================================================================
// EventType Tests
// =============================================================================

// TestEventType_Constants verifies that all EventType constants are defined and distinct.
func TestEventType_Constants(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		wantStr   string
	}{
		{
			name:      "acquired event",
			eventType: EventAcquired,
			wantStr:   "acquired",
		},
		{
			name:      "released event",
			eventType: EventReleased,
			wantStr:   "released",
		},
		{
			name:      "waiting event",
			eventType: EventWaiting,
			wantStr:   "waiting",
		},
		{
			name:      "capacity changed event",
			eventType: EventCapacityChanged,
			wantStr:   "capacity_changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantStr, string(tt.eventType), "EventType constant should have expected string value")
		})
	}
}

// TestEventType_Distinctness verifies all EventType constants have unique values.
func TestEventType_Distinctness(t *testing.T) {
	eventTypes := []EventType{
		EventAcquired,
		EventReleased,
		EventWaiting,
		EventCapacityChanged,
	}

	seen := make(map[EventType]bool)
	for _, et := range eventTypes {
		if seen[et] {
			t.Errorf("EventType %q is not unique", et)
		}
		seen[et] = true
	}

	assert.Equal(t, 4, len(seen), "Expected 4 distinct EventType values")
}

// =============================================================================
// LockInfo Tests
// =============================================================================

// TestLockInfo_Fields verifies that LockInfo has all required fields.
func TestLockInfo_Fields(t *testing.T) {
	now := time.Now()

	info := LockInfo{
		ID:         "lock-123",
		Type:       LockTypeSemaphore,
		Name:       "test-semaphore",
		AcquiredAt: now,
		Holder:     "goroutine-42",
		Capacity:   10,
		Used:       3,
		Waiting:    2,
	}

	assert.Equal(t, "lock-123", info.ID, "ID field should be set")
	assert.Equal(t, LockTypeSemaphore, info.Type, "Type field should be set")
	assert.Equal(t, "test-semaphore", info.Name, "Name field should be set")
	assert.Equal(t, now, info.AcquiredAt, "AcquiredAt field should be set")
	assert.Equal(t, "goroutine-42", info.Holder, "Holder field should be set")
	assert.Equal(t, int64(10), info.Capacity, "Capacity field should be set")
	assert.Equal(t, int64(3), info.Used, "Used field should be set")
	assert.Equal(t, int64(2), info.Waiting, "Waiting field should be set")
}

// TestLockInfo_ZeroValue verifies the zero value of LockInfo is valid.
func TestLockInfo_ZeroValue(t *testing.T) {
	var info LockInfo

	assert.Equal(t, "", info.ID, "ID zero value should be empty string")
	assert.Equal(t, LockType(""), info.Type, "Type zero value should be empty")
	assert.Equal(t, "", info.Name, "Name zero value should be empty string")
	assert.True(t, info.AcquiredAt.IsZero(), "AcquiredAt zero value should be zero time")
	assert.Equal(t, "", info.Holder, "Holder zero value should be empty string")
	assert.Equal(t, int64(0), info.Capacity, "Capacity zero value should be 0")
	assert.Equal(t, int64(0), info.Used, "Used zero value should be 0")
	assert.Equal(t, int64(0), info.Waiting, "Waiting zero value should be 0")
}

// TestLockInfo_MutexType verifies LockInfo works correctly for mutex type locks.
func TestLockInfo_MutexType(t *testing.T) {
	info := LockInfo{
		ID:         "mutex-1",
		Type:       LockTypeMutex,
		Name:       "config-mutex",
		AcquiredAt: time.Now(),
		Holder:     "main-goroutine",
		Capacity:   1, // Mutex has capacity of 1
		Used:       1, // Either 0 or 1
		Waiting:    0,
	}

	assert.Equal(t, LockTypeMutex, info.Type)
	assert.Equal(t, int64(1), info.Capacity, "Mutex should have capacity of 1")
	assert.LessOrEqual(t, info.Used, info.Capacity, "Used should not exceed capacity")
}

// TestLockInfo_SemaphoreType verifies LockInfo works correctly for semaphore type locks.
func TestLockInfo_SemaphoreType(t *testing.T) {
	info := LockInfo{
		ID:         "sem-1",
		Type:       LockTypeSemaphore,
		Name:       "worker-pool",
		AcquiredAt: time.Now(),
		Holder:     "",
		Capacity:   5,
		Used:       3,
		Waiting:    2,
	}

	assert.Equal(t, LockTypeSemaphore, info.Type)
	assert.Equal(t, int64(5), info.Capacity)
	assert.Equal(t, int64(3), info.Used)
	assert.Equal(t, int64(2), info.Waiting)
	assert.LessOrEqual(t, info.Used, info.Capacity, "Used should not exceed capacity")
}

// TestLockInfo_Copy verifies that LockInfo can be copied without side effects.
func TestLockInfo_Copy(t *testing.T) {
	original := LockInfo{
		ID:         "lock-1",
		Type:       LockTypeSemaphore,
		Name:       "test",
		AcquiredAt: time.Now(),
		Holder:     "holder-1",
		Capacity:   10,
		Used:       5,
		Waiting:    3,
	}

	// Copy by value
	copied := original

	// Modify the copy
	copied.ID = "lock-2"
	copied.Used = 7
	copied.Waiting = 1

	// Verify original is unchanged
	assert.Equal(t, "lock-1", original.ID, "Original ID should be unchanged")
	assert.Equal(t, int64(5), original.Used, "Original Used should be unchanged")
	assert.Equal(t, int64(3), original.Waiting, "Original Waiting should be unchanged")

	// Verify copy is modified
	assert.Equal(t, "lock-2", copied.ID, "Copy ID should be modified")
	assert.Equal(t, int64(7), copied.Used, "Copy Used should be modified")
	assert.Equal(t, int64(1), copied.Waiting, "Copy Waiting should be modified")
}

// =============================================================================
// LockEvent Tests
// =============================================================================

// TestLockEvent_Fields verifies that LockEvent has all required fields.
func TestLockEvent_Fields(t *testing.T) {
	now := time.Now()
	lockInfo := LockInfo{
		ID:   "lock-1",
		Type: LockTypeMutex,
		Name: "test-mutex",
	}

	event := LockEvent{
		Type:      EventAcquired,
		Lock:      lockInfo,
		Timestamp: now,
	}

	assert.Equal(t, EventAcquired, event.Type, "Type field should be set")
	assert.Equal(t, lockInfo, event.Lock, "Lock field should be set")
	assert.Equal(t, now, event.Timestamp, "Timestamp field should be set")
}

// TestLockEvent_AllEventTypes verifies LockEvent works with all event types.
func TestLockEvent_AllEventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
	}{
		{"acquired", EventAcquired},
		{"released", EventReleased},
		{"waiting", EventWaiting},
		{"capacity_changed", EventCapacityChanged},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := LockEvent{
				Type: tt.eventType,
				Lock: LockInfo{
					ID:   "test-lock",
					Type: LockTypeSemaphore,
				},
				Timestamp: time.Now(),
			}

			assert.Equal(t, tt.eventType, event.Type)
			assert.False(t, event.Timestamp.IsZero())
		})
	}
}

// TestLockEvent_ZeroValue verifies the zero value of LockEvent is valid.
func TestLockEvent_ZeroValue(t *testing.T) {
	var event LockEvent

	assert.Equal(t, EventType(""), event.Type, "Type zero value should be empty")
	assert.Equal(t, LockInfo{}, event.Lock, "Lock zero value should be empty LockInfo")
	assert.True(t, event.Timestamp.IsZero(), "Timestamp zero value should be zero time")
}

// =============================================================================
// LockSummary Tests
// =============================================================================

// TestLockSummary_Fields verifies that LockSummary has all required fields.
func TestLockSummary_Fields(t *testing.T) {
	summary := LockSummary{
		Total: 10,
		ByType: map[LockType]int{
			LockTypeMutex:     3,
			LockTypeSemaphore: 5,
			LockTypeFileLock:  2,
		},
		TotalCapacity: 50,
		TotalUsed:     25,
		TotalWaiting:  5,
	}

	assert.Equal(t, 10, summary.Total, "Total field should be set")
	assert.NotNil(t, summary.ByType, "ByType field should not be nil")
	assert.Equal(t, 3, summary.ByType[LockTypeMutex], "ByType[mutex] should be 3")
	assert.Equal(t, 5, summary.ByType[LockTypeSemaphore], "ByType[semaphore] should be 5")
	assert.Equal(t, 2, summary.ByType[LockTypeFileLock], "ByType[filelock] should be 2")
	assert.Equal(t, int64(50), summary.TotalCapacity, "TotalCapacity field should be set")
	assert.Equal(t, int64(25), summary.TotalUsed, "TotalUsed field should be set")
	assert.Equal(t, int64(5), summary.TotalWaiting, "TotalWaiting field should be set")
}

// TestLockSummary_ZeroValue verifies the zero value of LockSummary is valid.
func TestLockSummary_ZeroValue(t *testing.T) {
	var summary LockSummary

	assert.Equal(t, 0, summary.Total, "Total zero value should be 0")
	assert.Nil(t, summary.ByType, "ByType zero value should be nil")
	assert.Equal(t, int64(0), summary.TotalCapacity, "TotalCapacity zero value should be 0")
	assert.Equal(t, int64(0), summary.TotalUsed, "TotalUsed zero value should be 0")
	assert.Equal(t, int64(0), summary.TotalWaiting, "TotalWaiting zero value should be 0")
}

// TestLockSummary_EmptyByType verifies LockSummary works with empty ByType map.
func TestLockSummary_EmptyByType(t *testing.T) {
	summary := LockSummary{
		Total:  0,
		ByType: make(map[LockType]int),
	}

	assert.Equal(t, 0, summary.Total)
	assert.NotNil(t, summary.ByType)
	assert.Equal(t, 0, len(summary.ByType))
}

// TestLockSummary_AllLockTypes verifies LockSummary can track all lock types.
func TestLockSummary_AllLockTypes(t *testing.T) {
	summary := LockSummary{
		Total: 15,
		ByType: map[LockType]int{
			LockTypeMutex:     2,
			LockTypeRWMutex:   3,
			LockTypeSemaphore: 5,
			LockTypeFileLock:  3,
			LockTypeWeighted:  2,
		},
		TotalCapacity: 100,
		TotalUsed:     40,
		TotalWaiting:  10,
	}

	assert.Equal(t, 15, summary.Total)
	assert.Equal(t, 5, len(summary.ByType), "Should have 5 lock types")

	// Verify total matches sum of ByType
	sum := 0
	for _, count := range summary.ByType {
		sum += count
	}
	assert.Equal(t, summary.Total, sum, "Total should equal sum of ByType counts")
}

// TestLockSummary_Copy verifies that LockSummary can be safely copied.
func TestLockSummary_Copy(t *testing.T) {
	original := LockSummary{
		Total: 5,
		ByType: map[LockType]int{
			LockTypeMutex:     2,
			LockTypeSemaphore: 3,
		},
		TotalCapacity: 20,
		TotalUsed:     10,
		TotalWaiting:  2,
	}

	// Copy by value (note: map is a reference type)
	copied := original
	copied.ByType = make(map[LockType]int)
	for k, v := range original.ByType {
		copied.ByType[k] = v
	}

	// Modify copy
	copied.Total = 10
	copied.ByType[LockTypeFileLock] = 5
	copied.TotalUsed = 15

	// Verify original is unchanged
	assert.Equal(t, 5, original.Total, "Original Total should be unchanged")
	assert.Equal(t, 2, len(original.ByType), "Original ByType should have 2 entries")
	assert.Equal(t, int64(10), original.TotalUsed, "Original TotalUsed should be unchanged")

	// Verify copy is modified
	assert.Equal(t, 10, copied.Total, "Copy Total should be modified")
	assert.Equal(t, 3, len(copied.ByType), "Copy ByType should have 3 entries")
	assert.Equal(t, int64(15), copied.TotalUsed, "Copy TotalUsed should be modified")
}

// TestLockSummary_CapacityUtilization verifies capacity utilization calculations.
func TestLockSummary_CapacityUtilization(t *testing.T) {
	tests := []struct {
		name            string
		capacity        int64
		used            int64
		wantUtilization float64
	}{
		{
			name:            "no capacity",
			capacity:        0,
			used:            0,
			wantUtilization: 0.0,
		},
		{
			name:            "empty utilization",
			capacity:        100,
			used:            0,
			wantUtilization: 0.0,
		},
		{
			name:            "half utilized",
			capacity:        100,
			used:            50,
			wantUtilization: 0.5,
		},
		{
			name:            "fully utilized",
			capacity:        100,
			used:            100,
			wantUtilization: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := LockSummary{
				TotalCapacity: tt.capacity,
				TotalUsed:     tt.used,
			}

			var utilization float64
			if summary.TotalCapacity > 0 {
				utilization = float64(summary.TotalUsed) / float64(summary.TotalCapacity)
			}

			assert.InDelta(t, tt.wantUtilization, utilization, 0.001)
		})
	}
}
