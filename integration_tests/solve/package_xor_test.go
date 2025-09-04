// nolint:lll
package solve

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
//	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
//	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
//)
//
//// Test_exactlyOnePackage_selectNotPreferredThenPreferred_shouldReturnPreferred
//// Ref: test_select_package_when_xor_between_packages_and_larger_package_is_selected
//// Description: Two packages A and B exists, with B being the larger one
//// and exactly one of them has to be selected.
//// B has been preselected and we select A. We know expect
//// A to be selected without nothing left from B.
//func Test_exactlyOnePackage_selectNotPreferredThenPreferred_shouldReturnPreferred(t *testing.T) {
//	model, invertedPreferred := exactlyOnePackageOfTwoAvailableWithLargerNotPreferred()
//	polyhedron := model.GeneratePolyhedron()
//	client := glpk.NewClient(url)
//	selections := puan.Selections{
//		{
//			ID:     "packageB",
//			action: puan.ADD,
//		},
//		{
//			ID:     "packageA",
//			action: puan.ADD,
//		},
//	}
//
//	selectionsIDs := selections.getImpactingSelectionIDS()
//	objective := puan.CalculateObjective(
//		model.PrimitiveVariables(),
//		selectionsIDs,
//		invertedPreferred,
//	)
//
//	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
//	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
//	assert.Equal(
//		t,
//		puan.Solution{
//			"packageA": 1,
//			"packageB": 0,
//			"itemX":    1,
//			"itemY":    1,
//			"itemZ":    0,
//		},
//		primitiveSolution,
//	)
//}
//
//// Test_exactlyOnePackage_selectNotPreferred_shouldReturnNotPreferred
//// Ref: test_select_package_when_xor_between_packages
//// Description: Two packages, A and B, exist with B being the larger one.
//// They both share a subset of variables, and exactly one
//// of A and B must be selected, but with A as preferred.
//// With nothing being preselected, we select B and expect to get B.
//func Test_exactlyOnePackage_selectNotPreferred_shouldReturnNotPreferred(t *testing.T) {
//	model, invertedPreferred := exactlyOnePackageOfTwoAvailableWithLargerNotPreferred()
//	polyhedron := model.GeneratePolyhedron()
//	client := glpk.NewClient(url)
//	selections := puan.Selections{
//		{
//			ID:     "packageB",
//			action: puan.ADD,
//		},
//	}
//
//	selectionsIDs := selections.getImpactingSelectionIDS()
//	objective := puan.CalculateObjective(
//		model.PrimitiveVariables(),
//		selectionsIDs,
//		invertedPreferred,
//	)
//
//	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
//	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
//	assert.Equal(
//		t,
//		puan.Solution{
//			"packageA": 0,
//			"packageB": 1,
//			"itemX":    1,
//			"itemY":    1,
//			"itemZ":    1,
//		},
//		primitiveSolution,
//	)
//}
//
//func exactlyOnePackageOfTwoAvailableWithLargerNotPreferred() (*pldag.Model, []string) {
//	model := pldag.New()
//	model.SetPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")
//
//	includedItemsInA, _ := model.SetAnd("itemX", "itemY")
//	includedItemsInB, _ := model.SetAnd("itemX", "itemY", "itemZ")
//
//	packageARequiredItems, _ := model.SetImply("packageA", includedItemsInA)
//	packageBRequiredItems, _ := model.SetEquivalent("packageB", includedItemsInB)
//
//	exactlyOnePackage, _ := model.SetXor("packageA", "packageB")
//
//	anyOfThePackages, _ := model.SetOr("packageA", "packageB")
//	itemsInAllPackages, _ := model.SetImply(includedItemsInA, anyOfThePackages)
//	reversedPackageB, _ := model.SetImply(includedItemsInB, "packageB")
//
//	invertedPreferred, _ := model.SetNot("packageA")
//
//	root, _ := model.SetAnd(
//		exactlyOnePackage,
//		packageARequiredItems,
//		packageBRequiredItems,
//		itemsInAllPackages,
//		reversedPackageB,
//	)
//	_ = model.Assume(root)
//
//	return model, []string{invertedPreferred}
//}
