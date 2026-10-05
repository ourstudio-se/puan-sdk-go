package main

import (
	"fmt"
	"net/http"

	"github.com/ourstudio-se/puan-sdk-go/puan"
	"github.com/ourstudio-se/puan-sdk-go/solver"
)

//nolint:gocyclo
func main() {
	// Initialize the ruleset creator
	creator := puan.NewRulesetCreator()

	// Adds x, y, z as boolean primitive variables
	_ = creator.AddPrimitives([]string{"x", "y", "z"}...)

	// Create an XOR between x and y
	rule1, err := creator.SetXor("x", "y")
	if err != nil {
		panic(err)
	}

	// Enforces the rule
	err = creator.Assume(rule1)
	if err != nil {
		panic(err)
	}

	// Create the ruleset
	ruleset, err := creator.Create()
	if err != nil {
		panic(err)
	}

	// 2 selection groups: [-x, z] and [y]
	selectionGroups := []puan.Selections{
		{
			puan.NewSelectionBuilder("y").Build(),
		},
		{
			puan.NewSelectionBuilder("x").WithAction(puan.REMOVE).Build(),
			puan.NewSelectionBuilder("z").Build(),
		},
	}

	// Create a solution creator with a solver client
	solverClient := solver.NewClient("http://127.0.0.1:9000", "1234567890", &http.Client{})
	solutionCreator := puan.NewSolutionCreator(solverClient)

	// Create solutions
	query, err := puan.NewManySolutionsQuery(
		selectionGroups,
		ruleset,
		nil,
		nil,
	)
	if err != nil {
		panic(err)
	}

	envelope, err := solutionCreator.CreateManySolutions(query)
	if err != nil {
		panic(err)
	}

	// Solutions for the first selection group: [y]
	solutionForSelections1 := envelope.SolutionsBySelectionGroup()[0]
	fmt.Println(solutionForSelections1.Solution()) // = {x: 0, y: 1, z: 0}

	// Solutions for the second selection group: [-x, z]
	solutionForSelections2 := envelope.SolutionsBySelectionGroup()[1]
	fmt.Println(solutionForSelections2.Solution()) // = {x: 0, y: 1, z: 1}
}
