package scheduling

import (
	"fmt"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// DuplicateUnitIDError indicates duplicate Longname values in the work slice.
type DuplicateUnitIDError struct {
	FirstIndex  int
	SecondIndex int
	Longname    string
}

func (e *DuplicateUnitIDError) Error() string {
	return fmt.Sprintf("duplicate UnitID at indices %d and %d: %s", e.FirstIndex, e.SecondIndex, e.Longname)
}

// CircularDependencyError indicates a cycle in the dependency graph.
type CircularDependencyError struct {
	Units []string
}

func (e *CircularDependencyError) Error() string {
	return "circular dependency detected among units: " + joinStrings(e.Units)
}

func joinStrings(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	result := "["
	for i, str := range s {
		if i > 0 {
			result += ", "
		}
		result += str
	}
	return result + "]"
}

// validateNoCycles checks for circular dependencies using Kahn's algorithm.
// Returns CircularDependencyError if cycles are detected.
func validateNoCycles(work []workunit.UnitSpec) error {
	if len(work) == 0 {
		return nil
	}

	// Build in-degree map and unit set
	unitSet := make(map[string]bool)
	inDegree := make(map[string]int)
	processedSet := make(map[string]bool) // Track processed units explicitly

	for _, w := range work {
		key := w.ID.Longname()
		unitSet[key] = true
		inDegree[key] = 0
	}

	// Count in-degrees (only for deps within the work set)
	for _, w := range work {
		key := w.ID.Longname()
		for _, dep := range w.DependsOn {
			depKey := dep.Longname()
			if unitSet[depKey] {
				inDegree[key]++
			}
		}
	}

	// Process nodes with in-degree 0
	var queue []string
	for key, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, key)
		}
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		processedSet[node] = true // Mark as processed

		// Find units that depend on this node and decrement their in-degree
		for _, w := range work {
			for _, dep := range w.DependsOn {
				if dep.Longname() == node {
					key := w.ID.Longname()
					inDegree[key]--
					if inDegree[key] == 0 {
						queue = append(queue, key)
					}
				}
			}
		}
	}

	// Compare against unique unit count (unitSet), not work slice length
	// This handles cases where work contains duplicate Longname() values
	if len(processedSet) < len(unitSet) {
		// Circular dependency detected - find ALL unprocessed units
		var remaining []string
		for key := range unitSet {
			if !processedSet[key] {
				remaining = append(remaining, key)
			}
		}
		return &CircularDependencyError{Units: remaining}
	}

	return nil
}

// validateNoDuplicates checks for duplicate Longname values in the work slice.
// Returns DuplicateUnitIDError if duplicates are found, with the indices of the first pair.
func validateNoDuplicates(work []workunit.UnitSpec) error {
	if len(work) == 0 {
		return nil
	}

	seen := make(map[string]int) // Longname -> first index
	for i, w := range work {
		key := w.ID.Longname()
		if firstIdx, exists := seen[key]; exists {
			return &DuplicateUnitIDError{
				FirstIndex:  firstIdx,
				SecondIndex: i,
				Longname:    key,
			}
		}
		seen[key] = i
	}
	return nil
}
