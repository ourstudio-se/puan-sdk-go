//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_optionalPackagesWithForbids_changeToSmallerPackage
// Ref: test_will_delete_package_from_selected_actions_when_new_conflicting_package
// Description: Following rules are applied
// packageA -> !(packageB & packageC)
// packageB -> !(packageA & packageC)
// packageC -> !(packageA & packageB)
// packageA -> xor(itemN, itemM)
// packageA -> xor(itemX, itemY, itemZ)
// packageB -> xor(itemX, itemY)
// packageC -> xor(itemX)
// We expect the variant of 'package A' to disappear from selected actions when selecting 'package C'
func Test_optionalPackagesWithForbids_changeToSmallerPackage(t *testing.T) {
	ruleset := optionalPackagesWithForbids()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemN").Build(),
		puan.NewSelectionBuilder("packageC").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 0,
			"packageC": 1,
			"itemM":    0,
			"itemN":    0,
			"itemX":    1,
			"itemY":    0,
			"itemZ":    0,
		},
		solution,
	)
}

// Test_optionalPackagesWithForbids_changeToLargerPackage
// Ref: test_will_delete_package_from_selected_actions_when_new_conflicting_package_reversed_order
// Description: Following rules are applied
// packageA -> !(packageB & packageC)
// packageB -> !(packageA & packageC)
// packageC -> !(packageA & packageB)
// packageA -> xor(itemN, itemM)
// packageA -> xor(itemX, itemY, itemZ)
// packageB -> xor(itemX, itemY)
// packageC -> xor(itemX)
func Test_optionalPackagesWithForbids_changeToLargerPackage(t *testing.T) {
	ruleset := optionalPackagesWithForbids()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageC").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemN").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemM":    0,
			"itemN":    1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
		},
		solution,
	)
}

func Test_optionalPackagesWithForbids_noSelection(t *testing.T) {
	ruleset := optionalPackagesWithForbids()

	selections := puan.Selections{}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 0,
			"packageC": 0,
			"itemM":    0,
			"itemN":    0,
			"itemX":    0,
			"itemY":    0,
			"itemZ":    0,
		},
		solution,
	)
}

func optionalPackagesWithForbids() puan.Ruleset {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "packageB", "packageC", "itemN", "itemM", "itemX", "itemY", "itemZ")

	notPackageB, _ := creator.SetNot("packageB")
	notPackageC, _ := creator.SetNot("packageC")

	// Note: Law of implication A -> !B is equivalent to !A v !B
	packageAForbidsPackageB, _ := creator.SetImply("packageA", notPackageB)
	packageAForbidsPackageC, _ := creator.SetImply("packageA", notPackageC)
	packageBForbidsPackageC, _ := creator.SetImply("packageB", notPackageC)

	exactlyOneOfTheItemsNM, _ := creator.SetXor("itemN", "itemM")
	packageARequiresExactlyOneOfItemsNM, _ := creator.SetImply("packageA", exactlyOneOfTheItemsNM)

	itemsXYZ, _ := creator.SetAnd("itemX", "itemY", "itemZ")
	packageARequiresItemsXYZ, _ := creator.SetImply("packageA", itemsXYZ)

	itemsXY, _ := creator.SetAnd("itemX", "itemY")
	packageBRequiresItemsXY, _ := creator.SetImply("packageB", itemsXY)

	packageCRequiresItemsX, _ := creator.SetImply("packageC", "itemX")

	_ = creator.Assume(
		packageAForbidsPackageB,
		packageAForbidsPackageC,
		packageBForbidsPackageC,
		packageARequiresExactlyOneOfItemsNM,
		packageARequiresItemsXYZ,
		packageBRequiresItemsXY,
		packageCRequiresItemsX,
	)

	ruleset, _ := creator.Create()

	return ruleset
}
