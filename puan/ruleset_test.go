package puan

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
)

func Test_RuleSet_copy_shouldBeEqual(t *testing.T) {
	aMatrix := fake.New[[][]int]()
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)
	dependentVariables := fake.New[[]string]()
	independentVariables := fake.New[[]string]()
	selectableVariables := fake.New[[]string]()
	preferredVariables := fake.New[[]string]()
	periodVariables := fake.New[[]TimeBoundVariable]()

	original := Ruleset{}
	original.polyhedron = polyhedron
	original.dependentVariables = dependentVariables
	original.selectableVariables = selectableVariables
	original.independentVariables = independentVariables
	original.preferredVariables = preferredVariables
	original.periodVariables = periodVariables
	ccopy := original.copy()

	assert.True(t, reflect.DeepEqual(original, ccopy))
}

func Test_RuleSet_copy_givenChangeToCopy_shouldNotChangeOriginal(t *testing.T) {
	aMatrix := fake.New[[][]int]()
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	original := Ruleset{}
	original.polyhedron = polyhedron

	ccopy := original.copy()
	ccopy.polyhedron.AddEmptyColumn()

	assert.NotEqual(t, ccopy.polyhedron, original.polyhedron)
}

func Test_RuleSet_copy_givenChangeToOriginal_shouldNotChangeCopy(t *testing.T) {
	aMatrix := fake.New[[][]int]()
	bVector := fake.New[[]int]()
	polyhedron := pldag.NewPolyhedron(aMatrix, bVector)

	original := Ruleset{}
	original.polyhedron = polyhedron

	ccopy := original.copy()
	original.polyhedron.AddEmptyColumn()

	assert.NotEqual(t, original.polyhedron, ccopy.polyhedron)
}

func Test_RuleSet_obtainSelectionID_givenStandaloneSelection_shouldReturnSelectionID(t *testing.T) {
	want := uuid.New().String()
	selection := NewSelectionBuilder(want).Build()

	ruleset := Ruleset{}
	got, err := ruleset.obtainQuerySelectionID(selection)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

// nolint:lll
func Test_RuleSet_setCompositeSelectionConstraint_givenConstraintDoesNotExist_shouldSetNewConstraint(
	t *testing.T,
) {
	primaryID := uuid.New().String()
	subID := uuid.New().String()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)

	// need to create a constraint to have both primaryID and subID
	// as dependent variables in the Ruleset otherwise the new constraint
	// cannot be created
	_, _ = creator.SetImply(primaryID, subID)
	ruleset, _ := creator.Create()

	selection := NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build()

	id, err := ruleset.setCompositeSelectionConstraint(selection.ids())

	assert.NoError(t, err)
	assert.Equal(t, id, ruleset.dependentVariables[4])
	assert.Len(t, ruleset.dependentVariables, 5)
	assert.Len(t, ruleset.polyhedron.B(), 6)
	assert.Len(t, ruleset.polyhedron.A(), 6)
	assert.Len(t, ruleset.polyhedron.A()[0], 5)
}

func Test_RuleSet_setCompositeSelectionConstraint_givenConstraintExists_shouldNotSetNewConstraint(
	t *testing.T,
) {
	primaryID := uuid.New().String()
	subID := uuid.New().String()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_, _ = creator.SetAnd(primaryID, subID)
	ruleset, _ := creator.Create()

	wantVariables := ruleset.dependentVariables
	wantPolyhedron := ruleset.polyhedron

	selection := NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build()

	_, err := ruleset.setCompositeSelectionConstraint(selection.ids())

	assert.NoError(t, err)
	assert.Equal(t, wantVariables, ruleset.dependentVariables)
	assert.Equal(t, wantPolyhedron, ruleset.polyhedron)
}

