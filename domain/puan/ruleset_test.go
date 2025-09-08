package puan

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/fake"
	"github.com/stretchr/testify/assert"
)

func Test_RuleSet_copy_shouldBeEqual(t *testing.T) {
	aMatrix := fake.New[[][]int]()
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	variables := fake.New[[]string]()
	primitiveVariables := fake.New[[]string]()
	preferredVariables := fake.New[[]string]()

	original := &RuleSet{}
	original.polyhedron = polyhedron
	original.variables = variables
	original.primitiveVariables = primitiveVariables
	original.preferredVariables = preferredVariables

	copy := original.copy()

	assert.Equal(t, polyhedron, copy.polyhedron)
	assert.Equal(t, variables, copy.variables)
	assert.Equal(t, primitiveVariables, copy.primitiveVariables)
	assert.Equal(t, preferredVariables, copy.preferredVariables)
}

func Test_RuleSet_copy_givenChangeToCopy_shouldNotChangeOriginal(t *testing.T) {
	aMatrix := fake.New[[][]int]()
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	original := &RuleSet{}
	original.polyhedron = polyhedron

	copy := original.copy()
	copy.polyhedron.AddEmptyColumn()

	assert.NotEqual(t, copy.polyhedron, original.polyhedron)
}

func Test_RuleSet_copy_givenChangeToOriginal_shouldNotChangeCopy(t *testing.T) {
	aMatrix := fake.New[[][]int]()
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	original := &RuleSet{}
	original.polyhedron = polyhedron

	copy := original.copy()
	original.polyhedron.AddEmptyColumn()

	assert.NotEqual(t, original.polyhedron, copy.polyhedron)
}

func Test_RuleSet_obtainSelectionID_givenStandaloneSelection_shouldReturnSelectionID(t *testing.T) {
	want := faker.Word()
	selection := NewSelectionBuilder(want).Build()

	ruleSet := &RuleSet{}
	got, err := ruleSet.obtainSelectionID(selection)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_RuleSet_obtainSelectionID_givenCompositeSelection_shouldNotUseSelectionIDs(t *testing.T) {
	selection := NewSelectionBuilder(faker.Word()).WithSubSelectionID(faker.Word()).Build()

	polyhedron := pldag.NewPolyhedron([][]int{{1, 1}}, []int{2})
	ruleSet := &RuleSet{}
	ruleSet.polyhedron = polyhedron

	id, err := ruleSet.obtainSelectionID(selection)

	assert.NoError(t, err)
	assert.NotEqual(t, selection.ID(), id)
	assert.NotEqual(t, selection.subSelectionID, id)
}
