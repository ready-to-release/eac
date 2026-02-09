package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFactoryRepositoryDefaults_Singleton(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	cfg1, err := getFactoryRepositoryDefaults()
	require.NoError(t, err)
	require.NotNil(t, cfg1)

	cfg2, err := getFactoryRepositoryDefaults()
	require.NoError(t, err)

	// Same pointer: singleton behavior
	assert.Same(t, cfg1, cfg2, "getFactoryRepositoryDefaults should return the same pointer")
}

func TestCloneRepositoryDefaults_DeepCopy(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	clone1, err := cloneRepositoryDefaults()
	require.NoError(t, err)
	require.NotNil(t, clone1)

	clone2, err := cloneRepositoryDefaults()
	require.NoError(t, err)
	require.NotNil(t, clone2)

	// Different pointers: deep copy
	assert.NotSame(t, clone1, clone2, "cloneRepositoryDefaults should return different pointers")

	// Same content
	assert.Equal(t, clone1, clone2, "clones should have equal content")
}

func TestCloneRepositoryDefaults_MutationSafe(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	factory, err := getFactoryRepositoryDefaults()
	require.NoError(t, err)
	originalType := factory.Repository.Type

	// Mutate a clone
	clone, err := cloneRepositoryDefaults()
	require.NoError(t, err)
	clone.Repository.Type = "mutated-type"

	// Factory should be unchanged
	factory2, err := getFactoryRepositoryDefaults()
	require.NoError(t, err)
	assert.Equal(t, originalType, factory2.Repository.Type, "factory default should not be mutated")
}

func TestGetFactoryEnvironmentsDefaults_Singleton(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	cfg1, err := getFactoryEnvironmentsDefaults()
	require.NoError(t, err)
	require.NotNil(t, cfg1)

	cfg2, err := getFactoryEnvironmentsDefaults()
	require.NoError(t, err)

	assert.Same(t, cfg1, cfg2)
}

func TestGetFactoryTestingTagsDefaults_Singleton(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	cfg1, err := getFactoryTestingTagsDefaults()
	require.NoError(t, err)
	require.NotNil(t, cfg1)

	cfg2, err := getFactoryTestingTagsDefaults()
	require.NoError(t, err)

	assert.Same(t, cfg1, cfg2)
}

func TestGetFactoryTimeoutsDefaults_Singleton(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	cfg1, err := getFactoryTimeoutsDefaults()
	require.NoError(t, err)
	require.NotNil(t, cfg1)

	cfg2, err := getFactoryTimeoutsDefaults()
	require.NoError(t, err)

	assert.Same(t, cfg1, cfg2)
}

func TestGetFactoryTestSuitesDefaults_Singleton(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	cfg1, err := getFactoryTestSuitesDefaults()
	require.NoError(t, err)
	require.NotNil(t, cfg1)

	cfg2, err := getFactoryTestSuitesDefaults()
	require.NoError(t, err)

	assert.Same(t, cfg1, cfg2)
}

func TestCloneTestSuitesDefaults_HasSuiteMap(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	clone, err := cloneTestSuitesDefaults()
	require.NoError(t, err)
	require.NotNil(t, clone)

	// Verify buildSuiteMap was called on the clone
	if len(clone.Suites) > 0 {
		first := clone.Suites[0]
		got := clone.Get(first.Moniker)
		assert.NotNil(t, got, "suite map should be populated after cloning")
	}
}

func TestGetFactoryBlueprintsDefaults_Singleton(t *testing.T) {
	t.Cleanup(ResetFactoryDefaultsForTesting)

	cfg1, err := getFactoryBlueprintsDefaults()
	require.NoError(t, err)
	require.NotNil(t, cfg1)

	cfg2, err := getFactoryBlueprintsDefaults()
	require.NoError(t, err)

	assert.Same(t, cfg1, cfg2)
}

func TestResetFactoryDefaultsForTesting_AllowsReInit(t *testing.T) {
	cfg1, err := getFactoryRepositoryDefaults()
	require.NoError(t, err)

	ResetFactoryDefaultsForTesting()

	cfg2, err := getFactoryRepositoryDefaults()
	require.NoError(t, err)

	// After reset, a new pointer should be returned (re-parsed)
	assert.NotSame(t, cfg1, cfg2, "after reset, singleton should be re-initialized")
	assert.Equal(t, cfg1, cfg2, "content should be identical after re-init")
}

