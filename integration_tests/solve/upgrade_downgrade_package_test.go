// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

// Test_exactlyOnePackage_upgrade_case1
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is only A.
func Test_exactlyOnePackage_upgrade_case1(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
	}

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_upgrade_case2
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is A, then B.
func Test_exactlyOnePackage_upgrade_case2(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
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
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
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
			"itemY":    1,
			"itemZ":    1,
			"itemK":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_upgrade_case3
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is A, then C.
func Test_exactlyOnePackage_upgrade_case3(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
		{
			ID:     "packageC",
			Action: puan.ADD,
		},
	}
	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 0,
			"packageC": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemK":    1,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_upgrade_case4
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is B, then C.
func Test_exactlyOnePackage_upgrade_case4(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
		{
			ID:     "packageC",
			Action: puan.ADD,
		},
	}
	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 0,
			"packageC": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemK":    1,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_downgrade_case1
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is C, then A.
func Test_exactlyOnePackage_downgrade_case1(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{
		{
			ID:     "packageC",
			Action: puan.ADD,
		},
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
	}
	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_downgrade_case2
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is B, then A.
func Test_exactlyOnePackage_downgrade_case2(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
		{
			ID:     "packageA",
			Action: puan.ADD,
		},
	}
	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_downgrade_case3
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is C, then B.
func Test_exactlyOnePackage_downgrade_case3(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{
		{
			ID:     "packageC",
			Action: puan.ADD,
		},
		{
			ID:     "packageB",
			Action: puan.ADD,
		},
	}

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
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
			"itemY":    1,
			"itemZ":    1,
			"itemK":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_downgrade_case4
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Nothing is selected, expect the preferred package.
func Test_exactlyOnePackage_downgrade_case4(t *testing.T) {
	model, invertedPreferred := exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest()
	polyhedron := model.GeneratePolyhedron()
	client := glpk.NewClient(url)

	selections := puan.Selections{}

	selectionsIDs := selections.GetImpactingSelectionIDS()
	objective := puan.CalculateObjective3(
		model.PrimitiveVariables(),
		selectionsIDs,
		invertedPreferred,
	)

	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		primitiveSolution,
	)
}

func exactlyOnePackageOfThreeAvailableWithPreferredAsSmallest() (*pldag.Model, []string) {
	model := pldag.New()
	model.SetPrimitives("packageA", "packageB", "packageC", "itemX", "itemY", "itemZ", "itemK")

	includedItemsInA, _ := model.SetAnd("itemX", "itemY")
	includedItemsInB, _ := model.SetAnd("itemX", "itemY", "itemZ")
	includedItemsInC, _ := model.SetAnd("itemX", "itemY", "itemZ", "itemK")

	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := model.SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := model.SetEquivalent("packageC", includedItemsInC)

	exactlyOnePackage, _ := model.SetXor("packageA", "packageB", "packageC")
	anyOfThePackages, _ := model.SetOr("packageA", "packageB", "packageC")
	packageBOrC, _ := model.SetOr("packageB", "packageC")

	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageBOrC, _ := model.SetImply(includedItemsInB, packageBOrC)
	reversedPackageC, _ := model.SetImply(includedItemsInC, "packageC")

	invertedPreferred, _ := model.SetNot("packageA")

	root, _ := model.SetAnd(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		packageCRequiredItems,
		itemsInAllPackages,
		itemsInPackageBOrC,
		reversedPackageC,
	)
	_ = model.Assume(root)

	return model, []string{invertedPreferred}
}
