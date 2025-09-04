package puan

import "github.com/ourstudio-se/puan-sdk-go/domain/pldag"

type Query struct {
	polyhedron *pldag.Polyhedron
	variables  []string
	objective  Weights
}

func NewQuery(polyhedron *pldag.Polyhedron, variables []string, objective Weights) *Query {
	return &Query{
		polyhedron: polyhedron,
		variables:  variables,
		objective:  objective,
	}
}
