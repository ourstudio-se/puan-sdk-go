//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
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

	_ = creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemZ", "itemN", "itemM")

	sharedItems, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")
	packageRequiresItems, _ := creator.PLDAG().SetImply("packageA", sharedItems)

	exactlyOneOfItemNAndM, _ := creator.PLDAG().SetXor("itemN", "itemM")
	packageRequiresExactlyOneOfItemNAndM, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfItemNAndM)

	includedItemsInVariantOne, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ", "itemN")
	includedItemsInVariantTwo, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ", "itemM")

	reversedPackageVariantOne, _ := creator.PLDAG().SetImply(includedItemsInVariantOne, "packageA")
	reversedPackageVariantTwo, _ := creator.PLDAG().SetImply(includedItemsInVariantTwo, "packageA")

	_ = creator.SetAssumedVariables("packageA", packageRequiresItems, packageRequiresExactlyOneOfItemNAndM, reversedPackageVariantOne, reversedPackageVariantTwo)

	preferred, _ := creator.PLDAG().SetImply("packageA", "itemN")
	_ = creator.SetPreferreds(preferred)

	ruleSet, err := creator.Create()
	assert.NoError(t, err)

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemM").Build(),
		puan.NewSelectionBuilder("itemN").Build(),
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
			"itemY":    1,
			"itemZ":    1,
			"itemN":    1,
			"itemM":    0,
		},
		primitiveSolution,
	)
}

// Test_optionalPackageWithSmallPreferred_selectNotPreferred
// Ref: test_will_delete_package_variant_from_pre_selected_actions_when_conflicting
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). itemX is preferred oved (itemY, itemZ).
// We first select the preferred package variant and the change to the not preferred variant.
func Test_optionalPackageWithSmallPreferred_selectNotPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItemXItemY, _ := creator.PLDAG().SetXor("itemX", "itemY")
	xorItemXItemZ, _ := creator.PLDAG().SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := creator.PLDAG().SetImply("packageA", xorItemXItemY)
	packageExactlyOneOfItem1Item3, _ := creator.PLDAG().SetImply("packageA", xorItemXItemZ)

	reversePackageVariantOne, _ := creator.PLDAG().SetImply("itemX", "packageA")
	includedItemsInVariantTwo, _ := creator.PLDAG().SetAnd("itemY", "itemZ")
	reversePackageVariantTwo, _ := creator.PLDAG().SetImply(includedItemsInVariantTwo, "packageA")

	_ = creator.SetAssumedVariables(
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
		reversePackageVariantOne,
		reversePackageVariantTwo,
	)

	preferredVariant, _ := creator.PLDAG().SetImply("packageA", "itemX")
	_ = creator.SetPreferreds(preferredVariant)

	ruleSet, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
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

	_ = creator.PLDAG().SetPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	includedItemsInB, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")

	packageARequiresItems, _ := creator.PLDAG().SetImply("packageA", includedItemsInA)
	packageBRequiresItems, _ := creator.PLDAG().SetImply("packageB", includedItemsInB)

	notPackageB, _ := creator.PLDAG().SetNot("packageB")
	packageAForbidsB, _ := creator.PLDAG().SetImply("packageA", notPackageB)

	packageAOrB, _ := creator.PLDAG().SetOr("packageA", "packageB")
	reversedPackageAOrB, _ := creator.PLDAG().SetImply(includedItemsInA, packageAOrB)
	reversedPackageB, _ := creator.PLDAG().SetImply(includedItemsInB, "packageB")

	_ = creator.SetAssumedVariables(
		packageARequiresItems,
		packageBRequiresItems,
		packageAForbidsB,
		reversedPackageAOrB,
		reversedPackageB,
	)

	ruleSet, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}
