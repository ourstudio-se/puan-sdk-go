package main

import (
	"fmt"

	"github.com/ourstudio-se/puan-sdk-go/pldag"
)

func main() {
	model := pldag.New()
	model.SetPrimitives([]string{"x", "y"}...)
	_, err := model.SetAnd([]string{"x", "y"}...)
	if err != nil {
		panic(err)
	}

	polyhedron := model.GeneratePolyhedron()
	fmt.Println(polyhedron)
}
