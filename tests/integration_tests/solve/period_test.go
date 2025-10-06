package solve

import (
	"testing"
	"time"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/stretchr/testify/assert"
)

func Test_optionRequiredInPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	creator.EnableTime(startTime, endTime)
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

func Test_manyOptionsRequiredInPeriod_shouldChooseLaterPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX", "itemY", "itemZ")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	creator.EnableTime(startTime, endTime)
	xyAndZ, _ := creator.SetAnd("itemX", "itemY", "itemZ")
	_ = creator.AssumeInPeriod(
		xyAndZ,
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
			"itemX":    0,
			"itemY":    0,
			"itemZ":    0,
			"period_0": 0,
			"period_1": 1,
		},
		cleanedSolution,
	)
}

func Test_optionRequiredInLaterPeriod_shouldChooseEarlierPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	creator.EnableTime(startTime, endTime)
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

func Test_optionRequiredInLaterPeriod_andFromInLaterPeriod_shouldChooseLaterPeriod(t *testing.T) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	creator.EnableTime(startTime, endTime)
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

func Test_optionRequiredInLaterPeriod_andFromInEarlierPeriod_shouldChooseEarlierPeriod(
	t *testing.T,
) {
	creator := puan.NewRuleSetCreator()

	_ = creator.AddPrimitives("itemX")

	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	creator.EnableTime(startTime, endTime)
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
