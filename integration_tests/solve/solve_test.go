// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

const url = "http://127.0.0.1:9000"

// Test_exactlyOnePackage_selectPreferredThenNotPreferred_shouldReturnNotPreferred
// Ref: test_select_exactly_one_constrainted_component_with_additional_requirements
// Description: Exactly one of package A, B or C must be selected. A is preferred. B requires another
// variable itemX. Now, A is preselected and we select B. We expect (B, itemX) as result.
func Test_exactlyOnePackage_selectPreferredThenNotPreferred_shouldReturnNotPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "packageB", "packageC", "itemX")
	exactlyOnePackage, _ := creator.PLDAG().SetXor("packageA", "packageB", "packageC")

	packageBRequiresItemX, _ := creator.PLDAG().SetImply("packageB", "itemX")

	root, _ := creator.PLDAG().SetAnd(exactlyOnePackage, packageBRequiresItemX)
	_ = creator.PLDAG().Assume(root)

	invertedPreferred, _ := creator.PLDAG().SetNot("packageA")
	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

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
			"packageC": 0,
			"itemX":    1,
		},
		primitiveSolution,
	)
}

// Test_packageImpliesAnotherPackage_selectedAndDeselect_shouldReturnCheapestSolution
// Ref: test_select_same_not_constrainted_selected_component
// Description: package A requires B. B has been preselected and is then removed.
// We now expect the empty set as the result.
func Test_packageImpliesAnotherPackage_selectedAndDeselect_shouldReturnCheapestSolution(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "packageB")
	packageARequiredPackageB, _ := creator.PLDAG().SetImply("packageA", "packageB")

	_ = creator.PLDAG().Assume(packageARequiredPackageB)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageB").WithAction(puan.REMOVE).Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_selectAndDeselectNotPreferred_shouldReturnPreferred
// Ref: test_select_same_selected_exactly_one_constrainted_component
// Description: Exactly one of package A, B or C must be selected, but A is preferred.
// B has been preselected but is selected again. We now expect A to be selected.
func Test_exactlyOnePackage_selectAndDeselectNotPreferred_shouldReturnPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "packageB", "packageC")

	exactlyOnePackage, _ := creator.PLDAG().SetXor("packageA", "packageB", "packageC")

	root, _ := creator.PLDAG().SetAnd(exactlyOnePackage)
	_ = creator.PLDAG().Assume(root)

	invertedPreferred, _ := creator.PLDAG().SetNot("packageA")
	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageB").WithAction(puan.REMOVE).Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_nothingIsSelected_shouldReturnPreferred
// Ref: test_default_component_in_package_when_part_in_multiple_xors
// Description: Package A has two variants: (A, itemX, itemY, itemN) and (A, itemX, itemY, itemM, itemO)
// with preferred on the former.
// Nothing is preselected and we expect (A, itemX, itemY, itemN) as our result configuration.
func Test_exactlyOnePackage_nothingIsSelected_shouldReturnPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemM", "itemN", "itemO")

	itemsXAndY, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	packageARequiresItems, _ := creator.PLDAG().SetImply("packageA", itemsXAndY)

	exactlyOneOfItemMAndM, _ := creator.PLDAG().SetXor("itemN", "itemM")
	packageARequiresExactlyOneOfItemMAndN, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfItemMAndM)

	exactlyOneOfItemOAndM, _ := creator.PLDAG().SetXor("itemN", "itemO")
	packageARequiresExactlyOneOfItemOAndN, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfItemOAndM)

	root, _ := creator.PLDAG().SetAnd("packageA", packageARequiresItems, packageARequiresExactlyOneOfItemMAndN, packageARequiresExactlyOneOfItemOAndN)
	_ = creator.PLDAG().Assume(root)

	invertedPreferred, _ := creator.PLDAG().SetNot("itemN")
	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

	selections := puan.Selections{}

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
			"itemN":    1,
			"itemM":    0,
			"itemO":    0,
		},
		primitiveSolution,
	)
}

