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
	ref, err := model.SetAnd([]string{"x", "y"}...)
	if err != nil {
		panic(err)
	}

	_ = ref

	system := model.GenerateSystem()
	fmt.Println(system)
}
