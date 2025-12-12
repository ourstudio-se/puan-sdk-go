//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageX": 1,
			"itemA":    0,
			"itemB":    0,
			"itemC":    1,
			"itemD":    1,
		},
		solution,
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

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageX": 1,
			"itemA":    1,
			"itemB":    1,
			"itemC":    0,
			"itemD":    0,
		},
		solution,
	)
}

func optionalVariantsWithForbids() puan.Ruleset {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemA", "itemB", "itemC", "itemD", "packageX")

	reversedItemA, _ := creator.SetImply("itemA", "packageX")

	exactlyOneOfItemCAndA, _ := creator.SetXor("itemC", "itemA")
	exactlyOneOfItemCAndAInX, _ := creator.SetImply("packageX", exactlyOneOfItemCAndA)

	exactlyOneOfItemDAndB, _ := creator.SetXor("itemD", "itemB")
	exactlyOneOfItemDAndBInX, _ := creator.SetImply("packageX", exactlyOneOfItemDAndB)

	notItemC, _ := creator.SetNot("itemC")
	itemAForbidsItemC, _ := creator.SetImply("itemA", notItemC)

	exactlyOneOfItemCAndAWithB, _ := creator.SetImply("itemB", exactlyOneOfItemCAndA)

	itemARequiresItemB, _ := creator.SetImply("itemA", "itemB")

	notItemD, _ := creator.SetNot("itemD")
	itemAForbidsItemD, _ := creator.SetImply("itemA", notItemD)

	_ = creator.Assume(
		reversedItemA,
		exactlyOneOfItemCAndAInX,
		exactlyOneOfItemDAndBInX,
		itemAForbidsItemC,
		exactlyOneOfItemCAndAWithB,
		itemARequiresItemB,
		itemAForbidsItemD,
	)

	preferredPackageXItemC, _ := creator.SetImply("packageX", "itemC")
	preferredPackageXItemD, _ := creator.SetImply("packageX", "itemD")

	_ = creator.Prefer(preferredPackageXItemC, preferredPackageXItemD)

	ruleset, _ := creator.Create()

	return ruleset
}
