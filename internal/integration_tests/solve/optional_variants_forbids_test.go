//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_optionalVariantsWithForbids_shouldReturnPreferred
// Ref: test_will_change_package_variant_when_package_is_preselected_with_component_requiring_package
// Description: Following rules are applied (with preferreds on the left xor-component)
// itemA -> packageX
// itemA -> itemB
// itemA -> ~itemC
// itemA -> ~itemD
// itemC -> ~itemA
// itemB -> xor(itemC, itemA)
// packageX -> xor(itemC, itemA)
// packageX -> xor(itemD, itemB)
// Our case is that itemA is already selected, which indirectly will add
// package X with its preferred components itemC and itemD
// Then we select (X, itemC, itemD) and we expect itemA to be removed from solution.
func Test_optionalVariantsWithForbids_shouldReturnPreferred(t *testing.T) {
	ruleset := optionalVariantsWithForbids()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemA").Build(),
		puan.NewSelectionBuilder("packageX").WithSubSelectionID("itemC").WithSubSelectionID("itemD").Build(),
	}

	query, _ := ruleset.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageX": 1,
			"itemA":    0,
			"itemB":    0,
			"itemC":    1,
			"itemD":    1,
		},
		primitiveSolution,
	)
}

// Test_optionalVariantsWithForbids_shouldReturnNOTPreferred
// Ref: test_will_change_package_variant_when_single_component_is_preselected
// Description: Following rules are applied (with preferreds on the left xor-component)
// itemA -> packageX
// itemA -> itemB
// itemA -> ~itemC
// itemA -> ~itemD
// itemC -> ~itemA
// itemB -> xor(itemC, itemA)
// packageX -> xor(itemC, itemA)
// packageX -> xor(itemD, itemB)
// Our case is that itemC is already selected.
// Then we select (X, itemC, itemB) and we expect itemC to be removed from solution.
func Test_optionalVariantsWithForbids_shouldReturnNOTPreferred(t *testing.T) {
	ruleset := optionalVariantsWithForbids()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemC").Build(),
		puan.NewSelectionBuilder("packageX").WithSubSelectionID("itemA").WithSubSelectionID("itemB").Build(),
	}

	query, _ := ruleset.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageX": 1,
			"itemA":    1,
			"itemB":    1,
			"itemC":    0,
			"itemD":    0,
		},
		primitiveSolution,
	)
}

func optionalVariantsWithForbids() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("itemA", "itemB", "itemC", "itemD", "packageX")

	reversedItemA, _ := creator.PLDAG().SetImply("itemA", "packageX")

	exactlyOneOfItemCAndA, _ := creator.PLDAG().SetXor("itemC", "itemA")
	exactlyOneOfItemCAndAInX, _ := creator.PLDAG().SetImply("packageX", exactlyOneOfItemCAndA)

	exactlyOneOfItemDAndB, _ := creator.PLDAG().SetXor("itemD", "itemB")
	exactlyOneOfItemDAndBInX, _ := creator.PLDAG().SetImply("packageX", exactlyOneOfItemDAndB)

	notItemC, _ := creator.PLDAG().SetNot("itemC")
	itemAForbidsItemC, _ := creator.PLDAG().SetImply("itemA", notItemC)

	exactlyOneOfItemCAndAWithB, _ := creator.PLDAG().SetImply("itemB", exactlyOneOfItemCAndA)

	itemARequiresItemB, _ := creator.PLDAG().SetImply("itemA", "itemB")

	notItemD, _ := creator.PLDAG().SetNot("itemD")
	itemAForbidsItemD, _ := creator.PLDAG().SetImply("itemA", notItemD)

	root, _ := creator.PLDAG().SetAnd(
		reversedItemA,
		exactlyOneOfItemCAndAInX,
		exactlyOneOfItemDAndBInX,
		itemAForbidsItemC,
		exactlyOneOfItemCAndAWithB,
		itemARequiresItemB,
		itemAForbidsItemD,
	)

	_ = creator.PLDAG().Assume(root)

	preferredPackageXItemC, _ := creator.PLDAG().SetImply("packageX", "itemC")
	preferredPackageXItemD, _ := creator.PLDAG().SetImply("packageX", "itemD")

	_ = creator.SetPreferreds(preferredPackageXItemC, preferredPackageXItemD)

	ruleSet := creator.Create()

	return ruleSet
}
