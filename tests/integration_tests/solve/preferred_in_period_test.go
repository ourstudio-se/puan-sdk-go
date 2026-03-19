// nolint:lll
package solve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

// Global conditional rule with preferred the changes in periods. When selecting the condition in the second period,
// the preferred item should be the one for the second period.
func Test_givenConditionalRuleWithDifferentPreferreds_shouldReturnPreferredItemForCurrentPeriod(t *testing.T) {
	item1 := fake.New[string]()
	item2 := fake.New[string]()
	condition := fake.New[string]()

	creator := puan.NewRulesetCreator()
	_ = creator.AddPrimitives(item1, item2, condition)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	xorID, _ := creator.SetXor(item1, item2)
	id, _ := creator.SetImply(condition, xorID)

	_ = creator.Assume(id)

	preferredInPeriodOne, _ := creator.SetImply(id, item1)
	preferredInPeriodTwo, _ := creator.SetImply(id, item2)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)

	_ = creator.PreferInPeriod(preferredInPeriodOne, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(preferredInPeriodTwo, endOfFirstPeriod, endTime)

	ruleset, _ := creator.Create()

	inSecondPeriod := endOfFirstPeriod.Add(5 * time.Minute)
	envelope, _ := solutionCreator.Create(
		puan.Selections{puan.NewSelectionBuilder(condition).Build()},
		ruleset,
		&inSecondPeriod,
	)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      1,
			condition:  1,
			"period_0": 0,
			"period_1": 1,
		},
		solution,
	)
}

// Global XOR rule with preferred the changes in periods.
// When solving in the second period, the preferred item should be the one for the second period.
func Test_givenXORRuleWithDifferentPreferred_shouldReturnPreferredItemForCurrentPeriod(t *testing.T) {
	item1 := fake.New[string]()
	item2 := fake.New[string]()

	creator := puan.NewRulesetCreator()
	_ = creator.AddPrimitives(item1, item2)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	xorID, _ := creator.SetXor(item1, item2)
	_ = creator.Assume(xorID)

	preferredInPeriodOne, _ := creator.SetImply(xorID, item1)
	preferredInPeriodTwo, _ := creator.SetImply(xorID, item2)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)

	_ = creator.PreferInPeriod(preferredInPeriodOne, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(preferredInPeriodTwo, endOfFirstPeriod, endTime)

	ruleset, _ := creator.Create()

	inSecondPeriod := endOfFirstPeriod.Add(5 * time.Minute)
	envelope, _ := solutionCreator.Create(nil, ruleset, &inSecondPeriod)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      1,
			"period_0": 0,
			"period_1": 1,
		},
		solution,
	)
}
