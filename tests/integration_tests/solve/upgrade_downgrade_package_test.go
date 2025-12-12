// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_exactlyOnePackage_selectSmallestPackage
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is only A.
func Test_exactlyOnePackage_selectSmallestPackage(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		solution,
	)
}

// Test_exactlyOnePackage_upgradeToLargerPackage_case2
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is A, then B.
func Test_exactlyOnePackage_upgradeToLargerPackage_case2(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemK":    0,
		},
		solution,
	)
}

// Test_exactlyOnePackage_upgradeToLargerPackage_case3
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is A, then C.
func Test_exactlyOnePackage_upgradeToLargerPackage_case3(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
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
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemK":    1,
		},
		solution,
	)
}

// Test_exactlyOnePackage_upgradeToLargerPackage_case4
// Ref: test_upgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is B, then C.
func Test_exactlyOnePackage_upgradeToLargerPackage_case4(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
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
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemK":    1,
		},
		solution,
	)
}

// Test_exactlyOnePackage_downgradeToSmallerPackage_case1
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is C, then A.
func Test_exactlyOnePackage_downgradeToSmallerPackage_case1(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageC").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		solution,
	)
}

// Test_exactlyOnePackage_downgradeToSmallerPackage_case2
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is B, then A.
func Test_exactlyOnePackage_downgradeToSmallerPackage_case2(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		solution,
	)
}

// Test_exactlyOnePackage_downgradeToSmallerPackage_case3
// Ref: test_downgrade_package_when_xor_between_multiple_packages
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Selected package is C, then B.
func Test_exactlyOnePackage_downgradeToSmallerPackage_case3(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageC").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemK":    0,
		},
		solution,
	)
}

// Test_exactlyOnePackage_noSelection_shouldGivePreferred
// Ref:
// Description: Here are three packages, A, B and C, exist.
// C is larger than B, and B is larger than A.
// Nothing is selected, expect the preferred package.
func Test_exactlyOnePackage_noSelection_shouldGivePreferred(t *testing.T) {
	ruleset := upgradeDowngradePackageWithSharedItemsSmallestPreferred()

	selections := puan.Selections{}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"packageC": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    0,
			"itemK":    0,
		},
		solution,
	)
}

func upgradeDowngradePackageWithSharedItemsSmallestPreferred() puan.Ruleset {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "packageB", "packageC", "itemX", "itemY", "itemZ", "itemK")

	includedItemsInA, _ := creator.SetAnd("itemX", "itemY")
	includedItemsInB, _ := creator.SetAnd("itemX", "itemY", "itemZ")
	includedItemsInC, _ := creator.SetAnd("itemX", "itemY", "itemZ", "itemK")

	packageARequiredItems, _ := creator.SetImply("packageA", includedItemsInA)
	packageBRequiredItems, _ := creator.SetImply("packageB", includedItemsInB)
	packageCRequiredItems, _ := creator.SetImply("packageC", includedItemsInC)

	exactlyOnePackage, _ := creator.SetXor("packageA", "packageB", "packageC")
	anyOfThePackages, _ := creator.SetOr("packageA", "packageB", "packageC")
	packageBOrC, _ := creator.SetOr("packageB", "packageC")

	itemsInAllPackages, _ := creator.SetImply(includedItemsInA, anyOfThePackages)
	itemsInPackageBOrC, _ := creator.SetImply(includedItemsInB, packageBOrC)
	reversedPackageC, _ := creator.SetImply(includedItemsInC, "packageC")

	_ = creator.Assume(
		exactlyOnePackage,
		packageARequiredItems,
		packageBRequiredItems,
		packageCRequiredItems,
		itemsInAllPackages,
		itemsInPackageBOrC,
		reversedPackageC,
	)

	_ = creator.Prefer("packageA")

	ruleset, _ := creator.Create()

	return ruleset
}
