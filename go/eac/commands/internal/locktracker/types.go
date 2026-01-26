// Package locktracker provides lock visualization by tracking all active synchronization primitives.
package locktracker

import "time"

// LockType represents the type of synchronization primitive.
type LockType string

// LockType constants.
const (
	LockTypeMutex     LockType = "mutex"
	LockTypeRWMutex   LockType = "rwmutex"
	LockTypeSemaphore LockType = "semaphore"
	LockTypeFileLock  LockType = "filelock"
	LockTypeWeighted  LockType = "weighted"
)

// EventType represents the type of lock event.
type EventType string

// EventType constants.
const (
	EventAcquired        EventType = "acquired"
	EventReleased        EventType = "released"
	EventWaiting         EventType = "waiting"
	EventCapacityChanged EventType = "capacity_changed"
)

// LockInfo contains information about a tracked lock.
type LockInfo struct {
	ID         string
	Type       LockType
	Name       string
	AcquiredAt time.Time
	Holder     string
	Capacity   int64
	Used       int64
	Waiting    int64
}

// LockEvent represents an event related to a lock.
type LockEvent struct {
	Type      EventType
	Lock      LockInfo
	Timestamp time.Time
}

// LockSummary contains aggregated statistics about all tracked locks.
type LockSummary struct {
	Total         int
	ByType        map[LockType]int
	TotalCapacity int64
	TotalUsed     int64
	TotalWaiting  int64
}
