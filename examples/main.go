package main

import (
	"fmt"
	"log"

	"github.com/ourstudio-se/puan-sdk-go/domain/pldag"
	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
)

func main() {
	model := pldag.New()
	model.SetPrimitives([]string{"x", "y"}...)
	_, err := model.SetAnd([]string{"x", "y"}...)
	if err != nil {
		panic(err)
	}

	variables := model.Variables()
	polyhedron := model.GeneratePolyhedron()

	tmpObjective := glpk.Objective{}
	for _, v := range variables {
		tmpObjective[v] = 1
	}

	client := glpk.NewClient("http://127.0.0.1:9000")
	solution, err := client.Solve(polyhedron, model.Variables(), tmpObjective)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("solution: ", solution)
}
