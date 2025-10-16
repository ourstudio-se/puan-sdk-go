// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_exactlyOnePackage_selectNotPreferredThenPreferred_shouldGivePreferred
// Ref: test_select_package_when_xor_between_packages_and_larger_package_is_selected
// Description: Two packages A and B exists, with B being the larger one
// and exactly one of them has to be selected.
// B has been preselected and we select A. We know expect
// A to be selected without nothing left from B.
func Test_exactlyOnePackage_selectNotPreferredThenPreferred_shouldGivePreferred(t *testing.T) {
	ruleset := packagesWithSharedItemsSmallerPackagePreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
	}

	solutionCreator := puan.NewSolutionCreator(glpk.NewClient(url))
	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
		},
		solution,
	)
}

// Test_exactlyOnePackage_selectNotPreferred
// Ref: test_select_package_when_xor_between_packages
// Description: Two packages, A and B, exist with B being the larger one.
// They both share a subset of variables, and exactly one
// of A and B must be selected, but with A as preferred.
// With nothing being preselected, we select B and expect to get B.
func Test_exactlyOnePackage_selectNotPreferred(t *testing.T) {
	ruleset := packagesWithSharedItemsSmallerPackagePreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
	}

	solutionCreator := puan.NewSolutionCreator(glpk.NewClient(url))
	solution, _ := solutionCreator.Create(selections, *ruleset, nil)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
		},
		solution,
	)
}

func packagesWithSharedItemsSmallerPackagePreferred() *puan.Ruleset {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := creator.SetAnd("itemX", "itemY")
	includedItemsInB, _ := creator.SetAnd("itemX", "itemY", "itemZ")

	packageARequiredItems, _ := creator.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := creator.SetImply("packageB", includedItemsInB)

	exactlyOnePackage, _ := creator.SetXor("packageA", "packageB")

	anyOfThePackages, _ := creator.SetOr("packageA", "packageB")
	itemsInAllPackages, _ := creator.SetImply(includedItemsInA, anyOfThePackages)
	reversedPackageB, _ := creator.SetImply(includedItemsInB, "packageB")

	_ = creator.Assume(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		itemsInAllPackages,
		reversedPackageB,
	)

	_ = creator.Prefer("packageA")

	ruleset, _ := creator.Create()

	return ruleset
}
