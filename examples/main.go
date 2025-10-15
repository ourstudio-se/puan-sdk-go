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
	// and free-variable as a variable which will be freely selected
	// since it is not part of any logic.
	_ = creator.AddPrimitives([]string{"x", "y", "z", "free-variable"}...)

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
		puan.NewSelectionBuilder("free-variable").Build(),
	}

	// Client
	client := glpk.NewClient("http://127.0.0.1:9000")

	// Solve with the client, ruleset and selections
	solution, err := puan.Solve(client, ruleSet, selections, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("x: ", solution["x"])                         // = 1
	fmt.Println("y: ", solution["y"])                         // = 1
	fmt.Println("z: ", solution["z"])                         // = 0
	fmt.Println("free-variable: ", solution["free-variable"]) // = 1
}
