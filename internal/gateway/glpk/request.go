package glpk

import (
	"github.com/ourstudio-se/puan-sdk-go/internal/pldag"
	"github.com/ourstudio-se/puan-sdk-go/puan"
)

const DefaultDirection = "maximize"

func newSolveRequestFromQuery(
	query *puan.SolverQuery,
) SolveRequest {
	objective := Objective(query.Weights())

	request := newSolveRequest(
		query.Polyhedron(),
		query.Variables(),
		objective,
	)

	return request
}

func newSolveRequestFromMultiQuery(
	query *puan.MultiWeightSolverQuery,
) SolveRequest {
	objectives := make([]Objective, len(query.WeightGroups()))
	for i, group := range query.WeightGroups() {
		objectives[i] = Objective(group)
	}

	request := newSolveRequest(
		query.Polyhedron(),
		query.Variables(),
		objectives...,
	)

	return request
}

func newSolveRequest(
	polyhedron *pldag.Polyhedron,
	variableIDs []string,
	objectives ...Objective,
) SolveRequest {
	A := toSparseMatrix(polyhedron.SparseMatrix())
	b := polyhedron.B()
	variables := toBooleanVariables(variableIDs)

	request := SolveRequest{
		Polyhedron: Polyhedron{
			A:         A,
			B:         b,
			Variables: variables,
		},
		Objectives: objectives,
		Direction:  DefaultDirection,
	}

	return request
}
