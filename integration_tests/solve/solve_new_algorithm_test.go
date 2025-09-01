// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

func Test_select_exactly_one_constrainted_component_with_additional_requirements_new_algorithms(t *testing.T) {
	/*
		Exactly one of (a), (b) or (c) must be select. (a) is preferred. (b) requires another
		variable x. Now, (a) is preselected and we select (b). We expect (b,item1) as result.
	*/

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

func Test_select_same_not_constrainted_selected_component_new_algorithms(t *testing.T) {
	/*
		(a) requires (b). (b) has been preselected and we select (b)
		again. We now expect the empty set as the result.
	*/

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
	assert.Equal(t, 0, solution["packageA"])
	assert.Equal(t, 0, solution["packageB"])
}
