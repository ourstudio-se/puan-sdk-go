package solve

import (
	"testing"
	"time"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
)

// An item is included, but later is not. The solver should choose the earlier period
// with the item.
func Test_itemIncludedInPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime,
		startTime.Add(30*time.Minute))

	ruleSet, _ := creator.Create()

	query, _ := ruleSet.NewQuery(puan.QueryInput{})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"period_0": 1,
			"period_1": 0,
		},
		primitiveSolution,
	)
}

// Many items are included, but later non of them are. The solver should choose the later period,
// since the cost-savings of not having the items out weights the punishment of choosing the
// later period.
func Test_manyItemsIncludedInPeriod_shouldChooseLaterPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	items := []string{
		"item1",
		"item2",
		"item3",
		"item4",
		"item5",
		"item6",
	}
	_ = creator.AddPrimitives(items...)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	includeAll, _ := creator.SetAnd(items...)
	_ = creator.AssumeInPeriod(
		includeAll,
		startTime,
		startTime.Add(30*time.Minute))

	ruleSet, _ := creator.Create()

	query, _ := ruleSet.NewQuery(puan.QueryInput{})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	cleanedSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"item1":    0,
			"item2":    0,
			"item3":    0,
			"item4":    0,
			"item5":    0,
			"item6":    0,
			"period_0": 0,
			"period_1": 1,
		},
		cleanedSolution,
	)
}

// Item is included in later period. The solver should choose the earlier period,
// since it is cheaper both because of less items and because it is an earlier period.
func Test_itemIncludedInLaterPeriod_shouldChooseEarlierPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime.Add(30*time.Minute),
		endTime)

	ruleSet, _ := creator.Create()

	query, _ := ruleSet.NewQuery(puan.QueryInput{})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    0,
			"period_0": 1,
			"period_1": 0,
		},
		primitiveSolution,
	)
}

// Item is included in later period, and `from` is within that later period. The solver
// should choose the later period, as the earlier is forbidden.
func Test_itemIncludedInLaterPeriod_andFromInLaterPeriod_shouldChooseLaterPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime.Add(30*time.Minute),
		endTime)

	ruleSet, _ := creator.Create()

	from := startTime.Add(45 * time.Minute)
	query, _ := ruleSet.NewQuery(puan.QueryInput{From: &from})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	cleanedSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"period_0": 0,
			"period_1": 1,
		},
		cleanedSolution,
	)
}

// Item is included in later period. `from` is within the earlier period. The solver should
// choose the earlier period, since it is cheaper both because of less items and because
// it is an earlier period.
func Test_itemIncludedInLaterPeriod_andFromInEarlierPeriod_shouldChooseEarlierPeriod(
	t *testing.T,
) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime.Add(30*time.Minute),
		endTime)

	ruleSet, _ := creator.Create()

	from := startTime.Add(15 * time.Minute)
	query, _ := ruleSet.NewQuery(puan.QueryInput{From: &from})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	cleanedSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    0,
			"period_0": 1,
			"period_1": 0,
		},
		cleanedSolution,
	)
}

// An item is only available during a period. When the item is selected, the solver should
// choose that period.
func Test_itemSelectableInPeriod_givenItemSelected_shouldChoosePeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)

	notX, _ := creator.SetNot("itemX")
	_ = creator.AssumeInPeriod(
		notX,
		startTime,
		startTime.Add(15*time.Minute))
	_ = creator.AssumeInPeriod(
		notX,
		endTime.Add(-15*time.Minute),
		endTime)

	ruleSet, _ := creator.Create()

	query, _ := ruleSet.NewQuery(puan.QueryInput{
		Selections: puan.Selections{
			puan.NewSelectionBuilder("itemX").Build(),
		},
	})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"period_0": 0,
			"period_1": 1,
			"period_2": 0,
		},
		primitiveSolution,
	)
}

// An item is only available during a period. In that period, many items are included.
// When the item is selected, the solver should choose that period, even though
// all other items are included.
// nolint:lll
func Test_itemSelectableInPeriod_andManyItemsIncludedInThatPeriod_givenItemSelected_shouldChoosePeriod(
	t *testing.T,
) {
	creator := puan.NewRuleSetCreator()

	selectableItems := []string{"itemX"}
	includedItems := []string{
		"item1",
		"item2",
		"item3",
		"item4",
		"item5",
		"item6",
		"item7",
		"item8",
		"item9",
		"item10",
		"item11",
		"item12",
	}
	_ = creator.AddPrimitives(selectableItems...)
	_ = creator.AddPrimitives(includedItems...)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)

	periodWithRules, _ := puan.NewPeriod(
		startTime.Add(15*time.Minute),
		endTime.Add(-15*time.Minute),
	)

	// Make item selectable in period with rules
	notX, _ := creator.SetNot("itemX")
	_ = creator.AssumeInPeriod(
		notX,
		startTime,
		periodWithRules.From())
	_ = creator.AssumeInPeriod(
		notX,
		periodWithRules.To(),
		endTime)

	// Include all items in period with rules
	includeAll, _ := creator.SetAnd(includedItems...)
	_ = creator.AssumeInPeriod(
		includeAll,
		periodWithRules.From(),
		periodWithRules.To(),
	)

	ruleSet, _ := creator.Create()

	query, _ := ruleSet.NewQuery(puan.QueryInput{
		Selections: puan.Selections{
			puan.NewSelectionBuilder("itemX").Build(),
		},
	})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"item1":    1,
			"item2":    1,
			"item3":    1,
			"item4":    1,
			"item5":    1,
			"item6":    1,
			"item7":    1,
			"item8":    1,
			"item9":    1,
			"item10":   1,
			"item11":   1,
			"item12":   1,
			"period_0": 0,
			"period_1": 1,
			"period_2": 0,
		},
		primitiveSolution,
	)
}

// Exactly one of two packages are included in a period, where the expensive package is preferred.
// The solver should choose the earlier period with the preferred package.
// nolint:lll
func Test_includedPackageInEarlierPeriod_withPreferred_shouldChooseEarlierPeriodWithPreferredPackage(
	t *testing.T,
) {
	creator := puan.NewRuleSetCreator()

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AddPrimitives("packageA", "packageB", "itemX", "itemY", "itemZ", "itemW")

	packageAContent, _ := creator.SetAnd("itemX", "itemY", "itemZ")
	packageARequiresContent, _ := creator.SetImply("packageA", packageAContent)

	packageBContent := "itemW"
	packageBRequiresContent, _ := creator.SetImply("packageB", packageBContent)

	includeExactlyOnePackage, _ := creator.SetXor("packageA", "packageB")

	_ = creator.Assume(packageARequiresContent, packageBRequiresContent)

	preferredPackage, _ := creator.SetImply(includeExactlyOnePackage, "packageA")
	_ = creator.Prefer(preferredPackage)

	_ = creator.AssumeInPeriod(
		includeExactlyOnePackage,
		startTime,
		startTime.Add(30*time.Minute))

	ruleSet, _ := creator.Create()

	query, _ := ruleSet.NewQuery(puan.QueryInput{})

	client := glpk.NewClient(url)
	solution, _ := client.Solve(query)
	primitiveSolution, _ := ruleSet.RemoveSupportVariables(solution)
	assert.Equal(
		t,
		puan.Solution{
			"packageA": 1,
			"packageB": 0,
			"itemX":    1,
			"itemY":    1,
			"itemZ":    1,
			"itemW":    0,
			"period_0": 1,
			"period_1": 0,
		},
		primitiveSolution,
	)
}
