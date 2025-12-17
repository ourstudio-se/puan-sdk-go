// nolint:lll
package solve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Test_heavyPreferredWithOr_emptySelection
// Description: Package A requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select nothing and expect the preferred package A.
func Test_heavyPreferredWithOr_emptySelection(t *testing.T) {
	ruleset := heavyPreferredWithOr()

	selections := puan.Selections{}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    1,
			"itemY":    1,
		},
		solution,
	)
}

// Test_heavyPreferredWithOr_emptySelection
// Description: Package A requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select package A and expect package A.
func Test_heavyPreferredWithOr_preferSelection(t *testing.T) {
	ruleset := heavyPreferredWithOr()

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
			"itemX":    1,
			"itemY":    1,
		},
		solution,
	)
}

// Test_heavyPreferredWithOr_notPreferSelection
// Description: Package A requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select package B and expect only package B.
func Test_heavyPreferredWithOr_notPreferSelection(t *testing.T) {
	ruleset := heavyPreferredWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageB").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 0,
			"packageB": 1,
			"itemX":    0,
			"itemY":    0,
		},
		solution,
	)
}

// Test_heavyPreferredWithOr_bothPackagesSelection
// Description: Package A requires (itemX AND itemY) and (A OR B) with A
// being preferred. We select package B and Package A expect both.
func Test_heavyPreferredWithOr_bothPackagesSelection(t *testing.T) {
	ruleset := heavyPreferredWithOr()

	selections := puan.Selections{
		puan.NewSelectionBuilder("packageA").Build(),
		puan.NewSelectionBuilder("packageB").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 1,
			"itemX":    1,
			"itemY":    1,
		},
		solution,
	)
}

func heavyPreferredWithOr() puan.Ruleset {
	creator := puan.NewRulesetCreator()

	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "itemY")
	itemXAndY, _ := creator.SetAnd("itemX", "itemY")
	packageARequiresItemXAndItemY, _ := creator.SetImply("packageA", itemXAndY)

	packageAOrPackageB, _ := creator.SetOr("packageA", "packageB")
	_ = creator.Assume(
		packageARequiresItemXAndItemY,
		packageAOrPackageB,
	)

	notPackageB, _ := creator.SetNot("packageB")
	packageAAndNotB, _ := creator.SetAnd("packageA", notPackageB)
	_ = creator.Prefer(packageAAndNotB)

	ruleset, _ := creator.Create()

	return ruleset
}
