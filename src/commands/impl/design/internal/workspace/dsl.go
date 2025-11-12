package workspace

import (
	"fmt"
	"strings"
)

// GenerateBaseDSL generates the base Structurizr DSL for a new workspace
func GenerateBaseDSL(name, description string) string {
	return fmt.Sprintf(`workspace "%s" "%s" {

    model {
        # Define your software systems, containers, and components here

        system = softwareSystem "System" "Main software system" {
            # Containers will be added here
        }

        # Define relationships here
    }

    views {
        systemContext system "SystemContext" {
            include *
            autoLayout
        }

        container system "Containers" {
            include *
            autoLayout
        }

        styles {
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
        }
    }

}
`, name, description)
}

// SanitizeID converts a name to a valid DSL identifier
// - Converts to lowercase
// - Replaces spaces and hyphens with underscores
// - Removes non-alphanumeric characters (except underscores)
func SanitizeID(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "-", "_")

	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
