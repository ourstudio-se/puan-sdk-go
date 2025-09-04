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
//// Test_changeHeavyVariant_shouldReturnSelectedVariant
//// Ref: test_will_change_heavy_package_variant_when_single_option_is_preselected
//// Description: Following rules are applied
//// packageA -> xor(itemX, itemY)
//// packageA -> itemM, itemN, itemO, itemP, itemQ, itemR, itemS
//// We give pre selected action [itemX] and selects [packageA, itemY] and
//// expects solution [packageA, itemX] and pre selected [[itemX], [itemA, itemX]]
//func Test_changeHeavyVariant_shouldReturnSelectedVariant(t *testing.T) {
//	model := heavyVariantSetup()
//
//	selections := puan.Selections{
//		{
//			ID:     "itemX",
//			action: puan.ADD,
//		},
//		{
//			ID:     "packageA",
//			action: puan.ADD,
//		},
//		{
//			ID:     "itemY",
//			action: puan.ADD,
//		},
//	}
//
//	polyhedron := model.GeneratePolyhedron()
//	client := glpk.NewClient(url)
//
//	selectionsIDs := selections.getImpactingSelectionIDS()
//	objective := puan.CalculateObjective(
//		model.PrimitiveVariables(),
//		selectionsIDs,
//		nil,
//	)
//
//	solution, _ := client.Solve(polyhedron, model.Variables(), objective)
//	primitiveSolution, _ := solution.Extract(model.PrimitiveVariables()...)
//	assert.Equal(
//		t,
//		puan.Solution{
//			"packageA": 1,
//			"itemX":    0,
//			"itemY":    1,
//			"itemM":    1,
//			"itemN":    1,
//			"itemO":    1,
//			"itemP":    1,
//			"itemQ":    1,
//			"itemR":    1,
//			"itemS":    1,
//		},
//		primitiveSolution,
//	)
//}
//
//// Test_changeHeavyVariant_withVariantSelection_shouldReturnSelectedVariant
//// Ref: test_will_change_heavy_package_variant_is_pre_selected_and_other_package_variant_option_is_selected
//func Test_changeHeavyVariant_withVariantSelection_shouldReturnSelectedVariant(t *testing.T) {
//	// TODO: Same as Test_changeHeavyVariant_shouldReturnSelectedVariant, selection of variants needs to be implemented.
//	t.Skip()
//}
//
//func heavyVariantSetup() *pldag.Model {
//	model := pldag.New()
//	model.SetPrimitives("packageA", "itemX", "itemY", "itemM", "itemN", "itemO", "itemP", "itemQ", "itemR", "itemS")
//
//	exactlyOneOfItemXAndY, _ := model.SetXor("itemX", "itemY")
//	packageARequiresExactlyOneOfXAndY, _ := model.SetImply("packageA", exactlyOneOfItemXAndY)
//
//	includedItemsInA, _ := model.SetAnd("itemM", "itemN", "itemO", "itemP", "itemQ", "itemR", "itemS")
//	packageARequiresItems, _ := model.SetImply("packageA", includedItemsInA)
//
//	root, _ := model.SetAnd(
//		packageARequiresExactlyOneOfXAndY,
//		packageARequiresItems,
//	)
//
//	_ = model.Assume(root)
//
//	return model
//}
