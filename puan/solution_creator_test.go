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
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(selections, ruleset)

	assert.Error(t, err)
}

func Test_validateSelections_givenIndependentVariableSelectionWithSubSelection_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(subID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(selections, ruleset)

	assert.Error(t, err)
}

func Test_validateSelections_givenNotExistingID_shouldReturnError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	invalidID := fake.New[string]()
	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(invalidID).Build(),
	}

	err := validateSelections(selections, ruleset)

	assert.Error(t, err)
}

func Test_validateSelections_givenEmptySelection_shouldReturnNoError(t *testing.T) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRuleSetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleset, _ := creator.Create()

	selections := Selections{}

	err := validateSelections(selections, ruleset)

	assert.NoError(t, err)
}

func Test_categorizeSelections(t *testing.T) {
	independentID := fake.New[string]()
	dependentID := fake.New[string]()

	selections := Selections{
		NewSelectionBuilder(independentID).Build(),
		NewSelectionBuilder(dependentID).Build(),
	}

	independentVariables := []string{independentID}

	dependentSelections, independentSelections :=
		categorizeSelections(selections, independentVariables)

	assert.Len(t, dependentSelections, 1)
	assert.Equal(t, dependentID, dependentSelections[0].id)

	assert.Len(t, independentSelections, 1)
	assert.Equal(t, independentID, independentSelections[0].id)
}
