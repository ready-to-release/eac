package orchestrator

import (
	"sort"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// emitResourceStatus sends the current lock/resource status to observers.
// This should be called after semaphore acquire/release to update the TUI's Resources pane.
func (us *UnitScheduler) emitResourceStatus() {
	if us.emitFunc == nil || us.registry == nil {
		return
	}

	snapshot := us.registry.Snapshot()
	resources := make([]core.ResourceInfo, 0, len(snapshot))
	for _, lock := range snapshot {
		// Derive pool from lock name (component-scheduler → host, docker-scheduler → docker)
		pool := ""
		switch lock.Name {
		case "component-scheduler":
			pool = "host"
		case "docker-scheduler":
			pool = "docker"
		}

		resources = append(resources, core.ResourceInfo{
			Name:     lock.Name,
			Type:     string(lock.Type),
			Pool:     pool,
			Capacity: int(lock.Capacity),
			Used:     int(lock.Used),
			Waiting:  int(lock.Waiting),
		})
	}

	us.emit(core.ResourceStatusEvent{
		Time:      time.Now(),
		Resources: resources,
	})
}

// emitToolStatus sends the current tool/container status to observers.
// This should be called after tool activation/deactivation to update the TUI's Tools and Containers lamps.
func (us *UnitScheduler) emitToolStatus() {
	if us.emitFunc == nil {
		return
	}

	running, total := us.getContainerInstanceCounts()
	sysRunning, sysTotal := us.getSystemInstanceCounts()

	us.emit(core.ToolStatusEvent{
		Time:                      time.Now(),
		ActiveContainerTools:      us.getActiveContainerToolsList(),
		UsedContainerTools:        us.getUsedContainerToolsList(),
		ActiveSystem:              us.getActiveSystemToolsList(),
		UsedSystem:                us.getUsedSystemToolsList(),
		ContainerInstancesRunning: running,
		ContainerInstancesTotal:   total,
		SystemInvocationsRunning:  sysRunning,
		SystemInvocationsTotal:    sysTotal,
	})
}

// SetCacheTimes sets the cache times for modules that are up-to-date.
// These times are displayed in the TUI for cached modules to show when
// the cached artifacts were originally built.
func (us *UnitScheduler) SetCacheTimes(times map[string]time.Time) {
	us.cacheTimesMu.Lock()
	defer us.cacheTimesMu.Unlock()
	us.cacheTimes = times
}

// getCacheTime returns the cache time for a module, or zero time if not cached.
func (us *UnitScheduler) getCacheTime(module string) time.Time {
	us.cacheTimesMu.RLock()
	defer us.cacheTimesMu.RUnlock()
	if us.cacheTimes == nil {
		return time.Time{}
	}
	return us.cacheTimes[module]
}

// addActiveTool increments the usage count for a tool/handler.
// Also tracks docker as active when container tools are used.
// moniker is used to track which container instance owns which lamp position.
func (us *UnitScheduler) addActiveTool(handler string, isContainer bool, moniker string) {
	if handler == "" {
		return
	}

	us.toolsMu.Lock()
	if isContainer {
		us.activeContainerTools[handler]++
		us.usedContainerTools[handler] = true
		// Container tools require docker - track it as active system tool
		us.activeSystem["docker"]++
		us.usedSystem["docker"] = true
		// Assign a new lamp position for this container instance
		lampIdx := len(us.containerLamps)
		us.containerLamps = append(us.containerLamps, true) // true = running (lit)
		us.containerLampMap[moniker] = lampIdx
	} else {
		us.activeSystem[handler]++
		us.usedSystem[handler] = true
		// Track system invocation lamp
		us.systemLamps = append(us.systemLamps, true) // true = running (lit)
	}
	us.toolsMu.Unlock()

	// Emit tool status update
	us.emitToolStatus()
}

// removeActiveTool decrements the usage count for a tool/handler.
// Also decrements docker count when container tools finish.
// moniker identifies which container lamp to turn off.
func (us *UnitScheduler) removeActiveTool(handler string, isContainer bool, moniker string) {
	if handler == "" {
		return
	}

	us.toolsMu.Lock()
	if isContainer {
		if us.activeContainerTools[handler] > 0 {
			us.activeContainerTools[handler]--
			if us.activeContainerTools[handler] == 0 {
				delete(us.activeContainerTools, handler)
			}
		}
		// Decrement docker usage count
		if us.activeSystem["docker"] > 0 {
			us.activeSystem["docker"]--
			if us.activeSystem["docker"] == 0 {
				delete(us.activeSystem, "docker")
			}
		}
		// Turn off this container's lamp (mark as completed/dim)
		if idx, ok := us.containerLampMap[moniker]; ok && idx < len(us.containerLamps) {
			us.containerLamps[idx] = false // false = completed (dim)
		}
	} else {
		if us.activeSystem[handler] > 0 {
			us.activeSystem[handler]--
			if us.activeSystem[handler] == 0 {
				delete(us.activeSystem, handler)
			}
		}
		// Turn off one running system lamp (last running → dim)
		for i := len(us.systemLamps) - 1; i >= 0; i-- {
			if us.systemLamps[i] {
				us.systemLamps[i] = false
				break
			}
		}
	}
	us.toolsMu.Unlock()

	// Emit tool status update
	us.emitToolStatus()
}

// getActiveContainerToolsList returns a sorted list of currently active container tools.
func (us *UnitScheduler) getActiveContainerToolsList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.activeContainerTools) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.activeContainerTools))
	for tool := range us.activeContainerTools {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getUsedContainerToolsList returns a sorted list of all container tools that have been used.
func (us *UnitScheduler) getUsedContainerToolsList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.usedContainerTools) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.usedContainerTools))
	for tool := range us.usedContainerTools {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getActiveSystemToolsList returns a sorted list of currently active system tools.
func (us *UnitScheduler) getActiveSystemToolsList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.activeSystem) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.activeSystem))
	for tool := range us.activeSystem {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getUsedSystemToolsList returns a sorted list of all system tools that have been used.
func (us *UnitScheduler) getUsedSystemToolsList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.usedSystem) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.usedSystem))
	for tool := range us.usedSystem {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getContainerInstanceCounts returns running and total container instance counts.
// Running = currently lit lamps, Total = all lamps (lit + dim).
func (us *UnitScheduler) getContainerInstanceCounts() (running, total int) {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	total = len(us.containerLamps)
	for _, active := range us.containerLamps {
		if active {
			running++
		}
	}
	return running, total
}

// getSystemInstanceCounts returns running and total system tool invocation counts.
// Running = currently lit lamps, Total = all lamps (lit + dim).
func (us *UnitScheduler) getSystemInstanceCounts() (running, total int) {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	total = len(us.systemLamps)
	for _, active := range us.systemLamps {
		if active {
			running++
		}
	}
	return running, total
}

// tuiMarkPending creates a component tab in pending state (scheduled, waiting for slot).
// moniker is the globally unique ID (Longname) for matching; displayName is for tab labels.
// Emits UnitQueuedEvent to observers.
// Skips emission if InitSummary was already sent (tabs are pre-registered by TUI).
func (us *UnitScheduler) tuiMarkPending(moniker, displayName string, weight int, tags workunit.TagSummary) {
	if us.initSummaryEmitted {
		return
	}
	// Emit event to observers
	us.emit(core.UnitQueuedEvent{
		Time:        time.Now(),
		ID:          moniker,
		DisplayName: displayName,
		Weight:      weight,
		Tags:        tags,
	})
}

// tuiMarkRunning marks a component as actively running (slot acquired).
// If the item was already marked complete by background cache detection, this is a no-op.
// moniker is the globally unique ID (Longname) for matching.
// Emits UnitStartedEvent and ProgressUpdateEvent to observers.
func (us *UnitScheduler) tuiMarkRunning(moniker string) {
	// Skip if already marked complete by background cache detection
	// This prevents blue tabs from flashing orange
	if _, wasEarlyCached := us.earlyCached.Load(moniker); wasEarlyCached {
		return
	}

	us.tuiMu.Lock()
	us.tuiRunning = append(us.tuiRunning, moniker)
	running := make([]string, len(us.tuiRunning))
	copy(running, us.tuiRunning)
	completed := us.tuiCompleted
	total := us.tuiTotal
	us.tuiMu.Unlock()

	// Emit events to observers
	us.emit(core.UnitStartedEvent{
		Time: time.Now(),
		ID:   moniker,
	})
	pressureTarget := 0
	if us.semaphore != nil {
		pressureTarget = us.semaphore.HostCapacity().Total
	}
	us.emit(core.ProgressUpdateEvent{
		Time:           time.Now(),
		Running:        running,
		Completed:      completed,
		Total:          total,
		Roof:           us.roof,
		PressureTarget: pressureTarget,
	})
}

// tuiMarkCompleted removes a component from running, increments completed, and reports exit code.
// Counter updates happen regardless of observer presence - workers are the sole source of truth.
// moniker is the globally unique ID (Longname) for matching.
// Emits UnitCompletedEvent and ProgressUpdateEvent to observers.
func (us *UnitScheduler) tuiMarkCompleted(moniker string, exitCode int) {
	// Update counters under lock (always)
	// Workers are the sole source of truth for completion counting.
	// Background cache detection marks items visually but does NOT count.
	us.tuiMu.Lock()
	// Remove from running list
	for i, m := range us.tuiRunning {
		if m == moniker {
			us.tuiRunning = append(us.tuiRunning[:i], us.tuiRunning[i+1:]...)
			break
		}
	}
	// Always increment counter - background detection no longer counts
	us.tuiCompleted++
	running := make([]string, len(us.tuiRunning))
	copy(running, us.tuiRunning)
	completed := us.tuiCompleted
	total := us.tuiTotal
	us.tuiMu.Unlock()

	// Emit events to observers
	us.emit(core.UnitCompletedEvent{
		Time:     time.Now(),
		ID:       moniker,
		ExitCode: exitCode,
	})
	pressureTarget := 0
	if us.semaphore != nil {
		pressureTarget = us.semaphore.HostCapacity().Total
	}
	us.emit(core.ProgressUpdateEvent{
		Time:           time.Now(),
		Running:        running,
		Completed:      completed,
		Total:          total,
		Roof:           us.roof,
		PressureTarget: pressureTarget,
	})
}
