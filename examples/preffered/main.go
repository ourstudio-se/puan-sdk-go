package main

import (
	"fmt"

	"github.com/ourstudio-se/puan-sdk-go/internal/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

//nolint:gocyclo
func main() {
	// Initialize the ruleset creator
	creator := puan.NewRuleSetCreator()

	// Adds x, y, z as boolean primitive variables
	_ = creator.AddPrimitives([]string{"x", "y", "z"}...)

	// Create a simple and between x and y
	xyID, err := creator.SetAnd("x", "y")
	if err != nil {
		panic(err)
	}

	// Create a simple and between x and z
	xzID, err := creator.SetAnd("x", "z")
	if err != nil {
		panic(err)
	}

	// Either x with y or x with z
	xorID, err := creator.SetXor(xyID, xzID)
	if err != nil {
		panic(err)
	}

	// Enforces the connective to be true
	err = creator.Assume(xorID)
	if err != nil {
		panic(err)
	}

	// Set z as preferred if no variable is selected
	err = creator.Prefer("z")
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
	solutionCreator := puan.NewSolutionCreator(glpk.NewClient("http://127.0.0.1:9000"))

	// Create the solution
	envelope, err := solutionCreator.Create(selections, ruleset, nil)
	if err != nil {
		panic(err)
	}
	solution := envelope.Solution()

	fmt.Println("x: ", solution["x"]) // = 1
	fmt.Println("z: ", solution["z"]) // = 0
	fmt.Println("y: ", solution["y"]) // = 1
}
