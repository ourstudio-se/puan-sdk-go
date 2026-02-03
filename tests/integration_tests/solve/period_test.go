package solve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// An item is included, but later is not. The solver should choose the earlier period
// with the item.
func Test_itemIncludedInPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime,
		startTime.Add(30*time.Minute))

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(nil, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"period_0": 1,
			"period_1": 0,
		},
		solution,
	)
}

// Many items are included, but later none of them are. The solver should choose the later period,
// since the cost-savings of not having the items out weights the punishment of choosing the
// later period.
func Test_manyItemsIncludedInPeriod_shouldChooseLaterPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

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

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(nil, ruleset, nil)
	solution := envelope.Solution()
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
		solution,
	)
}

// Item is included in later period. The solver should choose the earlier period,
// since it is cheaper both because of less items and because it is an earlier period.
func Test_itemIncludedInLaterPeriod_shouldChooseEarlierPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime.Add(30*time.Minute),
		endTime)

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(nil, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    0,
			"period_0": 1,
			"period_1": 0,
		},
		solution,
	)
}

// Item is included in later period, and `from` is within that later period. The solver
// should choose the later period, as the earlier is forbidden.
func Test_itemIncludedInLaterPeriod_andFromInLaterPeriod_shouldChooseLaterPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime.Add(30*time.Minute),
		endTime)

	ruleset, _ := creator.Create()

	from := startTime.Add(45 * time.Minute)

	envelope, _ := solutionCreator.Create(nil, ruleset, &from)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"period_0": 0,
			"period_1": 1,
		},
		solution,
	)
}

// Item is included in later period. `from` is within the earlier period. The solver should
// choose the earlier period, since it is cheaper both because of less items and because
// it is an earlier period.
func Test_itemIncludedInLaterPeriod_andFromInEarlierPeriod_shouldChooseEarlierPeriod(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	_ = creator.EnableTime(startTime, endTime)
	_ = creator.AssumeInPeriod(
		"itemX",
		startTime.Add(30*time.Minute),
		endTime)

	ruleset, _ := creator.Create()

	from := startTime.Add(15 * time.Minute)

	envelope, _ := solutionCreator.Create(nil, ruleset, &from)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    0,
			"period_0": 1,
			"period_1": 0,
		},
		solution,
	)
}

// An item is only available during a period. When the item is selected, the solver should
// choose that period.
func Test_itemSelectableInPeriod_givenItemSelected_shouldChoosePeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

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

	ruleset, _ := creator.Create()
	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"period_0": 0,
			"period_1": 1,
			"period_2": 0,
		},
		solution,
	)
}

// An item is only available during a period. In that period, many items are included.
// When the item is selected, the solver should choose that period, even though
// all other items are included.
// nolint:lll
func Test_itemSelectableInPeriod_andManyItemsIncludedInThatPeriod_givenItemSelected_shouldChoosePeriod(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

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

	ruleset, _ := creator.Create()

	selections := puan.Selections{
		puan.NewSelectionBuilder("itemX").Build(),
	}

	envelope, _ := solutionCreator.Create(selections, ruleset, nil)
	solution := envelope.Solution()
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
		solution,
	)
}

// Exactly one of two packages are included in a period, where the expensive package is preferred.
// The solver should choose the earlier period with the preferred package.
// nolint:lll
func Test_includedPackageInEarlierPeriod_withPreferred_shouldChooseEarlierPeriodWithPreferredPackage(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

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

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(nil, ruleset, nil)
	solution := envelope.Solution()

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
		solution,
	)
}

// Time is enabled, but no variables are assumed in a period.
// `from` is not specified. The solution should contain the ruleset's
// period as default.
func Test_givenTimeEnabledWithoutTimeboundConstraints_andNoFromSpecified_shouldGetRulesetPeriod(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AddPrimitives("itemX", "itemY")
	xOrY, _ := creator.SetOr("itemX", "itemY")
	_ = creator.Assume(xOrY)

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(nil, ruleset, nil)
	solution := envelope.Solution()

	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"itemY":    0,
			"period_0": 1,
		},
		solution,
	)
}

// Time is enabled, but no variables are assumed in a period.
// `from` is before the ruleset's period. The solution should
// "jump forward" to the ruleset's period.
func Test_givenTimeEnabledWithoutTimeboundConstraints_andEarlyFromSpecified_shouldGetRulesetPeriod(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AddPrimitives("itemX", "itemY")
	xOrY, _ := creator.SetOr("itemX", "itemY")
	_ = creator.Assume(xOrY)

	ruleset, _ := creator.Create()

	beforeStart := startTime.Add(-1 * time.Hour)
	envelope, _ := solutionCreator.Create(nil, ruleset, &beforeStart)
	solution := envelope.Solution()

	assert.Equal(
		t,
		puan.Solution{
			"itemX":    1,
			"itemY":    0,
			"period_0": 1,
		},
		solution,
	)
}

// Time is enabled, but no variables are assumed in a period.
// `from` is after the ruleset's period. The solution should
// return an error, since this is not allowed.
func Test_givenTimeEnabledWithoutTimeboundConstraints_andLateFromSpecified_shouldReturnError(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AddPrimitives("itemX", "itemY")
	xOrY, _ := creator.SetOr("itemX", "itemY")
	_ = creator.Assume(xOrY)

	ruleset, _ := creator.Create()

	afterEnd := endTime.Add(1 * time.Hour)
	_, err := solutionCreator.Create(nil, ruleset, &afterEnd)
	assert.Error(t, err)
}
