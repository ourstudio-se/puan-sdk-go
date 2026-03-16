// nolint:lll
package solve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func Test_givenImplyRuleWithPreferredPeriods_shouldReturnPreferredItemForCurrentPeriod(t *testing.T) {
	item1 := fake.New[string]()
	item2 := fake.New[string]()
	condition := fake.New[string]()

	creator := puan.NewRulesetCreator()
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AddPrimitives(item1, item2, condition)

	xorID, _ := creator.SetXor(item1, item2)
	id, _ := creator.SetImply(condition, xorID)
	_ = creator.Assume(id)

	preferredInFirstPeriod, _ := creator.SetImply(condition, item1)
	preferredInSecondPeriod, _ := creator.SetImply(condition, item2)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)
	_ = creator.PreferInPeriod(preferredInFirstPeriod, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(preferredInSecondPeriod, endOfFirstPeriod, endTime)

	ruleset, _ := creator.Create()

	envelope1, _ := solutionCreator.Create(
		puan.Selections{puan.NewSelectionBuilder(condition).Build()},
		ruleset,
		&startTime,
	)
	solution1 := envelope1.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      1,
			item2:      0,
			condition:  1,
			"period_0": 1,
			"period_1": 0,
		},
		solution1,
	)

	inSecondPeriod := endOfFirstPeriod.Add(5 * time.Minute)
	envelope2, _ := solutionCreator.Create(
		puan.Selections{puan.NewSelectionBuilder(condition).Build()},
		ruleset,
		&inSecondPeriod,
	)
	solution2 := envelope2.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      1,
			condition:  1,
			"period_0": 0,
			"period_1": 1,
		},
		solution2,
	)
}

func Test_givenImplyRulePreferredPeriodsWithGaps_shouldReturnPreferredItemForCurrentPeriod(t *testing.T) {
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

	preferredInPeriodOneAndThree, _ := creator.SetImply(id, item1)
	preferredInPeriodTwo, _ := creator.SetImply(id, item2)

	endOfFirstPeriod := startTime.Add(15 * time.Minute)
	// The second period is between 'endOfFirstPeriod' and 'startOfThirdPeriod'.
	startOfThirdPeriod := endTime.Add(-15 * time.Minute)

	_ = creator.PreferInPeriod(preferredInPeriodOneAndThree, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(preferredInPeriodTwo, endOfFirstPeriod, startOfThirdPeriod)
	_ = creator.PreferInPeriod(preferredInPeriodOneAndThree, startOfThirdPeriod, endTime)

	ruleset, _ := creator.Create()

	envelope1, _ := solutionCreator.Create(
		puan.Selections{puan.NewSelectionBuilder(condition).Build()},
		ruleset,
		&startTime)
	solution1 := envelope1.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      1,
			item2:      0,
			condition:  1,
			"period_0": 1,
			"period_1": 0,
			"period_2": 0,
		},
		solution1,
	)

	inSecondPeriod := endOfFirstPeriod.Add(5 * time.Minute)
	envelope2, _ := solutionCreator.Create(
		puan.Selections{puan.NewSelectionBuilder(condition).Build()},
		ruleset,
		&inSecondPeriod,
	)
	solution2 := envelope2.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      1,
			condition:  1,
			"period_0": 0,
			"period_1": 1,
			"period_2": 0,
		},
		solution2,
	)

	inThirdPeriod := startOfThirdPeriod.Add(5 * time.Minute)
	envelope3, _ := solutionCreator.Create(
		puan.Selections{puan.NewSelectionBuilder(condition).Build()},
		ruleset,
		&inThirdPeriod)
	solution3 := envelope3.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      1,
			item2:      0,
			condition:  1,
			"period_0": 0,
			"period_1": 0,
			"period_2": 1,
		},
		solution3,
	)
}

func Test_givenImplyRuleWithPreferredPeriods_noSelection_shouldOnlyReturnPeriod(t *testing.T) {
	item1 := fake.New[string]()
	item2 := fake.New[string]()
	condition := fake.New[string]()

	creator := puan.NewRulesetCreator()
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AddPrimitives(item1, item2, condition)

	xorID, _ := creator.SetXor(item1, item2)
	id, _ := creator.SetImply(condition, xorID)
	_ = creator.Assume(id)

	preferredInFirstPeriod, _ := creator.SetImply(condition, item1)
	preferredInSecondPeriod, _ := creator.SetImply(condition, item2)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)
	_ = creator.PreferInPeriod(preferredInFirstPeriod, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(preferredInSecondPeriod, endOfFirstPeriod, endTime)

	ruleset, _ := creator.Create()

	envelope, _ := solutionCreator.Create(
		nil,
		ruleset,
		nil,
	)
	solution := envelope.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      0,
			condition:  0,
			"period_0": 1,
			"period_1": 0,
		},
		solution,
	)
}
