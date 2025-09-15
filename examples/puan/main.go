package main

import (
	"fmt"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	puan2 "github.com/ourstudio-se/puan-sdk-go/puan"
)

//nolint:gocyclo
func main() {
	creator := puan2.NewRuleSetCreator()
	creator.PLDAG().SetPrimitives([]string{"a", "x", "y"}...)
	variant1, _ := creator.PLDAG().SetAnd("a", "x")
	variant2, _ := creator.PLDAG().SetAnd("a", "y")
	exactlyOneVariant, _ := creator.PLDAG().SetXor(variant1, variant2)
	a, err := creator.PLDAG().SetImply("a", exactlyOneVariant)
	if err != nil {
		panic(err)
	}

	err = creator.SetPreferreds(variant2)
	if err != nil {
		panic(err)
	}

	err = creator.PLDAG().Assume(a)
	if err != nil {
		panic(err)
	}

	ruleSet := creator.Create()
	x := "x"

	selections := puan2.Selections{
		puan2.NewSelectionBuilder("a").WithSubSelectionID(x).Build(),
	}

	query, err := ruleSet.NewQuery(selections)
	if err != nil {
		panic(err)
	}

	fmt.Println("variables length: ", len(query.Variables()))
	fmt.Println("B length: ", len(query.Polyhedron().B()))
	fmt.Println("A column length: ", query.Polyhedron().SparseMatrix().Shape().NrOfColumns())

	client := glpk.NewClient("http://127.0.0.1:9000")
	solution, err := client.Solve(query)
	if err != nil {
		panic(err)
	}

	primitiveSolution, err := solution.Extract(ruleSet.PrimitiveVariables()...)
	if err != nil {
		panic(err)
	}

	fmt.Println("primitiveSolution: ", primitiveSolution)
}
