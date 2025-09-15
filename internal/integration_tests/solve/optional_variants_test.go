// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	puan2 "github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_optionalVariant_selectNotPreferred
// Ref: test_select_single_xor_component_when_another_xor_pair_is_preferred
// Description: Package A has two variants: (A, itemX) and (A, itemY, itemZ) with the latter
// being preferred. We select (A, itemX) and expect the result configuration (A, itemX)
func Test_optionalVariant_selectNotPreferred(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargeVariantPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
			"packageA": 1,
			"itemX":    1,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_optionalVariant_selectPreferred
// Ref: test_select_xor_pair_when_xor_pair_is_preferred
// Description: Package A has two variants: (A, itemX) and (A, itemY, itemZ) with the latter
// being preferred. We select (A, itemY, itemZ) and expect the result configuration
// (A, itemY, itemZ).
func Test_optionalVariant_selectPreferred(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargeVariantPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

// Test_optionalVariant_deselectingVariant_shouldGiveEmptySolution
// Ref: test_deselect_package_when_xor_pair_is_preferred_over_single_xor_component
// Description: Given rules package A -> xor(itemX, itemY),
// package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If (A, itemY, itemZ) is already selected, check that we will remove package A when deselecting A.
// Comment: this test fails. We get another variant of packageA instead of nothing.
func Test_optionalVariant_deselectingVariant_shouldGiveEmptySolution(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargeVariantPreferred()
	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
		puan2.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").WithAction(puan2.REMOVE).Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_optionalVariant_changeVariant
// Ref: test_select_single_xor_component_when_xor_pair_is_already_selected
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If (A, itemY, itemZ) is already selected, check that we will select (A, itemX) variant when selecting (A, itemX)
func Test_optionalVariant_changeVariant(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargeVariantPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
		puan2.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
			"packageA": 1,
			"itemX":    1,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_optionalVariant_selectItemInAnotherVariant_shouldChangeVariant
// Ref: test_select_one_component_in_xor_pair_when_single_xor_component_is_already_selected
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If package A and itemX are selected, check that we will get (A, itemY, itemZ) config when selecting itemY (or itemZ) independently.
func Test_optionalVariant_selectItemInAnotherVariant_shouldChangeVariant(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargeVariantPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
		puan2.NewSelectionBuilder("itemY").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

// Test_optionalVariant_noSelection_shouldGiveEmptySolution
func Test_optionalVariant_noSelection_shouldGiveEmptySolution(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargeVariantPreferred()

	selections := puan2.Selections{}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

func optionalVariantsWithXORBetweenItemsLargeVariantPreferred() *puan2.RuleSet {
	creator := puan2.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItem1Item2, _ := creator.PLDAG().SetXor("itemX", "itemY")
	xorItem1Item3, _ := creator.PLDAG().SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := creator.PLDAG().SetImply("packageA", xorItem1Item2)
	packageExactlyOneOfItem1Item3, _ := creator.PLDAG().SetImply("packageA", xorItem1Item3)

	root, _ := creator.PLDAG().SetAnd(
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
	)

	_ = creator.PLDAG().Assume(root)

	preferredItems, _ := creator.PLDAG().SetAnd("itemY", "itemZ")
	packagePreferredVariant, _ := creator.PLDAG().SetImply("packageA", preferredItems)
	invertedPreferred, _ := creator.PLDAG().SetNot(packagePreferredVariant)

	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

	return ruleSet
}
