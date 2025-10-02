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

	// Sets x, y, z as boolean primitive variables
	_ = creator.SetPrimitives([]string{"x", "y", "z"}...)

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
	ruleSet, err := creator.Create()
	if err != nil {
		panic(err)
	}

	// Custom selections, which in this specific case will override the preferred variable z
	selections := puan.Selections{
		puan.NewSelectionBuilder("y").Build(),
	}

	// Create the query for solver
	query, err := ruleSet.NewQuery(selections)
	if err != nil {
		panic(err)
	}

	// Solve
	client := glpk.NewClient("http://127.0.0.1:9000")
	solution, err := client.Solve(query)
	if err != nil {
		panic(err)
	}

	// Extract the solution for the primitive variables
	primitiveSolution, err := solution.Extract(ruleSet.PrimitiveVariables()...)
	if err != nil {
		panic(err)
	}
	fmt.Println("x: ", primitiveSolution["x"]) // = 1
	fmt.Println("y: ", primitiveSolution["y"]) // = 1
	fmt.Println("z: ", primitiveSolution["z"]) // = 0
}
