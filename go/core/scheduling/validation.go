package scheduling

import "github.com/ready-to-release/eac/go/core/workunit"

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
	// Build in-degree map
	unitSet := make(map[string]bool)
	inDegree := make(map[string]int)

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

	processed := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		processed++

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

	if processed < len(work) {
		// Circular dependency detected - find the cycle members
		var remaining []string
		for key, deg := range inDegree {
			if deg > 0 {
				remaining = append(remaining, key)
			}
		}
		return &CircularDependencyError{Units: remaining}
	}

	return nil
}
