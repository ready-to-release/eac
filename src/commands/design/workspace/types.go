package workspace

// Workspace represents a Structurizr workspace
type Workspace struct {
	Module      string
	Name        string
	Description string
	Path        string
}

// Container represents a container in the architecture
type Container struct {
	ID          string
	Name        string
	Technology  string
	Description string
}

// Relationship represents a relationship between elements
type Relationship struct {
	Source      string
	Destination string
	Description string
	Technology  string
}
