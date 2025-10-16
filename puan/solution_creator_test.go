// nolint:lll
package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
)

func Test_validateSelections_givenIndependentVariableInSubSelection_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(primaryID)
	ruleSet, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(selections, *ruleSet)

	assert.Error(t, err)
}

func Test_validateSelections_givenIndependentVariableSelectionWithSubSelection_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(subID)
	ruleSet, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(selections, *ruleSet)

	assert.Error(t, err)
}

func Test_validateSelections_givenNotExistingID_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	invalidID := fake.New[string]()
	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleSet, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(invalidID).Build(),
	}

	err := validateSelections(selections, *ruleSet)

	assert.Error(t, err)
}

func Test_validateSelections_givenEmptySelection_shouldReturnNoError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleSet, _ := creator.Create()

	selections := Selections{}

	err := validateSelections(selections, *ruleSet)

	assert.NoError(t, err)
}

func Test_extractIndependentSelection_noIndependentSelection_shouldReturnNil(t *testing.T) {
	independentVariableIDs := fake.New[[]string]()
	selections := Selections{
		NewSelectionBuilder(fake.New[string]()).Build(),
	}

	selection := extractIndependentSelection(selections, independentVariableIDs[0])
	assert.Nil(t, selection)
}

func Test_extractIndependentSelection_independentSelection_shouldReturnSelection(t *testing.T) {
	id := fake.New[string]()
	independentVariableIDs := []string{id}
	selections := Selections{
		{id: id, action: ADD},
	}

	actual := extractIndependentSelection(selections, independentVariableIDs[0])
	expected := Selection{id: id, action: ADD}

	assert.Equal(t, expected, *actual)
}

func Test_extractIndependentSelection_independentSelection_shouldReturnLastActionSelection(t *testing.T) {
	id := fake.New[string]()
	independentVariableIDs := []string{id}
	selections := Selections{
		{id: id, action: ADD},
		{id: id, action: REMOVE},
	}

	actual := extractIndependentSelection(selections, independentVariableIDs[0])
	expected := Selection{
		id:     id,
		action: REMOVE,
	}

	assert.Equal(t, expected, *actual)
}
