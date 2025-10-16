//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_variantsWithXORBetweenTwoItems_selectVariantThenItemInOtherVariant_shouldGiveNewVariant
// Ref: test_package_variant_will_change_when_selecting_another_xor_component
// Description: Given package A -> and(itemX, itemY, itemZ), xor(itemN,itemM)), reversed package rules
// and(itemX, itemY, itemZ, itemN) -> A, and(itemX, itemY, itemZ, itemM) -> A) and with preferred
// on variant (A,itemN), we test that if variant (A, itemX, itemY, itemZ, itemM) is preselected,
// and we select single variable itemN, then we will change into the other
// package variant (A, itemX, itemY, itemZ, itemN) (and not select single itemN)
// Note: package A is mandatory according to rule set.
func Test_variantsWithXORBetweenTwoItems_selectVariantThenItemInOtherVariant_shouldGiveNewVariant(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "itemX", "itemY", "itemZ", "itemN", "itemM")

	sharedItems, _ := creator.SetAnd("itemX", "itemY", "itemZ")
	packageRequiresItems, _ := creator.SetImply("packageA", sharedItems)

	exactlyOneOfItemNAndM, _ := creator.SetXor("itemN", "itemM")
	packageRequiresExactlyOneOfItemNAndM, _ := creator.SetImply("packageA", exactlyOneOfItemNAndM)

	includedItemsInVariantOne, _ := creator.SetAnd("itemX", "itemY", "itemZ", "itemN")
	includedItemsInVariantTwo, _ := creator.SetAnd("itemX", "itemY", "itemZ", "itemM")

	reversedPackageVariantOne, _ := creator.SetImply(includedItemsInVariantOne, "packageA")
	reversedPackageVariantTwo, _ := creator.SetImply(includedItemsInVariantTwo, "packageA")

	_ = creator.Assume("packageA", packageRequiresItems, packageRequiresExactlyOneOfItemNAndM, reversedPackageVariantOne, reversedPackageVariantTwo)

	preferred, _ := creator.SetImply("packageA", "itemN")
	_ = creator.Prefer(preferred)

	ruleset, err := creator.Create()
	assert.NoError(t, err)

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemM").Build(),
		puan.NewSelectionBuilder("itemN").Build(),
	}

	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemN":    1,
			"itemM":    0,
		},
		solution,
	)
}

// Test_optionalPackageWithSmallPreferred_selectNotPreferred
// Ref: test_will_delete_package_variant_from_pre_selected_actions_when_conflicting
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). itemX is preferred oved (itemY, itemZ).
// We first select the preferred package variant and the change to the not preferred variant.
func Test_optionalPackageWithSmallPreferred_selectNotPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItemXItemY, _ := creator.SetXor("itemX", "itemY")
	xorItemXItemZ, _ := creator.SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := creator.SetImply("packageA", xorItemXItemY)
	packageExactlyOneOfItem1Item3, _ := creator.SetImply("packageA", xorItemXItemZ)

	reversePackageVariantOne, _ := creator.SetImply("itemX", "packageA")
	includedItemsInVariantTwo, _ := creator.SetAnd("itemY", "itemZ")
	reversePackageVariantTwo, _ := creator.SetImply(includedItemsInVariantTwo, "packageA")

	_ = creator.Assume(
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
		reversePackageVariantOne,
		reversePackageVariantTwo,
	)

	preferredVariant, _ := creator.SetImply("packageA", "itemX")
	_ = creator.Prefer(preferredVariant)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").WithSubSelectionID("itemZ").Build(),
	}

	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    1,
		},
		solution,
	)
}

// Test_twoPackagesWithSharedItems_selectLargestPackage
// Ref: test_will_delete_package_from_selected_actions_when_adding_upgrading_package
// Description: Following rules are applied (with preferreds on the left xor-component)
// packageA -> (itemX, itemY)
// packageB -> (itemX, itemY, itemZ)
// packageA -> -packageB
// packageB -> -packageA
// (itemX, itemY) -> or(packageA, packageB)
// (itemX, itemY, itemX) -> packageB
// We have already selected packageA and now we select packageB. We expect packageB to be selected.
func Test_twoPackagesWithSharedItems_selectLargestPackage(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := creator.SetAnd("itemX", "itemY")
	includedItemsInB, _ := creator.SetAnd("itemX", "itemY", "itemZ")

	packageARequiresItems, _ := creator.SetImply("packageA", includedItemsInA)
	packageBRequiresItems, _ := creator.SetImply("packageB", includedItemsInB)

	notPackageB, _ := creator.SetNot("packageB")
	packageAForbidsB, _ := creator.SetImply("packageA", notPackageB)

	packageAOrB, _ := creator.SetOr("packageA", "packageB")
	reversedPackageAOrB, _ := creator.SetImply(includedItemsInA, packageAOrB)
	reversedPackageB, _ := creator.SetImply(includedItemsInB, "packageB")

	_ = creator.Assume(
		packageARequiresItems,
		packageBRequiresItems,
		packageAForbidsB,
		reversedPackageAOrB,
		reversedPackageB,
	)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
	}

	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
		},
		solution,
	)
}
