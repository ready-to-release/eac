//go:build L0 && ov

package schema_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/domain/schema"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getRepoRoot is a test helper that returns the repository root
func getRepoRoot(t *testing.T) string {
	t.Helper()
	repoRoot, err := repository.GetRepositoryRoot("")
	require.NoError(t, err)
	return repoRoot
}

func TestNewValidator(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)
	assert.NotNil(t, v)
}

func TestValidator_ValidateRepositoryYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Load the actual repository.yml from the repository
	repoPath := filepath.Join(repoRoot, ".eac", "repository.yml")
	data, err := os.ReadFile(repoPath)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaRepository, data)
	assert.NoError(t, err, "repository.yml should be valid against schema")
}

func TestValidator_ValidateEnvironmentsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Load the actual environments.yml from the repository
	envPath := filepath.Join(repoRoot, ".eac", "environments.yml")
	data, err := os.ReadFile(envPath)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaEnvironments, data)
	assert.NoError(t, err, "environments.yml should be valid against schema")
}

func TestValidator_ValidateTestingTagsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Load the actual testing-tags.yml from the repository
	tagsPath := filepath.Join(repoRoot, ".eac", "testing-tags.yml")
	data, err := os.ReadFile(tagsPath)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaTestingTags, data)
	assert.NoError(t, err, "testing-tags.yml should be valid against schema")
}

func TestValidator_InvalidRepositoryYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Missing required 'modules' field
	invalidYAML := []byte(`
something_else:
  - name: test
`)

	err = v.ValidateYAML(schema.SchemaRepository, invalidYAML)
	assert.Error(t, err, "should fail validation for missing 'modules' field")

	var validErr *schema.ValidationError
	assert.ErrorAs(t, err, &validErr)
	assert.Equal(t, schema.SchemaRepository, validErr.SchemaType)
}

func TestValidator_InvalidEnvironmentsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Missing required 'environments' field
	invalidYAML := []byte(`
something_else:
  - moniker: test
`)

	err = v.ValidateYAML(schema.SchemaEnvironments, invalidYAML)
	assert.Error(t, err, "should fail validation for missing 'environments' field")

	var validErr *schema.ValidationError
	assert.ErrorAs(t, err, &validErr)
	assert.Equal(t, schema.SchemaEnvironments, validErr.SchemaType)
}

func TestValidator_InvalidEnvironmentLevel(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Invalid level value
	invalidYAML := []byte(`
environments:
  - moniker: test
    name: Test Environment
    level: INVALID
    type: unit
`)

	err = v.ValidateYAML(schema.SchemaEnvironments, invalidYAML)
	assert.Error(t, err, "should fail validation for invalid level")
}

func TestValidator_InvalidModuleMoniker(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Invalid moniker (uppercase not allowed)
	invalidYAML := []byte(`
modules:
  - moniker: INVALID_UPPER
    name: Test Module
`)

	err = v.ValidateYAML(schema.SchemaRepository, invalidYAML)
	assert.Error(t, err, "should fail validation for invalid moniker pattern")
}

func TestValidator_ValidMinimalRepositoryYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Minimal valid repository config with modules
	validYAML := []byte(`
modules:
  - moniker: test-module
    name: Test Module
`)

	err = v.ValidateYAML(schema.SchemaRepository, validYAML)
	assert.NoError(t, err, "minimal valid repository config should pass")
}

func TestValidator_ValidMinimalEnvironmentsYAML(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Minimal valid environments config (metadata is optional)
	validYAML := []byte(`
environments:
  - moniker: test-env
    name: Test Environment
    level: L0
    type: unit
`)

	err = v.ValidateYAML(schema.SchemaEnvironments, validYAML)
	assert.NoError(t, err, "minimal valid environments config should pass")
}

func TestValidator_UnknownSchemaType(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	err = v.ValidateYAML(schema.SchemaType("unknown"), []byte(`test: value`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown schema type")
}

func TestValidator_InvalidYAMLSyntax(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	// Invalid YAML syntax
	invalidYAML := []byte(`
modules:
  - moniker: test
    name: [unclosed bracket
`)

	err = v.ValidateYAML(schema.SchemaRepository, invalidYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestValidator_ValidateJSON(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	validJSON := []byte(`{
		"modules": [
			{
				"moniker": "test-module",
				"name": "Test Module"
			}
		]
	}`)

	err = v.ValidateJSON(schema.SchemaRepository, validJSON)
	assert.NoError(t, err)
}

func TestValidator_InvalidJSON(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	invalidJSON := []byte(`{ invalid json }`)

	err = v.ValidateJSON(schema.SchemaRepository, invalidJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestGetSchemaTypes(t *testing.T) {
	types := schema.GetSchemaTypes()
	assert.Len(t, types, 10)
	assert.Contains(t, types, schema.SchemaComponentTypes)
	assert.Contains(t, types, schema.SchemaEnvironments)
	assert.Contains(t, types, schema.SchemaTestingTags)
	assert.Contains(t, types, schema.SchemaTestSuites)
	assert.Contains(t, types, schema.SchemaRepository)
	assert.Contains(t, types, schema.SchemaEACConfig)
	assert.Contains(t, types, schema.SchemaBooks)
	assert.Contains(t, types, schema.SchemaToolConfig)
}

func TestValidator_GetSchemaPath(t *testing.T) {
	repoRoot := getRepoRoot(t)
	v, err := schema.NewValidator(repoRoot)
	require.NoError(t, err)

	schemaPath := v.GetSchemaPath()
	assert.Contains(t, schemaPath, "contracts")
	assert.Contains(t, schemaPath, "core")
	assert.Contains(t, schemaPath, schema.ContractVersion)
}