// Test_implicationChain_shouldReturnAllAsTrue
// Ref: test_select_component_with_indirect_package_requirement
// Description: There exists a chain of requirements: E -> F -> A -> (itemX, itemY,itemZ).
// We select E and expect our result configuration to (E, F, A, itemX, itemY, itemZ)
func Test_implicationChain_shouldReturnAllAsTrue(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "packageE", "packageF", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")
	packageARequiresItems, _ := creator.PLDAG().SetImply("packageA", includedItemsInA)

	packageERequiresF, _ := creator.PLDAG().SetImply("packageE", "packageF")
	packageFRequiresA, _ := creator.PLDAG().SetImply("packageF", "packageA")

	root, _ := creator.PLDAG().SetAnd(
		packageERequiresF,
		packageFRequiresA,
		packageARequiresItems,
	)
	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageE").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageE": 1,
			"packageF": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

// Test_variantsWithXORBetweenTwoItems_selectedPreferredXOR_shouldReturnPreferred
// Ref: test_package_variant_will_change_when_selecting_another_xor_component
// Description: Given package A -> and(itemX, itemY, itemZ), xor(itemN,itemM)), reversed package rules
// and(itemX, itemY, itemZ, itemN) -> A, and(itemX, itemY, itemZ, itemM) -> A) and with preferred
// on variant (A,itemN), we test that if variant (A, itemX, itemY, itemZ, itemM) is preselected,
// and we select single variable itemN, then we will change into the other
// package variant (A, itemX, itemY, itemZ, itemN) (and not select single itemN)
// Note: package A is mandatory according to rule set.
func Test_variantsWithXORBetweenTwoItems_selectXORItemAfterAnotherVariant_shouldReturnNewVariant(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemZ", "itemN", "itemM")

	sharedItems, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")
	packageRequiresItems, _ := creator.PLDAG().SetImply("packageA", sharedItems)

	exactlyOneOfItemNAndM, _ := creator.PLDAG().SetXor("itemN", "itemM")
	packageRequiresExactlyOneOfItemNAndM, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfItemNAndM)

	includedItemsInVariantOne, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ", "itemN")
	includedItemsInVariantTwo, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ", "itemM")

	reversedPackageVariantOne, _ := creator.PLDAG().SetImply(includedItemsInVariantOne, "packageA")
	reversedPackageVariantTwo, _ := creator.PLDAG().SetImply(includedItemsInVariantTwo, "packageA")

	root, _ := creator.PLDAG().SetAnd("packageA", packageRequiresItems, packageRequiresExactlyOneOfItemNAndM, reversedPackageVariantOne, reversedPackageVariantTwo)

	_ = creator.PLDAG().Assume(root)

	preferred, _ := creator.PLDAG().SetImply("packageA", "itemN")
	invertedPreferred, _ := creator.PLDAG().SetNot(preferred)

	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

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

// Test_multiplePackagesWithXOR_shouldReturnSelected
// Ref: test_deselect_exactly_one_constrainted_variables_from_sequence
// Description: Following rules are applied (with preferreds on the left xor-component)
// xor(packageA, packageB, packageC, packageD, packageE)
// We have already selected packageA and now we select packageB.
// We expect packageB to be the only one in configuration
func Test_multiplePackagesWithXOR_shouldReturnSelected(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "packageB", "packageC", "packageD", "packageE")
	exactlyOnePackage, _ := creator.PLDAG().SetXor("packageA", "packageB", "packageC", "packageD", "packageE")

	root, _ := creator.PLDAG().SetAnd(exactlyOnePackage)
	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

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
			"packageC": 0,
			"packageD": 0,
			"packageE": 0,
		},
		primitiveSolution,
	)
}

