// Command: design list
// Description: List available modules with architecture documentation
package design

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ready-to-release/eac/src/commands/impl/design/internal"
	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register("design list", DesignList)
}

// DesignList lists available modules with architecture documentation
func DesignList() int {
	client, err := design.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	modules, err := client.ListModules()
	if err != nil {
		fmt.Printf("❌ Failed to list modules: %v\n", err)
		return 1
	}

	if len(modules) == 0 {
		fmt.Println("ℹ️  No modules with architecture documentation found")
		fmt.Println("\nExpected location: docs/reference/design/<module>/workspace.dsl")
		return 0
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Println("Available modules with architecture documentation:")
	fmt.Println()
	fmt.Fprintln(w, "MODULE\tSTATUS\tVIEWS\tDOCS\tDECISIONS\tPATH")
	fmt.Fprintln(w, "──────\t──────\t─────\t────\t─────────\t────")

	for _, module := range modules {
		docsCount := "-"
		if module.HasDocs {
			docsCount = fmt.Sprintf("%d", module.DocCount)
		}

		decisionsCount := "-"
		if module.HasDecisions {
			decisionsCount = fmt.Sprintf("%d", module.DecisionCount)
		}

		viewsCount := fmt.Sprintf("%d", module.ViewCount)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			module.Name,
			module.GetStatus(),
			viewsCount,
			docsCount,
			decisionsCount,
			module.Path,
		)
	}

	w.Flush()

	fmt.Println("\n💡 To view documentation:")
	fmt.Println("  go run . design serve <module>")
	return 0
}
