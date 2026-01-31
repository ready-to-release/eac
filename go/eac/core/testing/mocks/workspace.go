package mocks

import "github.com/ready-to-release/eac/contracts/eac-core-interfaces"

// MockWorkspace implements interfaces.WorkspacePort for testing.
type MockWorkspace struct {
	root        string
	source      string
	isContainer bool
	distRoot    string
}

// NewMockWorkspace creates a new MockWorkspace with sensible defaults.
func NewMockWorkspace() *MockWorkspace {
	return &MockWorkspace{
		root:     "/mock/workspace",
		source:   "git",
		distRoot: "/mock/workspace",
	}
}

// WithRoot sets the workspace root path.
func (m *MockWorkspace) WithRoot(root string) *MockWorkspace {
	m.root = root
	return m
}

// WithSource sets the detection source.
func (m *MockWorkspace) WithSource(source string) *MockWorkspace {
	m.source = source
	return m
}

// WithIsContainer sets the container flag.
func (m *MockWorkspace) WithIsContainer(isContainer bool) *MockWorkspace {
	m.isContainer = isContainer
	return m
}

// WithDistRoot sets the distribution root.
func (m *MockWorkspace) WithDistRoot(distRoot string) *MockWorkspace {
	m.distRoot = distRoot
	return m
}

// Root implements interfaces.WorkspacePort.
func (m *MockWorkspace) Root() string {
	return m.root
}

// Source implements interfaces.WorkspacePort.
func (m *MockWorkspace) Source() string {
	return m.source
}

// IsContainer implements interfaces.WorkspacePort.
func (m *MockWorkspace) IsContainer() bool {
	return m.isContainer
}

// DistRoot implements interfaces.WorkspacePort.
func (m *MockWorkspace) DistRoot() string {
	return m.distRoot
}

// Interface compliance check
var _ interfaces.WorkspacePort = (*MockWorkspace)(nil)