// Test_optionalPackageWithLightPreferred_selectNotPreferred_shouldReturnNotPreferred
// Ref: test_will_delete_package_variant_from_pre_selected_actions_when_conflicting
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). itemX is preferred oved (itemY, itemZ).
// We first select the preferred package variant and the change to the not preferred variant.
func Test_optionalPackageWithLightPreferred_selectNotPreferred_shouldReturnNotPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItemXItemY, _ := creator.PLDAG().SetXor("itemX", "itemY")
	xorItemXItemZ, _ := creator.PLDAG().SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := creator.PLDAG().SetImply("packageA", xorItemXItemY)
	packageExactlyOneOfItem1Item3, _ := creator.PLDAG().SetImply("packageA", xorItemXItemZ)

	reversePackageVariantOne, _ := creator.PLDAG().SetImply("itemX", "packageA")
	includedItemsInVariantTwo, _ := creator.PLDAG().SetAnd("itemY", "itemZ")
	reversePackageVariantTwo, _ := creator.PLDAG().SetImply(includedItemsInVariantTwo, "packageA")

	root, _ := creator.PLDAG().SetAnd(
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
		reversePackageVariantOne,
		reversePackageVariantTwo,
	)

	_ = creator.PLDAG().Assume(root)

	preferredVariant, _ := creator.PLDAG().SetImply("packageA", "itemX")
	invertedPreferred, _ := creator.PLDAG().SetNot(preferredVariant)

	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

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

// Test_twoPackagesWithSharedItems_selectLargestPackage_shouldReturnSelectedPackage
// Ref: test_will_delete_package_from_selected_actions_when_adding_upgrading_package
// Description: Following rules are applied (with preferreds on the left xor-component)
// packageA -> (itemX, itemY)
// packageB -> (itemX, itemY, itemZ)
// packageA -> -packageB
// packageB -> -packageA
// (itemX, itemY) -> or(packageA, packageB)
// (itemX, itemY, itemX) -> packageB
// We have already selected packageA and now we select packageB. We expect packageB to be selected
// and packageA deleted from pre selected actions
func Test_twoPackagesWithSharedItems_selectLargestPackage_shouldReturnSelectedPackage(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	includedItemsInB, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")

	packageARequiresItems, _ := creator.PLDAG().SetImply("packageA", includedItemsInA)
	packageBRequiresItems, _ := creator.PLDAG().SetImply("packageB", includedItemsInB)

	notPackageB, _ := creator.PLDAG().SetNot("packageB")
	packageAForbidsB, _ := creator.PLDAG().SetImply("packageA", notPackageB)

	packageAOrB, _ := creator.PLDAG().SetOr("packageA", "packageB")
	reversedPackageAOrB, _ := creator.PLDAG().SetImply(includedItemsInA, packageAOrB)
	reversedPackageB, _ := creator.PLDAG().SetImply(includedItemsInB, "packageB")

	root, _ := creator.PLDAG().SetAnd(
		packageARequiresItems,
		packageBRequiresItems,
		packageAForbidsB,
		reversedPackageAOrB,
		reversedPackageB,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

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

// Test_ignoreNotExistingVariable_shouldReturnValidSelection
// Ref: test_will_ignore_pre_selected_actions_not_existing_in_action_space
// Description: Following rules are applied (with preferreds on the left xor-component)
// packageA -> (itemX, itemY)
// We give pre selected action ['itemZ'] (which is not in action space) and
// expects solution to ignore it
func Test_ignoreNotExistingVariable_shouldReturnValidSelection(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY")

	includedItemsInA, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	packageARequiresItems, _ := creator.PLDAG().SetEquivalent("packageA", includedItemsInA)

	root, _ := creator.PLDAG().SetAnd(
		packageARequiresItems,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("notExistingID").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
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
		},
		primitiveSolution,
	)
}

// Test_notPreferCombinationsWithRequires_exclusively
// Ref: test_will_not_prefer_preferred_combinations_for_requires_exclusivelies
// Description: Let
// packageZ -> xor(itemX, itemY) (pref itemX)
// packageZ -> itemM & itemN & itemO
// packageA -> itemB
// We preselect packageA and selects itemX.
// We do not expect packageZ to be selected
// Comment: From python test preferreds are packageZ and itemX without condition.
// Here the preferred is modeled packageZ -> itemX.
// How should we interpret the python test, with defaultconfiguration?
func Test_notPreferCombinationsWithRequires_exclusively(t *testing.T) {
	t.Skip()
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "packageZ", "itemB", "itemX", "itemY", "itemM", "itemN", "itemO")

	exactlyOneIfItemXAndY, _ := creator.PLDAG().SetXor("itemX", "itemY")
	packageZRequiresExactlyOneOfItemXOrY, _ := creator.PLDAG().SetImply("packageZ", exactlyOneIfItemXAndY)

	requiredItemsInZ, _ := creator.PLDAG().SetAnd("itemM", "itemN", "itemO")
	packageZRequiresItems, _ := creator.PLDAG().SetImply("packageZ", requiredItemsInZ)

	packageARequiresItemB, _ := creator.PLDAG().SetImply("packageA", "itemB")

	root, _ := creator.PLDAG().SetAnd(
		packageZRequiresExactlyOneOfItemXOrY,
		packageZRequiresItems,
		packageARequiresItemB,
	)

	_ = creator.PLDAG().Assume(root)

	preferredZWithX, _ := creator.PLDAG().SetImply("packageZ", "itemX")
	invertedPreferred, _ := creator.PLDAG().SetNot(preferredZWithX)
	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("itemX").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageZ": 0,
			"itemB":    1,
			"itemX":    1,
			"itemY":    0,
			"itemM":    0,
			"itemN":    0,
			"itemO":    0,
		},
		primitiveSolution,
	)
}

