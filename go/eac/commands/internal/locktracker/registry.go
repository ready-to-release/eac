package locktracker

import (
	"sync"
	"time"
)

// Registry tracks all active locks.
type Registry struct {
	mu          sync.RWMutex
	locks       map[string]LockInfo
	subscribers []chan LockEvent
	subMu       sync.RWMutex
}

var (
	globalRegistry     *Registry
	globalRegistryOnce sync.Once
)

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		locks:       make(map[string]LockInfo),
		subscribers: make([]chan LockEvent, 0),
	}
}

// Get returns the global registry singleton.
func Get() *Registry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}

// Register adds a lock to the registry and triggers EventAcquired notification.
func (r *Registry) Register(info LockInfo) {
	r.mu.Lock()
	r.locks[info.ID] = info
	r.mu.Unlock()

	r.notify(LockEvent{
		Type:      EventAcquired,
		Lock:      info,
		Timestamp: time.Now(),
	})
}

// Unregister removes a lock from the registry and triggers EventReleased notification.
func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	info, exists := r.locks[id]
	if exists {
		delete(r.locks, id)
	}
	r.mu.Unlock()

	if exists {
		r.notify(LockEvent{
			Type:      EventReleased,
			Lock:      info,
			Timestamp: time.Now(),
		})
	}
}

// UpdateSemaphore updates the used and waiting counts for a semaphore and triggers EventCapacityChanged.
func (r *Registry) UpdateSemaphore(id string, used, waiting int64) {
	r.mu.Lock()
	info, exists := r.locks[id]
	if exists {
		info.Used = used
		info.Waiting = waiting
		r.locks[id] = info
	}
	r.mu.Unlock()

	if exists {
		r.notify(LockEvent{
			Type:      EventCapacityChanged,
			Lock:      info,
			Timestamp: time.Now(),
		})
	}
}

// Snapshot returns a copy of all lock information.
func (r *Registry) Snapshot() map[string]LockInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]LockInfo, len(r.locks))
	for k, v := range r.locks {
		result[k] = v
	}
	return result
}

// Summary returns aggregated statistics about all locks.
func (r *Registry) Summary() LockSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary := LockSummary{
		Total:  len(r.locks),
		ByType: make(map[LockType]int),
	}

	for _, lock := range r.locks {
		summary.ByType[lock.Type]++
		summary.TotalCapacity += lock.Capacity
		summary.TotalUsed += lock.Used
		summary.TotalWaiting += lock.Waiting
	}

	return summary
}

// Subscribe registers a channel to receive lock events. Returns an unsubscribe function.
func (r *Registry) Subscribe(ch chan LockEvent) func() {
	r.subMu.Lock()
	r.subscribers = append(r.subscribers, ch)
	r.subMu.Unlock()

	return func() {
		r.subMu.Lock()
		defer r.subMu.Unlock()

		for i, sub := range r.subscribers {
			if sub == ch {
				r.subscribers = append(r.subscribers[:i], r.subscribers[i+1:]...)
				break
			}
		}
	}
}

// notify sends an event to all subscribers.
func (r *Registry) notify(event LockEvent) {
	r.subMu.RLock()
	defer r.subMu.RUnlock()

	for _, ch := range r.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}
