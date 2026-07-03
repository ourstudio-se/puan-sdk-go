// nolint:lll
package puan

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
)

func Test_SolutionQuery_validateSelections_givenEmptySelection_shouldReturnNoError(
	t *testing.T,
) {
	primaryID := fake.New[string]()
	subID := fake.New[string]()

	creator := NewRulesetCreator()
	_ = creator.AddPrimitives(primaryID, subID)
	ruleset, _ := creator.Create()

	selections := Selections{}

	query := NewSolutionQueryBuilder().
		WithSelections(selections).
		WithRuleset(ruleset).
		Build()

	err := query.validateSelections()

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