// Test_selectPackageAfterItemSelection_shouldReturnPackage
// Ref: test_will_select_package_when_variant_component_in_selections
// Description: Let
// packageP -> xor(itemX, itemY)
// packageA -> itemB
// We preselect itemX and selects itemB.
// We expect (packageP, itemY) and packageA to be selected
func Test_selectPackageAfterItemSelection_shouldReturnPackage(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "packageP", "itemB", "itemX", "itemY")

	exactlyOneOfItemXAndY, _ := creator.PLDAG().SetXor("itemX", "itemY")
	packagePRequiresExactlyOneOfItemXOrY, _ := creator.PLDAG().SetImply("packageP", exactlyOneOfItemXAndY)

	packageARequiresItemB, _ := creator.PLDAG().SetImply("packageA", "itemB")

	root, _ := creator.PLDAG().SetAnd(
		packagePRequiresExactlyOneOfItemXOrY,
		packageARequiresItemB,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemB").Build(),
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("packageP").WithSubSelectionID("itemY").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageP": 1,
			"itemB":    1,
			"itemX":    0,
			"itemY":    1,
		},
		primitiveSolution,
	)
}

// Test_changeVariant_shouldReturnSelected
// Ref: test_select_package_variant_x_when_package_variant_y_is_selected
// Description: Let
// packageP -> itemX xor itemY
// packageP -> itemA & itemB & itemC
// we preselect (packageP, itemX) and select (packageP, itemY). We
// expects (packageP, itemX) to be removed from
// selected variants.
func Test_changeVariant_shouldReturnSelected(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageP", "itemX", "itemY", "itemA", "itemB", "itemC")

	includedItemsInPackage, _ := creator.PLDAG().SetAnd("itemA", "itemB", "itemC")
	packageRequiresItems, _ := creator.PLDAG().SetEquivalent("packageP", includedItemsInPackage)

	exactlyOneOfItemXAndY, _ := creator.PLDAG().SetXor("itemX", "itemY")
	packageRequiresExactlyOneOfItemXOrY, _ := creator.PLDAG().SetImply("packageP", exactlyOneOfItemXAndY)

	root, _ := creator.PLDAG().SetAnd(
		packageRequiresItems,
		packageRequiresExactlyOneOfItemXOrY,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageP").WithSubSelectionID("itemY").Build(),
		puan.NewSelectionBuilder("packageP").WithSubSelectionID("itemX").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageP": 1,
			"itemX":    1,
			"itemY":    0,
			"itemA":    1,
			"itemB":    1,
			"itemC":    1,
		},
		primitiveSolution,
	)
}

