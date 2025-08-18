package main

import (
	"fmt"
	"log"

	"github.com/ourstudio-se/puan-sdk-go/gateway/glpk"
	"github.com/ourstudio-se/puan-sdk-go/pldag"
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
	resp, err := client.Solve(polyhedron, model.Variables(), tmpObjective)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("error: ", resp.Solutions[0].Error)
	fmt.Println("objective: ", resp.Solutions[0].Objective)
	fmt.Println("solution: ", resp.Solutions[0].Solution)
	fmt.Println("status: ", resp.Solutions[0].Status)
}
