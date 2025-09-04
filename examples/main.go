package main

import (
	"fmt"

	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

func main() {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives([]string{"a", "x", "y"}...)
	creator.PLDAG().Assume()

	ruleSet := creator.Create()
	x := "x"

	selections := puan.Selections{
		puan.NewSelection(puan.ADD, "a", &x),
	}

	selectedIDs, err := ruleSet.CalculateSelectedIDs(selections)
	if err != nil {
		panic(err)
	}

	objective := puan.CalculateObjective(
		ruleSet.PrimitiveVariables(),
		selectedIDs,
		nil,
	)

	client := glpk.NewClient("http://127.0.0.1:9000")
	solution, err := client.Solve(ruleSet.Polyhedron(), ruleSet.Variables(), objective)
	if err != nil {
		panic(err)
	}

	fmt.Println("solution: ", solution)
}
