package main

import (
	"fmt"

	"github.com/ourstudio-se/puan-sdk-go/pldag"
)

func main() {
	//model := pldag.New()
	//model.SetPrimities([]string{"x", "y", "z"}...)
	//model.SetAnd([]string{"x", "y"}...)
	//
	//fmt.Println(model)

	model := pldag.New()
	model.SetPrimitives([]string{"x", "y"}...)
	ref := model.SetAnd([]string{"x", "y"}...)

	system := model.GenerateSystem([]string{ref}...)
	fmt.Println(system)
}
