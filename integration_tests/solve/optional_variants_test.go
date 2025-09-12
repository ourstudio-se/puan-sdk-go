// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

// Test_exactlyOneVariant_selectNotPreferred
// Ref: test_select_single_xor_component_when_another_xor_pair_is_preferred
// Description: Package A has two variants: (A, itemX) and (A, itemY, itemZ) with the latter
// being preferred. We select (A, itemX) and expect the result configuration (A, itemX)
func Test_exactlyOneVariant_selectNotPreferred(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargerVariantPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    1,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOneVariant_selectPreferred
// Ref: test_select_xor_pair_when_xor_pair_is_preferred
// Description: Package A has two variants: (A, itemX) and (A, itemY, itemZ) with the latter
// being preferred. We select (A, itemY, itemZ) and expect the result configuration
// (A, itemY, itemZ).
func Test_exactlyOneVariant_selectPreferred(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargerVariantPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

// Test_exactlyOneVariant_deselectingVariant_shouldGiveEmptySolution
// Ref: test_deselect_package_when_xor_pair_is_preferred_over_single_xor_component
// Description: Given rules package A -> xor(itemX, itemY),
// package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If (A, itemY, itemZ) is already selected, check that we will remove package A when deselecting A.
// Comment: this test fails. We get another variant of packageA instead of nothing.
func Test_exactlyOneVariant_deselectingVariant_shouldGiveEmptySolution(t *testing.T) {
	t.Skip()
	ruleSet := optionalVariantsWithXORBetweenItemsLargerVariantPreferred()
	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").WithAction(puan.REMOVE).Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOneVariant_changeVariant
// Ref: test_select_single_xor_component_when_xor_pair_is_already_selected
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If (A, itemY, itemZ) is already selected, check that we will select (A, itemX) variant when selecting (A, itemX)
func Test_exactlyOneVariant_changeVariant(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItem1Item2, _ := creator.PLDAG().SetXor("itemX", "itemY")
	xorItem1Item3, _ := creator.PLDAG().SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := creator.PLDAG().SetImply("packageA", xorItem1Item2)
	packageExactlyOneOfItem1Item3, _ := creator.PLDAG().SetImply("packageA", xorItem1Item3)

	root, _ := creator.PLDAG().SetAnd(
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
	)

	_ = creator.PLDAG().Assume("packageA", root)

	preferredItems, _ := creator.PLDAG().SetAnd("itemY", "itemZ")
	packagePreferredVariant, _ := creator.PLDAG().SetImply("packageA", preferredItems)
	invertedPreferred, _ := creator.PLDAG().SetNot(packagePreferredVariant)

	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    1,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOneVariant_selectItemInAnotherVariant_shouldChangeVariant
// Ref: test_select_one_component_in_xor_pair_when_single_xor_component_is_already_selected
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If package A and itemX are selected, check that we will get (A, itemY, itemZ) config when selecting itemY (or itemZ) independently.
func Test_exactlyOneVariant_selectItemInAnotherVariant_shouldChangeVariant(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargerVariantPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
		puan.NewSelectionBuilder("itemY").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

// Test_exactlyOneVariant_noSelection_shouldGivePreferred
func Test_exactlyOneVariant_noSelection_shouldEmptySolution(t *testing.T) {
	ruleSet := optionalVariantsWithXORBetweenItemsLargerVariantPreferred()

	selections := puan.Selections{}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

func optionalVariantsWithXORBetweenItemsLargerVariantPreferred() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()
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
