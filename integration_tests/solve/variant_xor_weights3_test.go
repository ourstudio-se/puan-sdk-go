// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

// Test_exactlyOneVariant_selectNotPreferred_shouldReturnSelected
// Ref: test_select_single_xor_component_when_another_xor_pair_is_preferred
// Description: Package A has two variants: (A, itemX) and (A, itemY, itemZ) with the latter
// being preferred. We select (A, itemX) and expect the result configuration (A, itemX)
func Test_exactlyOneVariant_selectNotPreferred_shouldReturnSelected_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

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
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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

// Test_exactlyOneVariant_selectPreferred_shouldReturnPreferred
// Ref: test_select_xor_pair_when_xor_pair_is_preferred
// Description: Package A has two variants: (A, itemX) and (A, itemY, itemZ) with the latter
// being preferred. We select (A, itemY, itemZ) and expect the result configuration
// (A, itemY, itemZ). This test is just to make sure that there is no weird behavior.
func Test_exactlyOneVariant_selectPreferred_shouldReturnPreferred_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

	selections := puan.Selections{
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
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
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

// Test_exactlyOneVariant_deselecting_shouldReturnCheapestSolution
// Ref: test_deselect_package_when_xor_pair_is_preferred_over_single_xor_component
// Description: Given rules package A -> xor(itemX, itemY),
// package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If (A, itemY, itemZ) is already selected, check that we will remove package A when deselecting A.
func Test_exactlyOneVariant_deselecting_shouldReturnCheapestSolution_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

	selections := puan.Selections{
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
		{
			ID:     "packageA",
			Action: puan.REMOVE,
		},
		{
			ID:     "itemY",
			Action: puan.REMOVE,
		},
		{
			ID:     "itemZ",
			Action: puan.REMOVE,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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

// Test_exactlyOneVariant_selectItemXAfterPreferred_shouldReturnVariantWithItemX
// Ref: test_select_single_xor_component_when_xor_pair_is_already_selected
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If (A, itemY, itemZ) is already selected, check that we will select (A, itemX) variant when selecting itemX
func Test_exactlyOneVariant_selectItemXAfterPreferred_shouldReturnVariantWithItemX_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

	selections := puan.Selections{
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
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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

// Test_exactlyOneVariant_onlySelectedPackage_shouldReturnPreferredVariant
// Ref:
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If package A is selected, check that we get the preferred variant.
func Test_exactlyOneVariant_onlySelectedPackage_shouldReturnPreferredVariant_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
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

// Test_exactlyOneVariant_selectPreferredItemAfterNotPreferredItem_shouldReturnPreferredVariant
// Ref: test_select_one_component_in_xor_pair_when_single_xor_component_is_already_selected
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If package A and itemX are selected, check that we will get (A, itemY, itemZ) config when selecting item2 (or item3).
func Test_exactlyOneVariant_selectPreferredItemAfterNotPreferredItem_shouldReturnPreferredVariant_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

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
			ID:     "itemY",
			Action: puan.ADD,
		},
	}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
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

// Test_exactlyOneVariant_selectEverythingWithPreferredItemLast_shouldReturnPreferredVariant
// Ref:
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If everything is selected with itemY last, check that we will get (A, itemY, itemZ).
func Test_exactlyOneVariant_selectEverythingWithPreferredItemLast_shouldReturnPreferredVariant_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "itemZ",
			Action: puan.ADD,
		},
		{
			ID:     "itemX",
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
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
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

// Test_exactlyOneVariant_nothingIsSelected_shouldReturnCheapestSolution
// Ref:
// Description: Given rules package A -> xor(itemX, itemY), package A -> xor(itemX, itemZ). (itemY, itemZ) is preferred oved itemX.
// If nothing is selected, check that we get the cheapest solution.
func Test_exactlyOneVariant_nothingIsSelected_shouldReturnCheapestSolution_w3(t *testing.T) {
	model, xorWithPreference := exactlyOnePackageVariantWithXORBetweenItems_w3()

	selections := puan.Selections{}

	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		xorWithPreference,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
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

func exactlyOnePackageVariantWithXORBetweenItems_w3() (*pldag.Model, []string) {
	model := pldag.New()
	model.SetPrimitives("packageA", "itemX", "itemY", "itemZ")

	xorItem1Item2, _ := model.SetXor("itemX", "itemY")
	xorItem1Item3, _ := model.SetXor("itemX", "itemZ")

	packageExactlyOneOfItem1Item2, _ := model.SetImply("packageA", xorItem1Item2)
	packageExactlyOneOfItem1Item3, _ := model.SetImply("packageA", xorItem1Item3)

	includedItemsInVariantOne, _ := model.SetAnd("itemY", "itemZ")
	packageVariantOne, _ := model.SetAnd("packageA", includedItemsInVariantOne)
	packageVariantTwo, _ := model.SetAnd("packageA", "itemX")
	exactlyOneVariant, _ := model.SetXor(packageVariantOne, packageVariantTwo)

	packageA, _ := model.SetImply("packageA", exactlyOneVariant)
	reversePackageVariantOne, _ := model.SetImply(includedItemsInVariantOne, "packageA")
	reversePackageVariantTwo, _ := model.SetImply("itemX", "packageA")

	root, _ := model.SetAnd(
		packageA,
		packageExactlyOneOfItem1Item2,
		packageExactlyOneOfItem1Item3,
		reversePackageVariantOne,
		reversePackageVariantTwo,
	)

	_ = model.Assume(root)

	negatedPreferred, _ := model.SetNot(packageVariantOne)
	invertedPreferred, _ := model.SetAnd("packageA", negatedPreferred)

	return model, []string{invertedPreferred}
}
