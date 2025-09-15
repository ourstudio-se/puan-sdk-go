// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	puan2 "github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_exactlyOnePackage_selectSmallestPackage
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is only A.
func Test_exactlyOnePackage_selectSmallestPackage(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

// Test_exactlyOnePackage_upgradeToLargerPackage_case2
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is A, then B.
func Test_exactlyOnePackage_upgradeToLargerPackage_case2(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").Build(),
		puan2.NewSelectionBuilder("packageB").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

// Test_exactlyOnePackage_upgradeToLargerPackage_case3
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is A, then C.
func Test_exactlyOnePackage_upgradeToLargerPackage_case3(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageA").Build(),
		puan2.NewSelectionBuilder("packageC").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

// Test_exactlyOnePackage_upgradeToLargerPackage_case4
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is B, then C.
func Test_exactlyOnePackage_upgradeToLargerPackage_case4(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageB").Build(),
		puan2.NewSelectionBuilder("packageC").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

// Test_exactlyOnePackage_downgradeToSmallerPackage_case1
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is C, then A.
func Test_exactlyOnePackage_downgradeToSmallerPackage_case1(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageC").Build(),
		puan2.NewSelectionBuilder("packageA").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

// Test_exactlyOnePackage_downgradeToSmallerPackage_case2
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is B, then A.
func Test_exactlyOnePackage_downgradeToSmallerPackage_case2(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageB").Build(),
		puan2.NewSelectionBuilder("packageA").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

// Test_exactlyOnePackage_downgradeToSmallerPackage_case3
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is C, then B.
func Test_exactlyOnePackage_downgradeToSmallerPackage_case3(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("packageC").Build(),
		puan2.NewSelectionBuilder("packageB").Build(),
	}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

// Test_exactlyOnePackage_noSelection_shouldGivePreferred
// Ref:
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Nothing is selected, expect the preferred package.
func Test_exactlyOnePackage_noSelection_shouldGivePreferred(t *testing.T) {
	ruleSet := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan2.Selections{}

	query, _ := ruleSet.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleSet.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan2.Solution{
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

func upgradeDowngradePackageWithSharedItemsSmallestPreferred() *puan2.RuleSet {
	creator := puan2.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives("packageA", "packageB", "packageC", "itemX", "itemY", "itemZ", "itemK")

	includedItemsInA, _ := creator.PLDAG().SetAnd("itemX", "itemY")
	includedItemsInB, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ")
	includedItemsInC, _ := creator.PLDAG().SetAnd("itemX", "itemY", "itemZ", "itemK")

	packageARequiredItems, _ := creator.PLDAG().SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := creator.PLDAG().SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := creator.PLDAG().SetImply("packageC", includedItemsInC)

	exactlyOnePackage, _ := creator.PLDAG().SetXor("packageA", "packageB", "packageC")
	anyOfThePackages, _ := creator.PLDAG().SetOr("packageA", "packageB", "packageC")
	packageBOrC, _ := creator.PLDAG().SetOr("packageB", "packageC")

	itemsInAllPackages, _ := creator.PLDAG().SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageBOrC, _ := creator.PLDAG().SetImply(includedItemsInB, packageBOrC)
	reversedPackageC, _ := creator.PLDAG().SetImply(includedItemsInC, "packageC")

	root, _ := creator.PLDAG().SetAnd(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		packageCRequiredItems,
		itemsInAllPackages,
		itemsInPackageBOrC,
		reversedPackageC,
	)
	_ = creator.PLDAG().Assume(root)

	invertedPreferred, _ := creator.PLDAG().SetNot("packageA")
	_ = creator.SetPreferreds(invertedPreferred)

	ruleSet := creator.Create()

	return ruleSet
}