func Test_RuleSet_constraintExists_givenVariablesExists_shouldReturnTrue(
	t *testing.T,
) {
	constraint, _ := pldag.NewAtLeastConstraint([]string{uuid.New().String()}, 1)

	ruleset := Ruleset{}
	ruleset.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleset.dependentVariables = []string{constraint.ID()}

	got := ruleset.constraintExists(constraint)

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

func Test_newCompositeSelectionConstraint_shouldCreateConstraintWithoutDuplicates(
	t *testing.T,
) {
	primaryID := uuid.New().String()
	subID := "a"
	subID2 := "a"

	got, err := newCompositeSelectionConstraint([]string{primaryID, subID, subID2})

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
	coefficients := pldag.Coefficients{
		id1: value1,
		id2: value2,
	}

	ruleset := Ruleset{}
	ruleset.dependentVariables = []string{
		uuid.New().String(),
		id1,
		id2,
		uuid.New().String(),
	}

	got, err := ruleset.newRow(coefficients)

	assert.NoError(t, err)
	assert.Equal(t, []int{0, value1, value2, 0}, got)
}

func Test_RuleSet_setConstraint_shouldAddColumnOnExistingRows(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleset := Ruleset{}
	ruleset.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleset.dependentVariables = []string{primitiveID}

	err := ruleset.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Len(t, ruleset.polyhedron.A()[0], 2)
}

func Test_RuleSet_setConstraint_shouldAddConstraintIDToVariables(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleset := Ruleset{}
	ruleset.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleset.dependentVariables = []string{primitiveID}

	err := ruleset.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Equal(t, constraint.ID(), ruleset.dependentVariables[1])
}

func Test_RuleSet_setConstraint_shouldAddTwoRowsToPolyhedron(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleset := Ruleset{}
	ruleset.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleset.dependentVariables = []string{primitiveID}

	err := ruleset.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Len(t, ruleset.polyhedron.A(), 2)
}

func Test_RuleSet_setConstraint_shouldAddTwoBiases(t *testing.T) {
	primitiveID := uuid.New().String()
	constraint, _ := pldag.NewAtLeastConstraint([]string{primitiveID}, 1)

	ruleset := Ruleset{}
	ruleset.polyhedron = pldag.NewPolyhedron(nil, nil)
	ruleset.dependentVariables = []string{primitiveID}

	err := ruleset.setConstraint(constraint)

	assert.NoError(t, err)
	assert.Len(t, ruleset.polyhedron.B(), 2)
}

func Test_RuleSet_FindPeriodInSolution_givenSingleMatchingPeriod_shouldReturnPeriod(
	t *testing.T,
) {
	period1 := Period{
		from: newTestTime("2024-01-01T00:00:00Z"),
		to:   newTestTime("2024-01-31T00:00:00Z"),
	}
	period2 := Period{
		from: newTestTime("2024-02-01T00:00:00Z"),
		to:   newTestTime("2024-02-28T00:00:00Z"),
	}

	ruleset := &Ruleset{
		periodVariables: TimeBoundVariables{
			{variable: "period1", period: period1},
			{variable: "period2", period: period2},
		},
	}

	solution := Solution{
		"period1": 1,
		"period2": 0,
	}

	result, err := ruleset.FindPeriodInSolution(solution)

	assert.NoError(t, err)
	assert.Equal(t, period1, result)
}

func Test_RuleSet_FindPeriodInSolution_givenNoMatchingPeriod_shouldReturnError(
	t *testing.T,
) {
	period := Period{
		from: newTestTime("2024-01-01T00:00:00Z"),
		to:   newTestTime("2024-01-31T00:00:00Z"),
	}

	ruleset := &Ruleset{
		periodVariables: TimeBoundVariables{
			{variable: "period1", period: period},
		},
	}

	solution := Solution{
		"period1": 0,
	}

	_, err := ruleset.FindPeriodInSolution(solution)

	assert.Error(t, err)
}

func Test_RuleSet_FindPeriodInSolution_givenMultipleMatchingPeriods_shouldReturnError(
	t *testing.T,
) {
	period1 := Period{
		from: newTestTime("2024-01-01T00:00:00Z"),
		to:   newTestTime("2024-01-31T00:00:00Z"),
	}
	period2 := Period{
		from: newTestTime("2024-02-01T00:00:00Z"),
		to:   newTestTime("2024-02-28T00:00:00Z"),
	}

	ruleset := &Ruleset{
		periodVariables: TimeBoundVariables{
			{variable: "period1", period: period1},
			{variable: "period2", period: period2},
		},
	}

	solution := Solution{
		"period1": 1,
		"period2": 1,
	}

	_, err := ruleset.FindPeriodInSolution(solution)

	assert.Error(t, err)
}

func Test_validateVariables_givenNilVariables_shouldReturnError(t *testing.T) {
	err := validateVariables(nil, nil, nil, nil, nil)
	assert.Error(t, err)
}

func Test_validateVariables_givenEmptyVariables_shouldReturnError(t *testing.T) {
	err := validateVariables([]string{}, []string{}, []string{}, []string{}, []string{})
	assert.Error(t, err)
}

func Test_validateVariables_givenDependentAndIndependentOverlap_shouldReturnError(t *testing.T) {
	independent := fake.New[[]string]()
	dependent := append([]string{}, independent[0])

	err := validateVariables(nil, dependent, independent, nil, nil)
	assert.Error(t, err)
}

func Test_validateVariables_givenUniqueSelectable_shouldReturnError(t *testing.T) {
	independent := fake.New[[]string]()
	dependent := fake.New[[]string]()
	selectable := append([]string{}, uuid.New().String())

	err := validateVariables(selectable, dependent, independent, nil, nil)
	assert.Error(t, err)
}

func Test_validateVariables_givenUniquePreferred_shouldReturnError(t *testing.T) {
	independent := fake.New[[]string]()
	dependent := fake.New[[]string]()
	preferred := append([]string{}, uuid.New().String())

	err := validateVariables(nil, dependent, independent, preferred, nil)
	assert.Error(t, err)
}

func Test_validateVariables_givenUniquePeriod_shouldReturnError(t *testing.T) {
	independent := fake.New[[]string]()
	dependent := fake.New[[]string]()
	period := append([]string{}, uuid.New().String())

	err := validateVariables(nil, dependent, independent, nil, period)
	assert.Error(t, err)
}

func Test_RuleSet_isValidTime_givenNilTimestamp_shouldReturnTrue(t *testing.T) {
	ruleset := Ruleset{}

	got := ruleset.isValidTime(nil)

	assert.True(t, got)
}

func Test_RuleSet_isValidTime_givenNoTimeboundVariablesInRuleset_shouldReturnTrue(t *testing.T) {
	ruleset := Ruleset{}
	timestamp := fake.New[time.Time]()

	got := ruleset.isValidTime(&timestamp)

	assert.True(t, got)
}

func Test_RuleSet_isValidTime_givenTimeboundVariablesEndingAfterTimestamp_shouldReturnTrue(t *testing.T) {
	timestamp := fake.New[time.Time]()
	afterTimestamp := timestamp.Add(1 * time.Hour)
	timeboundVariables := TimeBoundVariable{
		period: Period{
			to: afterTimestamp,
		},
	}
	ruleset := Ruleset{
		periodVariables: TimeBoundVariables{
			timeboundVariables,
		},
	}

	got := ruleset.isValidTime(&timestamp)

	assert.True(t, got)
}

func Test_RuleSet_isValidTime_givenTimeboundVariablesEndingAtTimestamp_shouldReturnTrue(t *testing.T) {
	timestamp := fake.New[time.Time]()
	timeboundVariables := TimeBoundVariable{
		period: Period{
			to: timestamp,
		},
	}
	ruleset := Ruleset{
		periodVariables: TimeBoundVariables{
			timeboundVariables,
		},
	}

	got := ruleset.isValidTime(&timestamp)

	assert.True(t, got)
}
func Test_RuleSet_isValidTime_givenOnlyTimeboundVariablesBeforeTimestamp_shouldReturnFalse(t *testing.T) {
	timestamp := fake.New[time.Time]()
	beforeTimestamp := timestamp.Add(-1 * time.Hour)
	timeboundVariables := TimeBoundVariable{
		period: Period{
			to: beforeTimestamp,
		},
	}
	ruleset := Ruleset{
		periodVariables: TimeBoundVariables{
			timeboundVariables,
		},
	}

	got := ruleset.isValidTime(&timestamp)

	assert.False(t, got)
}