// Test_notPreferPackage_whenXorComponentsInVariantsHasBeenSelected
// Ref: test_will_not_prefer_package_when_xor_components_in_variants
// Description: Following rules are applied
// xor(itemX, itemY, itemZ)
// packageP -> xor(itemX, itemY)
// packageP -> xor(itemA, itemB)
// pref(itemZ)
// pref(packageP, itemX)
// pref(packageP, itemA)
// pref(packageP, itemA, itemX)
// We expect package to not be included on initial state
// but itemZ to be selected alone
// If nothin is selected initially, preferred packageP, itemX, itemA should be chosen.
// If itemX has been selected and unselected, we expect itemZ to be selected alone since it is not the initial state anymore.
// Comment: How should we interpret the python test, with defaultconfiguration?
func Test_notPreferPackage_whenXorComponentsInVariantsHasBeenSelected(t *testing.T) {
	t.Skip()
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageP", "itemA", "itemB", "itemX", "itemY", "itemZ")

	exactlyOneOfTheItemsXYZ, _ := creator.PLDAG().SetXor("itemX", "itemY", "itemZ")

	exactlyOneOfItemXAndY, _ := creator.PLDAG().SetXor("itemX", "itemY")
	packagePRequiresExactlyOneOfTheItems, _ := creator.PLDAG().SetImply("packageP", exactlyOneOfItemXAndY)

	exactlyOneOfItemAAndB, _ := creator.PLDAG().SetXor("itemA", "itemB")
	packagePRequiresExactlyOneOfItemAAndB, _ := creator.PLDAG().SetImply("packageP", exactlyOneOfItemAAndB)

	root, _ := creator.PLDAG().SetAnd(
		exactlyOneOfTheItemsXYZ,
		packagePRequiresExactlyOneOfTheItems,
		packagePRequiresExactlyOneOfItemAAndB,
	)

	_ = creator.PLDAG().Assume(root)

	invertedZPreferred, _ := creator.PLDAG().SetNot("itemZ")

	prefItemsInPackageP, _ := creator.PLDAG().SetAnd("itemA", "itemX")
	prefPackagePImpliesAAndItemX, _ := creator.PLDAG().SetImply("packageP", prefItemsInPackageP)
	invertedPackagePAndPackageAAndItemX, _ := creator.PLDAG().SetNot(prefPackagePImpliesAAndItemX)

	_ = creator.SetPreferreds(invertedZPreferred, invertedPackagePAndPackageAAndItemX)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("itemX").WithAction(puan.REMOVE).Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageP": 0,
			"itemA":    0,
			"itemB":    0,
			"itemX":    0,
			"itemY":    0,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

// Test_changeItem_shouldReturnSelectedItem
// Ref: test_will_delete_package_from_selected_actions_when_new_conflicting_package
// Description: Following rules are applied
// a -> !(b & c)
// b -> !(a & c)
// c -> !(a & b)
// a -> xor(n,m)
// a -> xor(x,y,z)
// b -> xor(x,y)
// c -> xor(x)
// A variant of 'a' is pre selected and we select 'c'.
// We expect the variant of 'a' to disappear from selected actions
func Test_changeItem_shouldReturnSelectedItem(t *testing.T) {
	ruleSet := forbidsBetweenItemsWithAdditionalXORsSetup()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemA").WithSubSelectionID("itemN").Build(),
		puan.NewSelectionBuilder("itemC").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"itemA": 0,
			"itemB": 0,
			"itemC": 1,
			"itemM": 0,
			"itemN": 0,
			"itemX": 1,
			"itemY": 0,
			"itemZ": 0,
		},
		primitiveSolution,
	)
}

// Test_changeItemWithHeavyConsequence_shouldReturnSelectedItem
// Ref: test_will_delete_package_from_selected_actions_when_new_conflicting_package_reversed_order
// Description: Following rules are applied
// a -> !(b & c)
// b -> !(a & c)
// c -> !(a & b)
// a -> xor(n,m)
// a -> xor(x,y,z)
// b -> xor(x,y)
// c -> xor(x)
// A variant of 'a' is pre selected and we select 'c'.
// We expect the variant of 'a' to disappear from selected actions
func Test_changeItemWithHeavyConsequence_shouldReturnSelectedItem(t *testing.T) {
	ruleSet := forbidsBetweenItemsWithAdditionalXORsSetup()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemC").Build(),
		puan.NewSelectionBuilder("itemA").WithSubSelectionID("itemN").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"itemA": 1,
			"itemB": 0,
			"itemC": 0,
			"itemM": 0,
			"itemN": 1,
			"itemX": 1,
			"itemY": 1,
			"itemZ": 1,
		},
		primitiveSolution,
	)
}

