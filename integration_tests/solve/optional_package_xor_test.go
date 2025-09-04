//nolint:lll
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
//// Test_optionalVariantWithXORsBetweenItemsAndForbids_shouldReturnPreferred
//// Ref: test_will_change_package_variant_when_package_is_preselected_with_component_requiring_package
//// Description: Following rules are applied (with preferreds on the left xor-component)
//// itemA -> packageX
//// itemA -> itemB
//// itemA -> ~itemC
//// itemA -> ~itemD
//// itemC -> ~itemA
//// itemB -> xor(itemC, itemA)
//// packageX -> xor(itemC, itemA)
//// packageX -> xor(itemD, itemB)
//// Our case is that itemA is already selected, which indirectly will add
//// package X with its preferred components itemC and itemD
//// Then we select (X, itemC, itemD) and we expect itemA to be replaced
//func Test_optionalVariantWithXORsBetweenItemsAndForbids_shouldReturnPreferred(t *testing.T) {
//	model, preferredIDs := optionalPackageWithItemsWithXORsAndForbids()
//
//	selections := puan.Selections{
//		{
//			ID:     "itemA",
//			action: puan.ADD,
//		},
//		{
//			ID:     "packageX",
//			action: puan.ADD,
//		},
//		{
//			ID:     "itemC",
//			action: puan.ADD,
//		},
//		{
//			ID:     "itemD",
//			action: puan.ADD,
//		},
//	}
//
//	polyhedron := model.GeneratePolyhedron()
//	client := glpk.NewClient(url)
//
//	selectionsIDs := selections.getImpactingSelectionIDS()
//	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, preferredIDs)
//	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
//	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
//	assert.Equal(
//		t,
//		puan.Solution{
//			"packageX": 1,
//			"itemA":    0,
//			"itemB":    0,
//			"itemC":    1,
//			"itemD":    1,
//		},
//		primitiveSolution,
//	)
//}
//
//// Test_optionalVariantWithXORsBetweenItemsAndForbids_shouldReturnNotPreferred
//// Ref: test_will_change_package_variant_when_single_component_is_preselected
//// Description: Following rules are applied (with preferreds on the left xor-component)
//// itemA -> packageX
//// itemA -> itemB
//// itemA -> ~itemC
//// itemA -> ~itemD
//// itemC -> ~itemA
//// itemB -> xor(itemC, itemA)
//// packageX -> xor(itemC, itemA)
//// packageX -> xor(itemD, itemB)
//// Our case is that itemA is already selected, which indirectly will add
//// package X with its preferred components itemC and itemD
//// Then we select (X, itemC, itemD) and we expect itemA to be replaced
//func Test_optionalVariantWithXORsBetweenItemsAndForbids_shouldReturnNOTPreferred(t *testing.T) {
//	model, preferredIDs := optionalPackageWithItemsWithXORsAndForbids()
//
//	selections := puan.Selections{
//		{
//			ID:     "itemC",
//			action: puan.ADD,
//		},
//		{
//			ID:     "packageX",
//			action: puan.ADD,
//		},
//		{
//			ID:     "itemA",
//			action: puan.ADD,
//		},
//		{
//			ID:     "itemB",
//			action: puan.ADD,
//		},
//	}
//
//	polyhedron := model.GeneratePolyhedron()
//	client := glpk.NewClient(url)
//
//	selectionsIDs := selections.getImpactingSelectionIDS()
//	objective := puan.CalculateObjective(model.PrimitiveVariables(), selectionsIDs, preferredIDs)
//	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
//	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
//	assert.Equal(
//		t,
//		puan.Solution{
//			"packageX": 1,
//			"itemA":    1,
//			"itemB":    1,
//			"itemC":    0,
//			"itemD":    0,
//		},
//		primitiveSolution,
//	)
//}
//
//func optionalPackageWithItemsWithXORsAndForbids() (*pldag.Model, []string) {
//	model := pldag.New()
//	model.SetPrimitives("itemA", "itemB", "itemC", "itemD", "packageX")
//
//	reversedItemA, _ := model.SetImply("itemA", "packageX")
//
//	exactlyOneOfItemCAndA, _ := model.SetXor("itemC", "itemA")
//	exactlyOneOfItemCAndAInX, _ := model.SetImply("packageX", exactlyOneOfItemCAndA)
//
//	exactlyOneOfItemDAndB, _ := model.SetXor("itemD", "itemB")
//	exactlyOneOfItemDAndBInX, _ := model.SetImply("packageX", exactlyOneOfItemDAndB)
//
//	notItemC, _ := model.SetNot("itemC")
//	itemAForbidsItemC, _ := model.SetImply("itemA", notItemC)
//
//	exactlyOneOfItemCAndAWithB, _ := model.SetImply("itemB", exactlyOneOfItemCAndA)
//
//	itemARequiresItemB, _ := model.SetImply("itemA", "itemB")
//
//	notItemD, _ := model.SetNot("itemD")
//	itemAForbidsItemD, _ := model.SetImply("itemA", notItemD)
//
//	root, _ := model.SetAnd(
//		reversedItemA,
//		exactlyOneOfItemCAndAInX,
//		exactlyOneOfItemDAndBInX,
//		itemAForbidsItemC,
//		exactlyOneOfItemCAndAWithB,
//		itemARequiresItemB,
//		itemAForbidsItemD,
//	)
//
//	_ = model.Assume(root)
//
//	preferredVariant, _ := model.SetAnd("itemC", "itemD")
//	notPreferredVariant, _ := model.SetNot(preferredVariant)
//	invertedPreferred, _ := model.SetAnd("packageX", notPreferredVariant)
//
//	return model, []string{invertedPreferred}
//}
