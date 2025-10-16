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
