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
	ruleSet := packagesWithSharedItemsSmallerPackagePreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
		},
		primitiveSolution,
	)
}

// Test_exactlyOnePackage_selectNotPreferred
// Ref: test_select_package_when_xor_between_packages
// Description: Two packages, A and B, exist with B being the larger one.
// They both share a subset of variables, and exactly one
// of A and B must be selected, but with A as preferred.
// With nothing being preselected, we select B and expect to get B.
func Test_exactlyOnePackage_selectNotPreferred(t *testing.T) {
	ruleSet := packagesWithSharedItemsSmallerPackagePreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

func packagesWithSharedItemsSmallerPackagePreferred() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()
	_ = creator.PLDAG().SetPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ")

	includedItemsInA, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	includedItemsInB, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")

	packageARequiredItems, _ := creator.PLDAG().SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := creator.PLDAG().SetImply("packageB", includedItemsInB)

	exactlyOnePackage, _ := creator.PLDAG().SetXor("packageA", "packageB")

	anyOfThePackages, _ := creator.PLDAG().SetOr("packageA", "packageB")
	itemsInAllPackages, _ := creator.PLDAG().SetImply(includedItemsInA, anyOfThePackages)
	reversedPackageB, _ := creator.PLDAG().SetImply(includedItemsInB, "packageB")

	root, _ := creator.PLDAG().SetAnd(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		itemsInAllPackages,
		reversedPackageB,
	)
	_ = creator.PLDAG().Assume(root)

	_ = creator.SetPreferreds("packageA")

	ruleSet := creator.Create()

	return ruleSet
}
