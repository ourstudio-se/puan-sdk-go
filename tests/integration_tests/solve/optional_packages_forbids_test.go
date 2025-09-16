//nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
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
	ruleSet := optionalPackagesWithForbids()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemN").Build(),
		puan.NewSelectionBuilder("packageC").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
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
		primitiveSolution,
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
	ruleSet := optionalPackagesWithForbids()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageC").Build(),
		puan.NewSelectionBuilder("packageA").WithSubSelectionID("itemN").Build(),
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
			"packageC": 0,
			"itemM":    0,
			"itemN":    1,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
		},
		primitiveSolution,
	)
}

func Test_optionalPackagesWithForbids_noSelection(t *testing.T) {
	ruleSet := optionalPackagesWithForbids()

	selections := puan.Selections{}

	query, _ := ruleSet.NewQuery(selections)
	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
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
		primitiveSolution,
	)
}

func optionalPackagesWithForbids() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "packageB", "packageC", "itemN", "itemM", "itemX", "itemY", "itemZ")

	notPackageB, _ := creator.PLDAG().SetNot("packageB")
	notPackageC, _ := creator.PLDAG().SetNot("packageC")

	// Note: Law of implication A -> !B is equivalent to !A v !B
	packageAForbidsPackageB, _ := creator.PLDAG().SetImply("packageA", notPackageB)
	packageAForbidsPackageC, _ := creator.PLDAG().SetImply("packageA", notPackageC)
	packageBForbidsPackageC, _ := creator.PLDAG().SetImply("packageB", notPackageC)

	exactlyOneOfTheItemsNM, _ := creator.PLDAG().SetXor("itemN", "itemM")
	packageARequiresExactlyOneOfItemsNM, _ := creator.PLDAG().SetImply("packageA", exactlyOneOfTheItemsNM)

	itemsXYZ, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")
	packageARequiresItemsXYZ, _ := creator.PLDAG().SetImply("packageA", itemsXYZ)

	itemsXY, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	packageBRequiresItemsXY, _ := creator.PLDAG().SetImply("packageB", itemsXY)

	packageCRequiresItemsX, _ := creator.PLDAG().SetImply("packageC", "itemX")

	root, _ := creator.PLDAG().SetAnd(
		packageAForbidsPackageB,
		packageAForbidsPackageC,
		packageBForbidsPackageC,
		packageARequiresExactlyOneOfItemsNM,
		packageARequiresItemsXYZ,
		packageBRequiresItemsXY,
		packageCRequiresItemsX,
	)

	_ = creator.PLDAG().Assume(root)

	ruleSet := creator.Create()

	return ruleSet
}