// Test_preferredWithImply_selectItemInPreferred_shouldOnlyReturnItem
// Ref: test_preferred_components_order_when_having_duplicated_rules
// Description:
// A preferred rule's components will end up in the weight
// polytope where order is lost. Now we try to retain order
// but we want to check that it also works for duplicated rules.
// TODO: how should we mock this test? Duplicated variables should be handled
func Test_preferredWithImply_selectItemInPreferred_shouldOnlyReturnItem(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("itemA", "itemB", "itemC", "itemX", "itemY")

	exactlyOneOfItemAAndB, _ := creator.PLDAG().SetXor("itemA", "itemB")
	exactlyOneOfItemBAndC, _ := creator.PLDAG().SetXor("itemB", "itemC")

	itemXRequiresExactlyOneOfItemAAndB, _ := creator.PLDAG().SetImply("itemX", exactlyOneOfItemAAndB)
	itemYRequiresExactlyOneOfItemBAndC, _ := creator.PLDAG().SetImply("itemY", exactlyOneOfItemBAndC)

	root, _ := creator.PLDAG().SetAnd(
		itemXRequiresExactlyOneOfItemAAndB,
		itemYRequiresExactlyOneOfItemBAndC,
	)

	_ = creator.PLDAG().Assume(root)

	preferredX, _ := creator.PLDAG().SetImply("itemX", "itemA")
	invertedX, _ := creator.PLDAG().SetNot(preferredX)

	preferredXDuplicated, _ := creator.PLDAG().SetImply("itemX", "itemA")
	invertedXDuplicated, _ := creator.PLDAG().SetNot(preferredXDuplicated)

	preferredY, _ := creator.PLDAG().SetImply("itemY", "itemB")
	invertedY, _ := creator.PLDAG().SetNot(preferredY)

	// Comment: returns error due to duplicated variables.
	_ = creator.SetPreferreds(invertedX, invertedY, invertedXDuplicated)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemA").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"itemA": 1,
			"itemB": 0,
			"itemC": 0,
			"itemX": 0,
			"itemY": 0,
		},
		primitiveSolution,
	)
}

