package main

import (
	"fmt"
	"time"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

//nolint:gocyclo
func main() {
	creator := puan.NewRulesetCreator()

	// Enable time, and set start and end time
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)
	_ = creator.EnableTime(startTime, endTime)

	_ = creator.AddPrimitives([]string{"x", "y", "z"}...)

	xyID, err := creator.SetAnd("x", "y")
	if err != nil {
		panic(err)
	}

	xzID, err := creator.SetAnd("x", "z")
	if err != nil {
		panic(err)
	}

	endOfFirstPeriod := startTime.Add(30 * time.Minute)

	// Assume x AND y in the first period
	err = creator.AssumeInPeriod(xyID, startTime, endOfFirstPeriod)
	if err != nil {
		panic(err)
	}

	// Assume x AND z in the later period
	err = creator.AssumeInPeriod(xzID, endOfFirstPeriod, endTime)
	if err != nil {
		panic(err)
	}

	ruleSet, err := creator.Create()
	if err != nil {
		panic(err)
	}

	solutionCreator := puan.NewSolutionCreator(glpk.NewClient("http://127.0.0.1:9000"))

	inSecondPeriod := endOfFirstPeriod.Add(5 * time.Minute)
	envelope, err := solutionCreator.Create(
		nil,
		ruleSet,
		&inSecondPeriod,
	)
	if err != nil {
		panic(err)
	}
	solution := envelope.Solution()

	fmt.Println("x: ", solution["x"]) // = 1
	fmt.Println("y: ", solution["y"]) // = 0
	fmt.Println("z: ", solution["z"]) // = 1

	period, err := ruleSet.FindPeriodInSolution(solution)
	if err != nil {
		panic(err)
	}
	fmt.Printf("period: %s - %s\n", period.From(), period.To()) // endOfFirstPeriod-endTime
}
