package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func Test_givenNoFreeSelection_shouldGiveZeroValueFreeVariables(t *testing.T) {
	ruleset := rulesetWithIndependentPrimitives()

	selections := puan.Selections{}

	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":            1,
			"itemY":            0,
			"itemZ":            0,
			"independentItem1": 0,
			"independentItem2": 0,
		},
		solution,
	)
}

func Test_givenOneFreeSelection_shouldGiveSelectedFreeVariable(t *testing.T) {
	ruleset := rulesetWithIndependentPrimitives()

	selections := puan.Selections{
		puan.NewSelectionBuilder("independentItem1").Build(),
	}

	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":            1,
			"itemY":            0,
			"itemZ":            0,
			"independentItem1": 1,
			"independentItem2": 0,
		},
		solution,
	)
}

func Test_givenAllFreeSelection_shouldGiveAllSelectedFreeVariable(t *testing.T) {
	ruleset := rulesetWithIndependentPrimitives()

	selections := puan.Selections{
		puan.NewSelectionBuilder("independentItem1").Build(),
		puan.NewSelectionBuilder("independentItem2").Build(),
	}

	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":            1,
			"itemY":            0,
			"itemZ":            0,
			"independentItem1": 1,
			"independentItem2": 1,
		},
		solution,
	)
}

func Test_givenAllFreeSelectionAndDependantSelection_shouldGiveAllSelectedVariables(t *testing.T) {
	ruleset := rulesetWithIndependentPrimitives()

	selections := puan.Selections{
		puan.NewSelectionBuilder("independentItem1").Build(),
		puan.NewSelectionBuilder("independentItem2").Build(),
		puan.NewSelectionBuilder("itemZ").Build(),
	}

	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":            0,
			"itemY":            0,
			"itemZ":            1,
			"independentItem1": 1,
			"independentItem2": 1,
		},
		solution,
	)
}

func rulesetWithIndependentPrimitives() *puan.Ruleset {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("itemX", "itemY", "itemZ", "independentItem1", "independentItem2")
	exactlyOnePackage, _ := creator.SetXor("itemX", "itemY", "itemZ")

	_ = creator.Assume(
		exactlyOnePackage,
	)

	_ = creator.Prefer("itemX")

	ruleset, _ := creator.Create()

	return ruleset
}
