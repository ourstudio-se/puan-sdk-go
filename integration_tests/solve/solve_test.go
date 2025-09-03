// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

const url = "http://127.0.0.1:9000"

// Test_exactlyOnePackage_selectPreferredThenNotPreferred_shouldReturnNotPreferred
// Ref: test_select_exactly_one_constrainted_component_with_additional_requirements
// Description: Exactly one of package A, B or C must be selected. A is preferred. B requires another
// variable itemX. Now, A is preselected and we select B. We expect (B, itemX) as result.
func Test_exactlyOnePackage_selectPreferredThenNotPreferred_shouldReturnNotPreferred(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "itemX")
	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	packageB, _ := model.SetEquivalent("packageB", "itemX")

	root, _ := model.SetAnd(exactlyOnePackage, packageB)
	_ = model.Assume(root)

	invertedPreferred, _ := model.SetNot("packageA")

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
	}

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]string{invertedPreferred},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB")

	packageARequiredPackageB, _ := model.SetImply("packageA", "packageB")

	_ = model.Assume(packageARequiredPackageB)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := puan.Selections{
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
		{
			ID:     "packageB",
			Action: puan.REMOVE,
		},
	}

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, nil)
	solution, _ := client.Solve(polyhedron, model.Variables(), objective)

	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC")

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")

	root, _ := model.SetAnd(exactlyOnePackage)
	_ = model.Assume(root)

	invertedPreferred, _ := model.SetNot("packageA")

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	selections := puan.Selections{
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
		{
			ID:     "packageB",
			Action: puan.REMOVE,
		},
	}

	selectionsIDs := selections.GetImpactingSelectionIDS()

	objective := puan.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]string{invertedPreferred},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "itemX", "itemY", "itemM", "itemN", "itemO")

	includedItemsInVariantOne, _ := model.SetAnd("itemX", "itemY", "itemN")
	includedItemsInVariantTwo, _ := model.SetAnd("itemX", "itemY", "itemM", "itemO")

	packageVariantOne, _ := model.SetAnd("packageA", includedItemsInVariantOne)
	packageVariantTwo, _ := model.SetAnd("packageA", includedItemsInVariantTwo)

	exactlyOnePackage, _ := model.SetXor(packageVariantOne, packageVariantTwo)

	reversedPackageVariantOne, _ := model.SetImply(includedItemsInVariantOne, "packageA")
	reversedPackageVariantTwo, _ := model.SetImply(includedItemsInVariantTwo, "packageA")

	root, _ := model.SetAnd(exactlyOnePackage, reversedPackageVariantOne, reversedPackageVariantTwo)
	_ = model.Assume(root)

	invertedPreferred, _ := model.SetNot(packageVariantOne)

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{}

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]string{invertedPreferred},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "packageE", "packageF", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := model.SetAnd("itemX", "itemY", "itemZ")
	packageARequiresItems, _ := model.SetImply("packageA", includedItemsInA)

	packageERequiresF, _ := model.SetImply("packageE", "packageF")
	packageFRequiresA, _ := model.SetImply("packageF", "packageA")

	reversedPackageA, _ := model.SetImply(includedItemsInA, "packageA")

	root, _ := model.SetAnd(
		packageERequiresF,
		packageFRequiresA,
		packageARequiresItems,
		reversedPackageA,
	)
	_ = model.Assume(root)

	selections := puan.Selections{
		{
			ID:     "packageE",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, nil)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
func Test_variantsWithXORBetweenTwoItems_selectedPreferredXOR_shouldReturnPreferred(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives("packageA", "itemX", "itemY", "itemZ", "itemN", "itemM")

	sharedItems, _ := model.SetAnd("itemX", "itemY", "itemZ")
	packageRequiresItems, _ := model.SetImply("packageA", sharedItems)

	exactlyOneOfTheItems, _ := model.SetXor("itemN", "itemM")
	variants, _ := model.SetImply("packageA", exactlyOneOfTheItems)

	negatedPreferred, _ := model.SetNot("itemN")
	invertedPreferred, _ := model.SetAnd("packageA", negatedPreferred)

	includedItemsInVariantOne, _ := model.SetAnd("itemX", "itemY", "itemZ", "itemN")
	includedItemsInVariantTwo, _ := model.SetAnd("itemX", "itemY", "itemZ", "itemM")

	reversedPackageVariantOne, _ := model.SetImply(includedItemsInVariantOne, "packageA")
	reversedPackageVariantTwo, _ := model.SetImply(includedItemsInVariantTwo, "packageA")

	root, _ := model.SetAnd("packageA", packageRequiresItems, variants, reversedPackageVariantOne, reversedPackageVariantTwo)

	_ = model.Assume(root)

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "itemM",
			Action: puan.ADD,
		},
		{
			ID:     "itemN",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, []string{invertedPreferred})

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "packageD", "packageE")
	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC", "packageD", "packageE")

	root, _ := model.SetAnd(exactlyOnePackage)
	_ = model.Assume(root)

	invertedPreferred, _ := model.SetNot("packageA")

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, []string{invertedPreferred})

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItemXItemY, _ := model.SetXor("itemX", "itemY")
	xorItemXItemZ, _ := model.SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := model.SetImply("packageA", xorItemXItemY)
	packageExactlyOneOfItem1Item3, _ := model.SetImply("packageA", xorItemXItemZ)

	reversePackageVariantOne, _ := model.SetImply("itemX", "packageA")
	includedItemsInVariantTwo, _ := model.SetAnd("itemY", "itemZ")
	reversePackageVariantTwo, _ := model.SetImply(includedItemsInVariantTwo, "packageA")

	root, _ := model.SetAnd(
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
		reversePackageVariantOne,
		reversePackageVariantTwo,
	)

	_ = model.Assume(root)

	negatedPreferred, _ := model.SetNot("itemX")
	invertedPreferred, _ := model.SetAnd("packageA", negatedPreferred)

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "itemX",
			Action: puan.ADD,
		},
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "itemY",
			Action: puan.ADD,
		},
		{
			ID:     "itemZ",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]string{invertedPreferred},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := model.SetAnd("itemX", "itemY")
	includedItemsInB, _ := model.SetAnd("itemX", "itemY", "itemZ")

	packageARequiresItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiresItems, _ := model.SetImply("packageB", includedItemsInB)

	notPackageB, _ := model.SetNot("packageB")
	packageAForbidsB, _ := model.SetImply("packageA", notPackageB) // Law of implication

	packageAOrB, _ := model.SetOr("packageA", "packageB")
	reversedPackageAOrB, _ := model.SetImply(includedItemsInA, packageAOrB)
	reversedPackageB, _ := model.SetImply(includedItemsInB, "packageB")

	root, _ := model.SetAnd(
		packageARequiresItems,
		packageBRequiresItems,
		packageAForbidsB,
		reversedPackageAOrB,
		reversedPackageB,
	)

	_ = model.Assume(root)

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		nil,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "itemX", "itemY")

	includedItemsInA, _ := model.SetAnd("itemX", "itemY")
	packageARequiresItems, _ := model.SetEquivalent("packageA", includedItemsInA)

	root, _ := model.SetAnd(
		packageARequiresItems,
	)

	_ = model.Assume(root)

	selections := puan.Selections{
		{
			ID:     "notExistingID",
			Action: puan.ADD,
		},
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)
	model.Variables()
	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(
		model.PrimitiveVariables(),
		selectionsIDs,
		nil,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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

// TODO: This test is skipped for now, since it is not possible to construct atm.
func Test_variable_will_be_removed_after_chosen_with_many_variables_in_selected(t *testing.T) {
	/*
		This is a quite special case, since it yet cannot be constructed
		using polytope builder. We need a free-selectable variable x that
		later are selected among y and z. When we then select x again,
		we expect it to be removed
	*/
	t.Skip()
}

// Test_notPreferCombinationsWithRequires_exclusively
// Ref: test_will_not_prefer_preferred_combinations_for_requires_exclusivelies
// Description: Let
// packageZ -> xor(itemX, itemY) (pref itemX)
// packageZ -> itemM & itemN & itemO
// packageA -> itemB
// We preselect packageA and selects itemX.
// We do not expect packageZ to be selected
func Test_notPreferCombinationsWithRequires_exclusively(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives("packageA", "packageZ", "itemB", "itemX", "itemY", "itemM", "itemN", "itemO")

	exactlyOneIfItemXAndY, _ := model.SetXor("itemX", "itemY")
	packageZRequiresExactlyOneOfItemXOrY, _ := model.SetImply("packageZ", exactlyOneIfItemXAndY)

	requiredItemsInZ, _ := model.SetAnd("itemM", "itemN", "itemO")
	packageZRequiresItems, _ := model.SetImply("packageZ", requiredItemsInZ)

	packageARequiresItemB, _ := model.SetImply("packageA", "itemB")

	root, _ := model.SetAnd(
		packageZRequiresExactlyOneOfItemXOrY,
		packageZRequiresItems,
		packageARequiresItemB,
	)

	_ = model.Assume(root)

	negatedPreferred, _ := model.SetNot("itemX")
	invertedPreferred, _ := model.SetAnd("packageZ", negatedPreferred)

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "itemX",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, []string{invertedPreferred})

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageA", "packageP", "itemB", "itemX", "itemY")

	exactlyOneOfItemXAndY, _ := model.SetXor("itemX", "itemY")
	packagePRequiresExactlyOneOfItemXOrY, _ := model.SetImply("packageP", exactlyOneOfItemXAndY)

	packageARequiresItemB, _ := model.SetImply("packageA", "itemB")

	root, _ := model.SetAnd(
		packagePRequiresExactlyOneOfItemXOrY,
		packageARequiresItemB,
	)

	_ = model.Assume(root)
	selections := puan.Selections{
		{
			ID:     "itemB",
			Action: puan.ADD,
		},
		{
			ID:     "itemX",
			Action: puan.ADD,
		},
		{
			ID:     "packageP",
			Action: puan.ADD,
		},
		{
			ID:     "itemY",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, nil)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
	model := pldag.New()
	model.SetPrimitives("packageP", "itemX", "itemY", "itemA", "itemB", "itemC")

	includedItemsInPackage, _ := model.SetAnd("itemA", "itemB", "itemC")
	packageRequiresItems, _ := model.SetEquivalent("packageP", includedItemsInPackage)

	exactlyOneOfItemXAndY, _ := model.SetXor("itemX", "itemY")
	packageRequiresExactlyOneOfItemXOrY, _ := model.SetImply("packageP", exactlyOneOfItemXAndY)

	root, _ := model.SetAnd(
		packageRequiresItems,
		packageRequiresExactlyOneOfItemXOrY,
	)

	_ = model.Assume(root)
	selections := puan.Selections{
		{
			ID:     "packageP",
			Action: puan.ADD,
		},
		{
			ID:     "itemY",
			Action: puan.ADD,
		},
		{
			ID:     "packageP",
			Action: puan.ADD,
		},
		{
			ID:     "itemX",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, nil)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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
