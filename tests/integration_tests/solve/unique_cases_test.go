// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

var solutionCreator = puan.NewSolutionCreator(glpk.NewClient("http://127.0.0.1:9000"))

// Test_exactlyOnePackage_selectPreferredThenNotPreferred
// Ref: test_select_exactly_one_constrainted_component_with_additional_requirements
// Description: Exactly one of package A, B or C must be selected. A is preferred. B requires another
// variable itemX. Now, A is preselected and we select B. We expect (B, itemX) as result.
func Test_exactlyOnePackage_selectPreferredThenNotPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "packageB", "packageC", "itemX")
	exactlyOnePackage, _ := creator.SetXor("packageA", "packageB", "packageC")

	packageBRequiresItemX, _ := creator.SetImply("packageB", "itemX")

	_ = creator.Assume(exactlyOnePackage, packageBRequiresItemX)
	_ = creator.Prefer("packageA")

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"packageC": 0,
			"itemX":    1,
		},
		solution.Solution,
	)
}

// Test_packageImpliesAnotherPackage_addAndRemove_shouldGiveEmptySolution
// Ref: test_select_same_not_constrainted_selected_component
// Description: package A requires B. B has been preselected and is then removed.
func Test_packageImpliesAnotherPackage_addAndRemove_shouldGiveEmptySolution(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "packageB")
	packageARequiredPackageB, _ := creator.SetImply("packageA", "packageB")

	_ = creator.Assume(packageARequiredPackageB)
	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageB").WithAction(puan.REMOVE).Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 0,
		},
		solution.Solution,
	)
}

// Test_exactlyOnePackage_selectAndDeselectNotPreferred_shouldGivePreferred
// Ref: test_select_same_selected_exactly_one_constrainted_component
// Description: Exactly one of package A, B or C must be selected, but A is preferred.
// B has been preselected but is removed again. We now expect A to be selected.
func Test_exactlyOnePackage_selectAndDeselectNotPreferred_shouldGivePreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "packageB", "packageC")

	exactlyOnePackage, _ := creator.SetXor("packageA", "packageB", "packageC")

	_ = creator.Assume(exactlyOnePackage)
	_ = creator.Prefer("packageA")

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageB").WithAction(puan.REMOVE).Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
		},
		solution.Solution,
	)
}

// Test_exactlyOnePackage_nothingIsSelected_shouldGivePreferred
// Ref: test_default_component_in_package_when_part_in_multiple_xors
// Description: Package A has two variants: (A, itemX, itemY, itemN) and (A, itemX, itemY, itemM, itemO)
// with preferred on the former.
// Nothing is preselected and we expect (A, itemX, itemY, itemN) as our result configuration.
func Test_exactlyOnePackage_nothingIsSelected_shouldGivePreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "itemX", "itemY", "itemM", "itemN", "itemO")

	itemsXAndY, _ := creator.SetAnd("itemX", "itemY")
	packageARequiresItems, _ := creator.SetImply("packageA", itemsXAndY)

	exactlyOneOfItemMAndM, _ := creator.SetXor("itemN", "itemM")
	packageARequiresExactlyOneOfItemMAndN, _ := creator.SetImply("packageA", exactlyOneOfItemMAndM)

	exactlyOneOfItemOAndM, _ := creator.SetXor("itemN", "itemO")
	packageARequiresExactlyOneOfItemOAndN, _ := creator.SetImply("packageA", exactlyOneOfItemOAndM)

	_ = creator.Assume(
		"packageA",
		packageARequiresItems,
		packageARequiresExactlyOneOfItemMAndN,
		packageARequiresExactlyOneOfItemOAndN,
	)

	_ = creator.Prefer("itemN")

	ruleset, _ := creator.Create()

	selections := puan.Selections{}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution.Solution,
	)
}

// Test_implicationChain_shouldGiveAll
// Ref: test_select_component_with_indirect_package_requirement
// Description: There exists a chain of requirements: E -> F -> A -> (itemX, itemY,itemZ).
// We select E and expect our result configuration to (E, F, A, itemX, itemY, itemZ)
func Test_implicationChain_shouldGiveAll(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "packageE", "packageF", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := creator.SetAnd("itemX", "itemY", "itemZ")
	packageARequiresItems, _ := creator.SetImply("packageA", includedItemsInA)

	packageERequiresF, _ := creator.SetImply("packageE", "packageF")
	packageFRequiresA, _ := creator.SetImply("packageF", "packageA")

	_ = creator.Assume(
		packageERequiresF,
		packageFRequiresA,
		packageARequiresItems,
	)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageE").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution.Solution,
	)
}

