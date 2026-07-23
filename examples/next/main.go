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

	// Create a simple and between x and y
	rule1, err := creator.SetXor("x", "y")
	if err != nil {
		panic(err)
	}

	// Enforces the connective to be true
	err = creator.Assume(rule1)
	if err != nil {
		panic(err)
	}

	// Create the ruleset
	ruleset, err := creator.Create()
	if err != nil {
		panic(err)
	}

	// Custom selections, which in this specific case will override the preferred variable z
	selections := puan.Selections{
		puan.NewSelectionBuilder("y").Build(),
	}

	// Create a solution creator with a solver client
	solverClient := solver.NewClient("http://127.0.0.1:9000", "1234567890", &http.Client{})
	solutionCreator := puan.NewSolutionCreator(solverClient)

	// Create the solution
	query := puan.NewSolutionQueryBuilder().
		WithSelections(selections).
		WithRuleset(ruleset).
		Build()
	envelope, err := solutionCreator.CreateNextSolutions(query)
	if err != nil {
		panic(err)
	}

	solutionForX, _ := envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("x").Build(),
	)
	fmt.Println("x: ", solutionForX.Solution()) // = {x: 1, y: 0, z: 0}

	solutionForY, _ := envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("y").WithAction(puan.REMOVE).Build(),
	)
	fmt.Println("y: ", solutionForY.Solution()) // = {x: 1, y: 0, z: 0}

	solutionForZ, _ := envelope.GetSolutionBySelection(
		puan.NewSelectionBuilder("z").Build(),
	)
	fmt.Println("z: ", solutionForZ.Solution()) // = {x: 0, y: 1, z: 1}
}
