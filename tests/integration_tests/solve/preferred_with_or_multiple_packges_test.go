package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_multiplePackagesWithOr_allSelectedExceptA
// Description: There are four packages (A, B, C, D) with A being preferred.
// Package B requires itemX. We select nothing and expect the result
// configuration (A) since package A is preferred.
func Test_multiplePackagesWithOr_noSelectionExpectPreferred(t *testing.T) {
	ruleset := multiplePackagesWithOr()

	selections := puan.Selections{}

	query, _ := ruleset.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    0,
			"packageC": 0,
			"packageD": 0,
		},
		primitiveSolution,
	)
}

// Test_multiplePackagesWithOr_selectPackageB
// Description: There are four packages (A, B, C, D) with A being preferred.
// Package B requires itemX. We select package B and expect the result
// configuration (B, itemX).
func Test_multiplePackagesWithOr_selectPackageB(t *testing.T) {
	ruleset := multiplePackagesWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
	}

	query, _ := ruleset.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    1,
			"packageC": 0,
			"packageD": 0,
		},
		primitiveSolution,
	)
}

// Test_multiplePackagesWithOr_selectPackageC
// Description: There are four packages (A, B, C, D) with A being preferred.
// Package B requires itemX. We select package C and expect the result
// configuration (C).
func Test_multiplePackagesWithOr_selectPackageC(t *testing.T) {
	ruleset := multiplePackagesWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageC").Build(),
	}

	query, _ := ruleset.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 0,
			"itemX":    0,
			"packageC": 1,
			"packageD": 0,
		},
		primitiveSolution,
	)
}

// Test_multiplePackagesWithOr_allSelectedExceptA
// Description: There are four packages (A, B, C, D) with A being preferred.
// Package B requires itemX. We select packages B, C and D and expect the result
// configuration (B, C, D, itemX).
func Test_multiplePackagesWithOr_allSelectedExceptA(t *testing.T) {
	ruleset := multiplePackagesWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageC").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageD").Build(),
	}

	query, _ := ruleset.NewQuery(selections)

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    1,
			"packageC": 1,
			"packageD": 1,
		},
		primitiveSolution,
	)
}

func multiplePackagesWithOr() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()
	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "packageC", "packageD")

	anyOfThePackages, _ := creator.SetOr("packageA", "packageB", "packageC", "packageD")
	packageBRequiresItemX, _ := creator.SetImply("packageB", "itemX")

	_ = creator.Assume(anyOfThePackages, packageBRequiresItemX)

	unPreferredPackages, _ := creator.SetOr("packageB", "packageC", "packageD")
	notPreferred, _ := creator.SetNot(unPreferredPackages)
	packageAAndNotPreferred, _ := creator.SetAnd("packageA", notPreferred)

	_ = creator.Prefer(packageAAndNotPreferred)

	ruleset, _ := creator.Create()

	return ruleset
}