// Test_multiplePackagesWithXOR_shouldGiveLastSelected
// Ref: test_deselect_exactly_one_constrainted_variables_from_sequence
// Description: Following rules are applied (with preferreds on the left xor-component)
// xor(packageA, packageB, packageC, packageD, packageE)
// We have already selected packageA and now we select packageB.
// We expect packageB to be the only one in configuration
func Test_multiplePackagesWithXOR_shouldGiveLastSelected(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "packageB", "packageC", "packageD", "packageE")
	exactlyOnePackage, _ := creator.SetXor("packageA", "packageB", "packageC", "packageD", "packageE")

	_ = creator.Assume(exactlyOnePackage)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"packageC": 0,
			"packageD": 0,
			"packageE": 0,
		},
		solution.Solution,
	)
}

// Test_ignoreNotExistingVariable_shouldGiveValidSolution
// Ref: test_will_ignore_pre_selected_actions_not_existing_in_action_space
// Description: Following rules are applied (with preferreds on the left xor-component)
// packageA -> (itemX, itemY)
// We give pre selected action ['notExistingID'], expects error
func Test_notExistingVariable_shouldGiveError(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "itemX", "itemY")

	includedItemsInA, _ := creator.SetAnd("itemX", "itemY")
	packageARequiresItems, _ := creator.SetEquivalent("packageA", includedItemsInA)

	_ = creator.Assume(packageARequiresItems)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("notExistingID").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
	}

	_, err := solutionCreator.Create(selections, ruleset, nil)
	assert.Error(t, err)
}

// Test_packageInDefaultConfig
// Ref: test_will_not_prefer_preferred_combinations_for_requires_exclusivelies
// Description: Let
// packageZ -> xor(itemX, itemY) (pref itemX)
// packageZ -> itemM & itemN & itemO
// packageA -> itemB
// We preselect packageA and selects itemX.
// We do not expect packageZ to be selected
// Comment: From python test preferreds are packageZ and itemX without condition.
// Here the preferred is modeled packageZ -> itemX.
// Comment: How should we interpret the python test, with defaultconfiguration?
func Test_packageInDefaultConfig(t *testing.T) {
	t.Skip()
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "packageZ", "itemB", "itemX", "itemY", "itemM", "itemN", "itemO")

	exactlyOneIfItemXAndY, _ := creator.SetXor("itemX", "itemY")
	packageZRequiresExactlyOneOfItemXOrY, _ := creator.SetImply("packageZ", exactlyOneIfItemXAndY)

	requiredItemsInZ, _ := creator.SetAnd("itemM", "itemN", "itemO")
	packageZRequiresItems, _ := creator.SetImply("packageZ", requiredItemsInZ)

	packageARequiresItemB, _ := creator.SetImply("packageA", "itemB")

	_ = creator.Assume(
		packageZRequiresExactlyOneOfItemXOrY,
		packageZRequiresItems,
		packageARequiresItemB,
	)

	preferredZWithX, _ := creator.SetImply("packageZ", "itemX")
	_ = creator.Prefer(preferredZWithX)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("itemX").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution.Solution,
	)
}

// Test_selectPackageWithItemAfterSingleConflictingItemSelection_shouldGivePackage
// Ref: test_will_select_package_when_variant_component_in_selections
// Description: Let
// packageP -> xor(itemX, itemY)
// packageA -> itemB
// We preselect itemB and itemX then selects package P.
// We expect (packageP, itemY) and itemB to be selected
func Test_selectPackageWithItemAfterSingleConflictingItemSelection_shouldGivePackage(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "packageP", "itemB", "itemX", "itemY")

	exactlyOneOfItemXAndY, _ := creator.SetXor("itemX", "itemY")
	packagePRequiresExactlyOneOfItemXOrY, _ := creator.SetImply("packageP", exactlyOneOfItemXAndY)

	packageARequiresItemB, _ := creator.SetImply("packageA", "itemB")

	_ = creator.Assume(
		packagePRequiresExactlyOneOfItemXOrY,
		packageARequiresItemB,
	)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemB").Build(),
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("packageP").WithSubSelectionID("itemY").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageP": 1,
			"itemB":    1,
			"itemX":    0,
			"itemY":    1,
		},
		solution.Solution,
	)
}

