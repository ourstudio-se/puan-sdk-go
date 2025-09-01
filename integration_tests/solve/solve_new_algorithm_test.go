// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

// Test_exactlyOnePackage_selectPreferredThenNotPreferred_shouldReturnNotPreferred_newAlgorithm
// Ref: test_select_exactly_one_constrainted_component_with_additional_requirements
// Description: Exactly one of package A, B or C must be selected. A is preferred. B requires another
// variable itemX. Now, A is preselected and we select B. We expect (B, itemX) as result.
func Test_exactlyOnePackage_selectPreferredThenNotPreferred_shouldReturnNotPreferred_newAlgorithm(t *testing.T) {
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "item1")
	packageA, _ := model.SetAnd("packageA")
	packageB, _ := model.SetAnd("packageB")
	packageC, _ := model.SetAnd("packageC")
	exactlyOnePackage, _ := model.SetXor(packageA, packageB, packageC)

	xorWithPreference := puan.XORWithPreference2{
		PreferredID: packageA,
		IDs:         []string{packageA, packageB, packageC},
	}

	packageBRequiredItems, _ := model.SetEquivalent(packageB, "item1")

	root, _ := model.SetAnd(exactlyOnePackage, packageBRequiredItems)
	_ = model.Assume(root)

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
	objective := puan.CalculateObjective2(
		model.PrimitiveVariables(),
		selectionsIDs,
		[]puan.XORWithPreference2{xorWithPreference},
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)

	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"packageC": 0,
			"item1":    1,
		},
		primitiveSolution,
	)
}

// Test_packageImpliesAnotherPackage_selectedAndDeselect_shouldReturnCheapestSolution_newAlgorithm
// Ref: test_select_same_not_constrainted_selected_component
// Description: package A requires B. B has been preselected and is then removed.
// We now expect the empty set as the result.
func Test_packageImpliesAnotherPackage_selectedAndDeselect_shouldReturnCheapestSolution_newAlgorithm(t *testing.T) {
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