// Test_xorBetweenPackagesAndItems_shouldGiveLastSelection
// Ref: test_will_only_remove_one_selection_if_three_or_more_are_conflicting
// Description: Following rules are applied
// packageA -> xor(itemX, itemY, itemZ)
// packageB -> xor(itemX, itemY, itemZ)
// itemX -> xor(itemA, itemB)
// itemY -> xor(packageA, packageB)
// itemZ -> xor(packageA, packageB)
// itemX and packageA is selected first and then itemY is selected.
// This lead to a selection list of [itemX, packageA, itemY]
// When itemY is selected, we expect the list to become [packageA, itemY]
// This is because itemX was selected firstly and has most less
// priority.
func Test_xorBetweenPackagesAndItems_shouldGiveLastSelection(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	exactlyOneOfItemXYZ, _ := creator.PLDAG().SetXor("itemX", "itemY", "itemZ")
	packageARequiresExactlyOneOfItemXYZ, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfItemXYZ)
	packageBRequiresExactlyOneOfItemXYZ, _ := creator.PLDAG().SetImply("packageB", exactlyOneOfItemXYZ)

	exactlyOneOfPackageAAndB, _ := creator.PLDAG().SetXor("packageA", "packageB")
	itemXRequiresExactlyOneOfPackageAAndB, _ := creator.PLDAG().SetImply("itemX", exactlyOneOfPackageAAndB)
	itemYRequiresExactlyOneOfPackageAAndB, _ := creator.PLDAG().SetImply("itemY", exactlyOneOfPackageAAndB)
	itemZRequiresExactlyOneOfPackageAAndB, _ := creator.PLDAG().SetImply("itemZ", exactlyOneOfPackageAAndB)

	root, _ := creator.PLDAG().SetAnd(
		packageARequiresExactlyOneOfItemXYZ,
		packageBRequiresExactlyOneOfItemXYZ,
		itemXRequiresExactlyOneOfPackageAAndB,
		itemYRequiresExactlyOneOfPackageAAndB,
		itemZRequiresExactlyOneOfPackageAAndB,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
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
			"packageB": 0,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_xorBetweenPackagesAndItemsWithPreferred_shouldGiveLastSelection
// Ref: test_will_only_remove_one_selection_if_three_or_more_are_conflicting_reverse_selections
// Description: Following rules are applied
// packageA -> xor(itemX, itemY, itemZ)
// packageB -> xor(itemX, itemY, itemZ)
// itemX -> xor(itemA, itemB)
// itemY -> xor(packageA, packageB)
// itemZ -> xor(packageA, packageB)
// Preferred(packageA)
// packageA and itemX is selected first and then itemY is selected.
// This lead to a selection list of [packageA, itemX, itemY]
// When itemY is selected, we expect the list to become [itemX, itemY]
// This is because packageA was selected firstly and has most less
// priority. The configuration is expected as (packageA, itemY), since
// packageA is preferred over packageB, and itemY since it was later selected than itemX.
func Test_xorBetweenPackagesAndItemsWithPreferred_shouldGiveLastSelection(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	exactlyOneOfItemXYZ, _ := creator.PLDAG().SetXor("itemX", "itemY", "itemZ")
	packageARequiresExactlyOneOfItemXYZ, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfItemXYZ)
	packageBRequiresExactlyOneOfItemXYZ, _ := creator.PLDAG().SetImply("packageB", exactlyOneOfItemXYZ)

	exactlyOneOfPackageAAndB, _ := creator.PLDAG().SetXor("packageA", "packageB")
	itemXRequiresExactlyOneOfPackageAAndB, _ := creator.PLDAG().SetImply("itemX", exactlyOneOfPackageAAndB)
	itemYRequiresExactlyOneOfPackageAAndB, _ := creator.PLDAG().SetImply("itemY", exactlyOneOfPackageAAndB)
	itemZRequiresExactlyOneOfPackageAAndB, _ := creator.PLDAG().SetImply("itemZ", exactlyOneOfPackageAAndB)

	root, _ := creator.PLDAG().SetAnd(
		packageARequiresExactlyOneOfItemXYZ,
		packageBRequiresExactlyOneOfItemXYZ,
		itemXRequiresExactlyOneOfPackageAAndB,
		itemYRequiresExactlyOneOfPackageAAndB,
		itemZRequiresExactlyOneOfPackageAAndB,
	)

	_ = creator.PLDAG().Assume(root)

	invertedPreferredPackageA, _ := creator.PLDAG().SetNot("packageA")

	_ = creator.SetPreferreds(invertedPreferredPackageA)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("itemX").Build(),
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
			"packageB": 0,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_checkConflictingPreferred_shouldReturnSelectionsWithUnselectedPreferred
// Ref: test_will_not_change_variant_when_variant_should_not_be_choosable
// Description: Following rules are applied
// xor(itemX, itemY, itemZ)
// itemX -> xor(itemA, itemB, itemC)
// itemY -> xor(itemA, itemB, itemC)
// itemZ -> xor(itemA, itemB, itemC)
// itemN -> forb(itemA)
// pref(itemX)
// pref(itemA)
// pref(itemN, itemB)
// pre_selected = [itemB]
// We will check and see that selected actions
// for action itemN won't include [itemN, itemB], but just
// [itemN]. Since pref(itemN, itemB), there will be a variant
// action [itemN, itemB] which should not be able to select.
// In other cases, we would want to change variant
// to [itemN, itemB] but only if it is choosable.
// Comment: How should we interpret the python test, with defaultconfiguration?
func Test_checkConflictingPreferred_shouldReturnSelectionsWithUnselectedPreferred(t *testing.T) {
	t.Skip()
	creator := puan.NewRuleSetCreator()

	creator.PLDAG().SetPrimitives("itemA", "itemB", "itemC", "itemN", "itemX", "itemY", "itemZ")

	exactlyOneOfItemXYZ, _ := creator.PLDAG().SetXor("itemX", "itemY", "itemZ")

	exactlyOneOfItemABC, _ := creator.PLDAG().SetXor("itemA", "itemB", "itemC")
	itemXRequiresExactlyOneOfItemABC, _ := creator.PLDAG().SetImply("itemX", exactlyOneOfItemABC)
	itemYRequiresExactlyOneOfItemABC, _ := creator.PLDAG().SetImply("itemY", exactlyOneOfItemABC)
	itemZRequiresExactlyOneOfItemABC, _ := creator.PLDAG().SetImply("itemZ", exactlyOneOfItemABC)

	notItemA, _ := creator.PLDAG().SetNot("itemA")
	itemNForbidsItemA, _ := creator.PLDAG().SetImply("itemN", notItemA)

	root, _ := creator.PLDAG().SetAnd(
		exactlyOneOfItemXYZ,
		itemXRequiresExactlyOneOfItemABC,
		itemYRequiresExactlyOneOfItemABC,
		itemZRequiresExactlyOneOfItemABC,
		itemNForbidsItemA,
	)

	_ = creator.PLDAG().Assume(root)

	preferredItemNWithItemB, _ := creator.PLDAG().SetImply("itemN", "itemB")
	invertedPreferredItemNWithItemB, _ := creator.PLDAG().SetNot(preferredItemNWithItemB)

	invertedItemB, _ := creator.PLDAG().SetNot("itemB")
	invertedItemX, _ := creator.PLDAG().SetNot("itemX")
	invertedItemA, _ := creator.PLDAG().SetNot("itemA")

	_ = creator.SetPreferreds(invertedPreferredItemNWithItemB, invertedItemB, invertedItemX, invertedItemA)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemB").Build(),
		puan.NewSelectionBuilder("itemN").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"itemA": 0,
			"itemB": 1,
			"itemC": 0,
			"itemN": 1,
			"itemX": 1,
			"itemY": 0,
			"itemZ": 0,
		},
		primitiveSolution,
	)
}

