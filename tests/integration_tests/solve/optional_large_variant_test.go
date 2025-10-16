//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_optionalLargeVariantWithXOR_shouldReturnSelectedVariant
// Ref: test_will_change_heavy_package_variant_when_single_option_is_preselected
// Description: Following rules are applied
// packageA -> xor(itemX, itemY)
// packageA -> itemM, itemN, itemO, itemP, itemQ, itemR, itemS
// We give pre selected action [itemX] and selects [packageA, itemY] and
// expects solution [packageA, itemY, ...]
func Test_optionalLargeVariantWithXOR_removePreselectedItem(t *testing.T) {
	ruleset := optionalLargeVariantWithXOR()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemY").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution,
	)
}

// Test_optionalLargeVariantWithXOR_shouldChangeVariant
// Ref: test_will_change_heavy_package_variant_is_pre_selected_and_other_package_variant_option_is_selected
func Test_optionalLargeVariantWithXOR_shouldChangeVariant(t *testing.T) {
	ruleset := optionalLargeVariantWithXOR()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemX").Build(),
		puan.NewSelectionBuilder("itemY").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
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
		solution,
	)
}

func Test_optionalLargeVariantWithXOR_noSelection(t *testing.T) {
	ruleset := optionalLargeVariantWithXOR()

	selections := puan.Selections{}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
			"itemM":    0,
			"itemN":    0,
			"itemO":    0,
			"itemP":    0,
			"itemQ":    0,
			"itemR":    0,
			"itemS":    0,
		},
		solution,
	)
}

func Test_optionalLargeVariantWithXOR_singleItemSelection(t *testing.T) {
	ruleset := optionalLargeVariantWithXOR()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemM").Build(),
	}

	solution, _ := solutionCreator.Create(selections, ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"itemX":    0,
			"itemY":    0,
			"itemM":    1,
			"itemN":    0,
			"itemO":    0,
			"itemP":    0,
			"itemQ":    0,
			"itemR":    0,
			"itemS":    0,
		},
		solution,
	)
}

func optionalLargeVariantWithXOR() puan.Ruleset {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "itemX", "itemY", "itemM", "itemN", "itemO", "itemP", "itemQ", "itemR", "itemS")

	exactlyOneOfItemXAndY, _ := creator.SetXor("itemX", "itemY")
	packageARequiresExactlyOneOfXAndY, _ := creator.SetImply("packageA", exactlyOneOfItemXAndY)

	includedItemsInA, _ := creator.SetAnd("itemM", "itemN", "itemO", "itemP", "itemQ", "itemR", "itemS")
	packageARequiresItems, _ := creator.SetImply("packageA", includedItemsInA)

	_ = creator.Assume(packageARequiresExactlyOneOfXAndY, packageARequiresItems)

	ruleset, _ := creator.Create()

	return ruleset
}
