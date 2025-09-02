// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

// Test_exactlyOnePackage_selectPreferredThenNotPreferred_shouldReturnNotPreferred
// Ref: test_select_exactly_one_constrainted_component_with_additional_requirements
// Description: Exactly one of package A, B or C must be selected. A is preferred. B requires another
// variable itemX. Now, A is preselected and we select B. We expect (B, itemX) as result.

const url = "http://127.0.0.1:9000"

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