func Test_removingItemInAddedPackage_shouldRemovePackageAsWell(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY")

	itemXAndY, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	packageARequiresItemXAndY, _ := creator.PLDAG().SetImply("packageA", itemXAndY)

	root, _ := creator.PLDAG().SetAnd(packageARequiresItemXAndY)
	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("itemX").WithAction(puan.REMOVE).Build(),
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
		},
		primitiveSolution,
	)
}

func forbidsBetweenItemsWithAdditionalXORsSetup() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("itemA", "itemB", "itemC", "itemN", "itemM", "itemX", "itemY", "itemZ")

	notItemB, _ := creator.PLDAG().SetNot("itemB")
	notItemC, _ := creator.PLDAG().SetNot("itemC")

	// Note: Law of implication A -> !B is equivalent to !A v !B
	itemAForbidsItemB, _ := creator.PLDAG().SetImply("itemA", notItemB)
	itemAForbidsItemC, _ := creator.PLDAG().SetImply("itemA", notItemC)
	itemBForbidsItemC, _ := creator.PLDAG().SetImply("itemB", notItemC)

	exactlyOneOfTheItemsNM, _ := creator.PLDAG().SetXor("itemN", "itemM")
	itemARequiresExactlyOneOfItemsNM, _ := creator.PLDAG().SetImply("itemA", exactlyOneOfTheItemsNM)

	itemsXYZ, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")
	itemARequiresItemsXYZ, _ := creator.PLDAG().SetImply("itemA", itemsXYZ)

	itemsXY, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	itemBRequiresItemsXY, _ := creator.PLDAG().SetImply("itemB", itemsXY)

	itemCRequiresItemsX, _ := creator.PLDAG().SetImply("itemC", "itemX")

	root, _ := creator.PLDAG().SetAnd(
		itemAForbidsItemB,
		itemAForbidsItemC,
		itemBForbidsItemC,
		itemARequiresExactlyOneOfItemsNM,
		itemARequiresItemsXYZ,
		itemBRequiresItemsXY,
		itemCRequiresItemsX,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	return ruleSet
}
