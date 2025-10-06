// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_heavyNonPreferredWithOr_emptySelection
// Description: Package B requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select nothing and expect the preferred package A.
func Test_heavyNonPreferredWithOr_emptySelection(t *testing.T) {
	ruleset := heavyNonPreferredWithOr()

	selections := puan.Selections{}

	query, _ := ruleset.NewQuery(puan.QueryInput{Selections: selections})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    0,
			"itemY":    0,
		},
		primitiveSolution,
	)
}

// Test_heavyNonPreferredWithOr_preferSelection
// Description: Package B requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select package A and expect package A.
func Test_heavyNonPreferredWithOr_preferSelection(t *testing.T) {
	ruleset := heavyNonPreferredWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
	}

	query, _ := ruleset.NewQuery(puan.QueryInput{Selections: selections})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    0,
			"itemY":    0,
		},
		primitiveSolution,
	)
}

// Test_heavyNonPreferredWithOr_notPreferSelection
// Description: Package B requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select package B and expect only package B.
func Test_heavyNonPreferredWithOr_notPreferSelection(t *testing.T) {
	ruleset := heavyNonPreferredWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
	}

	query, _ := ruleset.NewQuery(puan.QueryInput{Selections: selections})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    1,
			"itemY":    1,
		},
		primitiveSolution,
	)
}

// Test_heavyNonPreferredWithOr_bothPackagesSelection
// Description: Package B requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select package B and Package A expect both.
func Test_heavyNonPreferredWithOr_bothPackagesSelection(t *testing.T) {
	ruleset := heavyNonPreferredWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
		puan.NewSelectionBuilder("packageA").Build(),
	}

	query, _ := ruleset.NewQuery(puan.QueryInput{Selections: selections})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := solution.Extract(ruleset.PrimitiveVariables()...)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 1,
			"itemX":    1,
			"itemY":    1,
		},
		primitiveSolution,
	)
}

func heavyNonPreferredWithOr() *puan.RuleSet {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "itemY")
	itemXAndY, _ := creator.SetAnd("itemX", "itemY")
	packageBRequiresItemXAndItemY, _ := creator.SetImply("packageB", itemXAndY)

	packageAOrPackageB, _ := creator.SetOr("packageA", "packageB")
	_ = creator.Assume(
		packageBRequiresItemXAndItemY,
		packageAOrPackageB,
	)

	notPackageB, _ := creator.SetNot("packageB")
	packageAAndNotB, _ := creator.SetAnd("packageA", notPackageB)
	_ = creator.Prefer(packageAAndNotB)

	ruleSet, _ := creator.Create()

	return ruleSet
}
