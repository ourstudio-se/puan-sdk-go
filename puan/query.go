package puan

import (
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/internal/weights"
)

type Query struct {
	polyhedron *pldag.Polyhedron
	variables  []string
	weights    weights.Weights
}

func NewQuery(polyhedron *pldag.Polyhedron, variables []string, weights weights.Weights) *Query {
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

func (q *Query) Weights() weights.Weights {
	return q.weights
}
