// nolint:lll
package solve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ourstudio-se/puan-sdk-go/internal/fake"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

func Test_givenXORWithDifferentPreferredPeriods_shouldReturnPreferredItemForCurrentPeriod(t *testing.T) {
	creator := puan.NewRulesetCreator()
	item1 := fake.New[string]()
	item2 := fake.New[string]()

	_ = creator.AddPrimitives(item1, item2)

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	xorID, _ := creator.SetXor(item1, item2)
	_ = creator.Assume(xorID)

	endOfFirstPeriod := startTime.Add(30 * time.Minute)
	_ = creator.PreferInPeriod(item1, startTime, endOfFirstPeriod)
	_ = creator.PreferInPeriod(item2, endOfFirstPeriod, endTime)

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
