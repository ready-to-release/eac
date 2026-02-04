// Package scheduling provides work unit scheduling with dependency resolution.
//
// The core abstraction is [WorkScheduler], which provides pull-based scheduling
// with LPT (Longest Processing Time First) ordering. Workers pull items from
// the scheduler; the scheduler evaluates readiness dynamically based on
// dependency completion.
//
// # Separation of Concerns
//
// This package owns "what to run next" based on:
//   - Dependencies: Only items with all dependencies complete are ready
//   - LPT ordering: Among ready items, heaviest (longest) is returned first
//
// The orchestrator (clibase/orchestrator) owns "when to run" via:
//   - Capacity gating: Semaphores control concurrent execution slots
//   - Resource management: Memory, CPU, Docker pool constraints
//
// # In-Flight Tracking
//
// In-flight tracking is implicit in the design:
//   - [WorkScheduler.Next] or [WorkScheduler.WaitForReady] removes item from queue
//   - Item is now "in-flight" until [WorkScheduler.MarkComplete] or [WorkScheduler.MarkFailed]
//   - Dependencies only unblock when MarkComplete is called
//
// # Thread Safety
//
// All [WorkScheduler] implementations are safe for concurrent use by multiple
// worker goroutines.
//
// # Example Usage
//
//	scheduler, err := scheduling.NewDependencyScheduler(workUnits)
//	if err != nil {
//	    // Handle circular dependency error
//	}
//	defer scheduler.Close()
//
//	// Worker loop
//	for {
//	    spec := scheduler.WaitForReady()
//	    if spec == nil {
//	        break // Queue exhausted
//	    }
//
//	    // Execute work...
//	    if success {
//	        scheduler.MarkComplete(spec.ID)
//	    } else {
//	        scheduler.MarkFailed(spec.ID)
//	    }
//	}
package scheduling
