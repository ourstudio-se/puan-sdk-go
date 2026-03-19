// nolint:lll
package solve

import (
	"testing"
	"time"

	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Many items are included, but later not.
// The solver should choose the earliest period despite the many items.
func Test_manyItemsIncludedInPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

	items := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
	_ = creator.AddPrimitives(items...)
	allIncluded, _ := creator.SetAnd(items...)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AssumeInPeriod(
		allIncluded,
		startTime,
		startTime.Add(30*time.Minute),
	)

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(nil, ruleset, nil)
	solution := envelope.Solution()

	asserter := newSolutionAsserter(solution)
	asserter.assertActive(t, "period_0")
	asserter.assertInactive(t, "period_1")
	asserter.assertActive(t, items...)
}

// Items are included in later period. The solver should choose the earlier period.
func Test_itemsIncludedInLaterPeriod_shouldChooseEarlierPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

	items := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
	_ = creator.AddPrimitives(items...)
	allIncluded, _ := creator.SetAnd(items...)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AssumeInPeriod(
		allIncluded,
		startTime.Add(30*time.Minute),
		endTime,
	)

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(nil, ruleset, nil)
	solution := envelope.Solution()

	asserter := newSolutionAsserter(solution)
	asserter.assertActive(t, "period_0")
	asserter.assertInactive(t, "period_1")
	asserter.assertInactive(t, items...)
}

// Item is included in later period, and `from` is within that later period. The solver
// should choose the later period, as the earlier is forbidden.
func Test_itemsIncludedInLaterPeriod_andFromInLaterPeriod_shouldChooseLaterPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()

	items := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
	_ = creator.AddPrimitives(items...)
	allIncluded, _ := creator.SetAnd(items...)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AssumeInPeriod(
		allIncluded,
		startTime.Add(30*time.Minute),
		endTime,
	)

	ruleset, _ := creator.Create()

	from := startTime.Add(45 * time.Minute)

	envelope, _ := solutionCreator.Create(nil, ruleset, &from)
	solution := envelope.Solution()

	asserter := newSolutionAsserter(solution)
	asserter.assertInactive(t, "period_0")
	asserter.assertActive(t, "period_1")
	asserter.assertActive(t, items...)
}

// Items are included in later period. `from` is within the earlier period.
// The solver should choose the earlier period.
func Test_itemsIncludedInLaterPeriod_andFromInEarlierPeriod_shouldChooseEarlierPeriod(
	t *testing.T,
) {
	creator := puan.NewRulesetCreator()

	items := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
	_ = creator.AddPrimitives(items...)
	allIncluded, _ := creator.SetAnd(items...)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AssumeInPeriod(
		allIncluded,
		startTime.Add(30*time.Minute),
		endTime,
	)

	ruleset, _ := creator.Create()

	from := startTime.Add(15 * time.Minute)

	envelope, _ := solutionCreator.Create(nil, ruleset, &from)
	solution := envelope.Solution()

	asserter := newSolutionAsserter(solution)
	asserter.assertActive(t, "period_0")
	asserter.assertInactive(t, "period_1")
	asserter.assertInactive(t, items...)
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
	includedItems := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
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

	asserter := newSolutionAsserter(solution)
	asserter.assertActive(t, "itemX")
	asserter.assertActive(t, "period_1")
	asserter.assertActive(t, includedItems...)
	asserter.assertInactive(t, "period_0")
	asserter.assertInactive(t, "period_2")
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

	_ = creator.AddPrimitives("packageA", "packageB")

	itemsInPackageA := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
	_ = creator.AddPrimitives(itemsInPackageA...)
	packageAContent, _ := creator.SetAnd(itemsInPackageA...)
	packageARequiresContent, _ := creator.SetImply("packageA", packageAContent)

	packageBContent := fake.New[string]()
	_ = creator.AddPrimitives(packageBContent)
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

	asserter := newSolutionAsserter(solution)
	asserter.assertActive(t, "packageA")
	asserter.assertActive(t, "period_0")
	asserter.assertActive(t, itemsInPackageA...)
	asserter.assertInactive(t, "packageB")
	asserter.assertInactive(t, packageBContent)
	asserter.assertInactive(t, "period_1")
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

// Global XOR rule for item1 and item2,
// item2 has many consequences in the first period.
// The solver should choose the first period when item2 is selected.
func Test_givenXORWithManyConsequencesInFirstPeriod_selectExpensiveItem_shouldChooseFirstPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	item1 := fake.New[string]()
	item2 := fake.New[string]()
	_ = creator.AddPrimitives(item1, item2)
	xorID, _ := creator.SetXor(item1, item2)
	_ = creator.Assume(xorID)

	item2Consequences := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
	_ = creator.AddPrimitives(item2Consequences...)

	consequenceID, _ := creator.SetAnd(item2Consequences...)
	item2Consequence, _ := creator.SetImply(item2, consequenceID)
	endOfFirstPeriod := startTime.Add(30 * time.Minute)
	_ = creator.AssumeInPeriod(item2Consequence, startTime, endOfFirstPeriod)

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(
		puan.Selections{
			puan.NewSelectionBuilder(item2).Build(),
		},
		ruleset,
		&startTime,
	)

	solution := envelope.Solution()

	asserter := newSolutionAsserter(solution)
	asserter.assertActive(t, item2)
	asserter.assertActive(t, "period_0")
	asserter.assertActive(t, item2Consequences...)
	asserter.assertInactive(t, item1)
	asserter.assertInactive(t, "period_1")
}

// Global XOR rules for item1 and many other items,
// all other items are preferred in the first period.
// The solver should choose the first period when item1 is selected.
func Test_givenXORWithManyPreferredInFirstPeriod_selectNonPreferredItem_shouldChooseFirstPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	item1 := fake.New[string]()
	_ = creator.AddPrimitives(item1)

	otherItems := fake.New[[]string](
		func(oo *options.Options) {
			oo.RandomMinSliceSize = 50
			oo.RandomMaxSliceSize = 50
		},
	)
	_ = creator.AddPrimitives(otherItems...)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)
	for _, otherItem := range otherItems {
		xorID, _ := creator.SetXor(item1, otherItem)
		_ = creator.Assume(xorID)
		preferredOtherItem, _ := creator.SetImply(xorID, otherItem)
		_ = creator.PreferInPeriod(preferredOtherItem, startTime, endOfFirstPeriod)
	}

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(
		puan.Selections{
			puan.NewSelectionBuilder(item1).Build(),
		},
		ruleset,
		&startTime,
	)

	solution := envelope.Solution()

	asserter := newSolutionAsserter(solution)
	asserter.assertActive(t, item1)
	asserter.assertActive(t, "period_0")
	asserter.assertInactive(t, otherItems...)
	asserter.assertInactive(t, "period_1")
}

type solutionAsserter struct {
	puan.Solution
}

func newSolutionAsserter(solution puan.Solution) solutionAsserter {
	return solutionAsserter{solution}
}

func (s solutionAsserter) assertActive(t *testing.T, variables ...string) {
	solution := s.Extract(variables...)
	for _, variable := range variables {
		value, ok := solution[variable]
		if !ok {
			assert.Failf(t, "variable %s not found in solution", variable)
		}

		assert.Equal(t, 1, value, "expected %s to be active", variable)
	}
}

func (s solutionAsserter) assertInactive(t *testing.T, variables ...string) {
	solution := s.Extract(variables...)
	for _, variable := range variables {
		value, ok := solution[variable]
		if !ok {
			assert.Failf(t, "variable %s not found in solution", variable)
		}

		assert.Equal(t, 0, value, "expected %s to be inactive", variable)
	}
}
