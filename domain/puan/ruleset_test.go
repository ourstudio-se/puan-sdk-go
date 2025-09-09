package puan

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/fake"
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
	want := uuid.New().String()
	selection := NewSelectionBuilder(want).Build()

	ruleSet := &RuleSet{}
	got, err := ruleSet.obtainSelectionID(selection)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// nolint:lll
func Test_RuleSet_setCompositeSelectionConstraint_givenConstraintDoesNotExist_shouldSetNewConstraint(
	t *testing.T,
) {
	primaryID := uuid.New().String()
	subID := uuid.New().String()

	creator := NewRuleSetCreator()
	creator.PLDAG().SetPrimitives(primaryID, subID)
	ruleSet := creator.Create()

	id, err := ruleSet.setCompositeSelectionConstraint([]string{primaryID, subID})

	assert.NoError(t, err)
	assert.Equal(t, id, ruleSet.variables[2])
	assert.Len(t, ruleSet.variables, 3)
	assert.Len(t, ruleSet.polyhedron.B(), 2)
	assert.Len(t, ruleSet.polyhedron.A(), 2)
	assert.Len(t, ruleSet.polyhedron.A()[0], 3)
}

func Test_RuleSet_setCompositeSelectionConstraint_givenConstraintExists_shouldNotSetNewConstraint(
	t *testing.T,
) {
	primaryID := uuid.New().String()
	subID := uuid.New().String()

	creator := NewRuleSetCreator()
	creator.PLDAG().SetPrimitives(primaryID, subID)
	_, _ = creator.PLDAG().SetAnd(primaryID, subID)
	ruleSet := creator.Create()

	wantVariables := ruleSet.variables
	wantPolyhedron := ruleSet.polyhedron

	_, err := ruleSet.setCompositeSelectionConstraint([]string{primaryID, subID})

	assert.NoError(t, err)
	assert.Equal(t, wantVariables, ruleSet.variables)
	assert.Equal(t, wantPolyhedron, ruleSet.polyhedron)
}

func Test_RuleSet_constraintExists_givenVariablesExists_shouldReturnTrue(
	t *testing.T,
) {
	constraint, _ := pldag.NewAtLeastConstraint([]string{uuid.New().String()}, 1)

	ruleSet := &RuleSet{}
	ruleSet.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleSet.variables = []string{constraint.ID()}

	got := ruleSet.constraintExists(constraint)

	assert.True(t, got)
}

func Test_newCompositeSelectionConstraint_shouldCreateConstraint(
	t *testing.T,
) {
	primaryID := uuid.New().String()
	subID := uuid.New().String()

	got, err := newCompositeSelectionConstraint([]string{primaryID, subID})

	want, _ := pldag.NewAtLeastConstraint([]string{primaryID, subID}, 2)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_RuleSet_newRow(
	t *testing.T,
) {
	id1 := uuid.New().String()
	id2 := uuid.New().String()
	value1 := fake.New[int]()
	value2 := fake.New[int]()
	coefficients := pldag.CoefficientValues{
		id1: value1,
		id2: value2,
	}

	ruleSet := &RuleSet{}
	ruleSet.variables = []string{
		uuid.New().String(),
		id1,
		id2,
		uuid.New().String(),
	}

	got, err := ruleSet.newRow(coefficients)

	assert.NoError(t, err)
	assert.Equal(t, []int{0, value1, value2, 0}, got)
}

func Test_RuleSet_setConstraint_shouldAddColumnOnExistingRows(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleSet := &RuleSet{}
	ruleSet.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleSet.variables = []string{primitiveID}

	err := ruleSet.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Len(t, ruleSet.polyhedron.A()[0], 2)
}

func Test_RuleSet_setConstraint_shouldAddConstraintIDToVariables(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleSet := &RuleSet{}
	ruleSet.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleSet.variables = []string{primitiveID}

	err := ruleSet.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Equal(t, constraint.ID(), ruleSet.variables[1])
}

func Test_RuleSet_setConstraint_shouldAddTwoRowsToPolyhedron(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleSet := &RuleSet{}
	ruleSet.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleSet.variables = []string{primitiveID}

	err := ruleSet.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Len(t, ruleSet.polyhedron.A(), 2)
}

func Test_RuleSet_setConstraint_shouldAddTwoBiases(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleSet := &RuleSet{}
	ruleSet.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleSet.variables = []string{primitiveID}

	err := ruleSet.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Len(t, ruleSet.polyhedron.B(), 2)
}
