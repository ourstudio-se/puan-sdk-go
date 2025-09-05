package main

import (
	"fmt"

	"github.com/ourstudio-se/puan-sdk-go/domain/puan"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

func main() {
	creator := puan.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives([]string{"a", "x", "y"}...)
	preferred, _ := creator.PLDAG().SetAnd("a", "y")
	_ = creator.SetPreferreds(preferred)
	creator.PLDAG().Assume(preferred)

	ruleSet := creator.Create()
	x := "x"

	selections := puan.Selections{
		puan.NewSelectionBuilder("a").WithSubSelectionID(x).Build(),
	}

	query, err := ruleSet.NewQuery(selections)
	if err != nil {
		panic(err)
	}

	client := glpk.NewClient("http://127.0.0.1:9000")
	solution, err := client.Solve(query.Polyhedron(), query.Variables(), query.Objective())
	if err != nil {
		panic(err)
	}

	primitiveSolution, err := solution.Extract(ruleSet.PrimitiveVariables()...)
	if err != nil {
		panic(err)
	}

	fmt.Println("primitiveSolution: ", primitiveSolution)
}
