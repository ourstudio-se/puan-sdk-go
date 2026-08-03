// nolint:lll
package puan

import (
	"testing"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puanerror"
	"github.com/stretchr/testify/assert"
)

func Test_validateRuleset_givenMissingRuleset_shouldReturnError(
	t *testing.T,
) {
	ruleset := Ruleset{}

	err := validateRuleset(ruleset)

	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_validateRuleset_givenValidRuleset_shouldReturnNoError(
	t *testing.T,
) {
	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(fake.New[string]())
	ruleset, _ := creator.Create()

	err := validateRuleset(ruleset)

	assert.NoError(t, err)
}

func Test_validateTimestamps_givenFromAfterTo_shouldReturnError(
	t *testing.T,
) {
	from := newTestTime("2024-01-02T00:00:00Z")
	to := newTestTime("2024-01-01T00:00:00Z")

	err := validateTimestamps(&from, &to)

	assert.ErrorIs(t, err, puanerror.InvalidArgument)
}

func Test_validateTimestamps_givenFromBeforeTo_shouldReturnNoError(
	t *testing.T,
) {
	from := newTestTime("2024-01-01T00:00:00Z")
	to := newTestTime("2024-01-02T00:00:00Z")

	err := validateTimestamps(&from, &to)

	assert.NoError(t, err)
}

func Test_validateSelections_givenIndependentVariableInSubSelection_shouldReturnError(
	t *testing.T,
) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(primaryID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(ruleset, selections)

	assert.Error(t, err)
}

func Test_validateSelections_givenIndependentVariableSelectionWithSubSelection_shouldReturnError(
	t *testing.T,
) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	_ = creator.Assume(subID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(primaryID).WithSubSelectionID(subID).Build(),
	}

	err := validateSelections(ruleset, selections)

	assert.Error(t, err)
}

func Test_validateSelections_givenNotExistingID_shouldReturnError(
	t *testing.T,
) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	invalidID := fake.New[string]()
	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleset, _ := creator.Create()

	selections := Selections{
		NewSelectionBuilder(invalidID).Build(),
	}

	err := validateSelections(ruleset, selections)

	assert.Error(t, err)
}
func Test_validateSelections_givenEmptySelection_shouldReturnNoError(
	t *testing.T,
) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleset, _ := creator.Create()

	selections := Selections{}

	err := validateSelections(ruleset, selections)

	assert.NoError(t, err)
}
