// nolint:lll
package solve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func Test_givenXORRuleWithDifferentPreferredPeriods_shouldReturnPreferredItemForCurrentPeriod(t *testing.T) {
	item1 := fake.New[string]()
	item2 := fake.New[string]()

	creator := puan.NewRulesetCreator()
	_ = creator.AddPrimitives(item1, item2)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	xorID, _ := creator.SetXor(item1, item2)
	_ = creator.Assume(xorID)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)

	preferredInPeriodOne, _ := creator.SetImply(xorID, item1)
	preferredInPeriodTwo, _ := creator.SetImply(xorID, item2)

	_ = creator.PreferInPeriod(preferredInPeriodOne, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(preferredInPeriodTwo, endOfFirstPeriod, endTime)

	ruleset, _ := creator.Create()

	envelope1, _ := solutionCreator.Create(nil, ruleset, &startTime)
	solution1 := envelope1.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      1,
			item2:      0,
			"period_0": 1,
			"period_1": 0,
		},
		solution1,
	)

	inSecondPeriod := endOfFirstPeriod.Add(5 * time.Minute)
	envelope2, _ := solutionCreator.Create(nil, ruleset, &inSecondPeriod)
	solution2 := envelope2.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      1,
			"period_0": 0,
			"period_1": 1,
		},
		solution2,
	)
}

func Test_givenXORRulePreferredPeriodsWithGaps_shouldReturnPreferredItemForCurrentPeriod(t *testing.T) {
	item1 := fake.New[string]()
	item2 := fake.New[string]()

	creator := puan.NewRulesetCreator()
	_ = creator.AddPrimitives(item1, item2)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	xorID, _ := creator.SetXor(item1, item2)
	_ = creator.Assume(xorID)

	preferredInPeriodOneAndThree, _ := creator.SetImply(xorID, item1)
	preferredInPeriodTwo, _ := creator.SetImply(xorID, item2)

	endOfFirstPeriod := startTime.Add(15 * time.Minute)
	// The second period is between the endOfFirstPeriod and the startOfThirdPeriod,
	startOfThirdPeriod := endTime.Add(-15 * time.Minute)

	_ = creator.PreferInPeriod(preferredInPeriodOneAndThree, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(preferredInPeriodTwo, endOfFirstPeriod, startOfThirdPeriod)
	_ = creator.PreferInPeriod(preferredInPeriodOneAndThree, startOfThirdPeriod, endTime)

	ruleset, _ := creator.Create()

	envelope1, _ := solutionCreator.Create(nil, ruleset, &startTime)
	solution1 := envelope1.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      1,
			item2:      0,
			"period_0": 1,
			"period_1": 0,
			"period_2": 0,
		},
		solution1,
	)

	inSecondPeriod := endOfFirstPeriod.Add(5 * time.Minute)
	envelope2, _ := solutionCreator.Create(nil, ruleset, &inSecondPeriod)
	solution2 := envelope2.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      1,
			"period_0": 0,
			"period_1": 1,
			"period_2": 0,
		},
		solution2,
	)

	inThirdPeriod := startOfThirdPeriod.Add(5 * time.Minute)
	envelope3, _ := solutionCreator.Create(nil, ruleset, &inThirdPeriod)
	solution3 := envelope3.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      1,
			item2:      0,
			"period_0": 0,
			"period_1": 0,
			"period_2": 1,
		},
		solution3,
	)
}

func Test_oscar(t *testing.T) {
	item1 := fake.New[string]()
	item2 := fake.New[string]()
	item3 := fake.New[string]()
	item4 := fake.New[string]()
	item5 := fake.New[string]()
	item6 := fake.New[string]()
	item7 := fake.New[string]()
	item8 := fake.New[string]()
	item9 := fake.New[string]()
	item10 := fake.New[string]()

	creator := puan.NewRulesetCreator()
	_ = creator.AddPrimitives(item1, item2, item3, item4, item5, item6, item7, item8, item9, item10)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	xorID, _ := creator.SetXor(item1, item2)

	orID, _ := creator.SetOr(item1, item2, item3, item4, item5, item6, item7, item8, item9, item10)
	_ = creator.Assume(xorID, orID)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)

	preferredInPeriodOne, _ := creator.SetImply(xorID, item1)

	_ = creator.PreferInPeriod(preferredInPeriodOne, startTime, endOfFirstPeriod)

	ruleset, _ := creator.Create()

	envelope1, _ := solutionCreator.Create(
		puan.Selections{
			puan.NewSelectionBuilder(item2).Build(),
		},
		ruleset, &startTime)
	solution1 := envelope1.Solution()
	assert.Equal(
		t,
		puan.Solution{
			item1:      0,
			item2:      1,
			item3:      0,
			item4:      0,
			item5:      0,
			item6:      0,
			item7:      0,
			item8:      0,
			item9:      0,
			item10:     0,
			"period_0": 1,
			"period_1": 0,
		},
		solution1,
	)
}
