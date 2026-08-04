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

	// `y` is the current selection
	currentSelections := puan.Selections{
		puan.NewSelectionBuilder("y").Build(),
	}
	// Calculate many next selections, after `y` is already selected
	addX := puan.NewSelectionBuilder("x").Build()
	removeY := puan.NewSelectionBuilder("y").WithAction(puan.REMOVE).Build()
	addZ := puan.NewSelectionBuilder("z").Build()
	nextSelections := puan.Selections{
		addX,
		removeY,
		addZ,
	}

	// Create a solution creator with a solver client
	solverClient := solver.NewClient("http://127.0.0.1:9000", "1234567890", &http.Client{})
	solutionCreator := puan.NewSolutionCreator(solverClient)

	// Create solutions
	query := puan.NewNextSolutionsQuery(
		currentSelections,
		nextSelections,
		ruleset,
		nil,
		nil,
	)
	envelope, err := solutionCreator.CreateNextSolutions(query)
	if err != nil {
		panic(err)
	}

	solutionForX, _ := envelope.GetSolutionBySelection(addX)
	fmt.Println("x: ", solutionForX.Solution()) // = {x: 1, y: 0, z: 0}

	solutionForY, _ := envelope.GetSolutionBySelection(removeY)
	fmt.Println("y: ", solutionForY.Solution()) // = {x: 1, y: 0, z: 0}

	solutionForZ, _ := envelope.GetSolutionBySelection(addZ)
	fmt.Println("z: ", solutionForZ.Solution()) // = {x: 0, y: 1, z: 1}
}