// Test_changeVariant_shouldGiveLastSelected
// Ref: test_select_package_variant_x_when_package_variant_y_is_selected
// Description: Let
// packageP -> itemX xor itemY
// packageP -> itemA & itemB & itemC
// we preselect (packageP, itemY) and select (packageP, itemX). We
// expects (packageP, itemY) to be removed from selected variants.
func Test_changeVariant_shouldGiveLastSelected(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageP", "itemX", "itemY", "itemA", "itemB", "itemC")

	includedItemsInPackage, _ := creator.SetAnd("itemA", "itemB", "itemC")
	packageRequiresItems, _ := creator.SetImply("packageP", includedItemsInPackage)

	exactlyOneOfItemXAndY, _ := creator.SetXor("itemX", "itemY")
	packageRequiresExactlyOneOfItemXOrY, _ := creator.SetImply("packageP", exactlyOneOfItemXAndY)

	_ = creator.Assume(
		packageRequiresItems,
		packageRequiresExactlyOneOfItemXOrY,
	)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageP").WithSubSelectionID("itemY").Build(),
		puan.NewSelectionBuilder("packageP").WithSubSelectionID("itemX").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution.Solution,
	)
}

// Test_subComponentsAndPackageInDefaultConfig
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
func Test_subComponentsAndPackageInDefaultConfig(t *testing.T) {
	t.Skip()
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageP", "itemA", "itemB", "itemX", "itemY", "itemZ")

	exactlyOneOfTheItemsXYZ, _ := creator.SetXor("itemX", "itemY", "itemZ")

	exactlyOneOfItemXAndY, _ := creator.SetXor("itemX", "itemY")
	packagePRequiresExactlyOneOfTheItems, _ := creator.SetImply("packageP", exactlyOneOfItemXAndY)

	exactlyOneOfItemAAndB, _ := creator.SetXor("itemA", "itemB")
	packagePRequiresExactlyOneOfItemAAndB, _ := creator.SetImply("packageP", exactlyOneOfItemAAndB)

	_ = creator.Assume(
		exactlyOneOfTheItemsXYZ,
		packagePRequiresExactlyOneOfTheItems,
		packagePRequiresExactlyOneOfItemAAndB,
	)

	prefItemsInPackageP, _ := creator.SetAnd("itemA", "itemX")
	prefPackagePImpliesAAndItemX, _ := creator.SetImply("packageP", prefItemsInPackageP)

	_ = creator.Prefer("itemZ", prefPackagePImpliesAAndItemX)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("itemX").WithAction(puan.REMOVE).Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution.Solution,
	)
}

