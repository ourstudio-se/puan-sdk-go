//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

// Test_changeHeavyVariant_shouldReturnSelectedVariant
// Ref: test_will_change_heavy_package_variant_when_single_option_is_preselected
// Description: Following rules are applied
// packageA -> xor(itemX, itemY)
// packageA -> itemM, itemN, itemO, itemP, itemQ, itemR, itemS
// We give pre selected action [itemX] and selects [packageA, itemY] and
// expects solution [packageA, itemX] and pre selected [[itemX], [itemA, itemX]]
func Test_changeHeavyVariant_shouldReturnSelectedVariant(t *testing.T) {
	ruleSet := heavyVariantSetup()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemM":    1,
			"itemN":    1,
			"itemO":    1,
			"itemP":    1,
			"itemQ":    1,
			"itemR":    1,
			"itemS":    1,
		},
		primitiveSolution,
	)
}

// Test_changeHeavyVariant_withVariantSelection_shouldReturnSelectedVariant
// Ref: test_will_change_heavy_package_variant_is_pre_selected_and_other_package_variant_option_is_selected
func Test_changeHeavyVariant_withVariantSelection_shouldReturnSelectedVariant(t *testing.T) {
	ruleSet := heavyVariantSetup()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
		puan.NewSelectionBuilder("itemY").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"itemX":    0,
			"itemY":    1,
			"itemM":    1,
			"itemN":    1,
			"itemO":    1,
			"itemP":    1,
			"itemQ":    1,
			"itemR":    1,
			"itemS":    1,
		},
		primitiveSolution,
	)
}

func heavyVariantSetup() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "itemX", "itemY", "itemM", "itemN", "itemO", "itemP", "itemQ", "itemR", "itemS")

	exactlyOneOfItemXAndY, _ := creator.PLDAG().SetXor("itemX", "itemY")
	packageARequiresExactlyOneOfXAndY, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfItemXAndY)

	includedItemsInA, _ := creator.PLDAG().SetAnd("itemM", "itemN", "itemO", "itemP", "itemQ", "itemR", "itemS")
	packageARequiresItems, _ := creator.PLDAG().SetImply("packageA", includedItemsInA)

	root, _ := creator.PLDAG().SetAnd(
		packageARequiresExactlyOneOfXAndY,
		packageARequiresItems,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	return ruleSet
}
