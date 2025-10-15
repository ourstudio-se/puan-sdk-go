package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
)

// nolint:lll
func Test_validateSelectionIDs_givenIndependentVariableInSubSelection_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(primaryID)
	ruleSet, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(selections, ruleSet)

	assert.Error(t, err)
}

func Test_validateSelectionIDs_givenInvalidSelection(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	invalidID := "invalid-id"
	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleSet, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(invalidID).Build(),
	}

	err := validateSelections(selections, ruleSet)

	assert.Error(t, err)
}

func Test_validateSelectionIDs_givenEmptySelection(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleSet, _ := creator.Create()

	selections := Selections{}

	err := validateSelections(selections, ruleSet)

	assert.NoError(t, err)
}