// Test_duplicatedPreferred
// Ref: test_preferred_components_order_when_having_duplicated_rules
// Description:
// A preferred rule's components will end up in the weight
// polytope where order is lost. Now we try to retain order
// but we want to check that it also works for duplicated rules.
// Comment: returns error due to duplicated variables.
func Test_duplicatedPreferred(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemA", "itemB", "itemC", "itemX", "itemY")

	exactlyOneOfItemAAndB, _ := creator.SetXor("itemA", "itemB")
	exactlyOneOfItemBAndC, _ := creator.SetXor("itemB", "itemC")

	itemXRequiresExactlyOneOfItemAAndB, _ := creator.SetImply("itemX", exactlyOneOfItemAAndB)
	itemYRequiresExactlyOneOfItemBAndC, _ := creator.SetImply("itemY", exactlyOneOfItemBAndC)

	_ = creator.Assume(
		itemXRequiresExactlyOneOfItemAAndB,
		itemYRequiresExactlyOneOfItemBAndC,
	)

	preferredX, _ := creator.SetImply("itemX", "itemA")
	preferredXDuplicated, _ := creator.SetImply("itemX", "itemA")
	preferredY, _ := creator.SetImply("itemY", "itemB")

	_ = creator.Prefer(preferredX, preferredXDuplicated, preferredY)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemA").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"itemA": 1,
			"itemB": 0,
			"itemC": 0,
			"itemX": 0,
			"itemY": 0,
		},
		solution.Solution,
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

	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	exactlyOneOfItemXYZ, _ := creator.SetXor("itemX", "itemY", "itemZ")
	packageARequiresExactlyOneOfItemXYZ, _ := creator.SetImply("packageA", exactlyOneOfItemXYZ)
	packageBRequiresExactlyOneOfItemXYZ, _ := creator.SetImply("packageB", exactlyOneOfItemXYZ)

	exactlyOneOfPackageAAndB, _ := creator.SetXor("packageA", "packageB")
	itemXRequiresExactlyOneOfPackageAAndB, _ := creator.SetImply("itemX", exactlyOneOfPackageAAndB)
	itemYRequiresExactlyOneOfPackageAAndB, _ := creator.SetImply("itemY", exactlyOneOfPackageAAndB)
	itemZRequiresExactlyOneOfPackageAAndB, _ := creator.SetImply("itemZ", exactlyOneOfPackageAAndB)

	_ = creator.Assume(
		packageARequiresExactlyOneOfItemXYZ,
		packageBRequiresExactlyOneOfItemXYZ,
		itemXRequiresExactlyOneOfPackageAAndB,
		itemYRequiresExactlyOneOfPackageAAndB,
		itemZRequiresExactlyOneOfPackageAAndB,
	)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("itemY").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    0,
		},
		solution.Solution,
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

	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	exactlyOneOfItemXYZ, _ := creator.SetXor("itemX", "itemY", "itemZ")
	packageARequiresExactlyOneOfItemXYZ, _ := creator.SetImply("packageA", exactlyOneOfItemXYZ)
	packageBRequiresExactlyOneOfItemXYZ, _ := creator.SetImply("packageB", exactlyOneOfItemXYZ)

	exactlyOneOfPackageAAndB, _ := creator.SetXor("packageA", "packageB")
	itemXRequiresExactlyOneOfPackageAAndB, _ := creator.SetImply("itemX", exactlyOneOfPackageAAndB)
	itemYRequiresExactlyOneOfPackageAAndB, _ := creator.SetImply("itemY", exactlyOneOfPackageAAndB)
	itemZRequiresExactlyOneOfPackageAAndB, _ := creator.SetImply("itemZ", exactlyOneOfPackageAAndB)

	_ = creator.Assume(
		packageARequiresExactlyOneOfItemXYZ,
		packageBRequiresExactlyOneOfItemXYZ,
		itemXRequiresExactlyOneOfPackageAAndB,
		itemYRequiresExactlyOneOfPackageAAndB,
		itemZRequiresExactlyOneOfPackageAAndB,
	)

	_ = creator.Prefer("packageA")

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("itemY").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    0,
			"itemY":    1,
			"itemZ":    0,
		},
		solution.Solution,
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

	_ = creator.AddPrimitives("itemA", "itemB", "itemC", "itemN", "itemX", "itemY", "itemZ")

	exactlyOneOfItemXYZ, _ := creator.SetXor("itemX", "itemY", "itemZ")

	exactlyOneOfItemABC, _ := creator.SetXor("itemA", "itemB", "itemC")
	itemXRequiresExactlyOneOfItemABC, _ := creator.SetImply("itemX", exactlyOneOfItemABC)
	itemYRequiresExactlyOneOfItemABC, _ := creator.SetImply("itemY", exactlyOneOfItemABC)
	itemZRequiresExactlyOneOfItemABC, _ := creator.SetImply("itemZ", exactlyOneOfItemABC)

	notItemA, _ := creator.SetNot("itemA")
	itemNForbidsItemA, _ := creator.SetImply("itemN", notItemA)

	_ = creator.Assume(
		exactlyOneOfItemXYZ,
		itemXRequiresExactlyOneOfItemABC,
		itemYRequiresExactlyOneOfItemABC,
		itemZRequiresExactlyOneOfItemABC,
		itemNForbidsItemA,
	)

	preferredItemNWithItemB, _ := creator.SetImply("itemN", "itemB")

	_ = creator.Prefer(preferredItemNWithItemB, "itemB", "itemX", "itemA")

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemB").Build(),
		puan.NewSelectionBuilder("itemN").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution.Solution,
	)
}

func Test_removingItemInAddedPackage_shouldRemovePackageAsWell(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "itemX", "itemY")

	itemXAndY, _ := creator.SetAnd("itemX", "itemY")
	packageARequiresItemXAndY, _ := creator.SetImply("packageA", itemXAndY)

	_ = creator.Assume(packageARequiresItemXAndY)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("itemX").WithAction(puan.REMOVE).Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
		},
		solution.Solution,
	)
}

func Test_removePackageWithSubselection_shouldGiveEmptySolution(t *testing.T) {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "itemX", "itemY", "itemZ", "itemM", "itemN")

	exactlyOneOfItemXYZ, _ := creator.SetXor("itemX", "itemY", "itemZ")
	anyOfItems, _ := creator.SetOr("itemM", "itemN")
	itemARequiresAnyOfItems, _ := creator.SetImply("packageA", anyOfItems)
	packageARequiresExactlyOneOfItemXYZ, _ := creator.SetImply("packageA", exactlyOneOfItemXYZ)

	_ = creator.Assume(
		packageARequiresExactlyOneOfItemXYZ,
		itemARequiresAnyOfItems,
	)

	preferred, _ := creator.SetImply("packageA", "itemX")
	_ = creator.Prefer(preferred)

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemZ").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemZ").WithAction(puan.REMOVE).Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
			"itemZ":    0,
			"itemM":    0,
			"itemN":    0,
		},
		solution.Solution,
	)
}
