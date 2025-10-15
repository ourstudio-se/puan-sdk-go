package puan

import (
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	weights2 "github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type Query struct {
	polyhedron *pldag.Polyhedron
	variables  []string
	weights    weights2.Weights
}

func NewQuery(polyhedron *pldag.Polyhedron, variables []string, weights weights2.Weights) *Query {
	return &Query{
		polyhedron: polyhedron,
		variables:  variables,
		weights:    weights,
	}
}

func (q *Query) Polyhedron() *pldag.Polyhedron {
	return q.polyhedron
}

func (q *Query) Variables() []string {
	return q.variables
}

func (q *Query) Weights() weights2.Weights {
	return q.weights
}
