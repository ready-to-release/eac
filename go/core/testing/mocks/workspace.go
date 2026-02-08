package mocks

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

// MockWorkspace implements core.WorkspacePort for testing.
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

// Root implements core.WorkspacePort.
func (m *MockWorkspace) Root() string {
	return m.root
}

// Source implements core.WorkspacePort.
func (m *MockWorkspace) Source() string {
	return m.source
}

// IsContainer implements core.WorkspacePort.
func (m *MockWorkspace) IsContainer() bool {
	return m.isContainer
}

// DistRoot implements core.WorkspacePort.
func (m *MockWorkspace) DistRoot() string {
	return m.distRoot
}

// Interface compliance check
var _ core.WorkspacePort = (*MockWorkspace)(nil)
